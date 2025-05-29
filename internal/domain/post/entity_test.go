package post_test

import (
	"strings"
	"testing"
	"time"

	"github.com/alnah/fla/internal/domain/category"
	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/post"
	"github.com/alnah/fla/internal/domain/shared"
	"github.com/alnah/fla/internal/domain/user"
	"slices"
)

// Mock clock for testing
type mockClock struct {
	now time.Time
}

func (m *mockClock) Now() time.Time {
	return m.now
}

// Mock user for testing post permissions
type mockUser struct {
	id    kernel.ID[user.User]
	roles []user.Role
}

func (m *mockUser) HasRole(role user.Role) bool {
	return slices.Contains(m.roles, role)
}

func (m *mockUser) HasAnyRole(roles ...user.Role) bool {
	return slices.ContainsFunc(roles, m.HasRole)
}

func (m *mockUser) GetID() kernel.ID[user.User] {
	return m.id
}

func (m *mockUser) CanEditPost(p user.PostInterface) bool {
	if m.HasAnyRole(user.RoleAdmin, user.RoleEditor) {
		return true
	}
	return p.GetOwner() == m.id && m.HasRole(user.RoleAuthor)
}

// Helper function to create a test category
func createTestCategory(t *testing.T, clock kernel.Clock) category.Category {
	t.Helper()

	categoryID, err := kernel.NewID[category.Category]("test-category-id")
	if err != nil {
		t.Fatalf("failed to create category ID: %v", err)
	}

	categoryName, err := category.NewCategoryName("Test Category")
	if err != nil {
		t.Fatalf("failed to create category name: %v", err)
	}

	userID, err := kernel.NewID[user.User]("user-123")
	if err != nil {
		t.Fatalf("failed to create user ID: %v", err)
	}

	cat, err := category.NewCategory(category.NewCategoryParams{
		CategoryID: categoryID,
		Name:       categoryName,
		CreatedBy:  userID,
		Clock:      clock,
	})
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	return cat
}

