package shared_test

import (
	"fmt"
	"testing"

	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/shared"
)

func TestNewPagination(t *testing.T) {
	t.Run("creates pagination with valid input", func(t *testing.T) {
		page := 2
		limit := 20
		totalItems := 100

		got, err := shared.NewPagination(page, limit, totalItems)

		assertNoError(t, err)
		if got.Page != page {
			t.Errorf("page: got %d, want %d", got.Page, page)
		}
		if got.Limit != limit {
			t.Errorf("limit: got %d, want %d", got.Limit, limit)
		}
		if got.TotalItems != totalItems {
			t.Errorf("totalItems: got %d, want %d", got.TotalItems, totalItems)
		}
		if got.TotalPages != 5 {
			t.Errorf("totalPages: got %d, want %d", got.TotalPages, 5)
		}
	})

	t.Run("uses default limit when zero", func(t *testing.T) {
		got, err := shared.NewPagination(1, 0, 100)

		assertNoError(t, err)
		if got.Limit != shared.DefaultPageLimit {
			t.Errorf("limit: got %d, want %d", got.Limit, shared.DefaultPageLimit)
		}
	})

	t.Run("uses default limit when negative", func(t *testing.T) {
		got, err := shared.NewPagination(1, -5, 100)

		assertNoError(t, err)
		if got.Limit != shared.DefaultPageLimit {
			t.Errorf("limit: got %d, want %d", got.Limit, shared.DefaultPageLimit)
		}
	})

	t.Run("uses first page when zero", func(t *testing.T) {
		got, err := shared.NewPagination(0, 10, 100)

		assertNoError(t, err)
		if got.Page != 1 {
			t.Errorf("page: got %d, want %d", got.Page, 1)
		}
	})

	t.Run("uses first page when negative", func(t *testing.T) {
		got, err := shared.NewPagination(-5, 10, 100)

		assertNoError(t, err)
		if got.Page != 1 {
			t.Errorf("page: got %d, want %d", got.Page, 1)
		}
	})

	t.Run("calculates total pages correctly", func(t *testing.T) {
		tests := []struct {
			totalItems int
			limit      int
			wantPages  int
		}{
			{0, 10, 0},
			{1, 10, 1},
			{10, 10, 1},
			{11, 10, 2},
			{99, 10, 10},
			{100, 10, 10},
			{101, 10, 11},
			{25, 5, 5},
			{26, 5, 6},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("%d items, %d per page", tt.totalItems, tt.limit), func(t *testing.T) {
				got, err := shared.NewPagination(1, tt.limit, tt.totalItems)

				assertNoError(t, err)
				if got.TotalPages != tt.wantPages {
					t.Errorf("totalPages: got %d, want %d", got.TotalPages, tt.wantPages)
				}
			})
		}
	})

	t.Run("rejects limit above maximum", func(t *testing.T) {
		_, err := shared.NewPagination(1, shared.MaxPageLimit+1, 100)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("rejects negative total items", func(t *testing.T) {
		_, err := shared.NewPagination(1, 10, -1)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("accepts minimum and maximum limits", func(t *testing.T) {
		t.Run("minimum limit", func(t *testing.T) {
			got, err := shared.NewPagination(1, shared.MinPageLimit, 100)

			assertNoError(t, err)
			if got.Limit != shared.MinPageLimit {
				t.Errorf("limit: got %d, want %d", got.Limit, shared.MinPageLimit)
			}
		})

		t.Run("maximum limit", func(t *testing.T) {
			got, err := shared.NewPagination(1, shared.MaxPageLimit, 100)

			assertNoError(t, err)
			if got.Limit != shared.MaxPageLimit {
				t.Errorf("limit: got %d, want %d", got.Limit, shared.MaxPageLimit)
			}
		})
	})
}

func TestPagination_Validate(t *testing.T) {
	t.Run("valid pagination passes", func(t *testing.T) {
		p, _ := shared.NewPagination(1, 10, 100)

		err := p.Validate()

		assertNoError(t, err)
	})

	t.Run("page less than 1 fails", func(t *testing.T) {
		// Create invalid pagination manually
		p := shared.Pagination{Page: 0, Limit: 10, TotalItems: 100}

		err := p.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("limit out of range fails", func(t *testing.T) {
		tests := []struct {
			name  string
			limit int
		}{
			{"below minimum", 0},
			{"above maximum", shared.MaxPageLimit + 1},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				p := shared.Pagination{Page: 1, Limit: tt.limit, TotalItems: 100}

				err := p.Validate()

				assertError(t, err)
				assertErrorCode(t, err, kernel.EInvalid)
			})
		}
	})

	t.Run("negative total items fails", func(t *testing.T) {
		p := shared.Pagination{Page: 1, Limit: 10, TotalItems: -1}

		err := p.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})
}

func TestPagination_String(t *testing.T) {
	p, _ := shared.NewPagination(2, 10, 100)

	got := p.String()
	want := "Pagination{Page: 2, Limit: 10, TotalItems: 100, TotalPages: 10}"

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPagination_Offset(t *testing.T) {
	tests := []struct {
		page   int
		limit  int
		offset int
	}{
		{1, 10, 0},
		{2, 10, 10},
		{3, 10, 20},
		{5, 20, 80},
		{1, 100, 0},
		{10, 5, 45},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("page %d, limit %d", tt.page, tt.limit), func(t *testing.T) {
			p, _ := shared.NewPagination(tt.page, tt.limit, 1000)

			got := p.Offset()

			if got != tt.offset {
				t.Errorf("got %d, want %d", got, tt.offset)
			}
		})
	}
}

func TestPagination_Navigation(t *testing.T) {
	t.Run("HasNextPage", func(t *testing.T) {
		tests := []struct {
			page       int
			totalPages int
			hasNext    bool
		}{
			{1, 5, true},
			{3, 5, true},
			{5, 5, false},
			{1, 1, false},
			{1, 0, false},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("page %d of %d", tt.page, tt.totalPages), func(t *testing.T) {
				totalItems := tt.totalPages * 10
				if tt.totalPages == 0 {
					totalItems = 0
				}
				p, _ := shared.NewPagination(tt.page, 10, totalItems)

				got := p.HasNextPage()

				if got != tt.hasNext {
					t.Errorf("got %v, want %v", got, tt.hasNext)
				}
			})
		}
	})

	t.Run("HasPreviousPage", func(t *testing.T) {
		tests := []struct {
			page    int
			hasPrev bool
		}{
			{1, false},
			{2, true},
			{5, true},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("page %d", tt.page), func(t *testing.T) {
				p, _ := shared.NewPagination(tt.page, 10, 100)

				got := p.HasPreviousPage()

				if got != tt.hasPrev {
					t.Errorf("got %v, want %v", got, tt.hasPrev)
				}
			})
		}
	})

	t.Run("NextPage", func(t *testing.T) {
		tests := []struct {
			page       int
			totalPages int
			wantNext   int
		}{
			{1, 5, 2},
			{3, 5, 4},
			{5, 5, 5}, // stays on last page
			{1, 1, 1}, // stays on only page
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("page %d of %d", tt.page, tt.totalPages), func(t *testing.T) {
				p, _ := shared.NewPagination(tt.page, 10, tt.totalPages*10)

				got := p.NextPage()

				if got != tt.wantNext {
					t.Errorf("got %d, want %d", got, tt.wantNext)
				}
			})
		}
	})

	t.Run("PreviousPage", func(t *testing.T) {
		tests := []struct {
			page     int
			wantPrev int
		}{
			{1, 1}, // stays on first page
			{2, 1},
			{5, 4},
			{10, 9},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("page %d", tt.page), func(t *testing.T) {
				p, _ := shared.NewPagination(tt.page, 10, 100)

				got := p.PreviousPage()

				if got != tt.wantPrev {
					t.Errorf("got %d, want %d", got, tt.wantPrev)
				}
			})
		}
	})
}

