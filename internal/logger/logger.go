package logger

import (
	"log/slog"
	"os"
	"strings"

	"forsaken-mail/internal/config"

	"gopkg.in/natefinch/lumberjack.v2"
)

func Setup(cfg *config.Config) {
	var level slog.Level
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if cfg.LogFile != "" {
		w := &lumberjack.Logger{
			Filename:   cfg.LogFile,
			MaxSize:    cfg.LogMaxSizeMB,
			MaxBackups: cfg.LogMaxBackups,
		}
		handler = slog.NewJSONHandler(w, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}
