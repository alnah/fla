package tag

import (
	"strings"
	"time"

	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/user"
)

const (
	MinTagNameLength int = 1
	MaxTagNameLength int = 50
)

const (
	MTagNameMissing string = "Missing tag name."
)

// TagName represents descriptive labels for content discovery and organization.
// Enables flexible content categorization beyond hierarchical structure.
type TagName string

// NewTagName creates validated tag label with appropriate length constraints.
// Ensures tags are meaningful while fitting within UI and database limits.
func NewTagName(name string) (TagName, error) {
	const op = "NewTagName"

	t := TagName(strings.TrimSpace(name))
	if err := t.Validate(); err != nil {
		return "", &kernel.Error{Operation: op, Cause: err}
	}

	return t, nil
}

func (t TagName) String() string { return string(t) }

// Validate ensures tag name meets requirements for effective content labeling.
// Balances descriptive value with practical display constraints.
func (t TagName) Validate() error {
	const op = "TagName.Validate"

	if err := kernel.ValidatePresence("tag name", t.String(), op); err != nil {
		return err
	}

	if err := kernel.ValidateLength("tag name", t.String(), MinTagNameLength, MaxTagNameLength, op); err != nil {
		return err
	}

	return nil
}

// Tag represents a descriptive label for categorizing and discovering blog content.
// Tags enable cross-cutting content organization beyond hierarchical categories.
type Tag struct {
	// Identity
	TagID kernel.ID[Tag]

	// Data
	Name TagName

	// Meta
	CreatedBy kernel.ID[user.User]
	CreatedAt time.Time
}

// NewTag creates a validated tag with proper metadata tracking.
// Ensures tag consistency and audit trail for content organization.
func NewTag(t Tag) (Tag, error) {
	const op = "NewTag"

	if err := t.Validate(); err != nil {
		return Tag{}, &kernel.Error{Operation: op, Cause: err}
	}

	return t, nil
}

// Validate enforces business rules for tag data integrity and consistency.
// Prevents invalid tags that would compromise content organization.
func (t Tag) Validate() error {
	const op = "Tag.Validate"

	if err := t.TagID.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := t.Name.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := t.CreatedBy.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	return nil
}