func TestPagination_Status(t *testing.T) {
	t.Run("IsEmpty", func(t *testing.T) {
		tests := []struct {
			totalItems int
			isEmpty    bool
		}{
			{0, true},
			{1, false},
			{100, false},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("%d items", tt.totalItems), func(t *testing.T) {
				p, _ := shared.NewPagination(1, 10, tt.totalItems)

				got := p.IsEmpty()

				if got != tt.isEmpty {
					t.Errorf("got %v, want %v", got, tt.isEmpty)
				}
			})
		}
	})

	t.Run("IsFirstPage", func(t *testing.T) {
		tests := []struct {
			page    int
			isFirst bool
		}{
			{1, true},
			{2, false},
			{10, false},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("page %d", tt.page), func(t *testing.T) {
				p, _ := shared.NewPagination(tt.page, 10, 100)

				got := p.IsFirstPage()

				if got != tt.isFirst {
					t.Errorf("got %v, want %v", got, tt.isFirst)
				}
			})
		}
	})

	t.Run("IsLastPage", func(t *testing.T) {
		tests := []struct {
			page       int
			totalPages int
			isLast     bool
		}{
			{1, 1, true},
			{1, 5, false},
			{5, 5, true},
			{10, 10, true},
			{9, 10, false},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("page %d of %d", tt.page, tt.totalPages), func(t *testing.T) {
				p, _ := shared.NewPagination(tt.page, 10, tt.totalPages*10)

				got := p.IsLastPage()

				if got != tt.isLast {
					t.Errorf("got %v, want %v", got, tt.isLast)
				}
			})
		}
	})
}

