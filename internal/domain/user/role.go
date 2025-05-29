package user

import (
	"github.com/alnah/fla/internal/domain/kernel"
)

const MRoleInvalid string = "Invalid role."

// Role defines permission levels for system access and content management.
// Enables fine-grained access control for collaborative blogging workflows.
type Role string

const (
	RoleAdmin      Role = "admin"      // Full system access and user management
	RoleEditor     Role = "editor"     // Content management and publication control
	RoleAuthor     Role = "author"     // Content creation and own post management
	RoleSubscriber Role = "subscriber" // Basic access for content consumption
	RoleVisitor    Role = "visitor"    // Anonymous read-only access
	RoleMachine    Role = "machine"    // Automated system access for integrations
)

func (r Role) String() string { return string(r) }

// Validate ensures role assignment uses defined permission levels.
// Prevents privilege escalation through invalid role specifications.
func (r Role) Validate() error {
	const op = "Role.Validate"

	switch r {
	case RoleAdmin, RoleEditor, RoleAuthor, RoleVisitor, RoleSubscriber, RoleMachine:
		return nil
	default:
		return &kernel.Error{
			Code:      kernel.EInvalid,
			Message:   MRoleInvalid,
			Operation: op,
		}
	}
}
