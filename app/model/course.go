package model

// Course adalah representasi baris di tabel courses.
type Course struct {
	ID       int    `json:"id"`
	KodeMK   string `json:"kode_mk"`
	NamaMK   string `json:"nama_mk"`
	SKS      int    `json:"sks"`
	Semester int    `json:"semester"`
	Kuota    int    `json:"kuota"`
}

// CourseWithQuota adalah Course + info kuota real-time dari enrollments.
// Terisi dan SisaKuota dihitung saat query, bukan disimpan di database.
// Kenapa tidak disimpan? Agar tidak ada risiko data tidak sinkron.
type CourseWithQuota struct {
	ID        int    `json:"id"`
	KodeMK    string `json:"kode_mk"`
	NamaMK    string `json:"nama_mk"`
	SKS       int    `json:"sks"`
	Semester  int    `json:"semester"`
	Kuota     int    `json:"kuota"`
	Terisi    int    `json:"terisi"`     // dihitung dari COUNT enrollments
	SisaKuota int    `json:"sisa_kuota"` // = kuota - terisi
}

// ListCourseQuery adalah parameter query string untuk GET /courses
type ListCourseQuery struct {
	Semester  int    // filter by semester
	Search    string // cari kode_mk atau nama_mk
	Available bool   // true = hanya yang sisa_kuota > 0
}
