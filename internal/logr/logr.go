package logr

import (
	"io"
	"log/slog"
	"os"
)

func LogDisable() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(nil, nil)))
}

func LogInfo() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)
	return logger
}

func LogDebug() *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, opts))
	slog.SetDefault(logger)
	return logger
}

func LogRemote(w io.Writer) *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	logger := slog.New(slog.NewJSONHandler(w, opts))
	slog.SetDefault(logger)
	return logger
}
