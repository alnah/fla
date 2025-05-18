package aiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alnah/fla/breaker"
	"github.com/alnah/fla/logger"
	"github.com/alnah/fla/retrier"
	"github.com/alnah/fla/transport"
)

const anthropicVersion string = "2023-06-01"

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

func (c *ChatClient) BaseClient() *baseClient { return c.base }

func WithTemperature(t Temperature) option[*ChatClient] {
	return func(c *ChatClient) { c.temperature = t }
}
func WithMessages(ms Messages) option[*ChatClient] {
	return func(c *ChatClient) { c.messages = ms }
}
func WithMaxTokens(mt MaxTokens) option[*ChatClient] {
	return func(c *ChatClient) { c.maxTokens = mt }
}

// NewChatClient builds a ChatClient with default transport chain.
func NewChatClient(opts ...option[*ChatClient]) (*ChatClient, error) {
	c := &ChatClient{base: &baseClient{}}
	for _, opt := range opts {
		opt(c)
	}

	if err := c.applyDefaults().setProviderFlag().validate(); err != nil {
		return nil, NewChatClientError(c.base.provider, "failed to build chat client", err)
	}

	return c, nil
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

// MarshalJSON handles provider-specific JSON fields.
func (c ChatClient) MarshalJSON() ([]byte, error) {
	v := c.maxTokens.Int()
	switch {
	case c.useOpenAI:
		type openaiPayload struct {
			Model       string    `json:"model"`
			Temperature float32   `json:"temperature"`
			Messages    []Message `json:"messages"`
			MaxTokens   *int      `json:"max_completion_tokens,omitempty"`
		}
		payload := openaiPayload{
			Model:       c.base.model.String(),
			Temperature: c.temperature.Float32(),
			Messages:    append([]Message{{Role: RoleSystem, Content: c.system}}, c.messages...),
			MaxTokens:   (*int)(&c.maxTokens),
		}
		return json.Marshal(payload)
	case c.useAnthropic:
		type anthropicPayload struct {
			Model               string    `json:"model"`
			System              string    `json:"system"`
			Messages            []Message `json:"messages"`
			MaxCompletionTokens *int      `json:"max_tokens,omitempty"`
			Temperature         float32   `json:"temperature,omitempty"`
		}
		payload := anthropicPayload{
			Model:               c.base.model.String(),
			System:              c.system,
			Messages:            c.messages,
			MaxCompletionTokens: &v,
			Temperature:         c.temperature.Float32(),
		}
		return json.Marshal(payload)
	default:
		return nil, fmt.Errorf("no provider configured")
	}
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

func (c *ChatClient) validate() error {
	if c.base.ctx == nil {
		return errors.New("context must be provided")
	}
	if c.base.log == nil {
		return errors.New("logger must be set")
	}
	if c.base.httpClient == nil {
		return errors.New("http client must be set")
	}
	if err := c.base.provider.Validate(); err != nil {
		return err
	}
	if err := c.base.url.Validate(); err != nil {
		return err
	}
	if err := c.base.apiKey.Validate(); err != nil {
		return err
	}
	if err := c.base.httpMethod.Validate(); err != nil {
		return err
	}
	if err := c.base.model.Validate(); err != nil {
		return err
	}
	if err := c.maxTokens.Validate(); err != nil {
		return err
	}
	if err := c.temperature.Validate(c.base.model); err != nil {
		return err
	}
	if err := c.messages.Validate(); err != nil {
		return err
	}
	if c.useOpenAI == c.useAnthropic {
		return errors.New("must configure exactly one provider: openai or anthropic")
	}
	if c.useOpenAI && c.base.provider != ProviderOpenAI {
		return fmt.Errorf("url indicates openai but provider is %s", c.base.provider)
	}
	if c.useAnthropic && c.base.provider != ProviderAnthropic {
		return fmt.Errorf("url indicates anthropic but provider is %s", c.base.provider)
	}
	switch {
	case c.useOpenAI:
		switch c.base.model {
		case ModelReasoningOpenAI, ModelFlagshipOpenAI, ModelCheapOpenAI:
			// ok
		default:
			return fmt.Errorf("model %s not supported by openai", c.base.model)
		}

	case c.useAnthropic:
		switch c.base.model {
		case ModelReasoningAnthropic, ModelCheapAnthropic:
			// ok
		default:
			return fmt.Errorf("model %s not supported by anthropic", c.base.model)
		}
	}
	if c.useAnthropic {
		for _, m := range c.messages {
			if m.Role == RoleSystem {
				return errors.New("system message must be passed via system field, not via messages, when using anthropic")
			}
		}
	}
	if c.useOpenAI {
		var systemCount int
		for _, m := range c.messages {
			if m.Role == RoleSystem {
				systemCount++
			}
		}
		if systemCount > 1 {
			return errors.New("system message must be passed once, when using openai")
		}
	}
	return nil
}

func (c *ChatClient) newTransportChain() http.RoundTripper {
	return transport.Chain(
		c.base.httpClient.Transport,
		transport.AddHeader("Content-Type", "application/json"),
		c.addAuthHeader(),
		c.addSpecHeader(),
		transport.AddHeader("User-Agent", "Fla/1.0"),
		transport.UseStatusClassifier(func(sc int) bool { return sc == 429 || sc >= 500 }, c.buildError()),
		transport.UseCircuitBreaker(breaker.New(breaker.ThirdPartyConfig())),
		transport.UseRetrier(retrier.NewExpBackoffJitter(), isRetryable),
		transport.UseLogger(c.base.log),
	)
}

func (c *ChatClient) addAuthHeader() transport.Middleware {
	if c.base.provider == ProviderOpenAI {
		return transport.AddHeader("Authorization", "Bearer "+c.base.apiKey.GetEnv())
	}
	return transport.AddHeader("x-api-key", c.base.apiKey.GetEnv())
}

func (c *ChatClient) addSpecHeader() transport.Middleware {
	if c.useAnthropic {
		return transport.AddHeader("anthropic-version", anthropicVersion)
	}
	return transport.AddHeader("key", "")
}

func (c *ChatClient) buildError() transport.ErrorFactoryFunc {
	if c.useOpenAI {
		return buildOpenaiError
	}
	return buildAnthropicError
}

// chatResponse holds the result of chat completions for chat completion providers.
type chatResponse struct {
	content string
}

func (cc chatResponse) Content() string {
	if cc.content != "" {
		return cc.content
	}
	return ""
}

// parseResponse extracts a ChatCompletion from raw JSON depending on provider.
func (c *ChatClient) parseResponse(byt []byte) (chatResponse, error) {
	switch {
	case c.useOpenAI:
		type openaiPayload struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		var payload openaiPayload
		if err := json.Unmarshal(byt, &payload); err != nil {
			return chatResponse{}, fmt.Errorf("failed to unmarshal response body: %w", err)
		}
		if len(payload.Choices) == 0 {
			return chatResponse{}, fmt.Errorf("no choices in OpenAI response")
		}
		return chatResponse{content: payload.Choices[0].Message.Content}, nil

	case c.useAnthropic:
		type anthropicPayload struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		}
		var payload anthropicPayload
		if err := json.Unmarshal(byt, &payload); err != nil {
			return chatResponse{}, fmt.Errorf("failed to unmarshal response body: %w", err)
		}
		if len(payload.Content) == 0 {
			return chatResponse{}, fmt.Errorf("no content in Anthropic response")
		}
		return chatResponse{content: payload.Content[0].Text}, nil

	default:
		return chatResponse{}, fmt.Errorf("no provider configured")
	}
}
