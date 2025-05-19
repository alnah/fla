package filesystem

import "os"

const (
	// User permissions
	PermUserRWX os.FileMode = 0o700
	PermUserRW  os.FileMode = 0o600
	PermUserR   os.FileMode = 0o400
)

type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Stat(path string) (os.FileInfo, error)
}
