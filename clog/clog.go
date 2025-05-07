package clog

import (
	"log/slog"
	"os"
)

// Logger wraps slog.Logger to offer structured logging.
type Logger struct{ *slog.Logger }

// New returns a Logger with a default TextHandler to os.Stdout.
func New() *Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})
	return NewWithHandler(handler)
}

// NewWithHandler creates a Logger using a custom slog.Handler.
// This allows tests to inject their own log collector.
func NewWithHandler(h slog.Handler) *Logger {
	return &Logger{slog.New(h)}
}
