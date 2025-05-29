package user_test

import (
	"testing"
	"time"

	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/shared"
	"github.com/alnah/fla/internal/domain/user"
)

// mockPost implements user.PostInterface for testing
type mockPost struct {
	owner  kernel.ID[user.User]
	status string
}

func (m *mockPost) GetOwner() kernel.ID[user.User] {
	return m.owner
}

func (m *mockPost) GetStatus() string {
	return m.status
}

func createTestUser(id string, roles ...user.Role) user.User {
	clock := &stubClock{t: time.Now()}

	userID, _ := kernel.NewID[user.User](id)
	username, _ := shared.NewUsername("testuser")
	email, _ := shared.NewEmail("test@example.com")

	u, _ := user.NewUser(user.NewUserParams{
		UserID:   userID,
		Username: username,
		Email:    email,
		Roles:    roles,
		Clock:    clock,
	})

	return u
}

func TestUser_GetID(t *testing.T) {
	u := createTestUser("user-123", user.RoleAuthor)

	got := u.GetID()
	want, _ := kernel.NewID[user.User]("user-123")

	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestUser_CanCreatePost(t *testing.T) {
	tests := []struct {
		name  string
		roles []user.Role
		want  bool
	}{
		{"admin can create", []user.Role{user.RoleAdmin}, true},
		{"editor can create", []user.Role{user.RoleEditor}, true},
		{"author can create", []user.Role{user.RoleAuthor}, true},
		{"visitor cannot create", []user.Role{user.RoleVisitor}, false},
		{"subscriber cannot create", []user.Role{user.RoleSubscriber}, false},
		{"machine cannot create", []user.Role{user.RoleMachine}, false},
		{"multiple roles with author", []user.Role{user.RoleSubscriber, user.RoleAuthor}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := createTestUser("user-123", tt.roles...)

			got := u.CanCreatePost()

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_CanViewPost(t *testing.T) {
	ownerID, _ := kernel.NewID[user.User]("owner-123")

	tests := []struct {
		name       string
		userID     string
		userRoles  []user.Role
		postOwner  kernel.ID[user.User]
		postStatus string
		want       bool
	}{
		// Published posts - everyone can view
		{"anyone can view published", "user-123", []user.Role{user.RoleVisitor}, ownerID, "published", true},
		{"subscriber can view published", "user-123", []user.Role{user.RoleSubscriber}, ownerID, "published", true},
		{"author can view published", "user-123", []user.Role{user.RoleAuthor}, ownerID, "published", true},

		// Draft posts - owner, admin, editor can view
		{"owner can view own draft", "owner-123", []user.Role{user.RoleAuthor}, ownerID, "draft", true},
		{"admin can view any draft", "admin-123", []user.Role{user.RoleAdmin}, ownerID, "draft", true},
		{"editor can view any draft", "editor-123", []user.Role{user.RoleEditor}, ownerID, "draft", true},
		{"non-owner author cannot view draft", "other-123", []user.Role{user.RoleAuthor}, ownerID, "draft", false},
		{"visitor cannot view draft", "visitor-123", []user.Role{user.RoleVisitor}, ownerID, "draft", false},

		// Other statuses follow same rules as draft
		{"owner can view own scheduled", "owner-123", []user.Role{user.RoleAuthor}, ownerID, "scheduled", true},
		{"non-owner cannot view scheduled", "other-123", []user.Role{user.RoleAuthor}, ownerID, "scheduled", false},
		{"admin can view archived", "admin-123", []user.Role{user.RoleAdmin}, ownerID, "archived", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := createTestUser(tt.userID, tt.userRoles...)
			post := &mockPost{owner: tt.postOwner, status: tt.postStatus}

			got := u.CanViewPost(post)

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_CanEditPost(t *testing.T) {
	ownerID, _ := kernel.NewID[user.User]("owner-123")

	tests := []struct {
		name      string
		userID    string
		userRoles []user.Role
		postOwner kernel.ID[user.User]
		want      bool
	}{
		// Admins and editors can edit any post
		{"admin can edit any post", "admin-123", []user.Role{user.RoleAdmin}, ownerID, true},
		{"editor can edit any post", "editor-123", []user.Role{user.RoleEditor}, ownerID, true},

		// Authors can edit their own posts
		{"author can edit own post", "owner-123", []user.Role{user.RoleAuthor}, ownerID, true},
		{"author cannot edit others' post", "other-123", []user.Role{user.RoleAuthor}, ownerID, false},

		// Non-authors cannot edit
		{"visitor cannot edit", "user-123", []user.Role{user.RoleVisitor}, ownerID, false},
		{"subscriber cannot edit", "user-123", []user.Role{user.RoleSubscriber}, ownerID, false},
		{"owner without author role cannot edit", "owner-123", []user.Role{user.RoleSubscriber}, ownerID, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := createTestUser(tt.userID, tt.userRoles...)
			post := &mockPost{owner: tt.postOwner, status: "draft"}

			got := u.CanEditPost(post)

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_CanDeletePost(t *testing.T) {
	ownerID, _ := kernel.NewID[user.User]("owner-123")

	tests := []struct {
		name       string
		userID     string
		userRoles  []user.Role
		postOwner  kernel.ID[user.User]
		postStatus string
		want       bool
	}{
		// Admin can delete any post
		{"admin can delete any post", "admin-123", []user.Role{user.RoleAdmin}, ownerID, "draft", true},
		{"admin can delete published post", "admin-123", []user.Role{user.RoleAdmin}, ownerID, "published", true},

		// Others can only delete their own drafts
		{"owner can delete own draft", "owner-123", []user.Role{user.RoleAuthor}, ownerID, "draft", true},
		{"owner cannot delete own published", "owner-123", []user.Role{user.RoleAuthor}, ownerID, "published", false},
		{"owner cannot delete own scheduled", "owner-123", []user.Role{user.RoleAuthor}, ownerID, "scheduled", false},

		// Non-owners and non-admins cannot delete
		{"editor cannot delete", "editor-123", []user.Role{user.RoleEditor}, ownerID, "draft", false},
		{"non-owner cannot delete draft", "other-123", []user.Role{user.RoleAuthor}, ownerID, "draft", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := createTestUser(tt.userID, tt.userRoles...)
			post := &mockPost{owner: tt.postOwner, status: tt.postStatus}

			got := u.CanDeletePost(post)

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_CanPublishPost(t *testing.T) {
	ownerID, _ := kernel.NewID[user.User]("owner-123")

	tests := []struct {
		name      string
		userID    string
		userRoles []user.Role
		postOwner kernel.ID[user.User]
		want      bool
	}{
		// Admin and editor can publish any post
		{"admin can publish any post", "admin-123", []user.Role{user.RoleAdmin}, ownerID, true},
		{"editor can publish any post", "editor-123", []user.Role{user.RoleEditor}, ownerID, true},

		// Author can publish own post
		{"author can publish own post", "owner-123", []user.Role{user.RoleAuthor}, ownerID, true},
		{"author cannot publish others' post", "other-123", []user.Role{user.RoleAuthor}, ownerID, false},

		// Others cannot publish
		{"visitor cannot publish", "user-123", []user.Role{user.RoleVisitor}, ownerID, false},
		{"subscriber cannot publish", "user-123", []user.Role{user.RoleSubscriber}, ownerID, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := createTestUser(tt.userID, tt.userRoles...)
			post := &mockPost{owner: tt.postOwner, status: "draft"}

			got := u.CanPublishPost(post)

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_CanSchedulePost(t *testing.T) {
	// CanSchedulePost has same rules as CanPublishPost
	ownerID, _ := kernel.NewID[user.User]("owner-123")

	tests := []struct {
		name      string
		userID    string
		userRoles []user.Role
		postOwner kernel.ID[user.User]
		want      bool
	}{
		{"admin can schedule any post", "admin-123", []user.Role{user.RoleAdmin}, ownerID, true},
		{"editor can schedule any post", "editor-123", []user.Role{user.RoleEditor}, ownerID, true},
		{"author can schedule own post", "owner-123", []user.Role{user.RoleAuthor}, ownerID, true},
		{"author cannot schedule others' post", "other-123", []user.Role{user.RoleAuthor}, ownerID, false},
		{"visitor cannot schedule", "user-123", []user.Role{user.RoleVisitor}, ownerID, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := createTestUser(tt.userID, tt.userRoles...)
			post := &mockPost{owner: tt.postOwner, status: "draft"}

			got := u.CanSchedulePost(post)

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_CanArchivePost(t *testing.T) {
	ownerID, _ := kernel.NewID[user.User]("owner-123")

	tests := []struct {
		name      string
		userRoles []user.Role
		want      bool
	}{
		{"admin can archive", []user.Role{user.RoleAdmin}, true},
		{"editor can archive", []user.Role{user.RoleEditor}, true},
		{"author cannot archive", []user.Role{user.RoleAuthor}, false},
		{"visitor cannot archive", []user.Role{user.RoleVisitor}, false},
		{"subscriber cannot archive", []user.Role{user.RoleSubscriber}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := createTestUser("user-123", tt.userRoles...)
			post := &mockPost{owner: ownerID, status: "published"}

			got := u.CanArchivePost(post)

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_CanChangePostStatus(t *testing.T) {
	ownerID, _ := kernel.NewID[user.User]("owner-123")

	tests := []struct {
		name      string
		userID    string
		userRoles []user.Role
		postOwner kernel.ID[user.User]
		newStatus string
		want      bool
	}{
		// Draft status - same as edit permissions
		{"owner can change to draft", "owner-123", []user.Role{user.RoleAuthor}, ownerID, "draft", true},
		{"admin can change to draft", "admin-123", []user.Role{user.RoleAdmin}, ownerID, "draft", true},
		{"non-owner cannot change to draft", "other-123", []user.Role{user.RoleAuthor}, ownerID, "draft", false},

		// Published status - same as publish permissions
		{"owner can change to published", "owner-123", []user.Role{user.RoleAuthor}, ownerID, "published", true},
		{"non-owner cannot change to published", "other-123", []user.Role{user.RoleAuthor}, ownerID, "published", false},

		// Scheduled status - same as schedule permissions
		{"owner can change to scheduled", "owner-123", []user.Role{user.RoleAuthor}, ownerID, "scheduled", true},

		// Archived status - same as archive permissions
		{"admin can change to archived", "admin-123", []user.Role{user.RoleAdmin}, ownerID, "archived", true},
		{"author cannot change to archived", "author-123", []user.Role{user.RoleAuthor}, ownerID, "archived", false},

		// Unknown status
		{"no one can change to unknown status", "admin-123", []user.Role{user.RoleAdmin}, ownerID, "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := createTestUser(tt.userID, tt.userRoles...)
			post := &mockPost{owner: tt.postOwner, status: "draft"}

			got := u.CanChangePostStatus(post, tt.newStatus)

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_CanManageCategories(t *testing.T) {
	tests := []struct {
		name  string
		roles []user.Role
		want  bool
	}{
		{"admin can manage", []user.Role{user.RoleAdmin}, true},
		{"editor can manage", []user.Role{user.RoleEditor}, true},
		{"author cannot manage", []user.Role{user.RoleAuthor}, false},
		{"visitor cannot manage", []user.Role{user.RoleVisitor}, false},
		{"multiple roles with editor", []user.Role{user.RoleAuthor, user.RoleEditor}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := createTestUser("user-123", tt.roles...)

			got := u.CanManageCategories()

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_CanManageTags(t *testing.T) {
	tests := []struct {
		name  string
		roles []user.Role
		want  bool
	}{
		{"admin can manage", []user.Role{user.RoleAdmin}, true},
		{"editor can manage", []user.Role{user.RoleEditor}, true},
		{"author cannot manage", []user.Role{user.RoleAuthor}, false},
		{"visitor cannot manage", []user.Role{user.RoleVisitor}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := createTestUser("user-123", tt.roles...)

			got := u.CanManageTags()

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_CanAddTagToPost(t *testing.T) {
	ownerID, _ := kernel.NewID[user.User]("owner-123")

	// CanAddTagToPost has same rules as CanEditPost
	tests := []struct {
		name      string
		userID    string
		userRoles []user.Role
		postOwner kernel.ID[user.User]
		want      bool
	}{
		{"admin can add tag", "admin-123", []user.Role{user.RoleAdmin}, ownerID, true},
		{"editor can add tag", "editor-123", []user.Role{user.RoleEditor}, ownerID, true},
		{"owner can add tag to own post", "owner-123", []user.Role{user.RoleAuthor}, ownerID, true},
		{"non-owner cannot add tag", "other-123", []user.Role{user.RoleAuthor}, ownerID, false},
		{"visitor cannot add tag", "user-123", []user.Role{user.RoleVisitor}, ownerID, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := createTestUser(tt.userID, tt.userRoles...)
			post := &mockPost{owner: tt.postOwner, status: "draft"}

			got := u.CanAddTagToPost(post)

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_CanChangePostCategory(t *testing.T) {
	ownerID, _ := kernel.NewID[user.User]("owner-123")

	// CanChangePostCategory has same rules as CanEditPost
	tests := []struct {
		name      string
		userID    string
		userRoles []user.Role
		postOwner kernel.ID[user.User]
		want      bool
	}{
		{"admin can change category", "admin-123", []user.Role{user.RoleAdmin}, ownerID, true},
		{"editor can change category", "editor-123", []user.Role{user.RoleEditor}, ownerID, true},
		{"owner can change own post category", "owner-123", []user.Role{user.RoleAuthor}, ownerID, true},
		{"non-owner cannot change category", "other-123", []user.Role{user.RoleAuthor}, ownerID, false},
		{"visitor cannot change category", "user-123", []user.Role{user.RoleVisitor}, ownerID, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := createTestUser(tt.userID, tt.userRoles...)
			post := &mockPost{owner: tt.postOwner, status: "draft"}

			got := u.CanChangePostCategory(post)

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
