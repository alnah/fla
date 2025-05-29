package domain_test

import (
	"strings"
	"testing"
	"time"

	"github.com/alnah/fla/internal/domain"
)

// TestDomainTypeAliases verifies that all type aliases are correctly exported
func TestDomainTypeAliases(t *testing.T) {
	t.Run("kernel types", func(t *testing.T) {
		// Clock interface
		var _ domain.Clock = (*stubClock)(nil)

		// Error type
		err := &domain.Error{
			Code:      domain.EInvalid,
			Message:   "test error",
			Operation: "TestOp",
		}
		if err.Error() != "TestOp: <invalid> test error" {
			t.Errorf("Error formatting incorrect: %s", err.Error())
		}
	})

	t.Run("error constants", func(t *testing.T) {
		constants := map[string]string{
			"EConflict":  domain.EConflict,
			"EInternal":  domain.EInternal,
			"EInvalid":   domain.EInvalid,
			"EForbidden": domain.EForbidden,
			"ENotFound":  domain.ENotFound,
		}

		expected := map[string]string{
			"EConflict":  "conflict",
			"EInternal":  "internal",
			"EInvalid":   "invalid",
			"EForbidden": "forbidden",
			"ENotFound":  "not_found",
		}

		for name, value := range constants {
			if value != expected[name] {
				t.Errorf("%s: got %q, want %q", name, value, expected[name])
			}
		}
	})

	t.Run("error functions", func(t *testing.T) {
		err := &domain.Error{Code: domain.EInvalid, Message: "test message"}

		if domain.ErrorCode(err) != domain.EInvalid {
			t.Errorf("ErrorCode: got %q, want %q", domain.ErrorCode(err), domain.EInvalid)
		}

		if domain.ErrorMessage(err) != "test message" {
			t.Errorf("ErrorMessage: got %q, want %q", domain.ErrorMessage(err), "test message")
		}
	})

	t.Run("shared types", func(t *testing.T) {
		// Email
		email, err := domain.NewEmail("test@example.com")
		assertNoError(t, err)
		if email.String() != "test@example.com" {
			t.Errorf("Email: got %q, want %q", email, "test@example.com")
		}

		// FirstName
		firstName, err := domain.NewFirstName("John")
		assertNoError(t, err)
		if firstName.String() != "John" {
			t.Errorf("FirstName: got %q, want %q", firstName, "John")
		}

		// LastName
		lastName, err := domain.NewLastName("Doe")
		assertNoError(t, err)
		if lastName.String() != "Doe" {
			t.Errorf("LastName: got %q, want %q", lastName, "Doe")
		}

		// Username
		username, err := domain.NewUsername("johndoe")
		assertNoError(t, err)
		if username.String() != "johndoe" {
			t.Errorf("Username: got %q, want %q", username, "johndoe")
		}

		// Title
		title, err := domain.NewTitle("Test Title Here")
		assertNoError(t, err)
		if title.String() != "Test Title Here" {
			t.Errorf("Title: got %q, want %q", title, "Test Title Here")
		}

		// Description
		desc, err := domain.NewDescription("A test description")
		assertNoError(t, err)
		if desc.String() != "A test description" {
			t.Errorf("Description: got %q, want %q", desc, "A test description")
		}

		// Slug
		slug, err := domain.NewSlug("Test Title Here")
		assertNoError(t, err)
		if slug.String() != "test-title-here" {
			t.Errorf("Slug: got %q, want %q", slug, "test-title-here")
		}

		// Datetime
		now := time.Now()
		dt, err := domain.NewDatetime(now)
		assertNoError(t, err)
		if !dt.Time().Equal(now.UTC()) {
			t.Error("Datetime not properly created")
		}

		// DatetimeNow
		dtNow, err := domain.NewDatetimeNow()
		assertNoError(t, err)
		if dtNow.Time().After(time.Now()) {
			t.Error("DatetimeNow should not be in future")
		}

		// Pagination
		pagination, err := domain.NewPagination(1, 10, 100)
		assertNoError(t, err)
		if pagination.Page != 1 || pagination.Limit != 10 || pagination.TotalItems != 100 {
			t.Error("Pagination not properly created")
		}
	})

	t.Run("post types", func(t *testing.T) {
		// Status constants
		if domain.StatusDraft != "draft" {
			t.Errorf("StatusDraft: got %q, want %q", domain.StatusDraft, "draft")
		}
		if domain.StatusPublished != "published" {
			t.Errorf("StatusPublished: got %q, want %q", domain.StatusPublished, "published")
		}
		if domain.StatusArchived != "archived" {
			t.Errorf("StatusArchived: got %q, want %q", domain.StatusArchived, "archived")
		}
		if domain.StatusScheduled != "scheduled" {
			t.Errorf("StatusScheduled: got %q, want %q", domain.StatusScheduled, "scheduled")
		}

		// PostContent
		content, err := domain.NewPostContent(strings.Repeat("a", 300))
		assertNoError(t, err)
		if len(content.String()) != 300 {
			t.Error("PostContent not properly created")
		}

		// PostID
		postID, err := domain.NewPostID("post-123")
		assertNoError(t, err)
		if postID.String() != "post-123" {
			t.Errorf("PostID: got %q, want %q", postID, "post-123")
		}
	})

	t.Run("user types", func(t *testing.T) {
		// Role constants
		roles := map[string]domain.Role{
			"Admin":      domain.RoleAdmin,
			"Editor":     domain.RoleEditor,
			"Author":     domain.RoleAuthor,
			"Subscriber": domain.RoleSubscriber,
			"Visitor":    domain.RoleVisitor,
			"Machine":    domain.RoleMachine,
		}

		expected := map[string]string{
			"Admin":      "admin",
			"Editor":     "editor",
			"Author":     "author",
			"Subscriber": "subscriber",
			"Commenter":  "commenter",
			"Visitor":    "visitor",
			"Machine":    "machine",
		}

		for name, role := range roles {
			if string(role) != expected[name] {
				t.Errorf("Role%s: got %q, want %q", name, role, expected[name])
			}
		}

		// UserID
		userID, err := domain.NewUserID("user-123")
		assertNoError(t, err)
		if userID.String() != "user-123" {
			t.Errorf("UserID: got %q, want %q", userID, "user-123")
		}

		// SocialProfile
		profile, err := domain.NewSocialProfile(domain.SocialMediaGitHub, "https://github.com/user")
		assertNoError(t, err)
		if profile.Platform != domain.SocialMediaGitHub {
			t.Error("SocialProfile platform mismatch")
		}
	})

	t.Run("category types", func(t *testing.T) {
		// CategoryName
		catName, err := domain.NewCategoryName("Test Category")
		assertNoError(t, err)
		if catName.String() != "Test Category" {
			t.Errorf("CategoryName: got %q, want %q", catName, "Test Category")
		}

		// CategoryID
		catID, err := domain.NewCategoryID("cat-123")
		assertNoError(t, err)
		if catID.String() != "cat-123" {
			t.Errorf("CategoryID: got %q, want %q", catID, "cat-123")
		}
	})

	t.Run("tag types", func(t *testing.T) {
		// TagName
		tagName, err := domain.NewTagName("test-tag")
		assertNoError(t, err)
		if tagName.String() != "test-tag" {
			t.Errorf("TagName: got %q, want %q", tagName, "test-tag")
		}

		// TagID
		tagID, err := domain.NewTagID("tag-123")
		assertNoError(t, err)
		if tagID.String() != "tag-123" {
			t.Errorf("TagID: got %q, want %q", tagID, "tag-123")
		}
	})

	t.Run("subscription types", func(t *testing.T) {
		// Status constants
		if domain.SubscriptionStatusActive != "active" {
			t.Errorf("SubscriptionStatusActive: got %q, want %q",
				domain.SubscriptionStatusActive, "active")
		}
		if domain.SubscriptionStatusUnsubscribed != "unsubscribed" {
			t.Errorf("SubscriptionStatusUnsubscribed: got %q, want %q",
				domain.SubscriptionStatusUnsubscribed, "unsubscribed")
		}
		if domain.SubscriptionStatusBounced != "bounced" {
			t.Errorf("SubscriptionStatusBounced: got %q, want %q",
				domain.SubscriptionStatusBounced, "bounced")
		}
		if domain.SubscriptionStatusComplained != "complained" {
			t.Errorf("SubscriptionStatusComplained: got %q, want %q",
				domain.SubscriptionStatusComplained, "complained")
		}

		// SubscriptionID
		subID, err := domain.NewSubscriptionID("sub-123")
		assertNoError(t, err)
		if subID.String() != "sub-123" {
			t.Errorf("SubscriptionID: got %q, want %q", subID, "sub-123")
		}
	})

	t.Run("URL types", func(t *testing.T) {
		// Generic URL
		url, err := domain.NewURL("https://example.com")
		assertNoError(t, err)
		if url.String() != "https://example.com" {
			t.Errorf("URL: got %q, want %q", url, "https://example.com")
		}
	})
}

