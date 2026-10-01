package service

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"siakad/app/model"
	"siakad/app/repository"
	"siakad/helper"
)

// EnrollmentService menangani 2 endpoint:
// - POST   /api/v1/enrollments      → Ambil matkul / tambah KRS (mahasiswa)
// - DELETE /api/v1/enrollments/{id} → Batalkan KRS (mahasiswa, milik sendiri)
type EnrollmentService struct {
	enrollRepo  repository.EnrollmentRepository
	studentRepo repository.StudentRepository
	courseRepo  repository.CourseRepository
}

func NewEnrollmentService(
	enrollRepo repository.EnrollmentRepository,
	studentRepo repository.StudentRepository,
	courseRepo repository.CourseRepository,
) *EnrollmentService {
	return &EnrollmentService{
		enrollRepo:  enrollRepo,
		studentRepo: studentRepo,
		courseRepo:  courseRepo,
	}
}

// Enroll menangani POST /api/v1/enrollments
//
// Business rules yang dicek (sesuai soal):
// 1. course_id harus ada
// 2. Mahasiswa tidak bisa ambil matkul yang sama dua kali pada tahun akademik yang sama (→ 409)
// 3. Kuota matkul tidak boleh penuh (→ 422)
// 4. Total SKS tidak boleh melebihi batas berdasarkan IPK (→ 422, sertakan sisa SKS)
func (s *EnrollmentService) Enroll(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	// Baca request body
	var req model.EnrollRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrBadRequest("Body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.ErrValidation(errs)
	}

	// Validasi format tahun akademik: "2026/2027-Ganjil" atau "2026/2027-Genap"
	if !helper.ValidateTahunAkademik(req.TahunAkademik) {
		return helper.ErrValidation(map[string]string{
			"tahun_akademik": "Format harus: YYYY/YYYY-Ganjil atau YYYY/YYYY-Genap",
		})
	}

	// Ambil data mahasiswa yang sedang login
	currentUser := helper.CurrentUser(c)
	student, err := s.studentRepo.FindByUserID(ctx, currentUser.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.ErrForbidden("Anda tidak terdaftar sebagai mahasiswa")
		}
		return helper.ErrInternal(err)
	}

	// Cek course_id ada di database
	course, err := s.courseRepo.FindByID(ctx, req.CourseID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.ErrValidation(map[string]string{
				"course_id": "Mata kuliah tidak ditemukan",
			})
		}
		return helper.ErrInternal(err)
	}

	// ── Business Rule #4: Cek batas SKS ───────────────────────
	// Hitung total SKS yang sudah diambil di tahun akademik ini
	totalSKS, err := s.enrollRepo.SumSKS(ctx, student.ID, req.TahunAkademik)
	if err != nil {
		return helper.ErrInternal(err)
	}

	batas := model.BatasSKS(student.IPKTerakhir)
	sisaSKS := batas - totalSKS

	// Cek apakah penambahan matkul ini akan melebihi batas
	if totalSKS+course.SKS > batas {
		return helper.ErrValidation(map[string]string{
			"sks": fmt.Sprintf(
				"Penambahan %d SKS melebihi batas. Sisa SKS yang dapat diambil: %d SKS",
				course.SKS, sisaSKS,
			),
		})
	}

	// ── Enroll dalam transaction (cek duplikasi + kuota di repo) ─
	enrollment, err := s.enrollRepo.Enroll(ctx, student.ID, req.CourseID, req.TahunAkademik)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicate):
			// Business Rule #2: sudah pernah ambil matkul ini di tahun yang sama
			return helper.ErrConflict("Anda sudah mengambil mata kuliah ini pada tahun akademik yang sama")
		case errors.Is(err, repository.ErrQuotaFull):
			// Business Rule #3: kuota penuh
			return helper.ErrValidation(map[string]string{
				"course_id": "Kuota mata kuliah sudah penuh",
			})
		case errors.Is(err, repository.ErrNotFound):
			return helper.NotFound("Mata kuliah tidak ditemukan")
		default:
			return helper.ErrInternal(err)
		}
	}

	return helper.Created(c, "Mata kuliah berhasil ditambahkan ke KRS", enrollment)
}

// CancelEnroll menangani DELETE /api/v1/enrollments/{id}
//
// Mahasiswa hanya bisa membatalkan KRS miliknya sendiri.
// Kalau enrollment milik orang lain → 403.
// Kalau tidak ditemukan → 404.
func (s *EnrollmentService) CancelEnroll(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	enrollID, ok := helper.ParamInt(c, "id")
	if !ok {
		return helper.ErrBadRequest("ID enrollment harus berupa angka positif")
	}

	// Ambil data mahasiswa yang sedang login
	currentUser := helper.CurrentUser(c)
	student, err := s.studentRepo.FindByUserID(ctx, currentUser.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.ErrForbidden("Anda tidak terdaftar sebagai mahasiswa")
		}
		return helper.ErrInternal(err)
	}

	// Cek dulu apakah enrollment ini ada
	enrollment, err := s.enrollRepo.FindByID(ctx, enrollID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("Data KRS tidak ditemukan")
		}
		return helper.ErrInternal(err)
	}

	// Cek kepemilikan: enrollment harus milik mahasiswa yang login
	if enrollment.StudentID != student.ID {
		return helper.ErrForbidden("Anda tidak dapat membatalkan KRS milik mahasiswa lain")
	}

	// Hapus enrollment — kuota otomatis kembali bertambah
	// (kuota tidak disimpan di courses, dihitung dari COUNT enrollments)
	if err := s.enrollRepo.Delete(ctx, enrollID, student.ID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("Data KRS tidak ditemukan")
		}
		return helper.ErrInternal(err)
	}

	return helper.NoContent(c)
}
