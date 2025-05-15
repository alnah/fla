package aiclient

import (
	"context"
	"net/http"

	"github.com/alnah/fla/logger"
)

type Base struct {
	// api fields
	Model    aiModel
	provider provider
	// infra fields
	ctx        context.Context
	logger     *logger.Logger
	httpClient *http.Client
	httpMethod httpMethod
	url        url
	apiKey     apiKey
}

type hasBase interface{ BaseClient() *Base }

type option[T hasBase] func(T)
