package shared_test

import (
	"strings"
	"testing"

	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/shared"
)

func TestNewFirstName(t *testing.T) {
	t.Run("creates first name with valid input", func(t *testing.T) {
		names := []string{
			"John",
			"Marie-Claire",
			"José",
			"李",
			"Владимир",
			"", // empty is valid (optional field)
		}

		for _, name := range names {
			t.Run(name, func(t *testing.T) {
				got, err := shared.NewFirstName(name)

				assertNoError(t, err)
				if got.String() != name {
					t.Errorf("got %q, want %q", got, name)
				}
			})
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		input := "  John  "
		want := "John"

		got, err := shared.NewFirstName(input)

		assertNoError(t, err)
		if got.String() != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("allows empty name", func(t *testing.T) {
		got, err := shared.NewFirstName("")

		assertNoError(t, err)
		if got.String() != "" {
			t.Errorf("got %q, want empty", got)
		}
	})

	t.Run("rejects name exceeding max length", func(t *testing.T) {
		longName := strings.Repeat("a", shared.MaxFirstNameLength+1)

		_, err := shared.NewFirstName(longName)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("accepts name at max length", func(t *testing.T) {
		maxName := strings.Repeat("a", shared.MaxFirstNameLength)

		got, err := shared.NewFirstName(maxName)

		assertNoError(t, err)
		if len(got.String()) != shared.MaxFirstNameLength {
			t.Errorf("got length %d, want %d", len(got.String()), shared.MaxFirstNameLength)
		}
	})
}

func TestFirstName_Validate(t *testing.T) {
	t.Run("valid name passes", func(t *testing.T) {
		name := shared.FirstName("John")

		err := name.Validate()

		assertNoError(t, err)
	})

	t.Run("empty name passes", func(t *testing.T) {
		name := shared.FirstName("")

		err := name.Validate()

		assertNoError(t, err)
	})

	t.Run("too long name fails", func(t *testing.T) {
		name := shared.FirstName(strings.Repeat("a", shared.MaxFirstNameLength+1))

		err := name.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})
}

func TestNewLastName(t *testing.T) {
	t.Run("creates last name with valid input", func(t *testing.T) {
		names := []string{
			"Smith",
			"van der Berg",
			"O'Connor",
			"García-López",
			"王",
			"", // empty is valid (optional field)
		}

		for _, name := range names {
			t.Run(name, func(t *testing.T) {
				got, err := shared.NewLastName(name)

				assertNoError(t, err)
				if got.String() != name {
					t.Errorf("got %q, want %q", got, name)
				}
			})
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		input := "  Smith  "
		want := "Smith"

		got, err := shared.NewLastName(input)

		assertNoError(t, err)
		if got.String() != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("rejects name exceeding max length", func(t *testing.T) {
		longName := strings.Repeat("a", shared.MaxLastNameLength+1)

		_, err := shared.NewLastName(longName)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})
}

func TestLastName_Validate(t *testing.T) {
	t.Run("valid name passes", func(t *testing.T) {
		name := shared.LastName("Smith")

		err := name.Validate()

		assertNoError(t, err)
	})

	t.Run("empty name passes", func(t *testing.T) {
		name := shared.LastName("")

		err := name.Validate()

		assertNoError(t, err)
	})

	t.Run("too long name fails", func(t *testing.T) {
		name := shared.LastName(strings.Repeat("a", shared.MaxLastNameLength+1))

		err := name.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})
}

func TestNewUsername(t *testing.T) {
	t.Run("creates username with valid input", func(t *testing.T) {
		usernames := []string{
			"john123",
			"user_name",
			"test-user",
			"ABC",
			"123",
			"a_b-c",
			"UPPER_lower_123",
		}

		for _, username := range usernames {
			t.Run(username, func(t *testing.T) {
				got, err := shared.NewUsername(username)

				assertNoError(t, err)
				if got.String() != username {
					t.Errorf("got %q, want %q", got, username)
				}
			})
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		input := "  john123  "
		want := "john123"

		got, err := shared.NewUsername(input)

		assertNoError(t, err)
		if got.String() != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("rejects empty username", func(t *testing.T) {
		_, err := shared.NewUsername("")

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("rejects whitespace only", func(t *testing.T) {
		_, err := shared.NewUsername("   ")

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("rejects username below min length", func(t *testing.T) {
		_, err := shared.NewUsername("ab")

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("accepts username at min length", func(t *testing.T) {
		got, err := shared.NewUsername("abc")

		assertNoError(t, err)
		if got.String() != "abc" {
			t.Errorf("got %q, want %q", got, "abc")
		}
	})

	t.Run("rejects username exceeding max length", func(t *testing.T) {
		longUsername := strings.Repeat("a", shared.MaxUsernameLength+1)

		_, err := shared.NewUsername(longUsername)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("accepts username at max length", func(t *testing.T) {
		maxUsername := strings.Repeat("a", shared.MaxUsernameLength)

		got, err := shared.NewUsername(maxUsername)

		assertNoError(t, err)
		if len(got.String()) != shared.MaxUsernameLength {
			t.Errorf("got length %d, want %d", len(got.String()), shared.MaxUsernameLength)
		}
	})

	t.Run("rejects invalid characters", func(t *testing.T) {
		invalidUsernames := []string{
			"user name",  // space
			"user@name",  // @
			"user.name",  // dot
			"user!name",  // exclamation
			"user#name",  // hash
			"user$name",  // dollar
			"user%name",  // percent
			"user&name",  // ampersand
			"user*name",  // asterisk
			"user(name",  // parenthesis
			"user+name",  // plus
			"user=name",  // equals
			"user/name",  // slash
			"user\\name", // backslash
			"user[name",  // bracket
			"user{name",  // brace
			"user|name",  // pipe
			"user~name",  // tilde
			"user`name",  // backtick
			"user'name",  // apostrophe
			"user\"name", // quote
			"user,name",  // comma
			"user<name",  // less than
			"user>name",  // greater than
			"user?name",  // question mark
			"user:name",  // colon
			"user;name",  // semicolon
			"user\nname", // newline
			"user\tname", // tab
			"用户名",        // non-ASCII
		}

		for _, username := range invalidUsernames {
			t.Run(username, func(t *testing.T) {
				_, err := shared.NewUsername(username)

				assertError(t, err)
				assertErrorCode(t, err, kernel.EInvalid)
			})
		}
	})
}

func TestUsername_Validate(t *testing.T) {
	t.Run("valid username passes", func(t *testing.T) {
		username := shared.Username("john123")

		err := username.Validate()

		assertNoError(t, err)
	})

	t.Run("empty username fails", func(t *testing.T) {
		username := shared.Username("")

		err := username.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("invalid characters fail", func(t *testing.T) {
		username := shared.Username("user@name")

		err := username.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("too short username fails", func(t *testing.T) {
		username := shared.Username("ab")

		err := username.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("too long username fails", func(t *testing.T) {
		username := shared.Username(strings.Repeat("a", shared.MaxUsernameLength+1))

		err := username.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})
}
