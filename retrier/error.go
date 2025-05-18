package retrier

import "fmt"

// RetrierError is returned when an operation fails after exhausting all retry attempts.
type RetrierError struct {
	attempts int   // total attempts made (≥ maxAttempts)
	wrapped  error // last error encountered
}

func (e *RetrierError) Error() string {
	return fmt.Sprintf("retrier: after %d attempt(s): %v", e.attempts, e.wrapped)
}

func (e *RetrierError) Unwrap() error { return e.wrapped }
