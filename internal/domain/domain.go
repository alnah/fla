// Package domain provides the core business logic and entities for a language learning blog.
//
// This facade re-exports the main types from sub-packages for backward compatibility
// and convenient access to the most commonly used domain types.
package domain

import (
	"github.com/alnah/fla/internal/domain/category"
	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/post"
	"github.com/alnah/fla/internal/domain/shared"
	"github.com/alnah/fla/internal/domain/subscription"
	"github.com/alnah/fla/internal/domain/tag"
	"github.com/alnah/fla/internal/domain/user"
)

// Re-export kernel types
type (
	// Clock exposes the current UTC time.
	// Domain code depends only on this interface.
	Clock = kernel.Clock

	// Error provides structured error handling with operation context and error chaining.
	// Enables precise error diagnosis and consistent error responses across the domain.
	Error = kernel.Error
)

// Re-export error codes
const (
	EConflict  = kernel.EConflict  // Action cannot be performed due to business rule conflicts
	EInternal  = kernel.EInternal  // Internal system error requiring technical investigation
	EInvalid   = kernel.EInvalid   // Validation failed on user input or data constraints
	EForbidden = kernel.EForbidden // Action not allowed due to permission restrictions
	ENotFound  = kernel.ENotFound  // Requested entity does not exist in the system
)

// Re-export error functions
var (
	// ErrorCode extracts the machine-readable error classification for handling logic.
	// Returns the most specific error code available in the error chain.
	ErrorCode = kernel.ErrorCode

	// ErrorMessage retrieves the human-readable error description for user display.
	// Provides clear, actionable error messages while maintaining security.
	ErrorMessage = kernel.ErrorMessage
)

// Re-export shared types
type (
	// Email represents validated email addresses for user communication.
	// Ensures deliverable addresses for notifications and account management.
	Email = shared.Email

	// FirstName represents personal given name with optional validation.
	// Enables personalization while accommodating diverse naming conventions.
	FirstName = shared.FirstName

	// LastName represents family name with flexible validation requirements.
	// Supports diverse naming traditions while maintaining data consistency.
	LastName = shared.LastName

	// Username represents unique user identifiers for login and public display.
	// Ensures usernames are URL-safe and meet platform conventions.
	Username = shared.Username

	// Title represents content headlines with length validation for readability.
	// Ensures titles are descriptive enough while maintaining display compatibility.
	Title = shared.Title

	// Description provides explanatory text for entities with length constraints.
	// Enables rich metadata while maintaining UI display boundaries.
	Description = shared.Description

	// Slug represents URL-friendly identifiers for content addressing.
	// Enables SEO-optimized URLs and clean content navigation paths.
	Slug = shared.Slug

	// Datetime wraps time with domain-specific validation for audit trails.
	// Ensures temporal data integrity and prevents logical inconsistencies.
	Datetime = shared.Datetime

	// Pagination handles content listing with page-based navigation for improved user experience.
	// Provides offset calculations and navigation state for repository queries.
	Pagination = shared.Pagination
)

// Re-export shared constructors
var (
	// NewEmail creates validated email address with format verification.
	// Prevents invalid addresses that would cause delivery failures.
	NewEmail = shared.NewEmail

	// NewFirstName creates validated first name with cultural sensitivity.
	// Supports optional names while preventing excessively long values.
	NewFirstName = shared.NewFirstName

	// NewLastName creates validated family name with cultural accommodation.
	// Balances name validation with international naming flexibility.
	NewLastName = shared.NewLastName

	// NewUsername creates validated username with character and length restrictions.
	// Prevents conflicts and ensures usernames work across web systems.
	NewUsername = shared.NewUsername

	// NewTitle creates validated content title with readability requirements.
	// Balances descriptive content with practical length constraints for SEO and UI.
	NewTitle = shared.NewTitle

	// NewDescription creates validated descriptive text with length checking.
	// Ensures descriptions fit within meta tag limits and UI components.
	NewDescription = shared.NewDescription

	// NewSlug generates URL-safe slug from input text with automatic formatting.
	// Creates clean URLs while preserving content meaning and SEO value.
	NewSlug = shared.NewSlug

	// NewDatetime creates validated datetime for historical record keeping.
	// Enforces past-only timestamps for audit trail integrity.
	NewDatetime = shared.NewDatetime

	// NewDatetimeNow captures current moment for timestamp generation.
	// Provides consistent time source for creation and modification tracking.
	NewDatetimeNow = shared.NewDatetimeNow

	// NewPagination creates a new pagination with validation
	NewPagination = shared.NewPagination
)

