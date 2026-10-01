package config

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad/app/service"
	"siakad/helper"
	"siakad/middleware"
	"siakad/route"
)

// NewApp merakit aplikasi Fiber lengkap: middleware global + semua route.
func NewApp(
	logger *slog.Logger,
	pool *pgxpool.Pool,
	jwt *helper.JWTManager,
	authSvc *service.AuthService,
	studentSvc *service.StudentService,
	courseSvc *service.CourseService,
	enrollmentSvc *service.EnrollmentService,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      GetEnv("APP_NAME", "SIAKAD Mini API"),
		ErrorHandler: newErrorHandler(logger),
	})

	// Middleware global: recover dari panic + request logger
	middleware.Register(app)

	// Daftarkan semua 10 endpoint
	route.Register(app, jwt, authSvc, studentSvc, courseSvc, enrollmentSvc)

	// Tangkap URL yang tidak dikenal
	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("Endpoint tidak ditemukan")
	})

	return app
}

// newErrorHandler adalah satu-satunya tempat error diubah jadi response JSON.
// Handler tidak pernah menulis response error sendiri — mereka cukup return error.
func newErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		var appErr *helper.AppError

		switch {
		case errors.As(err, &appErr):
			// Error yang sudah kita rencanakan (validasi, not found, dll)
		default:
			// Error tak terduga — cek apakah dari Fiber sendiri
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				appErr = &helper.AppError{
					Status:  fiberErr.Code,
					Message: fiberErr.Message,
				}
			} else {
				// Error server yang tidak terduga → log detail, kirim pesan generik
				logger.Error("unhandled error",
					slog.String("path", c.Path()),
					slog.String("method", c.Method()),
					slog.String("error", err.Error()),
				)
				appErr = helper.ErrInternal(err)
			}
		}

		// Semua error dikembalikan dengan format JSON yang seragam
		return c.Status(appErr.Status).JSON(fiber.Map{
			"success": false,
			"message": appErr.Message,
			"errors":  appErr.Fields,
		})
	}
}