func TestNewPost(t *testing.T) {
	clock := &mockClock{now: time.Now()}

	t.Run("creates post with minimal fields", func(t *testing.T) {
		// Create required fields
		postID, err := kernel.NewID[post.Post]("post-123")
		if err != nil {
			t.Fatalf("failed to create post ID: %v", err)
		}

		ownerID, err := kernel.NewID[user.User]("user-123")
		if err != nil {
			t.Fatalf("failed to create owner ID: %v", err)
		}

		title, err := shared.NewTitle("Test Post Title Example") // Ensure > 10 chars
		if err != nil {
			t.Fatalf("failed to create title: %v", err)
		}

		content, err := post.NewPostContent(strings.Repeat("This is test content. ", 20)) // ~20 words * 5 chars = 100 chars, repeat to get 300+ chars
		if err != nil {
			t.Fatalf("failed to create content: %v", err)
		}

		cat := createTestCategory(t, clock)
		featuredImage, err := kernel.NewURL[post.FeaturedImage]("")
		if err != nil {
			t.Fatalf("failed to create featured image URL: %v", err)
		}

		// Create post
		p, err := post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        post.StatusDraft,
			Category:      cat,
			Clock:         clock,
		})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
			return
		}

		// Verify fields
		if p.PostID != postID {
			t.Errorf("expected PostID %q, got %q", postID, p.PostID)
		}
		if p.Owner != ownerID {
			t.Errorf("expected Owner %q, got %q", ownerID, p.Owner)
		}
		if p.Title != title {
			t.Errorf("expected Title %q, got %q", title, p.Title)
		}
		if p.Content != content {
			t.Errorf("expected Content %q, got %q", content, p.Content)
		}
		if p.Status != post.StatusDraft {
			t.Errorf("expected Status %q, got %q", post.StatusDraft, p.Status)
		}
	})

	t.Run("creates post with all optional fields", func(t *testing.T) {
		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("user-123")
		title, _ := shared.NewTitle("Test Post Title Example")
		content, _ := post.NewPostContent(strings.Repeat("This is test content. ", 20))
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("https://example.com/image.jpg")
		seoTitle, _ := shared.NewTitle("SEO Title Example")
		seoDescription, _ := shared.NewDescription("SEO description for search engines")
		ogTitle, _ := shared.NewTitle("Open Graph Title")
		ogDescription, _ := shared.NewDescription("Open Graph description for social media")
		ogImage, _ := kernel.NewURL[post.OpenGraphImage]("https://example.com/og-image.jpg")
		canonicalURL, _ := kernel.NewURL[post.Canonical]("https://example.com/canonical")
		publishedAt := time.Now().Add(-1 * time.Hour) // Past time for published post

		p, err := post.NewPost(post.NewPostParams{
			PostID:               postID,
			Owner:                ownerID,
			Title:                title,
			Content:              content,
			FeaturedImage:        featuredImage,
			Status:               post.StatusPublished,
			Category:             cat,
			PublishedAt:          &publishedAt,
			SEOTitle:             seoTitle,
			SEODescription:       seoDescription,
			OpenGraphTitle:       ogTitle,
			OpenGraphDescription: ogDescription,
			OpenGraphImage:       ogImage,
			CanonicalURL:         canonicalURL,
			SchemaType:           post.SchemaTypeArticle,
			Clock:                clock,
		})

		assertNoError(t, err)

		// Verify optional fields are set
		if p.SEOTitle != seoTitle {
			t.Errorf("SEOTitle: got %v, want %v", p.SEOTitle, seoTitle)
		}
		if p.SEODescription != seoDescription {
			t.Errorf("SEODescription: got %v, want %v", p.SEODescription, seoDescription)
		}
		if p.OpenGraphTitle != ogTitle {
			t.Errorf("OpenGraphTitle: got %v, want %v", p.OpenGraphTitle, ogTitle)
		}
		if p.OpenGraphDescription != ogDescription {
			t.Errorf("OpenGraphDescription: got %v, want %v", p.OpenGraphDescription, ogDescription)
		}
		if p.OpenGraphImage != ogImage {
			t.Errorf("OpenGraphImage: got %v, want %v", p.OpenGraphImage, ogImage)
		}
		if p.CanonicalURL != canonicalURL {
			t.Errorf("CanonicalURL: got %v, want %v", p.CanonicalURL, canonicalURL)
		}
		if p.SchemaType != post.SchemaTypeArticle {
			t.Errorf("SchemaType: got %v, want %v", p.SchemaType, post.SchemaTypeArticle)
		}
	})

	t.Run("generates slug from title", func(t *testing.T) {
		testCases := []struct {
			name         string
			title        string
			expectedSlug string
		}{
			{"Simple Title", "Simple Title Example", "simple-title-example"},
			{"Français: Leçon 1", "Français: Leçon Number One", "francais-lecon-number-one"},
			{"Title with Numbers 123", "Title with Numbers 123", "title-with-numbers-123"},
			{"UPPERCASE TITLE", "UPPERCASE TITLE EXAMPLE", "uppercase-title-example"},
			{"Title!!!With???Punctuation", "Title!!!With???Punctuation!!!", "title-with-punctuation"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				postID, _ := kernel.NewID[post.Post]("post-123")
				ownerID, _ := kernel.NewID[user.User]("user-123")
				title, err := shared.NewTitle(tc.title)
				if err != nil {
					t.Fatalf("failed to create title: %v", err)
				}
				content, _ := post.NewPostContent(strings.Repeat("This is test content. ", 20))
				cat := createTestCategory(t, clock)
				featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

				p, err := post.NewPost(post.NewPostParams{
					PostID:        postID,
					Owner:         ownerID,
					Title:         title,
					Content:       content,
					FeaturedImage: featuredImage,
					Status:        post.StatusDraft,
					Category:      cat,
					Clock:         clock,
				})

				if err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}

				if p.Slug.String() != tc.expectedSlug {
					t.Errorf("expected slug %q, got %q", tc.expectedSlug, p.Slug)
				}
			})
		}
	})

	t.Run("validates required fields", func(t *testing.T) {
		// Create valid parameters first
		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("user-123")
		title, _ := shared.NewTitle("Test Post Title Example")
		content, _ := post.NewPostContent(strings.Repeat("This is test content. ", 20))
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		tests := []struct {
			name   string
			modify func(*post.NewPostParams)
		}{
			{
				name: "invalid post ID",
				modify: func(p *post.NewPostParams) {
					p.PostID = kernel.ID[post.Post]("")
				},
			},
			{
				name: "invalid owner ID",
				modify: func(p *post.NewPostParams) {
					p.Owner = kernel.ID[user.User]("")
				},
			},
			{
				name: "invalid title",
				modify: func(p *post.NewPostParams) {
					p.Title = shared.Title("")
				},
			},
			{
				name: "invalid content",
				modify: func(p *post.NewPostParams) {
					p.Content = post.PostContent("")
				},
			},
			{
				name: "invalid status",
				modify: func(p *post.NewPostParams) {
					p.Status = post.Status("invalid")
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				params := post.NewPostParams{
					PostID:        postID,
					Owner:         ownerID,
					Title:         title,
					Content:       content,
					FeaturedImage: featuredImage,
					Status:        post.StatusDraft,
					Category:      cat,
					Clock:         clock,
				}

				tt.modify(&params)

				_, err := post.NewPost(params)

				assertError(t, err)
				assertErrorCode(t, err, kernel.EInvalid)
			})
		}
	})

	t.Run("validates optional SEO fields when present", func(t *testing.T) {
		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("user-123")
		title, _ := shared.NewTitle("Test Post Title Example")
		content, _ := post.NewPostContent(strings.Repeat("This is test content. ", 20))
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		// Test invalid SEO title
		_, err := post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        post.StatusDraft,
			Category:      cat,
			SEOTitle:      shared.Title("a"), // Invalid too short SEO title
			Clock:         clock,
		})

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)

		// Test invalid Open Graph title
		_, err = post.NewPost(post.NewPostParams{
			PostID:         postID,
			Owner:          ownerID,
			Title:          title,
			Content:        content,
			FeaturedImage:  featuredImage,
			Status:         post.StatusDraft,
			Category:       cat,
			OpenGraphTitle: shared.Title("a"), // Invalid too short OG title
			Clock:          clock,
		})

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("validates scheduled post requirements", func(t *testing.T) {
		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("user-123")
		title, _ := shared.NewTitle("Test Post Title Example")
		content, _ := post.NewPostContent(strings.Repeat("This is test content. ", 20))
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		// Test scheduled post without published date
		_, err := post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        post.StatusScheduled,
			Category:      cat,
			PublishedAt:   nil, // Missing required date
			Clock:         clock,
		})

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)

		// Test scheduled post with past date
		pastTime := clock.Now().Add(-1 * time.Hour)
		_, err = post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        post.StatusScheduled,
			Category:      cat,
			PublishedAt:   &pastTime, // Past date
			Clock:         clock,
		})

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})
}

func TestPost_String(t *testing.T) {
	clock := &mockClock{now: time.Now()}

	// Create a post with all fields
	postID, err := kernel.NewID[post.Post]("post-123")
	if err != nil {
		t.Fatalf("failed to create post ID: %v", err)
	}

	ownerID, err := kernel.NewID[user.User]("user-123")
	if err != nil {
		t.Fatalf("failed to create owner ID: %v", err)
	}

	title, err := shared.NewTitle("Test Post Title Example")
	if err != nil {
		t.Fatalf("failed to create title: %v", err)
	}

	// Create content that's long enough to be truncated in String()
	longContent := strings.Repeat("This is a test post. ", 50) // Will be > 100 chars
	content, err := post.NewPostContent(longContent)
	if err != nil {
		t.Fatalf("failed to create content: %v", err)
	}

	cat := createTestCategory(t, clock)
	featuredImage, err := kernel.NewURL[post.FeaturedImage]("https://example.com/image.jpg")
	if err != nil {
		t.Fatalf("failed to create featured image URL: %v", err)
	}

	p, err := post.NewPost(post.NewPostParams{
		PostID:        postID,
		Owner:         ownerID,
		Title:         title,
		Content:       content,
		FeaturedImage: featuredImage,
		Status:        post.StatusDraft,
		Category:      cat,
		Clock:         clock,
	})

	if err != nil {
		t.Fatalf("failed to create post: %v", err)
	}

	str := p.String()

	// Check that all expected content is in the string representation
	expectedParts := []string{
		`ID: "post-123"`,
		`Title: "Test Post Title Example"`,
		`Status: "draft"`,
		`Slug: "test-post-title-example"`,
		`Owner: "user-123"`,
		`Category: "Test Category"`,
		`Content: "This is a test post. This is a test post. This is a test post. This is a test post. This is a test p..."`,
		`WordCount: 250`,
		`HasFeaturedImage: true`,
	}

	for _, part := range expectedParts {
		if !strings.Contains(str, part) {
			t.Errorf("String() missing expected content: %q\n\tGot: %s", part, str)
		}
	}
}

