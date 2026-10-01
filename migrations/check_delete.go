//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "127.0.0.1"), getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"), getEnv("DB_PASSWORD", ""),
		getEnv("DB_NAME", "siakad"), getEnv("DB_SSLMODE", "disable"))

	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil { log.Fatal(err) }
	defer conn.Close(context.Background())

	// Cek status mahasiswa20
	var id int
	var nim, nama string
	var deletedAt interface{}
	conn.QueryRow(context.Background(),
		`SELECT s.id, s.nim, s.nama, s.deleted_at 
		 FROM students s JOIN users u ON u.id = s.user_id 
		 WHERE u.email = 'mahasiswa20@siakad.ac.id'`,
	).Scan(&id, &nim, &nama, &deletedAt)
	fmt.Printf("Student id=%d nim=%s nama=%s deleted_at=%v\n", id, nim, nama, deletedAt)

	// Kalau deleted_at masih NULL, soft delete sekarang
	if deletedAt == nil {
		_, err = conn.Exec(context.Background(),
			`UPDATE students SET deleted_at = NOW() WHERE id = $1`, id)
		if err != nil { log.Fatal(err) }
		fmt.Println("Soft delete berhasil dilakukan")
	} else {
		fmt.Println("Sudah di-soft delete sebelumnya")
	}
}

func getEnv(k, fb string) string {
	if v := os.Getenv(k); v != "" { return v }
	return fb
}
