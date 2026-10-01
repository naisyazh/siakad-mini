package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"siakad/app/model"
)

// Sentinel errors — lapisan atas cukup mengenal dua ini,
// tidak perlu tahu detail error PostgreSQL.
var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data sudah ada")
)

// UserRepository adalah kontrak (interface) untuk operasi tabel users.
// Service hanya kenal interface ini, bukan struct konkretnya.
// Manfaat: mudah diganti (misal untuk testing pakai mock).
type UserRepository interface {
	FindByID(ctx context.Context, id int) (model.User, error)
	FindByEmail(ctx context.Context, email string) (model.User, error)
	Create(ctx context.Context, email, hashedPassword, role string) (model.User, error)
}

type userRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepo{pool: pool}
}

func (r *userRepo) FindByID(ctx context.Context, id int) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password, role, created_at FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("FindByID user: %w", err)
	}
	return u, nil
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (model.User, error) {
	var u model.User
	// JOIN ke students untuk cek soft delete.
	// Kalau role mahasiswa tapi students.deleted_at terisi → tidak bisa login (401).
	// Kalau role admin → tidak ada baris di students, LEFT JOIN tetap jalan.
	err := r.pool.QueryRow(ctx,
		`SELECT u.id, u.email, u.password, u.role, u.created_at
		 FROM users u
		 LEFT JOIN students s ON s.user_id = u.id
		 WHERE LOWER(u.email) = LOWER($1)
		 AND (u.role = 'admin' OR s.deleted_at IS NULL)`, email,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("FindByEmail: %w", err)
	}
	return u, nil
}

func (r *userRepo) Create(ctx context.Context, email, hashedPassword, role string) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (email, password, role)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, password, role, created_at`,
		email, hashedPassword, role,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return model.User{}, ErrDuplicate
		}
		return model.User{}, fmt.Errorf("Create user: %w", err)
	}
	return u, nil
}

// isUniqueViolation memeriksa apakah error berasal dari pelanggaran UNIQUE constraint.
// Kode 23505 adalah kode resmi PostgreSQL untuk unique violation.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
