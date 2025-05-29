package user_test

import (
	"strings"
	"testing"
	"time"

	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/shared"
	"github.com/alnah/fla/internal/domain/user"
)

type stubClock struct {
	t time.Time
}

func (s *stubClock) Now() time.Time { return s.t }

func TestNewUser(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	clock := &stubClock{t: fixedTime}

	validUserID, _ := kernel.NewID[user.User]("user-123")
	validUsername, _ := shared.NewUsername("johndoe")
	validEmail, _ := shared.NewEmail("john@example.com")
	validFirstName, _ := shared.NewFirstName("John")
	validLastName, _ := shared.NewLastName("Doe")
	validDescription, _ := shared.NewDescription("Software developer")
	validPictureURL, _ := kernel.NewURL[user.ProfilePicture]("https://example.com/pic.jpg")
	validSocialProfile, _ := user.NewSocialProfile(user.SocialMediaGitHub, "https://github.com/johndoe")

	t.Run("creates user with minimal required fields", func(t *testing.T) {
		params := user.NewUserParams{
			UserID:   validUserID,
			Username: validUsername,
			Email:    validEmail,
			Roles:    []user.Role{user.RoleAuthor},
			Clock:    clock,
		}

		got, err := user.NewUser(params)

		assertNoError(t, err)

		if got.ID != validUserID {
			t.Errorf("ID: got %v, want %v", got.ID, validUserID)
		}
		if got.Username != validUsername {
			t.Errorf("Username: got %v, want %v", got.Username, validUsername)
		}
		if got.Email != validEmail {
			t.Errorf("Email: got %v, want %v", got.Email, validEmail)
		}
		if len(got.Roles) != 1 || got.Roles[0] != user.RoleAuthor {
			t.Errorf("Roles: got %v, want [%v]", got.Roles, user.RoleAuthor)
		}
		if !got.CreatedAt.Equal(fixedTime) {
			t.Errorf("CreatedAt: got %v, want %v", got.CreatedAt, fixedTime)
		}
		if !got.UpdatedAt.Equal(fixedTime) {
			t.Errorf("UpdatedAt: got %v, want %v", got.UpdatedAt, fixedTime)
		}
	})

	t.Run("creates user with all fields", func(t *testing.T) {
		params := user.NewUserParams{
			UserID:         validUserID,
			Username:       validUsername,
			Email:          validEmail,
			Roles:          []user.Role{user.RoleAdmin, user.RoleEditor},
			FirstName:      validFirstName,
			LastName:       validLastName,
			Description:    validDescription,
			PictureURL:     validPictureURL,
			SocialProfiles: []user.SocialProfile{validSocialProfile},
			Clock:          clock,
		}

		got, err := user.NewUser(params)

		assertNoError(t, err)

		if got.FirstName != validFirstName {
			t.Errorf("FirstName: got %v, want %v", got.FirstName, validFirstName)
		}
		if got.LastName != validLastName {
			t.Errorf("LastName: got %v, want %v", got.LastName, validLastName)
		}
		if got.Description != validDescription {
			t.Errorf("Description: got %v, want %v", got.Description, validDescription)
		}
		if got.PictureURL != validPictureURL {
			t.Errorf("PictureURL: got %v, want %v", got.PictureURL, validPictureURL)
		}
		if len(got.SocialProfiles) != 1 {
			t.Fatalf("SocialProfiles: got %d profiles, want 1", len(got.SocialProfiles))
		}
		if got.SocialProfiles[0].Platform != user.SocialMediaGitHub {
			t.Errorf("SocialProfile platform: got %v, want %v",
				got.SocialProfiles[0].Platform, user.SocialMediaGitHub)
		}
	})

	t.Run("rejects invalid parameters", func(t *testing.T) {
		tests := []struct {
			name   string
			params user.NewUserParams
		}{
			{
				name: "empty user ID",
				params: user.NewUserParams{
					UserID:   kernel.ID[user.User](""),
					Username: validUsername,
					Email:    validEmail,
					Roles:    []user.Role{user.RoleAuthor},
					Clock:    clock,
				},
			},
			{
				name: "empty username",
				params: user.NewUserParams{
					UserID:   validUserID,
					Username: shared.Username(""),
					Email:    validEmail,
					Roles:    []user.Role{user.RoleAuthor},
					Clock:    clock,
				},
			},
			{
				name: "empty email",
				params: user.NewUserParams{
					UserID:   validUserID,
					Username: validUsername,
					Email:    shared.Email(""),
					Roles:    []user.Role{user.RoleAuthor},
					Clock:    clock,
				},
			},
			{
				name: "no roles",
				params: user.NewUserParams{
					UserID:   validUserID,
					Username: validUsername,
					Email:    validEmail,
					Roles:    []user.Role{},
					Clock:    clock,
				},
			},
			{
				name: "invalid role",
				params: user.NewUserParams{
					UserID:   validUserID,
					Username: validUsername,
					Email:    validEmail,
					Roles:    []user.Role{"invalid-role"},
					Clock:    clock,
				},
			},
			{
				name: "duplicate social media platforms",
				params: user.NewUserParams{
					UserID:   validUserID,
					Username: validUsername,
					Email:    validEmail,
					Roles:    []user.Role{user.RoleAuthor},
					SocialProfiles: []user.SocialProfile{
						{Platform: user.SocialMediaTwitter, URL: "https://twitter.com/user1"},
						{Platform: user.SocialMediaTwitter, URL: "https://twitter.com/user2"},
					},
					Clock: clock,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := user.NewUser(tt.params)

				assertError(t, err)
				assertErrorCode(t, err, kernel.EInvalid)
			})
		}
	})

	t.Run("accepts multiple roles", func(t *testing.T) {
		params := user.NewUserParams{
			UserID:   validUserID,
			Username: validUsername,
			Email:    validEmail,
			Roles:    []user.Role{user.RoleAdmin, user.RoleEditor, user.RoleAuthor},
			Clock:    clock,
		}

		got, err := user.NewUser(params)

		assertNoError(t, err)
		if len(got.Roles) != 3 {
			t.Errorf("expected 3 roles, got %d", len(got.Roles))
		}
	})

	t.Run("accepts multiple social profiles", func(t *testing.T) {
		twitter, _ := user.NewSocialProfile(user.SocialMediaTwitter, "https://twitter.com/user")
		github, _ := user.NewSocialProfile(user.SocialMediaGitHub, "https://github.com/user")
		linkedin, _ := user.NewSocialProfile(user.SocialMediaLinkedIn, "https://linkedin.com/in/user")

		params := user.NewUserParams{
			UserID:         validUserID,
			Username:       validUsername,
			Email:          validEmail,
			Roles:          []user.Role{user.RoleAuthor},
			SocialProfiles: []user.SocialProfile{twitter, github, linkedin},
			Clock:          clock,
		}

		got, err := user.NewUser(params)

		assertNoError(t, err)
		if len(got.SocialProfiles) != 3 {
			t.Errorf("expected 3 social profiles, got %d", len(got.SocialProfiles))
		}
	})
}

