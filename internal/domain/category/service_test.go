package category_test

import (
	"fmt"
	"testing"

	"github.com/alnah/fla/internal/domain/category"
	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/shared"
)

// Mock repository for testing
type mockRepository struct {
	categories     map[string]category.Category
	paths          map[string]category.CategoryPath
	buildPathFunc  func(kernel.ID[category.Category]) (category.CategoryPath, error)
	findByPathFunc func([]string) (*category.Category, error)
}

func (m *mockRepository) Create(cat category.Category) error {
	return nil
}

func (m *mockRepository) GetByID(catID kernel.ID[category.Category]) (*category.Category, error) {
	if cat, ok := m.categories[catID.String()]; ok {
		return &cat, nil
	}
	return nil, &kernel.Error{Code: kernel.ENotFound, Message: "category not found"}
}

func (m *mockRepository) GetAll() ([]category.Category, error) {
	return nil, nil
}

func (m *mockRepository) Update(cat category.Category) error {
	return nil
}

func (m *mockRepository) Delete(catID kernel.ID[category.Category]) error {
	return nil
}

func (m *mockRepository) GetChildren(catID kernel.ID[category.Category]) ([]category.Category, error) {
	return nil, nil
}

func (m *mockRepository) GetRootCategories() ([]category.Category, error) {
	return nil, nil
}

func (m *mockRepository) BuildPath(catID kernel.ID[category.Category]) (category.CategoryPath, error) {
	if m.buildPathFunc != nil {
		return m.buildPathFunc(catID)
	}
	if path, ok := m.paths[catID.String()]; ok {
		return path, nil
	}
	return nil, &kernel.Error{Code: kernel.ENotFound, Message: "category not found"}
}

func (m *mockRepository) FindByPath(pathSegments []string) (*category.Category, error) {
	if m.findByPathFunc != nil {
		return m.findByPathFunc(pathSegments)
	}
	return nil, &kernel.Error{Code: kernel.ENotFound, Message: "category not found"}
}

func (m *mockRepository) IsSlugUniqueInParent(slug shared.Slug, parentID *kernel.ID[category.Category]) (bool, error) {
	return true, nil
}

