package config

import "os"

type UserDirs interface {
	ConfigDir() (string, error)
	CacheDir() (string, error)
	HomeDir() (string, error)
}

type userDirs struct{}

var user = userDirs{}

func (u *userDirs) ConfigDir() (string, error) { return os.UserConfigDir() }
func (u *userDirs) CacheDir() (string, error)  { return os.UserCacheDir() }
func (u *userDirs) HomeDir() (string, error)   { return os.UserHomeDir() }
