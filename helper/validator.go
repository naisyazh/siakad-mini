package helper

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

// ValidateStruct memvalidasi struct berdasarkan tag `validate`.
// Jika ada error, dikembalikan sebagai map[field]pesan untuk ErrValidation.
func ValidateStruct(s any) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	fields := make(map[string]string)
	for _, e := range err.(validator.ValidationErrors) {
		field := strings.ToLower(e.Field())
		fields[field] = humanizeError(e)
	}
	return fields
}

// humanizeError mengubah pesan error validator jadi bahasa Indonesia yang lebih ramah.
func humanizeError(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "Field ini wajib diisi"
	case "email":
		return "Format email tidak valid"
	case "min":
		if e.Type().Kind().String() == "string" {
			return "Minimal " + e.Param() + " karakter"
		}
		return "Nilai minimal " + e.Param()
	case "max":
		if e.Type().Kind().String() == "string" {
			return "Maksimal " + e.Param() + " karakter"
		}
		return "Nilai maksimal " + e.Param()
	case "len":
		return "Harus tepat " + e.Param() + " karakter"
	case "numeric":
		return "Harus berupa angka"
	case "alphanum":
		return "Hanya boleh huruf dan angka"
	default:
		return "Nilai tidak valid"
	}
}

// ── Helper untuk membaca query string ─────────────────────────

// QueryInt membaca query string sebagai int, dengan nilai default jika kosong/invalid.
func QueryInt(c *fiber.Ctx, key string, def int) int {
	val := c.Query(key)
	if val == "" {
		return def
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return n
}

// QueryBool membaca query string sebagai bool.
// Mengembalikan (value, true) jika ada, (false, false) jika tidak ada.
func QueryBool(c *fiber.Ctx, key string) (bool, bool) {
	val := c.Query(key)
	if val == "" {
		return false, false
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return false, false
	}
	return b, true
}

// ParamInt membaca parameter path :id sebagai int positif.
// Mengembalikan (0, false) jika tidak valid.
func ParamInt(c *fiber.Ctx, name string) (int, bool) {
	id, err := strconv.Atoi(c.Params(name))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

// tahunAkademikRegex memvalidasi format tahun akademik: "2026/2027-Ganjil" atau "2026/2027-Genap"
var tahunAkademikRegex = regexp.MustCompile(`^\d{4}/\d{4}-(Ganjil|Genap)$`)

// ValidateTahunAkademik memastikan format tahun akademik sesuai ketentuan.
func ValidateTahunAkademik(s string) bool {
	return tahunAkademikRegex.MatchString(s)
}

// ValidateAngkatan memastikan angkatan tidak melebihi tahun berjalan.
func ValidateAngkatan(angkatan int) bool {
	currentYear := time.Now().Year()
	return angkatan >= 2000 && angkatan <= currentYear
}

// TrimRequest adalah helper untuk bersihkan whitespace dari string field.
func TrimString(s string) string {
	return strings.TrimSpace(s)
}
