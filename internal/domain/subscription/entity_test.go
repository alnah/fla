package subscription_test

import (
	"testing"
	"time"

	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/shared"
	"github.com/alnah/fla/internal/domain/subscription"
)

type stubClock struct {
	t time.Time
}

func (s *stubClock) Now() time.Time { return s.t }

func TestNewSubscription(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	clock := &stubClock{t: fixedTime}

	validSubscriptionID, _ := kernel.NewID[subscription.Subscription]("sub-123")
	validFirstName, _ := shared.NewFirstName("John")
	validEmail, _ := shared.NewEmail("john@example.com")

	t.Run("creates active subscription", func(t *testing.T) {
		params := subscription.NewSubscriptionParams{
			SubscriptionID: validSubscriptionID,
			FirstName:      validFirstName,
			Email:          validEmail,
			Clock:          clock,
		}

		got, err := subscription.NewSubscription(params)

		assertNoError(t, err)

		if got.SubscriptionID != validSubscriptionID {
			t.Errorf("SubscriptionID: got %v, want %v", got.SubscriptionID, validSubscriptionID)
		}
		if got.FirstName != validFirstName {
			t.Errorf("FirstName: got %v, want %v", got.FirstName, validFirstName)
		}
		if got.Email != validEmail {
			t.Errorf("Email: got %v, want %v", got.Email, validEmail)
		}
		if got.Status != subscription.StatusActive {
			t.Errorf("Status: got %v, want %v", got.Status, subscription.StatusActive)
		}
		if !got.IsActive {
			t.Error("expected IsActive to be true")
		}
		if !got.SubscribedAt.Equal(fixedTime) {
			t.Errorf("SubscribedAt: got %v, want %v", got.SubscribedAt, fixedTime)
		}
		if got.UnsubscribedAt != nil {
			t.Error("expected UnsubscribedAt to be nil")
		}
		if !got.UpdatedAt.Equal(fixedTime) {
			t.Errorf("UpdatedAt: got %v, want %v", got.UpdatedAt, fixedTime)
		}
	})

	t.Run("creates subscription with empty first name", func(t *testing.T) {
		emptyFirstName, _ := shared.NewFirstName("")

		params := subscription.NewSubscriptionParams{
			SubscriptionID: validSubscriptionID,
			FirstName:      emptyFirstName,
			Email:          validEmail,
			Clock:          clock,
		}

		got, err := subscription.NewSubscription(params)

		assertNoError(t, err)
		if got.FirstName.String() != "" {
			t.Errorf("expected empty first name, got %q", got.FirstName)
		}
	})

	t.Run("rejects invalid parameters", func(t *testing.T) {
		tests := []struct {
			name   string
			params subscription.NewSubscriptionParams
		}{
			{
				name: "empty subscription ID",
				params: subscription.NewSubscriptionParams{
					SubscriptionID: kernel.ID[subscription.Subscription](""),
					FirstName:      validFirstName,
					Email:          validEmail,
					Clock:          clock,
				},
			},
			{
				name: "invalid first name",
				params: subscription.NewSubscriptionParams{
					SubscriptionID: validSubscriptionID,
					FirstName:      shared.FirstName("a very long name that exceeds the maximum allowed length for first names"),
					Email:          validEmail,
					Clock:          clock,
				},
			},
			{
				name: "empty email",
				params: subscription.NewSubscriptionParams{
					SubscriptionID: validSubscriptionID,
					FirstName:      validFirstName,
					Email:          shared.Email(""),
					Clock:          clock,
				},
			},
			{
				name: "invalid email",
				params: subscription.NewSubscriptionParams{
					SubscriptionID: validSubscriptionID,
					FirstName:      validFirstName,
					Email:          shared.Email("invalid-email"),
					Clock:          clock,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := subscription.NewSubscription(tt.params)

				assertError(t, err)
				assertErrorCode(t, err, kernel.EInvalid)
			})
		}
	})
}

