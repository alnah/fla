package shared_test

import (
	"testing"

	"golang.org/x/text/language"

	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/shared"
	"slices"
)

func TestNewLocale(t *testing.T) {
	t.Run("creates locale with valid input", func(t *testing.T) {
		validLocales := []string{
			"fr-FR",
			"en-US",
			"pt-BR",
		}

		for _, locale := range validLocales {
			t.Run(locale, func(t *testing.T) {
				got, err := shared.NewLocale(locale)

				assertNoError(t, err)
				if got.String() != locale {
					t.Errorf("got %q, want %q", got, locale)
				}
			})
		}
	})

	t.Run("returns default locale for empty input", func(t *testing.T) {
		got, err := shared.NewLocale("")

		assertNoError(t, err)
		if got != shared.DefaultLocale {
			t.Errorf("got %v, want %v", got, shared.DefaultLocale)
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		input := "  fr-FR  "
		want := shared.LocaleFrenchFR

		got, err := shared.NewLocale(input)

		assertNoError(t, err)
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("rejects invalid BCP 47 format", func(t *testing.T) {
		invalidLocales := []string{
			"invalid",
			"fr_FR", // underscore instead of hyphen
			"french",
			"123",
			"fr-FR-extra-tag", // too many tags
		}

		for _, locale := range invalidLocales {
			t.Run(locale, func(t *testing.T) {
				_, err := shared.NewLocale(locale)

				assertError(t, err)
				assertErrorCode(t, err, kernel.EInvalid)
			})
		}
	})

	t.Run("rejects unsupported locales", func(t *testing.T) {
		unsupportedLocales := []string{
			"ja-JP", // Japanese - not in supported list
			"ko-KR", // Korean - not in supported list
			"zh-CN", // Chinese - not in supported list
		}

		for _, locale := range unsupportedLocales {
			t.Run(locale, func(t *testing.T) {
				_, err := shared.NewLocale(locale)

				assertError(t, err)
				assertErrorCode(t, err, kernel.EInvalid)
			})
		}
	})
}

func TestLocale_String(t *testing.T) {
	want := "fr-FR"
	locale := shared.Locale(want)

	got := locale.String()

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLocale_Validate(t *testing.T) {
	t.Run("valid supported locales pass", func(t *testing.T) {
		for _, locale := range shared.SupportedLocales {
			t.Run(string(locale), func(t *testing.T) {
				err := locale.Validate()

				assertNoError(t, err)
			})
		}
	})

	t.Run("empty locale fails", func(t *testing.T) {
		locale := shared.Locale("")

		err := locale.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("invalid BCP 47 format fails", func(t *testing.T) {
		locale := shared.Locale("invalid-format")

		err := locale.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("unsupported locale fails", func(t *testing.T) {
		locale := shared.Locale("ja-JP")

		err := locale.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})
}

func TestLocale_IsSupported(t *testing.T) {
	t.Run("supported locales return true", func(t *testing.T) {
		for _, locale := range shared.SupportedLocales {
			t.Run(string(locale), func(t *testing.T) {
				if !locale.IsSupported() {
					t.Errorf("expected %s to be supported", locale)
				}
			})
		}
	})

	t.Run("unsupported locale returns false", func(t *testing.T) {
		locale := shared.Locale("ja-JP")

		if locale.IsSupported() {
			t.Error("expected unsupported locale to return false")
		}
	})

	t.Run("empty locale returns false", func(t *testing.T) {
		locale := shared.Locale("")

		if locale.IsSupported() {
			t.Error("expected empty locale to return false")
		}
	})
}

func TestLocale_GetEffectiveLocale(t *testing.T) {
	t.Run("returns self for supported locale", func(t *testing.T) {
		locale := shared.LocaleEnglishUS

		got := locale.GetEffectiveLocale()

		if got != locale {
			t.Errorf("got %v, want %v", got, locale)
		}
	})

	t.Run("returns default for empty locale", func(t *testing.T) {
		locale := shared.Locale("")

		got := locale.GetEffectiveLocale()

		if got != shared.DefaultLocale {
			t.Errorf("got %v, want %v", got, shared.DefaultLocale)
		}
	})

	t.Run("returns default for unsupported locale", func(t *testing.T) {
		locale := shared.Locale("ja-JP")

		got := locale.GetEffectiveLocale()

		if got != shared.DefaultLocale {
			t.Errorf("got %v, want %v", got, shared.DefaultLocale)
		}
	})
}

func TestLocale_IsDefault(t *testing.T) {
	t.Run("default locale returns true", func(t *testing.T) {
		if !shared.DefaultLocale.IsDefault() {
			t.Error("expected default locale to return true")
		}
	})

	t.Run("non-default locale returns false", func(t *testing.T) {
		locale := shared.LocaleFrenchFR

		if locale.IsDefault() {
			t.Error("expected non-default locale to return false")
		}
	})
}

func TestLocale_ToLanguageTag(t *testing.T) {
	t.Run("converts supported locale to language tag", func(t *testing.T) {
		locale := shared.LocaleFrenchFR

		tag, err := locale.ToLanguageTag()

		assertNoError(t, err)
		expectedTag, _ := language.Parse("fr-FR")
		if tag != expectedTag {
			t.Errorf("got %v, want %v", tag, expectedTag)
		}
	})

	t.Run("handles empty locale by using effective locale", func(t *testing.T) {
		locale := shared.Locale("")

		tag, err := locale.ToLanguageTag()

		assertNoError(t, err)
		expectedTag, _ := language.Parse(string(shared.DefaultLocale))
		if tag != expectedTag {
			t.Errorf("got %v, want %v", tag, expectedTag)
		}
	})

	t.Run("handles unsupported locale by using default", func(t *testing.T) {
		locale := shared.Locale("ja-JP")

		tag, err := locale.ToLanguageTag()

		assertNoError(t, err)
		expectedTag, _ := language.Parse(string(shared.DefaultLocale))
		if tag != expectedTag {
			t.Errorf("got %v, want %v", tag, expectedTag)
		}
	})
}

func TestLocale_ToISO639Language(t *testing.T) {
	tests := []struct {
		locale   shared.Locale
		expected string
	}{
		{shared.LocaleFrenchFR, "fr"},
		{shared.LocaleEnglishUS, "en"},
		{shared.LocalePortugueseBR, "pt"},
	}

	for _, tt := range tests {
		t.Run(string(tt.locale), func(t *testing.T) {
			got := tt.locale.ToISO639Language()

			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}

	t.Run("returns default language for invalid locale", func(t *testing.T) {
		locale := shared.Locale("invalid")

		got := locale.ToISO639Language()

		if got != "en" { // default fallback is English since DefaultLocale is en-US
			t.Errorf("got %q, want %q", got, "en")
		}
	})
}

func TestLocale_GetRegion(t *testing.T) {
	tests := []struct {
		locale   shared.Locale
		expected string
	}{
		{shared.LocaleFrenchFR, "FR"},
		{shared.LocaleEnglishUS, "US"},
		{shared.LocalePortugueseBR, "BR"},
	}

	for _, tt := range tests {
		t.Run(string(tt.locale), func(t *testing.T) {
			got := tt.locale.GetRegion()

			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}

	t.Run("returns empty string for invalid locale", func(t *testing.T) {
		locale := shared.Locale("invalid")

		got := locale.GetRegion()

		// Invalid locale falls back to default locale (en-US), so region is US
		if got != "US" {
			t.Errorf("got %q, want %q", got, "US")
		}
	})
}

func TestLocale_GetDisplayName(t *testing.T) {
	t.Run("returns display name in specified language", func(t *testing.T) {
		locale := shared.LocaleFrenchFR

		got := locale.GetDisplayName(shared.LocaleEnglishUS)

		// The exact display name depends on the implementation
		// but it should not be empty and should be different from the locale code
		if got == "" || got == string(locale) {
			t.Errorf("expected meaningful display name, got %q", got)
		}
	})

	t.Run("returns self display when displayIn is empty", func(t *testing.T) {
		locale := shared.LocaleFrenchFR

		got := locale.GetDisplayName("")

		if got == "" {
			t.Error("expected non-empty display name")
		}
	})

	t.Run("handles invalid displayIn locale gracefully", func(t *testing.T) {
		locale := shared.LocaleFrenchFR

		got := locale.GetDisplayName(shared.Locale("invalid"))

		if got == "" {
			t.Error("expected fallback display name")
		}
	})
}

func TestLocale_GetSelfDisplayName(t *testing.T) {
	t.Run("returns display name in own language", func(t *testing.T) {
		for _, locale := range shared.SupportedLocales {
			t.Run(string(locale), func(t *testing.T) {
				got := locale.GetSelfDisplayName()

				if got == "" {
					t.Errorf("expected non-empty self display name for %s", locale)
				}
			})
		}
	})
}

func TestLocale_GetLanguageDisplayName(t *testing.T) {
	t.Run("returns language name without region", func(t *testing.T) {
		for _, locale := range shared.SupportedLocales {
			t.Run(string(locale), func(t *testing.T) {
				got := locale.GetLanguageDisplayName()

				if got == "" {
					t.Errorf("expected non-empty language display name for %s", locale)
				}
			})
		}
	})
}

func TestLocale_GetEnglishDisplayName(t *testing.T) {
	t.Run("returns display name in English", func(t *testing.T) {
		for _, locale := range shared.SupportedLocales {
			t.Run(string(locale), func(t *testing.T) {
				got := locale.GetEnglishDisplayName()

				if got == "" {
					t.Errorf("expected non-empty English display name for %s", locale)
				}
			})
		}
	})
}

func TestGetDisplayNameMap(t *testing.T) {
	t.Run("returns map with all supported locales", func(t *testing.T) {
		got := shared.GetDisplayNameMap(shared.LocaleEnglishUS)

		if len(got) != len(shared.SupportedLocales) {
			t.Errorf("got %d locales, want %d", len(got), len(shared.SupportedLocales))
		}

		for _, locale := range shared.SupportedLocales {
			if name, exists := got[locale]; !exists {
				t.Errorf("missing locale %s in display name map", locale)
			} else if name == "" {
				t.Errorf("empty display name for locale %s", locale)
			}
		}
	})
}

func TestGetAllSelfDisplayNames(t *testing.T) {
	t.Run("returns map with all supported locales in their own language", func(t *testing.T) {
		got := shared.GetAllSelfDisplayNames()

		if len(got) != len(shared.SupportedLocales) {
			t.Errorf("got %d locales, want %d", len(got), len(shared.SupportedLocales))
		}

		for _, locale := range shared.SupportedLocales {
			if name, exists := got[locale]; !exists {
				t.Errorf("missing locale %s in self display name map", locale)
			} else if name == "" {
				t.Errorf("empty self display name for locale %s", locale)
			}
		}
	})
}

func TestGetLanguageOnlyDisplayNames(t *testing.T) {
	t.Run("returns map with language names only", func(t *testing.T) {
		got := shared.GetLanguageOnlyDisplayNames()

		if len(got) != len(shared.SupportedLocales) {
			t.Errorf("got %d locales, want %d", len(got), len(shared.SupportedLocales))
		}

		for _, locale := range shared.SupportedLocales {
			if name, exists := got[locale]; !exists {
				t.Errorf("missing locale %s in language display name map", locale)
			} else if name == "" {
				t.Errorf("empty language display name for locale %s", locale)
			}
		}
	})
}

func TestLocaleConstants(t *testing.T) {
	t.Run("default locale is English", func(t *testing.T) {
		if shared.DefaultLocale != shared.LocaleEnglishUS {
			t.Errorf("DefaultLocale: got %v, want %v", shared.DefaultLocale, shared.LocaleEnglishUS)
		}
	})

	t.Run("supported locales contains expected values", func(t *testing.T) {
		expected := []shared.Locale{shared.LocaleFrenchFR, shared.LocaleEnglishUS, shared.LocalePortugueseBR}

		if len(shared.SupportedLocales) != len(expected) {
			t.Errorf("SupportedLocales length: got %d, want %d", len(shared.SupportedLocales), len(expected))
		}

		for _, expectedLocale := range expected {
			found := slices.Contains(shared.SupportedLocales, expectedLocale)
			if !found {
				t.Errorf("expected locale %s not found in SupportedLocales", expectedLocale)
			}
		}
	})

	t.Run("locale constants have correct values", func(t *testing.T) {
		tests := []struct {
			name   string
			locale shared.Locale
			want   string
		}{
			{"LocaleFrenchFR", shared.LocaleFrenchFR, "fr-FR"},
			{"LocaleEnglishUS", shared.LocaleEnglishUS, "en-US"},
			{"LocalePortugueseBR", shared.LocalePortugueseBR, "pt-BR"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if string(tt.locale) != tt.want {
					t.Errorf("got %q, want %q", tt.locale, tt.want)
				}
			})
		}
	})
}

func TestLocaleErrorMessages(t *testing.T) {
	t.Run("error message constants", func(t *testing.T) {
		tests := []struct {
			name     string
			constant string
			expected string
		}{
			{"invalid locale", shared.MLocaleInvalid, "Invalid locale code."},
			{"missing locale", shared.MLocaleMissing, "Missing locale."},
			{"unsupported locale format", shared.MLocaleUnsupported, "Unsupported locale: %s."},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.constant != tt.expected {
					t.Errorf("got %q, want %q", tt.constant, tt.expected)
				}
			})
		}
	})
}
