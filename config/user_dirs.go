package config

import "os"

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
