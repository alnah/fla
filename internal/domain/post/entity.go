package post

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/alnah/fla/internal/domain/category"
	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/shared"
	"github.com/alnah/fla/internal/domain/user"
)

const (
	MPostInvalid                 string = "Invalid post."
	MPostInvalidStatusTransition string = "Invalid status transition from %s to %s."
	MPostCannotPublish           string = "User cannot publish this post."
	MPostCannotApprove           string = "User cannot approve this post."
	MPostCannotSchedule          string = "User cannot schedule this post."
	MPostScheduledDateRequired   string = "Scheduled date is required for scheduled posts."
	MPostScheduledDatePast       string = "Scheduled date must be in the future."
	AverageWordsPerMinute               = 200 // Average reading speed for adults
)

// Type markers for URL generics
type (
	FeaturedImage  struct{}
	OpenGraphImage struct{}
	Canonical      struct{}
)

// Post represents a complete learning article with content, metadata, and publishing workflow.
// Designed for educational blogs with SEO optimization and approval processes.
type Post struct {
	// Identity
	PostID kernel.ID[Post]
	Owner  kernel.ID[user.User]

	// Data
	Title         shared.Title
	Content       PostContent
	FeaturedImage kernel.URL[FeaturedImage] // Optional: featured image for the post
	Status        Status
	Slug          shared.Slug

	// SEO & Social Media
	SEOTitle             shared.Title               // Optional: SEO-optimized title (defaults to Title if empty)
	SEODescription       shared.Description         // Optional: Meta description for search results
	OpenGraphTitle       shared.Title               // Optional: Social media title (defaults to SEOTitle if empty)
	OpenGraphDescription shared.Description         // Optional: Social media description (defaults to SEODescription if empty)
	OpenGraphImage       kernel.URL[OpenGraphImage] // Optional: Social media image (defaults to FeaturedImage if empty)

	// Advanced SEO
	CanonicalURL kernel.URL[Canonical] // Optional: Canonical URL for duplicate content prevention
	SchemaType   SchemaType            // Schema.org markup type for structured data

	// Publishing workflow
	PublishedAt *time.Time            // When post was/will be published (nil = not published)
	ApprovedBy  *kernel.ID[user.User] // Who approved the post for publishing (nil = not approved)
	ApprovedAt  *time.Time            // When post was approved (nil = not approved)

	// Meta
	CreatedAt time.Time
	UpdatedAt time.Time
	Category  category.Category // Post must have one Category

	// DI
	Clock kernel.Clock
}

// NewPostParams holds the parameters needed to create a new post.
type NewPostParams struct {
	// Required
	PostID        kernel.ID[Post]
	Owner         kernel.ID[user.User]
	Title         shared.Title
	Content       PostContent
	FeaturedImage kernel.URL[FeaturedImage]
	Status        Status
	Category      category.Category

	// Optional
	PublishedAt *time.Time

	// Optional SEO & Social Media (all optional)
	SEOTitle       shared.Title
	SEODescription shared.Description

	// Optional Social Media
	OpenGraphTitle       shared.Title
	OpenGraphDescription shared.Description
	OpenGraphImage       kernel.URL[OpenGraphImage]

	// Optional advanced SEO
	CanonicalURL kernel.URL[Canonical] // Canonical URL for duplicate content
	SchemaType   SchemaType            // Schema.org markup type

	// DI
	Clock kernel.Clock
}

// NewPost creates a validated post with automatic slug generation and workflow initialization.
// Ensures all required content and metadata are properly structured for publication.
func NewPost(p NewPostParams) (Post, error) {
	const op = "NewPost"

	now := p.Clock.Now()

	slug, err := shared.NewSlug(p.Title.String())
	if err != nil {
		return Post{}, &kernel.Error{Operation: op, Cause: err}
	}

	post := Post{
		PostID:               p.PostID,
		Owner:                p.Owner,
		Title:                p.Title,
		Content:              p.Content,
		FeaturedImage:        p.FeaturedImage,
		Status:               p.Status,
		Slug:                 slug,
		SEOTitle:             p.SEOTitle,
		SEODescription:       p.SEODescription,
		OpenGraphTitle:       p.OpenGraphTitle,
		OpenGraphDescription: p.OpenGraphDescription,
		OpenGraphImage:       p.OpenGraphImage,
		CanonicalURL:         p.CanonicalURL,
		SchemaType:           p.SchemaType,
		PublishedAt:          p.PublishedAt,
		ApprovedBy:           nil, // New posts are not approved
		ApprovedAt:           nil,
		CreatedAt:            now,
		UpdatedAt:            now,
		Category:             p.Category,
		Clock:                p.Clock,
	}

	if err := post.Validate(); err != nil {
		return Post{}, &kernel.Error{Operation: op, Cause: err}
	}

	return post, nil
}