// Re-export post types
type (
	// Post represents a complete learning article with content, metadata, and publishing workflow.
	// Designed for educational blogs with SEO optimization and approval processes.
	Post = post.Post

	// PostContent represents the main body text of educational blog posts.
	// Enforces minimum length for substantial content and maximum for readability.
	PostContent = post.PostContent

	// Status represents the publication state of blog content in the editorial workflow.
	// Controls content visibility and enables staged publication processes.
	Status = post.Status

	// SchemaType represents Schema.org markup types for structured data
	SchemaType = post.SchemaType

	// PostsList combines paginated posts with navigation metadata for content browsing.
	// Enables efficient content listing with proper page controls and item counts.
	PostsList = post.PostsList

	// NewPostParams holds the parameters needed to create a new post.
	NewPostParams = post.NewPostParams
)

// PostID provides unique identification for blog post entities.
// Enables post retrieval, linking, and relationship management.
type PostID = kernel.ID[post.Post]

// NewPostID creates validated post identifier for entity referencing.
// Ensures post identity integrity throughout the system.
var NewPostID = func(id string) (PostID, error) {
	return kernel.NewID[post.Post](id)
}

// Re-export post constructors and constants
var (
	// NewPost creates a validated post with automatic slug generation and workflow initialization.
	// Ensures all required content and metadata are properly structured for publication.
	NewPost = post.NewPost

	// NewPostContent creates validated post content with educational length requirements.
	// Ensures posts provide sufficient learning value while remaining digestible.
	NewPostContent = post.NewPostContent

	// NewPostsList creates a new paginated posts list
	NewPostsList = post.NewPostsList
)

const (
	StatusDraft     = post.StatusDraft     // Content in development, not visible to public
	StatusPublished = post.StatusPublished // Live content available to all readers
	StatusArchived  = post.StatusArchived  // Historical content removed from active circulation
	StatusScheduled = post.StatusScheduled // Content queued for future publication
)

// PostRepository defines essential data operations for post management.
// Provides clean interface between domain logic and data persistence layer.
type PostRepository = post.Repository

// Re-export user types
type (
	// User represents an authenticated person with role-based permissions in the blogging system.
	// Manages identity, profile information, and content access controls.
	User = user.User

	// Role defines permission levels for system access and content management.
	// Enables fine-grained access control for collaborative blogging workflows.
	Role = user.Role

	// SocialProfile represents validated social media profile links.
	// Ensures profile URLs are correctly formatted and platform-appropriate.
	SocialProfile = user.SocialProfile

	// SocialMediaURL defines supported social media platforms for user profiles.
	// Enables standardized social media integration across the platform.
	SocialMediaURL = user.SocialMediaURL

	// NewUserParams holds essential information for creating user accounts.
	// Separates required identity fields from optional profile information.
	NewUserParams = user.NewUserParams
)

// UserID provides unique identification for user accounts in the system.
// Enables user authentication, authorization, and content ownership tracking.
type UserID = kernel.ID[user.User]

// NewUserID creates validated user identifier for account management.
// Ensures user identity integrity throughout authentication and authorization.
var NewUserID = func(id string) (UserID, error) {
	return kernel.NewID[user.User](id)
}

// Re-export user constructors and constants
var (
	// NewUser creates a validated user account with proper role assignment.
	// Ensures user data integrity and permission system initialization.
	NewUser = user.NewUser

	// NewSocialProfile creates validated social media profile with platform-specific rules.
	// Prevents broken social links and ensures proper platform URL formatting.
	NewSocialProfile = user.NewSocialProfile
)

const (
	RoleAdmin      = user.RoleAdmin      // Full system access and user management
	RoleEditor     = user.RoleEditor     // Content management and publication control
	RoleAuthor     = user.RoleAuthor     // Content creation and own post management
	RoleSubscriber = user.RoleSubscriber // Basic access for content consumption
	RoleVisitor    = user.RoleVisitor    // Anonymous read-only access
	RoleMachine    = user.RoleMachine    // Automated system access for integrations
)

// Re-export category types
type (
	// Category represents a hierarchical content organization unit for educational blogs.
	// Categories enable structured navigation through learning materials (Level → Skill → Topic).
	Category = category.Category

	// CategoryName represents user-facing category titles with length validation.
	// Ensures category names are meaningful and fit within UI constraints.
	CategoryName = category.CategoryName

	// CategoryPath represents the complete hierarchy trail from root to target category.
	// Enables URL generation and breadcrumb navigation for educational content structure.
	CategoryPath = category.CategoryPath

	// CategoryBreadcrumb represents navigation trail elements for hierarchical browsing.
	// Enables users to understand their location and navigate back through category levels.
	CategoryBreadcrumb = category.CategoryBreadcrumb

	// NewCategoryParams holds the essential information needed to create a learning category.
	// Used to ensure all required fields are provided during category creation.
	NewCategoryParams = category.NewCategoryParams

	// PathService handles URL generation and parsing for hierarchical navigation.
	// Enables clean URLs and breadcrumb navigation for educational content structure.
	CategoryPathService = category.PathService
)

