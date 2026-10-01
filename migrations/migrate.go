//go:build ignore
// +build ignore

// Jalankan dengan: go run migrations/migrate.go
// Script ini membuat database 'siakad', menjalankan semua migration SQL,
// lalu mengisi seeder dengan hash bcrypt yang benar.

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Muat .env
	if err := godotenv.Load(); err != nil {
		log.Println("tidak ada .env, pakai environment sistem")
	}

	host     := getEnv("DB_HOST", "127.0.0.1")
	port     := getEnv("DB_PORT", "5432")
	user     := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "")
	dbname   := getEnv("DB_NAME", "siakad")
	sslmode  := getEnv("DB_SSLMODE", "disable")

	// 1. Buat database jika belum ada (connect ke 'postgres' dulu)
	adminDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		host, port, user, password, sslmode)
	adminConn, err := pgx.Connect(context.Background(), adminDSN)
	if err != nil {
		log.Fatalf("gagal connect ke postgres: %v", err)
	}
	var exists bool
	_ = adminConn.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)", dbname).Scan(&exists)
	if !exists {
		_, err = adminConn.Exec(context.Background(),
			fmt.Sprintf("CREATE DATABASE %s", dbname))
		if err != nil {
			log.Fatalf("gagal buat database: %v", err)
		}
		fmt.Printf("database '%s' berhasil dibuat\n", dbname)
	} else {
		fmt.Printf("database '%s' sudah ada, skip\n", dbname)
	}
	adminConn.Close(context.Background())

	// 2. Connect ke database siakad
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatalf("gagal connect ke %s: %v", dbname, err)
	}
	defer conn.Close(context.Background())

	// 3. Jalankan migration SQL berurutan
	migrations := []string{
		"migrations/001_users.sql",
		"migrations/002_students.sql",
		"migrations/003_courses.sql",
		"migrations/004_enrollments.sql",
	}
	for _, f := range migrations {
		sql, err := os.ReadFile(f)
		if err != nil {
			log.Fatalf("gagal baca %s: %v", f, err)
		}
		if _, err := conn.Exec(context.Background(), string(sql)); err != nil {
			log.Fatalf("gagal jalankan %s: %v", f, err)
		}
		fmt.Printf("✓ %s\n", f)
	}

	// 4. Bersihkan data lama
	_, _ = conn.Exec(context.Background(),
		"TRUNCATE enrollments, students, users, courses RESTART IDENTITY CASCADE")

	// 5. Insert 10 mata kuliah
	courses := []struct{ kode, nama string; sks, sem, kuota int }{
		{"IF101", "Algoritma dan Pemrograman",    3, 1, 40},
		{"IF102", "Matematika Diskrit",           3, 1, 35},
		{"IF201", "Struktur Data",                3, 2, 35},
		{"IF202", "Basis Data",                   3, 2, 40},
		{"IF301", "Pemrograman Web",              3, 3, 40},
		{"IF302", "Rekayasa Perangkat Lunak",     3, 3, 30},
		{"IF401", "Jaringan Komputer",            3, 4, 30},
		{"IF402", "Pemrograman Mobile",           3, 4, 25},
		{"IF501", "Kecerdasan Buatan",            3, 5, 25},
		{"IF502", "Keamanan Sistem Informasi",    3, 5, 20},
	}
	for _, c := range courses {
		_, err := conn.Exec(context.Background(),
			"INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota) VALUES ($1,$2,$3,$4,$5)",
			c.kode, c.nama, c.sks, c.sem, c.kuota)
		if err != nil {
			log.Fatalf("gagal insert course %s: %v", c.kode, err)
		}
	}
	fmt.Println("✓ 10 mata kuliah berhasil dimasukkan")

	// 6. Insert 1 admin (password: admin123)
	adminHash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	var adminID int
	err = conn.QueryRow(context.Background(),
		"INSERT INTO users (email, password, role) VALUES ($1,$2,$3) RETURNING id",
		"admin@siakad.ac.id", string(adminHash), "admin").Scan(&adminID)
	if err != nil {
		log.Fatalf("gagal insert admin: %v", err)
	}
	fmt.Printf("✓ Admin dibuat (id=%d, email=admin@siakad.ac.id, password=admin123)\n", adminID)

	// 7. Insert 20 mahasiswa (password awal = NIM)
	mahasiswaData := []struct{ email, nim, nama, prodi string; angkatan int; ipk float64 }{
		{"mahasiswa01@siakad.ac.id", "434241001", "Rina Putri Andari",    "Sistem Informasi",      2022, 3.75},
		{"mahasiswa02@siakad.ac.id", "434241002", "Budi Santoso",         "Teknik Informatika",    2022, 3.20},
		{"mahasiswa03@siakad.ac.id", "434241003", "Citra Dewi Lestari",   "Sistem Informasi",      2022, 2.85},
		{"mahasiswa04@siakad.ac.id", "434241004", "Dimas Arya Pratama",   "Teknik Informatika",    2022, 3.50},
		{"mahasiswa05@siakad.ac.id", "434241005", "Eka Fitriana",         "Manajemen Informatika", 2023, 3.10},
		{"mahasiswa06@siakad.ac.id", "434241006", "Fajar Nugroho",        "Teknik Informatika",    2023, 2.40},
		{"mahasiswa07@siakad.ac.id", "434241007", "Gita Puspita Sari",    "Sistem Informasi",      2023, 3.90},
		{"mahasiswa08@siakad.ac.id", "434241008", "Hendra Wijaya",        "Manajemen Informatika", 2023, 2.60},
		{"mahasiswa09@siakad.ac.id", "434241009", "Indah Permatasari",    "Teknik Informatika",    2024, 3.45},
		{"mahasiswa10@siakad.ac.id", "434241010", "Joko Susilo",          "Sistem Informasi",      2024, 3.00},
		{"mahasiswa11@siakad.ac.id", "434241011", "Kartika Wulandari",    "Manajemen Informatika", 2024, 3.80},
		{"mahasiswa12@siakad.ac.id", "434241012", "Lukman Hakim",         "Teknik Informatika",    2024, 2.95},
		{"mahasiswa13@siakad.ac.id", "434241013", "Maya Anggraini",       "Sistem Informasi",      2024, 3.15},
		{"mahasiswa14@siakad.ac.id", "434241014", "Nanda Saputra",        "Teknik Informatika",    2022, 2.70},
		{"mahasiswa15@siakad.ac.id", "434241015", "Olivia Rachma",        "Manajemen Informatika", 2022, 3.60},
		{"mahasiswa16@siakad.ac.id", "434241016", "Putra Ramadhan",       "Teknik Informatika",    2023, 3.35},
		{"mahasiswa17@siakad.ac.id", "434241017", "Qori Amelia",          "Sistem Informasi",      2023, 2.55},
		{"mahasiswa18@siakad.ac.id", "434241018", "Rizky Firmansyah",     "Teknik Informatika",    2024, 3.70},
		{"mahasiswa19@siakad.ac.id", "434241019", "Sari Wahyuni",         "Manajemen Informatika", 2024, 3.05},
		{"mahasiswa20@siakad.ac.id", "434241020", "Taufik Hidayat",       "Sistem Informasi",      2022, 2.45},
	}

	for _, m := range mahasiswaData {
		// Password awal = NIM mahasiswa (sesuai ketentuan soal)
		nimHash, _ := bcrypt.GenerateFromPassword([]byte(m.nim), bcrypt.DefaultCost)

		var userID int
		err := conn.QueryRow(context.Background(),
			"INSERT INTO users (email, password, role) VALUES ($1,$2,'mahasiswa') RETURNING id",
			m.email, string(nimHash)).Scan(&userID)
		if err != nil {
			log.Fatalf("gagal insert user %s: %v", m.email, err)
		}

		_, err = conn.Exec(context.Background(),
			"INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES ($1,$2,$3,$4,$5,$6)",
			userID, m.nim, m.nama, m.prodi, m.angkatan, m.ipk)
		if err != nil {
			log.Fatalf("gagal insert student %s: %v", m.nim, err)
		}
	}
	fmt.Println("✓ 20 mahasiswa berhasil dimasukkan (password = NIM masing-masing)")
	fmt.Println("\nSeeder selesai! Database siap dipakai.")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
