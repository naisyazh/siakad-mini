package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv memuat variabel dari berkas .env ke environment proses.
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("peringatan: berkas .env tidak ditemukan, memakai environment sistem")
	}
}

// GetEnv mengambil nilai environment. Bila kosong, kembalikan nilai bawaan.
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