func TestPost_WordCount(t *testing.T) {
	clock := &mockClock{now: time.Now()}

	testCases := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "simple text",
			content:  strings.Repeat("word ", 100), // 100 words, 500 chars - definitely > 300
			expected: 100,
		},
		{
			name:     "text with punctuation",
			content:  strings.Repeat("hello, world! ", 30), // 60 words, 420 chars
			expected: 60,
		},
		{
			name:     "text with markdown",
			content:  "**" + strings.Repeat("word ", 100) + "**", // 100 words, 500+ chars
			expected: 100,
		},
		{
			name:     "text with code blocks",
			content:  strings.Repeat("word ", 100) + "```code```", // 100 words remain
			expected: 100,
		},
		{
			name:     "text with links",
			content:  strings.Repeat("[word](http://x.com) ", 50), // 50 words, 1000+ chars
			expected: 50,
		},
		{
			name:     "text with multiple spaces",
			content:  strings.Repeat("word   ", 100), // 100 words, 700 chars
			expected: 100,
		},
		{
			name:     "unicode text",
			content:  strings.Repeat("café naïve 北京 ", 30), // 90 words, 450+ chars
			expected: 90,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Ensure content is long enough (300+ chars)
			if len(tc.content) < 300 {
				tc.content = tc.content + strings.Repeat(" padding ", 50)
			}

			postID, _ := kernel.NewID[post.Post]("post-123")
			ownerID, _ := kernel.NewID[user.User]("user-123")
			title, err := shared.NewTitle("Test Post Title Example")
			if err != nil {
				t.Fatalf("failed to create title: %v", err)
			}
			content, err := post.NewPostContent(tc.content)
			if err != nil {
				t.Fatalf("failed to create content: %v", err)
			}
			cat := createTestCategory(t, clock)
			featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

			p, err := post.NewPost(post.NewPostParams{
				PostID:        postID,
				Owner:         ownerID,
				Title:         title,
				Content:       content,
				FeaturedImage: featuredImage,
				Status:        post.StatusDraft,
				Category:      cat,
				Clock:         clock,
			})

			if err != nil {
				t.Fatalf("failed to create post: %v", err)
			}

			got := p.WordCount()
			if got != tc.expected {
				t.Errorf("got %d, want %d", got, tc.expected)
			}
		})
	}
}

func TestPost_EstimatedReadingTime(t *testing.T) {
	clock := &mockClock{now: time.Now()}

	testCases := []struct {
		name        string
		wordCount   int
		expectedMin int
	}{
		{"very short content", 10, 1}, // Always at least 1 minute
		{"50 words", 50, 1},           // 50/200 = 0.25, rounds up to 1
		{"100 words", 100, 1},         // 100/200 = 0.5, rounds up to 1
		{"200 words", 200, 1},         // 200/200 = 1.0, exactly 1
		{"250 words", 250, 2},         // 250/200 = 1.25, rounds up to 2
		{"400 words", 400, 2},         // 400/200 = 2.0, exactly 2
		{"500 words", 500, 3},         // 500/200 = 2.5, rounds up to 3
		{"1000 words", 1000, 5},       // 1000/200 = 5.0, exactly 5
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create content with approximately the desired word count
			content := strings.Repeat("word ", tc.wordCount)
			// Ensure minimum length requirement is met
			if len(content) < 300 {
				content = content + strings.Repeat(" padding ", 50)
			}

			postID, _ := kernel.NewID[post.Post]("post-123")
			ownerID, _ := kernel.NewID[user.User]("user-123")
			title, _ := shared.NewTitle("Test Post Title Example")
			postContent, _ := post.NewPostContent(content)
			cat := createTestCategory(t, clock)
			featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

			p, _ := post.NewPost(post.NewPostParams{
				PostID:        postID,
				Owner:         ownerID,
				Title:         title,
				Content:       postContent,
				FeaturedImage: featuredImage,
				Status:        post.StatusDraft,
				Category:      cat,
				Clock:         clock,
			})

			got := p.EstimatedReadingTime()
			if got != tc.expectedMin {
				t.Errorf("got %d minutes, want %d minutes", got, tc.expectedMin)
			}
		})
	}
}

func TestPost_StateChecks(t *testing.T) {
	clock := &mockClock{now: time.Now()}

	createPost := func(status post.Status) post.Post {
		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("user-123")
		title, _ := shared.NewTitle("Test Post Title Example")
		content, _ := post.NewPostContent(strings.Repeat("Test content. ", 25))
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		params := post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        status,
			Category:      cat,
			Clock:         clock,
		}

		// For scheduled posts, set a future PublishedAt date
		if status == post.StatusScheduled {
			futureTime := clock.Now().Add(24 * time.Hour) // 24 hours in the future
			params.PublishedAt = &futureTime
		}

		p, _ := post.NewPost(params)
		return p
	}

	t.Run("draft post states", func(t *testing.T) {
		p := createPost(post.StatusDraft)

		if !p.IsDraft() {
			t.Error("expected IsDraft to be true")
		}
		if p.IsPublished() {
			t.Error("expected IsPublished to be false")
		}
		if p.IsScheduled() {
			t.Error("expected IsScheduled to be false")
		}
	})

	t.Run("published post states", func(t *testing.T) {
		p := createPost(post.StatusPublished)

		if p.IsDraft() {
			t.Error("expected IsDraft to be false")
		}
		if !p.IsPublished() {
			t.Error("expected IsPublished to be true")
		}
		if p.IsScheduled() {
			t.Error("expected IsScheduled to be false")
		}
	})

	t.Run("scheduled post states", func(t *testing.T) {
		p := createPost(post.StatusScheduled)

		if p.IsDraft() {
			t.Error("expected IsDraft to be false")
		}
		if p.IsPublished() {
			t.Error("expected IsPublished to be false")
		}
		if !p.IsScheduled() {
			t.Error("expected IsScheduled to be true")
		}
	})

	t.Run("archived post states", func(t *testing.T) {
		p := createPost(post.StatusArchived)

		if p.IsDraft() {
			t.Error("expected IsDraft to be false")
		}
		if p.IsPublished() {
			t.Error("expected IsPublished to be false")
		}
		if p.IsScheduled() {
			t.Error("expected IsScheduled to be false")
		}
	})
}

