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
	// http fields
	base *BaseClient
	// api fields
	Temperature Temperature
	System      string // Anthropic only
	Messages    Messages
	MaxTokens   MaxTokens
	// provider fields
	UseOpenAI    bool // internal
	UseAnthropic bool // internal

}

// NewChatClient builds a ChatClient with default transport chain.
func NewChatClient(options ...Option[*ChatClient]) (*ChatClient, error) {
	c := &ChatClient{base: &BaseClient{}}
	for _, opt := range options {
		opt(c)
	}

	if err := c.applyDefaults().setProviderFlag().validate(); err != nil {
		return nil, fmt.Errorf("failed to build chat client: %w", err)
	}

	c.base.httpClient.Transport = c.configureTransportChain()
	return c, nil
}

func (c *ChatClient) applyDefaults() *ChatClient {
	if c.base.ctx == nil {
		c.base.ctx = context.Background()
	}
	if c.base.logger == nil {
		c.base.logger = logger.New()
	}
	if c.base.httpClient == nil {
		c.base.httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	if c.base.httpMethod == "" {
		c.base.httpMethod = HTTPMethod(http.MethodPost)
	}
	if c.MaxTokens < 1 {
		c.MaxTokens = MaxTokens(8192)
	}
	return c
}

func (c *ChatClient) setProviderFlag() *ChatClient {
	c.UseOpenAI = strings.Contains(c.base.url.String(), ProviderOpenAI.String())
	c.UseAnthropic = strings.Contains(c.base.url.String(), ProviderAnthropic.String())
	return c
}

func (c *ChatClient) Do() (ChatCompletion, error) {
	byt, err := json.Marshal(c)
	if err != nil {
		return ChatCompletion{}, NewChatClientError(c.base.provider, "failed to marshal payload", err)
	}

	req, err := http.NewRequestWithContext(c.base.ctx, c.base.httpMethod.String(), c.base.url.String(), bytes.NewBuffer(byt))
	if err != nil {
		return ChatCompletion{}, NewChatClientError(c.base.provider, "failed to build http request", err)
	}

	res, err := c.base.httpClient.Do(req)
	if err != nil {
		return ChatCompletion{}, NewChatClientError(c.base.provider, "failed to send http request", err)
	}
	defer func() { _ = res.Body.Close() }()

	byt, err = io.ReadAll(res.Body)
	res.Body = io.NopCloser(bytes.NewReader(byt))
	if err != nil {
		return ChatCompletion{}, NewChatClientError(c.base.provider, "failed to read response body", err)
	}

	if res.StatusCode != 200 {
		return ChatCompletion{}, BuildProviderError(c.base.provider, res)
	}

	completion, err := c.ParseResponse(byt)
	if err != nil {
		return ChatCompletion{}, NewChatClientError(c.base.provider, "failed to parse response payload: %w", err)
	}

	return completion, nil
}
