package database

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"siakad/config"
)

// NewPool membuat connection pool ke PostgreSQL.
// Pool di-share ke seluruh aplikasi — satu pool untuk semua request.
func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
	host     := config.GetEnv("DB_HOST", "127.0.0.1")
	port     := config.GetEnv("DB_PORT", "5432")
	user     := config.GetEnv("DB_USER", "postgres")
	password := config.GetEnv("DB_PASSWORD", "")
	dbname   := config.GetEnv("DB_NAME", "siakad")
	sslmode  := config.GetEnv("DB_SSLMODE", "disable")
	maxConns := config.GetEnv("DB_MAX_CONNS", "10")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s pool_max_conns=%s",
		host, port, user, password, dbname, sslmode, maxConns,
	)

	// Validasi maxConns bisa diparse (mencegah konfigurasi salah)
	if _, err := strconv.Atoi(maxConns); err != nil {
		return nil, fmt.Errorf("DB_MAX_CONNS bukan angka: %w", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat pool: %w", err)
	}

	// Pastikan koneksi benar-benar berhasil sebelum lanjut.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("gagal ping database: %w", err)
	}

	return pool, nil
}
