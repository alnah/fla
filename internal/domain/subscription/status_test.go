package subscription_test

import (
	"testing"

	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/subscription"
)

func TestStatus_String(t *testing.T) {
	tests := []struct {
		status subscription.Status
		want   string
	}{
		{subscription.StatusActive, "active"},
		{subscription.StatusUnsubscribed, "unsubscribed"},
		{subscription.StatusBounced, "bounced"},
		{subscription.StatusComplained, "complained"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.status.String()

			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStatus_Validate(t *testing.T) {
	t.Run("valid statuses pass", func(t *testing.T) {
		validStatuses := []subscription.Status{
			subscription.StatusActive,
			subscription.StatusUnsubscribed,
			subscription.StatusBounced,
			subscription.StatusComplained,
		}

		for _, status := range validStatuses {
			t.Run(string(status), func(t *testing.T) {
				err := status.Validate()

				assertNoError(t, err)
			})
		}
	})

	t.Run("invalid status fails", func(t *testing.T) {
		invalidStatuses := []subscription.Status{
			"",
			"pending",
			"verified",
			"blocked",
			"ACTIVE", // case sensitive
			"Active",
			"active ",
			" active",
		}

		for _, status := range invalidStatuses {
			t.Run(string(status), func(t *testing.T) {
				err := status.Validate()

				assertError(t, err)
				assertErrorCode(t, err, kernel.EInvalid)
			})
		}
	})
}

func TestStatusConstants(t *testing.T) {
	// Ensure status constants have expected values
	tests := []struct {
		name   string
		status subscription.Status
		want   string
	}{
		{"StatusActive", subscription.StatusActive, "active"},
		{"StatusUnsubscribed", subscription.StatusUnsubscribed, "unsubscribed"},
		{"StatusBounced", subscription.StatusBounced, "bounced"},
		{"StatusComplained", subscription.StatusComplained, "complained"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("got %q, want %q", tt.status, tt.want)
			}
		})
	}
}
