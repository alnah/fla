package subscription

import (
	"github.com/alnah/fla/internal/domain/kernel"
)

// Status represents the status of a subscription
type Status string

const (
	StatusActive       Status = "active"
	StatusUnsubscribed Status = "unsubscribed"
	StatusBounced      Status = "bounced"    // Email bounced
	StatusComplained   Status = "complained" // Spam complaint
)

func (s Status) String() string { return string(s) }

func (s Status) Validate() error {
	const op = "Status.Validate"

	switch s {
	case StatusActive, StatusUnsubscribed, StatusBounced, StatusComplained:
		return nil
	default:
		return &kernel.Error{
			Code:      kernel.EInvalid,
			Message:   "Invalid subscription status.",
			Operation: op,
		}
	}
}
