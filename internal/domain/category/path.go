package category

import "strings"

// CategoryPath represents the complete hierarchy trail from root to target category.
// Enables URL generation and breadcrumb navigation for educational content structure.
type CategoryPath []Category

// String generates URL-friendly path representation for web routing.
// Creates clean URLs like "a1/comprehension-ecrite/sports" for SEO and usability.
func (cp CategoryPath) String() string {
	if len(cp) == 0 {
		return ""
	}

	segments := make([]string, len(cp))
	for i, category := range cp {
		segments[i] = category.Slug.String()
	}

	return strings.Join(segments, "/")
}

// Depth calculates hierarchy level for validation and display purposes.
// Enables depth-based restrictions and navigation level awareness.
func (cp CategoryPath) Depth() int {
	return len(cp) - 1
}

// IsValidDepth verifies path stays within configured hierarchy limits.
// Prevents overly deep category structures that complicate navigation.
func (cp CategoryPath) IsValidDepth() bool {
	return cp.Depth() <= MaxCategoryDepth-1
}

// Leaf returns the final category in the path for content association.
// Identifies the most specific category for post categorization.
func (cp CategoryPath) Leaf() *Category {
	if len(cp) == 0 {
		return nil
	}
	return &cp[len(cp)-1]
}

// CategoryBreadcrumb represents navigation trail elements for hierarchical browsing.
// Enables users to understand their location and navigate back through category levels.
type CategoryBreadcrumb struct {
	Category Category
	IsLast   bool // True if this is the last item in breadcrumb
	Level    int  // 0-based level in hierarchy (0=A1, 1=Compréhension écrite, 2=Sports)
}
