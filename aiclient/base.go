package aiclient

import (
	"context"
	"net/http"

	"github.com/alnah/fla/logger"
)

type baseClient struct {
	// api fields
	model    aiModel
	provider provider
	// infra fields
	ctx        context.Context
	logger     *logger.Logger
	httpClient *http.Client
	httpMethod httpMethod
	url        url
	apiKey     apiKey
}

type hasBase interface{ BaseClient() *baseClient }

type option[T hasBase] func(T)