func TestSubscription_String(t *testing.T) {
	clock := &stubClock{t: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)}

	subscriptionID, _ := kernel.NewID[subscription.Subscription]("sub-123")
	firstName, _ := shared.NewFirstName("John")
	email, _ := shared.NewEmail("john@example.com")

	params := subscription.NewSubscriptionParams{
		SubscriptionID: subscriptionID,
		FirstName:      firstName,
		Email:          email,
		Clock:          clock,
	}

	sub, _ := subscription.NewSubscription(params)

	got := sub.String()
	want := `Subscription{ID: "sub-123", Name: "John", Email: "john@example.com", Status: "active", Active: true, SubscribedAt: 2024-01-15T10:00:00Z}`

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSubscription_Validate(t *testing.T) {
	clock := &stubClock{t: time.Now()}

	t.Run("valid subscription passes", func(t *testing.T) {
		subscriptionID, _ := kernel.NewID[subscription.Subscription]("sub-123")
		firstName, _ := shared.NewFirstName("John")
		email, _ := shared.NewEmail("john@example.com")

		sub, _ := subscription.NewSubscription(subscription.NewSubscriptionParams{
			SubscriptionID: subscriptionID,
			FirstName:      firstName,
			Email:          email,
			Clock:          clock,
		})

		err := sub.Validate()

		assertNoError(t, err)
	})

	t.Run("invalid fields fail", func(t *testing.T) {
		tests := []struct {
			name     string
			modifier func(*subscription.Subscription)
		}{
			{
				name: "empty ID",
				modifier: func(s *subscription.Subscription) {
					s.SubscriptionID = kernel.ID[subscription.Subscription]("")
				},
			},
			{
				name: "invalid first name",
				modifier: func(s *subscription.Subscription) {
					s.FirstName = shared.FirstName("a very long name that exceeds the maximum allowed length")
				},
			},
			{
				name: "empty email",
				modifier: func(s *subscription.Subscription) {
					s.Email = shared.Email("")
				},
			},
			{
				name: "invalid email",
				modifier: func(s *subscription.Subscription) {
					s.Email = shared.Email("invalid-email")
				},
			},
			{
				name: "invalid status",
				modifier: func(s *subscription.Subscription) {
					s.Status = subscription.Status("invalid")
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Create valid subscription
				subscriptionID, _ := kernel.NewID[subscription.Subscription]("sub-123")
				firstName, _ := shared.NewFirstName("John")
				email, _ := shared.NewEmail("john@example.com")

				sub, _ := subscription.NewSubscription(subscription.NewSubscriptionParams{
					SubscriptionID: subscriptionID,
					FirstName:      firstName,
					Email:          email,
					Clock:          clock,
				})

				// Apply modifier to make it invalid
				tt.modifier(&sub)

				err := sub.Validate()

				assertError(t, err)
				assertErrorCode(t, err, kernel.EInvalid)
			})
		}
	})
}

func TestSubscription_Unsubscribe(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	clock := &stubClock{t: fixedTime}

	createActiveSubscription := func() subscription.Subscription {
		subscriptionID, _ := kernel.NewID[subscription.Subscription]("sub-123")
		firstName, _ := shared.NewFirstName("John")
		email, _ := shared.NewEmail("john@example.com")

		sub, _ := subscription.NewSubscription(subscription.NewSubscriptionParams{
			SubscriptionID: subscriptionID,
			FirstName:      firstName,
			Email:          email,
			Clock:          clock,
		})

		return sub
	}

	t.Run("unsubscribe active subscription", func(t *testing.T) {
		sub := createActiveSubscription()
		unsubscribeTime := fixedTime.Add(24 * time.Hour)
		clock.t = unsubscribeTime

		got, err := sub.Unsubscribe()

		assertNoError(t, err)
		if got.Status != subscription.StatusUnsubscribed {
			t.Errorf("Status: got %v, want %v", got.Status, subscription.StatusUnsubscribed)
		}
		if got.IsActive {
			t.Error("expected IsActive to be false")
		}
		if got.UnsubscribedAt == nil || !got.UnsubscribedAt.Equal(unsubscribeTime) {
			t.Errorf("UnsubscribedAt: got %v, want %v", got.UnsubscribedAt, unsubscribeTime)
		}
		if !got.UpdatedAt.Equal(unsubscribeTime) {
			t.Errorf("UpdatedAt: got %v, want %v", got.UpdatedAt, unsubscribeTime)
		}
	})

	t.Run("cannot unsubscribe already unsubscribed", func(t *testing.T) {
		sub := createActiveSubscription()
		sub.Status = subscription.StatusUnsubscribed
		sub.IsActive = false

		_, err := sub.Unsubscribe()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EConflict)
	})

	t.Run("cannot unsubscribe bounced subscription", func(t *testing.T) {
		sub := createActiveSubscription()
		sub.Status = subscription.StatusBounced
		sub.IsActive = false

		_, err := sub.Unsubscribe()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EConflict)
	})

	t.Run("cannot unsubscribe complained subscription", func(t *testing.T) {
		sub := createActiveSubscription()
		sub.Status = subscription.StatusComplained
		sub.IsActive = false

		_, err := sub.Unsubscribe()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EConflict)
	})
}

