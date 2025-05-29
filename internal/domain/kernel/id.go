package kernel

import (
	"fmt"
	"strings"
)

// ID provides unique identification for entities in the system.
// Generic type parameter T indicates the entity type this ID belongs to.
type ID[T any] string

// NewID creates a validated identifier with presence checking.
// Prevents empty or invalid identifiers that would break entity references.
func NewID[T any](id string) (ID[T], error) {
	const op = "NewID"

	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		var zero T
		return "", &Error{
			Code:      EInvalid,
			Message:   fmt.Sprintf("Missing %T ID.", zero),
			Operation: op,
		}
	}

	return ID[T](trimmed), nil
}

// String returns the string representation of the ID.
func (id ID[T]) String() string { return string(id) }

// Validate ensures ID meets system requirements for entity identification.
// Prevents database constraint violations and broken entity relationships.
func (id ID[T]) Validate() error {
	const op = "ID.Validate"

	if strings.TrimSpace(string(id)) == "" {
		var zero T
		return &Error{
			Code:      EInvalid,
			Message:   fmt.Sprintf("Missing %T ID.", zero),
			Operation: op,
		}
	}

	return nil
}
