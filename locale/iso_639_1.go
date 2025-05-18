package locale

import (
	"fmt"

	"golang.org/x/text/language"
)

// IETF represents a full IETF BCP 47 language tag (e.g. "en-US").
type IETF string

// String returns the IETF tag as a plain string.
func (t IETF) String() string { return string(t) }

// IsValid reports whether the IETF tag parses without error under BCP 47.
func (t IETF) IsValid() bool {
	_, err := language.Parse(t.String())
	return err == nil
}

// Validate returns an error if the IETF tag is not a valid BCP 47 tag.
func (t IETF) Validate() error {
	if !t.IsValid() {
		return fmt.Errorf("invalid IETF tag: %q", t.String())
	}
	return nil
}
