//go:build integration

package locale

import (
	"encoding/json"
	"flag"
	"testing"
)

func TestLang_FlagValue(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		wantErr  bool
		wantLang Lang
	}{
		{name: "Valid", input: "-lang=fr", wantErr: false, wantLang: FR},
		{name: "Invalid", input: "-lang=zz", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)

			var lang Lang
			fs.Var(&lang, "lang", "language")

			err := fs.Parse([]string{tc.input})
			if tc.wantErr {
				if err == nil {
					t.Fatalf("parsing: want error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("parsing: want no error, got %v", err)
				}
				if lang != tc.wantLang {
					t.Errorf("post-parsing: want %v, got %v", lang, PT)
				}
			}
		})
	}
}

func TestLang_JSONUnmarshal(t *testing.T) {
	type wrapper struct {
		Language Lang `json:"language"`
	}

	testCases := []struct {
		name     string
		byt      []byte
		wantErr  bool
		wantLang Lang
	}{
		{name: "Valid", byt: []byte(`{"language": "fr"}`), wantErr: false, wantLang: FR},
		{name: "Invalid", byt: []byte(`{"language": "zz"}`), wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var w wrapper
			err := json.Unmarshal(tc.byt, &w)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("unmarshaling: want error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unmarshaling: want no error, got %v", err)
				}
				if w.Language != tc.wantLang {
					t.Errorf("unmarshaling: want %v, got %v", tc.wantLang, w.Language)
				}
			}
		})
	}
}
