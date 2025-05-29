package post

import (
	"strings"

	"github.com/alnah/fla/internal/domain/kernel"
)

const (
	MinPostContentLength int = 300
	MaxPostContentLength int = 10000
)

// PostContent represents the main body text of educational blog posts.
// Enforces minimum length for substantial content and maximum for readability.
type PostContent string

// NewPostContent creates validated post content with educational length requirements.
// Ensures posts provide sufficient learning value while remaining digestible.
func NewPostContent(content string) (PostContent, error) {
	const op = "NewPostContent"

	t := PostContent(strings.TrimSpace(content))
	if err := t.Validate(); err != nil {
		return "", &kernel.Error{Operation: op, Cause: err}
	}

	return t, nil
}

func (p PostContent) String() string {
	return string(p)
}

// Validate enforces content length standards for educational effectiveness.
// Balances comprehensive learning material with reader attention spans.
func (p PostContent) Validate() error {
	const op = "PostContent.Validate"

	if err := kernel.ValidatePresence("post content", p.String(), op); err != nil {
		return err
	}

	if err := kernel.ValidateMinLength("post content", p.String(), MinPostContentLength, op); err != nil {
		return err
	}

	if err := kernel.ValidateMaxLength("post content", p.String(), MaxPostContentLength, op); err != nil {
		return err
	}

	return nil
}
