package post

import (
	"fmt"

	"github.com/alnah/fla/internal/domain/shared"
)

// PostsList combines paginated posts with navigation metadata for content browsing.
// Enables efficient content listing with proper page controls and item counts.
type PostsList struct {
	Posts      []Post
	Pagination shared.Pagination
}

// NewPostsList creates a new paginated posts list
func NewPostsList(posts []Post, pagination shared.Pagination) PostsList {
	// Create a new slice to avoid potential issues with the original slice
	postsCopy := make([]Post, len(posts))
	copy(postsCopy, posts)

	return PostsList{
		Posts:      postsCopy,
		Pagination: pagination,
	}
}

// IsEmpty returns true if the list has no posts
func (pl PostsList) IsEmpty() bool {
	return len(pl.Posts) == 0
}

// Count returns the number of posts in current page
func (pl PostsList) Count() int {
	return len(pl.Posts)
}

// String returns a string representation of the posts list
func (pl PostsList) String() string {
	return fmt.Sprintf("PostsList{Count: %d, %s}",
		len(pl.Posts), pl.Pagination.String())
}
