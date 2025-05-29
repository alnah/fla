package shared_test

import (
	"strings"
	"testing"

	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/shared"
)

func TestNewTitle(t *testing.T) {
	t.Run("creates title with valid input", func(t *testing.T) {
		titles := []string{
			"Learn French in 10 Days",
			"A Complete Guide to French Grammar",
			"How to Master French Pronunciation: Tips & Tricks",
			strings.Repeat("a", shared.MinTitleLength),
			strings.Repeat("a", shared.MaxTitleLength),
		}

		for _, title := range titles {
			t.Run(title, func(t *testing.T) {
				got, err := shared.NewTitle(title)

				assertNoError(t, err)
				if got.String() != title {
					t.Errorf("got %q, want %q", got, title)
				}
			})
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		input := "  Learn French  "
		want := "Learn French"

		got, err := shared.NewTitle(input)

		assertNoError(t, err)
		if got.String() != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("rejects empty title", func(t *testing.T) {
		_, err := shared.NewTitle("")

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("rejects title below min length", func(t *testing.T) {
		shortTitle := strings.Repeat("a", shared.MinTitleLength-1)

		_, err := shared.NewTitle(shortTitle)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("rejects title exceeding max length", func(t *testing.T) {
		longTitle := strings.Repeat("a", shared.MaxTitleLength+1)

		_, err := shared.NewTitle(longTitle)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("accepts unicode characters", func(t *testing.T) {
		title := "Apprendre le français 🇫🇷"

		got, err := shared.NewTitle(title)

		assertNoError(t, err)
		if got.String() != title {
			t.Errorf("got %q, want %q", got, title)
		}
	})
}

func TestTitle_Validate(t *testing.T) {
	t.Run("valid title passes", func(t *testing.T) {
		title := shared.Title("Learn French Today")

		err := title.Validate()

		assertNoError(t, err)
	})

	t.Run("empty title fails", func(t *testing.T) {
		title := shared.Title("")

		err := title.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("too short title fails", func(t *testing.T) {
		title := shared.Title("Short")

		err := title.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("too long title fails", func(t *testing.T) {
		title := shared.Title(strings.Repeat("a", shared.MaxTitleLength+1))

		err := title.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})
}
