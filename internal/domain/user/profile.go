package user

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/alnah/fla/internal/domain/kernel"
)

// SocialMediaURL defines supported social media platforms for user profiles.
// Enables standardized social media integration across the platform.
type SocialMediaURL string

const (
	SocialMediaTwitter   SocialMediaURL = "twitter"   // Twitter/X platform integration
	SocialMediaLinkedIn  SocialMediaURL = "linkedin"  // Professional networking platform
	SocialMediaInstagram SocialMediaURL = "instagram" // Visual content platform
	SocialMediaTikTok    SocialMediaURL = "tiktok"    // Short-form video platform
	SocialMediaYouTube   SocialMediaURL = "youtube"   // Video content platform
	SocialMediaGitHub    SocialMediaURL = "github"    // Code repository platform
)

const (
	MSocialProfileInvalid      string = "Invalid social media profile."
	MSocialURLRequired         string = "Social media URL is required."
	MSocialURLInvalidFormat    string = "Invalid URL format."
	MSocialURLInvalidScheme    string = "Social media URL must use http or https scheme."
	MSocialPlatformUnsupported string = "Unsupported social media platform."
)

// SocialProfile represents validated social media profile links.
// Ensures profile URLs are correctly formatted and platform-appropriate.
type SocialProfile struct {
	Platform SocialMediaURL
	URL      string
}

// NewSocialProfile creates validated social media profile with platform-specific rules.
// Prevents broken social links and ensures proper platform URL formatting.
func NewSocialProfile(platform SocialMediaURL, profileURL string) (SocialProfile, error) {
	const op = "NewSocialProfile"

	profile := SocialProfile{
		Platform: platform,
		URL:      strings.TrimSpace(profileURL),
	}

	if err := profile.Validate(); err != nil {
		return SocialProfile{}, &kernel.Error{Operation: op, Cause: err}
	}

	return profile, nil
}

func (sp SocialProfile) String() string {
	return fmt.Sprintf("SocialProfile{Platform: %q, URL: %q}", sp.Platform, sp.URL)
}

// Validate ensures social profile meets platform-specific URL requirements.
// Prevents broken social media links and maintains platform consistency.
func (sp SocialProfile) Validate() error {
	const op = "SocialProfile.Validate"

	if err := sp.validateURLPresence(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := sp.validateURLFormat(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := sp.validateURLScheme(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := sp.validatePlatform(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	return nil
}

func (sp SocialProfile) validateURLPresence() error {
	const op = "SocialProfile.validateURLPresence"

	if sp.URL == "" {
		return &kernel.Error{
			Code:      kernel.EInvalid,
			Message:   MSocialURLRequired,
			Operation: op,
		}
	}

	return nil
}

func (sp SocialProfile) validateURLFormat() error {
	const op = "SocialProfile.validateURLFormat"

	u, err := url.Parse(sp.URL)
	if err != nil || u.Host == "" {
		return &kernel.Error{
			Code:      kernel.EInvalid,
			Message:   MSocialURLInvalidFormat,
			Operation: op,
			Cause:     err,
		}
	}

	return nil
}

func (sp SocialProfile) validateURLScheme() error {
	const op = "SocialProfile.validateURLScheme"

	parsedURL, _ := url.Parse(sp.URL)
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return &kernel.Error{
			Code:      kernel.EInvalid,
			Message:   MSocialURLInvalidScheme,
			Operation: op,
		}
	}

	return nil
}

func (sp SocialProfile) validatePlatform() error {
	const op = "SocialProfile.validatePlatform"

	switch sp.Platform {
	case SocialMediaLinkedIn, SocialMediaInstagram, SocialMediaTwitter,
		SocialMediaTikTok, SocialMediaYouTube, SocialMediaGitHub:
		return nil
	default:
		return &kernel.Error{
			Code:      kernel.EConflict,
			Message:   MSocialPlatformUnsupported,
			Operation: op,
		}
	}
}

// ProfilePicture type marker for URL generic
type ProfilePicture struct{}
