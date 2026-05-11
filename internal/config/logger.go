package config

import (
	"log/slog"
	"os"
)

func (cfg *Config) Logger() *slog.Logger {
	var level slog.Level
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var format slog.Handler
	switch cfg.LogFormat {
	case "json":
		format = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	case "text":
		format = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	default:
		format = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	logger := slog.New(format)
	slog.SetDefault(logger)
	return logger
}
