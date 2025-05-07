package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/alnah/fla/ai"
	"github.com/alnah/fla/ai/breaker"
	"github.com/alnah/fla/ai/retrier"
	"github.com/alnah/fla/ai/transport"
	"github.com/alnah/fla/clog"
)

const (
	apiKeyFromEnv    = "ANTHROPIC_API_KEY" // #nosec G101: this is only the ENV var name, not a secret
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
func WithLogger(log *clog.Logger) option { return func(c *Chat) { c.log = log } }

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
	log           *clog.Logger
	ctx           context.Context
	transportOnce sync.Once
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
	c.System = instructions
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

type content struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type completion struct {
	Result []content `json:"content"`
	ID     string    `json:"id"`
}

// Content extracts the main text of the first choice, or returns empty
// if no content is available.
func (res *completion) Content() string {
	if len(res.Result) == 0 || res.Result[0].Text == "" {
		return ""
	}
	return res.Result[0].Text
}

// AnthropicCompletionError indicates a request failure due to invalid input
// or an API-side error, preserving the descriptive error message.
type AnthropicCompletionError struct{ message string }

func (e *AnthropicCompletionError) Error() string {
	return fmt.Sprintf("AnthropicCompletionError: %s", e.message)
}

// Completion sends the assembled chat request to the Anthropic API.
func (c *Chat) Completion() (completion, error) {
	if len(c.Messages) == 0 {
		return completion{}, &AnthropicCompletionError{message: "messages required"}
	}

	byt, err := json.Marshal(c)
	if err != nil {
		return completion{}, &AnthropicCompletionError{
			message: fmt.Sprintf("failed to marshal chat: %v", err),
		}
	}

	req, err := http.NewRequestWithContext(c.ctx, c.method, c.url, bytes.NewBuffer(byt))
	if err != nil {
		return completion{}, &AnthropicCompletionError{
			message: fmt.Sprintf("failed to create request: %v", err),
		}
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
			transport.ClassifyStatus,
			transport.UseCircuitBreaker(breaker.New()),
			transport.UseRetrier(retrier.New(), isRetryable),
			transport.UseLogger(c.log),
		)
	})

	res, err := c.hc.Do(req)
	if err != nil {
		return completion{}, &AnthropicCompletionError{
			message: fmt.Sprintf("failed to send HTTP request: %v", err),
		}
	}

	defer func() { _ = res.Body.Close() }()

	byt, err = io.ReadAll(res.Body)
	if err != nil {
		return completion{}, &AnthropicCompletionError{message: fmt.Sprintf("failed to read response body: %v", err)}
	}

	if res.StatusCode != http.StatusOK {
		var apiErr transport.APIError
		_ = json.Unmarshal(byt, &apiErr)

		return completion{}, &transport.HTTPError{
			Status:  res.StatusCode,
			Type:    apiErr.Error.Type,
			Code:    apiErr.Error.Code,
			Message: apiErr.Error.Message,
		}
	}

	var body completion
	if err = json.Unmarshal(byt, &body); err != nil {
		return completion{}, &AnthropicCompletionError{message: fmt.Sprintf("failed to unmarshal response body: %v", err)}
	}

	return body, nil

}

const (
	errRateLimited transport.ErrType = "rate_limit_error"
	errAPI         transport.ErrType = "api_error"
	errOverloaded  transport.ErrType = "overloaded_error"
)

var retryable = map[transport.ErrType]struct{}{
	errRateLimited: {},
	errAPI:         {},
	errOverloaded:  {},
}

var isRetryable = transport.NewRetryClassifier(retryable)
