package config

import "fmt"

// ConfigError wraps any error that occurs during a specific
// configuration stage, providing context for troubleshooting.
type ConfigError struct {
	Stage string // operation being performed (e.g., "reading json")
	Cause error  // underlying error
}

// newConfigError tags an error with its configuration stage.
func newConfigError(stage string, err error) *ConfigError {
	return &ConfigError{
		Stage: stage,
		Cause: err,
	}
}

// Error formats the ConfigError to include both stage and cause
// so logs clearly indicate where configuration failed.
func (e *ConfigError) Error() string {
	if e.Stage != "" && e.Cause != nil {
		return fmt.Sprintf("config: %s: %v", e.Stage, e.Cause)
	}
	if e.Stage != "" {
		return fmt.Sprintf("config: %s", e.Stage)
	}
	if e.Cause != nil {
		return fmt.Sprintf("config: %v", e.Cause)
	}
	return "config: unknown error"
}

// Unwrap allows errors.Unwrap to retrieve the underlying cause.
func (e *ConfigError) Unwrap() error {
	return e.Cause
}
