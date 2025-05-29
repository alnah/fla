package shared_test

import (
	"testing"

	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/shared"
)

func TestNewEmail(t *testing.T) {
	t.Run("creates email with valid format", func(t *testing.T) {
		validEmails := []string{
			"user@example.com",
			"user.name@example.com",
			"user+tag@example.co.uk",
			"user_name@example-domain.com",
			"123@example.com",
			"user@123.456.789.012",
			"user%name@example.com",
		}

		for _, email := range validEmails {
			t.Run(email, func(t *testing.T) {
				got, err := shared.NewEmail(email)

				assertNoError(t, err)
				if got.String() != email {
					t.Errorf("got %q, want %q", got, email)
				}
			})
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		input := "  user@example.com  "
		want := "user@example.com"

		got, err := shared.NewEmail(input)

		assertNoError(t, err)
		if got.String() != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("rejects empty email", func(t *testing.T) {
		_, err := shared.NewEmail("")

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("rejects whitespace only", func(t *testing.T) {
		_, err := shared.NewEmail("   ")

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("rejects invalid email formats", func(t *testing.T) {
		invalidEmails := []string{
			"user",                   // no @ symbol
			"@example.com",           // no local part
			"user@",                  // no domain
			"user@.com",              // domain starts with dot
			"user@example",           // no TLD
			"user @example.com",      // space in local part
			"user@exam ple.com",      // space in domain
			"user..name@example.com", // consecutive dots
			"user@example..com",      // consecutive dots in domain
			"user@example.com.",      // trailing dot
			".user@example.com",      // leading dot
			"user@.example.com",      // leading dot in domain
			"user@@example.com",      // double @
			"user@exam@ple.com",      // multiple @
		}

		for _, email := range invalidEmails {
			t.Run(email, func(t *testing.T) {
				_, err := shared.NewEmail(email)

				assertError(t, err)
				assertErrorCode(t, err, kernel.EInvalid)
			})
		}
	})
}

func TestEmail_String(t *testing.T) {
	want := "user@example.com"
	email := shared.Email(want)

	got := email.String()

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEmail_Validate(t *testing.T) {
	t.Run("valid email passes", func(t *testing.T) {
		email := shared.Email("user@example.com")

		err := email.Validate()

		assertNoError(t, err)
	})

	t.Run("empty email fails", func(t *testing.T) {
		email := shared.Email("")

		err := email.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("invalid format fails", func(t *testing.T) {
		email := shared.Email("invalid-email")

		err := email.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})
}

func TestEmail_EdgeCases(t *testing.T) {
	t.Run("handles international domains", func(t *testing.T) {
		// Currently the regex doesn't support IDN, this should fail
		_, err := shared.NewEmail("user@例え.jp")

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("handles very long valid email", func(t *testing.T) {
		// Max 64 chars for local part
		longLocal := "abcdefghijklmnopqrstuvwxyz0123456789.abcdefghijklmnopqrstuvwxy"
		email := longLocal + "@example.com"

		got, err := shared.NewEmail(email)

		assertNoError(t, err)
		if got.String() != email {
			t.Errorf("got %q, want %q", got, email)
		}
	})
}
