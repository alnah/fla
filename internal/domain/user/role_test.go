package user_test

import (
	"testing"

	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/user"
)

func TestRole_String(t *testing.T) {
	tests := []struct {
		role user.Role
		want string
	}{
		{user.RoleAdmin, "admin"},
		{user.RoleEditor, "editor"},
		{user.RoleAuthor, "author"},
		{user.RoleSubscriber, "subscriber"},
		{user.RoleVisitor, "visitor"},
		{user.RoleMachine, "machine"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.role.String()

			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRole_Validate(t *testing.T) {
	t.Run("valid roles pass", func(t *testing.T) {
		validRoles := []user.Role{
			user.RoleAdmin,
			user.RoleEditor,
			user.RoleAuthor,
			user.RoleSubscriber,
			user.RoleVisitor,
			user.RoleMachine,
		}

		for _, role := range validRoles {
			t.Run(string(role), func(t *testing.T) {
				err := role.Validate()

				assertNoError(t, err)
			})
		}
	})

	t.Run("invalid role fails", func(t *testing.T) {
		invalidRoles := []user.Role{
			"",
			"superadmin",
			"moderator",
			"guest",
			"user",
			"ADMIN", // case sensitive
			"Admin",
			"admin ",
			" admin",
		}

		for _, role := range invalidRoles {
			t.Run(string(role), func(t *testing.T) {
				err := role.Validate()

				assertError(t, err)
				assertErrorCode(t, err, kernel.EInvalid)
				assertErrorMessage(t, err, user.MRoleInvalid)
			})
		}
	})
}

func TestRoleConstants(t *testing.T) {
	// Ensure role constants have expected values
	tests := []struct {
		name string
		role user.Role
		want string
	}{
		{"RoleAdmin", user.RoleAdmin, "admin"},
		{"RoleEditor", user.RoleEditor, "editor"},
		{"RoleAuthor", user.RoleAuthor, "author"},
		{"RoleSubscriber", user.RoleSubscriber, "subscriber"},
		{"RoleVisitor", user.RoleVisitor, "visitor"},
		{"RoleMachine", user.RoleMachine, "machine"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.role) != tt.want {
				t.Errorf("got %q, want %q", tt.role, tt.want)
			}
		})
	}
}
