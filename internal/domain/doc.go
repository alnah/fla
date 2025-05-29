// Package domain provides the core business logic and entities for a language learning blog.
//
// This domain implements a complete blogging system optimized for language education,
// featuring hierarchical categories, content management, email subscriptions, and SEO optimization.
//
// # Architecture
//
// The domain follows Domain-Driven Design principles with a modular structure:
//
//	domain/
//	├── kernel/        # Core types and utilities (Clock, Error, ID[T], URL[T], validators)
//	├── shared/        # Shared value objects (Email, Title, Pagination, etc.)
//	├── post/          # Post aggregate (Post, Status, SEO types)
//	├── user/          # User aggregate (User, Role, permissions)
//	├── category/      # Category aggregate (Category, path services)
//	├── subscription/  # Subscription aggregate (email management)
//	├── tag/           # Tag aggregate (content tagging)
//	└── domain.go      # Facade for backward compatibility
//
// # Core Features
//
// Content Management:
//   - Hierarchical categories (Level → Skill → Topic: A1 → Reading → Sports)
//   - Rich post content with markdown support
//   - Comprehensive SEO and social media optimization
//   - Approval workflow for collaborative editing
//   - Scheduled publishing
//
// User System:
//   - Role-based permissions (Admin, Editor, Author)
//   - Profile management with social media links
//   - Content ownership and editing rules
//
// Email Subscriptions:
//   - Anonymous subscriptions with first name and email
//   - Subscription lifecycle management (subscribe, unsubscribe, resubscribe)
//   - Email bounce and complaint handling
//
// # Usage Examples
//
// Creating a hierarchical category structure:
//
//	// Create root category (A1 level)
//	levelA1, err := domain.NewCategory(domain.NewCategoryParams{
//	    CategoryID: domain.NewCategoryID("a1-level"),
//	    Name:       domain.NewCategoryName("A1"),
//	    CreatedBy:  userID,
//	    Clock:      clock,
//	})
//
//	// Create skill category under A1
//	readingSkill, err := domain.NewCategory(domain.NewCategoryParams{
//	    CategoryID: domain.NewCategoryID("a1-reading"),
//	    Name:       domain.NewCategoryName("Compréhension écrite"),
//	    ParentID:   &levelA1.CategoryID,
//	    CreatedBy:  userID,
//	    Clock:      clock,
//	})
//
//	// Create topic category under reading skill
//	sportsCategory, err := domain.NewCategory(domain.NewCategoryParams{
//	    CategoryID: domain.NewCategoryID("a1-reading-sports"),
//	    Name:       domain.NewCategoryName("Sports"),
//	    ParentID:   &readingSkill.CategoryID,
//	    CreatedBy:  userID,
//	    Clock:      clock,
//	})
//
// Creating and publishing a blog post:
//
//	// Create a new post
//	newPost, err := domain.NewPost(domain.NewPostParams{
//	    PostID:  domain.NewPostID("football-lesson"),
//	    Owner:   authorID,
//	    Title:   domain.NewTitle("Jouer au Football"),
//	    Content: domain.NewPostContent("Le football est un sport..."),
//	    Status:  domain.StatusDraft,
//	    Category: sportsCategory,
//	    SEOTitle: domain.NewTitle("French A1 Reading: Playing Football - Learn French"),
//	    SEODescription: domain.NewDescription("Practice French reading comprehension..."),
//	    Clock:   clock,
//	})
//
//	// Editor approves the post
//	approvedPost, err := newPost.Approve(editor)
//
//	// Admin publishes the post
//	publishedPost, err := approvedPost.Publish(admin)
//
// Managing email subscriptions:
//
//	// User subscribes to newsletter
//	sub, err := domain.NewSubscription(domain.NewSubscriptionParams{
//	    SubscriptionID: domain.NewSubscriptionID("sub-123"),
//	    FirstName:      domain.NewFirstName("Marie"),
//	    Email:          domain.NewEmail("marie@example.com"),
//	    Clock:          clock,
//	})
//
//	// User unsubscribes
//	unsubscribed, err := sub.Unsubscribe()
//
//	// User resubscribes later
//	resubscribed, err := unsubscribed.Resubscribe()
//
// URL generation and breadcrumbs:
//
//	// Build category URL: "a1/comprehension-ecrite/sports"
//	pathService := domain.NewCategoryPathService(repository)
//	url, err := pathService.BuildURL(sportsCategory.CategoryID)
//
//	// Generate breadcrumbs for navigation
//	breadcrumbs, err := pathService.GetBreadcrumbs(sportsCategory.CategoryID)
//	// Result: [A1] > [Compréhension écrite] > [Sports]
//
// Using generic ID types:
//
//	// Type-safe IDs prevent mixing different entity IDs
//	var postID domain.PostID = "post-123"
//	var userID domain.UserID = "user-456"
//	// postID = userID // Compile error: type mismatch
//
// Pagination for post listings:
//
//	// Create pagination for category page
//	pagination, err := domain.NewPagination(1, 10, 45) // Page 1, 10 per page, 45 total
//
//	// Get posts from repository
//	posts := repository.GetPostsByCategory(categoryID, pagination)
//
//	// Create paginated result
//	result := domain.NewPostsList(posts, pagination)
//
// # Business Rules
//
// Content Publishing:
//   - Authors create posts but cannot publish directly
//   - Editors and Admins can approve and publish posts
//   - Editors cannot approve their own posts (only Admin can)
//   - Published posts can be unpublished back to draft (Admin/Editor only)
//   - Scheduled posts are automatically published when the time arrives
//   - Status transitions follow a defined state machine (see post/status.go)
//
// Category Hierarchy:
//   - Maximum 3 levels deep (Level → Skill → Topic)
//   - Slug must be unique within the same parent category
//   - Categories cannot be deleted if they have posts or child categories
//   - URL paths are generated automatically: /a1/comprehension-ecrite/sports
//
// User Permissions:
//   - Admin: Full system access
//   - Editor: Can manage content and approve posts
//   - Author: Can create and edit own posts
//   - Visitor: Can view published content
//   - Permissions are checked through methods like CanEditPost, CanPublish, etc.
//
// Email Subscriptions:
//   - One subscription per email address
//   - Subscribers can unsubscribe and resubscribe
//   - Bounced emails and spam complaints automatically disable subscriptions
//   - Only active subscribers receive new post notifications
//
// # Error Handling
//
// All domain operations return descriptive errors with operation context:
//
//	post, err := post.NewPost(params)
//	if err != nil {
//	    code := kernel.ErrorCode(err)        // "invalid", "forbidden", etc.
//	    message := kernel.ErrorMessage(err)  // Human-readable message
//	    // Handle based on error type
//	}
//
// # SEO and Social Media
//
// Posts support comprehensive SEO optimization:
//   - Custom SEO titles and descriptions
//   - Open Graph metadata for social sharing
//   - Canonical URLs for duplicate content prevention
//   - Schema.org structured data for search engines
//   - Automatic fallbacks when custom values aren't provided
//
// # Integration Points
//
// The domain provides clean interfaces for external systems:
//   - Repository interfaces in each package (category.Repository, subscription.Repository, etc.)
//   - Email services: Send notifications via Postmark, Mailchimp, etc.
//   - Static site generators: Export content for JAMstack deployment
//   - Search engines: Full-text search integration
//   - Event system: Can be extended with domain events for async processing
//
// # Performance Optimizations
//
// This modular structure provides several performance benefits:
//   - Smaller compilation units reduce build times
//   - Generic types eliminate code duplication
//   - Table-driven validation reduces function call overhead
//   - Direct time.Time usage removes wrapper allocation
//   - Precompiled regex patterns for slug generation
//
// This domain is specifically optimized for language learning blogs but can be
// adapted for other educational content or general blogging needs.
package domain
