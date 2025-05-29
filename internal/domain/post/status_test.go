package post_test

import (
	"testing"

	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/post"
)

func TestStatus_String(t *testing.T) {
	tests := []struct {
		status post.Status
		want   string
	}{
		{post.StatusDraft, "draft"},
		{post.StatusPublished, "published"},
		{post.StatusArchived, "archived"},
		{post.StatusScheduled, "scheduled"},
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
		validStatuses := []post.Status{
			post.StatusDraft,
			post.StatusPublished,
			post.StatusArchived,
			post.StatusScheduled,
		}

		for _, status := range validStatuses {
			t.Run(string(status), func(t *testing.T) {
				err := status.Validate()

				assertNoError(t, err)
			})
		}
	})

	t.Run("invalid status fails", func(t *testing.T) {
		invalidStatuses := []post.Status{
			"",
			"pending",
			"review",
			"deleted",
			"DRAFT", // case sensitive
			"Published",
			"draft ",
			" draft",
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

func TestStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name  string
		from  post.Status
		to    post.Status
		canDo bool
	}{
		// From Draft
		{"draft to published", post.StatusDraft, post.StatusPublished, true},
		{"draft to scheduled", post.StatusDraft, post.StatusScheduled, true},
		{"draft to archived", post.StatusDraft, post.StatusArchived, false},
		{"draft to draft", post.StatusDraft, post.StatusDraft, true}, // same status always allowed

		// From Published
		{"published to draft", post.StatusPublished, post.StatusDraft, true},
		{"published to archived", post.StatusPublished, post.StatusArchived, true},
		{"published to scheduled", post.StatusPublished, post.StatusScheduled, false},
		{"published to published", post.StatusPublished, post.StatusPublished, true},

		// From Scheduled
		{"scheduled to draft", post.StatusScheduled, post.StatusDraft, true},
		{"scheduled to published", post.StatusScheduled, post.StatusPublished, true},
		{"scheduled to archived", post.StatusScheduled, post.StatusArchived, false},
		{"scheduled to scheduled", post.StatusScheduled, post.StatusScheduled, true},

		// From Archived
		{"archived to published", post.StatusArchived, post.StatusPublished, true},
		{"archived to draft", post.StatusArchived, post.StatusDraft, false},
		{"archived to scheduled", post.StatusArchived, post.StatusScheduled, false},
		{"archived to archived", post.StatusArchived, post.StatusArchived, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.from.CanTransitionTo(tt.to)

			if got != tt.canDo {
				t.Errorf("got %v, want %v", got, tt.canDo)
			}
		})
	}
}

func TestStatusConstants(t *testing.T) {
	// Ensure status constants have expected values
	tests := []struct {
		name   string
		status post.Status
		want   string
	}{
		{"StatusDraft", post.StatusDraft, "draft"},
		{"StatusPublished", post.StatusPublished, "published"},
		{"StatusArchived", post.StatusArchived, "archived"},
		{"StatusScheduled", post.StatusScheduled, "scheduled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("got %q, want %q", tt.status, tt.want)
			}
		})
	}
}

func TestStatus_InvalidTransitions(t *testing.T) {
	// Test with invalid status values to ensure they return false
	invalidStatus := post.Status("invalid")
	validStatus := post.StatusDraft

	t.Run("invalid status cannot transition", func(t *testing.T) {
		got := invalidStatus.CanTransitionTo(validStatus)

		if got {
			t.Error("expected invalid status to not allow transitions")
		}
	})

	t.Run("cannot transition to invalid status", func(t *testing.T) {
		got := validStatus.CanTransitionTo(invalidStatus)

		if got {
			t.Error("expected transition to invalid status to be disallowed")
		}
	})
}
