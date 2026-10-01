package service

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"siakad/app/model"
	"siakad/app/repository"
	"siakad/helper"
	"siakad/middleware"
)

// AuthService menangani 2 endpoint:
// - POST /api/v1/auth/login  → Login
// - GET  /api/v1/auth/me     → Profil user yang sedang login
type AuthService struct {
	userRepo    repository.UserRepository
	studentRepo repository.StudentRepository
	jwt         *helper.JWTManager
}

func NewAuthService(
	userRepo repository.UserRepository,
	studentRepo repository.StudentRepository,
	jwt *helper.JWTManager,
) *AuthService {
	return &AuthService{userRepo: userRepo, studentRepo: studentRepo, jwt: jwt}
}

// Login menangani POST /api/v1/auth/login
//
// Alur:
// 1. Validasi request body (email + password)
// 2. Cari user berdasarkan email
// 3. Bandingkan password dengan hash bcrypt
// 4. Generate JWT token
// 5. Reset rate limit counter karena login berhasil
func (s *AuthService) Login(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	// Baca dan validasi request body
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrBadRequest("Body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.ErrValidation(errs)
	}

	// Cari user berdasarkan email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// Sengaja pakai pesan generik — jangan bocorkan apakah email terdaftar
			return helper.ErrUnauthorized("Email atau password salah")
		}
		return helper.ErrInternal(err)
	}

	// Bandingkan password yang dikirim dengan hash di database
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return helper.ErrUnauthorized("Email atau password salah")
	}

	// Generate JWT access token
	token, err := s.jwt.Generate(user.ID, user.Role)
	if err != nil {
		return helper.ErrInternal(err)
	}

	// Reset counter rate limit untuk IP ini karena login berhasil
	middleware.ResetLoginCounter(c.IP())

	return helper.OK(c, "Login berhasil", model.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   s.jwt.TTLSeconds(),
		User: model.UserInfo{
			ID:    user.ID,
			Email: user.Email,
			Role:  user.Role,
		},
	})
}

// Me menangani GET /api/v1/auth/me
//
// Alur:
// 1. Ambil user dari context (sudah diisi oleh middleware RequireAuth)
// 2. Query ulang ke DB untuk mendapat data lengkap (email dll)
// 3. Jika role mahasiswa → sertakan data student (nim, nama, prodi, angkatan)
func (s *AuthService) Me(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	// Ambil user dari context yang sudah diisi middleware
	currentUser := helper.CurrentUser(c)
	if currentUser == nil {
		return helper.ErrUnauthorized("Silakan login terlebih dahulu")
	}

	// Query data lengkap user dari DB (token hanya simpan ID + role)
	user, err := s.userRepo.FindByID(ctx, currentUser.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.ErrUnauthorized("Akun tidak ditemukan")
		}
		return helper.ErrInternal(err)
	}

	resp := model.MeResponse{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}

	// Jika mahasiswa → tambahkan data student
	if user.Role == "mahasiswa" {
		student, err := s.studentRepo.FindByUserID(ctx, user.ID)
		if err == nil {
			resp.StudentData = &model.StudentInfo{
				NIM:      student.NIM,
				Nama:     student.Nama,
				Prodi:    student.Prodi,
				Angkatan: student.Angkatan,
			}
		}
		// Jika student tidak ditemukan, tetap lanjut (data user tetap dikembalikan)
	}

	return helper.OK(c, "Data profil berhasil diambil", resp)
}
