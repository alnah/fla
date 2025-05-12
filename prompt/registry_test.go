// registry_test.go
package prompt

import (
	"embed"
	"testing"

	"github.com/alnah/fla/locale"
)

//go:embed test/ok/*/*/*.md
var okFS embed.FS

//go:embed test/ok/*/*/*.md
var fallbackFS embed.FS

//go:embed test/sad/no_front_matter_v1.md
var noFrontMatterFS embed.FS

//go:embed test/sad/bad_json_header_v1.md
var badJsonFS embed.FS

//go:embed test/sad/bad_level_v1.md
var badLevelFS embed.FS

//go:embed test/sad/bad_lang_v1.md
var BadLangFS embed.FS

//go:embed test/sad/missing_key_v1.md
var missingKeyFS embed.FS

/********* Helpers *********/

func assertNoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
}

func assertValue(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func assertError(t testing.TB, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("want error, got nil")
	}
}

/********* Unit tests *********/

func TestPromptLoading_Success(t *testing.T) {
	rt, err := registry(okFS)
	assertNoError(t, err)

	data, err := rt.Prompt(locale.FR, A1, promptID("test"), map[string]any{"input": "is success"})
	assertNoError(t, err)
	assertValue(t, data, "Test is success")
}

func TestPromptRuntime_ErrorFormat(t *testing.T) {
	testCases := []struct {
		name    string
		embedFS embed.FS
	}{
		{name: "NoFrontMatter", embedFS: noFrontMatterFS},
		{name: "BadJSON_Header", embedFS: badJsonFS},
		{name: "BadLevel", embedFS: badLevelFS},
		{name: "BadLang", embedFS: BadLangFS},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := registry(tc.embedFS)
			assertError(t, err)
		})
	}
}

func TestPromptLoading_MissingKey(t *testing.T) {
	rt, err := registry(missingKeyFS)
	assertNoError(t, err)

	_, err = rt.Prompt(locale.FR, A1, promptID("missing_key"), map[string]any{"missing": "key"})
	assertError(t, err)
}

func TestPromptLoading_FallbackLogic(t *testing.T) {
	testCases := []struct {
		name     string
		embedFS  embed.FS
		lang     locale.Lang
		level    level
		promptID promptID
		input    string
		want     string
	}{
		{
			name:     "OverrideVersion",
			embedFS:  fallbackFS,
			lang:     locale.PT,
			level:    A1,
			promptID: promptID("version"),
			input:    "ok",
			want:     "Version 2 ok",
		},
		{
			name:     "FallbackPrevLevelSameLang",
			embedFS:  fallbackFS,
			lang:     locale.FR,
			level:    C1,
			promptID: promptID("fallback"),
			input:    "b2 ok",
			want:     "Fallback b2 ok",
		},
		{
			name:     "WildcardAnyLevel",
			embedFS:  fallbackFS,
			lang:     locale.FR,
			level:    AnyLevel,
			promptID: promptID("wildcard"),
			input:    "ok",
			want:     "Any level ok",
		},
	}
	for _, tc := range testCases {
		rt, err := registry(tc.embedFS)
		assertNoError(t, err)

		data, err := rt.Prompt(tc.lang, tc.level, tc.promptID, map[string]any{"input": tc.input})
		assertNoError(t, err)
		assertValue(t, data, tc.want)

	}
}

func TestPromptLoading_NotFound(t *testing.T) {
	rt, err := registry(fallbackFS)
	assertNoError(t, err)

	_, err = rt.Prompt(locale.PT, C2, promptID("not_found"), map[string]any{"input": "irrelevant"})
	assertError(t, err)
}
