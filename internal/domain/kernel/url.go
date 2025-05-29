package kernel

import (
	"net/url"
	"strings"
)

const (
	MInvalidURL       string = "Invalid URL."
	MInvalidURLFormat string = "Invalid URL format."
	MInvalidURLScheme string = "URL must use http or https scheme."
)

// URL represents validated URLs for resources with security validation.
// Generic type parameter T indicates the context this URL is used in.
type URL[T any] string

// NewURL creates validated URL with security and format checking.
// Prevents broken links and ensures secure resource loading.
func NewURL[T any](urlStr string) (URL[T], error) {
	const op = "NewURL"

	if urlStr == "" {
		return "", nil // Optional field
	}

	u := URL[T](strings.TrimSpace(urlStr))
	if err := u.Validate(); err != nil {
		return "", &Error{Operation: op, Cause: err}
	}

	return u, nil
}

func (u URL[T]) String() string { return string(u) }

// Validate ensures URL is properly formatted and uses secure protocols.
// Prevents broken references and potential security vulnerabilities.
func (u URL[T]) Validate() error {
	const op = "URL.Validate"

	if u.String() == "" {
		return nil // Optional field
	}

	if err := u.validateFormat(); err != nil {
		return &Error{Operation: op, Cause: err}
	}

	if err := u.validateScheme(); err != nil {
		return &Error{Operation: op, Cause: err}
	}

	return nil
}

func (u URL[T]) validateFormat() error {
	const op = "URL.validateFormat"

	_, err := url.Parse(u.String())
	if err != nil {
		return &Error{
			Code:      EInvalid,
			Message:   MInvalidURLFormat,
			Operation: op,
			Cause:     err,
		}
	}

	return nil
}

func (u URL[T]) validateScheme() error {
	const op = "URL.validateScheme"

	parsedURL, _ := url.Parse(u.String())
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return &Error{
			Code:      EInvalid,
			Message:   MInvalidURLScheme,
			Operation: op,
		}
	}

	return nil
}
