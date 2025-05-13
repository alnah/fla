package logger

import (
	"log/slog"
	"os"
	"sync"
)

var (
	defaultLogger *Logger
	once          sync.Once
)

// Default returns a singleton Logger.
func Default() *Logger {
	once.Do(func() {
		defaultLogger = New()
	})
	return defaultLogger
}

// Logger wraps slog.Logger to offer structured, leveled logging.
type Logger struct{ *slog.Logger }

// New returns a Logger emitting human-readable text logs to stdout at debug level by default.
func New() *Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})
	return NewWithHandler(handler)
}

// NewWithHandler creates a Logger using the provided slog.Handler.
// Useful for tests (inject a buffer) or alternate formats (JSON).
func NewWithHandler(h slog.Handler) *Logger {
	return &Logger{slog.New(h)}
}
