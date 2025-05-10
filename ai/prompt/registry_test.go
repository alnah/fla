// registry_test.go
package prompt

import (
	"embed"
	"testing"
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

func assertNoErrorLoadingEmbedFS(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("loading embed filesystem: didn't want error, got %v", err)
	}
}

func assertNoErrorLoadingPromptData(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("loading prompt data: didn't want error, got: %v", err)
	}
}

func assertPromptData(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("loading prompt data: got %q, want %q", got, want)
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
	assertNoErrorLoadingEmbedFS(t, err)

	data, err := rt.Prompt(FR, A1, promptID("test"), map[string]any{"input": "is success"})
	assertNoErrorLoadingPromptData(t, err)

	assertPromptData(t, data, "Test is success")
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
	assertNoErrorLoadingEmbedFS(t, err)

	_, err = rt.Prompt(FR, A1, promptID("missing_key"), map[string]any{"missing": "key"})
	assertError(t, err)
}

func TestPromptLoading_FallbackLogic(t *testing.T) {
	testCases := []struct {
		name     string
		embedFS  embed.FS
		lang     language
		level    level
		promptID promptID
		input    string
		want     string
	}{
		{
			name:     "OverrideVersion",
			embedFS:  fallbackFS,
			lang:     PT,
			level:    A1,
			promptID: promptID("version"),
			input:    "ok",
			want:     "Version 2 ok",
		},
		{
			name:     "FallackToInferiorLevelForSameLanguage",
			embedFS:  fallbackFS,
			lang:     FR,
			level:    C1,
			promptID: promptID("fallback"),
			input:    "b2 ok",
			want:     "Fallback b2 ok",
		},
		{
			name:     "WildcardAnyLevel",
			embedFS:  fallbackFS,
			lang:     FR,
			level:    AnyLevel,
			promptID: promptID("wildcard"),
			input:    "ok",
			want:     "Any level ok",
		},
	}
	for _, tc := range testCases {
		rt, err := registry(tc.embedFS)
		assertNoErrorLoadingEmbedFS(t, err)

		data, err := rt.Prompt(tc.lang, tc.level, tc.promptID, map[string]any{"input": tc.input})
		assertNoErrorLoadingPromptData(t, err)
		assertPromptData(t, data, tc.want)

	}
}

func TestPromptLoading_NotFound(t *testing.T) {
	rt, err := registry(fallbackFS)
	assertNoErrorLoadingEmbedFS(t, err)

	_, err = rt.Prompt(PT, C2, promptID("not_found"), map[string]any{"input": "irrelevant"})
	assertError(t, err)
}
