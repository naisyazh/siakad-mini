package config

import (
	"log/slog"
	"os"
)

// NewLogger membuat structured logger.
// Di mode production, level INFO ke atas yang dicetak.
func NewLogger() *slog.Logger {
	level := slog.LevelDebug
	if GetEnv("APP_ENV", "development") == "production" {
		level = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	return slog.New(handler)
}
