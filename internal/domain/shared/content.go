package shared

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/alnah/fla/internal/domain/kernel"
)

const (
	MinTitleLength       int = 10
	MaxTitleLength       int = 100
	MinDescriptionLength int = 0 // Optional field
	MaxDescriptionLength int = 300
	MaxSlugLength        int = MaxTitleLength + 10
)

const (
	MSlugInvalidChars string = "Slug contains invalid characters."
	MSlugGeneration   string = "Slug could not be generated."
)

// Title represents content headlines with length validation for readability.
// Ensures titles are descriptive enough while maintaining display compatibility.
type Title string

// NewTitle creates validated content title with readability requirements.
// Balances descriptive content with practical length constraints for SEO and UI.
func NewTitle(title string) (Title, error) {
	const op = "NewTitle"

	t := Title(strings.TrimSpace(title))
	if err := t.Validate(); err != nil {
		return "", &kernel.Error{Operation: op, Cause: err}
	}

	return t, nil
}

func (t Title) String() string { return string(t) }

// Validate ensures title meets editorial standards for effective communication.
// Enforces minimum descriptiveness while preventing overly long headlines.
func (t Title) Validate() error {
	const op = "Title.Validate"

	if err := kernel.ValidatePresence("title", t.String(), op); err != nil {
		return err
	}

	if err := kernel.ValidateLength("title", t.String(), MinTitleLength, MaxTitleLength, op); err != nil {
		return err
	}

	return nil
}

// Description provides explanatory text for entities with length constraints.
// Enables rich metadata while maintaining UI display boundaries.
type Description string

// NewDescription creates validated descriptive text with length checking.
// Ensures descriptions fit within meta tag limits and UI components.
func NewDescription(desc string) (Description, error) {
	const op = "NewDescription"

	d := Description(strings.TrimSpace(desc))
	if err := d.Validate(); err != nil {
		return "", &kernel.Error{Operation: op, Cause: err}
	}

	return d, nil
}

func (d Description) String() string { return string(d) }

// Validate ensures description length stays within practical display limits.
// Prevents overly long descriptions that break UI layouts and meta tags.
func (d Description) Validate() error {
	const op = "Description.Validate"

	if err := kernel.ValidateLength("description", d.String(), MinDescriptionLength, MaxDescriptionLength, op); err != nil {
		return err
	}

	return nil
}

// Slug represents URL-friendly identifiers for content addressing.
// Enables SEO-optimized URLs and clean content navigation paths.
type Slug string

// NewSlug generates URL-safe slug from input text with automatic formatting.
// Creates clean URLs while preserving content meaning and SEO value.
func NewSlug(input string) (Slug, error) {
	const op = "NewSlug"

	slug, err := generateSlug(input)
	if err != nil {
		return "", &kernel.Error{Operation: op, Cause: err}
	}

	s := Slug(slug)
	if err := s.Validate(); err != nil {
		return "", &kernel.Error{Operation: op, Cause: err}
	}

	return s, nil
}

func (s Slug) String() string { return string(s) }

