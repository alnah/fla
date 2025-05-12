package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	ai "github.com/alnah/fla/aiclient"
	"github.com/alnah/fla/aiclient/breaker"
	"github.com/alnah/fla/aiclient/retrier"
	"github.com/alnah/fla/aiclient/transport"
	"github.com/alnah/fla/logger"
)

const (
	apiKeyFromEnv    = ai.EnvAPIKeyAnthropic
	gateway          = "https://api.anthropic.com/v1/"
	pathMessages     = "messages"
	anthropicVersion = "2023-06-01"
)

type model string

func (m model) String() string { return string(m) }

const (
	// ModelReasoning selects the compact reasoning-optimized LLM
	ModelReasoning model = "claude-3-7-sonnet-latest"
	// ModelCostOptimized selects a lower-cost variant of the reasoning LLM.
	ModelCostOptimized model = "claude-3-5-haiku-latest"
)

type option func(*Chat)

// WithModel sets which model to call.
func WithModel(m model) option { return func(c *Chat) { c.Model = m } }

// WithLogger attaches structured logging of request/response events.
func WithLogger(log *logger.Logger) option { return func(c *Chat) { c.log = log } }

// WithHTTPClient overrides the default HTTP client for custom transport configuration.
func WithHTTPClient(hc *http.Client) option { return func(c *Chat) { c.hc = hc } }

// WithContext supplies a context for cancellation and deadlines.
func WithContext(ctx context.Context) option { return func(c *Chat) { c.ctx = ctx } }

// WithAPIKey provides the OpenAI authentication token explicitly.
func WithAPIKey(s string) option { return func(c *Chat) { c.apiKey = s } }

// WithURL directs requests to a custom endpoint, useful for testing.
func WithURL(s string) option { return func(c *Chat) { c.url = s } }

// WithMessages initiliazes the chat history to send in the request.
func WithMessages(m []ai.Message) option { return func(c *Chat) { c.Messages = m } }

// WithMaxTokens limits how many tokens the model may generate.
func WithMaxTokens(n int) option { return func(c *Chat) { c.MaxTokens = n } }

// WithSystem guides the model with general instructions.
func WithSystem(s string) option { return func(c *Chat) { c.System = s } }

// WithTemperature adjusts response variability; higher values yield more diverse outputs.
func WithTemperature(n float32) option { return func(c *Chat) { c.Temperature = n } }

// Chat manages a conversation request to the Anthropic API.
type Chat struct {
	Model         model        `json:"model"`
	Messages      []ai.Message `json:"messages"`
	MaxTokens     int          `json:"max_tokens"`
	System        string       `json:"system,omitempty"`
	Temperature   float32      `json:"temperature,omitempty"`
	hc            *http.Client
	log           *logger.Logger
	ctx           context.Context
	transportOnce sync.Once
	systemOnce    sync.Once
	version       string
	apiKey        string
	method        string
	url           string
}

// NewChat constructs a chat client. Defaults model to cost optimized.
// Callers can override settings via options to customize transcription.
func NewChat(opts ...option) *Chat {
	c := &Chat{
		Model:     ModelCostOptimized,
		MaxTokens: 8192,
		hc:        &http.Client{Timeout: 30 * time.Second},
		ctx:       context.Background(),
		version:   anthropicVersion,
		apiKey:    os.Getenv(apiKeyFromEnv),
		method:    http.MethodPost,
		url:       gateway + pathMessages,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// SetSystem guides the model with general instructions.
func (c *Chat) SetSystem(instructions string) *Chat {
	c.systemOnce.Do(func() {
		c.System = instructions
	})
	return c
}

// AddMessage appends a new role/content pair to the conversation history.
// It returns the same chat to enable fluent chaining.
func (c *Chat) AddMessage(role ai.Role, content string) *Chat {
	c.Messages = append(c.Messages, ai.Message{
		Role:    role,
		Content: content,
	})
	return c
}

// Completion sends the assembled chat request to the Anthropic API.
func (c *Chat) Completion() (ai.Completion, error) {
	if len(c.Messages) == 0 {
		return ai.Completion{}, ai.NewChatError(ai.ProviderAnthropic, "messages required", nil)
	}

	byt, err := json.Marshal(c)
	if err != nil {
		return ai.Completion{}, ai.NewChatError(ai.ProviderAnthropic, "failed to marshal payload", err)
	}

	req, err := http.NewRequestWithContext(c.ctx, c.method, c.url, bytes.NewBuffer(byt))
	if err != nil {
		return ai.Completion{}, ai.NewChatError(ai.ProviderAnthropic, "failed to build http request", err)
	}

	c.transportOnce.Do(func() {
		orig := c.hc.Transport
		if orig == nil {
			orig = http.DefaultTransport
		}
		c.hc.Transport = transport.Chain(
			orig,
			transport.AddHeader("Content-Type", "application/json"),
			transport.AddHeader("x-api-key", c.apiKey),
			transport.AddHeader("User-Agent", "Fla/1.0"),
			transport.AddHeader("anthropic-version", c.version),
			transport.ClassifyStatus(ai.ProviderAnthropic),
			transport.UseCircuitBreaker(breaker.New()),
			transport.UseRetrier(retrier.New(), isRetryable),
			transport.UseLogger(c.log),
		)
	})

	res, err := c.hc.Do(req)
	if err != nil {
		return ai.Completion{}, ai.NewChatError(ai.ProviderAnthropic, "failed to send http request", err)
	}
	defer func() { _ = res.Body.Close() }()

	byt, err = io.ReadAll(res.Body)
	res.Body = io.NopCloser(bytes.NewReader(byt))
	if err != nil {
		return ai.Completion{}, ai.NewChatError(ai.ProviderAnthropic, "failed to read response body", err)
	}

	if res.StatusCode != http.StatusOK {
		httpErr := ai.NewHTTPError(ai.ProviderOpenAI, res)
		return ai.Completion{}, ai.NewChatError(ai.ProviderOpenAI, "http request failed", httpErr)
	}

	var completion ai.Completion
	if err = json.Unmarshal(byt, &completion); err != nil {
		return ai.Completion{}, ai.NewChatError(ai.ProviderAnthropic, "failed to unmarshal response body", err)
	}

	return completion, nil
}

const (
	errRateLimited ai.ErrType = "rate_limit_error"
	errAPI         ai.ErrType = "api_error"
	errOverloaded  ai.ErrType = "overloaded_error"
)

var retryable = map[ai.ErrType]struct{}{
	errRateLimited: {},
	errAPI:         {},
	errOverloaded:  {},
}

var isRetryable = ai.NewRetryClassifier(retryable)
