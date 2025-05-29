package subscription

import (
	"fmt"
	"time"

	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/shared"
)

const (
	MSubscriptionEmailExists   string = "Email is already subscribed."
	MSubscriptionNotFound      string = "Subscription not found."
	MSubscriptionAlreadyActive string = "Subscription is already active."
	MSubscriptionNotActive     string = "Subscription is not active."
)

// Subscription manages email newsletter enrollment for blog content notifications.
// Supports subscription lifecycle with unsubscribe/resubscribe and bounce handling.
type Subscription struct {
	// Identity
	SubscriptionID kernel.ID[Subscription]

	// Subscriber Info
	FirstName shared.FirstName
	Email     shared.Email

	// Status
	Status Status

	// Preferences
	IsActive bool // Quick check for active subscriptions

	// Meta
	SubscribedAt   time.Time
	UnsubscribedAt *time.Time // When they unsubscribed (nil if still subscribed)
	UpdatedAt      time.Time

	// DI
	Clock kernel.Clock
}

// NewSubscriptionParams holds the parameters needed to create a new subscription
type NewSubscriptionParams struct {
	// Required
	SubscriptionID kernel.ID[Subscription]
	FirstName      shared.FirstName
	Email          shared.Email

	// DI
	Clock kernel.Clock
}

// NewSubscription creates an active email subscription with immediate notification enrollment.
// Validates email format and subscriber information for reliable delivery.
func NewSubscription(p NewSubscriptionParams) (Subscription, error) {
	const op = "NewSubscription"

	now := p.Clock.Now()

	subscription := Subscription{
		SubscriptionID: p.SubscriptionID,
		FirstName:      p.FirstName,
		Email:          p.Email,
		Status:         StatusActive,
		IsActive:       true,
		SubscribedAt:   now,
		UnsubscribedAt: nil,
		UpdatedAt:      now,
		Clock:          p.Clock,
	}

	if err := subscription.Validate(); err != nil {
		return Subscription{}, &kernel.Error{Operation: op, Cause: err}
	}

	return subscription, nil
}

// Validate performs validation on the subscription
func (s Subscription) Validate() error {
	const op = "Subscription.Validate"

	if err := s.SubscriptionID.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := s.FirstName.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := s.Email.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := s.Status.Validate(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	return nil
}

// String returns a string representation of the subscription
func (s Subscription) String() string {
	return fmt.Sprintf("Subscription{"+
		"ID: %q, "+
		"Name: %q, "+
		"Email: %q, "+
		"Status: %q, "+
		"Active: %t, "+
		"SubscribedAt: %s"+
		"}",
		s.SubscriptionID,
		s.FirstName,
		s.Email,
		s.Status,
		s.IsActive,
		s.SubscribedAt.Format(time.RFC3339),
	)
}

// Unsubscribe marks the subscription as unsubscribed
func (s Subscription) Unsubscribe() (Subscription, error) {
	const op = "Subscription.Unsubscribe"

	// Can only unsubscribe active subscriptions
	if s.Status != StatusActive {
		return s, &kernel.Error{
			Code:      kernel.EConflict,
			Message:   MSubscriptionNotActive,
			Operation: op,
		}
	}

	now := s.Clock.Now()

	updated := s
	updated.Status = StatusUnsubscribed
	updated.IsActive = false
	updated.UnsubscribedAt = &now
	updated.UpdatedAt = now

	return updated, nil
}

// Resubscribe reactivates an unsubscribed subscription
func (s Subscription) Resubscribe() (Subscription, error) {
	const op = "Subscription.Resubscribe"

	if s.Status == StatusActive {
		return s, &kernel.Error{
			Code:      kernel.EConflict,
			Message:   MSubscriptionAlreadyActive,
			Operation: op,
		}
	}

	// Can only resubscribe if previously unsubscribed (not bounced/complained)
	if s.Status != StatusUnsubscribed {
		return s, &kernel.Error{
			Code:      kernel.EInvalid,
			Message:   "Cannot resubscribe: subscription was not voluntarily unsubscribed",
			Operation: op,
		}
	}

	now := s.Clock.Now()

	updated := s
	updated.Status = StatusActive
	updated.IsActive = true
	updated.UnsubscribedAt = nil
	updated.UpdatedAt = now

	return updated, nil
}

// MarkAsBounced marks the subscription as bounced (bad email)
func (s Subscription) MarkAsBounced() (Subscription, error) {
	now := s.Clock.Now()

	updated := s
	updated.Status = StatusBounced
	updated.IsActive = false
	updated.UpdatedAt = now

	return updated, nil
}

// MarkAsComplained marks the subscription as complained (spam report)
func (s Subscription) MarkAsComplained() (Subscription, error) {
	now := s.Clock.Now()

	updated := s
	updated.Status = StatusComplained
	updated.IsActive = false
	updated.UpdatedAt = now

	return updated, nil
}

// IsSubscribed returns true if subscription is active
func (s Subscription) IsSubscribed() bool {
	return s.IsActive && s.Status == StatusActive
}

// CanReceiveEmails returns true if subscription can receive emails
func (s Subscription) CanReceiveEmails() bool {
	return s.IsSubscribed()
}

// GetDisplayName returns the display name for the subscriber
func (s Subscription) GetDisplayName() string {
	if s.FirstName.String() != "" {
		return s.FirstName.String()
	}
	return s.Email.String()
}
