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

	// infra fields
	ctx        context.Context
	logger     *logger.Logger
	httpClient *http.Client
	httpMethod HTTPMethod
	provider   Provider
	url        URL
	apiKey     APIKey

	// internal fields
	UseOpenAI    bool // internal
	UseAnthropic bool // internal

}

// NewChatClient builds a ChatClient with default transport chain.
func NewChatClient(options ...Option) (*ChatClient, error) {
	c := &ChatClient{}
	for _, opt := range options {
		opt(c)
	}

	if err := c.applyDefaults().setProviderFlag().validate(); err != nil {
		return nil, fmt.Errorf("failed to build chat client: %w", err)
	}

	c.httpClient.Transport = c.configureTransportChain()
	return c, nil
}

func (c *ChatClient) applyDefaults() *ChatClient {
	if c.logger == nil {
		c.logger = logger.New()
	}
	if c.httpClient == nil {
		c.httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	if c.httpMethod == "" {
		c.httpMethod = HTTPMethod(http.MethodPost)
	}
	if c.MaxTokens < 1 {
		c.MaxTokens = MaxTokens(8192)
	}
	return c
}

func (c *ChatClient) setProviderFlag() *ChatClient {
	c.UseOpenAI = strings.Contains(c.url.String(), ProviderOpenAI.String())
	c.UseAnthropic = strings.Contains(c.url.String(), ProviderAnthropic.String())
	return c
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

	res, err := c.httpClient.Do(req)
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
