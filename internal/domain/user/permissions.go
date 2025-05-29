package user

import (
	"github.com/alnah/fla/internal/domain/kernel"
)

// PostInterface represents the minimal interface needed for permission checks.
// Avoids circular dependencies between user and post packages.
type PostInterface interface {
	GetOwner() kernel.ID[User]
	GetStatus() string
}

// PostPermissionChecker represents a user that can check permissions on posts.
// This interface is implemented by User to avoid circular dependencies.
type PostPermissionChecker interface {
	HasRole(role Role) bool
	HasAnyRole(roles ...Role) bool
	GetID() kernel.ID[User]
	CanEditPost(post PostInterface) bool
}

// GetID returns the user's ID for permission checks.
func (u User) GetID() kernel.ID[User] {
	return u.ID
}

// CanCreatePost determines if user has permission to create new blog posts.
// Authors, editors, and admins can create content in the system.
func (u User) CanCreatePost() bool {
	return u.HasAnyRole(RoleAdmin, RoleEditor, RoleAuthor)
}

// CanViewPost checks if user can access post content based on publication status.
// Published content is public; draft content requires ownership or editorial roles.
func (u User) CanViewPost(post PostInterface) bool {
	if post.GetStatus() == "published" {
		return true
	}

	return post.GetOwner() == u.ID || u.HasAnyRole(RoleAdmin, RoleEditor)
}

// CanEditPost determines editing permissions based on ownership and role hierarchy.
// Admins and editors can edit any post; authors can edit their own content.
func (u User) CanEditPost(post PostInterface) bool {
	if u.HasAnyRole(RoleAdmin, RoleEditor) {
		return true
	}

	return post.GetOwner() == u.ID && u.HasRole(RoleAuthor)
}

// CanDeletePost restricts deletion to appropriate users based on content status.
// Prevents accidental loss of published content while allowing draft cleanup.
func (u User) CanDeletePost(post PostInterface) bool {
	if u.HasRole(RoleAdmin) {
		return true
	}

	return post.GetOwner() == u.ID && post.GetStatus() == "draft"
}

// CanPublishPost determines publication permissions in the editorial workflow.
// Maintains content quality through role-based publication controls.
func (u User) CanPublishPost(post PostInterface) bool {
	if u.HasAnyRole(RoleAdmin, RoleEditor) {
		return true
	}

	return post.GetOwner() == u.ID && u.HasRole(RoleAuthor)
}

// CanSchedulePost checks permissions for delayed publication features.
// Enables content planning while maintaining editorial oversight.
func (u User) CanSchedulePost(post PostInterface) bool {
	return u.CanPublishPost(post)
}

// CanArchivePost determines who can remove content from active circulation.
// Restricts archiving to editorial roles to prevent content loss.
func (u User) CanArchivePost(post PostInterface) bool {
	return u.HasAnyRole(RoleAdmin, RoleEditor)
}

// CanChangePostStatus validates status transition permissions for workflow control.
// Ensures proper content lifecycle management through role-based restrictions.
func (u User) CanChangePostStatus(post PostInterface, newStatus string) bool {
	switch newStatus {
	case "draft":
		return u.CanEditPost(post)
	case "published":
		return u.CanPublishPost(post)
	case "scheduled":
		return u.CanSchedulePost(post)
	case "archived":
		return u.CanArchivePost(post)
	default:
		return false
	}
}

// CanManageCategories determines who can create and modify the content taxonomy.
// Restricts category management to prevent structural chaos in content organization.
func (u User) CanManageCategories() bool {
	return u.HasAnyRole(RoleAdmin, RoleEditor)
}

// CanManageTags controls who can create and modify content tags.
// Maintains tag consistency while allowing editorial content organization.
func (u User) CanManageTags() bool {
	return u.HasAnyRole(RoleAdmin, RoleEditor)
}

// CanAddTagToPost checks if user can associate tags with specific posts.
// Links tag management to content editing permissions for consistency.
func (u User) CanAddTagToPost(post PostInterface) bool {
	return u.CanEditPost(post)
}

// CanChangePostCategory determines who can move posts between categories.
// Prevents content misclassification while enabling editorial organization.
func (u User) CanChangePostCategory(post PostInterface) bool {
	return u.CanEditPost(post)
}