// TestDomainEntityCreation tests that entities can be created through the domain facade
func TestDomainEntityCreation(t *testing.T) {
	clock := &stubClock{t: time.Now()}

	t.Run("create user", func(t *testing.T) {
		userID, _ := domain.NewUserID("user-123")
		username, _ := domain.NewUsername("johndoe")
		email, _ := domain.NewEmail("john@example.com")

		user, err := domain.NewUser(domain.NewUserParams{
			UserID:   userID,
			Username: username,
			Email:    email,
			Roles:    []domain.Role{domain.RoleAuthor},
			Clock:    clock,
		})

		assertNoError(t, err)
		if user.ID != userID {
			t.Error("User ID mismatch")
		}
	})

	t.Run("create category", func(t *testing.T) {
		categoryID, _ := domain.NewCategoryID("cat-123")
		categoryName, _ := domain.NewCategoryName("Test Category")
		userID, _ := domain.NewUserID("user-123")

		category, err := domain.NewCategory(domain.NewCategoryParams{
			CategoryID: categoryID,
			Name:       categoryName,
			CreatedBy:  userID,
			Clock:      clock,
		})

		assertNoError(t, err)
		if category.CategoryID != categoryID {
			t.Error("Category ID mismatch")
		}
	})

	t.Run("create post", func(t *testing.T) {
		postID, _ := domain.NewPostID("post-123")
		ownerID, _ := domain.NewUserID("user-123")
		title, _ := domain.NewTitle("Test Post Title")
		content, _ := domain.NewPostContent(strings.Repeat("a", 300))

		// Create category for post
		categoryID, _ := domain.NewCategoryID("cat-123")
		categoryName, _ := domain.NewCategoryName("Test Category")
		userID, _ := domain.NewUserID("user-123")

		category, _ := domain.NewCategory(domain.NewCategoryParams{
			CategoryID: categoryID,
			Name:       categoryName,
			CreatedBy:  userID,
			Clock:      clock,
		})

		post, err := domain.NewPost(domain.NewPostParams{
			PostID:   postID,
			Owner:    ownerID,
			Title:    title,
			Content:  content,
			Status:   domain.StatusDraft,
			Category: category,
			Clock:    clock,
		})

		assertNoError(t, err)
		if post.PostID != postID {
			t.Error("Post ID mismatch")
		}
	})

	t.Run("create tag", func(t *testing.T) {
		tagID, _ := domain.NewTagID("tag-123")
		tagName, _ := domain.NewTagName("test-tag")
		userID, _ := domain.NewUserID("user-123")

		tag, err := domain.NewTag(domain.Tag{
			TagID:     tagID,
			Name:      tagName,
			CreatedBy: userID,
			CreatedAt: clock.Now(),
		})

		assertNoError(t, err)
		if tag.TagID != tagID {
			t.Error("Tag ID mismatch")
		}
	})

	t.Run("create subscription", func(t *testing.T) {
		subID, _ := domain.NewSubscriptionID("sub-123")
		firstName, _ := domain.NewFirstName("John")
		email, _ := domain.NewEmail("john@example.com")

		sub, err := domain.NewSubscription(domain.NewSubscriptionParams{
			SubscriptionID: subID,
			FirstName:      firstName,
			Email:          email,
			Clock:          clock,
		})

		assertNoError(t, err)
		if sub.SubscriptionID != subID {
			t.Error("Subscription ID mismatch")
		}
	})
}