func TestPost_HasFeaturedImage(t *testing.T) {
	clock := &mockClock{now: time.Now()}

	t.Run("with featured image", func(t *testing.T) {
		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("user-123")
		title, _ := shared.NewTitle("Test Post Title Example")
		content, _ := post.NewPostContent(strings.Repeat("Test content. ", 25))
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("https://example.com/image.jpg")

		p, _ := post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        post.StatusDraft,
			Category:      cat,
			Clock:         clock,
		})

		if !p.HasFeaturedImage() {
			t.Error("expected HasFeaturedImage to be true")
		}
	})

	t.Run("without featured image", func(t *testing.T) {
		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("user-123")
		title, _ := shared.NewTitle("Test Post Title Example")
		content, _ := post.NewPostContent(strings.Repeat("Test content. ", 25))
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		p, _ := post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        post.StatusDraft,
			Category:      cat,
			Clock:         clock,
		})

		if p.HasFeaturedImage() {
			t.Error("expected HasFeaturedImage to be false")
		}
	})
}

func TestPost_GetExcerpt(t *testing.T) {
	clock := &mockClock{now: time.Now()}

	createPostWithContent := func(content string) post.Post {
		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("user-123")
		title, _ := shared.NewTitle("Test Post Title Example")

		// Ensure content meets minimum length
		if len(content) < 300 {
			content = content + strings.Repeat(" padding ", 50)
		}

		postContent, _ := post.NewPostContent(content)
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		p, _ := post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       postContent,
			FeaturedImage: featuredImage,
			Status:        post.StatusDraft,
			Category:      cat,
			Clock:         clock,
		})

		return p
	}

	testCases := []struct {
		name     string
		content  string
		maxLen   int
		expected string
	}{
		{
			name:     "short content",
			content:  strings.Repeat("a", 300),
			maxLen:   50,
			expected: strings.Repeat("a", 50) + "...",
		},
		{
			name:     "long content with word boundary",
			content:  "This is a long post content that should be truncated at a word boundary for better readability." + strings.Repeat(" More content here to reach minimum length.", 10),
			maxLen:   50,
			expected: "This is a long post content that should be...",
		},
		{
			name:     "markdown content stripped",
			content:  "# Title\n\n**Bold text** and *italic text*. [Link](https://example.com). " + strings.Repeat("Additional content to meet minimum length requirements. ", 10),
			maxLen:   50,
			expected: "Bold text and italic text. Link. Additional...",
		},
		{
			name:     "no word boundary available",
			content:  strings.Repeat("a", 10) + " " + strings.Repeat("b", 300),
			maxLen:   15,
			expected: "aaaaaaaaaa...",
		},
		{
			name:     "content exactly at max length",
			content:  strings.Repeat("a", 50) + strings.Repeat(" padding ", 50),
			maxLen:   50,
			expected: strings.Repeat("a", 50) + "...",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := createPostWithContent(tc.content)
			got := p.GetExcerpt(tc.maxLen)

			if got != tc.expected {
				t.Errorf("got %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestPost_IsReadyToPublish(t *testing.T) {
	now := time.Now()

	t.Run("scheduled post ready to publish", func(t *testing.T) {
		// Create a clock that's in the future relative to the scheduled time
		scheduledTime := time.Now()
		futureTime := scheduledTime.Add(1 * time.Hour)
		clock := &mockClock{now: futureTime} // Clock is 1 hour after scheduled time

		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("user-123")
		title, _ := shared.NewTitle("Test Post Title Example")
		content, _ := post.NewPostContent(strings.Repeat("Test content. ", 25))
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		// First create with a clock where scheduled time is in the future
		createClock := &mockClock{now: scheduledTime.Add(-1 * time.Minute)}

		p, err := post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        post.StatusScheduled,
			PublishedAt:   &scheduledTime,
			Category:      cat,
			Clock:         createClock, // Use clock where scheduled time is future
		})

		if err != nil {
			t.Fatalf("failed to create post: %v", err)
		}

		// Now update the post's clock to simulate time passing
		p.Clock = clock

		if !p.IsReadyToPublish() {
			t.Error("expected IsReadyToPublish to be true")
		}
	})

	t.Run("scheduled post not ready", func(t *testing.T) {
		// Create a scheduled post with PublishedAt in the future
		futureTime := now.Add(1 * time.Hour)
		clock := &mockClock{now: now}

		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("user-123")
		title, _ := shared.NewTitle("Test Post Title Example")
		content, _ := post.NewPostContent(strings.Repeat("Test content. ", 25))
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		p, _ := post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        post.StatusScheduled,
			PublishedAt:   &futureTime,
			Category:      cat,
			Clock:         clock,
		})

		if p.IsReadyToPublish() {
			t.Error("expected IsReadyToPublish to be false")
		}
	})

	t.Run("non-scheduled post returns false", func(t *testing.T) {
		clock := &mockClock{now: now}

		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("user-123")
		title, _ := shared.NewTitle("Test Post Title Example")
		content, _ := post.NewPostContent(strings.Repeat("Test content. ", 25))
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		p, _ := post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        post.StatusDraft,
			Category:      cat,
			Clock:         clock,
		})

		if p.IsReadyToPublish() {
			t.Error("expected IsReadyToPublish to be false for non-scheduled post")
		}
	})

	t.Run("scheduled post without PublishedAt returns false", func(t *testing.T) {
		clock := &mockClock{now: now}

		// Create a scheduled post manually to test this edge case
		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("user-123")
		title, _ := shared.NewTitle("Test Post Title Example")
		content, _ := post.NewPostContent(strings.Repeat("Test content. ", 25))
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		p, _ := post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        post.StatusDraft,
			Category:      cat,
			Clock:         clock,
		})

		// Manually set status to scheduled but leave PublishedAt nil
		p.Status = post.StatusScheduled
		p.PublishedAt = nil

		if p.IsReadyToPublish() {
			t.Error("expected IsReadyToPublish to be false for scheduled post without PublishedAt")
		}
	})
}

