package aiclient

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/alnah/fla/logger"
)

func TestChatClientNew_WithOptions(t *testing.T) {
	temperature := Temperature(1)
	messages := Messages{
		Message{
			Role:    RoleSystem,
			Content: "system",
		},
		Message{
			Role:    RoleUser,
			Content: "say test",
		},
		Message{
			Role:    RoleAssistant,
			Content: "test",
		},
	}
	ctx := context.Background()
	httpClient := &http.Client{Timeout: 30 * time.Second}
	logger := logger.New()
	httpMethod := HTTPMethod(http.MethodPost)
	maxTokens := MaxTokens(8192)
	chat, err := NewChatClient(
		WithProvider[*ChatClient](ProviderOpenAI),
		WithURL[*ChatClient](URLChatCompletionOpenAI),
		WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
		WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
		WithContext[*ChatClient](ctx),
		WithHTTPClient[*ChatClient](httpClient),
		WithLogger[*ChatClient](logger),
		WithHTTPMethod[*ChatClient](httpMethod),
		WithTemperature(temperature),
		WithMessages(messages),
		WithMaxTokens(maxTokens),
	)
	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if chat.base.provider != ProviderOpenAI {
		t.Errorf("provider: want %v, got %v", ProviderOpenAI, chat.base.provider)
	}
	if chat.base.url != URLChatCompletionOpenAI {
		t.Errorf("url: want %v, got %v", URLChatCompletionOpenAI, chat.base.url)
	}
	if chat.base.apiKey != EnvOpenAIAPIKey {
		t.Errorf("api key: want %v, got %v", EnvOpenAIAPIKey, chat.base.apiKey)
	}
	if chat.base.Model != AIModelCostOptimizedOpenAI {
		t.Errorf("model: want %v, got %v", AIModelCostOptimizedOpenAI, chat.base.Model)
	}
	testCases := []struct {
		role    Role
		content string
	}{
		{role: RoleSystem, content: "system"},
		{role: RoleUser, content: "say test"},
		{role: RoleAssistant, content: "test"},
	}
	for i, tc := range testCases {
		if chat.Messages[i].Role != tc.role {
			t.Errorf("role: want %v, got %v", tc.role, chat.Messages[i].Role)
		}
		if chat.Messages[i].Content != tc.content {
			t.Errorf("content: want %v, got %v", tc.content, chat.Messages[i].Content)
		}
	}
	if chat.MaxTokens != maxTokens {
		t.Errorf("max tokens: want %v, got %v", maxTokens, chat.MaxTokens)
	}
	if chat.base.ctx != ctx {
		t.Errorf("ctx: want %v, got %v", ctx, chat.base.ctx)
	}
	if chat.base.httpClient != httpClient {
		t.Errorf("http client: want %v, got %v", httpClient, chat.base)
	}
	if chat.base.logger != logger {
		t.Errorf("logger: want %v, got %v", logger, chat.base.logger)
	}
	if chat.base.httpMethod != httpMethod {
		t.Errorf("http method: want %v, got %v", httpMethod, chat.base.httpMethod)
	}
	if chat.base.httpClient.Transport == nil {
		t.Errorf("transport chain: want it set, got nil")
	}
}

func TestChatClientNew_Apply_Defaults(t *testing.T) {
	chat, err := NewChatClient(
		WithContext[*ChatClient](context.Background()),
		WithProvider[*ChatClient](ProviderOpenAI),
		WithURL[*ChatClient](URLChatCompletionOpenAI),
		WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
		WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
	)
	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	testCases := map[string]any{
		"logger":      chat.base.logger,
		"http client": chat.base.httpClient,
		"http method": chat.base.httpMethod,
		"max tokens":  chat.MaxTokens,
	}
	for field, value := range testCases {
		if value == nil {
			t.Errorf("%s: want it set", field)
		}
	}
}

