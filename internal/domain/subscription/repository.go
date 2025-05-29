package subscription

import (
	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/shared"
)

// Repository defines email subscription data operations.
// Manages subscriber lifecycle and campaign targeting for content notifications.
type Repository interface {
	// Basic CRUD operations for subscription management
	Create(subscription Subscription) error
	GetByID(subscriptionID kernel.ID[Subscription]) (*Subscription, error)
	GetByEmail(email shared.Email) (*Subscription, error)
	Update(subscription Subscription) error
	Delete(subscriptionID kernel.ID[Subscription]) error

	// Subscription management for campaign targeting
	GetActiveSubscriptions() ([]Subscription, error)
	GetAllSubscriptions() ([]Subscription, error)

	// Duplicate prevention for email uniqueness
	ExistsByEmail(email shared.Email) (bool, error)

	// Campaign targeting for content notifications
	GetSubscribersForNewPost() ([]Subscription, error)
}
