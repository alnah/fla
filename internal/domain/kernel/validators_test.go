package kernel_test

import (
	"strings"
	"testing"

	"github.com/alnah/fla/internal/domain/kernel"
)

func TestErrLen(t *testing.T) {
	got := kernel.ErrLen("username", 3, 20)
	want := "username must be between 3 and 20 characters."

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestErrGt(t *testing.T) {
	got := kernel.ErrGt("password", 8)
	want := "password must be greater than 8 characters."

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestErrLt(t *testing.T) {
	got := kernel.ErrLt("title", 100)
	want := "title must be less than 100 characters."

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestErrMissing(t *testing.T) {
	got := kernel.ErrMissing("email")
	want := "Missing email."

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestValidatePresence(t *testing.T) {
	const operation = "TestOp"

	t.Run("accepts non-empty string", func(t *testing.T) {
		err := kernel.ValidatePresence("field", "value", operation)

		assertNoError(t, err)
	})

	t.Run("rejects empty string", func(t *testing.T) {
		err := kernel.ValidatePresence("field", "", operation)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("rejects whitespace only", func(t *testing.T) {
		err := kernel.ValidatePresence("field", "   ", operation)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("accepts string with whitespace and content", func(t *testing.T) {
		err := kernel.ValidatePresence("field", "  content  ", operation)

		assertNoError(t, err)
	})
}

func TestValidateLength(t *testing.T) {
	const operation = "TestOp"

	t.Run("accepts string within bounds", func(t *testing.T) {
		err := kernel.ValidateLength("field", "hello", 3, 10, operation)

		assertNoError(t, err)
	})

	t.Run("accepts string at minimum length", func(t *testing.T) {
		err := kernel.ValidateLength("field", "abc", 3, 10, operation)

		assertNoError(t, err)
	})

	t.Run("accepts string at maximum length", func(t *testing.T) {
		err := kernel.ValidateLength("field", "1234567890", 3, 10, operation)

		assertNoError(t, err)
	})

	t.Run("rejects string below minimum", func(t *testing.T) {
		err := kernel.ValidateLength("field", "ab", 3, 10, operation)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("rejects string above maximum", func(t *testing.T) {
		err := kernel.ValidateLength("field", "12345678901", 3, 10, operation)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("counts unicode characters correctly", func(t *testing.T) {
		// "café" has 4 unicode characters
		err := kernel.ValidateLength("field", "café", 4, 4, operation)

		assertNoError(t, err)
	})

	t.Run("handles emojis correctly", func(t *testing.T) {
		// Each emoji counts as one character
		err := kernel.ValidateLength("field", "👋🌍", 2, 2, operation)

		assertNoError(t, err)
	})
}

func TestValidateMinLength(t *testing.T) {
	const operation = "TestOp"

	t.Run("accepts string above minimum", func(t *testing.T) {
		err := kernel.ValidateMinLength("field", "hello world", 5, operation)

		assertNoError(t, err)
	})

	t.Run("accepts string at minimum", func(t *testing.T) {
		err := kernel.ValidateMinLength("field", "hello", 5, operation)

		assertNoError(t, err)
	})

	t.Run("rejects string below minimum", func(t *testing.T) {
		err := kernel.ValidateMinLength("field", "hi", 5, operation)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("accepts empty string when minimum is 0", func(t *testing.T) {
		err := kernel.ValidateMinLength("field", "", 0, operation)

		assertNoError(t, err)
	})

	t.Run("handles unicode correctly", func(t *testing.T) {
		err := kernel.ValidateMinLength("field", "café☕", 5, operation)

		assertNoError(t, err)
	})
}

func TestValidateMaxLength(t *testing.T) {
	const operation = "TestOp"

	t.Run("accepts string below maximum", func(t *testing.T) {
		err := kernel.ValidateMaxLength("field", "hello", 10, operation)

		assertNoError(t, err)
	})

	t.Run("accepts string at maximum", func(t *testing.T) {
		err := kernel.ValidateMaxLength("field", "1234567890", 10, operation)

		assertNoError(t, err)
	})

	t.Run("rejects string above maximum", func(t *testing.T) {
		err := kernel.ValidateMaxLength("field", "12345678901", 10, operation)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("accepts empty string", func(t *testing.T) {
		err := kernel.ValidateMaxLength("field", "", 10, operation)

		assertNoError(t, err)
	})

	t.Run("handles very long strings", func(t *testing.T) {
		longString := strings.Repeat("a", 1000)
		err := kernel.ValidateMaxLength("field", longString, 100, operation)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("handles unicode correctly", func(t *testing.T) {
		// 10 unicode characters including emojis
		err := kernel.ValidateMaxLength("field", "Hello 世界 🌍", 10, operation)

		assertNoError(t, err)
	})
}

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
