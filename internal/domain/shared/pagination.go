package shared

import (
	"fmt"
	"math"

	"github.com/alnah/fla/internal/domain/kernel"
)

const (
	MPaginationInvalidPage  string = "Page number must be greater than 0."
	MPaginationInvalidLimit string = "Limit must be between %d and %d."
	MPaginationInvalidTotal string = "Total items cannot be negative."
)

const (
	MinPageLimit     = 1
	MaxPageLimit     = 100
	DefaultPageLimit = 10
)

// Pagination handles content listing with page-based navigation for improved user experience.
// Provides offset calculations and navigation state for repository queries.
type Pagination struct {
	Page       int // Current page (1-based)
	Limit      int // Items per page
	TotalItems int // Total number of items
	TotalPages int // Total number of pages (calculated)
}

// NewPagination creates a new pagination with validation
func NewPagination(page, limit, totalItems int) (Pagination, error) {
	const op = "NewPagination"

	// Use default limit if not provided
	if limit <= 0 {
		limit = DefaultPageLimit
	}

	// Use first page if not provided
	if page <= 0 {
		page = 1
	}

	pagination := Pagination{
		Page:       page,
		Limit:      limit,
		TotalItems: totalItems,
		TotalPages: calculateTotalPages(totalItems, limit),
	}

	if err := pagination.Validate(); err != nil {
		return Pagination{}, &kernel.Error{Operation: op, Cause: err}
	}

	return pagination, nil
}

// Validate performs validation on pagination parameters
func (p Pagination) Validate() error {
	const op = "Pagination.Validate"

	if err := p.validatePage(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := p.validateLimit(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := p.validateTotalItems(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	return nil
}

// validatePage checks if page number is valid
func (p Pagination) validatePage() error {
	const op = "Pagination.validatePage"

	if p.Page < 1 {
		return &kernel.Error{
			Code:      kernel.EInvalid,
			Message:   MPaginationInvalidPage,
			Operation: op,
		}
	}

	return nil
}

// validateLimit checks if limit is within allowed range
func (p Pagination) validateLimit() error {
	const op = "Pagination.validateLimit"

	if p.Limit < MinPageLimit || p.Limit > MaxPageLimit {
		return &kernel.Error{
			Code:      kernel.EInvalid,
			Message:   fmt.Sprintf(MPaginationInvalidLimit, MinPageLimit, MaxPageLimit),
			Operation: op,
		}
	}

	return nil
}

// validateTotalItems checks if total items is not negative
func (p Pagination) validateTotalItems() error {
	const op = "Pagination.validateTotalItems"

	if p.TotalItems < 0 {
		return &kernel.Error{
			Code:      kernel.EInvalid,
			Message:   MPaginationInvalidTotal,
			Operation: op,
		}
	}

	return nil
}

// String returns a string representation of the pagination
func (p Pagination) String() string {
	return fmt.Sprintf("Pagination{Page: %d, Limit: %d, TotalItems: %d, TotalPages: %d}",
		p.Page, p.Limit, p.TotalItems, p.TotalPages)
}

// Offset returns the offset for database queries (0-based)
func (p Pagination) Offset() int {
	return (p.Page - 1) * p.Limit
}

// HasNextPage returns true if there's a next page
func (p Pagination) HasNextPage() bool {
	return p.Page < p.TotalPages
}

// HasPreviousPage returns true if there's a previous page
func (p Pagination) HasPreviousPage() bool {
	return p.Page > 1
}

// NextPage returns the next page number (or current if last page)
func (p Pagination) NextPage() int {
	if p.HasNextPage() {
		return p.Page + 1
	}
	return p.Page
}

// PreviousPage returns the previous page number (or 1 if first page)
func (p Pagination) PreviousPage() int {
	if p.HasPreviousPage() {
		return p.Page - 1
	}
	return 1
}

// IsEmpty returns true if there are no items
func (p Pagination) IsEmpty() bool {
	return p.TotalItems == 0
}

// IsLastPage returns true if this is the last page
func (p Pagination) IsLastPage() bool {
	return p.Page >= p.TotalPages
}

// IsFirstPage returns true if this is the first page
func (p Pagination) IsFirstPage() bool {
	return p.Page == 1
}

// ItemsOnCurrentPage returns the number of items on the current page
func (p Pagination) ItemsOnCurrentPage() int {
	if p.IsEmpty() {
		return 0
	}

	if p.IsLastPage() {
		// Last page might have fewer items
		remaining := p.TotalItems % p.Limit
		if remaining == 0 {
			return p.Limit
		}
		return remaining
	}

	return p.Limit
}

// StartItem returns the number of the first item on current page (1-based)
func (p Pagination) StartItem() int {
	if p.IsEmpty() {
		return 0
	}
	return p.Offset() + 1
}

// EndItem returns the number of the last item on current page (1-based)
func (p Pagination) EndItem() int {
	if p.IsEmpty() {
		return 0
	}
	return p.StartItem() + p.ItemsOnCurrentPage() - 1
}

// calculateTotalPages calculates total number of pages
func calculateTotalPages(totalItems, limit int) int {
	if totalItems <= 0 || limit <= 0 {
		return 0
	}
	return int(math.Ceil(float64(totalItems) / float64(limit)))
}