func TestUser_String(t *testing.T) {
	clock := &stubClock{t: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)}

	userID, _ := kernel.NewID[user.User]("user-123")
	username, _ := shared.NewUsername("johndoe")
	email, _ := shared.NewEmail("john@example.com")
	firstName, _ := shared.NewFirstName("John")
	lastName, _ := shared.NewLastName("Doe")
	description, _ := shared.NewDescription("A very long description that should be truncated in the string representation")
	pictureURL, _ := kernel.NewURL[user.ProfilePicture]("https://example.com/pic.jpg")
	socialProfile, _ := user.NewSocialProfile(user.SocialMediaGitHub, "https://github.com/johndoe")

	params := user.NewUserParams{
		UserID:         userID,
		Username:       username,
		Email:          email,
		Roles:          []user.Role{user.RoleAdmin},
		FirstName:      firstName,
		LastName:       lastName,
		Description:    description,
		PictureURL:     pictureURL,
		SocialProfiles: []user.SocialProfile{socialProfile},
		Clock:          clock,
	}

	u, _ := user.NewUser(params)

	got := u.String()

	// Check that it contains key information
	checks := []string{
		`UserID: "user-123"`,
		`Username: "johndoe"`,
		`Email: "john@example.com"`,
		`FirstName: "John"`,
		`LastName: "Doe"`,
		`Description: "A very long description that should be truncated i..."`, // truncated
		`PictureURL: "https://example.com/pic.jpg"`,
		`Roles: [admin]`,
		"2024-01-15T10:00:00Z",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Errorf("String() missing expected content: %q\nGot: %s", check, got)
		}
	}
}