func TestPost_IsApproved(t *testing.T) {
	clock := &mockClock{now: time.Now()}

	createBasePost := func() post.Post {
		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("user-123")
		title, _ := shared.NewTitle("Test Post Title Example")
		content, _ := post.NewPostContent(strings.Repeat("Test content. ", 25))
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		p, _ := post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        post.StatusDraft,
			Category:      cat,
			Clock:         clock,
		})

		return p
	}

	t.Run("new post is not approved", func(t *testing.T) {
		p := createBasePost()

		if p.IsApproved() {
			t.Error("expected IsApproved to be false for new post")
		}
	})

	t.Run("post with approval is approved", func(t *testing.T) {
		p := createBasePost()
		now := time.Now()
		approverID, _ := kernel.NewID[user.User]("approver-123")

		// Manually set approval fields
		p.ApprovedBy = &approverID
		p.ApprovedAt = &now

		if !p.IsApproved() {
			t.Error("expected IsApproved to be true for approved post")
		}
	})

	t.Run("post with only ApprovedBy is not approved", func(t *testing.T) {
		p := createBasePost()
		approverID, _ := kernel.NewID[user.User]("approver-123")

		// Only set ApprovedBy, not ApprovedAt
		p.ApprovedBy = &approverID

		if p.IsApproved() {
			t.Error("expected IsApproved to be false when only ApprovedBy is set")
		}
	})

	t.Run("post with only ApprovedAt is not approved", func(t *testing.T) {
		p := createBasePost()
		now := time.Now()

		// Only set ApprovedAt, not ApprovedBy
		p.ApprovedAt = &now

		if p.IsApproved() {
			t.Error("expected IsApproved to be false when only ApprovedAt is set")
		}
	})
}

func TestPost_CanBeEditedBy(t *testing.T) {
	clock := &mockClock{now: time.Now()}

	createPost := func(ownerID kernel.ID[user.User]) post.Post {
		postID, _ := kernel.NewID[post.Post]("post-123")
		title, _ := shared.NewTitle("Test Post Title Example")
		content, _ := post.NewPostContent(strings.Repeat("Test content. ", 25))
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		p, _ := post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        post.StatusDraft,
			Category:      cat,
			Clock:         clock,
		})

		return p
	}

	t.Run("admin can edit any post", func(t *testing.T) {
		ownerID, _ := kernel.NewID[user.User]("owner-123")
		adminID, _ := kernel.NewID[user.User]("admin-123")

		p := createPost(ownerID)
		admin := &mockUser{id: adminID, roles: []user.Role{user.RoleAdmin}}

		if !p.CanBeEditedBy(admin) {
			t.Error("expected admin to be able to edit post")
		}
	})

	t.Run("editor can edit any post", func(t *testing.T) {
		ownerID, _ := kernel.NewID[user.User]("owner-123")
		editorID, _ := kernel.NewID[user.User]("editor-123")

		p := createPost(ownerID)
		editor := &mockUser{id: editorID, roles: []user.Role{user.RoleEditor}}

		if !p.CanBeEditedBy(editor) {
			t.Error("expected editor to be able to edit post")
		}
	})

	t.Run("author can edit own post", func(t *testing.T) {
		authorID, _ := kernel.NewID[user.User]("author-123")

		p := createPost(authorID)
		author := &mockUser{id: authorID, roles: []user.Role{user.RoleAuthor}}

		if !p.CanBeEditedBy(author) {
			t.Error("expected author to be able to edit own post")
		}
	})

	t.Run("author cannot edit other's post", func(t *testing.T) {
		ownerID, _ := kernel.NewID[user.User]("owner-123")
		authorID, _ := kernel.NewID[user.User]("author-123")

		p := createPost(ownerID)
		author := &mockUser{id: authorID, roles: []user.Role{user.RoleAuthor}}

		if p.CanBeEditedBy(author) {
			t.Error("expected author not to be able to edit other's post")
		}
	})

	t.Run("visitor cannot edit any post", func(t *testing.T) {
		ownerID, _ := kernel.NewID[user.User]("owner-123")
		visitorID, _ := kernel.NewID[user.User]("visitor-123")

		p := createPost(ownerID)
		visitor := &mockUser{id: visitorID, roles: []user.Role{user.RoleVisitor}}

		if p.CanBeEditedBy(visitor) {
			t.Error("expected visitor not to be able to edit post")
		}
	})
}

