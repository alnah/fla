package aiclient

import (
	"context"
	"net/http"

	"github.com/alnah/fla/logger"
)

type baseClient struct {
	// api fields
	provider provider
	apiKey   apiKey
	url      url
	model    model
	// infra fields
	ctx        context.Context
	log        *logger.Logger
	httpClient *http.Client
	httpMethod httpMethod
}

type hasBase interface{ BaseClient() *baseClient }

type option[T hasBase] func(T)