// TestDomainServiceCreation tests that services can be created through the domain facade
func TestDomainServiceCreation(t *testing.T) {
	t.Run("create path service", func(t *testing.T) {
		repo := &mockCategoryRepository{}
		service := domain.NewCategoryPathService(repo)

		if service == nil {
			t.Error("expected path service to be created")
		}
	})
}

// TestPostsList tests the PostsList type
func TestPostsList(t *testing.T) {
	clock := &stubClock{t: time.Now()}

	// Create test posts
	posts := make([]domain.Post, 3)
	for i := range 3 {
		postID, _ := domain.NewPostID(string(rune('a' + i)))
		ownerID, _ := domain.NewUserID("user-123")
		title, _ := domain.NewTitle("Test Post")
		content, _ := domain.NewPostContent(strings.Repeat("a", 300))

		categoryID, _ := domain.NewCategoryID("cat-123")
		categoryName, _ := domain.NewCategoryName("Test Category")
		userID, _ := domain.NewUserID("user-123")

		category, _ := domain.NewCategory(domain.NewCategoryParams{
			CategoryID: categoryID,
			Name:       categoryName,
			CreatedBy:  userID,
			Clock:      clock,
		})

		posts[i], _ = domain.NewPost(domain.NewPostParams{
			PostID:   postID,
			Owner:    ownerID,
			Title:    title,
			Content:  content,
			Status:   domain.StatusDraft,
			Category: category,
			Clock:    clock,
		})
	}

	pagination, _ := domain.NewPagination(1, 10, 3)

	list := domain.NewPostsList(posts, pagination)

	if list.Count() != 3 {
		t.Errorf("Count: got %d, want %d", list.Count(), 3)
	}
	if list.IsEmpty() {
		t.Error("expected list not to be empty")
	}
}