func TestPost_CanTransitionTo(t *testing.T) {
	clock := &mockClock{now: time.Now()}

	createPost := func(status post.Status, approved bool) post.Post {
		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("owner-123")
		title, _ := shared.NewTitle("Test Post Title Example")
		content, _ := post.NewPostContent(strings.Repeat("Test content. ", 25))
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		p, _ := post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        status,
			Category:      cat,
			Clock:         clock,
		})

		if approved {
			now := time.Now()
			approverID, _ := kernel.NewID[user.User]("approver-123")
			p.ApprovedBy = &approverID
			p.ApprovedAt = &now
		}

		return p
	}

	t.Run("admin can transition to published when approved", func(t *testing.T) {
		p := createPost(post.StatusDraft, true) // Approved post
		adminID, _ := kernel.NewID[user.User]("admin-123")
		admin := &mockUser{id: adminID, roles: []user.Role{user.RoleAdmin}}

		err := p.CanTransitionTo(post.StatusPublished, admin)

		assertNoError(t, err)
	})

	t.Run("cannot transition to published when not approved", func(t *testing.T) {
		p := createPost(post.StatusDraft, false) // Not approved
		adminID, _ := kernel.NewID[user.User]("admin-123")
		admin := &mockUser{id: adminID, roles: []user.Role{user.RoleAdmin}}

		err := p.CanTransitionTo(post.StatusPublished, admin)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("author cannot publish even when approved", func(t *testing.T) {
		authorID, _ := kernel.NewID[user.User]("author-123")
		p := createPost(post.StatusDraft, true) // Approved post
		p.Owner = authorID                      // Set author as owner
		author := &mockUser{id: authorID, roles: []user.Role{user.RoleAuthor}}

		err := p.CanTransitionTo(post.StatusPublished, author)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EForbidden)
	})

	t.Run("admin can schedule post", func(t *testing.T) {
		p := createPost(post.StatusDraft, false)
		adminID, _ := kernel.NewID[user.User]("admin-123")
		admin := &mockUser{id: adminID, roles: []user.Role{user.RoleAdmin}}

		err := p.CanTransitionTo(post.StatusScheduled, admin)

		assertNoError(t, err)
	})

	t.Run("author cannot schedule post", func(t *testing.T) {
		authorID, _ := kernel.NewID[user.User]("author-123")
		p := createPost(post.StatusDraft, false)
		p.Owner = authorID
		author := &mockUser{id: authorID, roles: []user.Role{user.RoleAuthor}}

		err := p.CanTransitionTo(post.StatusScheduled, author)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EForbidden)
	})

	t.Run("admin can archive post", func(t *testing.T) {
		p := createPost(post.StatusPublished, true)
		adminID, _ := kernel.NewID[user.User]("admin-123")
		admin := &mockUser{id: adminID, roles: []user.Role{user.RoleAdmin}}

		err := p.CanTransitionTo(post.StatusArchived, admin)

		assertNoError(t, err)
	})

	t.Run("author cannot archive post", func(t *testing.T) {
		authorID, _ := kernel.NewID[user.User]("author-123")
		p := createPost(post.StatusPublished, true)
		p.Owner = authorID
		author := &mockUser{id: authorID, roles: []user.Role{user.RoleAuthor}}

		err := p.CanTransitionTo(post.StatusArchived, author)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EForbidden)
	})

	t.Run("invalid status transition", func(t *testing.T) {
		p := createPost(post.StatusDraft, false)
		adminID, _ := kernel.NewID[user.User]("admin-123")
		admin := &mockUser{id: adminID, roles: []user.Role{user.RoleAdmin}}

		// Try to transition from draft to archived (invalid)
		err := p.CanTransitionTo(post.StatusArchived, admin)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})
}

func TestPost_Approve(t *testing.T) {
	clock := &mockClock{now: time.Now()}

	createPost := func(ownerID kernel.ID[user.User]) post.Post {
		postID, _ := kernel.NewID[post.Post]("post-123")
		title, _ := shared.NewTitle("Test Post Title Example")
		content, _ := post.NewPostContent(strings.Repeat("Test content. ", 25))
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		p, _ := post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        post.StatusDraft,
			Category:      cat,
			Clock:         clock,
		})

		return p
	}

	t.Run("admin can approve any post", func(t *testing.T) {
		ownerID, _ := kernel.NewID[user.User]("owner-123")
		adminID, _ := kernel.NewID[user.User]("admin-123")

		p := createPost(ownerID)
		admin := &mockUser{id: adminID, roles: []user.Role{user.RoleAdmin}}

		approved, err := p.Approve(admin)

		assertNoError(t, err)
		if !approved.IsApproved() {
			t.Error("expected post to be approved")
		}
		if approved.ApprovedBy == nil || *approved.ApprovedBy != adminID {
			t.Error("expected ApprovedBy to be set to admin ID")
		}
		if approved.ApprovedAt == nil {
			t.Error("expected ApprovedAt to be set")
		}
	})

	t.Run("editor can approve other's post", func(t *testing.T) {
		ownerID, _ := kernel.NewID[user.User]("owner-123")
		editorID, _ := kernel.NewID[user.User]("editor-123")

		p := createPost(ownerID)
		editor := &mockUser{id: editorID, roles: []user.Role{user.RoleEditor}}

		approved, err := p.Approve(editor)

		assertNoError(t, err)
		if !approved.IsApproved() {
			t.Error("expected post to be approved")
		}
	})

	t.Run("editor cannot approve own post", func(t *testing.T) {
		editorID, _ := kernel.NewID[user.User]("editor-123")

		p := createPost(editorID)
		editor := &mockUser{id: editorID, roles: []user.Role{user.RoleEditor}}

		_, err := p.Approve(editor)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EForbidden)
	})

	t.Run("admin can approve own post", func(t *testing.T) {
		adminID, _ := kernel.NewID[user.User]("admin-123")

		p := createPost(adminID)
		admin := &mockUser{id: adminID, roles: []user.Role{user.RoleAdmin}}

		approved, err := p.Approve(admin)

		assertNoError(t, err)
		if !approved.IsApproved() {
			t.Error("expected post to be approved")
		}
	})

	t.Run("author cannot approve post", func(t *testing.T) {
		authorID, _ := kernel.NewID[user.User]("author-123")

		p := createPost(authorID)
		author := &mockUser{id: authorID, roles: []user.Role{user.RoleAuthor}}

		_, err := p.Approve(author)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EForbidden)
	})
}

func TestPost_Schedule(t *testing.T) {
	clock := &mockClock{now: time.Now()}

	createPost := func() post.Post {
		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("owner-123")
		title, _ := shared.NewTitle("Test Post Title Example")
		content, _ := post.NewPostContent(strings.Repeat("Test content. ", 25))
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		p, _ := post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        post.StatusDraft,
			Category:      cat,
			Clock:         clock,
		})

		return p
	}

	t.Run("admin can schedule post", func(t *testing.T) {
		p := createPost()
		adminID, _ := kernel.NewID[user.User]("admin-123")
		admin := &mockUser{id: adminID, roles: []user.Role{user.RoleAdmin}}
		futureTime := clock.Now().Add(24 * time.Hour)

		scheduled, err := p.Schedule(futureTime, admin)

		assertNoError(t, err)
		if scheduled.Status != post.StatusScheduled {
			t.Errorf("expected status to be scheduled, got %v", scheduled.Status)
		}
		if scheduled.PublishedAt == nil || !scheduled.PublishedAt.Equal(futureTime) {
			t.Error("expected PublishedAt to be set to future time")
		}
	})

	t.Run("cannot schedule with past time", func(t *testing.T) {
		p := createPost()
		adminID, _ := kernel.NewID[user.User]("admin-123")
		admin := &mockUser{id: adminID, roles: []user.Role{user.RoleAdmin}}
		pastTime := clock.Now().Add(-24 * time.Hour)

		_, err := p.Schedule(pastTime, admin)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("author cannot schedule post", func(t *testing.T) {
		p := createPost()
		authorID, _ := kernel.NewID[user.User]("author-123")
		author := &mockUser{id: authorID, roles: []user.Role{user.RoleAuthor}}
		futureTime := clock.Now().Add(24 * time.Hour)

		_, err := p.Schedule(futureTime, author)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EForbidden)
	})
}

