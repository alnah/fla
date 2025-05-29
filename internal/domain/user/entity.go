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

	// Data
	FirstName      shared.FirstName
	LastName       shared.LastName
	Description    shared.Description
	PictureURL     kernel.URL[ProfilePicture]
	SocialProfiles []SocialProfile

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

	// Optional
	FirstName      shared.FirstName
	LastName       shared.LastName
	Description    shared.Description
	PictureURL     kernel.URL[ProfilePicture]
	SocialProfiles []SocialProfile

	// DI
	Clock kernel.Clock
}

// NewUser creates a validated user account with proper role assignment.
// Ensures user data integrity and permission system initialization.
func NewUser(p NewUserParams) (User, error) {
	const op = "NewUser"

	now := p.Clock.Now()

	user := User{
		ID:             p.UserID,
		Username:       p.Username,
		Email:          p.Email,
		FirstName:      p.FirstName,
		LastName:       p.LastName,
		Description:    p.Description,
		PictureURL:     p.PictureURL,
		SocialProfiles: p.SocialProfiles,
		Roles:          p.Roles,
		CreatedAt:      now,
		UpdatedAt:      now,
		Clock:          p.Clock,
	}

	if err := user.Validate(); err != nil {
		return User{}, &kernel.Error{Operation: op, Cause: err}
	}

	return user, nil
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
