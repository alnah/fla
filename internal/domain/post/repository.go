package post

import (
	"github.com/alnah/fla/internal/domain/category"
	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/shared"
	"github.com/alnah/fla/internal/domain/tag"
	"github.com/alnah/fla/internal/domain/user"
)

// PostReader defines read-only operations for content consumption.
// Used by public website pages and content display systems.
type PostReader interface {
	// GetByID retrieves a specific post for editing, moderation, or display.
	// Used by admin interfaces, edit forms, and content management systems.
	GetByID(postID kernel.ID[Post]) (*Post, error)

	// GetBySlug finds posts by URL-friendly identifiers for public web pages.
	// Used by website routing to serve individual blog posts to visitors.
	GetBySlug(slug shared.Slug) (*Post, error)
}

// PostWriter defines write operations for content management.
// Used by content creation and editing workflows.
type PostWriter interface {
	// Create persists a new post to enable content publishing workflows.
	// Used when authors create drafts or import content from external sources.
	Create(post Post) error

	// Update saves changes to existing posts for content revision workflows.
	// Used when authors edit drafts, moderators approve content, or scheduled posts auto-publish.
	Update(post Post) error

	// Delete removes posts permanently for content cleanup and spam removal.
	// Used by admin tools and automated content moderation systems.
	Delete(postID kernel.ID[Post]) error
}

// PostLister provides paginated content browsing for public consumption.
// Used by website pages that show multiple posts to visitors.
type PostLister interface {
	// GetPublishedPosts returns paginated live content for public website display.
	// Used by homepage, blog listings, and RSS feeds to serve content to visitors.
	GetPublishedPosts(pagination shared.Pagination) (PostsList, error)

	// GetPostsByCategory filters content by learning topic for organized browsing.
	// Used by category pages (A1/Reading/Sports) to help learners find relevant content.
	GetPostsByCategory(categoryID kernel.ID[category.Category], pagination shared.Pagination) (PostsList, error)

	// GetPostsByTag finds related content across categories for cross-topic discovery.
	// Used by tag pages and "related posts" features to connect similar learning materials.
	GetPostsByTag(tagID kernel.ID[tag.Tag], pagination shared.Pagination) (PostsList, error)

	// GetPostsByAuthor returns content from specific writers for author profile pages.
	// Used by author bio pages and contributor portfolios in multi-author blogs.
	GetPostsByAuthor(authorID kernel.ID[user.User], pagination shared.Pagination) (PostsList, error)
}

// PostSearcher handles content discovery through queries.
// Used by search functionality and content recommendation systems.
type PostSearcher interface {
	// Search finds posts matching user queries for content discovery.
	// Used by site search functionality to help visitors find specific learning topics.
	Search(query string, pagination shared.Pagination) (PostsList, error)

	// GetRelatedPosts suggests similar content to keep readers engaged.
	// Used by individual post pages to recommend additional relevant learning materials.
	GetRelatedPosts(postID kernel.ID[Post], limit int) ([]Post, error)
}

// PostScheduler manages time-based publishing workflows.
// Used by background jobs and automated publishing systems.
type PostScheduler interface {
	// GetScheduledPosts returns all posts queued for future publication.
	// Used by background jobs to automatically publish content at scheduled times.
	GetScheduledPosts() ([]Post, error)
}

// PostValidator provides data integrity checks for content creation.
// Used by forms and APIs to prevent duplicate or invalid content.
type PostValidator interface {
	// IsSlugUnique prevents URL conflicts when creating or updating posts.
	// Used by content creation forms to ensure each post has a unique web address.
	IsSlugUnique(slug shared.Slug, excludeID *kernel.ID[Post]) (bool, error)
}

// Composed interfaces for common use cases

// PostManager combines read/write operations for content management systems.
// Used by admin dashboards and CMS interfaces that need full post control.
type PostManager interface {
	PostReader
	PostWriter
	PostValidator
}

// PostBrowser combines listing and search for public content discovery.
// Used by public website features that help visitors find and browse content.
type PostBrowser interface {
	PostReader
	PostLister
	PostSearcher
}

// PostPublisher handles the complete publishing workflow.
// Used by editorial systems and automated publishing processes.
type PostPublisher interface {
	PostReader
	PostWriter
	PostScheduler
	PostValidator
}

// Full repository interface for implementations that provide everything.
// Most concrete implementations (like PostgresPostRepository) will implement this.
type Repository interface {
	PostReader
	PostWriter
	PostLister
	PostSearcher
	PostScheduler
	PostValidator
}
