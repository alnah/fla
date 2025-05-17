package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilePath_Validate(t *testing.T) {
	// set up a temporary root directory
	root := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir(root)

	// create a file outside of root for the traversal test
	parent := filepath.Dir(root)
	outside := filepath.Join(parent, "evil.txt")
	os.WriteFile(outside, []byte("malicious"), 0o644)

	// create a subdirectory for test files
	sub := filepath.Join(root, "sub")
	os.Mkdir(sub, 0o755)

	// small valid file
	small := filepath.Join(sub, "ok.txt")
	os.WriteFile(small, []byte("hello"), 0o644)

	// large file exceeding 1 MB
	const oneMB = 1024 * 1024
	big := filepath.Join(sub, "big.log")
	os.WriteFile(big, make([]byte, oneMB+1), 0o644)

	type fields struct {
		path       string
		maxSizeMB  int64
		allowedExt []string
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr string
	}{
		{
			name: "ValidFile",
			fields: fields{
				path:       small,
				maxSizeMB:  1,
				allowedExt: []string{"txt"},
			},
		},
		{
			name: "ExtensionNotAllowed",
			fields: fields{
				path:       small,
				maxSizeMB:  1,
				allowedExt: []string{"log"},
			},
			wantErr: "extension txt not allowed",
		},
		{
			name: "FileTooBig",
			fields: fields{
				path:       big,
				maxSizeMB:  1,
				allowedExt: []string{"log"},
			},
			wantErr: "too big",
		},
		{
			name: "FileNameTooLong",
			fields: fields{
				path:       strings.Repeat("a", 300) + ".txt",
				maxSizeMB:  1,
				allowedExt: []string{"txt"},
			},
			wantErr: "file name too long",
		},
		{
			name: "DirectoryTraversalAttempt",
			fields: fields{
				path:       "../etc/evil.txt",
				maxSizeMB:  1,
				allowedExt: []string{"txt"},
			},
			wantErr: "unsafe path",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fp := FilePath(tc.fields.path)
			got, err := fp.Validate(tc.fields.maxSizeMB, tc.fields.allowedExt...)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("want error containing %s, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("want no error, got %v", err)
			}
			if !filepath.IsAbs(got) {
				t.Errorf("returned path %s; want absolute path", got)
			}
			if !strings.HasPrefix(got, root) {
				t.Errorf("returned path %s; want under %s", got, root)
			}
		})
	}
}
