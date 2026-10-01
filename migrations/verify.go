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
		getEnv("DB_HOST","127.0.0.1"), getEnv("DB_PORT","5432"),
		getEnv("DB_USER","postgres"), getEnv("DB_PASSWORD",""),
		getEnv("DB_NAME","siakad"), getEnv("DB_SSLMODE","disable"))

	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil { log.Fatal(err) }
	defer conn.Close(context.Background())

	tables := []string{"users", "students", "courses", "enrollments"}
	for _, t := range tables {
		var count int
		conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM "+t).Scan(&count)
		fmt.Printf("%-15s : %d baris\n", t, count)
	}

	// Tampilkan sample mahasiswa
	fmt.Println("\nSample mahasiswa:")
	rows, _ := conn.Query(context.Background(),
		"SELECT s.nim, s.nama, s.prodi, s.ipk_terakhir, u.email FROM students s JOIN users u ON u.id=s.user_id LIMIT 3")
	defer rows.Close()
	for rows.Next() {
		var nim, nama, prodi, email string
		var ipk float64
		rows.Scan(&nim, &nama, &prodi, &ipk, &email)
		fmt.Printf("  NIM:%-12s Nama:%-22s IPK:%.2f Email:%s\n", nim, nama, ipk, email)
	}
}

func getEnv(k, fb string) string {
	if v := os.Getenv(k); v != "" { return v }
	return fb
}
