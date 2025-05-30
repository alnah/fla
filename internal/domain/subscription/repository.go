package subscription

import (
	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/shared"
)

// SubscriptionReader defines read-only operations for subscription access.
// Used by subscription status pages and customer service tools.
type SubscriptionReader interface {
	// GetByID retrieves specific subscriptions for customer support and account management.
	// Used by help desk systems and subscription status verification tools.
	GetByID(subscriptionID kernel.ID[Subscription]) (*Subscription, error)

	// GetByEmail finds subscriptions for unsubscribe links and customer inquiries.
	// Used by email preference pages and customer self-service portals.
	GetByEmail(email shared.Email) (*Subscription, error)
}

// SubscriptionWriter defines modification operations for subscription lifecycle.
// Used by signup forms and subscription management workflows.
type SubscriptionWriter interface {
	// Create adds new subscribers to enable newsletter delivery and audience growth.
	// Used by signup forms, lead magnets, and subscription conversion funnels.
	Create(subscription Subscription) error

	// Update modifies subscription status for preference changes and lifecycle management.
	// Used by unsubscribe processes, preference updates, and bounce handling.
	Update(subscription Subscription) error

	// Delete permanently removes subscriptions for GDPR compliance and data cleanup.
	// Used by data retention policies and complete account deletion requests.
	Delete(subscriptionID kernel.ID[Subscription]) error
}

// SubscriptionLister provides bulk access to subscriber collections.
// Used by newsletter analytics and subscriber management dashboards.
type SubscriptionLister interface {
	// GetActiveSubscriptions returns engaged subscribers for newsletter delivery.
	// Used by email marketing systems to target active audience members.
	GetActiveSubscriptions() ([]Subscription, error)

	// GetAllSubscriptions returns complete subscriber database for analytics and reporting.
	// Used by dashboard metrics, subscriber growth analysis, and data exports.
	GetAllSubscriptions() ([]Subscription, error)
}

// SubscriptionValidator provides data integrity checks for email management.
// Used by signup forms and APIs to prevent duplicate subscriptions.
type SubscriptionValidator interface {
	// ExistsByEmail prevents duplicate subscriptions for email uniqueness enforcement.
	// Used by signup forms to check if email is already subscribed before creating new records.
	ExistsByEmail(email shared.Email) (bool, error)
}

// CampaignTargeter identifies subscribers for content distribution.
// Used by email marketing automation and newsletter delivery systems.
type CampaignTargeter interface {
	// GetSubscribersForNewPost returns active subscribers ready to receive content notifications.
	// Used by automated email systems when new blog posts are published.
	GetSubscribersForNewPost() ([]Subscription, error)
}

// Composed interfaces for common use cases

// SubscriptionService combines core operations for public subscription management.
// Used by public-facing subscription forms and user preference pages.
type SubscriptionService interface {
	SubscriptionReader
	SubscriptionWriter
	SubscriptionValidator
}

// NewsletterManager handles subscriber audience management for email campaigns.
// Used by marketing tools and email automation systems.
type NewsletterManager interface {
	SubscriptionLister
	CampaignTargeter
}

// SubscriptionAdmin provides complete subscriber database control.
// Used by admin dashboards and customer service management tools.
type SubscriptionAdmin interface {
	SubscriptionReader
	SubscriptionWriter
	SubscriptionLister
	SubscriptionValidator
}

// EmailMarketer combines targeting and management for campaign operations.
// Used by email marketing platforms and automated notification systems.
type EmailMarketer interface {
	SubscriptionLister
	CampaignTargeter
	SubscriptionValidator
}

// Full repository interface for implementations that provide everything.
// Most concrete implementations (like PostgresSubscriptionRepository) will implement this.
type Repository interface {
	SubscriptionReader
	SubscriptionWriter
	SubscriptionLister
	SubscriptionValidator
	CampaignTargeter
}
