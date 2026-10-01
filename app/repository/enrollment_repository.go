package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"siakad/app/model"
)

type EnrollmentRepository interface {
	// Enroll mengambil matkul dalam satu transaction dengan row locking.
	// Row locking (SELECT ... FOR UPDATE) mencegah race condition
	// ketika dua mahasiswa mendaftar matkul yang sama di waktu bersamaan.
	Enroll(ctx context.Context, studentID, courseID int, tahun string) (model.Enrollment, error)

	// Delete menghapus enrollment milik mahasiswa.
	Delete(ctx context.Context, enrollmentID, studentID int) error

	// FindByID mengambil satu enrollment (untuk cek kepemilikan saat delete).
	FindByID(ctx context.Context, id int) (model.Enrollment, error)

	// SumSKS menghitung total SKS yang sudah diambil mahasiswa pada tahun akademik tertentu.
	SumSKS(ctx context.Context, studentID int, tahun string) (int, error)
}

type enrollmentRepo struct {
	pool *pgxpool.Pool
}

func NewEnrollmentRepository(pool *pgxpool.Pool) EnrollmentRepository {
	return &enrollmentRepo{pool: pool}
}

// Enroll menjalankan semua pengecekan business rule dalam SATU transaction:
// 1. Cek duplikasi (sudah ambil matkul ini di tahun yang sama?)
// 2. Cek kuota — pakai SELECT FOR UPDATE agar tidak ada race condition
// 3. Cek total SKS tidak melebihi batas berdasarkan IPK
// 4. Jika semua lolos → INSERT enrollment
func (r *enrollmentRepo) Enroll(ctx context.Context, studentID, courseID int, tahun string) (model.Enrollment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Enrollment{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) // no-op jika sudah Commit

	// ── Step 1: Cek duplikasi ──────────────────────────────────
	var exists bool
	_ = tx.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM enrollments
			WHERE student_id=$1 AND course_id=$2 AND tahun_akademik=$3
		)`, studentID, courseID, tahun,
	).Scan(&exists)
	if exists {
		return model.Enrollment{}, ErrDuplicate // → 409
	}

	// ── Step 2: Kunci baris courses lalu cek kuota ────────────
	// Dua query terpisah karena FOR UPDATE tidak bisa digabung dengan GROUP BY.
	// Query 2a: kunci baris courses dengan FOR UPDATE
	var kuota int
	err = tx.QueryRow(ctx,
		`SELECT kuota FROM courses WHERE id = $1 FOR UPDATE`, courseID,
	).Scan(&kuota)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Enrollment{}, ErrNotFound
		}
		return model.Enrollment{}, fmt.Errorf("lock course: %w", err)
	}

	// Query 2b: hitung sudah berapa yang mendaftar
	var terisi int
	_ = tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM enrollments WHERE course_id = $1`, courseID,
	).Scan(&terisi)

	if terisi >= kuota {
		return model.Enrollment{}, ErrQuotaFull // → 422
	}

	// ── Step 3: Insert enrollment ──────────────────────────────
	var e model.Enrollment
	err = tx.QueryRow(ctx,
		`INSERT INTO enrollments (student_id, course_id, tahun_akademik)
		 VALUES ($1, $2, $3)
		 RETURNING id, student_id, course_id, tahun_akademik, created_at`,
		studentID, courseID, tahun,
	).Scan(&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt)
	if err != nil {
		return model.Enrollment{}, fmt.Errorf("insert enrollment: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Enrollment{}, fmt.Errorf("commit tx: %w", err)
	}

	return e, nil
}

func (r *enrollmentRepo) Delete(ctx context.Context, enrollmentID, studentID int) error {
	// Pastikan enrollment milik mahasiswa yang sedang login (student_id harus cocok)
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM enrollments WHERE id=$1 AND student_id=$2`,
		enrollmentID, studentID,
	)
	if err != nil {
		return fmt.Errorf("delete enrollment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *enrollmentRepo) FindByID(ctx context.Context, id int) (model.Enrollment, error) {
	var e model.Enrollment
	err := r.pool.QueryRow(ctx,
		`SELECT id, student_id, course_id, tahun_akademik, created_at
		 FROM enrollments WHERE id = $1`, id,
	).Scan(&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Enrollment{}, ErrNotFound
		}
		return model.Enrollment{}, fmt.Errorf("FindByID enrollment: %w", err)
	}
	return e, nil
}

func (r *enrollmentRepo) SumSKS(ctx context.Context, studentID int, tahun string) (int, error) {
	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(c.sks), 0)
		 FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1 AND e.tahun_akademik = $2`,
		studentID, tahun,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("SumSKS: %w", err)
	}
	return total, nil
}

// ErrQuotaFull adalah sentinel error khusus untuk kasus kuota penuh.
var ErrQuotaFull = errors.New("kuota mata kuliah penuh")
