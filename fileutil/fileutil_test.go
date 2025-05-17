package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilePath_Validate(t *testing.T) {
	const oneMB = 1024 * 1024

	root := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir(root)

	parent := filepath.Dir(root)

	// crafted files
	outside := filepath.Join(parent, "evil.txt")                    // traversal target
	small := filepath.Join(root, "ok.txt")                          // valid tiny file
	upper := filepath.Join(root, "DATA.TXT")                        // upper-case ext
	big := filepath.Join(root, "huge.log")                          // >1 MB
	tooLong := filepath.Join(root, strings.Repeat("a", 300)+".txt") // 300-rune name
	missing := filepath.Join(root, "ghost.bin")                     // absent file

	_ = os.WriteFile(outside, []byte("x"), 0o644)
	_ = os.WriteFile(small, []byte("ok"), 0o644)
	_ = os.WriteFile(upper, []byte("ok"), 0o644)
	_ = os.WriteFile(big, make([]byte, oneMB+1), 0o644)

	cases := []struct {
		name    string
		path    string
		maxMB   int64
		allowed []string
		wantErr bool
	}{
		{
			name:    "ValidFile",
			path:    small,
			maxMB:   1,
			allowed: []string{"txt"},
		},
		{
			name:    "ExtensionNotAllowed",
			path:    small,
			maxMB:   1,
			allowed: []string{"log"},
			wantErr: true,
		},
		{
			name:    "FileTooBig",
			path:    big,
			maxMB:   1,
			allowed: []string{"log"},
			wantErr: true,
		},
		{
			name:    "FileNameTooLong",
			path:    tooLong,
			maxMB:   1,
			allowed: []string{"txt"},
			wantErr: true,
		},
		{
			name:    "DirectoryTraversal",
			path:    "../evil.txt",
			maxMB:   1,
			allowed: []string{"txt"},
			wantErr: true,
		},
		{
			name:    "NoAllowedList",
			path:    upper,
			maxMB:   1,
			allowed: nil,
			wantErr: false,
		},
		{
			name:    "CaseInsensitiveExt",
			path:    upper,
			maxMB:   1,
			allowed: []string{"txt"},
			wantErr: false,
		},
		{
			name:    "NoSizeLimit",
			path:    big,
			maxMB:   0,
			allowed: []string{"log"},
			wantErr: false,
		},
		{
			name:    "FileNotFound",
			path:    missing,
			maxMB:   1,
			allowed: []string{"bin"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fp := FilePath(tc.path)
			got, err := fp.Validate(tc.maxMB, tc.allowed...)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !filepath.IsAbs(got) {
				t.Errorf("returned path %s is not absolute", got)
			}
		})
	}
}
