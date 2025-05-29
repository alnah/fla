package category_test

import (
	"strings"
	"testing"
	"time"

	"github.com/alnah/fla/internal/domain/category"
	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/shared"
	"github.com/alnah/fla/internal/domain/user"
)

func TestNewCategoryName(t *testing.T) {
	t.Run("creates category name with valid input", func(t *testing.T) {
		names := []string{
			"A1",
			"Compréhension écrite",
			"Sports",
			"Français Avancé",
			"Café & Culture",
			"a", // minimum length
			strings.Repeat("a", category.MaxCategoryNameLength), // maximum length
		}

		for _, name := range names {
			t.Run(name, func(t *testing.T) {
				got, err := category.NewCategoryName(name)

				assertNoError(t, err)
				if got.String() != name {
					t.Errorf("got %q, want %q", got, name)
				}
			})
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		input := "  A1 Level  "
		want := "A1 Level"

		got, err := category.NewCategoryName(input)

		assertNoError(t, err)
		if got.String() != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("rejects empty name", func(t *testing.T) {
		_, err := category.NewCategoryName("")

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("rejects whitespace only", func(t *testing.T) {
		_, err := category.NewCategoryName("   ")

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("rejects name exceeding max length", func(t *testing.T) {
		longName := strings.Repeat("a", category.MaxCategoryNameLength+1)

		_, err := category.NewCategoryName(longName)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("accepts name at exact max length", func(t *testing.T) {
		exactMaxName := strings.Repeat("a", category.MaxCategoryNameLength)

		got, err := category.NewCategoryName(exactMaxName)

		assertNoError(t, err)
		if len(got.String()) != category.MaxCategoryNameLength {
			t.Errorf("expected length %d, got %d", category.MaxCategoryNameLength, len(got.String()))
		}
	})
}

func TestCategoryName_String(t *testing.T) {
	want := "Test Category"
	name := category.CategoryName(want)

	got := name.String()

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCategoryName_Validate(t *testing.T) {
	t.Run("valid name passes", func(t *testing.T) {
		name := category.CategoryName("Valid Name")

		err := name.Validate()

		assertNoError(t, err)
	})

	t.Run("empty name fails", func(t *testing.T) {
		name := category.CategoryName("")

		err := name.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("too long name fails", func(t *testing.T) {
		name := category.CategoryName(strings.Repeat("a", category.MaxCategoryNameLength+1))

		err := name.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("name with only whitespace fails", func(t *testing.T) {
		name := category.CategoryName("   ")

		err := name.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("minimum length name passes", func(t *testing.T) {
		name := category.CategoryName("a")

		err := name.Validate()

		assertNoError(t, err)
	})

	t.Run("maximum length name passes", func(t *testing.T) {
		name := category.CategoryName(strings.Repeat("a", category.MaxCategoryNameLength))

		err := name.Validate()

		assertNoError(t, err)
	})
}

func TestNewCategory(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	clock := &stubClock{t: fixedTime}

	validCategoryID, _ := kernel.NewID[category.Category]("test-category-id")
	validName, _ := category.NewCategoryName("Test Category")
	validDescription, _ := shared.NewDescription("A test category description")
	validUserID, _ := kernel.NewID[user.User]("user-123")

	t.Run("creates category with minimal required fields", func(t *testing.T) {
		params := category.NewCategoryParams{
			CategoryID: validCategoryID,
			Name:       validName,
			CreatedBy:  validUserID,
			Clock:      clock,
		}

		got, err := category.NewCategory(params)

		assertNoError(t, err)

		if got.CategoryID != validCategoryID {
			t.Errorf("CategoryID: got %v, want %v", got.CategoryID, validCategoryID)
		}
		if got.Name != validName {
			t.Errorf("Name: got %v, want %v", got.Name, validName)
		}
		if got.CreatedBy != validUserID {
			t.Errorf("CreatedBy: got %v, want %v", got.CreatedBy, validUserID)
		}
		if !got.CreatedAt.Equal(fixedTime) {
			t.Errorf("CreatedAt: got %v, want %v", got.CreatedAt, fixedTime)
		}
		if got.ParentID != nil {
			t.Error("ParentID should be nil for root category")
		}
		if got.Description.String() != "" {
			t.Errorf("Description should be empty, got %q", got.Description)
		}
	})

	t.Run("creates category with all fields", func(t *testing.T) {
		parentID, _ := kernel.NewID[category.Category]("parent-category-id")

		params := category.NewCategoryParams{
			CategoryID:  validCategoryID,
			Name:        validName,
			Description: validDescription,
			ParentID:    &parentID,
			CreatedBy:   validUserID,
			Clock:       clock,
		}

		got, err := category.NewCategory(params)

		assertNoError(t, err)

		if got.Description != validDescription {
			t.Errorf("Description: got %v, want %v", got.Description, validDescription)
		}
		if got.ParentID == nil || *got.ParentID != parentID {
			t.Errorf("ParentID: got %v, want %v", got.ParentID, parentID)
		}
	})

	t.Run("automatically generates slug from name", func(t *testing.T) {
		testCases := []struct {
			name         string
			categoryName string
			expectedSlug string
		}{
			{"Simple name", "Simple Name", "simple-name"},
			{"French accents", "Compréhension écrite", "comprehension-ecrite"},
			{"Special characters", "Café & Culture", "cafe-culture"},
			{"Numbers", "A1 Level", "a1-level"},
			{"Multiple spaces", "Test   Category   Name", "test-category-name"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				catName, _ := category.NewCategoryName(tc.categoryName)
				params := category.NewCategoryParams{
					CategoryID: validCategoryID,
					Name:       catName,
					CreatedBy:  validUserID,
					Clock:      clock,
				}

				got, err := category.NewCategory(params)

				assertNoError(t, err)
				if got.Slug.String() != tc.expectedSlug {
					t.Errorf("Slug: got %q, want %q", got.Slug, tc.expectedSlug)
				}
			})
		}
	})

	t.Run("rejects invalid parameters", func(t *testing.T) {
		tests := []struct {
			name   string
			params category.NewCategoryParams
		}{
			{
				name: "empty category ID",
				params: category.NewCategoryParams{
					CategoryID: kernel.ID[category.Category](""),
					Name:       validName,
					CreatedBy:  validUserID,
					Clock:      clock,
				},
			},
			{
				name: "empty name",
				params: category.NewCategoryParams{
					CategoryID: validCategoryID,
					Name:       category.CategoryName(""),
					CreatedBy:  validUserID,
					Clock:      clock,
				},
			},
			{
				name: "empty created by",
				params: category.NewCategoryParams{
					CategoryID: validCategoryID,
					Name:       validName,
					CreatedBy:  kernel.ID[user.User](""),
					Clock:      clock,
				},
			},
			{
				name: "invalid description",
				params: category.NewCategoryParams{
					CategoryID:  validCategoryID,
					Name:        validName,
					Description: shared.Description(strings.Repeat("a", 301)), // exceeds max
					CreatedBy:   validUserID,
					Clock:       clock,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := category.NewCategory(tt.params)

				assertError(t, err)
				assertErrorCode(t, err, kernel.EInvalid)
			})
		}
	})

	t.Run("handles slug generation failure", func(t *testing.T) {
		// Create a name that would result in invalid slug (only special characters)
		invalidName := category.CategoryName("!!!")
		params := category.NewCategoryParams{
			CategoryID: validCategoryID,
			Name:       invalidName,
			CreatedBy:  validUserID,
			Clock:      clock,
		}

		_, err := category.NewCategory(params)

		assertError(t, err)
		// Should fail during slug generation
	})
}

func TestCategory_Validate(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	clock := &stubClock{t: fixedTime}

	validCategoryID, _ := kernel.NewID[category.Category]("test-category-id")
	validName, _ := category.NewCategoryName("Test Category")
	validUserID, _ := kernel.NewID[user.User]("user-123")

	t.Run("valid category passes", func(t *testing.T) {
		cat, _ := category.NewCategory(category.NewCategoryParams{
			CategoryID: validCategoryID,
			Name:       validName,
			CreatedBy:  validUserID,
			Clock:      clock,
		})

		err := cat.Validate()

		assertNoError(t, err)
	})

	t.Run("validates all fields", func(t *testing.T) {
		tests := []struct {
			name     string
			modifier func(*category.Category)
		}{
			{
				name: "empty category ID",
				modifier: func(c *category.Category) {
					c.CategoryID = kernel.ID[category.Category]("")
				},
			},
			{
				name: "empty name",
				modifier: func(c *category.Category) {
					c.Name = category.CategoryName("")
				},
			},
			{
				name: "invalid name",
				modifier: func(c *category.Category) {
					c.Name = category.CategoryName(strings.Repeat("a", category.MaxCategoryNameLength+1))
				},
			},
			{
				name: "empty slug",
				modifier: func(c *category.Category) {
					c.Slug = shared.Slug("")
				},
			},
			{
				name: "invalid slug",
				modifier: func(c *category.Category) {
					c.Slug = shared.Slug("Invalid Slug!")
				},
			},
			{
				name: "invalid description",
				modifier: func(c *category.Category) {
					c.Description = shared.Description(strings.Repeat("a", 301))
				},
			},
			{
				name: "empty created by",
				modifier: func(c *category.Category) {
					c.CreatedBy = kernel.ID[user.User]("")
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Create valid category
				cat, _ := category.NewCategory(category.NewCategoryParams{
					CategoryID: validCategoryID,
					Name:       validName,
					CreatedBy:  validUserID,
					Clock:      clock,
				})

				// Apply modifier to make it invalid
				tt.modifier(&cat)

				err := cat.Validate()

				assertError(t, err)
				assertErrorCode(t, err, kernel.EInvalid)
			})
		}
	})

	t.Run("validates parent ID when present", func(t *testing.T) {
		parentID, _ := kernel.NewID[category.Category]("parent-id")
		cat, _ := category.NewCategory(category.NewCategoryParams{
			CategoryID: validCategoryID,
			Name:       validName,
			ParentID:   &parentID,
			CreatedBy:  validUserID,
			Clock:      clock,
		})

		err := cat.Validate()

		assertNoError(t, err)
	})

	t.Run("rejects invalid parent ID", func(t *testing.T) {
		cat, _ := category.NewCategory(category.NewCategoryParams{
			CategoryID: validCategoryID,
			Name:       validName,
			CreatedBy:  validUserID,
			Clock:      clock,
		})

		// Manually set invalid parent ID
		invalidParentID := kernel.ID[category.Category]("")
		cat.ParentID = &invalidParentID

		err := cat.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("prevents self-referencing parent", func(t *testing.T) {
		cat, _ := category.NewCategory(category.NewCategoryParams{
			CategoryID: validCategoryID,
			Name:       validName,
			CreatedBy:  validUserID,
			Clock:      clock,
		})

		// Set parent ID to same as category ID
		cat.ParentID = &cat.CategoryID

		err := cat.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})
}

func TestCategory_IsRoot(t *testing.T) {
	clock := &stubClock{t: time.Now()}
	validCategoryID, _ := kernel.NewID[category.Category]("test-category-id")
	validName, _ := category.NewCategoryName("Test Category")
	validUserID, _ := kernel.NewID[user.User]("user-123")

	t.Run("returns true for root category", func(t *testing.T) {
		cat, _ := category.NewCategory(category.NewCategoryParams{
			CategoryID: validCategoryID,
			Name:       validName,
			CreatedBy:  validUserID,
			Clock:      clock,
		})

		if !cat.IsRoot() {
			t.Error("expected IsRoot to be true for category without parent")
		}
	})

	t.Run("returns false for child category", func(t *testing.T) {
		parentID, _ := kernel.NewID[category.Category]("parent-id")
		cat, _ := category.NewCategory(category.NewCategoryParams{
			CategoryID: validCategoryID,
			Name:       validName,
			ParentID:   &parentID,
			CreatedBy:  validUserID,
			Clock:      clock,
		})

		if cat.IsRoot() {
			t.Error("expected IsRoot to be false for category with parent")
		}
	})
}

func TestCategory_HasParent(t *testing.T) {
	clock := &stubClock{t: time.Now()}
	validCategoryID, _ := kernel.NewID[category.Category]("test-category-id")
	validName, _ := category.NewCategoryName("Test Category")
	validUserID, _ := kernel.NewID[user.User]("user-123")

	t.Run("returns false for root category", func(t *testing.T) {
		cat, _ := category.NewCategory(category.NewCategoryParams{
			CategoryID: validCategoryID,
			Name:       validName,
			CreatedBy:  validUserID,
			Clock:      clock,
		})

		if cat.HasParent() {
			t.Error("expected HasParent to be false for category without parent")
		}
	})

	t.Run("returns true for child category", func(t *testing.T) {
		parentID, _ := kernel.NewID[category.Category]("parent-id")
		cat, _ := category.NewCategory(category.NewCategoryParams{
			CategoryID: validCategoryID,
			Name:       validName,
			ParentID:   &parentID,
			CreatedBy:  validUserID,
			Clock:      clock,
		})

		if !cat.HasParent() {
			t.Error("expected HasParent to be true for category with parent")
		}
	})
}

func TestCategory_String(t *testing.T) {
	clock := &stubClock{t: time.Now()}
	validCategoryID, _ := kernel.NewID[category.Category]("test-category-id")
	validName, _ := category.NewCategoryName("Test Category")
	validUserID, _ := kernel.NewID[user.User]("user-123")

	t.Run("string representation for root category", func(t *testing.T) {
		cat, _ := category.NewCategory(category.NewCategoryParams{
			CategoryID: validCategoryID,
			Name:       validName,
			CreatedBy:  validUserID,
			Clock:      clock,
		})

		got := cat.String()
		expected := `Category{ID: "test-category-id", Name: "Test Category", Slug: "test-category", Root: true}`

		if got != expected {
			t.Errorf("got %q, want %q", got, expected)
		}
	})

	t.Run("string representation for child category", func(t *testing.T) {
		parentID, _ := kernel.NewID[category.Category]("parent-id")
		cat, _ := category.NewCategory(category.NewCategoryParams{
			CategoryID: validCategoryID,
			Name:       validName,
			ParentID:   &parentID,
			CreatedBy:  validUserID,
			Clock:      clock,
		})

		got := cat.String()
		expected := `Category{ID: "test-category-id", Name: "Test Category", Slug: "test-category", Parent: "parent-id"}`

		if got != expected {
			t.Errorf("got %q, want %q", got, expected)
		}
	})
}

func TestCategoryConstants(t *testing.T) {
	t.Run("max category depth constant", func(t *testing.T) {
		if category.MaxCategoryDepth != 3 {
			t.Errorf("MaxCategoryDepth: got %d, want %d", category.MaxCategoryDepth, 3)
		}
	})

	t.Run("category name length constants", func(t *testing.T) {
		if category.MinCategoryNameLength != 1 {
			t.Errorf("MinCategoryNameLength: got %d, want %d", category.MinCategoryNameLength, 1)
		}
		if category.MaxCategoryNameLength != 100 {
			t.Errorf("MaxCategoryNameLength: got %d, want %d", category.MaxCategoryNameLength, 100)
		}
	})
}

func TestCategoryErrorMessages(t *testing.T) {
	t.Run("error message constants", func(t *testing.T) {
		tests := []struct {
			name     string
			constant string
			expected string
		}{
			{"circular reference", category.MCategoryCircularReference, "Category cannot be its own parent."},
			{"max depth exceeded", category.MCategoryMaxDepthExceeded, "Category hierarchy cannot exceed 3 levels deep."},
			{"name not unique", category.MCategoryNameNotUnique, "Category name must be unique within parent."},
			{"slug not unique", category.MCategorySlugNotUnique, "Category slug must be unique within parent."},
			{"name missing", category.MCategoryNameMissing, "Missing category name."},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.constant != tt.expected {
					t.Errorf("got %q, want %q", tt.constant, tt.expected)
				}
			})
		}
	})
}
