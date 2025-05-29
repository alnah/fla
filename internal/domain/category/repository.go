package category

import (
	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/shared"
)

// Repository defines essential data operations for category management.
// Provides clean interface between domain logic and data persistence layer.
type Repository interface {
	// Basic CRUD operations for category lifecycle management
	Create(category Category) error
	GetByID(categoryID kernel.ID[Category]) (*Category, error)
	GetAll() ([]Category, error)
	Update(category Category) error
	Delete(categoryID kernel.ID[Category]) error

	// Hierarchy operations for educational content structure
	GetChildren(categoryID kernel.ID[Category]) ([]Category, error)
	GetRootCategories() ([]Category, error)
	BuildPath(categoryID kernel.ID[Category]) (CategoryPath, error)
	FindByPath(pathSegments []string) (*Category, error)

	// Validation support for business rule enforcement
	IsSlugUniqueInParent(slug shared.Slug, parentID *kernel.ID[Category]) (bool, error)
}
