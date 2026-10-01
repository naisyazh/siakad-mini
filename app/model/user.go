package model

import "time"

// User adalah representasi baris di tabel users.
// Field Password tidak pernah dikirim ke client (json:"-").
type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // JANGAN pernah kirim ini ke client
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// ── Request & Response ────────────────────────────────────────

// LoginRequest adalah body yang dikirim saat POST /auth/login
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginResponse adalah isi dari field "data" saat login berhasil
type LoginResponse struct {
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type"`
	ExpiresIn   int      `json:"expires_in"` // dalam detik
	User        UserInfo `json:"user"`
}

// UserInfo adalah subset data user yang aman dikirim ke client
type UserInfo struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// MeResponse adalah respons GET /auth/me
// Jika role mahasiswa, StudentData akan terisi. Jika admin, nil.
type MeResponse struct {
	ID          int          `json:"id"`
	Email       string       `json:"email"`
	Role        string       `json:"role"`
	StudentData *StudentInfo `json:"student,omitempty"`
}

// StudentInfo adalah subset data student untuk response /auth/me
type StudentInfo struct {
	NIM      string `json:"nim"`
	Nama     string `json:"nama"`
	Prodi    string `json:"prodi"`
	Angkatan int    `json:"angkatan"`
}
