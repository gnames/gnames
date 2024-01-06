package logr

import (
	"io"
	"log/slog"
	"os"
)

// LogDisable disables logging.
func LogDisable() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(io.Discard, nil)))
}

// LogDev returns a logger that writes to stderr, provides information about
// file and method used, and has debug level.
func LogDev() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, opts))
	slog.SetDefault(logger.With(
		slog.String("gnApp", "gnames"),
	))
	return logger
}

// LogInfo returns a logger that writes to stderr, and has info level.
func LogInfo() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)
	return logger
}

// LogDebug returns a logger that writes to stderr, and has debug level.
func LogDebug() *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, opts))
	slog.SetDefault(logger)
	return logger
}

// LogRemote returns a logger that writes to a given io.Writer.
func LogRemote(w io.Writer) *slog.Logger {
	opts := &slog.HandlerOptions{}
	logger := slog.New(slog.NewJSONHandler(w, opts))
	slog.SetDefault(logger)
	return logger
}
