package helper

import "github.com/gofiber/fiber/v2"

// AppError adalah satu-satunya bentuk error yang dikenal aplikasi ini.
// Handler tidak pernah menulis response error sendiri — mereka cukup
// return salah satu fungsi di bawah, lalu ErrorHandler di config/app.go
// yang mengubahnya jadi JSON.
type AppError struct {
	Status  int               // HTTP status code
	Message string            // pesan untuk client
	Fields  map[string]string // detail per-field, khusus validasi (422)
}

func (e *AppError) Error() string { return e.Message }

// ── Konstruktor error berdasarkan status code ─────────────────

func ErrBadRequest(msg string) *AppError {
	return &AppError{Status: fiber.StatusBadRequest, Message: msg}
}

func ErrUnauthorized(msg string) *AppError {
	return &AppError{Status: fiber.StatusUnauthorized, Message: msg}
}

func ErrForbidden(msg string) *AppError {
	return &AppError{Status: fiber.StatusForbidden, Message: msg}
}

func NotFound(msg string) *AppError {
	return &AppError{Status: fiber.StatusNotFound, Message: msg}
}

func ErrConflict(msg string) *AppError {
	return &AppError{Status: fiber.StatusConflict, Message: msg}
}

func ErrValidation(fields map[string]string) *AppError {
	return &AppError{
		Status:  fiber.StatusUnprocessableEntity,
		Message: "Validasi gagal",
		Fields:  fields,
	}
}

func ErrTooManyRequests(msg string) *AppError {
	return &AppError{Status: fiber.StatusTooManyRequests, Message: msg}
}

// ErrInternal untuk error server yang tidak terduga.
// Pesan yang dikirim ke client selalu generik — detail ada di log server.
func ErrInternal(cause error) *AppError {
	return &AppError{
		Status:  fiber.StatusInternalServerError,
		Message: "Terjadi kesalahan pada server",
	}
}
