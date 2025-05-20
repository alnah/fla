package filesystem

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	pu "github.com/alnah/fla/pathutil"
)

// DiskFS persists files under a real directory, ensuring production data
// lives on disk. Atomic writes guard against partial files on failures.
type DiskFS struct {
	root string
	fs   fs.FS
}

// New returns a DiskFS rooted at the given directory so
// all operations stay sandboxed within it.
func New(root string) *DiskFS {
	return &DiskFS{root: root, fs: os.DirFS(root)}
}

// ReadFile loads a file’s bytes from under root for reliable access.
func (d *DiskFS) ReadFile(path string) ([]byte, error) {
	return fs.ReadFile(d.fs, path)
}

// WriteFile writes data atomically to avoid corrupted files,
// using a temp file and rename under the same directory.
// If the file doesn't exist, it is created.
func (d *DiskFS) WriteFile(path string, data []byte, perm os.FileMode) (err error) {
	fullPath := filepath.Join(d.root, path)

	// refuse to overwrite non-regular files
	if fi, statErr := os.Stat(fullPath); statErr == nil && !fi.Mode().IsRegular() {
		return fmt.Errorf("%s exists and isn’t a regular file", fullPath)
	}

	dir := filepath.Join(d.root, filepath.Dir(path))
	base := filepath.Base(path)

	// create temp file alongside target for atomicity
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

	if _, err = f.Write(data); err != nil {
		return err
	}

	// preserve permissions for non-Windows systems
	if runtime.GOOS != "windows" {
		if err = f.Chmod(perm); err != nil {
			return err
		}
	}

	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}

	if err = os.Rename(tmpName, fullPath); err != nil {
		return err
	}

	// ensure directory entry is flushed for durability
	if secureDir, serr := pu.DirPath(dir).Secure(); serr == nil {
		if dirF, derr := os.Open(secureDir); derr == nil {
			_ = dirF.Sync()
			_ = dirF.Close()
		}
	}

	return nil
}

// MkdirAll creates nested directories under root so consumers
// needn't manage hierarchy themselves.
func (d *DiskFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(filepath.Join(d.root, path), perm)
}

// Stat reports file or directory info under root, mirroring os.Stat.
func (d *DiskFS) Stat(path string) (os.FileInfo, error) {
	return os.Stat(filepath.Join(d.root, path))
}
