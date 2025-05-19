package filesystem

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	pu "github.com/alnah/fla/pathutil"
)

// DiskFS implements FileSystem using the OS filesystem with a root directory.
type DiskFS struct {
	root string
	fs   fs.FS
}

// New returns a DiskFS rooted at the given directory.
func New(root string) *DiskFS {
	return &DiskFS{root: root, fs: os.DirFS(root)}
}

// ReadFile reads the file at the given path relative to the root.
func (d *DiskFS) ReadFile(path string) ([]byte, error) {
	return fs.ReadFile(d.fs, path)
}

// WriteFile writes data to path (relative to root) atomically with the given permissions.
// It uses a temporary file in the same directory under the root and renames it on success.
func (d *DiskFS) WriteFile(path string, data []byte, perm os.FileMode) (err error) {
	fullPath := filepath.Join(d.root, path)
	// ensure existing path is a regular file if it exists
	if fi, statErr := os.Stat(fullPath); statErr == nil && !fi.Mode().IsRegular() {
		return fmt.Errorf("%s already exists and is not a regular file", fullPath)
	}

	dir := filepath.Join(d.root, filepath.Dir(path))
	base := filepath.Base(path)
	// create a temporary file for atomic write
	f, err := os.CreateTemp(dir, base+".tmp")
	if err != nil {
		return err
	}
	tmpName := f.Name()
	defer func() {
		if err != nil {
			_ = f.Close()
			_ = os.Remove(tmpName)
		}
	}()

	// write data
	if _, err = f.Write(data); err != nil {
		return err
	}

	// apply permission bits (skip on windows if not supported)
	if runtime.GOOS != "windows" {
		if err = f.Chmod(perm); err != nil {
			return err
		}
	}

	// ensure content is flushed to disk
	if err = f.Sync(); err != nil {
		return err
	}

	// close before rename
	if err = f.Close(); err != nil {
		return err
	}

	err = os.Rename(tmpName, fullPath)
	if err != nil {
		return err
	}

	// optionally sync parent directory for durability
	if dir, err = pu.DirPath(dir).Secure(); err != nil {
		return err
	}
	// #nosec G304: `dir` has been canonicalized and secured via DirPath.Secure()
	if dirF, derr := os.Open(dir); derr == nil {
		_ = dirF.Sync()
		_ = dirF.Close()
	}

	return nil
}

// MkdirAll creates a directory (under root) and all necessary parents with the given permissions.
func (d *DiskFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(filepath.Join(d.root, path), perm)
}

// Stat returns a FileInfo describing the named file under the root.
func (d *DiskFS) Stat(path string) (os.FileInfo, error) {
	return os.Stat(filepath.Join(d.root, path))
}