// String returns a string representation of the post.
func (p Post) String() string {
	const maxContentLength = 100

	content := p.Content.String()
	if len(content) > maxContentLength {
		content = content[:maxContentLength] + "..."
	}

	return fmt.Sprintf("Post{"+
		"ID: %q, "+
		"Title: %q, "+
		"Status: %q, "+
		"Slug: %q, "+
		"Owner: %q, "+
		"Category: %q, "+
		"Content: %q, "+
		"WordCount: %d, "+
		"HasFeaturedImage: %t"+
		"}",
		p.PostID,
		p.Title,
		p.Status,
		p.Slug,
		p.Owner,
		p.Category.Name,
		content,
		p.WordCount(),
		p.HasFeaturedImage(),
	)
}

// Validate performs validation on the post.
func (p Post) Validate() error {
	const op = "Post.Validate"

	if err := p.PostID.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := p.Owner.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := p.Title.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := p.Content.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := p.FeaturedImage.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	// Fix: Only validate non-empty SEO titles
	if p.SEOTitle.String() != "" {
		if err := p.SEOTitle.Validate(); err != nil {
			return &kernel.Error{Operation: op, Cause: err}
		}
	}

	if err := p.SEODescription.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	// Fix: Only validate non-empty OpenGraph titles
	if p.OpenGraphTitle.String() != "" {
		if err := p.OpenGraphTitle.Validate(); err != nil {
			return &kernel.Error{Operation: op, Cause: err}
		}
	}

	if err := p.OpenGraphDescription.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := p.OpenGraphImage.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := p.CanonicalURL.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := p.SchemaType.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := p.Status.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := p.Slug.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := p.Category.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := p.validateWorkflowFields(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	return nil
}

// validateWorkflowFields validates publishing workflow fields.
func (p Post) validateWorkflowFields() error {
	const op = "Post.validateWorkflowFields"

	// Validate ApprovedBy if present
	if p.ApprovedBy != nil {
		if err := p.ApprovedBy.Validate(); err != nil {
			return &kernel.Error{Operation: op, Cause: err}
		}
	}

	// Scheduled posts must have a future PublishedAt date
	if p.Status == StatusScheduled {
		if p.PublishedAt == nil {
			return &kernel.Error{
				Code:      kernel.EInvalid,
				Message:   MPostScheduledDateRequired,
				Operation: op,
			}
		}

		if !p.PublishedAt.After(p.Clock.Now()) {
			return &kernel.Error{
				Code:      kernel.EInvalid,
				Message:   MPostScheduledDatePast,
				Operation: op,
			}
		}
	}

	return nil
}

// WordCount calculates content length for reading time estimation and content planning.
// Strips markdown formatting to provide accurate word counts for educational material.
func (p Post) WordCount() int {
	content := p.Content.String()

	// Remove markdown syntax and HTML tags for more accurate count
	content = kernel.StripMarkdown(content)

	// Split by whitespace and count non-empty strings
	words := strings.Fields(content)
	return len(words)
}

// EstimatedReadingTime helps learners plan study sessions by providing realistic time expectations.
// Calculated using average adult reading speed for educational content.
func (p Post) EstimatedReadingTime() int {
	wordCount := p.WordCount()
	minutes := float64(wordCount) / AverageWordsPerMinute

	// Round up to at least 1 minute
	return int(math.Max(1, math.Ceil(minutes)))
}

// IsPublished returns true if the post is published.
func (p Post) IsPublished() bool {
	return p.Status == StatusPublished
}

// IsDraft returns true if the post is a draft.
func (p Post) IsDraft() bool {
	return p.Status == StatusDraft
}

// CanBeEditedBy checks if a user can edit this post.
func (p Post) CanBeEditedBy(u user.PostPermissionChecker) bool {
	return u.CanEditPost(p)
}

// GetExcerpt returns a truncated version of the content for previews.
func (p Post) GetExcerpt(maxLength int) string {
	content := p.Content.String()

	// Strip markdown for cleaner excerpt.
	content = kernel.StripMarkdown(content)

	if len(content) <= maxLength {
		return content
	}

	// Truncate and add ellipsis, but try to break at word boundary.
	truncated := content[:maxLength]
	if lastSpace := strings.LastIndex(truncated, " "); lastSpace > maxLength/2 {
		truncated = truncated[:lastSpace]
	}

	return truncated + "..."
}

// HasFeaturedImage returns true if the post has a featured image.
func (p Post) HasFeaturedImage() bool {
	return p.FeaturedImage.String() != ""
}

// IsApproved returns true if the post has been approved.
func (p Post) IsApproved() bool {
	return p.ApprovedBy != nil && p.ApprovedAt != nil
}

// IsScheduled returns true if the post is scheduled for future publishing.
func (p Post) IsScheduled() bool {
	return p.Status == StatusScheduled
}

// IsReadyToPublish returns true if scheduled post should be published now.
func (p Post) IsReadyToPublish() bool {
	if !p.IsScheduled() || p.PublishedAt == nil {
		return false
	}

	return !p.PublishedAt.After(p.Clock.Now())
}

// CanTransitionTo checks if post can transition to new status.
func (p Post) CanTransitionTo(newStatus Status, u user.PostPermissionChecker) error {
	const op = "Post.CanTransitionTo"

	// Check if status transition is allowed
	if !p.Status.CanTransitionTo(newStatus) {
		return &kernel.Error{
			Code:      kernel.EInvalid,
			Message:   fmt.Sprintf(MPostInvalidStatusTransition, p.Status, newStatus),
			Operation: op,
		}
	}

	// Check user permissions for specific transitions
	switch newStatus {
	case StatusPublished:
		// Only approved posts can be published
		if !p.IsApproved() {
			return &kernel.Error{
				Code:      kernel.EInvalid,
				Message:   MPostCannotPublish,
				Operation: op,
			}
		}
		// Only admin/editor can publish
		if !u.HasAnyRole(user.RoleAdmin, user.RoleEditor) {
			return &kernel.Error{
				Code:      kernel.EForbidden,
				Message:   MPostCannotPublish,
				Operation: op,
			}
		}

	case StatusScheduled:
		// Only admin/editor can schedule
		if !u.HasAnyRole(user.RoleAdmin, user.RoleEditor) {
			return &kernel.Error{
				Code:      kernel.EForbidden,
				Message:   MPostCannotSchedule,
				Operation: op,
			}
		}

	case StatusArchived:
		// Only admin/editor can archive
		if !u.HasAnyRole(user.RoleAdmin, user.RoleEditor) {
			return &kernel.Error{
				Code:      kernel.EForbidden,
				Message:   fmt.Sprintf(MPostInvalidStatusTransition, p.Status, newStatus),
				Operation: op,
			}
		}

	case StatusDraft:
		// Published posts can go back to draft for major edits (admin/editor only)
		if p.Status == StatusPublished && !u.HasAnyRole(user.RoleAdmin, user.RoleEditor) {
			return &kernel.Error{
				Code:      kernel.EForbidden,
				Message:   fmt.Sprintf(MPostInvalidStatusTransition, p.Status, newStatus),
				Operation: op,
			}
		}
	}

	return nil
}

// Approve validates editorial approval for content publication in collaborative environments.
// Enforces business rules preventing self-approval and ensuring content quality control.
func (p Post) Approve(approver user.PostPermissionChecker) (Post, error) {
	const op = "Post.Approve"

	// Only admin/editor can approve.
	if !approver.HasAnyRole(user.RoleAdmin, user.RoleEditor) {
		return p, &kernel.Error{
			Code:      kernel.EForbidden,
			Message:   MPostCannotApprove,
			Operation: op,
		}
	}

	// Cannot approve own posts (unless admin).
	if p.Owner == approver.GetID() && !approver.HasRole(user.RoleAdmin) {
		return p, &kernel.Error{
			Code:      kernel.EForbidden,
			Message:   MPostCannotApprove,
			Operation: op,
		}
	}

	now := p.Clock.Now()
	approverID := approver.GetID()

	updatedPost := p
	updatedPost.ApprovedBy = &approverID
	updatedPost.ApprovedAt = &now
	updatedPost.UpdatedAt = now

	return updatedPost, nil
}

// Schedule schedules the post for future publishing.
func (p Post) Schedule(publishAt time.Time, u user.PostPermissionChecker) (Post, error) {
	const op = "Post.Schedule"

	// Check if user can schedule
	if err := p.CanTransitionTo(StatusScheduled, u); err != nil {
		return p, &kernel.Error{Operation: op, Cause: err}
	}

	// Validate future date
	if !publishAt.After(p.Clock.Now()) {
		return p, &kernel.Error{
			Code:      kernel.EInvalid,
			Message:   MPostScheduledDatePast,
			Operation: op,
		}
	}

	updatedPost := p
	updatedPost.Status = StatusScheduled
	updatedPost.PublishedAt = &publishAt
	updatedPost.UpdatedAt = p.Clock.Now()

	return updatedPost, nil
}

// Publish publishes the post immediately.
func (p Post) Publish(u user.PostPermissionChecker) (Post, error) {
	const op = "Post.Publish"

	// Check if user can publish.
	if err := p.CanTransitionTo(StatusPublished, u); err != nil {
		return p, &kernel.Error{Operation: op, Cause: err}
	}

	now := p.Clock.Now()

	updatedPost := p
	updatedPost.Status = StatusPublished
	updatedPost.PublishedAt = &now
	updatedPost.UpdatedAt = now

	return updatedPost, nil
}

// GetOwner returns the post owner ID for permission checks.
func (p Post) GetOwner() kernel.ID[user.User] {
	return p.Owner
}

// GetStatus returns the post status as string for permission checks.
func (p Post) GetStatus() string {
	return string(p.Status)
}
