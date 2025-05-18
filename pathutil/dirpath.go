package pathutil

import (
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"
)

// DirPath wraps a filesystem path and provides a validation helper.
type DirPath string

// Validate ensures a directory path is safe and sane to consume. It:
//   - Normalizes & canonicalizes the path.
//   - Ensures the base name isn't too long.
//   - Opens the parent directory via OpenRoot to avoid TOCTOU/traversal attacks.
//   - Checks that the target exists and is a directory.
//   - Returns the canonical absolute path.
func (d DirPath) Validate() (string, error) {
	// clean and normalize the path
	clean := filepath.Clean(string(d))

	// check base name length
	base := filepath.Base(clean)
	if utf8.RuneCountInString(base) > 255 {
		return "", fmt.Errorf("directory name too long: max 255 runes, got %d", utf8.RuneCountInString(base))
	}

	// resolve parent directory and relative component
	absParent, err := filepath.Abs(filepath.Dir(clean))
	if err != nil {
		return "", fmt.Errorf("abs parent dir: %w", err)
	}
	relName := filepath.Base(clean)

	// open root of parent directory
	root, err := os.OpenRoot(absParent)
	if err != nil {
		return "", fmt.Errorf("open root directory %s: %w", absParent, err)
	}
	defer func() { _ = root.Close() }()

	// open the target directory under the safe root
	dirFile, err := root.Open(relName)
	if err != nil {
		return "", fmt.Errorf("open directory %s: %w", relName, err)
	}
	defer func() { _ = dirFile.Close() }()

	// verify it's a directory
	info, err := dirFile.Stat()
	if err != nil {
		return "", fmt.Errorf("stat directory: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("path %s is not a directory", clean)
	}

	// return the canonical absolute path
	absPath, err := filepath.Abs(clean)
	if err != nil {
		return "", fmt.Errorf("abs path: %w", err)
	}
	return absPath, nil
}
