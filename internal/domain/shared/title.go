package shared

import (
	"strings"

	"github.com/alnah/fla/internal/domain/kernel"
)

const (
	MinTitleLength int = 10
	MaxTitleLength int = 100
)

// Title represents content headlines with length validation for readability.
// Ensures titles are descriptive enough while maintaining display compatibility.
type Title string

// NewTitle creates validated content title with readability requirements.
// Balances descriptive content with practical length constraints for SEO and UI.
func NewTitle(title string) (Title, error) {
	const op = "NewTitle"

	t := Title(strings.TrimSpace(title))
	if err := t.Validate(); err != nil {
		return "", &kernel.Error{Operation: op, Cause: err}
	}

	return t, nil
}

func (t Title) String() string { return string(t) }

// Validate ensures title meets editorial standards for effective communication.
// Enforces minimum descriptiveness while preventing overly long headlines.
func (t Title) Validate() error {
	const op = "Title.Validate"

	if err := kernel.ValidatePresence("title", t.String(), op); err != nil {
		return err
	}

	if err := kernel.ValidateLength("title", t.String(), MinTitleLength, MaxTitleLength, op); err != nil {
		return err
	}

	return nil
}
