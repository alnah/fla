package logger

import (
	"io"
	"log/slog"
)

// slogger wraps the standard slog.Logger to provide a production-ready
// implementation that writes human-readable text to stdout.
type slogger struct{ *slog.Logger }

// NewSlogger returns a Logger that can be configure for any kind of use.
func NewSlogger(w io.Writer, source bool, level slog.Level) *slogger {
	handler := slog.NewTextHandler(w, &slog.HandlerOptions{
		AddSource: source,
		Level:     level,
	})
	return &slogger{slog.New(handler)}
}

// Debug logs a message at debug level for detailed tracing.
func (l *slogger) Debug(msg string, args ...any) {
	l.Logger.Debug(msg, args...)
}

// Info logs a message at info level for general operational events.
func (l *slogger) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

// Warn logs a message at warn level to highlight potential issues.
func (l *slogger) Warn(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

// Error logs a message at error level for failures requiring attention.
func (l *slogger) Error(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}

// With attaches structured context fields and returns a new Logger
// so subsequent calls include those fields automatically.
func (l *slogger) With(args ...any) Logger {
	return &slogger{l.Logger.With(args...)}
}
