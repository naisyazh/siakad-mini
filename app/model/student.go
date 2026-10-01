package model

import "time"

// Student adalah representasi baris di tabel students.
// DeletedAt bertipe *time.Time (pointer) karena nilainya boleh NULL di database.
// Kalau bukan pointer, nil tidak bisa direpresentasikan.
type Student struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	NIM         string     `json:"nim"`
	Nama        string     `json:"nama"`
	Prodi       string     `json:"prodi"`
	Angkatan    int        `json:"angkatan"`
	IPKTerakhir float64    `json:"ipk_terakhir"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"` // nil = aktif
}

// StudentDetail adalah respons untuk GET /students/{id}
// Berisi data mahasiswa + daftar KRS + total SKS + batas SKS
type StudentDetail struct {
	ID          int          `json:"id"`
	NIM         string       `json:"nim"`
	Nama        string       `json:"nama"`
	Prodi       string       `json:"prodi"`
	Angkatan    int          `json:"angkatan"`
	IPKTerakhir float64      `json:"ipk_terakhir"`
	TotalSKS    int          `json:"total_sks"`
	BatasSKS    int          `json:"batas_sks"`
	Courses     []EnrolledCourse `json:"courses"`
}

// EnrolledCourse adalah data matkul yang sudah diambil mahasiswa (untuk detail student)
type EnrolledCourse struct {
	EnrollmentID  int    `json:"enrollment_id"`
	CourseID      int    `json:"course_id"`
	KodeMK        string `json:"kode_mk"`
	NamaMK        string `json:"nama_mk"`
	SKS           int    `json:"sks"`
	Semester      int    `json:"semester"`
	TahunAkademik string `json:"tahun_akademik"`
}

// ── Request ────────────────────────────────────────────────────

// CreateStudentRequest adalah body untuk POST /students
type CreateStudentRequest struct {
	NIM         string  `json:"nim"          validate:"required,len=12,numeric"`
	Nama        string  `json:"nama"         validate:"required,min=2,max=100"`
	Email       string  `json:"email"        validate:"required,email,max=120"`
	Prodi       string  `json:"prodi"        validate:"required,min=2,max=100"`
	Angkatan    int     `json:"angkatan"     validate:"required,min=2000"`
	IPKTerakhir float64 `json:"ipk_terakhir" validate:"omitempty,min=0,max=4"`
}

// UpdateStudentRequest adalah body untuk PUT /students/{id}
// NIM tidak boleh diubah sesuai ketentuan soal
type UpdateStudentRequest struct {
	Nama        string  `json:"nama"         validate:"required,min=2,max=100"`
	Prodi       string  `json:"prodi"        validate:"required,min=2,max=100"`
	Angkatan    int     `json:"angkatan"     validate:"required,min=2000"`
	IPKTerakhir float64 `json:"ipk_terakhir" validate:"omitempty,min=0,max=4"`
}

// ListStudentQuery adalah parameter query string untuk GET /students
type ListStudentQuery struct {
	Page     int    // default 1
	PerPage  int    // default 10, maks 50
	Prodi    string // filter by prodi
	Angkatan int    // filter by angkatan
	Search   string // cari nim atau nama
	Sort     string // "nama" atau "-ipk_terakhir"
}

// BatasSKS menghitung batas SKS berdasarkan IPK sesuai business rule
// IPK >= 3.00 -> 24 SKS
// IPK 2.50-2.99 -> 21 SKS
// IPK < 2.50 -> 18 SKS
func BatasSKS(ipk float64) int {
	switch {
	case ipk >= 3.00:
		return 24
	case ipk >= 2.50:
		return 21
	default:
		return 18
	}
}
