package model

import "time"

// Enrollment adalah representasi baris di tabel enrollments (KRS).
type Enrollment struct {
	ID            int       `json:"id"`
	StudentID     int       `json:"student_id"`
	CourseID      int       `json:"course_id"`
	TahunAkademik string    `json:"tahun_akademik"`
	CreatedAt     time.Time `json:"created_at"`
}

// ── Request ────────────────────────────────────────────────────

// EnrollRequest adalah body untuk POST /enrollments
type EnrollRequest struct {
	// course_id harus ada di database
	CourseID      int    `json:"course_id"      validate:"required,min=1"`
	// Format: "2026/2027-Ganjil" atau "2026/2027-Genap"
	TahunAkademik string `json:"tahun_akademik" validate:"required"`
}
