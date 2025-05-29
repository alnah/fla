package user_test

import (
	"testing"

	"github.com/alnah/fla/internal/domain/kernel"
)

// Test helpers
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func assertErrorCode(t *testing.T, err error, want string) {
	t.Helper()
	got := kernel.ErrorCode(err)
	if got != want {
		t.Errorf("error code: got %q, want %q", got, want)
	}
}

func assertErrorMessage(t *testing.T, err error, want string) {
	t.Helper()
	got := kernel.ErrorMessage(err)
	if got != want {
		t.Errorf("error message: got %q, want %q", got, want)
	}
}
