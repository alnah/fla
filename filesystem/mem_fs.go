package filesystem

import (
	"os"
	"path"
	"slices"
	"strings"
	"sync"
	"time"
)

// memFS holds file data and directory structure in memory to
// enable fast, isolated tests without external dependencies.
type memFS struct {
	files map[string][]byte    // file contents by path
	dirs  map[string]struct{}  // set of existing directory paths
	times map[string]time.Time // last-modified timestamps
	mu    sync.RWMutex         // guards access to internal maps
}

// NewMemFS returns a fresh in-memory FileSystem, ensuring
// tests can create, read, and manage files without disk I/O.
func NewMemFS() FileSystem {
	return &memFS{
		files: make(map[string][]byte),
		dirs:  map[string]struct{}{"": {}}, // root directory always exists
		times: make(map[string]time.Time),
	}
}

// Root returns a dummy root path.
func (m *memFS) Root() string { return "." }

// ReadFile returns a copy of the stored bytes for path or ErrNotExist.
// Copying prevents tests from accidentally mutating internal state.
func (m *memFS) ReadFile(p string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if data, ok := m.files[p]; ok {
		return slices.Clone(data), nil
	}
	return nil, os.ErrNotExist
}

// WriteFile stores data at path and records its modification time.
// It auto-creates any parent directories so tests need not manage them.
func (m *memFS) WriteFile(p string, data []byte, perm os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// ensure parent directories exist for nested writes
	dir := path.Dir(p)
	if dir != "." && dir != "" {
		if err := m.MkdirAll(dir, PermUserRWX); err != nil {
			return err
		}
	}

	// store a copy of data to avoid external mutations
	m.files[p] = slices.Clone(data)
	m.times[p] = time.Now()
	return nil
}

// MkdirAll creates the directory hierarchy for p in memory,
// recording timestamps so Stat can report modification times.
func (m *memFS) MkdirAll(p string, perm os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	parts := strings.Split(p, string(os.PathSeparator))
	for i := range parts {
		sub := strings.Join(parts[:i+1], string(os.PathSeparator))
		if _, exists := m.dirs[sub]; !exists {
			m.dirs[sub] = struct{}{}
			m.times[sub] = time.Now()
		}
	}
	return nil
}

// Stat returns file information for p, letting tests verify
// existence, size, mode, and directory status without touching disk.
func (m *memFS) Stat(p string) (os.FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, isDir := m.dirs[p]; isDir {
		return &memInfo{name: path.Base(p), mode: os.ModeDir | PermUserRWX, mod: m.times[p]}, nil
	}
	if data, ok := m.files[p]; ok {
		return &memInfo{name: path.Base(p), size: int64(len(data)), mode: PermUserRW, mod: m.times[p]}, nil
	}
	return nil, os.ErrNotExist
}

// memInfo implements os.FileInfo over in-memory entries so tests
// can assert metadata like Name, Size, Mode, and ModTime.
type memInfo struct {
	name string      // base name of file or directory
	size int64       // length in bytes for files
	mode os.FileMode // file mode bits (PermUserRWX or PermUserRW)
	mod  time.Time   // last modification timestamp
}

// Name returns the base name for the file or directory.
func (m *memInfo) Name() string { return m.name }

// Size reports the stored size in bytes, or zero for directories.
func (m *memInfo) Size() int64 { return m.size }

// Mode returns the mode bits, indicating permissions and dir-ness.
func (m *memInfo) Mode() os.FileMode { return m.mode }

// ModTime returns when the entry was last modified in this memFS.
func (m *memInfo) ModTime() time.Time { return m.mod }

// IsDir reports whether this entry represents a directory.
func (m *memInfo) IsDir() bool { return m.mode.IsDir() }

// Sys returns underlying data source (none for in-memory FS).
func (m *memInfo) Sys() any { return nil }
