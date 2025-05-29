package shared

import (
	"regexp"
	"strings"

	"github.com/alnah/fla/internal/domain/kernel"
)

const (
	MEmailInvalid       string = "Invalid email."
	MEmailMissing       string = "Missing email."
	MEmailFormatInvalid string = "Invalid email format."
)

// Email represents validated email addresses for user communication.
// Ensures deliverable addresses for notifications and account management.
type Email string

// NewEmail creates validated email address with format verification.
// Prevents invalid addresses that would cause delivery failures.
func NewEmail(email string) (Email, error) {
	const op = "NewEmail"

	e := Email(strings.TrimSpace(email))
	if err := e.Validate(); err != nil {
		return "", &kernel.Error{Operation: op, Cause: err}
	}

	return e, nil
}

func (e Email) String() string { return string(e) }

// Validate ensures email meets RFC standards for reliable delivery.
// Prevents communication failures due to malformed addresses.
func (e Email) Validate() error {
	const op = "Email.Validate"

	if err := e.validatePresence(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := e.validateFormat(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	return nil
}

func (e Email) validatePresence() error {
	const op = "Email.validatePresence"

	if e.String() == "" {
		return &kernel.Error{
			Code:      kernel.EInvalid,
			Message:   MEmailMissing,
			Operation: op,
		}
	}

	return nil
}

// Alternative: Using a more comprehensive but still maintainable regex
func (e Email) validateFormat() error {
	const op = "Email.validateFormat"

	// This pattern is more comprehensive while still being readable
	// It handles most common email formats according to RFC 5322
	pattern := `^[a-zA-Z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+(?:\.[a-zA-Z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+)*@(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?$`

	matched, err := regexp.MatchString(pattern, e.String())
	if err != nil {
		return &kernel.Error{
			Code:      kernel.EInternal,
			Message:   "Failed to match email pattern",
			Operation: op,
			Cause:     err,
		}
	}

	if !matched {
		return &kernel.Error{
			Code:      kernel.EInvalid,
			Message:   MEmailFormatInvalid,
			Operation: op,
		}
	}

	return nil
}
