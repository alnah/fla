package kernel_test

import (
	"testing"

	"github.com/alnah/fla/internal/domain/kernel"
)

// TestResource is a dummy type for testing generic URLs
type TestResource struct{}

func TestNewURL(t *testing.T) {
	t.Run("creates URL with valid HTTPS", func(t *testing.T) {
		input := "https://example.com"

		got, err := kernel.NewURL[TestResource](input)

		assertNoError(t, err)
		if got.String() != input {
			t.Errorf("got %q, want %q", got, input)
		}
	})

	t.Run("creates URL with valid HTTP", func(t *testing.T) {
		input := "http://example.com"

		got, err := kernel.NewURL[TestResource](input)

		assertNoError(t, err)
		if got.String() != input {
			t.Errorf("got %q, want %q", got, input)
		}
	})

	t.Run("allows empty URL as optional field", func(t *testing.T) {
		got, err := kernel.NewURL[TestResource]("")

		assertNoError(t, err)
		if got.String() != "" {
			t.Errorf("got %q, want empty string", got)
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		input := "  https://example.com  "
		want := "https://example.com"

		got, err := kernel.NewURL[TestResource](input)

		assertNoError(t, err)
		if got.String() != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("rejects invalid URL format", func(t *testing.T) {
		_, err := kernel.NewURL[TestResource]("not a url")

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("rejects non-HTTP(S) scheme", func(t *testing.T) {
		schemes := []string{
			"ftp://example.com",
			"file:///path/to/file",
			"javascript:alert('xss')",
			"data:text/plain,hello",
		}

		for _, scheme := range schemes {
			t.Run(scheme, func(t *testing.T) {
				_, err := kernel.NewURL[TestResource](scheme)

				assertError(t, err)
				assertErrorCode(t, err, kernel.EInvalid)
			})
		}
	})

	t.Run("accepts complex URLs", func(t *testing.T) {
		complexURLs := []string{
			"https://example.com/path/to/resource",
			"https://example.com:8080/path",
			"https://sub.example.com",
			"https://example.com/path?query=value&other=123",
			"https://example.com/path#fragment",
			"https://user:pass@example.com/path",
		}

		for _, url := range complexURLs {
			t.Run(url, func(t *testing.T) {
				got, err := kernel.NewURL[TestResource](url)

				assertNoError(t, err)
				if got.String() != url {
					t.Errorf("got %q, want %q", got, url)
				}
			})
		}
	})
}

func TestURL_String(t *testing.T) {
	want := "https://example.com"
	url := kernel.URL[TestResource](want)

	got := url.String()

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestURL_Validate(t *testing.T) {
	t.Run("valid HTTPS URL passes", func(t *testing.T) {
		url := kernel.URL[TestResource]("https://example.com")

		err := url.Validate()

		assertNoError(t, err)
	})

	t.Run("valid HTTP URL passes", func(t *testing.T) {
		url := kernel.URL[TestResource]("http://example.com")

		err := url.Validate()

		assertNoError(t, err)
	})

	t.Run("empty URL passes as optional", func(t *testing.T) {
		url := kernel.URL[TestResource]("")

		err := url.Validate()

		assertNoError(t, err)
	})

	t.Run("invalid format fails", func(t *testing.T) {
		url := kernel.URL[TestResource]("not a url")

		err := url.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("non-HTTP(S) scheme fails", func(t *testing.T) {
		url := kernel.URL[TestResource]("ftp://example.com")

		err := url.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("malformed URL fails", func(t *testing.T) {
		url := kernel.URL[TestResource]("https://[invalid")

		err := url.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})
}

func TestURL_TypeSafety(t *testing.T) {
	t.Run("URLs maintain type safety", func(t *testing.T) {
		type Resource1 struct{}
		type Resource2 struct{}

		url1 := kernel.URL[Resource1]("https://example.com")
		url2 := kernel.URL[Resource2]("https://example.com")

		// These are different types even with same value
		if url1.String() != url2.String() {
			t.Errorf("string values should be equal")
		}
	})
}
