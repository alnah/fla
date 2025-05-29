package post

import (
	"slices"

	"github.com/alnah/fla/internal/domain/kernel"
)

const (
	MPostContentInvalid string = "Invalid post content."
	MStatusInvalid      string = "Invalid status."
	MSchemaTypeInvalid  string = "Invalid schema type."
)

// Status represents the publication state of blog content in the editorial workflow.
// Controls content visibility and enables staged publication processes.
type Status string

const (
	StatusDraft     Status = "draft"     // Content in development, not visible to public
	StatusPublished Status = "published" // Live content available to all readers
	StatusArchived  Status = "archived"  // Historical content removed from active circulation
	StatusScheduled Status = "scheduled" // Content queued for future publication
)

// allowedTransitions defines valid status transitions in the workflow.
// Enforces editorial process and prevents invalid state changes.
var allowedTransitions = map[Status][]Status{
	StatusDraft:     {StatusPublished, StatusScheduled},
	StatusPublished: {StatusDraft, StatusArchived},
	StatusScheduled: {StatusDraft, StatusPublished},
	StatusArchived:  {StatusPublished},
}

func (s Status) String() string { return string(s) }

// Validate ensures status uses defined workflow states.
// Prevents invalid status assignments that would break publication flow.
func (s Status) Validate() error {
	const op = "Status.Validate"

	switch s {
	case StatusDraft, StatusPublished, StatusArchived, StatusScheduled:
		return nil
	default:
		return &kernel.Error{
			Code:      kernel.EInvalid,
			Message:   MStatusInvalid,
			Operation: op,
		}
	}
}

// CanTransitionTo checks if this status can transition to the target status.
// Uses the transition table to enforce workflow rules.
func (s Status) CanTransitionTo(target Status) bool {
	// Same status is always allowed
	if s == target {
		return true
	}

	// Check allowed transitions
	allowed, exists := allowedTransitions[s]
	if !exists {
		return false
	}

	return slices.Contains(allowed, target)
}
