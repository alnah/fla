package logger

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
)

const (
	LevelDebug LogLevel = LogLevel(slog.LevelDebug)
	LevelInfo  LogLevel = LogLevel(slog.LevelInfo)
	LevelWarn  LogLevel = LogLevel(slog.LevelWarn)
	LevelError LogLevel = LogLevel(slog.LevelError)
)

// validLevels lists all allowed LogLevel values.
var validLevels = []LogLevel{LevelDebug, LevelInfo, LevelWarn, LevelError}

// String returns the text representation of the log level.
func (l LogLevel) String() string { return slog.Level(l).String() }

// Int returns the numeric code of the log level.
func (l LogLevel) Int() int { return int(l) }

// IsValid reports whether l is one of the predefined constants.
func (l LogLevel) IsValid() bool { return slices.Contains(validLevels, l) }

// Validate returns an error if l is not a valid LogLevel.
func (l LogLevel) Validate() error {
	if !l.IsValid() {
		return fmt.Errorf("invalid log level: %d", l)
	}
	return nil
}

// MarshalText implements encoding.TextMarshaler, delegating to slog.Level.
func (l LogLevel) MarshalText() ([]byte, error) {
	return slog.Level(l).MarshalText()
}

// UnmarshalText implements encoding.TextUnmarshaler, delegating to slog.Level.
func (l *LogLevel) UnmarshalText(text []byte) error {
	var lvl slog.Level
	if err := lvl.UnmarshalText(text); err != nil {
		return err
	}
	*l = LogLevel(lvl)
	return nil
}

// MarshalJSON ensures JSON output is the level’s name (e.g. "INFO").
func (l LogLevel) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.String())
}

// UnmarshalJSON accepts either a quoted name ("debug", case-insensitive) or a numeric value (e.g. 4).
func (l *LogLevel) UnmarshalJSON(data []byte) error {
	// try as string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		lvl, err := ParseLogLevel(s)
		if err != nil {
			return err
		}
		*l = lvl
		return nil
	}

	// try as number
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*l = LogLevel(i)
		return nil
	}

	return fmt.Errorf("invalid log level %s", strings.TrimSpace(string(data)))
}

// ParseLogLevel converts s to a LogLevel, accepting names ("DEBUG","info",…) or numeric codes.
func ParseLogLevel(s string) (LogLevel, error) {
	u := strings.ToUpper(strings.TrimSpace(s))

	switch u {
	case "DEBUG":
		return LevelDebug, nil
	case "INFO":
		return LevelInfo, nil
	case "WARN", "WARNING":
		return LevelWarn, nil
	case "ERROR":
		return LevelError, nil
	default:

		if i, err := strconv.Atoi(u); err == nil {
			return LogLevel(i), nil
		}
		return 0, fmt.Errorf("invalid log level: %v", s)
	}
}
