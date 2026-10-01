package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"siakad/app/repository"
	"siakad/app/service"
	"siakad/config"
	"siakad/database"
	"siakad/helper"
)

func main() {
	// ── 1. Konfigurasi ────────────────────────────────────────────
	config.LoadEnv()
	logger := config.NewLogger()

	// ── 2. Database ───────────────────────────────────────────────
	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()
	logger.Info("database terhubung")

	// ── 3. JWT Manager ────────────────────────────────────────────
	jwtManager := helper.NewJWTManager(
		config.GetEnv("JWT_SECRET", "rahasia-ganti-di-production"),
		config.GetEnv("JWT_ISSUER", "siakad-api"),
		24*time.Hour, // access token berlaku 24 jam (untuk development)
	)

	// ── 4. Repository — lapisan akses database ────────────────────
	userRepo       := repository.NewUserRepository(pool)
	studentRepo    := repository.NewStudentRepository(pool)
	courseRepo     := repository.NewCourseRepository(pool)
	enrollmentRepo := repository.NewEnrollmentRepository(pool)

	// ── 5. Service — lapisan logika bisnis + handler ──────────────
	authSvc       := service.NewAuthService(userRepo, studentRepo, jwtManager)
	studentSvc    := service.NewStudentService(studentRepo, userRepo, pool)
	courseSvc     := service.NewCourseService(courseRepo)
	enrollmentSvc := service.NewEnrollmentService(enrollmentRepo, studentRepo, courseRepo)

	// ── 6. Rakit aplikasi Fiber ───────────────────────────────────
	app := config.NewApp(logger, pool, jwtManager, authSvc, studentSvc, courseSvc, enrollmentSvc)

	port := config.GetEnv("APP_PORT", "3001")

	// Jalankan server di goroutine agar tidak memblokir graceful shutdown
	go func() {
		logger.Info("server berjalan", slog.String("port", port))
		if err := app.Listen(":" + port); err != nil {
			logger.Error("server berhenti", slog.String("error", err.Error()))
		}
	}()

	// ── 7. Graceful Shutdown ──────────────────────────────────────
	// Tunggu sinyal Ctrl+C atau SIGTERM dari sistem
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("sinyal berhenti diterima, menutup server...")

	// Beri waktu 10 detik untuk request yang sedang berjalan selesai
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error("gagal menutup server", slog.String("error", err.Error()))
	}

	logger.Info("server berhenti dengan rapi")
}