// Test helpers
type stubClock struct {
	t time.Time
}

func (s *stubClock) Now() time.Time { return s.t }

type mockCategoryRepository struct{}

func (m *mockCategoryRepository) Create(category domain.Category) error { return nil }
func (m *mockCategoryRepository) GetByID(categoryID domain.CategoryID) (*domain.Category, error) {
	return nil, nil
}
func (m *mockCategoryRepository) GetAll() ([]domain.Category, error)        { return nil, nil }
func (m *mockCategoryRepository) Update(category domain.Category) error     { return nil }
func (m *mockCategoryRepository) Delete(categoryID domain.CategoryID) error { return nil }
func (m *mockCategoryRepository) GetChildren(categoryID domain.CategoryID) ([]domain.Category, error) {
	return nil, nil
}

func (m *mockCategoryRepository) GetRootCategories() ([]domain.Category, error) { return nil, nil }
func (m *mockCategoryRepository) BuildPath(categoryID domain.CategoryID) (domain.CategoryPath, error) {
	return nil, nil
}

func (m *mockCategoryRepository) FindByPath(pathSegments []string) (*domain.Category, error) {
	return nil, nil
}

func (m *mockCategoryRepository) IsSlugUniqueInParent(slug domain.Slug, parentID *domain.CategoryID) (bool, error) {
	return true, nil
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
