// ./internal/domain/user/entity.go
package user

import (
	"fmt"
	"slices"
	"time"

	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/shared"
)

const (
	MUserRoleMissing          string = "Missing roles. One role should be set."
	MUserInvalidRole          string = "Invalid role: %q."
	MUserInvalidSocialProfile string = "Invalid social profile: %+v."
	MUserDuplicateSocialMedia string = "Duplicate social media platform: %q."
)

// User represents an authenticated person with role-based permissions in the blogging system.
// Manages identity, profile information, and content access controls.
type User struct {
	// Identity
	ID       kernel.ID[User]
	Username shared.Username
	Email    shared.Email

	// Permissions
	Roles []Role

	// Profile Data
	FirstName      shared.FirstName
	LastName       shared.LastName
	Description    shared.Description
	PictureURL     kernel.URL[ProfilePicture]
	SocialProfiles []SocialProfile

	// Preferences
	LocalePreference shared.Locale // User's preferred interface language

	// Meta
	CreatedAt time.Time
	UpdatedAt time.Time

	// DI
	Clock kernel.Clock
}

// NewUserParams holds essential information for creating user accounts.
// Separates required identity fields from optional profile information.
type NewUserParams struct {
	// Required
	UserID   kernel.ID[User]
	Username shared.Username
	Email    shared.Email
	Roles    []Role

	// Optional Profile
	FirstName      shared.FirstName
	LastName       shared.LastName
	Description    shared.Description
	PictureURL     kernel.URL[ProfilePicture]
	SocialProfiles []SocialProfile

	// Optional Preferences
	LocalePreference shared.Locale // Defaults to system default if not provided

	// DI
	Clock kernel.Clock
}

// NewUser creates a validated user account with proper role assignment.
// Ensures user data integrity and permission system initialization.
func NewUser(p NewUserParams) (User, error) {
	const op = "NewUser"

	now := p.Clock.Now()

	// Use default locale if not specified
	locale := p.LocalePreference
	if locale == "" {
		locale = shared.DefaultLocale
	}

	user := User{
		ID:               p.UserID,
		Username:         p.Username,
		Email:            p.Email,
		FirstName:        p.FirstName,
		LastName:         p.LastName,
		Description:      p.Description,
		PictureURL:       p.PictureURL,
		SocialProfiles:   p.SocialProfiles,
		LocalePreference: locale,
		Roles:            p.Roles,
		CreatedAt:        now,
		UpdatedAt:        now,
		Clock:            p.Clock,
	}

	if err := user.Validate(); err != nil {
		return User{}, &kernel.Error{Operation: op, Cause: err}
	}

	return user, nil
}

// UpdateLocalePreference allows users to change their language preference
func (u User) UpdateLocalePreference(newLocale shared.Locale) (User, error) {
	const op = "User.UpdateLocalePreference"

	// Validate the new locale
	if err := newLocale.Validate(); err != nil {
		return u, &kernel.Error{Operation: op, Cause: err}
	}

	updated := u
	updated.LocalePreference = newLocale
	updated.UpdatedAt = u.Clock.Now()

	return updated, nil
}

// GetDisplayName returns the user's display name in their preferred language
func (u User) GetDisplayName() string {
	// Prioritize first name, then username, then email
	if u.FirstName.String() != "" {
		return u.FirstName.String()
	}
	if u.Username.String() != "" {
		return u.Username.String()
	}
	return u.Email.String()
}

// GetFullName returns the user's full name if available
func (u User) GetFullName() string {
	firstName := u.FirstName.String()
	lastName := u.LastName.String()

	if firstName != "" && lastName != "" {
		return firstName + " " + lastName
	} else if firstName != "" {
		return firstName
	} else if lastName != "" {
		return lastName
	}

	return ""
}

// String provides detailed user representation for debugging and logging.
// Truncates sensitive information while preserving diagnostic value.
func (u User) String() string {
	const truncateMaxLength = 50

	description := u.Description.String()
	if len(description) > truncateMaxLength {
		description = description[:truncateMaxLength] + "..."
	}

	return fmt.Sprintf("User{"+
		"UserID: %q, "+
		"Username: %q, "+
		"Email: %q, "+
		"FirstName: %q, "+
		"LastName: %q, "+
		"Description: %q, "+
		"PictureURL: %q, "+
		"SocialProfiles: %+v, "+
		"LocalePreference: %q, "+
		"Roles: %+v, "+
		"CreatedAt: %s, "+
		"UpdatedAt: %s"+
		"}",
		u.ID,
		u.Username,
		u.Email,
		u.FirstName,
		u.LastName,
		description,
		u.PictureURL,
		u.SocialProfiles,
		u.LocalePreference,
		u.Roles,
		u.CreatedAt.Format(time.RFC3339),
		u.UpdatedAt.Format(time.RFC3339),
	)
}

// Validate ensures user data meets all business requirements and constraints.
// Prevents creation of users that would violate system integrity or security.
func (u User) Validate() error {
	const op = "User.Validate"

	if err := u.validateIdentity(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := u.validateOptionalProfile(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := u.validatePreferences(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := u.validateRoles(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := u.validateSocialProfiles(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	return nil
}

func (u User) validateIdentity() error {
	const op = "User.validateIdentity"

	if err := u.ID.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := u.Username.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := u.Email.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	return nil
}

func (u User) validateOptionalProfile() error {
	const op = "User.validateOptionalProfile"

	if err := u.FirstName.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := u.LastName.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := u.Description.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := u.PictureURL.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	return nil
}

func (u User) validatePreferences() error {
	const op = "User.validatePreferences"

	if err := u.LocalePreference.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	return nil
}

func (u User) validateRoles() error {
	const op = "User.validateRoles"

	if len(u.Roles) == 0 {
		return &kernel.Error{Code: kernel.EInvalid, Message: MUserRoleMissing, Operation: op}
	}

	for _, role := range u.Roles {
		if err := role.Validate(); err != nil {
			return &kernel.Error{
				Code:      kernel.EInvalid,
				Message:   fmt.Sprintf(MUserInvalidRole, role),
				Operation: op,
				Cause:     err,
			}
		}
	}

	return nil
}

func (u User) validateSocialProfiles() error {
	const op = "User.validateSocialProfiles"

	for _, profile := range u.SocialProfiles {
		if err := profile.Validate(); err != nil {
			return &kernel.Error{
				Code:      kernel.EInvalid,
				Message:   fmt.Sprintf(MUserInvalidSocialProfile, profile),
				Operation: op,
				Cause:     err,
			}
		}
	}

	if err := u.validateUniqueSocialPlatforms(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	return nil
}

func (u User) validateUniqueSocialPlatforms() error {
	const op = "User.validateUniqueSocialPlatforms"

	platformCount := make(map[SocialMediaURL]int)
	for _, profile := range u.SocialProfiles {
		platformCount[profile.Platform]++
		if platformCount[profile.Platform] > 1 {
			return &kernel.Error{
				Code:      kernel.EInvalid,
				Message:   fmt.Sprintf(MUserDuplicateSocialMedia, profile.Platform),
				Operation: op,
			}
		}
	}

	return nil
}

// HasRole checks if user has a specific role.
func (u User) HasRole(role Role) bool {
	return slices.Contains(u.Roles, role)
}

// HasAnyRole checks if user has any of the specified roles.
func (u User) HasAnyRole(roles ...Role) bool {
	return slices.ContainsFunc(roles, u.HasRole)
}
