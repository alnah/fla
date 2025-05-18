package aiclient

import (
	"context"
	"net/http"

	"github.com/alnah/fla/logger"
)

// baseClient holds common configuration and infrastructure fields
// shared across multiple AI-related clients. This struct is meant to
// be embedded or referenced by specific client implementations.
type baseClient struct {
	// api specific configuration
	provider provider
	apiKey   apiKey
	url      url
	model    model

	// infrastructure-related deps
	ctx        context.Context
	log        *logger.Logger
	httpClient *http.Client
	httpMethod httpMethod
}

// hasBase is an interface that requires a BaseClient method returning
// a pointer to a baseClient. This allows generic operations on any type
// that exposes access to a baseClient.
type hasBase interface{ BaseClient() *baseClient }

// option is a generic function type used to apply configuration to any
// client type that implements hasBase. This enables the functional options
// pattern to be reused across different client types without naming conflicts.
type option[T hasBase] func(T)

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
