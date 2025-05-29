package post

import "github.com/alnah/fla/internal/domain/kernel"

// SchemaType represents Schema.org markup types for structured data
type SchemaType string

const (
	SchemaTypeArticle            SchemaType = "Article"
	SchemaTypeBlogPosting        SchemaType = "BlogPosting"
	SchemaTypeEducationalContent SchemaType = "EducationalContent"
	SchemaTypeLearningResource   SchemaType = "LearningResource"
	SchemaTypeHowTo              SchemaType = "HowTo"
	SchemaTypeDefault            SchemaType = SchemaTypeEducationalContent
)

func (s SchemaType) String() string { return string(s) }

func (s SchemaType) Validate() error {
	const op = "SchemaType.Validate"

	// Empty is allowed (will use default)
	if s == "" {
		return nil
	}

	switch s {
	case SchemaTypeArticle, SchemaTypeBlogPosting, SchemaTypeEducationalContent,
		SchemaTypeLearningResource, SchemaTypeHowTo:
		return nil
	default:
		return &kernel.Error{
			Code:      kernel.EInvalid,
			Message:   MSchemaTypeInvalid,
			Operation: op,
		}
	}
}

// GetEffectiveType returns the schema type to use, with default fallback
func (s SchemaType) GetEffectiveType() SchemaType {
	if s == "" {
		return SchemaTypeDefault
	}
	return s
}

// IsEducational returns true if this is an educational content type
func (s SchemaType) IsEducational() bool {
	return s == SchemaTypeEducationalContent || s == SchemaTypeLearningResource
}
