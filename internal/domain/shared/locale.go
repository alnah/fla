package shared

import (
	"fmt"
	"slices"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/language/display"

	"github.com/alnah/fla/internal/domain/kernel"
)

const (
	MLocaleInvalid     string = "Invalid locale code."
	MLocaleMissing     string = "Missing locale."
	MLocaleUnsupported string = "Unsupported locale: %s."
)

// Locale represents a language/region combination following IETF BCP 47 (e.g. fr-FR).
type Locale string

const (
	LocaleFrenchFR     Locale = "fr-FR" // French (France)
	LocaleEnglishUS    Locale = "en-US" // English (United States)
	LocalePortugueseBR Locale = "pt-BR" // Portuguese (Brazil)
)

// DefaultLocale is the fallback when no locale is specified
const DefaultLocale = LocaleEnglishUS

// SupportedLocales lists all supported interface languages
var SupportedLocales = []Locale{
	LocaleFrenchFR,
	LocaleEnglishUS,
	LocalePortugueseBR,
}

// NewLocale creates a validated locale with support checking.
// Ensures only supported languages are used in the application.
func NewLocale(locale string) (Locale, error) {
	const op = "NewLocale"

	if locale == "" {
		return DefaultLocale, nil
	}

	l := Locale(strings.TrimSpace(locale))
	if err := l.Validate(); err != nil {
		return "", &kernel.Error{Operation: op, Cause: err}
	}

	return l, nil
}

func (l Locale) String() string { return string(l) }

// Validate ensures locale is supported by the application.
// Prevents unsupported language codes that would break i18n.
func (l Locale) Validate() error {
	const op = "Locale.Validate"

	if err := kernel.ValidatePresence("locale", l.String(), op); err != nil {
		return err
	}

	if err := l.validateBCP47Format(); err != nil {
		return err
	}

	if err := l.validateSupported(); err != nil {
		return err
	}

	return nil
}

// validateBCP47Format checks if locale is a valid BCP 47 tag.
func (l Locale) validateBCP47Format() error {
	const op = "locale.validateBCP47Format"

	if _, err := language.Parse(string(l)); err != nil {
		return &kernel.Error{
			Code:      kernel.EInvalid,
			Message:   MLocaleInvalid,
			Operation: op,
			Cause:     err,
		}
	}
	return nil
}

// validateSupported checks if locale is supported by the application.
func (l Locale) validateSupported() error {
	const op = "locale.validateSupported"

	if !l.IsSupported() {
		return &kernel.Error{
			Code:      kernel.EInvalid,
			Message:   fmt.Sprintf(MLocaleUnsupported, l),
			Operation: op,
		}
	}
	return nil
}

// IsSupported checks if this locale is supported by the application.
func (l Locale) IsSupported() bool {
	return slices.Contains(SupportedLocales, l)
}

// GetEffectiveLocale returns the locale to use, with default fallback.
func (l Locale) GetEffectiveLocale() Locale {
	if l == "" || !l.IsSupported() {
		return DefaultLocale
	}
	return l
}

// IsDefault returns true if this is the default locale.
func (l Locale) IsDefault() bool {
	return l == DefaultLocale
}

// ToLanguageTag converts the locale to a language.Tag for use with golang.org/x/text.
func (l Locale) ToLanguageTag() (language.Tag, error) {
	return language.Parse(string(l.GetEffectiveLocale()))
}

// ToISO639Language extracts the ISO 639-1 language code from the BCP 47 locale.
// Returns "fr" from "fr-FR", "en" from "en-US", "pt" from "pt-BR".
// Useful for legacy systems that only support ISO 639-1.
func (l Locale) ToISO639Language() string {
	tag, err := l.ToLanguageTag()
	if err != nil {
		return "fr" // fallback to default language
	}

	base, _ := tag.Base()
	return base.String()
}

// GetRegion extracts the region code from the BCP 47 locale.
// Returns "FR" from "fr-FR", "US" from "en-US", "BR" from "pt-BR".
// Useful for region-specific formatting (dates, numbers, currency).
func (l Locale) GetRegion() string {
	tag, err := l.ToLanguageTag()
	if err != nil {
		return ""
	}

	region, _ := tag.Region()
	return region.String()
}

// GetDisplayName returns the locale's display name in the specified display language.
// If displayIn is empty, returns the name in the locale's own language.
func (l Locale) GetDisplayName(displayIn Locale) string {
	targetTag, err := l.ToLanguageTag()
	if err != nil {
		return string(l) // fallback to raw string
	}

	var namer display.Namer
	if displayIn == "" {
		// Display in the locale's own language (self-display)
		namer = display.Self
	} else {
		displayTag, err := displayIn.ToLanguageTag()
		if err != nil {
			namer = display.Self // fallback to self-display
		} else {
			namer = display.Languages(displayTag)
		}
	}

	return namer.Name(targetTag)
}

// GetSelfDisplayName returns the locale's display name in its own language.
// Example: "français (France)", "English (United States)", "português (Brasil)"
func (l Locale) GetSelfDisplayName() string {
	return l.GetDisplayName("")
}

// GetLanguageDisplayName returns only the language part in its own language.
// Example: "français", "English", "português" (without region)
func (l Locale) GetLanguageDisplayName() string {
	tag, err := l.ToLanguageTag()
	if err != nil {
		return string(l)
	}

	// Create language-only tag (without region)
	langOnlyTag, _ := tag.Base()
	langTag := language.Make(langOnlyTag.String())

	namer := display.Self
	return namer.Name(langTag)
}

// GetEnglishDisplayName returns the locale's display name in English.
func (l Locale) GetEnglishDisplayName() string {
	return l.GetDisplayName(LocaleEnglishUS)
}

// GetDisplayNameMap returns a map of all supported locales with their display names
// in the specified display language.
func GetDisplayNameMap(displayIn Locale) map[Locale]string {
	result := make(map[Locale]string)
	for _, locale := range SupportedLocales {
		result[locale] = locale.GetDisplayName(displayIn)
	}
	return result
}

// GetAllSelfDisplayNames returns a map of all supported locales with their
// self-display names (each locale displayed in its own language).
func GetAllSelfDisplayNames() map[Locale]string {
	result := make(map[Locale]string)
	for _, locale := range SupportedLocales {
		result[locale] = locale.GetSelfDisplayName()
	}
	return result
}

// GetLanguageOnlyDisplayNames returns language names without region info.
func GetLanguageOnlyDisplayNames() map[Locale]string {
	result := make(map[Locale]string)
	for _, locale := range SupportedLocales {
		result[locale] = locale.GetLanguageDisplayName()
	}
	return result
}
