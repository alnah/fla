package pathutil

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
	// clean and extract extension
	clean := filepath.Clean(string(f))
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(clean), "."))

	// build allowed set
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
	if utf8.RuneCountInString(base) > 255 {
		return "", fmt.Errorf("file name too long: max 255 runes, got %d", utf8.RuneCountInString(base))
	}

	// resolve directory and relative path
	absDir, err := filepath.Abs(filepath.Dir(clean))
	if err != nil {
		return "", fmt.Errorf("abs dir: %w", err)
	}
	relPath := filepath.Base(clean)

	// open root directory
	root, err := os.OpenRoot(absDir)
	if err != nil {
		return "", fmt.Errorf("open root directory %s: %w", absDir, err)
	}
	defer func() { _ = root.Close() }()

	// open file within root
	file, err := root.Open(relPath)
	if err != nil {
		return "", fmt.Errorf("open file %s: %w", relPath, err)
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
		return fmt.Errorf("file too big: max %d MB", maxSizeMB)
	}
	return nil
}
