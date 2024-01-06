package testhelpr

import (
	"io"
	"log/slog"
	"os"
)

func Setup() {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	slog.SetDefault(logger)
}

func Teardown() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)
	// Your teardown code here
}
