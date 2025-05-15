package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// FilePath wraps a filesystem path and provides a validation helper.
type FilePath string

// Validate ensures a file is safe and sane to consume. It:
//   - Normalizes & canonicalizes the path.
//   - Ensures the extension is in the allowed list (case-insensitive).
//   - Opens the file via openRoot to avoid TOCTOU/traversal attacks.
//   - Checks file size ≤ maxSizeMB (if >0).
//   - Returns the canonical absolute path.
func (f FilePath) Validate(maxSizeMB int64, allowedExt ...string) (string, error) {
	// clean, and extract extension
	clean := filepath.Clean(string(f))
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(clean), "."))

	// build allowed set once
	allowedSet := make(map[string]struct{}, len(allowedExt))
	for _, e := range allowedExt {
		allowedSet[strings.ToLower(strings.TrimPrefix(e, "."))] = struct{}{}
	}
	if len(allowedSet) > 0 {
		if _, ok := allowedSet[ext]; !ok {
			return "", fmt.Errorf("extension %s not allowed", ext)
		}
	}

	// check filename length
	base := filepath.Base(clean)
	n := utf8.RuneCountInString(base)
	if n > 255 {
		return "", fmt.Errorf("file name too long: %d runes > %d", n, 255)
	}

	// resolve directory
	dir := filepath.Dir(clean)
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("abs dir: %w", err)
	}

	// open directory and file via handle
	root, err := os.Open(absDir)
	if err != nil {
		return "", fmt.Errorf("open directory %s: unsafe path", absDir)
	}
	defer func() { _ = root.Close() }()

	file, err := os.Open(clean)
	if err != nil {
		return "", fmt.Errorf("open file %s: %w", clean, err)
	}
	defer func() { _ = file.Close() }()

	// size check
	if err := checkSize(file, maxSizeMB); err != nil {
		return "", fmt.Errorf("size check: %w", err)
	}

	// return absolute path
	absPath, err := filepath.Abs(clean)
	if err != nil {
		return "", fmt.Errorf("abs path: %w", err)
	}
	return absPath, nil
}

func checkSize(f *os.File, maxSizeMB int64) error {
	if maxSizeMB <= 0 {
		return nil
	}
	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}
	if info.Size() > maxSizeMB*1024*1024 {
		return fmt.Errorf("file too big: %.2f MB > %d MB", float64(info.Size())/1024.0/1024.0, maxSizeMB)
	}
	return nil
}
