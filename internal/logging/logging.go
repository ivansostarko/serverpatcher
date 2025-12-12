package logging

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/serverpatcher/serverpatcher/internal/config"
)

func New(cfg config.LoggingConfig) (*slog.Logger, func(), error) {
	level := parseLevel(cfg.Level)

	if err := os.MkdirAll(filepath.Dir(cfg.File), 0755); err != nil {
		return nil, nil, err
	}
	f, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, err
	}

	var w io.Writer = f
	if cfg.AlsoStdout {
		w = io.MultiWriter(os.Stdout, f)
	}

	opts := &slog.HandlerOptions{
		Level: level,
		// Ensure timestamps are present even in text mode.
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{Key: slog.TimeKey, Value: slog.StringValue(a.Value.Time().Format(time.RFC3339Nano))}
			}
			return a
		},
	}

	var handler slog.Handler
	if cfg.JSON {
		handler = slog.NewJSONHandler(w, opts)
	} else {
		handler = slog.NewTextHandler(w, opts)
	}

	logger := slog.New(handler).With("app", "serverpatcher")
	closeFn := func() { _ = f.Close() }
	return logger, closeFn, nil
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
