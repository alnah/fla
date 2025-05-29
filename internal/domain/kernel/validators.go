package kernel

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// ErrLen generates consistent length validation error messages.
// Reduces repetitive error message formatting across value objects.
func ErrLen(field string, minLen, maxLen int) string {
	return fmt.Sprintf("%s must be between %d and %d characters.", field, minLen, maxLen)
}

// ErrGt generates greater-than length validation error messages.
// Provides consistent minimum length error formatting.
func ErrGt(field string, minLen int) string {
	return fmt.Sprintf("%s must be greater than %d characters.", field, minLen)
}

// ErrLt generates less-than length validation error messages.
// Provides consistent maximum length error formatting.
func ErrLt(field string, maxLen int) string {
	return fmt.Sprintf("%s must be less than %d characters.", field, maxLen)
}

// ErrMissing generates missing field error messages.
// Standardizes presence validation errors.
func ErrMissing(field string) string {
	return fmt.Sprintf("Missing %s.", field)
}

// ValidatePresence ensures a field is not empty.
// Common validation for required string fields.
func ValidatePresence(field, value, operation string) error {
	if strings.TrimSpace(value) == "" {
		return &Error{
			Code:      EInvalid,
			Message:   ErrMissing(field),
			Operation: operation,
		}
	}
	return nil
}

// ValidateLength ensures a string is within min/max bounds.
// Common validation for length-constrained fields.
func ValidateLength(field, value string, minLen, maxLen int, operation string) error {
	length := utf8.RuneCountInString(value)
	if length < minLen || length > maxLen {
		return &Error{
			Code:      EInvalid,
			Message:   ErrLen(field, minLen, maxLen),
			Operation: operation,
		}
	}
	return nil
}

// ValidateMinLength ensures a string meets minimum length.
// Common validation for fields with only minimum constraints.
func ValidateMinLength(field, value string, minLen int, operation string) error {
	if utf8.RuneCountInString(value) < minLen {
		return &Error{
			Code:      EInvalid,
			Message:   ErrGt(field, minLen),
			Operation: operation,
		}
	}
	return nil
}

// ValidateMaxLength ensures a string doesn't exceed maximum length.
// Common validation for fields with only maximum constraints.
func ValidateMaxLength(field, value string, maxLen int, operation string) error {
	if utf8.RuneCountInString(value) > maxLen {
		return &Error{
			Code:      EInvalid,
			Message:   ErrLt(field, maxLen),
			Operation: operation,
		}
	}
	return nil
}
