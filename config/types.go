package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Env defines how to read environment variables and distinguish environments. Useful
// for testing.
type Env interface {
	Type() string                        // dev, test, or prod
	Get(key string) string               // retrieve raw value
	LookupEnv(key string) (string, bool) // detect presence
}

// configEnv uses the OS environment to implement Env,
// so settings can be overridden without code changes.
type configEnv struct{}

var env = &configEnv{}

// Type returns the CLI environment type to select appropriate defaults.
func (e *configEnv) Type() string {
	val := os.Getenv(envType)
	if val == "" {
		return "prod"
	}
	return val
}

// Get returns the value of key or empty if unset.
func (e *configEnv) Get(key string) string {
	return os.Getenv(key)
}

// LookupEnv checks if key is present and returns its value.
func (e *configEnv) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

// UserDirs abstracts locations for config, cache, and home directories
// so OS differences are hidden from the rest of the code.
type UserDirs interface {
	ConfigDir() (string, error) // where to store user config files
	CacheDir() (string, error)  // where to store transient cache
	HomeDir() (string, error)   // user’s home for embedding resources
}

// userDirs implements UserDirs using the standard library,
// so callers get the platform-correct paths.
type userDirs struct{}

var user = userDirs{}

// ConfigDir returns the OS-specific configuration directory.
func (u *userDirs) ConfigDir() (string, error) { return os.UserConfigDir() }

// CacheDir returns the OS-specific cache directory.
func (u *userDirs) CacheDir() (string, error) { return os.UserCacheDir() }

// HomeDir returns the current user’s home directory.
func (u *userDirs) HomeDir() (string, error) { return os.UserHomeDir() }

// Duration is a thin wrapper around time.Duration that
// unmarshals from and marshals to a string like "30s", "5m".
type Duration time.Duration

// UnmarshalJSON implements json.Unmarshaler.
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("duration should be a string, got %s: %w", string(b), err)
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	*d = Duration(dur)
	return nil
}

// MarshalJSON implements json.Marshaler.
func (d Duration) MarshalJSON() ([]byte, error) {
	s := time.Duration(d).String()
	return json.Marshal(s)
}

// ToTimeDuration casts Duration back to time.Duration.
func (d Duration) ToTimeDuration() time.Duration {
	return time.Duration(d)
}
