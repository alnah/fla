package aiclient

import (
	"context"
	"net/http"

	"github.com/alnah/fla/logger"
)

type BaseClient struct {
	// api fields
	Model    AIModel
	provider Provider
	// infra fields
	ctx        context.Context
	logger     *logger.Logger
	httpClient *http.Client
	httpMethod HTTPMethod
	url        URL
	apiKey     APIKey
}

type hasBase interface{ BaseClient() *BaseClient }

type Option[T hasBase] func(T)
