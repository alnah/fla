package locale

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

// Lang is a restricted enumeration of supported IETF tags.
type Lang string

// String returns the Lang value as a plain string.
func (l Lang) String() string { return string(l) }

// Supported language constants.
const (
	// LangFrFR stands for French as used in France.
	LangFrFR Lang = "fr-FR"
	// LangPtBR stands for Portuguese as used in Brazil.
	LangPtBR Lang = "pt-BR"
	// LangEnUS stands for English as used in the United States.
	LangEnUS Lang = "en-US"
)

var langPattern = regexp.MustCompile(`^[a-z]{2}-[A-Z]{2}$`)

// IsValid reports whether the Lang value matches the pattern xx-XX
// and is one of the supported constants.
func (l Lang) IsValid() bool {
	if !langPattern.MatchString(l.String()) {
		return false
	}
	switch l {
	case LangFrFR, LangPtBR, LangEnUS:
		return true
	}
	return false
}

// Validate returns an error if the Lang value is not supported.
func (l Lang) Validate() error {
	if !l.IsValid() {
		return fmt.Errorf(
			"unsupported language: %s (must be one of %s, %s, %s)",
			l.String(), LangFrFR, LangPtBR, LangEnUS,
		)
	}
	return nil
}

// Set implements flag.Value (and pflag.Value) for overriding env vars, or binding flags.
func (l *Lang) Set(s string) error {
	candidate := Lang(s)
	if err := candidate.Validate(); err != nil {
		return fmt.Errorf("invalid language %s: %w", s, err)
	}
	*l = candidate
	return nil
}

// Type returns a descriptive name of this flag’s type (used by pflag).
func (l *Lang) Type() string {
	return "Lang"
}

// ToISO6391 extracts the two‐letter ISO 639-1 code (lowercase) from the Lang tag.
func (l Lang) ToISO6391() ISO6391 {
	parts := strings.SplitN(l.String(), "-", 2)
	return ISO6391(strings.ToLower(parts[0]))
}

// ToIETF converts the Lang value into a generic IETF tag.
func (l Lang) ToIETF() IETF {
	return IETF(l.String())
}

var (
	isoToLang = map[ISO6391]Lang{
		"fr": LangFrFR,
		"pt": LangPtBR,
		"en": LangEnUS,
	}
	ietfToLang = map[IETF]Lang{
		LangFrFR.ToIETF(): LangFrFR,
		LangPtBR.ToIETF(): LangPtBR,
		LangEnUS.ToIETF(): LangEnUS,
	}
)

// FromISO6391 converts a valid ISO 639-1 code to one of the supported Lang tags.
// Returns an error if the code is invalid or not among the supported languages.
func FromISO6391(i ISO6391) (Lang, error) {
	if err := i.Validate(); err != nil {
		return "", err
	}
	if l, ok := isoToLang[i]; ok {
		return l, nil
	}
	return "", fmt.Errorf("no language mapping for ISO6391 %s", i)
}

// FromIETF converts a valid IETF tag to one of the supported Lang tags.
// Returns an error if the tag is invalid or not among the supported languages.
func FromIETF(t IETF) (Lang, error) {
	if err := t.Validate(); err != nil {
		return "", err
	}
	if l, ok := ietfToLang[t]; ok {
		return l, nil
	}
	return "", fmt.Errorf("no language mapping for IETF tag %s", t)
}

// DisplayName returns the human‐readable name of the Lang value
// localized into the given display language.
// For example, LangFrFR.DisplayName(LangEnUS) == "French (France)".
func (l Lang) DisplayName(displayLang Lang) string {
	tag := language.Make(string(l))
	dl := language.Make(string(displayLang))
	return display.Tags(dl).Name(tag)
}

// EnglishName returns the name of the language in English.
func (l Lang) EnglishName() string {
	return l.DisplayName(LangEnUS)
}

// FrenchName returns the name of the language in French.
func (l Lang) FrenchName() string {
	return l.DisplayName(LangFrFR)
}

// PortugueseName returns the name of the language in Portuguese (Brazil).
func (l Lang) PortugueseName() string {
	return l.DisplayName(LangPtBR)
}

// ParseLang converts s into a Lang by normalizing case, checking
// the xx-XX pattern, and validating against the supported set.
func ParseLang(s string) (Lang, error) {
	// normalize separators and case:
	t := strings.TrimSpace(strings.ReplaceAll(s, "_", "-"))
	parts := strings.SplitN(t, "-", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid language tag %s", s)
	}
	languageCode := strings.ToLower(parts[0])
	regionCode := strings.ToUpper(parts[1])
	candidate := Lang(languageCode + "-" + regionCode)

	// validate against your supported constants
	if err := candidate.Validate(); err != nil {
		return "", fmt.Errorf("unsupported language %s: %w", candidate, err)
	}

	return candidate, nil
}
