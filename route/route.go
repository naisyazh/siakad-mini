package route

import (
	"github.com/gofiber/fiber/v2"
	"siakad/app/service"
	"siakad/helper"
	"siakad/middleware"
)

// Register mendaftarkan semua 10 endpoint ke aplikasi Fiber.
// Setiap endpoint dikelompokkan sesuai resource-nya.
func Register(
	app *fiber.App,
	jwt *helper.JWTManager,
	authSvc *service.AuthService,
	studentSvc *service.StudentService,
	courseSvc *service.CourseService,
	enrollmentSvc *service.EnrollmentService,
) {
	api := app.Group("/api/v1")

	// ── 1-2: AUTH ────────────────────────────────────────────────
	// POST /api/v1/auth/login  → publik, ada rate limiter
	// GET  /api/v1/auth/me     → semua role yang sudah login
	auth := api.Group("/auth")
	auth.Post("/login", middleware.LoginRateLimiter(), authSvc.Login)
	auth.Get("/me", middleware.RequireAuth(jwt), authSvc.Me)

	// ── 3-7: STUDENTS ────────────────────────────────────────────
	// GET    /students       → admin only
	// POST   /students       → admin only
	// GET    /students/:id   → admin + mahasiswa (data sendiri)
	// PUT    /students/:id   → admin only
	// DELETE /students/:id   → admin only (soft delete)
	students := api.Group("/students", middleware.RequireAuth(jwt))
	students.Get("/", middleware.RequireRole("admin"), studentSvc.List)
	students.Post("/", middleware.RequireRole("admin"), studentSvc.Create)
	students.Get("/:id", studentSvc.Detail)   // ownership dicek di dalam service
	students.Put("/:id", middleware.RequireRole("admin"), studentSvc.Update)
	students.Delete("/:id", middleware.RequireRole("admin"), studentSvc.Delete)

	// ── 8: COURSES ───────────────────────────────────────────────
	// GET /courses → semua role yang sudah login
	api.Get("/courses", middleware.RequireAuth(jwt), courseSvc.List)

	// ── 9-10: ENROLLMENTS ────────────────────────────────────────
	// POST   /enrollments      → mahasiswa only
	// DELETE /enrollments/:id  → mahasiswa only (milik sendiri)
	enrollments := api.Group("/enrollments",
		middleware.RequireAuth(jwt),
		middleware.RequireRole("mahasiswa"),
	)
	enrollments.Post("/", enrollmentSvc.Enroll)
	enrollments.Delete("/:id", enrollmentSvc.CancelEnroll)
}
