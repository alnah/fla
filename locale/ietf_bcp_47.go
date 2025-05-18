package locale

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/text/language"
)

// ISO6391 represents a two‐letter ISO 639-1 language code (e.g. "en").
type ISO6391 string

// String returns the ISO 639-1 code as a plain string.
func (i ISO6391) String() string { return string(i) }

// IsValid reports whether the ISO 639-1 code is syntactically correct
// (exactly two ASCII letters) and corresponds to a known language.
func (i ISO6391) IsValid() bool {
	pat := regexp.MustCompile(`^[A-Za-z]{2}$`)
	s := strings.ToLower(i.String())
	if !pat.MatchString(s) {
		return false
	}
	tag := language.Make(s)
	return !tag.IsRoot()
}

// Validate returns an error if the ISO 639-1 code is not valid.
func (i ISO6391) Validate() error {
	if !i.IsValid() {
		return fmt.Errorf("invalid ISO 639-1 code: %q", i.String())
	}
	return nil
}
