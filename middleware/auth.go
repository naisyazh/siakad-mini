package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"siakad/app/model"
	"siakad/helper"
)

// RequireAuth adalah middleware yang memvalidasi JWT token dari header Authorization.
//
// Alurnya:
// 1. Ambil header "Authorization: Bearer <token>"
// 2. Verifikasi token dengan JWTManager
// 3. Simpan data user ke Fiber locals agar handler bisa akses tanpa query DB
//
// Kalau token tidak ada atau tidak valid → 401 langsung, request tidak dilanjutkan.
func RequireAuth(jwt *helper.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Ambil header Authorization
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return helper.ErrUnauthorized("Token tidak ditemukan, silakan login terlebih dahulu")
		}

		// Format harus "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return helper.ErrUnauthorized("Format token tidak valid, gunakan: Bearer <token>")
		}

		tokenStr := parts[1]

		// Verifikasi token — cek tanda tangan dan masa berlaku
		claims, err := jwt.Verify(tokenStr)
		if err != nil {
			return helper.ErrUnauthorized("Token tidak valid atau sudah kedaluwarsa")
		}

		// Simpan user ke context agar handler bisa akses via helper.CurrentUser(c)
		// Kita tidak query DB di sini karena data sudah ada di token (stateless)
		user := &model.User{
			ID:   claims.UserID,
			Role: claims.Role,
		}
		helper.SetCurrentUser(c, user)

		return c.Next()
	}
}

// RequireRole memastikan user yang login memiliki role tertentu.
// Harus dipasang SETELAH RequireAuth karena butuh current user dari context.
//
// Contoh penggunaan:
//
//	users.Get("/", RequireAuth(jwt), RequireRole("admin"), handler)
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := helper.CurrentUser(c)
		if user == nil {
			// Seharusnya tidak terjadi kalau RequireAuth dipasang lebih dulu
			return helper.ErrUnauthorized("Silakan login terlebih dahulu")
		}

		for _, role := range roles {
			if user.Role == role {
				return c.Next()
			}
		}

		return helper.ErrForbidden("Akses ditolak, role tidak sesuai")
	}
}
