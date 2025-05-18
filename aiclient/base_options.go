package aiclient

import (
	"context"
	"net/http"

	"github.com/alnah/fla/logger"
)

// WithContext sets the context for the base client.
// Useful for controlling cancellation, timeouts, etc.
func WithContext[T hasBase](c context.Context) option[T] {
	return func(t T) { t.BaseClient().ctx = c }
}

// WithLogger assigns a logger to the base client.
// Enables structured and contextual logging.
func WithLogger[T hasBase](l *logger.Logger) option[T] {
	return func(t T) { t.BaseClient().log = l }
}

// WithProvider sets the AI provider for the base client.
func WithProvider[T hasBase](p provider) option[T] {
	return func(t T) { t.BaseClient().provider = p }
}

// WithURL sets the base URL for the API the client interacts with.
func WithURL[T hasBase](u url) option[T] {
	return func(t T) { t.BaseClient().url = u }
}

// WithAPIKey assigns the API key used for authenticating requests.
func WithAPIKey[T hasBase](a apiKey) option[T] {
	return func(t T) { t.BaseClient().apiKey = a }
}

// WithModel selects the model used by the AI provider.
func WithModel[T hasBase](m model) option[T] {
	return func(t T) { t.BaseClient().model = m }
}

// WithHTTPClient sets a custom HTTP client for outgoing requests.
// This is useful for testing, and, perhaps, for customizing timeouts, transport, etc.
func WithHTTPClient[T hasBase](hc *http.Client) option[T] {
	return func(t T) { t.BaseClient().httpClient = hc }
}

// WithHTTPMethod defines the HTTP method to use (e.g., GET, POST).
func WithHTTPMethod[T hasBase](hm httpMethod) option[T] {
	return func(t T) { t.BaseClient().httpMethod = hm }
}