func TestPost_Publish(t *testing.T) {
	clock := &mockClock{now: time.Now()}

	createApprovedPost := func() post.Post {
		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("owner-123")
		title, _ := shared.NewTitle("Test Post Title Example")
		content, _ := post.NewPostContent(strings.Repeat("Test content. ", 25))
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		p, _ := post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        post.StatusDraft,
			Category:      cat,
			Clock:         clock,
		})

		// Approve the post
		now := time.Now()
		approverID, _ := kernel.NewID[user.User]("approver-123")
		p.ApprovedBy = &approverID
		p.ApprovedAt = &now

		return p
	}

	t.Run("admin can publish approved post", func(t *testing.T) {
		p := createApprovedPost()
		adminID, _ := kernel.NewID[user.User]("admin-123")
		admin := &mockUser{id: adminID, roles: []user.Role{user.RoleAdmin}}

		published, err := p.Publish(admin)

		assertNoError(t, err)
		if published.Status != post.StatusPublished {
			t.Errorf("expected status to be published, got %v", published.Status)
		}
		if published.PublishedAt == nil {
			t.Error("expected PublishedAt to be set")
		}
		if !published.IsPublished() {
			t.Error("expected post to be published")
		}
	})

	t.Run("editor can publish approved post", func(t *testing.T) {
		p := createApprovedPost()
		editorID, _ := kernel.NewID[user.User]("editor-123")
		editor := &mockUser{id: editorID, roles: []user.Role{user.RoleEditor}}

		published, err := p.Publish(editor)

		assertNoError(t, err)
		if published.Status != post.StatusPublished {
			t.Errorf("expected status to be published, got %v", published.Status)
		}
	})

	t.Run("cannot publish unapproved post", func(t *testing.T) {
		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("owner-123")
		title, _ := shared.NewTitle("Test Post Title Example")
		content, _ := post.NewPostContent(strings.Repeat("Test content. ", 25))
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		p, _ := post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        post.StatusDraft,
			Category:      cat,
			Clock:         clock,
		})

		adminID, _ := kernel.NewID[user.User]("admin-123")
		admin := &mockUser{id: adminID, roles: []user.Role{user.RoleAdmin}}

		_, err := p.Publish(admin)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("author cannot publish post", func(t *testing.T) {
		p := createApprovedPost()
		authorID, _ := kernel.NewID[user.User]("author-123")
		author := &mockUser{id: authorID, roles: []user.Role{user.RoleAuthor}}

		_, err := p.Publish(author)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EForbidden)
	})
}

func TestPost_GetOwner(t *testing.T) {
	clock := &mockClock{now: time.Now()}

	ownerID, _ := kernel.NewID[user.User]("owner-123")
	postID, _ := kernel.NewID[post.Post]("post-123")
	title, _ := shared.NewTitle("Test Post Title Example")
	content, _ := post.NewPostContent(strings.Repeat("Test content. ", 25))
	cat := createTestCategory(t, clock)
	featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

	p, _ := post.NewPost(post.NewPostParams{
		PostID:        postID,
		Owner:         ownerID,
		Title:         title,
		Content:       content,
		FeaturedImage: featuredImage,
		Status:        post.StatusPublished,
		Category:      cat,
		Clock:         clock,
	})

	got := p.GetStatus()

	if got != "published" {
		t.Errorf("expected status %q, got %q", "published", got)
	}
}

func TestPost_ValidateWorkflowFields(t *testing.T) {
	clock := &mockClock{now: time.Now()}

	t.Run("validates ApprovedBy when present", func(t *testing.T) {
		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("owner-123")
		title, _ := shared.NewTitle("Test Post Title Example")
		content, _ := post.NewPostContent(strings.Repeat("Test content. ", 25))
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		p, _ := post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        post.StatusDraft,
			Category:      cat,
			Clock:         clock,
		})

		// Set invalid ApprovedBy
		invalidApproverID := kernel.ID[user.User]("")
		p.ApprovedBy = &invalidApproverID

		err := p.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("scheduled post requires future PublishedAt", func(t *testing.T) {
		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("owner-123")
		title, _ := shared.NewTitle("Test Post Title Example")
		content, _ := post.NewPostContent(strings.Repeat("Test content. ", 25))
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		// Create with valid future time first
		futureTime := clock.Now().Add(24 * time.Hour)
		p, _ := post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        post.StatusScheduled,
			Category:      cat,
			PublishedAt:   &futureTime,
			Clock:         clock,
		})

		// Now modify to have past time
		pastTime := clock.Now().Add(-1 * time.Hour)
		p.PublishedAt = &pastTime

		err := p.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})
}

