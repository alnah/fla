package config

import "fmt"

// ConfigError wraps any error that occurs during a specific
// configuration stage, giving context for troubleshooting.
type ConfigError struct {
	Stage string // operation being performed (e.g., "reading json")
	Cause error  // underlying error
}

// NewConfigError tags an error with its configuration stage.
func NewConfigError(stage string, err error) *ConfigError {
	return &ConfigError{
		Stage: stage,
		Cause: err,
	}
}

// Error formats the ConfigError to include both stage and cause
// so logs clearly indicate where configuration failed.
func (e *ConfigError) Error() string {
	return fmt.Sprintf("configuration %s: %v", e.Stage, e.Cause)
}

// Unwrap allows errors.Unwrap to retrieve the underlying cause.
func (e *ConfigError) Unwrap() error { return e.Cause }
