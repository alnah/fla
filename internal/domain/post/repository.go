package post

import (
	"github.com/alnah/fla/internal/domain/category"
	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/shared"
	"github.com/alnah/fla/internal/domain/tag"
	"github.com/alnah/fla/internal/domain/user"
)

// Repository defines essential data operations for post management.
// Provides clean interface between domain logic and data persistence layer.
type Repository interface {
	// Basic CRUD operations for post lifecycle management
	Create(post Post) error
	GetByID(postID kernel.ID[Post]) (*Post, error)
	GetBySlug(slug shared.Slug) (*Post, error)
	Update(post Post) error
	Delete(postID kernel.ID[Post]) error

	// Query operations for content discovery
	GetPublishedPosts(pagination shared.Pagination) (PostsList, error)
	GetPostsByCategory(categoryID kernel.ID[category.Category], pagination shared.Pagination) (PostsList, error)
	GetPostsByTag(tagID kernel.ID[tag.Tag], pagination shared.Pagination) (PostsList, error)
	GetPostsByAuthor(authorID kernel.ID[user.User], pagination shared.Pagination) (PostsList, error)
	GetScheduledPosts() ([]Post, error)

	// Search and filtering
	Search(query string, pagination shared.Pagination) (PostsList, error)
	GetRelatedPosts(postID kernel.ID[Post], limit int) ([]Post, error)

	// Validation support
	IsSlugUnique(slug shared.Slug, excludeID *kernel.ID[Post]) (bool, error)
}
