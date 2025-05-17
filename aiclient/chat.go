package aiclient

import (
	"bytes"
	"context"
	"encoding/json"
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
	// shared fields
	base *baseClient
	// api fields
	temperature Temperature
	system      string // Anthropic only
	messages    Messages
	maxTokens   MaxTokens
	// provider fields
	useOpenAI    bool // internal
	useAnthropic bool // internal

}

// NewChatClient builds a ChatClient with default transport chain.
func NewChatClient(options ...option[*ChatClient]) (*ChatClient, error) {
	c := &ChatClient{base: &baseClient{}}
	for _, opt := range options {
		opt(c)
	}

	if err := c.applyDefaults().setProviderFlag().validate(); err != nil {
		return nil, NewChatClientError(c.base.provider, "failed to build chat client", err)
	}

	return c, nil
}

func (c *ChatClient) applyDefaults() *ChatClient {
	if c.base.ctx == nil {
		c.base.ctx = context.Background()
	}
	if c.base.log == nil {
		c.base.log = logger.New()
	}
	if c.base.httpClient == nil {
		c.base.httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	if c.base.httpMethod == "" {
		c.base.httpMethod = httpMethod(http.MethodPost)
	}
	if c.maxTokens < 1 {
		c.maxTokens = MaxTokens(8192)
	}
	return c
}

func (c *ChatClient) setProviderFlag() *ChatClient {
	c.useOpenAI = strings.Contains(c.base.url.String(), ProviderOpenAI.String())
	c.useAnthropic = strings.Contains(c.base.url.String(), ProviderAnthropic.String())
	return c
}

func (c *ChatClient) Completion() (chatResponse, error) {
	byt, err := json.Marshal(c)
	if err != nil {
		return chatResponse{}, NewChatClientError(c.base.provider, "failed to marshal payload", err)
	}

	req, err := http.NewRequestWithContext(c.base.ctx, c.base.httpMethod.String(), c.base.url.String(), bytes.NewBuffer(byt))
	if err != nil {
		return chatResponse{}, NewChatClientError(c.base.provider, "failed to build http request", err)
	}

	c.base.httpClient.Transport = c.newTransportChain()
	res, err := c.base.httpClient.Do(req)
	if err != nil {
		return chatResponse{}, NewChatClientError(c.base.provider, "failed to send http request", err)
	}
	defer func() { _ = res.Body.Close() }()

	byt, err = io.ReadAll(res.Body)
	res.Body = io.NopCloser(bytes.NewReader(byt))
	if err != nil {
		return chatResponse{}, NewChatClientError(c.base.provider, "failed to read response body", err)
	}

	if res.StatusCode != 200 {
		return chatResponse{}, buildProviderError(c.base.provider, res)
	}

	completion, err := c.parseResponse(byt)
	if err != nil {
		return chatResponse{}, NewChatClientError(c.base.provider, "failed to parse response payload: %w", err)
	}

	return completion, nil
}
