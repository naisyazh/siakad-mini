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

type CourseRepository interface {
	FindAll(ctx context.Context, q model.ListCourseQuery) ([]model.CourseWithQuota, error)
	FindByID(ctx context.Context, id int) (model.Course, error)
}

type courseRepo struct {
	pool *pgxpool.Pool
}

func NewCourseRepository(pool *pgxpool.Pool) CourseRepository {
	return &courseRepo{pool: pool}
}

// FindAll mengambil daftar matkul beserta info kuota real-time.
// terisi dihitung dengan COUNT dari enrollments — selalu akurat.
func (r *courseRepo) FindAll(ctx context.Context, q model.ListCourseQuery) ([]model.CourseWithQuota, error) {
	where := []string{"1=1"}
	args := []any{}

	if q.Semester > 0 {
		args = append(args, q.Semester)
		where = append(where, fmt.Sprintf("c.semester = $%d", len(args)))
	}
	if q.Search != "" {
		args = append(args, "%"+strings.ToLower(q.Search)+"%")
		where = append(where, fmt.Sprintf(
			"(LOWER(c.kode_mk) LIKE $%d OR LOWER(c.nama_mk) LIKE $%d)", len(args), len(args),
		))
	}

	whereSQL := "WHERE " + strings.Join(where, " AND ")

	// Jika available=true, tambahkan filter sisa kuota > 0
	havingSQL := ""
	if q.Available {
		havingSQL = "HAVING c.kuota - COUNT(e.id) > 0"
	}

	query := fmt.Sprintf(
		`SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota,
		        COUNT(e.id) AS terisi,
		        c.kuota - COUNT(e.id) AS sisa_kuota
		 FROM courses c
		 LEFT JOIN enrollments e ON e.course_id = c.id
		 %s
		 GROUP BY c.id
		 %s
		 ORDER BY c.semester ASC, c.kode_mk ASC`,
		whereSQL, havingSQL,
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("FindAll courses: %w", err)
	}
	defer rows.Close()

	result := []model.CourseWithQuota{}
	for rows.Next() {
		var c model.CourseWithQuota
		if err := rows.Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester,
			&c.Kuota, &c.Terisi, &c.SisaKuota); err != nil {
			return nil, fmt.Errorf("scan course: %w", err)
		}
		result = append(result, c)
	}

	return result, rows.Err()
}

func (r *courseRepo) FindByID(ctx context.Context, id int) (model.Course, error) {
	var c model.Course
	err := r.pool.QueryRow(ctx,
		`SELECT id, kode_mk, nama_mk, sks, semester, kuota FROM courses WHERE id = $1`, id,
	).Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Course{}, ErrNotFound
		}
		return model.Course{}, fmt.Errorf("FindByID course: %w", err)
	}
	return c, nil
}
