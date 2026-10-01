package helper

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"siakad/app/model"
)

// Key untuk menyimpan data di Fiber locals
const (
	LocalsUser = "current_user" // *model.User
)

// SetCurrentUser menyimpan user yang sedang login ke context request.
// Dipanggil oleh middleware RequireAuth setelah token diverifikasi.
func SetCurrentUser(c *fiber.Ctx, user *model.User) {
	c.Locals(LocalsUser, user)
}

// CurrentUser mengambil user yang sedang login dari context.
// Mengembalikan nil jika tidak ada (endpoint publik atau middleware belum jalan).
func CurrentUser(c *fiber.Ctx) *model.User {
	user := c.Locals(LocalsUser)
	if user == nil {
		return nil
	}
	u, _ := user.(*model.User)
	return u
}

// ReqCtx membuat context dengan timeout 5 detik untuk operasi database.
// Setiap handler wajib pakai ini agar request tidak menggantung selamanya
// kalau database lambat.
func ReqCtx(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}