func TestSubscription_Resubscribe(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	clock := &stubClock{t: fixedTime}

	createUnsubscribedSubscription := func() subscription.Subscription {
		subscriptionID, _ := kernel.NewID[subscription.Subscription]("sub-123")
		firstName, _ := shared.NewFirstName("John")
		email, _ := shared.NewEmail("john@example.com")

		sub, _ := subscription.NewSubscription(subscription.NewSubscriptionParams{
			SubscriptionID: subscriptionID,
			FirstName:      firstName,
			Email:          email,
			Clock:          clock,
		})

		// Unsubscribe it
		unsubscribeTime := fixedTime.Add(24 * time.Hour)
		sub.Status = subscription.StatusUnsubscribed
		sub.IsActive = false
		sub.UnsubscribedAt = &unsubscribeTime

		return sub
	}

	t.Run("resubscribe unsubscribed subscription", func(t *testing.T) {
		sub := createUnsubscribedSubscription()
		resubscribeTime := fixedTime.Add(48 * time.Hour)
		clock.t = resubscribeTime

		got, err := sub.Resubscribe()

		assertNoError(t, err)
		if got.Status != subscription.StatusActive {
			t.Errorf("Status: got %v, want %v", got.Status, subscription.StatusActive)
		}
		if !got.IsActive {
			t.Error("expected IsActive to be true")
		}
		if got.UnsubscribedAt != nil {
			t.Error("expected UnsubscribedAt to be nil")
		}
		if !got.UpdatedAt.Equal(resubscribeTime) {
			t.Errorf("UpdatedAt: got %v, want %v", got.UpdatedAt, resubscribeTime)
		}
	})

	t.Run("cannot resubscribe active subscription", func(t *testing.T) {
		subscriptionID, _ := kernel.NewID[subscription.Subscription]("sub-123")
		firstName, _ := shared.NewFirstName("John")
		email, _ := shared.NewEmail("john@example.com")

		sub, _ := subscription.NewSubscription(subscription.NewSubscriptionParams{
			SubscriptionID: subscriptionID,
			FirstName:      firstName,
			Email:          email,
			Clock:          clock,
		})

		_, err := sub.Resubscribe()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EConflict)
	})

	t.Run("cannot resubscribe bounced subscription", func(t *testing.T) {
		sub := createUnsubscribedSubscription()
		sub.Status = subscription.StatusBounced

		_, err := sub.Resubscribe()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("cannot resubscribe complained subscription", func(t *testing.T) {
		sub := createUnsubscribedSubscription()
		sub.Status = subscription.StatusComplained

		_, err := sub.Resubscribe()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})
}

func TestSubscription_MarkAsBounced(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	clock := &stubClock{t: fixedTime}

	createActiveSubscription := func() subscription.Subscription {
		subscriptionID, _ := kernel.NewID[subscription.Subscription]("sub-123")
		firstName, _ := shared.NewFirstName("John")
		email, _ := shared.NewEmail("john@example.com")

		sub, _ := subscription.NewSubscription(subscription.NewSubscriptionParams{
			SubscriptionID: subscriptionID,
			FirstName:      firstName,
			Email:          email,
			Clock:          clock,
		})

		return sub
	}

	t.Run("mark active subscription as bounced", func(t *testing.T) {
		sub := createActiveSubscription()
		bounceTime := fixedTime.Add(24 * time.Hour)
		clock.t = bounceTime

		got, err := sub.MarkAsBounced()

		assertNoError(t, err)
		if got.Status != subscription.StatusBounced {
			t.Errorf("Status: got %v, want %v", got.Status, subscription.StatusBounced)
		}
		if got.IsActive {
			t.Error("expected IsActive to be false")
		}
		if !got.UpdatedAt.Equal(bounceTime) {
			t.Errorf("UpdatedAt: got %v, want %v", got.UpdatedAt, bounceTime)
		}
	})

	t.Run("can mark any status as bounced", func(t *testing.T) {
		statuses := []subscription.Status{
			subscription.StatusActive,
			subscription.StatusUnsubscribed,
			subscription.StatusComplained,
			subscription.StatusBounced, // Can re-bounce
		}

		for _, status := range statuses {
			t.Run(string(status), func(t *testing.T) {
				sub := createActiveSubscription()
				sub.Status = status

				got, err := sub.MarkAsBounced()

				assertNoError(t, err)
				if got.Status != subscription.StatusBounced {
					t.Errorf("Status: got %v, want %v", got.Status, subscription.StatusBounced)
				}
			})
		}
	})
}

func TestSubscription_MarkAsComplained(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	clock := &stubClock{t: fixedTime}

	createActiveSubscription := func() subscription.Subscription {
		subscriptionID, _ := kernel.NewID[subscription.Subscription]("sub-123")
		firstName, _ := shared.NewFirstName("John")
		email, _ := shared.NewEmail("john@example.com")

		sub, _ := subscription.NewSubscription(subscription.NewSubscriptionParams{
			SubscriptionID: subscriptionID,
			FirstName:      firstName,
			Email:          email,
			Clock:          clock,
		})

		return sub
	}

	t.Run("mark active subscription as complained", func(t *testing.T) {
		sub := createActiveSubscription()
		complaintTime := fixedTime.Add(24 * time.Hour)
		clock.t = complaintTime

		got, err := sub.MarkAsComplained()

		assertNoError(t, err)
		if got.Status != subscription.StatusComplained {
			t.Errorf("Status: got %v, want %v", got.Status, subscription.StatusComplained)
		}
		if got.IsActive {
			t.Error("expected IsActive to be false")
		}
		if !got.UpdatedAt.Equal(complaintTime) {
			t.Errorf("UpdatedAt: got %v, want %v", got.UpdatedAt, complaintTime)
		}
	})
}

func TestSubscription_IsSubscribed(t *testing.T) {
	clock := &stubClock{t: time.Now()}

	subscriptionID, _ := kernel.NewID[subscription.Subscription]("sub-123")
	firstName, _ := shared.NewFirstName("John")
	email, _ := shared.NewEmail("john@example.com")

	tests := []struct {
		name     string
		status   subscription.Status
		isActive bool
		want     bool
	}{
		{"active and IsActive", subscription.StatusActive, true, true},
		{"active but not IsActive", subscription.StatusActive, false, false},
		{"unsubscribed", subscription.StatusUnsubscribed, false, false},
		{"bounced", subscription.StatusBounced, false, false},
		{"complained", subscription.StatusComplained, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub, _ := subscription.NewSubscription(subscription.NewSubscriptionParams{
				SubscriptionID: subscriptionID,
				FirstName:      firstName,
				Email:          email,
				Clock:          clock,
			})

			sub.Status = tt.status
			sub.IsActive = tt.isActive

			got := sub.IsSubscribed()

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubscription_CanReceiveEmails(t *testing.T) {
	// CanReceiveEmails has same logic as IsSubscribed
	clock := &stubClock{t: time.Now()}

	subscriptionID, _ := kernel.NewID[subscription.Subscription]("sub-123")
	firstName, _ := shared.NewFirstName("John")
	email, _ := shared.NewEmail("john@example.com")

	sub, _ := subscription.NewSubscription(subscription.NewSubscriptionParams{
		SubscriptionID: subscriptionID,
		FirstName:      firstName,
		Email:          email,
		Clock:          clock,
	})

	t.Run("active subscription can receive emails", func(t *testing.T) {
		if !sub.CanReceiveEmails() {
			t.Error("expected CanReceiveEmails to be true")
		}
	})

	t.Run("unsubscribed cannot receive emails", func(t *testing.T) {
		sub.Status = subscription.StatusUnsubscribed
		sub.IsActive = false

		if sub.CanReceiveEmails() {
			t.Error("expected CanReceiveEmails to be false")
		}
	})
}

func TestSubscription_GetDisplayName(t *testing.T) {
	clock := &stubClock{t: time.Now()}

	subscriptionID, _ := kernel.NewID[subscription.Subscription]("sub-123")
	email, _ := shared.NewEmail("john@example.com")

	t.Run("returns first name when available", func(t *testing.T) {
		firstName, _ := shared.NewFirstName("John")

		sub, _ := subscription.NewSubscription(subscription.NewSubscriptionParams{
			SubscriptionID: subscriptionID,
			FirstName:      firstName,
			Email:          email,
			Clock:          clock,
		})

		got := sub.GetDisplayName()
		want := "John"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("returns email when first name is empty", func(t *testing.T) {
		emptyFirstName, _ := shared.NewFirstName("")

		sub, _ := subscription.NewSubscription(subscription.NewSubscriptionParams{
			SubscriptionID: subscriptionID,
			FirstName:      emptyFirstName,
			Email:          email,
			Clock:          clock,
		})

		got := sub.GetDisplayName()
		want := "john@example.com"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
