package aiclient

import (
	"context"
	"net/http"

	"github.com/alnah/fla/logger"
)

type BaseClient struct {
	// api fields
	Provider Provider
	Model    AIModel
	// infra fields
	ctx        context.Context
	logger     *logger.Logger
	httpClient *http.Client
	httpMethod HTTPMethod
	provider   Provider
	url        URL
	apiKey     APIKey
}

type hasBase interface{ BaseClient() *BaseClient }

type Option[T hasBase] func(T)
