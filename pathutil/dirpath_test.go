package pathutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDirPath_Validate(t *testing.T) {
	root := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir(root)

	parent := filepath.Dir(root)

	// crafted dirs and files
	outside := filepath.Join(parent, "evil")
	valid := filepath.Join(root, "okdir")
	tooLong := filepath.Join(root, strings.Repeat("a", 300))
	missing := filepath.Join(root, "ghost")
	notDir := filepath.Join(root, "afile")

	_ = os.Mkdir(outside, 0o755)
	_ = os.Mkdir(valid, 0o755)
	_ = os.Mkdir(tooLong, 0o755)
	_ = os.WriteFile(notDir, []byte("nope"), 0o644)

	cases := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name: "ValidDir",
			path: valid,
		},
		{
			name:    "DirectoryTraversal",
			path:    "../evil",
			wantErr: true,
		},
		{
			name:    "DirNotFound",
			path:    missing,
			wantErr: true,
		},
		{
			name:    "NotADirectory",
			path:    notDir,
			wantErr: true,
		},
		{
			name:    "DirNameTooLong",
			path:    tooLong,
			wantErr: true,
		},
		{
			name:    "InvalidAbsDir",
			path:    string([]byte{0}),
			wantErr: true,
		},
		{
			name:    "UnreadableParent",
			path:    "/nonexistent/parentdir",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dp := DirPath(tc.path)
			got, err := dp.Secure()

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil for path %q", tc.path)
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

	t.Run("ClosedDirFileStatFails", func(t *testing.T) {
		// simulate a closed directory handle: open then close
		dir := t.TempDir()
		f, _ := os.Open(dir)
		f.Close()
		_, err := f.Stat()
		if err == nil {
			t.Fatal("expected error from closed directory file Stat()")
		}
	})
}