func TestPost_Validate_ComprehensiveValidation(t *testing.T) {
	clock := &mockClock{now: time.Now()}

	// Create a valid post first
	postID, _ := kernel.NewID[post.Post]("post-123")
	ownerID, _ := kernel.NewID[user.User]("owner-123")
	title, _ := shared.NewTitle("Test Post Title Example")
	content, _ := post.NewPostContent(strings.Repeat("Test content. ", 25))
	cat := createTestCategory(t, clock)
	featuredImage, _ := kernel.NewURL[post.FeaturedImage]("https://example.com/image.jpg")
	seoTitle, _ := shared.NewTitle("SEO Title Example")
	seoDescription, _ := shared.NewDescription("SEO description")
	ogTitle, _ := shared.NewTitle("OG Title Example")
	ogDescription, _ := shared.NewDescription("OG description")
	ogImage, _ := kernel.NewURL[post.OpenGraphImage]("https://example.com/og.jpg")
	canonicalURL, _ := kernel.NewURL[post.Canonical]("https://example.com/canonical")

	validPost, _ := post.NewPost(post.NewPostParams{
		PostID:               postID,
		Owner:                ownerID,
		Title:                title,
		Content:              content,
		FeaturedImage:        featuredImage,
		Status:               post.StatusDraft,
		Category:             cat,
		SEOTitle:             seoTitle,
		SEODescription:       seoDescription,
		OpenGraphTitle:       ogTitle,
		OpenGraphDescription: ogDescription,
		OpenGraphImage:       ogImage,
		CanonicalURL:         canonicalURL,
		SchemaType:           post.SchemaTypeArticle,
		Clock:                clock,
	})

	tests := []struct {
		name     string
		modifier func(*post.Post)
	}{
		{
			name: "invalid featured image URL",
			modifier: func(p *post.Post) {
				p.FeaturedImage = kernel.URL[post.FeaturedImage]("invalid-url")
			},
		},
		{
			name: "invalid SEO description",
			modifier: func(p *post.Post) {
				p.SEODescription = shared.Description(strings.Repeat("a", 301)) // Too long
			},
		},
		{
			name: "invalid Open Graph description",
			modifier: func(p *post.Post) {
				p.OpenGraphDescription = shared.Description(strings.Repeat("a", 301)) // Too long
			},
		},
		{
			name: "invalid OpenGraph image URL",
			modifier: func(p *post.Post) {
				p.OpenGraphImage = kernel.URL[post.OpenGraphImage]("invalid-url")
			},
		},
		{
			name: "invalid canonical URL",
			modifier: func(p *post.Post) {
				p.CanonicalURL = kernel.URL[post.Canonical]("invalid-url")
			},
		},
		{
			name: "invalid schema type",
			modifier: func(p *post.Post) {
				p.SchemaType = post.SchemaType("InvalidSchemaType")
			},
		},
		{
			name: "invalid slug",
			modifier: func(p *post.Post) {
				p.Slug = shared.Slug("Invalid Slug!")
			},
		},
		{
			name: "invalid category",
			modifier: func(p *post.Post) {
				p.Category.CategoryID = kernel.ID[category.Category]("")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the valid post
			testPost := validPost

			// Apply the modification
			tt.modifier(&testPost)

			err := testPost.Validate()

			assertError(t, err)
			assertErrorCode(t, err, kernel.EInvalid)
		})
	}
}

// Additional tests for edge cases and error conditions

func TestPost_EdgeCases(t *testing.T) {
	clock := &mockClock{now: time.Now()}

	t.Run("post with empty optional SEO fields validates", func(t *testing.T) {
		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("owner-123")
		title, _ := shared.NewTitle("Test Post Title Example")
		content, _ := post.NewPostContent(strings.Repeat("Test content. ", 25))
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		p, err := post.NewPost(post.NewPostParams{
			PostID:               postID,
			Owner:                ownerID,
			Title:                title,
			Content:              content,
			FeaturedImage:        featuredImage,
			Status:               post.StatusDraft,
			Category:             cat,
			SEOTitle:             shared.Title(""), // Empty but valid
			SEODescription:       shared.Description(""),
			OpenGraphTitle:       shared.Title(""), // Empty but valid
			OpenGraphDescription: shared.Description(""),
			OpenGraphImage:       kernel.URL[post.OpenGraphImage](""),
			CanonicalURL:         kernel.URL[post.Canonical](""),
			SchemaType:           post.SchemaType(""), // Empty, will use default
			Clock:                clock,
		})

		assertNoError(t, err)

		// Verify empty optional fields are handled correctly
		if p.SEOTitle.String() != "" {
			t.Error("expected empty SEO title to remain empty")
		}
		if p.OpenGraphTitle.String() != "" {
			t.Error("expected empty OG title to remain empty")
		}
	})

	t.Run("post creation with slug generation failure", func(t *testing.T) {
		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("owner-123")
		// Create a title that would result in empty slug (only special characters)
		title := shared.Title("!!!")
		content, _ := post.NewPostContent(strings.Repeat("Test content. ", 25))
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		_, err := post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        post.StatusDraft,
			Category:      cat,
			Clock:         clock,
		})

		assertError(t, err)
		// Should fail during slug generation or validation
	})

	t.Run("word count with empty content after markdown stripping", func(t *testing.T) {
		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("owner-123")
		title, _ := shared.NewTitle("Test Post Title Example")
		// Content with only markdown that gets stripped away, plus padding
		markdownContent := "```\ncode block\n```" + strings.Repeat(" ", 300)
		content, _ := post.NewPostContent(markdownContent)
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		p, _ := post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        post.StatusDraft,
			Category:      cat,
			Clock:         clock,
		})

		wordCount := p.WordCount()
		// Should handle case where content becomes mostly empty after markdown stripping
		if wordCount < 0 {
			t.Error("word count should not be negative")
		}
	})

	t.Run("estimated reading time edge cases", func(t *testing.T) {
		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("user-123")
		title, _ := shared.NewTitle("Test Post Title Example")

		// Test with exactly 1 word (should still be 1 minute minimum)
		content, _ := post.NewPostContent("word" + strings.Repeat(" ", 300)) // Pad to meet minimum
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		p, _ := post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        post.StatusDraft,
			Category:      cat,
			Clock:         clock,
		})

		readingTime := p.EstimatedReadingTime()
		if readingTime < 1 {
			t.Error("estimated reading time should be at least 1 minute")
		}
	})

	t.Run("excerpt with exact word boundary", func(t *testing.T) {
		postID, _ := kernel.NewID[post.Post]("post-123")
		ownerID, _ := kernel.NewID[user.User]("user-123")
		title, _ := shared.NewTitle("Test Post Title Example")

		// Create content where word boundary is exactly at half of maxLength
		testContent := "This is a test sentence that will be truncated exactly at the word boundary point." + strings.Repeat(" padding", 50)
		content, _ := post.NewPostContent(testContent)
		cat := createTestCategory(t, clock)
		featuredImage, _ := kernel.NewURL[post.FeaturedImage]("")

		p, _ := post.NewPost(post.NewPostParams{
			PostID:        postID,
			Owner:         ownerID,
			Title:         title,
			Content:       content,
			FeaturedImage: featuredImage,
			Status:        post.StatusDraft,
			Category:      cat,
			Clock:         clock,
		})

		excerpt := p.GetExcerpt(50)
		if !strings.HasSuffix(excerpt, "...") && len(excerpt) > 50 {
			t.Error("excerpt should be truncated properly")
		}
	})
}