func TestChatClientNew_Set_ProviderFlag(t *testing.T) {
	testCases := []struct {
		name          string
		provider      Provider
		url           URL
		apiKey        APIKey
		aiModel       AIModel
		flagOpenAI    bool
		flagAnthropic bool
	}{
		{
			name:          "OpenAI",
			provider:      ProviderOpenAI,
			url:           URLChatCompletionOpenAI,
			apiKey:        EnvOpenAIAPIKey,
			aiModel:       AIModelCostOptimizedOpenAI,
			flagOpenAI:    true,
			flagAnthropic: false,
		},
		{
			name:          "Anthropic",
			provider:      ProviderAnthropic,
			url:           URLChatCompletionAnthropic,
			apiKey:        EnvAnthropicAPIKey,
			aiModel:       AIModelCostOptimizedAnthropic,
			flagOpenAI:    false,
			flagAnthropic: true,
		},
	}
	for _, tc := range testCases {
		chat, err := NewChatClient(
			WithContext[*ChatClient](context.Background()),
			WithProvider[*ChatClient](tc.provider),
			WithURL[*ChatClient](tc.url),
			WithAPIKey[*ChatClient](tc.apiKey),
			WithModel[*ChatClient](tc.aiModel),
		)
		if err != nil {
			t.Fatalf("want no error, got %v", err)
		}
		switch tc.provider {
		case ProviderOpenAI:
			if chat.UseOpenAI != tc.flagOpenAI {
				t.Errorf("openai provider: want flag")
			}
			if chat.UseAnthropic != tc.flagAnthropic {
				t.Errorf("anthropic provider: want no flag")
			}
		case ProviderAnthropic:
			if chat.UseOpenAI != tc.flagOpenAI {
				t.Errorf("openai provider: want no flag")
			}
			if chat.UseAnthropic != tc.flagAnthropic {
				t.Errorf("anthropic provider: want flag")
			}
		}
	}
}

