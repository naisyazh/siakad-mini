package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"siakad/helper"
)

// Register memasang semua middleware global ke aplikasi Fiber.
// Dipanggil satu kali dari config/app.go sebelum route didaftarkan.
func Register(app *fiber.App) {
	// Recover mencegah server crash kalau ada panic yang tidak tertangkap
	app.Use(recover.New(recover.Config{
		EnableStackTrace: false, // jangan bocorkan stack trace ke client
	}))

	// Logger mencatat setiap request: method, path, status, waktu
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${method} ${path} → ${status} (${latency})\n",
	}))
}

// ── Rate Limiter untuk Login ───────────────────────────────────
//
// Soal minta: 429 jika gagal login lebih dari 5 kali per menit.
// Kita implementasi dengan in-memory counter per IP.
// Untuk production, gunakan Redis — tapi untuk UTS ini sudah cukup.

type loginCounter struct {
	count     int
	resetAt   time.Time
}

var (
	loginMu       sync.Mutex
	loginCounters = make(map[string]*loginCounter)
)

// LoginRateLimiter membatasi percobaan login maksimal 5 kali per menit per IP.
// Jika melebihi → 429 Too Many Requests.
//
// Cara kerja:
// - Setiap IP punya counter dan waktu reset
// - Counter di-increment setiap request masuk ke endpoint login
// - Setelah 1 menit, counter direset otomatis
// - Kalau counter > 5 → tolak dengan 429
func LoginRateLimiter() fiber.Handler {
	const maxAttempts = 5
	const window = time.Minute

	return func(c *fiber.Ctx) error {
		ip := c.IP()

		loginMu.Lock()
		entry, exists := loginCounters[ip]
		now := time.Now()

		if !exists || now.After(entry.resetAt) {
			// IP baru atau window sudah habis → reset counter
			loginCounters[ip] = &loginCounter{
				count:   1,
				resetAt: now.Add(window),
			}
			loginMu.Unlock()
			return c.Next()
		}

		entry.count++
		count := entry.count
		loginMu.Unlock()

		if count > maxAttempts {
			return helper.ErrTooManyRequests(
				"Terlalu banyak percobaan login, coba lagi dalam 1 menit",
			)
		}

		return c.Next()
	}
}

// ResetLoginCounter mereset counter IP tertentu saat login berhasil.
// Dipanggil dari auth service setelah login sukses.
func ResetLoginCounter(ip string) {
	loginMu.Lock()
	delete(loginCounters, ip)
	loginMu.Unlock()
}
