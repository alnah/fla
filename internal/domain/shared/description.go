package shared

import (
	"strings"

	"github.com/alnah/fla/internal/domain/kernel"
)

const (
	MinDescriptionLength int = 0 // Optional field
	MaxDescriptionLength int = 300
)

// Description provides explanatory text for entities with length constraints.
// Enables rich metadata while maintaining UI display boundaries.
type Description string

// NewDescription creates validated descriptive text with length checking.
// Ensures descriptions fit within meta tag limits and UI components.
func NewDescription(desc string) (Description, error) {
	const op = "NewDescription"

	d := Description(strings.TrimSpace(desc))
	if err := d.Validate(); err != nil {
		return "", &kernel.Error{Operation: op, Cause: err}
	}

	return d, nil
}

func (d Description) String() string { return string(d) }

// Validate ensures description length stays within practical display limits.
// Prevents overly long descriptions that break UI layouts and meta tags.
func (d Description) Validate() error {
	const op = "Description.Validate"

	if err := kernel.ValidateLength("description", d.String(), MinDescriptionLength, MaxDescriptionLength, op); err != nil {
		return err
	}

	return nil
}
