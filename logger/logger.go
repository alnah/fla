package logger

import (
	"log/slog"
	"os"
)

// New returns a Logger emitting human-readable text logs to stdout at debug level by default.
func New() *slog.Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})
	return NewWithHandler(handler)
}

func NewTestLogger() *slog.Logger { return New() }

// NewWithHandler creates a Logger using the provided slog.Handler.
// Useful for tests (inject a buffer) or alternate formats (JSON).
func NewWithHandler(h slog.Handler) *slog.Logger {
	return slog.New(h)
}
