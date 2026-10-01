package service

import (
	"github.com/gofiber/fiber/v2"
	"siakad/app/model"
	"siakad/app/repository"
	"siakad/helper"
)

// CourseService menangani 1 endpoint:
// - GET /api/v1/courses → Daftar matkul + sisa kuota (semua role)
type CourseService struct {
	courseRepo repository.CourseRepository
}

func NewCourseService(courseRepo repository.CourseRepository) *CourseService {
	return &CourseService{courseRepo: courseRepo}
}

// List menangani GET /api/v1/courses
// Query params: semester, search (kode_mk atau nama_mk), available=true
func (s *CourseService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	q := model.ListCourseQuery{
		Semester: helper.QueryInt(c, "semester", 0),
		Search:   c.Query("search"),
	}

	// available=true → hanya tampilkan yang masih ada sisa kuota
	if av, ok := helper.QueryBool(c, "available"); ok {
		q.Available = av
	}

	courses, err := s.courseRepo.FindAll(ctx, q)
	if err != nil {
		return helper.ErrInternal(err)
	}

	return helper.OK(c, "Daftar mata kuliah berhasil diambil", courses)
}
