package category_test

import (
	"testing"
	"time"

	"github.com/alnah/fla/internal/domain/category"
	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/user"
)

func createTestCategory(id, name string, parentID *string) category.Category {
	clock := &stubClock{t: time.Now()}

	categoryID, _ := kernel.NewID[category.Category](id)
	userID, _ := kernel.NewID[user.User]("user-123")
	categoryName, _ := category.NewCategoryName(name)

	params := category.NewCategoryParams{
		CategoryID: categoryID,
		Name:       categoryName,
		CreatedBy:  userID,
		Clock:      clock,
	}

	if parentID != nil {
		pid, _ := kernel.NewID[category.Category](*parentID)
		params.ParentID = &pid
	}

	cat, _ := category.NewCategory(params)
	return cat
}

func TestCategoryPath_String(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		path := category.CategoryPath{}

		got := path.String()
		want := ""

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("single category", func(t *testing.T) {
		cat := createTestCategory("a1", "A1", nil)
		path := category.CategoryPath{cat}

		got := path.String()
		want := "a1"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("multiple categories", func(t *testing.T) {
		a1ID := "a1"
		readingID := "reading"

		a1 := createTestCategory("a1", "A1", nil)
		reading := createTestCategory("reading", "Compréhension écrite", &a1ID)
		sports := createTestCategory("sports", "Sports", &readingID)

		path := category.CategoryPath{a1, reading, sports}

		got := path.String()
		want := "a1/comprehension-ecrite/sports"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("handles special characters in slugs", func(t *testing.T) {
		cat1 := createTestCategory("cat1", "Café & Culture", nil)
		cat2ID := "cat1"
		cat2 := createTestCategory("cat2", "Français Avancé", &cat2ID)

		path := category.CategoryPath{cat1, cat2}

		got := path.String()
		want := "cafe-culture/francais-avance"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestCategoryPath_Depth(t *testing.T) {
	tests := []struct {
		name      string
		pathLen   int
		wantDepth int
	}{
		{"empty path", 0, -1},
		{"single category", 1, 0},
		{"two categories", 2, 1},
		{"three categories", 3, 2},
		{"four categories", 4, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := make(category.CategoryPath, tt.pathLen)
			for i := range tt.pathLen {
				path[i] = createTestCategory("cat", "Category", nil)
			}

			got := path.Depth()

			if got != tt.wantDepth {
				t.Errorf("got %d, want %d", got, tt.wantDepth)
			}
		})
	}
}

func TestCategoryPath_IsValidDepth(t *testing.T) {
	tests := []struct {
		name    string
		pathLen int
		isValid bool
	}{
		{"empty path", 0, true},
		{"depth 0 (1 category)", 1, true},
		{"depth 1 (2 categories)", 2, true},
		{"depth 2 (3 categories)", 3, true},
		{"depth 3 (4 categories)", 4, false},
		{"depth 4 (5 categories)", 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := make(category.CategoryPath, tt.pathLen)
			for i := range tt.pathLen {
				path[i] = createTestCategory("cat", "Category", nil)
			}

			got := path.IsValidDepth()

			if got != tt.isValid {
				t.Errorf("got %v, want %v", got, tt.isValid)
			}
		})
	}

	t.Run("max depth constant", func(t *testing.T) {
		if category.MaxCategoryDepth != 3 {
			t.Errorf("MaxCategoryDepth: got %d, want %d", category.MaxCategoryDepth, 3)
		}
	})
}

func TestCategoryPath_Leaf(t *testing.T) {
	t.Run("empty path returns nil", func(t *testing.T) {
		path := category.CategoryPath{}

		got := path.Leaf()

		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("single category returns itself", func(t *testing.T) {
		cat := createTestCategory("a1", "A1", nil)
		path := category.CategoryPath{cat}

		got := path.Leaf()

		if got == nil {
			t.Fatal("expected non-nil leaf")
		}
		if got.CategoryID != cat.CategoryID {
			t.Errorf("got %v, want %v", got.CategoryID, cat.CategoryID)
		}
	})

	t.Run("multiple categories returns last", func(t *testing.T) {
		a1ID := "a1"
		readingID := "reading"

		a1 := createTestCategory("a1", "A1", nil)
		reading := createTestCategory("reading", "Compréhension écrite", &a1ID)
		sports := createTestCategory("sports", "Sports", &readingID)

		path := category.CategoryPath{a1, reading, sports}

		got := path.Leaf()

		if got == nil {
			t.Fatal("expected non-nil leaf")
		}
		if got.CategoryID != sports.CategoryID {
			t.Errorf("got %v, want %v", got.CategoryID, sports.CategoryID)
		}
		if got.Name != sports.Name {
			t.Errorf("got %v, want %v", got.Name, sports.Name)
		}
	})
}

func TestCategoryBreadcrumb(t *testing.T) {
	t.Run("breadcrumb properties", func(t *testing.T) {
		cat := createTestCategory("a1", "A1", nil)

		breadcrumb := category.CategoryBreadcrumb{
			Category: cat,
			IsLast:   false,
			Level:    0,
		}

		if breadcrumb.Category.CategoryID != cat.CategoryID {
			t.Errorf("Category ID mismatch")
		}
		if breadcrumb.IsLast != false {
			t.Errorf("IsLast: got %v, want %v", breadcrumb.IsLast, false)
		}
		if breadcrumb.Level != 0 {
			t.Errorf("Level: got %d, want %d", breadcrumb.Level, 0)
		}
	})

	t.Run("last breadcrumb", func(t *testing.T) {
		cat := createTestCategory("sports", "Sports", nil)

		breadcrumb := category.CategoryBreadcrumb{
			Category: cat,
			IsLast:   true,
			Level:    2,
		}

		if !breadcrumb.IsLast {
			t.Error("expected IsLast to be true")
		}
		if breadcrumb.Level != 2 {
			t.Errorf("Level: got %d, want %d", breadcrumb.Level, 2)
		}
		if breadcrumb.Category.Slug.String() != "sports" {
			t.Errorf("expected category slug to be 'sports', got %q", breadcrumb.Category.Slug.String())
		}
	})
}
