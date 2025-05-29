package shared_test

import (
	"strings"
	"testing"

	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/shared"
)

func TestNewDescription(t *testing.T) {
	t.Run("creates description with valid input", func(t *testing.T) {
		descriptions := []string{
			"", // empty is valid (optional field)
			"A short description",
			"This is a comprehensive guide to learning French grammar, including verb conjugations, noun genders, and sentence structure.",
			strings.Repeat("a", shared.MaxDescriptionLength),
		}

		for _, desc := range descriptions {
			t.Run(desc, func(t *testing.T) {
				got, err := shared.NewDescription(desc)

				assertNoError(t, err)
				if got.String() != desc {
					t.Errorf("got %q, want %q", got, desc)
				}
			})
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		input := "  A description  "
		want := "A description"

		got, err := shared.NewDescription(input)

		assertNoError(t, err)
		if got.String() != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("allows empty description", func(t *testing.T) {
		got, err := shared.NewDescription("")

		assertNoError(t, err)
		if got.String() != "" {
			t.Errorf("got %q, want empty", got)
		}
	})

	t.Run("rejects description exceeding max length", func(t *testing.T) {
		longDesc := strings.Repeat("a", shared.MaxDescriptionLength+1)

		_, err := shared.NewDescription(longDesc)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})
}

func TestDescription_Validate(t *testing.T) {
	t.Run("valid description passes", func(t *testing.T) {
		desc := shared.Description("A valid description")

		err := desc.Validate()

		assertNoError(t, err)
	})

	t.Run("empty description passes", func(t *testing.T) {
		desc := shared.Description("")

		err := desc.Validate()

		assertNoError(t, err)
	})

	t.Run("too long description fails", func(t *testing.T) {
		desc := shared.Description(strings.Repeat("a", shared.MaxDescriptionLength+1))

		err := desc.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})
}