func TestChatClientNew_Validate_Fields(t *testing.T) {
	testCases := []struct {
		name        string
		chatBuilder func() (*ChatClient, error)
		wantError   bool
	}{
		{
			name: "AllRequiredFieldsPass",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
				)
			},
			wantError: false,
		},
		{
			name: "AllRerequiredAndOptionalFieldsPass",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
					WithTemperature(Temperature(1)),
					WithMessages(Messages{Message{Role: RoleUser, Content: "test"}}),
					WithMaxTokens(MaxTokens(4096)),
					WithHTTPClient[*ChatClient](http.DefaultClient),
					WithHTTPMethod[*ChatClient](HTTPMethod(http.MethodPost)),
					WithLogger[*ChatClient](logger.New()),
				)
			},
			wantError: false,
		},
		{
			name: "RequireContext",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
				)
			},
			wantError: true,
		},
		{
			name: "RequireProvider",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
				)
			},
			wantError: true,
		},
		{
			name: "RequireURL",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
				)
			},
			wantError: true,
		},
		{
			name: "RequireAPIKey",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
				)
			},
			wantError: true,
		},
		{
			name: "RequireModel",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
				)
			},
			wantError: true,
		},
		{
			name: "InvalidProvider",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](Provider("invalid")),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
				)
			},
			wantError: true,
		},
		{
			name: "InvalidURL",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URL("invalid")),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
				)
			},
			wantError: true,
		},
		{
			name: "InvalidAPIKey",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](APIKey("NON_EXISTING_API_KEY")),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
				)
			},
			wantError: true,
		},
		{
			name: "InvalidModel",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModel("invalid")),
				)
			},
			wantError: true,
		},
		{
			name: "InvalidHTTPMethod",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithHTTPMethod[*ChatClient](HTTPMethod("INVALID")), // should fail
				)
			},
			wantError: true,
		},
		{
			name: "MaxTokensLt1",
			chatBuilder: func() (*ChatClient, error) {
				chat, _ := NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
				)
				// override default max tokens
				chat.MaxTokens = MaxTokens(-1)
				tweakErr := chat.MaxTokens.Validate()
				return chat, tweakErr
			},
			wantError: true,
		},
		{
			name: "OpenAITemperatureGt2",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
					// test
					WithTemperature(3),
				)
			},
			wantError: true,
		},
		{
			name: "OpenAITemperatureLt0",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
					// test
					WithTemperature(-1),
				)
			},
			wantError: true,
		},
		{
			name: "AnthropicTemperatureGt1",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderAnthropic),
					WithURL[*ChatClient](URLChatCompletionAnthropic),
					WithAPIKey[*ChatClient](EnvAnthropicAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedAnthropic),
					// test
					WithTemperature(2),
				)
			},
			wantError: true,
		},
		{
			name: "AnthropicTemperatureLt0",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderAnthropic),
					WithURL[*ChatClient](URLChatCompletionAnthropic),
					WithAPIKey[*ChatClient](EnvAnthropicAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedAnthropic),
					// test
					WithTemperature(-1),
				)
			},
			wantError: true,
		},
		{
			name: "MessagesEmptyContent",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderAnthropic),
					WithURL[*ChatClient](URLChatCompletionAnthropic),
					WithAPIKey[*ChatClient](EnvAnthropicAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedAnthropic),
					// test
					WithMessages(Messages{
						Message{
							Content: "",
							Role:    RoleSystem,
						},
						Message{
							Content: "test",
							Role:    RoleUser,
						},
					}),
				)
			},
			wantError: true,
		},
		{
			name: "MessagesInvalidRole",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderAnthropic),
					WithURL[*ChatClient](URLChatCompletionAnthropic),
					WithAPIKey[*ChatClient](EnvAnthropicAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedAnthropic),
					// test
					WithMessages(Messages{
						Message{
							Content: "system",
							Role:    RoleSystem,
						},
						Message{
							Content: "test",
							Role:    Role("invalid"),
						},
					}),
				)
			},
			wantError: true,
		},
		{
			name: "NilContext",
			chatBuilder: func() (*ChatClient, error) {
				chat, _ := NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
				)
				// override default context
				chat.base.ctx = nil
				return chat, chat.validate()
			},
			wantError: true,
		},
		{
			name: "NilLogger",
			chatBuilder: func() (*ChatClient, error) {
				chat, _ := NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
				)
				// override default logger
				chat.base.logger = nil
				return chat, chat.validate()
			},
			wantError: true,
		},
		{
			name: "NilHTTPClient",
			chatBuilder: func() (*ChatClient, error) {
				chat, _ := NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
				)
				// override default http client
				chat.base.httpClient = nil
				return chat, chat.validate()
			},
			wantError: true,
		},
		{
			name: "EnsureOneProviderOnly",
			chatBuilder: func() (*ChatClient, error) {
				chat, _ := NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
				)
				// override provider flags to create conflict
				chat.UseAnthropic = true
				chat.UseOpenAI = true
				return chat, chat.validate()
			},
			wantError: true,
		},
		{
			name: "ProviderOpenAIUnmatchingURL",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionAnthropic), // anthropic, want openai
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
				)
			},
			wantError: true,
		},
		{
			name: "ProviderAnthropicUnmatchingURL",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderAnthropic),
					WithURL[*ChatClient](URLChatCompletionOpenAI), // openai, want anthropic
					WithAPIKey[*ChatClient](EnvAnthropicAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedAnthropic),
				)
			},
			wantError: true,
		},
		{
			name: "UnsupportedOpenAIModel",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedAnthropic), // anthropic, want openai
				)
			},
			wantError: true,
		},
		{
			name: "UnsupportedAnthropicModel",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderAnthropic),
					WithURL[*ChatClient](URLChatCompletionAnthropic),
					WithAPIKey[*ChatClient](EnvAnthropicAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI), // openai, want anthropic
				)
			},
			wantError: true,
		},
		{
			name: "NoRoleSystemInMessagesForAnthropic",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderAnthropic),
					WithURL[*ChatClient](URLChatCompletionAnthropic),
					WithAPIKey[*ChatClient](EnvAnthropicAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedAnthropic),
					// test
					WithMessages(Messages{
						Message{
							Role:    RoleSystem, // fail
							Content: "test",
						},
					}),
				)

			},
			wantError: true,
		},
		{
			name: "EnsureOneSystemRoleForOpenAI",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
					// test
					WithMessages(Messages{
						Message{
							Role:    RoleSystem,
							Content: "test",
						},
						Message{
							Role:    RoleSystem, // fail
							Content: "test",
						},
					}),
				)

			},
			wantError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.chatBuilder()
			if tc.wantError && err == nil {
				t.Errorf("want error, got nil")
			}
			if !tc.wantError && err != nil {
				t.Errorf("want no error, got %v", err)
			}
		})
	}
}

type tripperware func(req *http.Request) (*http.Response, error)

func (t tripperware) RoundTrip(req *http.Request) (*http.Response, error) {
	return t(req)
}

