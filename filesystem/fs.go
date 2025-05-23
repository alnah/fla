package filesystem

import "os"

// Permission bits for FileSystem implementations to promote consistency.
const (
	PermUserRWX os.FileMode = 0o700 // owner read/write/execute
	PermUserRW  os.FileMode = 0o600 // owner read/write
	PermUserR   os.FileMode = 0o400 // owner read-only
)

// FileSystem abstracts file operations so code can swap in-memory
// or disk-based implementations without changing callers.
type FileSystem interface {
	Root() string
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Stat(path string) (os.FileInfo, error)
}
