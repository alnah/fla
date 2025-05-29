package user_test

import (
	"testing"

	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/user"
)

func TestNewSocialProfile(t *testing.T) {
	t.Run("creates social profile with valid input", func(t *testing.T) {
		tests := []struct {
			platform user.SocialMediaURL
			url      string
		}{
			{user.SocialMediaTwitter, "https://twitter.com/username"},
			{user.SocialMediaLinkedIn, "https://linkedin.com/in/username"},
			{user.SocialMediaInstagram, "https://instagram.com/username"},
			{user.SocialMediaTikTok, "https://tiktok.com/@username"},
			{user.SocialMediaYouTube, "https://youtube.com/channel/abc123"},
			{user.SocialMediaGitHub, "https://github.com/username"},
			{user.SocialMediaTwitter, "http://twitter.com/username"}, // HTTP also valid
		}

		for _, tt := range tests {
			t.Run(string(tt.platform), func(t *testing.T) {
				got, err := user.NewSocialProfile(tt.platform, tt.url)

				assertNoError(t, err)
				if got.Platform != tt.platform {
					t.Errorf("Platform: got %v, want %v", got.Platform, tt.platform)
				}
				if got.URL != tt.url {
					t.Errorf("URL: got %q, want %q", got.URL, tt.url)
				}
			})
		}
	})

	t.Run("trims whitespace from URL", func(t *testing.T) {
		input := "  https://twitter.com/username  "
		want := "https://twitter.com/username"

		got, err := user.NewSocialProfile(user.SocialMediaTwitter, input)

		assertNoError(t, err)
		if got.URL != want {
			t.Errorf("got %q, want %q", got.URL, want)
		}
	})

	t.Run("rejects empty URL", func(t *testing.T) {
		_, err := user.NewSocialProfile(user.SocialMediaTwitter, "")

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("rejects whitespace only URL", func(t *testing.T) {
		_, err := user.NewSocialProfile(user.SocialMediaTwitter, "   ")

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("rejects invalid URL format", func(t *testing.T) {
		invalidURLs := []string{
			"not a url",
			"twitter.com/username", // missing scheme
			"https://",             // incomplete URL
			"https://[invalid",     // malformed URL
		}

		for _, url := range invalidURLs {
			t.Run(url, func(t *testing.T) {
				_, err := user.NewSocialProfile(user.SocialMediaTwitter, url)

				assertError(t, err)
				assertErrorCode(t, err, kernel.EInvalid)
			})
		}
	})

	t.Run("rejects non-HTTP(S) schemes", func(t *testing.T) {
		invalidSchemes := []string{
			"ftp://twitter.com/username",
			"javascript:alert('xss')",
			"data:text/plain,hello",
			"file:///etc/passwd",
		}

		for _, url := range invalidSchemes {
			t.Run(url, func(t *testing.T) {
				_, err := user.NewSocialProfile(user.SocialMediaTwitter, url)

				assertError(t, err)
				assertErrorCode(t, err, kernel.EInvalid)
			})
		}
	})

	t.Run("rejects unsupported platforms", func(t *testing.T) {
		unsupportedPlatforms := []user.SocialMediaURL{
			"facebook",
			"snapchat",
			"reddit",
			"",
		}

		for _, platform := range unsupportedPlatforms {
			t.Run(string(platform), func(t *testing.T) {
				_, err := user.NewSocialProfile(platform, "https://example.com")

				assertError(t, err)
				assertErrorCode(t, err, kernel.EConflict)
			})
		}
	})
}

func TestSocialProfile_String(t *testing.T) {
	profile, _ := user.NewSocialProfile(user.SocialMediaTwitter, "https://twitter.com/username")

	got := profile.String()
	want := `SocialProfile{Platform: "twitter", URL: "https://twitter.com/username"}`

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSocialProfile_Validate(t *testing.T) {
	t.Run("valid profile passes", func(t *testing.T) {
		profile := user.SocialProfile{
			Platform: user.SocialMediaTwitter,
			URL:      "https://twitter.com/username",
		}

		err := profile.Validate()

		assertNoError(t, err)
	})

	t.Run("empty URL fails", func(t *testing.T) {
		profile := user.SocialProfile{
			Platform: user.SocialMediaTwitter,
			URL:      "",
		}

		err := profile.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("invalid URL format fails", func(t *testing.T) {
		profile := user.SocialProfile{
			Platform: user.SocialMediaTwitter,
			URL:      "not a url",
		}

		err := profile.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("non-HTTP(S) scheme fails", func(t *testing.T) {
		profile := user.SocialProfile{
			Platform: user.SocialMediaTwitter,
			URL:      "ftp://twitter.com/username",
		}

		err := profile.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("unsupported platform fails", func(t *testing.T) {
		profile := user.SocialProfile{
			Platform: "facebook",
			URL:      "https://facebook.com/username",
		}

		err := profile.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EConflict)
	})
}

func TestSocialMediaURLConstants(t *testing.T) {
	// Ensure constants have expected values
	tests := []struct {
		name     string
		platform user.SocialMediaURL
		want     string
	}{
		{"Twitter", user.SocialMediaTwitter, "twitter"},
		{"LinkedIn", user.SocialMediaLinkedIn, "linkedin"},
		{"Instagram", user.SocialMediaInstagram, "instagram"},
		{"TikTok", user.SocialMediaTikTok, "tiktok"},
		{"YouTube", user.SocialMediaYouTube, "youtube"},
		{"GitHub", user.SocialMediaGitHub, "github"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.platform) != tt.want {
				t.Errorf("got %q, want %q", tt.platform, tt.want)
			}
		})
	}
}
