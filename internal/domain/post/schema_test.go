package post_test

import (
	"testing"

	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/post"
)

func TestSchemaType_String(t *testing.T) {
	tests := []struct {
		schema post.SchemaType
		want   string
	}{
		{post.SchemaTypeArticle, "Article"},
		{post.SchemaTypeBlogPosting, "BlogPosting"},
		{post.SchemaTypeEducationalContent, "EducationalContent"},
		{post.SchemaTypeLearningResource, "LearningResource"},
		{post.SchemaTypeHowTo, "HowTo"},
		{post.SchemaTypeDefault, "EducationalContent"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.schema.String()

			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSchemaType_Validate(t *testing.T) {
	t.Run("valid schema types pass", func(t *testing.T) {
		validSchemas := []post.SchemaType{
			post.SchemaTypeArticle,
			post.SchemaTypeBlogPosting,
			post.SchemaTypeEducationalContent,
			post.SchemaTypeLearningResource,
			post.SchemaTypeHowTo,
			"", // empty is allowed (will use default)
		}

		for _, schema := range validSchemas {
			t.Run(string(schema), func(t *testing.T) {
				err := schema.Validate()

				assertNoError(t, err)
			})
		}
	})

	t.Run("invalid schema type fails", func(t *testing.T) {
		invalidSchemas := []post.SchemaType{
			"InvalidType",
			"article", // case sensitive
			"Blog",
			"Educational",
			"HowToGuide",
			"Tutorial",
		}

		for _, schema := range invalidSchemas {
			t.Run(string(schema), func(t *testing.T) {
				err := schema.Validate()

				assertError(t, err)
				assertErrorCode(t, err, kernel.EInvalid)
			})
		}
	})
}

func TestSchemaType_GetEffectiveType(t *testing.T) {
	t.Run("returns self when set", func(t *testing.T) {
		schemas := []post.SchemaType{
			post.SchemaTypeArticle,
			post.SchemaTypeBlogPosting,
			post.SchemaTypeEducationalContent,
			post.SchemaTypeLearningResource,
			post.SchemaTypeHowTo,
		}

		for _, schema := range schemas {
			t.Run(string(schema), func(t *testing.T) {
				got := schema.GetEffectiveType()

				if got != schema {
					t.Errorf("got %v, want %v", got, schema)
				}
			})
		}
	})

	t.Run("returns default when empty", func(t *testing.T) {
		var emptySchema post.SchemaType

		got := emptySchema.GetEffectiveType()
		want := post.SchemaTypeDefault

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("default is EducationalContent", func(t *testing.T) {
		if post.SchemaTypeDefault != post.SchemaTypeEducationalContent {
			t.Errorf("SchemaTypeDefault: got %v, want %v",
				post.SchemaTypeDefault, post.SchemaTypeEducationalContent)
		}
	})
}

func TestSchemaType_IsEducational(t *testing.T) {
	tests := []struct {
		schema        post.SchemaType
		isEducational bool
	}{
		{post.SchemaTypeArticle, false},
		{post.SchemaTypeBlogPosting, false},
		{post.SchemaTypeEducationalContent, true},
		{post.SchemaTypeLearningResource, true},
		{post.SchemaTypeHowTo, false},
		{"", false}, // empty
	}

	for _, tt := range tests {
		t.Run(string(tt.schema), func(t *testing.T) {
			got := tt.schema.IsEducational()

			if got != tt.isEducational {
				t.Errorf("got %v, want %v", got, tt.isEducational)
			}
		})
	}
}

func TestSchemaTypeConstants(t *testing.T) {
	// Ensure constants have expected values
	tests := []struct {
		name   string
		schema post.SchemaType
		want   string
	}{
		{"SchemaTypeArticle", post.SchemaTypeArticle, "Article"},
		{"SchemaTypeBlogPosting", post.SchemaTypeBlogPosting, "BlogPosting"},
		{"SchemaTypeEducationalContent", post.SchemaTypeEducationalContent, "EducationalContent"},
		{"SchemaTypeLearningResource", post.SchemaTypeLearningResource, "LearningResource"},
		{"SchemaTypeHowTo", post.SchemaTypeHowTo, "HowTo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.schema) != tt.want {
				t.Errorf("got %q, want %q", tt.schema, tt.want)
			}
		})
	}
}