// CategoryID provides unique identification for category entities in the system.
// Ensures referential integrity and enables efficient category lookups.
type CategoryID = kernel.ID[category.Category]

// NewCategoryID creates a validated category identifier with presence checking.
// Prevents empty or invalid identifiers that would break category references.
var NewCategoryID = func(id string) (CategoryID, error) {
	return kernel.NewID[category.Category](id)
}

// Re-export category constructors
var (
	// NewCategory creates a validated category with automatic slug generation.
	// Ensures category hierarchy rules and data integrity are maintained.
	NewCategory = category.NewCategory

	// NewCategoryName creates a validated category name with proper length limits.
	// Maintains consistent category naming and prevents UI layout issues.
	NewCategoryName = category.NewCategoryName

	// NewPathService creates path service with repository dependency.
	// Provides URL management capabilities for category-based content organization.
	NewCategoryPathService = category.NewPathService
)

// Re-export tag types
type (
	// Tag represents a descriptive label for categorizing and discovering blog content.
	// Tags enable cross-cutting content organization beyond hierarchical categories.
	Tag = tag.Tag

	// TagName represents descriptive labels for content discovery and organization.
	// Enables flexible content categorization beyond hierarchical structure.
	TagName = tag.TagName
)

// TagID provides unique identification for content tagging system.
// Enables tag-based content discovery and cross-reference functionality.
type TagID = kernel.ID[tag.Tag]

// NewTagID creates validated tag identifier for content organization.
// Ensures tag system integrity and efficient content categorization.
var NewTagID = func(id string) (TagID, error) {
	return kernel.NewID[tag.Tag](id)
}

// Re-export tag constructors
var (
	// NewTag creates a validated tag with proper metadata tracking.
	// Ensures tag consistency and audit trail for content organization.
	NewTag = tag.NewTag

	// NewTagName creates validated tag label with appropriate length constraints.
	// Ensures tags are meaningful while fitting within UI and database limits.
	NewTagName = tag.NewTagName
)

// Re-export subscription types
type (
	// Subscription manages email newsletter enrollment for blog content notifications.
	// Supports subscription lifecycle with unsubscribe/resubscribe and bounce handling.
	Subscription = subscription.Subscription

	// Status represents the status of a subscription
	SubscriptionStatus = subscription.Status

	// NewSubscriptionParams holds the parameters needed to create a new subscription
	NewSubscriptionParams = subscription.NewSubscriptionParams
)

// SubscriptionID provides unique identification for email subscription records.
// Enables subscription management and unsubscribe link generation.
type SubscriptionID = kernel.ID[subscription.Subscription]

// NewSubscriptionID creates validated subscription identifier for record tracking.
// Ensures subscription identity integrity for email campaign management.
var NewSubscriptionID = func(id string) (SubscriptionID, error) {
	return kernel.NewID[subscription.Subscription](id)
}

// Re-export subscription constructors and constants
// NewSubscription creates an active email subscription with immediate notification enrollment.
// Validates email format and subscriber information for reliable delivery.
var NewSubscription = subscription.NewSubscription

const (
	SubscriptionStatusActive       = subscription.StatusActive       // Subscription is active
	SubscriptionStatusUnsubscribed = subscription.StatusUnsubscribed // User has unsubscribed
	SubscriptionStatusBounced      = subscription.StatusBounced      // Email bounced
	SubscriptionStatusComplained   = subscription.StatusComplained   // Spam complaint
)

// Re-export repository interfaces
type (
	// CategoryRepository defines essential data operations for category management.
	// Provides clean interface between domain logic and data persistence layer.
	CategoryRepository = category.Repository

	// SubscriptionRepository defines email subscription data operations.
	// Manages subscriber lifecycle and campaign targeting for content notifications.
	SubscriptionRepository = subscription.Repository
)

// URL type aliases for backward compatibility
// URL represents validated URLs for resources with security validation.
// Ensures sources are accessible and use secure protocols.
type URL = kernel.URL[any]

// NewURL creates validated URL with security and format checking.
// Prevents broken links and ensures secure resource loading.
var NewURL = func(urlStr string) (URL, error) {
	return kernel.NewURL[any](urlStr)
}

const (
	SocialMediaTwitter   = user.SocialMediaTwitter   // SocialMediaTwitter represents Twitter/X platform.
	SocialMediaLinkedIn  = user.SocialMediaLinkedIn  // SocialMediaLinkedIn represents LinkedIn platform.
	SocialMediaInstagram = user.SocialMediaInstagram // SocialMediaInstagram represents Instagram platform.
	SocialMediaTikTok    = user.SocialMediaTikTok    // SocialMediaTikTok represents TikTok platform.
	SocialMediaYouTube   = user.SocialMediaYouTube   // SocialMediaYouTube represents YouTube platform.
	SocialMediaGitHub    = user.SocialMediaGitHub    // SocialMediaGitHub represents GitHub platform.
)