func TestUser_Validate(t *testing.T) {
	clock := &stubClock{t: time.Now()}

	t.Run("valid user passes", func(t *testing.T) {
		userID, _ := kernel.NewID[user.User]("user-123")
		username, _ := shared.NewUsername("johndoe")
		email, _ := shared.NewEmail("john@example.com")

		u, _ := user.NewUser(user.NewUserParams{
			UserID:   userID,
			Username: username,
			Email:    email,
			Roles:    []user.Role{user.RoleAuthor},
			Clock:    clock,
		})

		err := u.Validate()

		assertNoError(t, err)
	})

	t.Run("invalid fields fail", func(t *testing.T) {
		tests := []struct {
			name     string
			modifier func(*user.User)
		}{
			{
				name: "empty ID",
				modifier: func(u *user.User) {
					u.ID = kernel.ID[user.User]("")
				},
			},
			{
				name: "empty username",
				modifier: func(u *user.User) {
					u.Username = shared.Username("")
				},
			},
			{
				name: "empty email",
				modifier: func(u *user.User) {
					u.Email = shared.Email("")
				},
			},
			{
				name: "no roles",
				modifier: func(u *user.User) {
					u.Roles = []user.Role{}
				},
			},
			{
				name: "invalid role",
				modifier: func(u *user.User) {
					u.Roles = []user.Role{"invalid"}
				},
			},
			{
				name: "invalid first name",
				modifier: func(u *user.User) {
					u.FirstName = shared.FirstName(strings.Repeat("a", 51))
				},
			},
			{
				name: "invalid last name",
				modifier: func(u *user.User) {
					u.LastName = shared.LastName(strings.Repeat("a", 51))
				},
			},
			{
				name: "invalid description",
				modifier: func(u *user.User) {
					u.Description = shared.Description(strings.Repeat("a", 301))
				},
			},
			{
				name: "invalid picture URL",
				modifier: func(u *user.User) {
					u.PictureURL = kernel.URL[user.ProfilePicture]("not-a-url")
				},
			},
			{
				name: "invalid social profile",
				modifier: func(u *user.User) {
					u.SocialProfiles = []user.SocialProfile{
						{Platform: "invalid", URL: "https://example.com"},
					}
				},
			},
			{
				name: "duplicate social platforms",
				modifier: func(u *user.User) {
					u.SocialProfiles = []user.SocialProfile{
						{Platform: user.SocialMediaTwitter, URL: "https://twitter.com/user1"},
						{Platform: user.SocialMediaTwitter, URL: "https://twitter.com/user2"},
					}
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Create valid user
				userID, _ := kernel.NewID[user.User]("user-123")
				username, _ := shared.NewUsername("johndoe")
				email, _ := shared.NewEmail("john@example.com")

				u, _ := user.NewUser(user.NewUserParams{
					UserID:   userID,
					Username: username,
					Email:    email,
					Roles:    []user.Role{user.RoleAuthor},
					Clock:    clock,
				})

				// Apply modifier to make it invalid
				tt.modifier(&u)

				err := u.Validate()

				assertError(t, err)
				assertErrorCode(t, err, kernel.EInvalid)
			})
		}
	})
}

func TestUser_HasRole(t *testing.T) {
	clock := &stubClock{t: time.Now()}

	userID, _ := kernel.NewID[user.User]("user-123")
	username, _ := shared.NewUsername("johndoe")
	email, _ := shared.NewEmail("john@example.com")

	u, _ := user.NewUser(user.NewUserParams{
		UserID:   userID,
		Username: username,
		Email:    email,
		Roles:    []user.Role{user.RoleAdmin, user.RoleEditor},
		Clock:    clock,
	})

	t.Run("returns true for assigned roles", func(t *testing.T) {
		if !u.HasRole(user.RoleAdmin) {
			t.Error("expected HasRole(RoleAdmin) to be true")
		}
		if !u.HasRole(user.RoleEditor) {
			t.Error("expected HasRole(RoleEditor) to be true")
		}
	})

	t.Run("returns false for unassigned roles", func(t *testing.T) {
		if u.HasRole(user.RoleAuthor) {
			t.Error("expected HasRole(RoleAuthor) to be false")
		}
		if u.HasRole(user.RoleVisitor) {
			t.Error("expected HasRole(RoleVisitor) to be false")
		}
	})
}

func TestUser_HasAnyRole(t *testing.T) {
	clock := &stubClock{t: time.Now()}

	userID, _ := kernel.NewID[user.User]("user-123")
	username, _ := shared.NewUsername("johndoe")
	email, _ := shared.NewEmail("john@example.com")

	u, _ := user.NewUser(user.NewUserParams{
		UserID:   userID,
		Username: username,
		Email:    email,
		Roles:    []user.Role{user.RoleEditor, user.RoleAuthor},
		Clock:    clock,
	})

	t.Run("returns true if any role matches", func(t *testing.T) {
		if !u.HasAnyRole(user.RoleAdmin, user.RoleEditor) {
			t.Error("expected HasAnyRole to be true when Editor role is present")
		}
		if !u.HasAnyRole(user.RoleVisitor, user.RoleAuthor, user.RoleSubscriber) {
			t.Error("expected HasAnyRole to be true when Author role is present")
		}
	})

	t.Run("returns false if no roles match", func(t *testing.T) {
		if u.HasAnyRole(user.RoleAdmin, user.RoleVisitor) {
			t.Error("expected HasAnyRole to be false when no roles match")
		}
		if u.HasAnyRole(user.RoleSubscriber) {
			t.Error("expected HasAnyRole to be false for single unmatched role")
		}
	})

	t.Run("works with single role", func(t *testing.T) {
		if !u.HasAnyRole(user.RoleEditor) {
			t.Error("expected HasAnyRole to be true for single matched role")
		}
	})
}
