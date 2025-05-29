package shared

import (
	"regexp"
	"strings"

	"github.com/alnah/fla/internal/domain/kernel"
)

const (
	MinFirstNameLength = 0 // Optional field
	MaxFirstNameLength = 50
	MinLastNameLength  = 0 // Optional field
	MaxLastNameLength  = 50
	MaxUsernameLength  = 30
	MinUserNameLength  = 3
)

const (
	MUsernameInvalidChars string = "Username can only contain letters, numbers, underscores, and hyphens."
)

// FirstName represents personal given name with optional validation.
// Enables personalization while accommodating diverse naming conventions.
type FirstName string

// NewFirstName creates validated first name with cultural sensitivity.
// Supports optional names while preventing excessively long values.
func NewFirstName(name string) (FirstName, error) {
	const op = "NewFirstName"

	f := FirstName(strings.TrimSpace(name))
	if err := f.Validate(); err != nil {
		return "", &kernel.Error{Operation: op, Cause: err}
	}

	return f, nil
}

func (f FirstName) String() string { return string(f) }

// Validate ensures first name meets display and storage requirements.
// Accommodates cultural naming patterns while preventing UI issues.
func (f FirstName) Validate() error {
	const op = "FirstName.Validate"

	if err := kernel.ValidateLength("first name", f.String(), MinFirstNameLength, MaxFirstNameLength, op); err != nil {
		return err
	}

	return nil
}

// LastName represents family name with flexible validation requirements.
// Supports diverse naming traditions while maintaining data consistency.
type LastName string

// NewLastName creates validated family name with cultural accommodation.
// Balances name validation with international naming flexibility.
func NewLastName(name string) (LastName, error) {
	const op = "NewLastName"

	l := LastName(strings.TrimSpace(name))
	if err := l.Validate(); err != nil {
		return "", &kernel.Error{Operation: op, Cause: err}
	}

	return l, nil
}

func (l LastName) String() string { return string(l) }

// Validate ensures last name fits within system constraints and display limits.
// Prevents database overflow while respecting naming diversity.
func (l LastName) Validate() error {
	const op = "LastName.Validate"

	if err := kernel.ValidateLength("last name", l.String(), MinLastNameLength, MaxLastNameLength, op); err != nil {
		return err
	}

	return nil
}

// Username represents unique user identifiers for login and public display.
// Ensures usernames are URL-safe and meet platform conventions.
type Username string

// NewUsername creates validated username with character and length restrictions.
// Prevents conflicts and ensures usernames work across web systems.
func NewUsername(username string) (Username, error) {
	const op = "NewUsername"

	u := Username(strings.TrimSpace(username))
	if err := u.Validate(); err != nil {
		return "", &kernel.Error{Operation: op, Cause: err}
	}

	return u, nil
}

func (u Username) String() string { return string(u) }

// Validate ensures username meets platform standards for identification.
// Balances user choice with technical requirements and platform conventions.
func (u Username) Validate() error {
	const op = "Username.Validate"

	if err := kernel.ValidatePresence("username", u.String(), op); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := kernel.ValidateLength("username", u.String(), MinUserNameLength, MaxUsernameLength, op); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := u.validateCharacters(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	return nil
}

func (u Username) validateCharacters() error {
	const op = "Username.validateCharacters"

	pattern := `^[a-zA-Z0-9_-]+$`
	matched, _ := regexp.MatchString(pattern, u.String())
	if !matched {
		return &kernel.Error{
			Code:      kernel.EInvalid,
			Message:   MUsernameInvalidChars,
			Operation: op,
		}
	}

	return nil
}
