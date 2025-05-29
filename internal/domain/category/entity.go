package category

import (
	"fmt"
	"strings"
	"time"

	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/shared"
	"github.com/alnah/fla/internal/domain/user"
)

const (
	MCategoryCircularReference string = "Category cannot be its own parent."
	MCategoryMaxDepthExceeded  string = "Category hierarchy cannot exceed 3 levels deep."
	MCategoryNameNotUnique     string = "Category name must be unique within parent."
	MCategorySlugNotUnique     string = "Category slug must be unique within parent."
)

const MaxCategoryDepth = 3

// CategoryName represents user-facing category titles with length validation.
// Ensures category names are meaningful and fit within UI constraints.
type CategoryName string

const (
	MCategoryNameMissing  string = "Missing category name."
	MinCategoryNameLength int    = 1
	MaxCategoryNameLength int    = 100
)

// NewCategoryName creates a validated category name with proper length limits.
// Maintains consistent category naming and prevents UI layout issues.
func NewCategoryName(name string) (CategoryName, error) {
	const op = "NewCategoryName"

	c := CategoryName(strings.TrimSpace(name))
	if err := c.Validate(); err != nil {
		return "", &kernel.Error{Operation: op, Cause: err}
	}

	return c, nil
}

func (c CategoryName) String() string { return string(c) }

// Validate enforces category name requirements for display and consistency.
// Ensures names are present and within acceptable length boundaries.
func (c CategoryName) Validate() error {
	const op = "CategoryName.Validate"

	if err := kernel.ValidatePresence("category name", c.String(), op); err != nil {
		return err
	}

	if err := kernel.ValidateLength("category name", c.String(), MinCategoryNameLength, MaxCategoryNameLength, op); err != nil {
		return err
	}

	return nil
}

// Category represents a hierarchical content organization unit for educational blogs.
// Categories enable structured navigation through learning materials (Level → Skill → Topic).
type Category struct {
	// Identity
	CategoryID kernel.ID[Category]

	// Data
	Name        CategoryName
	Slug        shared.Slug
	Description shared.Description // Optional explanation of the category

	// Hierarchy
	ParentID *kernel.ID[Category] // nil for root categories

	// Meta
	CreatedBy kernel.ID[user.User]
	CreatedAt time.Time

	// DI
	Clock kernel.Clock
}

// NewCategoryParams holds the essential information needed to create a learning category.
// Used to ensure all required fields are provided during category creation.
type NewCategoryParams struct {
	// Required
	CategoryID kernel.ID[Category]
	Name       CategoryName
	CreatedBy  kernel.ID[user.User]

	// Optional
	Description shared.Description
	ParentID    *kernel.ID[Category] // nil for root categories

	// DI
	Clock kernel.Clock
}

// NewCategory creates a validated category with automatic slug generation.
// Ensures category hierarchy rules and data integrity are maintained.
func NewCategory(params NewCategoryParams) (Category, error) {
	const op = "NewCategory"

	now := params.Clock.Now()

	slug, err := shared.NewSlug(params.Name.String())
	if err != nil {
		return Category{}, &kernel.Error{Operation: op, Cause: err}
	}

	category := Category{
		CategoryID:  params.CategoryID,
		Name:        params.Name,
		Slug:        slug,
		Description: params.Description,
		ParentID:    params.ParentID,
		CreatedBy:   params.CreatedBy,
		CreatedAt:   now,
		Clock:       params.Clock,
	}

	if err := category.Validate(); err != nil {
		return Category{}, &kernel.Error{Operation: op, Cause: err}
	}

	return category, nil
}

// Validate enforces business rules for category data consistency and hierarchy constraints.
// Prevents invalid category structures that would break navigation or URL generation.
func (c Category) Validate() error {
	const op = "Category.Validate"

	if err := c.CategoryID.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := c.Name.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := c.Slug.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := c.Description.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := c.CreatedBy.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := c.validateBasicHierarchy(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	return nil
}

// validateBasicHierarchy performs basic hierarchy validation
func (c Category) validateBasicHierarchy() error {
	const op = "Category.validateBasicHierarchy"

	if c.ParentID != nil {
		if err := c.ParentID.Validate(); err != nil {
			return &kernel.Error{Operation: op, Cause: err}
		}

		// Cannot be its own parent (basic circular reference check)
		if *c.ParentID == c.CategoryID {
			return &kernel.Error{
				Code:      kernel.EInvalid,
				Message:   MCategoryCircularReference,
				Operation: op,
			}
		}
	}

	return nil
}

// IsRoot determines if this category is a top-level learning category (A1, A2, etc.).
// Root categories serve as main entry points in the educational hierarchy.
func (c Category) IsRoot() bool {
	return c.ParentID == nil
}

// HasParent indicates if this category belongs to a parent category in the hierarchy.
// Used to determine navigation structure and URL path generation.
func (c Category) HasParent() bool {
	return c.ParentID != nil
}

// String returns a string representation of the category
func (c Category) String() string {
	if c.ParentID == nil {
		return fmt.Sprintf("Category{ID: %q, Name: %q, Slug: %q, Root: true}",
			c.CategoryID, c.Name, c.Slug)
	}

	return fmt.Sprintf("Category{ID: %q, Name: %q, Slug: %q, Parent: %q}",
		c.CategoryID, c.Name, c.Slug, *c.ParentID)
}
