package aiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alnah/fla/logger"
)

// ChatClient wraps HTTP transport with middleware for different AI providers.
// It automatically injects headers, error classification, circuit breaker,
// retrier, and logging.
type ChatClient struct {
	// api fields
	Model       AIModel
	Temperature Temperature
	System      string // Anthropic only
	Messages    Messages
	MaxTokens   MaxTokens

	// internal fields
	UseOpenAI    bool // internal
	UseAnthropic bool // internal

	// infra fields
	ctx        context.Context
	hc         *http.Client
	logger     *logger.Logger
	provider   Provider
	url        URL
	apiKey     APIKey
	httpMethod HTTPMethod
}

// NewChatClient builds a ChatClient with default transport chain.
func NewChatClient(options ...Option) (*ChatClient, error) {
	client := &ChatClient{
		ctx:        context.Background(),
		hc:         &http.Client{Timeout: 30 * time.Second},
		httpMethod: http.MethodPost,
		logger:     logger.Default(),
		MaxTokens:  8192,
	}

	for _, opt := range options {
		opt(client)
	}

	client.determineProvider()
	if err := client.Validate(); err != nil {
		return nil, fmt.Errorf("failed to build chat client: %w", err)
	}

	client.hc.Transport = client.configureTransportChain()
	return client, nil
}

func (c *ChatClient) determineProvider() {
	c.UseOpenAI = strings.Contains(c.url.String(), ProviderOpenAI.String())
	c.UseAnthropic = strings.Contains(c.url.String(), ProviderAnthropic.String())
}

func (c *ChatClient) Do() (ChatCompletion, error) {
	byt, err := json.Marshal(c)
	if err != nil {
		return ChatCompletion{}, NewChatClientError(c.provider, "failed to marshal payload", err)
	}

	req, err := http.NewRequestWithContext(c.ctx, c.httpMethod.String(), c.url.String(), bytes.NewBuffer(byt))
	if err != nil {
		return ChatCompletion{}, NewChatClientError(c.provider, "failed to build http request", err)
	}

	res, err := c.hc.Do(req)
	if err != nil {
		return ChatCompletion{}, NewChatClientError(c.provider, "failed to send http request", err)
	}
	defer func() { _ = res.Body.Close() }()

	byt, err = io.ReadAll(res.Body)
	res.Body = io.NopCloser(bytes.NewReader(byt))
	if err != nil {
		return ChatCompletion{}, NewChatClientError(c.provider, "failed to read response body", err)
	}

	if res.StatusCode != 200 {
		return ChatCompletion{}, BuildProviderError(c.provider, res)
	}

	completion, err := c.ParseResponse(byt)
	if err != nil {
		return ChatCompletion{}, NewChatClientError(c.provider, "failed to parse response payload: %w", err)
	}

	return completion, nil
}
