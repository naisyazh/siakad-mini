package service

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"siakad/app/model"
	"siakad/app/repository"
	"siakad/helper"
)

// StudentService menangani 5 endpoint:
// - GET    /api/v1/students        → List (admin only)
// - POST   /api/v1/students        → Create (admin only)
// - GET    /api/v1/students/{id}   → Detail (admin + mahasiswa data sendiri)
// - PUT    /api/v1/students/{id}   → Update (admin only)
// - DELETE /api/v1/students/{id}   → Soft delete (admin only)
type StudentService struct {
	studentRepo repository.StudentRepository
	userRepo    repository.UserRepository
	pool        *pgxpool.Pool
}

func NewStudentService(
	studentRepo repository.StudentRepository,
	userRepo repository.UserRepository,
	pool *pgxpool.Pool,
) *StudentService {
	return &StudentService{studentRepo: studentRepo, userRepo: userRepo, pool: pool}
}

// List menangani GET /api/v1/students
// Hanya admin yang boleh akses (dijaga di route).
// Mendukung: pagination, filter prodi & angkatan, search nim/nama, sort.
func (s *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	// Baca semua query parameter dengan nilai default yang aman
	q := model.ListStudentQuery{
		Page:     helper.QueryInt(c, "page", 1),
		PerPage:  helper.QueryInt(c, "per_page", 10),
		Prodi:    c.Query("prodi"),
		Angkatan: helper.QueryInt(c, "angkatan", 0),
		Search:   c.Query("search"),
		Sort:     c.Query("sort"),
	}

	// Batasi per_page maksimal 50
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PerPage < 1 {
		q.PerPage = 10
	}
	if q.PerPage > 50 {
		q.PerPage = 50
	}

	students, total, err := s.studentRepo.FindAll(ctx, q)
	if err != nil {
		return helper.ErrInternal(err)
	}

	meta := model.NewMeta(q.Page, q.PerPage, total)
	return helper.OKWithMeta(c, "Data mahasiswa berhasil diambil", students, meta)
}

// Detail menangani GET /api/v1/students/{id}
// Admin bisa akses siapa saja. Mahasiswa hanya bisa akses data dirinya sendiri.
func (s *StudentService) Detail(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	id, ok := helper.ParamInt(c, "id")
	if !ok {
		return helper.ErrBadRequest("ID mahasiswa harus berupa angka positif")
	}

	currentUser := helper.CurrentUser(c)

	// Jika mahasiswa → cek apakah sedang mengakses data dirinya sendiri
	if currentUser.Role == "mahasiswa" {
		// Cari student berdasarkan user_id yang login
		myStudent, err := s.studentRepo.FindByUserID(ctx, currentUser.ID)
		if err != nil {
			return helper.ErrForbidden("Anda tidak memiliki data mahasiswa")
		}
		// Kalau id di URL bukan miliknya → tolak
		if myStudent.ID != id {
			return helper.ErrForbidden("Anda hanya dapat mengakses data milik sendiri")
		}
	}

	detail, err := s.studentRepo.FindDetailByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("Mahasiswa tidak ditemukan")
		}
		return helper.ErrInternal(err)
	}

	return helper.OK(c, "Detail mahasiswa berhasil diambil", detail)
}

// Create menangani POST /api/v1/students
// Membuat user (role mahasiswa) DAN student dalam satu transaction.
// Password awal = NIM yang di-hash bcrypt (sesuai ketentuan soal).
func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrBadRequest("Body harus berupa JSON yang valid")
	}

	// Validasi struct tag
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.ErrValidation(errs)
	}

	// Validasi tambahan yang tidak bisa dilakukan dengan tag
	if !helper.ValidateAngkatan(req.Angkatan) {
		return helper.ErrValidation(map[string]string{
			"angkatan": "Angkatan tidak valid atau melebihi tahun berjalan",
		})
	}

	// Hash NIM sebagai password awal
	hashedNIM, err := bcrypt.GenerateFromPassword([]byte(req.NIM), bcrypt.DefaultCost)
	if err != nil {
		return helper.ErrInternal(err)
	}

	// Mulai transaction — user dan student harus dibuat bersamaan.
	// Kalau salah satu gagal, keduanya di-rollback.
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return helper.ErrInternal(err)
	}
	defer tx.Rollback(ctx) // no-op kalau sudah Commit

	// Insert user baru
	var newUser model.User
	err = tx.QueryRow(ctx,
		`INSERT INTO users (email, password, role) VALUES ($1,$2,'mahasiswa')
		 RETURNING id, email, password, role, created_at`,
		req.Email, string(hashedNIM),
	).Scan(&newUser.ID, &newUser.Email, &newUser.Password, &newUser.Role, &newUser.CreatedAt)
	if err != nil {
		if isErrDuplicate(err) {
			return helper.ErrValidation(map[string]string{
				"email": "Email sudah terdaftar",
			})
		}
		return helper.ErrInternal(err)
	}

	// Insert student baru (dalam tx yang sama)
	newStudent, err := s.studentRepo.Create(ctx, tx, model.Student{
		UserID:      newUser.ID,
		NIM:         req.NIM,
		Nama:        req.Nama,
		Prodi:       req.Prodi,
		Angkatan:    req.Angkatan,
		IPKTerakhir: req.IPKTerakhir,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.ErrValidation(map[string]string{
				"nim": "NIM sudah terdaftar",
			})
		}
		return helper.ErrInternal(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return helper.ErrInternal(err)
	}

	return helper.Created(c, "Mahasiswa berhasil ditambahkan", newStudent)
}

// Update menangani PUT /api/v1/students/{id}
// NIM tidak boleh diubah. Hanya nama, prodi, angkatan, ipk_terakhir.
func (s *StudentService) Update(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	id, ok := helper.ParamInt(c, "id")
	if !ok {
		return helper.ErrBadRequest("ID mahasiswa harus berupa angka positif")
	}

	var req model.UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrBadRequest("Body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.ErrValidation(errs)
	}
	if !helper.ValidateAngkatan(req.Angkatan) {
		return helper.ErrValidation(map[string]string{
			"angkatan": "Angkatan tidak valid atau melebihi tahun berjalan",
		})
	}

	updated, err := s.studentRepo.Update(ctx, model.Student{
		ID:          id,
		Nama:        req.Nama,
		Prodi:       req.Prodi,
		Angkatan:    req.Angkatan,
		IPKTerakhir: req.IPKTerakhir,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("Mahasiswa tidak ditemukan")
		}
		return helper.ErrInternal(err)
	}

	return helper.OK(c, "Data mahasiswa berhasil diperbarui", updated)
}

// Delete menangani DELETE /api/v1/students/{id}
// Soft delete: mengisi deleted_at, tidak menghapus baris.
// Mahasiswa yang di-soft-delete tidak muncul di List dan tidak bisa login.
func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	id, ok := helper.ParamInt(c, "id")
	if !ok {
		return helper.ErrBadRequest("ID mahasiswa harus berupa angka positif")
	}

	if err := s.studentRepo.SoftDelete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("Mahasiswa tidak ditemukan")
		}
		return helper.ErrInternal(err)
	}

	return helper.NoContent(c)
}

// isErrDuplicate cek apakah error adalah unique constraint violation dari pgx (kode 23505)
func isErrDuplicate(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
