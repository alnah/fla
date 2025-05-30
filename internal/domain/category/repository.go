package category

import (
	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/shared"
)

// CategoryReader defines read-only operations for category access.
// Used by public navigation menus and content display systems.
type CategoryReader interface {
	// GetByID retrieves a specific category for content organization and display.
	// Used by post creation forms and category management interfaces.
	GetByID(categoryID kernel.ID[Category]) (*Category, error)

	// GetAll returns complete category catalog for admin overview and site maps.
	// Used by administration panels and SEO sitemap generation.
	GetAll() ([]Category, error)
}

// CategoryWriter defines modification operations for category management.
// Used by content management systems and category administration tools.
type CategoryWriter interface {
	// Create establishes new categories to expand content organization structure.
	// Used when adding new learning topics or reorganizing educational content.
	Create(category Category) error

	// Update modifies existing categories for content structure maintenance.
	// Used when renaming topics or updating category descriptions.
	Update(category Category) error

	// Delete removes obsolete categories for content organization cleanup.
	// Used by admin tools to eliminate unused or redundant categorization.
	Delete(categoryID kernel.ID[Category]) error
}

// CategoryHierarchy manages parent-child relationships for educational content structure.
// Used by navigation systems and content organization features.
type CategoryHierarchy interface {
	// GetChildren finds subcategories for hierarchical content browsing.
	// Used by navigation menus to show topic breakdowns (A1 → Reading, Writing).
	GetChildren(categoryID kernel.ID[Category]) ([]Category, error)

	// GetRootCategories returns top-level learning categories for main navigation.
	// Used by homepage menus and primary content organization (A1, A2, B1 levels).
	GetRootCategories() ([]Category, error)
}

// CategoryPathBuilder creates URL paths for hierarchical navigation.
// Used by URL routing and breadcrumb generation systems.
type CategoryPathBuilder interface {
	// BuildPath creates hierarchical URL structure for SEO-friendly navigation.
	// Used by URL generators to create paths like "a1/reading/sports" for web routing.
	BuildPath(categoryID kernel.ID[Category]) (CategoryPath, error)

	// FindByPath locates categories from URL segments for request routing.
	// Used by web routers to map URLs like "/a1/reading/sports" to category content.
	FindByPath(pathSegments []string) (*Category, error)
}

// CategoryValidator provides data integrity checks for category creation.
// Used by forms and APIs to prevent duplicate or invalid category structures.
type CategoryValidator interface {
	// IsSlugUniqueInParent prevents URL conflicts within the same category level.
	// Used by category creation forms to ensure unique names within parent categories.
	IsSlugUniqueInParent(slug shared.Slug, parentID *kernel.ID[Category]) (bool, error)
}

// Composed interfaces for common use cases

// CategoryBrowser combines reading and hierarchy for public navigation.
// Used by public website features that display category structure to visitors.
type CategoryBrowser interface {
	CategoryReader
	CategoryHierarchy
}

// CategoryManager combines CRUD and validation for administrative control.
// Used by admin panels and CMS interfaces that manage category structure.
type CategoryManager interface {
	CategoryReader
	CategoryWriter
	CategoryValidator
}

// CategoryNavigator handles complete navigation and URL management.
// Used by web routing systems and navigation menu generators.
type CategoryNavigator interface {
	CategoryReader
	CategoryHierarchy
	CategoryPathBuilder
}

// CategoryOrganizer provides full category structure management.
// Used by content management systems that need complete category control.
type CategoryOrganizer interface {
	CategoryReader
	CategoryWriter
	CategoryHierarchy
	CategoryValidator
}

// Full repository interface for implementations that provide everything.
// Most concrete implementations (like PostgresCategoryRepository) will implement this.
type Repository interface {
	CategoryReader
	CategoryWriter
	CategoryHierarchy
	CategoryPathBuilder
	CategoryValidator
}
