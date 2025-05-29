package kernel_test

import (
	"errors"
	"testing"

	"github.com/alnah/fla/internal/domain/kernel"
)

func TestError_Error(t *testing.T) {
	t.Run("with operation and message", func(t *testing.T) {
		err := &kernel.Error{
			Operation: "CreateUser",
			Code:      kernel.EInvalid,
			Message:   "invalid email",
		}

		want := "CreateUser: <invalid> invalid email"
		got := err.Error()

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("with operation and cause", func(t *testing.T) {
		cause := errors.New("database connection failed")
		err := &kernel.Error{
			Operation: "SaveUser",
			Cause:     cause,
		}

		want := "SaveUser: database connection failed"
		got := err.Error()

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("with message only", func(t *testing.T) {
		err := &kernel.Error{
			Code:    kernel.ENotFound,
			Message: "user not found",
		}

		want := "<not_found> user not found"
		got := err.Error()

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("with operation, code, message and cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := &kernel.Error{
			Operation: "UpdateUser",
			Code:      kernel.EConflict,
			Message:   "email already exists",
			Cause:     cause,
		}

		want := "UpdateUser: underlying error"
		got := err.Error()

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestErrorCode(t *testing.T) {
	t.Run("returns code from error", func(t *testing.T) {
		err := &kernel.Error{Code: kernel.EInvalid}

		got := kernel.ErrorCode(err)
		want := kernel.EInvalid

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("returns code from nested error", func(t *testing.T) {
		innerErr := &kernel.Error{Code: kernel.ENotFound}
		err := &kernel.Error{Cause: innerErr}

		got := kernel.ErrorCode(err)
		want := kernel.ENotFound

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("returns internal for non-kernel error", func(t *testing.T) {
		err := errors.New("standard error")

		got := kernel.ErrorCode(err)
		want := kernel.EInternal

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("returns empty string for nil error", func(t *testing.T) {
		got := kernel.ErrorCode(nil)
		want := ""

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("returns internal when no code in chain", func(t *testing.T) {
		err := &kernel.Error{Cause: errors.New("standard error")}

		got := kernel.ErrorCode(err)
		want := kernel.EInternal

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestErrorMessage(t *testing.T) {
	t.Run("returns message from error", func(t *testing.T) {
		err := &kernel.Error{Message: "validation failed"}

		got := kernel.ErrorMessage(err)
		want := "validation failed"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("returns message from nested error", func(t *testing.T) {
		innerErr := &kernel.Error{Message: "inner message"}
		err := &kernel.Error{Cause: innerErr}

		got := kernel.ErrorMessage(err)
		want := "inner message"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("returns internal message for non-kernel error", func(t *testing.T) {
		err := errors.New("standard error")

		got := kernel.ErrorMessage(err)
		want := kernel.MInternal

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("returns empty string for nil error", func(t *testing.T) {
		got := kernel.ErrorMessage(nil)
		want := ""

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("returns internal message when no message in chain", func(t *testing.T) {
		err := &kernel.Error{Cause: errors.New("standard error")}

		got := kernel.ErrorMessage(err)
		want := kernel.MInternal

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestErrorConstants(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"conflict code", kernel.EConflict, "conflict"},
		{"internal code", kernel.EInternal, "internal"},
		{"invalid code", kernel.EInvalid, "invalid"},
		{"forbidden code", kernel.EForbidden, "forbidden"},
		{"not found code", kernel.ENotFound, "not_found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.expected {
				t.Errorf("got %q, want %q", tt.code, tt.expected)
			}
		})
	}
}