func TestPagination_Items(t *testing.T) {
	t.Run("ItemsOnCurrentPage", func(t *testing.T) {
		tests := []struct {
			page       int
			limit      int
			totalItems int
			wantItems  int
		}{
			{1, 10, 100, 10},  // full page
			{10, 10, 100, 10}, // last full page
			{11, 10, 105, 5},  // partial last page
			{1, 10, 5, 5},     // partial only page
			{1, 10, 0, 0},     // empty
			{3, 20, 55, 15},   // partial last page
		}

		for _, tt := range tests {
			name := fmt.Sprintf("page %d, limit %d, total %d", tt.page, tt.limit, tt.totalItems)
			t.Run(name, func(t *testing.T) {
				p, _ := shared.NewPagination(tt.page, tt.limit, tt.totalItems)

				got := p.ItemsOnCurrentPage()

				if got != tt.wantItems {
					t.Errorf("got %d, want %d", got, tt.wantItems)
				}
			})
		}
	})

	t.Run("StartItem", func(t *testing.T) {
		tests := []struct {
			page       int
			limit      int
			totalItems int
			wantStart  int
		}{
			{1, 10, 100, 1},
			{2, 10, 100, 11},
			{3, 10, 100, 21},
			{5, 20, 100, 81},
			{1, 10, 0, 0}, // empty
		}

		for _, tt := range tests {
			name := fmt.Sprintf("page %d, limit %d", tt.page, tt.limit)
			t.Run(name, func(t *testing.T) {
				p, _ := shared.NewPagination(tt.page, tt.limit, tt.totalItems)

				got := p.StartItem()

				if got != tt.wantStart {
					t.Errorf("got %d, want %d", got, tt.wantStart)
				}
			})
		}
	})

	t.Run("EndItem", func(t *testing.T) {
		tests := []struct {
			page       int
			limit      int
			totalItems int
			wantEnd    int
		}{
			{1, 10, 100, 10},
			{2, 10, 100, 20},
			{10, 10, 100, 100},
			{11, 10, 105, 105}, // partial last page
			{1, 10, 5, 5},      // partial only page
			{1, 10, 0, 0},      // empty
		}

		for _, tt := range tests {
			name := fmt.Sprintf("page %d, limit %d, total %d", tt.page, tt.limit, tt.totalItems)
			t.Run(name, func(t *testing.T) {
				p, _ := shared.NewPagination(tt.page, tt.limit, tt.totalItems)

				got := p.EndItem()

				if got != tt.wantEnd {
					t.Errorf("got %d, want %d", got, tt.wantEnd)
				}
			})
		}
	})
}
