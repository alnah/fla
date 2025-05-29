package post_test

import (
	"strings"
	"testing"
	"time"

	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/post"
	"github.com/alnah/fla/internal/domain/shared"
	"github.com/alnah/fla/internal/domain/user"
)

func TestNewPostsList(t *testing.T) {
	clock := &mockClock{now: time.Now()}

	// Helper to create a test post
	createTestPost := func(id string) post.Post {
		postID, _ := kernel.NewID[post.Post](id)
		ownerID, _ := kernel.NewID[user.User]("user-123")
		title, _ := shared.NewTitle("Test Post Title")
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

	t.Run("creates posts list with nil slice", func(t *testing.T) {
		pagination, _ := shared.NewPagination(1, 10, 0)

		list := post.NewPostsList(nil, pagination)

		if list.Posts == nil {
			t.Error("expected Posts to be non-nil slice")
		}

		if len(list.Posts) != 0 {
			t.Errorf("expected empty slice, got %d posts", len(list.Posts))
		}

		if !list.IsEmpty() {
			t.Error("expected IsEmpty to be true")
		}

		if list.Count() != 0 {
			t.Errorf("expected Count to be 0, got %d", list.Count())
		}
	})

	t.Run("creates posts list with empty slice", func(t *testing.T) {
		pagination, _ := shared.NewPagination(1, 10, 0)
		posts := []post.Post{}

		list := post.NewPostsList(posts, pagination)

		if list.Posts == nil {
			t.Error("expected Posts to be non-nil")
		}

		if len(list.Posts) != 0 {
			t.Errorf("expected empty slice, got %d posts", len(list.Posts))
		}

		if !list.IsEmpty() {
			t.Error("expected IsEmpty to be true")
		}
	})

	t.Run("creates posts list with posts", func(t *testing.T) {
		posts := []post.Post{
			createTestPost("post-1"),
			createTestPost("post-2"),
			createTestPost("post-3"),
		}
		pagination, _ := shared.NewPagination(1, 10, 3)

		list := post.NewPostsList(posts, pagination)

		if len(list.Posts) != 3 {
			t.Errorf("expected 3 posts, got %d", len(list.Posts))
		}

		if list.IsEmpty() {
			t.Error("expected IsEmpty to be false")
		}

		if list.Count() != 3 {
			t.Errorf("expected Count to be 3, got %d", list.Count())
		}
	})

	t.Run("string representation", func(t *testing.T) {
		posts := []post.Post{
			createTestPost("post-1"),
			createTestPost("post-2"),
		}
		pagination, _ := shared.NewPagination(2, 10, 25)

		list := post.NewPostsList(posts, pagination)

		str := list.String()
		expectedStr := "PostsList{Count: 2, Pagination{Page: 2, Limit: 10, TotalItems: 25, TotalPages: 3}}"

		if str != expectedStr {
			t.Errorf("expected string %q, got %q", expectedStr, str)
		}
	})
}
