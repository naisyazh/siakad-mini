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
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())

	// Set kuota IF502 jadi 1 untuk test kuota penuh
	_, err = conn.Exec(context.Background(),
		"UPDATE courses SET kuota = 1 WHERE kode_mk = 'IF502'")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Kuota IF502 diset ke 1")

	// Cek hasilnya
	var kuota int
	conn.QueryRow(context.Background(),
		"SELECT kuota FROM courses WHERE kode_mk = 'IF502'").Scan(&kuota)
	fmt.Printf("Kuota IF502 sekarang: %d\n", kuota)
}

func getEnv(k, fb string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fb
}
