package kernel

import (
	"bytes"
	"fmt"
)

// Application error codes for categorizing different types of failures.
const (
	EConflict  string = "conflict"  // Action cannot be performed due to business rule conflicts
	EInternal  string = "internal"  // Internal system error requiring technical investigation
	EInvalid   string = "invalid"   // Validation failed on user input or data constraints
	EForbidden string = "forbidden" // Action not allowed due to permission restrictions
	ENotFound  string = "not_found" // Requested entity does not exist in the system
)

// Application generic message for internal errors to avoid exposing system details.
const MInternal string = "An internal error has occurred. Please contact technical support."

// Error provides structured error handling with operation context and error chaining.
// Enables precise error diagnosis and consistent error responses across the domain.
type Error struct {
	// Machine-readable error code for programmatic handling
	Code string

	// Human-readable message for user display
	Message string

	// Logical operation context for debugging and tracing
	Operation string

	// Underlying error cause for error chain traversal
	Cause error
}

// ErrorCode extracts the machine-readable error classification for handling logic.
// Returns the most specific error code available in the error chain.
func ErrorCode(err error) string {
	if err == nil {
		return ""
	} else if e, ok := err.(*Error); ok && e.Code != "" {
		return e.Code
	} else if ok && e.Cause != nil {
		return ErrorCode(e.Cause)
	}

	return EInternal
}

// ErrorMessage retrieves the human-readable error description for user display.
// Provides clear, actionable error messages while maintaining security.
func ErrorMessage(err error) string {
	if err == nil {
		return ""
	} else if e, ok := err.(*Error); ok && e.Message != "" {
		return e.Message
	} else if ok && e.Cause != nil {
		return ErrorMessage(e.Cause)
	}

	return MInternal
}

// Error returns the complete error representation including operation context.
// Provides detailed error information for logging and debugging purposes.
func (e *Error) Error() string {
	var buf bytes.Buffer

	// Include operation context for tracing error location
	if e.Operation != "" {
		fmt.Fprintf(&buf, "%s: ", e.Operation)
	}

	// Chain error messages or provide code and message
	if e.Cause != nil {
		buf.WriteString(e.Cause.Error())
	} else {
		if e.Code != "" {
			fmt.Fprintf(&buf, "<%s> ", e.Code)
		}
		buf.WriteString(e.Message)
	}

	return buf.String()
}