func TestPathService_BuildURL(t *testing.T) {
	t.Run("builds URL for single category", func(t *testing.T) {
		cat := createTestCategory("a1", "A1", nil)
		repo := &mockRepository{
			paths: map[string]category.CategoryPath{
				"a1": {cat},
			},
		}
		service := category.NewPathService(repo)

		catID, _ := kernel.NewID[category.Category]("a1")
		got, err := service.BuildURL(catID)

		assertNoError(t, err)
		want := "a1"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("builds URL for nested categories", func(t *testing.T) {
		a1ID := "a1"
		readingID := "reading"

		a1 := createTestCategory("a1", "A1", nil)
		reading := createTestCategory("reading", "Compréhension écrite", &a1ID)
		sports := createTestCategory("sports", "Sports", &readingID)

		repo := &mockRepository{
			paths: map[string]category.CategoryPath{
				"sports": {a1, reading, sports},
			},
		}
		service := category.NewPathService(repo)

		catID, _ := kernel.NewID[category.Category]("sports")
		got, err := service.BuildURL(catID)

		assertNoError(t, err)
		want := "a1/comprehension-ecrite/sports"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("returns error for non-existent category", func(t *testing.T) {
		repo := &mockRepository{
			paths: map[string]category.CategoryPath{},
		}
		service := category.NewPathService(repo)

		catID, _ := kernel.NewID[category.Category]("non-existent")
		_, err := service.BuildURL(catID)

		assertError(t, err)
		assertErrorCode(t, err, kernel.ENotFound)
	})

	t.Run("propagates repository errors", func(t *testing.T) {
		repo := &mockRepository{
			buildPathFunc: func(id kernel.ID[category.Category]) (category.CategoryPath, error) {
				return nil, &kernel.Error{Code: kernel.EInternal, Message: "database error"}
			},
		}
		service := category.NewPathService(repo)

		catID, _ := kernel.NewID[category.Category]("cat")
		_, err := service.BuildURL(catID)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInternal)
	})
}

func TestPathService_ParseURL(t *testing.T) {
	t.Run("parses single segment URL", func(t *testing.T) {
		cat := createTestCategory("a1", "A1", nil)
		repo := &mockRepository{
			findByPathFunc: func(segments []string) (*category.Category, error) {
				if len(segments) == 1 && segments[0] == "a1" {
					return &cat, nil
				}
				return nil, &kernel.Error{Code: kernel.ENotFound}
			},
		}
		service := category.NewPathService(repo)

		got, err := service.ParseURL("a1")

		assertNoError(t, err)
		if got.CategoryID != cat.CategoryID {
			t.Errorf("got %v, want %v", got.CategoryID, cat.CategoryID)
		}
	})

	t.Run("parses multi-segment URL", func(t *testing.T) {
		sports := createTestCategory("sports", "Sports", nil)
		repo := &mockRepository{
			findByPathFunc: func(segments []string) (*category.Category, error) {
				expected := []string{"a1", "comprehension-ecrite", "sports"}
				if len(segments) == len(expected) {
					for i, seg := range segments {
						if seg != expected[i] {
							return nil, &kernel.Error{Code: kernel.ENotFound}
						}
					}
					return &sports, nil
				}
				return nil, &kernel.Error{Code: kernel.ENotFound}
			},
		}
		service := category.NewPathService(repo)

		got, err := service.ParseURL("a1/comprehension-ecrite/sports")

		assertNoError(t, err)
		if got.CategoryID != sports.CategoryID {
			t.Errorf("got %v, want %v", got.CategoryID, sports.CategoryID)
		}
	})

	t.Run("handles URL encoding", func(t *testing.T) {
		cat := createTestCategory("cat", "Category", nil)
		repo := &mockRepository{
			findByPathFunc: func(segments []string) (*category.Category, error) {
				if len(segments) == 1 && segments[0] == "café culture" {
					return &cat, nil
				}
				return nil, &kernel.Error{Code: kernel.ENotFound}
			},
		}
		service := category.NewPathService(repo)

		got, err := service.ParseURL("caf%C3%A9%20culture")

		assertNoError(t, err)
		if got.CategoryID != cat.CategoryID {
			t.Errorf("got %v, want %v", got.CategoryID, cat.CategoryID)
		}
	})

	t.Run("trims slashes", func(t *testing.T) {
		cat := createTestCategory("a1", "A1", nil)
		repo := &mockRepository{
			findByPathFunc: func(segments []string) (*category.Category, error) {
				if len(segments) == 1 && segments[0] == "a1" {
					return &cat, nil
				}
				return nil, &kernel.Error{Code: kernel.ENotFound}
			},
		}
		service := category.NewPathService(repo)

		tests := []string{
			"/a1",
			"a1/",
			"/a1/",
			"//a1//",
		}

		for _, url := range tests {
			t.Run(url, func(t *testing.T) {
				got, err := service.ParseURL(url)

				assertNoError(t, err)
				if got.CategoryID != cat.CategoryID {
					t.Errorf("got %v, want %v", got.CategoryID, cat.CategoryID)
				}
			})
		}
	})

	t.Run("returns error for empty path", func(t *testing.T) {
		repo := &mockRepository{}
		service := category.NewPathService(repo)

		tests := []string{"", "/", "//"}

		for _, url := range tests {
			t.Run(fmt.Sprintf("url: %q", url), func(t *testing.T) {
				_, err := service.ParseURL(url)

				assertError(t, err)
			})
		}
	})

	t.Run("returns error for invalid URL encoding", func(t *testing.T) {
		repo := &mockRepository{}
		service := category.NewPathService(repo)

		_, err := service.ParseURL("invalid%encoding")

		assertError(t, err)
	})

	t.Run("returns error for non-existent path", func(t *testing.T) {
		repo := &mockRepository{
			findByPathFunc: func(segments []string) (*category.Category, error) {
				return nil, &kernel.Error{Code: kernel.ENotFound, Message: "category not found"}
			},
		}
		service := category.NewPathService(repo)

		_, err := service.ParseURL("non/existent/path")

		assertError(t, err)
		assertErrorCode(t, err, kernel.ENotFound)
	})
}

func TestPathService_GetBreadcrumbs(t *testing.T) {
	t.Run("returns breadcrumbs for single category", func(t *testing.T) {
		cat := createTestCategory("a1", "A1", nil)
		repo := &mockRepository{
			paths: map[string]category.CategoryPath{
				"a1": {cat},
			},
		}
		service := category.NewPathService(repo)

		catID, _ := kernel.NewID[category.Category]("a1")
		got, err := service.GetBreadcrumbs(catID)

		assertNoError(t, err)
		if len(got) != 1 {
			t.Fatalf("expected 1 breadcrumb, got %d", len(got))
		}

		if got[0].Category.CategoryID != cat.CategoryID {
			t.Errorf("Category ID mismatch")
		}
		if !got[0].IsLast {
			t.Error("expected last breadcrumb")
		}
		if got[0].Level != 0 {
			t.Errorf("Level: got %d, want %d", got[0].Level, 0)
		}
	})

	t.Run("returns breadcrumbs for nested categories", func(t *testing.T) {
		a1ID := "a1"
		readingID := "reading"

		a1 := createTestCategory("a1", "A1", nil)
		reading := createTestCategory("reading", "Compréhension écrite", &a1ID)
		sports := createTestCategory("sports", "Sports", &readingID)

		repo := &mockRepository{
			paths: map[string]category.CategoryPath{
				"sports": {a1, reading, sports},
			},
		}
		service := category.NewPathService(repo)

		catID, _ := kernel.NewID[category.Category]("sports")
		got, err := service.GetBreadcrumbs(catID)

		assertNoError(t, err)
		if len(got) != 3 {
			t.Fatalf("expected 3 breadcrumbs, got %d", len(got))
		}

		// Check first breadcrumb (A1)
		if got[0].Category.CategoryID != a1.CategoryID {
			t.Error("First breadcrumb category mismatch")
		}
		if got[0].IsLast {
			t.Error("First breadcrumb should not be last")
		}
		if got[0].Level != 0 {
			t.Errorf("First breadcrumb level: got %d, want %d", got[0].Level, 0)
		}

		// Check second breadcrumb (Reading)
		if got[1].Category.CategoryID != reading.CategoryID {
			t.Error("Second breadcrumb category mismatch")
		}
		if got[1].IsLast {
			t.Error("Second breadcrumb should not be last")
		}
		if got[1].Level != 1 {
			t.Errorf("Second breadcrumb level: got %d, want %d", got[1].Level, 1)
		}

		// Check third breadcrumb (Sports)
		if got[2].Category.CategoryID != sports.CategoryID {
			t.Error("Third breadcrumb category mismatch")
		}
		if !got[2].IsLast {
			t.Error("Third breadcrumb should be last")
		}
		if got[2].Level != 2 {
			t.Errorf("Third breadcrumb level: got %d, want %d", got[2].Level, 2)
		}
	})

	t.Run("returns error for non-existent category", func(t *testing.T) {
		repo := &mockRepository{
			paths: map[string]category.CategoryPath{},
		}
		service := category.NewPathService(repo)

		catID, _ := kernel.NewID[category.Category]("non-existent")
		_, err := service.GetBreadcrumbs(catID)

		assertError(t, err)
		assertErrorCode(t, err, kernel.ENotFound)
	})
}
