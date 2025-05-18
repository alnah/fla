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