// Validate ensures slug meets URL standards and length requirements.
// Prevents broken URLs and ensures compatibility with web routing.
func (s Slug) Validate() error {
	const op = "Slug.Validate"

	if err := kernel.ValidatePresence("slug", s.String(), op); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := kernel.ValidateMaxLength("slug", s.String(), MaxSlugLength, op); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := s.validateSlugFormat(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	return nil
}

func (s Slug) validateSlugFormat() error {
	const op = "Slug.validateSlugFormat"

	if !slugFormatRe.MatchString(s.String()) {
		return &kernel.Error{
			Code:      kernel.EInvalid,
			Message:   MSlugInvalidChars,
			Operation: op,
		}
	}

	return nil
}

// transliterationMap provides basic transliteration for common non-ASCII characters
var transliterationMap = map[rune]string{
	// French
	'À': "A", 'à': "a",
	'Â': "A", 'â': "a",
	'Ä': "A", 'ä': "a",
	'Æ': "AE", 'æ': "ae",
	'Ç': "C", 'ç': "c",
	'È': "E", 'è': "e",
	'É': "E", 'é': "e",
	'Ê': "E", 'ê': "e",
	'Ë': "E", 'ë': "e",
	'Î': "I", 'î': "i",
	'Ï': "I", 'ï': "i",
	'Ô': "O", 'ô': "o",
	'Œ': "OE", 'œ': "oe",
	'Ù': "U", 'ù': "u",
	'Û': "U", 'û': "u",
	'Ü': "U", 'ü': "u",
	'Ÿ': "Y", 'ÿ': "y",

	// Spanish
	'Á': "A", 'á': "a",
	'Í': "I", 'í': "i",
	'Ñ': "N", 'ñ': "n",
	'Ó': "O", 'ó': "o",
	'Ú': "U", 'ú': "u",
	'¿': "", '¡': "", // Remove inverted punctuation

	// Portuguese
	'Ã': "A", 'ã': "a",
	'Õ': "O", 'õ': "o",
	// À, Á, Â, Ç, É, Ê, Í, Ó, Ô, Ú already covered above

	// English (rare but sometimes used)
	'Ð': "D", 'ð': "d",
	'Þ': "TH", 'þ': "th",

	// Polish
	'Ł': "L", 'ł': "l",
	'Ą': "A", 'ą': "a",
	'Ć': "C", 'ć': "c",
	'Ę': "E", 'ę': "e",
	'Ń': "N", 'ń': "n",
	'Ś': "S", 'ś': "s",
	'Ź': "Z", 'ź': "z",
	'Ż': "Z", 'ż': "z",

	// German
	'Ö': "O", 'ö': "o",
	'ß': "ss",

	// Scandinavian
	'Å': "A", 'å': "a",
	'Ø': "O", 'ø': "o",

	// Common symbols and ligatures
	'&': "-",
	'@': "-",
	'°': "-",

	// Currency symbols removed completely (no replacement)
	'€': "",
	'£': "",
	'$': "",
}

// transliterate converts special characters to their ASCII equivalents
func transliterate(s string) string {
	var result strings.Builder
	for _, r := range s {
		if replacement, ok := transliterationMap[r]; ok {
			result.WriteString(replacement)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// generateSlug transforms text into URL-safe format with international support.
// Handles accents, special characters, and length constraints automatically.
func generateSlug(input string) (string, error) {
	const op = "generateSlug"

	// Trim whitespace first
	input = strings.TrimSpace(input)

	// Check for empty input after trimming
	if input == "" {
		return "", &kernel.Error{
			Code:      kernel.EInvalid,
			Message:   MSlugGeneration,
			Operation: op,
		}
	}

	// First apply transliteration for special characters
	s := transliterate(input)

	// Then lowercase
	s = strings.ToLower(s)

	// Strip remaining accents/diacritics
	var err error
	if s, _, err = transform.String(accentRemover, s); err != nil {
		return "", &kernel.Error{
			Code:      kernel.EInternal,
			Message:   MSlugGeneration,
			Operation: op,
			Cause:     err,
		}
	}

	// Replace any sequence of non-alphanumeric chars with a single hyphen
	s = nonAlphaRe.ReplaceAllString(s, "-")

	// Trim hyphens from the ends
	s = strings.Trim(s, "-")

	// FIXED: Check if we have an empty or invalid slug after processing
	if s == "" || !containsAlphanumeric(s) {
		return "", &kernel.Error{
			Code:      kernel.EInvalid,
			Message:   MSlugGeneration,
			Operation: op,
		}
	}

	// Enforce MaxSlugLength without breaking runes; trim trailing hyphens again
	if utf8.RuneCountInString(s) > MaxSlugLength {
		r := []rune(s)
		s = string(r[:MaxSlugLength])
		s = strings.TrimRight(s, "-")
	}

	return s, nil
}

// containsAlphanumeric checks if string contains at least one alphanumeric character
func containsAlphanumeric(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return true
		}
	}
	return false
}

// Precompile the accent‐removal transformer and non-alphanumeric regex.
var (
	slugFormatRe  = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	accentRemover = transform.Chain(
		norm.NFD,
		runes.Remove(runes.In(unicode.Mn)),
		norm.NFC,
	)
	nonAlphaRe = regexp.MustCompile(`[^a-z0-9]+`)
)
