package tag_test

import (
	"strings"
	"testing"
	"time"

	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/tag"
	"github.com/alnah/fla/internal/domain/user"
)

func TestNewTagName(t *testing.T) {
	t.Run("creates tag name with valid input", func(t *testing.T) {
		names := []string{
			"grammar",
			"french-verbs",
			"A1 Level",
			"日本語",
			"a", // minimum length
			strings.Repeat("a", tag.MaxTagNameLength),
		}

		for _, name := range names {
			t.Run(name, func(t *testing.T) {
				got, err := tag.NewTagName(name)

				assertNoError(t, err)
				if got.String() != name {
					t.Errorf("got %q, want %q", got, name)
				}
			})
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		input := "  grammar  "
		want := "grammar"

		got, err := tag.NewTagName(input)

		assertNoError(t, err)
		if got.String() != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("rejects empty name", func(t *testing.T) {
		_, err := tag.NewTagName("")

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("rejects whitespace only", func(t *testing.T) {
		_, err := tag.NewTagName("   ")

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("rejects name exceeding max length", func(t *testing.T) {
		longName := strings.Repeat("a", tag.MaxTagNameLength+1)

		_, err := tag.NewTagName(longName)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})
}

func TestNewTag(t *testing.T) {
	validTagID, _ := kernel.NewID[tag.Tag]("tag-123")
	validUserID, _ := kernel.NewID[user.User]("user-123")
	validName, _ := tag.NewTagName("grammar")
	validTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	t.Run("creates tag with valid input", func(t *testing.T) {
		tagEntity := tag.Tag{
			TagID:     validTagID,
			Name:      validName,
			CreatedBy: validUserID,
			CreatedAt: validTime,
		}

		got, err := tag.NewTag(tagEntity)

		assertNoError(t, err)

		if got.TagID != validTagID {
			t.Errorf("TagID: got %v, want %v", got.TagID, validTagID)
		}
		if got.Name != validName {
			t.Errorf("Name: got %v, want %v", got.Name, validName)
		}
		if got.CreatedBy != validUserID {
			t.Errorf("CreatedBy: got %v, want %v", got.CreatedBy, validUserID)
		}
		if !got.CreatedAt.Equal(validTime) {
			t.Errorf("CreatedAt: got %v, want %v", got.CreatedAt, validTime)
		}
	})

	t.Run("rejects invalid tag", func(t *testing.T) {
		tests := []struct {
			name      string
			tagEntity tag.Tag
		}{
			{
				name: "empty tag ID",
				tagEntity: tag.Tag{
					TagID:     kernel.ID[tag.Tag](""),
					Name:      validName,
					CreatedBy: validUserID,
					CreatedAt: validTime,
				},
			},
			{
				name: "empty name",
				tagEntity: tag.Tag{
					TagID:     validTagID,
					Name:      tag.TagName(""),
					CreatedBy: validUserID,
					CreatedAt: validTime,
				},
			},
			{
				name: "empty created by",
				tagEntity: tag.Tag{
					TagID:     validTagID,
					Name:      validName,
					CreatedBy: kernel.ID[user.User](""),
					CreatedAt: validTime,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := tag.NewTag(tt.tagEntity)

				assertError(t, err)
				assertErrorCode(t, err, kernel.EInvalid)
			})
		}
	})
}

func TestTag_Validate(t *testing.T) {
	validTagID, _ := kernel.NewID[tag.Tag]("tag-123")
	validUserID, _ := kernel.NewID[user.User]("user-123")
	validName, _ := tag.NewTagName("grammar")
	validTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	t.Run("valid tag passes", func(t *testing.T) {
		tagEntity := tag.Tag{
			TagID:     validTagID,
			Name:      validName,
			CreatedBy: validUserID,
			CreatedAt: validTime,
		}

		err := tagEntity.Validate()

		assertNoError(t, err)
	})

	t.Run("invalid fields fail", func(t *testing.T) {
		tests := []struct {
			name     string
			modifier func(*tag.Tag)
		}{
			{
				name: "empty ID",
				modifier: func(t *tag.Tag) {
					t.TagID = kernel.ID[tag.Tag]("")
				},
			},
			{
				name: "empty name",
				modifier: func(t *tag.Tag) {
					t.Name = tag.TagName("")
				},
			},
			{
				name: "name too long",
				modifier: func(t *tag.Tag) {
					t.Name = tag.TagName(strings.Repeat("a", tag.MaxTagNameLength+1))
				},
			},
			{
				name: "empty created by",
				modifier: func(t *tag.Tag) {
					t.CreatedBy = kernel.ID[user.User]("")
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Create valid tag
				tagEntity := tag.Tag{
					TagID:     validTagID,
					Name:      validName,
					CreatedBy: validUserID,
					CreatedAt: validTime,
				}

				// Apply modifier to make it invalid
				tt.modifier(&tagEntity)

				err := tagEntity.Validate()

				assertError(t, err)
				assertErrorCode(t, err, kernel.EInvalid)
			})
		}
	})
}

func TestTagName_String(t *testing.T) {
	want := "grammar"
	name := tag.TagName(want)

	got := name.String()

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTagName_Validate(t *testing.T) {
	t.Run("valid name passes", func(t *testing.T) {
		name := tag.TagName("grammar")

		err := name.Validate()

		assertNoError(t, err)
	})

	t.Run("empty name fails", func(t *testing.T) {
		name := tag.TagName("")

		err := name.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("too long name fails", func(t *testing.T) {
		name := tag.TagName(strings.Repeat("a", tag.MaxTagNameLength+1))

		err := name.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})
}

// Test helpers
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func assertErrorCode(t *testing.T, err error, want string) {
	t.Helper()
	got := kernel.ErrorCode(err)
	if got != want {
		t.Errorf("error code: got %q, want %q", got, want)
	}
}