func TestChatClient_Do(t *testing.T) {
	testCases := []struct {
		name            string
		chatBuilder     func() (*ChatClient, error)
		statusCode      int
		body            *bytes.Buffer
		want            string
		roundTripperErr error
		wantErr         bool
	}{
		{
			name: "SuccessAnthropic",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
				)
			},
			statusCode: 200,
			body:       bytes.NewBufferString(`{"choices":[{"message":{"content":"test"}}]}`),
			want:       "test",
			wantErr:    false,
		},
		{
			name: "SuccessOpenAI",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderAnthropic),
					WithURL[*ChatClient](URLChatCompletionAnthropic),
					WithAPIKey[*ChatClient](EnvAnthropicAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedAnthropic),
				)
			},
			statusCode: 200,
			body:       bytes.NewBufferString(`{"content":[{"text": "test"}]}`),
			want:       "test",
			wantErr:    false,
		},
		{
			name: "NoContentFinalCompletion",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
				)
			},
			statusCode: 200,
			body:       bytes.NewBufferString(`{"choices":[{"message":{"content":""}}]}`), // empty content
			want:       "",                                                                // empty string
			wantErr:    false,
		},
		{
			name: "MalformedURL",
			chatBuilder: func() (*ChatClient, error) {
				chat, _ := NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
				)
				// override chat.url with malformed url
				chat.base.url = "::::"
				return chat, nil
			},
			body:    bytes.NewBufferString(""),
			wantErr: true,
		},
		{
			name: "NetworkFailed",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
				)
			},
			body:            bytes.NewBufferString(""),
			roundTripperErr: errors.New("network error"),
			wantErr:         true,
		},
		{
			name: "StatusNotOKOpenAI",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
				)
			},
			statusCode: 401,
			body: bytes.NewBufferString(
				`{"error": {"message": "incorrect api key provided", "type": "invalid_api_key"}}`,
			),
			wantErr: true,
		},
		{
			name: "StatusNotOKAnthropic",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderAnthropic),
					WithURL[*ChatClient](URLChatCompletionAnthropic),
					WithAPIKey[*ChatClient](EnvAnthropicAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedAnthropic),
				)
			},
			statusCode: 401,
			body: bytes.NewBufferString(
				`{"error": {"message": "incorrect api key provided", "type": "invalid_api_key"}}`,
			),
			wantErr: true,
		},
		{
			name: "MalformedResponseBody",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
				)
			},
			body:    bytes.NewBufferString("{]]invalid[[}"),
			wantErr: true,
		},
		{
			name: "NoChoicesOpenAIPayload",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
				)
			},
			statusCode: 200,
			body:       bytes.NewBufferString(`{"choices":[]}`),
			wantErr:    true,
		},
		{
			name: "NoContentAnthropicResponseBody",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderAnthropic),
					WithURL[*ChatClient](URLChatCompletionAnthropic),
					WithAPIKey[*ChatClient](EnvAnthropicAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedAnthropic),
				)
			},
			statusCode: 200,
			body:       bytes.NewBufferString(`{"content":[]}`),
			wantErr:    true,
		},
		{
			name: "MalformedOpenAIStatusOKResponseBody",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
				)
			},
			statusCode: 200,
			body:       bytes.NewBufferString(`{"malformed:""}`),
			wantErr:    true,
		},
		{
			name: "MalformedAnthropicStatusOKResponseBody",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderAnthropic),
					WithURL[*ChatClient](URLChatCompletionAnthropic),
					WithAPIKey[*ChatClient](EnvAnthropicAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedAnthropic),
				)
			},
			statusCode: 200,
			body:       bytes.NewBufferString(`{"malformed:""}`),
			wantErr:    true,
		},
		{
			name: "MalformedOpenAIStatusNotOKResponseBody",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderOpenAI),
					WithURL[*ChatClient](URLChatCompletionOpenAI),
					WithAPIKey[*ChatClient](EnvOpenAIAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedOpenAI),
				)
			},
			statusCode: 401,
			body:       bytes.NewBufferString(`{"malformed:""}`),
			wantErr:    true,
		},
		{
			name: "MalformedAnthropicStatusNotOKResponseBody",
			chatBuilder: func() (*ChatClient, error) {
				return NewChatClient(
					WithContext[*ChatClient](context.Background()),
					WithProvider[*ChatClient](ProviderAnthropic),
					WithURL[*ChatClient](URLChatCompletionAnthropic),
					WithAPIKey[*ChatClient](EnvAnthropicAPIKey),
					WithModel[*ChatClient](AIModelCostOptimizedAnthropic),
				)
			},
			statusCode: 401,
			body:       bytes.NewBufferString(`{"malformed:""}`),
			wantErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			chat, _ := tc.chatBuilder()
			// mock response
			chat.base.httpClient.Transport = tripperware(func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.statusCode, Body: io.NopCloser(tc.body)}, tc.roundTripperErr
			})

			completion, err := chat.Do()
			switch {
			case tc.wantErr:
				if err == nil {
					t.Fatal("want error, got nil")
				}
			default:
				if err != nil {
					t.Fatalf("want no error, got %v", err)
				}
				if completion.String() != tc.want {
					t.Errorf("want %q, got %q", tc.want, completion.String())
				}
			}
		})
	}
}
