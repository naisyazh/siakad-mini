package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"siakad/app/model"
)

type StudentRepository interface {
	FindAll(ctx context.Context, q model.ListStudentQuery) ([]model.Student, int, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	FindByUserID(ctx context.Context, userID int) (model.Student, error)
	FindDetailByID(ctx context.Context, id int) (model.StudentDetail, error)
	Create(ctx context.Context, tx pgx.Tx, s model.Student) (model.Student, error)
	Update(ctx context.Context, s model.Student) (model.Student, error)
	SoftDelete(ctx context.Context, id int) error
}

type studentRepo struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentRepo{pool: pool}
}

func (r *studentRepo) FindAll(ctx context.Context, q model.ListStudentQuery) ([]model.Student, int, error) {
	// Bangun WHERE clause secara dinamis berdasarkan filter yang diberikan
	where := []string{"deleted_at IS NULL"} // soft delete: selalu kecualikan yang sudah dihapus
	args := []any{}

	if q.Prodi != "" {
		args = append(args, q.Prodi)
		where = append(where, fmt.Sprintf("prodi = $%d", len(args)))
	}
	if q.Angkatan > 0 {
		args = append(args, q.Angkatan)
		where = append(where, fmt.Sprintf("angkatan = $%d", len(args)))
	}
	if q.Search != "" {
		args = append(args, "%"+strings.ToLower(q.Search)+"%")
		where = append(where, fmt.Sprintf(
			"(LOWER(nim) LIKE $%d OR LOWER(nama) LIKE $%d)", len(args), len(args),
		))
	}

	whereSQL := "WHERE " + strings.Join(where, " AND ")

	// Hitung total dulu (untuk meta pagination)
	var total int
	err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM students "+whereSQL, args...,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count students: %w", err)
	}

	// Tentukan urutan
	orderSQL := "ORDER BY id ASC"
	switch q.Sort {
	case "nama":
		orderSQL = "ORDER BY nama ASC"
	case "-ipk_terakhir":
		orderSQL = "ORDER BY ipk_terakhir DESC"
	}

	// Hitung offset untuk pagination
	offset := (q.Page - 1) * q.PerPage
	args = append(args, q.PerPage, offset)

	rows, err := r.pool.Query(ctx,
		fmt.Sprintf(
			`SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at
			 FROM students %s %s
			 LIMIT $%d OFFSET $%d`,
			whereSQL, orderSQL, len(args)-1, len(args),
		),
		args...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("query students: %w", err)
	}
	defer rows.Close()

	result := []model.Student{}
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi,
			&s.Angkatan, &s.IPKTerakhir, &s.DeletedAt); err != nil {
			return nil, 0, fmt.Errorf("scan student: %w", err)
		}
		result = append(result, s)
	}

	return result, total, rows.Err()
}

func (r *studentRepo) FindByID(ctx context.Context, id int) (model.Student, error) {
	var s model.Student
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at
		 FROM students WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("FindByID student: %w", err)
	}
	return s, nil
}

func (r *studentRepo) FindByUserID(ctx context.Context, userID int) (model.Student, error) {
	var s model.Student
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at
		 FROM students WHERE user_id = $1 AND deleted_at IS NULL`, userID,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("FindByUserID student: %w", err)
	}
	return s, nil
}

// FindDetailByID mengambil data mahasiswa lengkap: profil + daftar KRS + total SKS.
// Satu fungsi ini menghasilkan semua yang dibutuhkan endpoint GET /students/{id}.
func (r *studentRepo) FindDetailByID(ctx context.Context, id int) (model.StudentDetail, error) {
	var d model.StudentDetail

	// Query 1: data dasar mahasiswa
	err := r.pool.QueryRow(ctx,
		`SELECT id, nim, nama, prodi, angkatan, ipk_terakhir
		 FROM students WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&d.ID, &d.NIM, &d.Nama, &d.Prodi, &d.Angkatan, &d.IPKTerakhir)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.StudentDetail{}, ErrNotFound
		}
		return model.StudentDetail{}, fmt.Errorf("FindDetailByID student: %w", err)
	}

	// Hitung batas SKS berdasarkan IPK (business rule)
	d.BatasSKS = model.BatasSKS(d.IPKTerakhir)

	// Query 2: daftar KRS beserta detail matkul
	rows, err := r.pool.Query(ctx,
		`SELECT e.id, e.course_id, c.kode_mk, c.nama_mk, c.sks, c.semester, e.tahun_akademik
		 FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1
		 ORDER BY e.created_at DESC`, id,
	)
	if err != nil {
		return model.StudentDetail{}, fmt.Errorf("query enrollments: %w", err)
	}
	defer rows.Close()

	d.Courses = []model.EnrolledCourse{}
	for rows.Next() {
		var ec model.EnrolledCourse
		if err := rows.Scan(&ec.EnrollmentID, &ec.CourseID, &ec.KodeMK,
			&ec.NamaMK, &ec.SKS, &ec.Semester, &ec.TahunAkademik); err != nil {
			return model.StudentDetail{}, fmt.Errorf("scan enrolled course: %w", err)
		}
		d.TotalSKS += ec.SKS // akumulasi total SKS
		d.Courses = append(d.Courses, ec)
	}

	return d, rows.Err()
}

func (r *studentRepo) Create(ctx context.Context, tx pgx.Tx, s model.Student) (model.Student, error) {
	// Create dijalankan dalam transaction (tx) karena bersamaan dengan Create user
	err := tx.QueryRow(ctx,
		`INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at`,
		s.UserID, s.NIM, s.Nama, s.Prodi, s.Angkatan, s.IPKTerakhir,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("Create student: %w", err)
	}
	return s, nil
}

func (r *studentRepo) Update(ctx context.Context, s model.Student) (model.Student, error) {
	// NIM tidak boleh diubah (sesuai ketentuan soal), hanya field lain yang diupdate
	err := r.pool.QueryRow(ctx,
		`UPDATE students SET nama=$1, prodi=$2, angkatan=$3, ipk_terakhir=$4
		 WHERE id=$5 AND deleted_at IS NULL
		 RETURNING id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at`,
		s.Nama, s.Prodi, s.Angkatan, s.IPKTerakhir, s.ID,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("Update student: %w", err)
	}
	return s, nil
}

func (r *studentRepo) SoftDelete(ctx context.Context, id int) error {
	// Soft delete: isi deleted_at dengan waktu sekarang, bukan hapus baris
	tag, err := r.pool.Exec(ctx,
		`UPDATE students SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		return fmt.Errorf("SoftDelete student: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
