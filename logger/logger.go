package logger

import (
	"log/slog"
	"os"
)

// Logger defines the core logging methods and context attachment
// so callers can log at various levels and enrich entries with fields.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
}

// LogLevel is an alias of slog.Level, allowing method extensions.
type LogLevel slog.Level

var defaultLogger = NewSlogger(os.Stdout, true, slog.LevelError)

// Default returns the package-level slogger configured to write human-readable
// text logs to stdout with source information enabled at error level.
func Default() *slogger { return defaultLogger }
