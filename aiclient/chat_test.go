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
	httpMethod := httpMethod(http.MethodPost)
	maxTokens := MaxTokens(8192)
	chat, err := NewChat(
		WithProvider[*Chat](ProviderOpenAI),
		WithURL[*Chat](URLChatOpenAI),
		WithAPIKey[*Chat](EnvOpenAIAPIKey),
		WithModel[*Chat](AIModelCostOptimizedOpenAI),
		WithContext[*Chat](ctx),
		WithHTTPClient[*Chat](httpClient),
		WithLogger[*Chat](logger),
		WithHTTPMethod[*Chat](httpMethod),
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
	if chat.base.url != URLChatOpenAI {
		t.Errorf("url: want %v, got %v", URLChatOpenAI, chat.base.url)
	}
	if chat.base.apiKey != EnvOpenAIAPIKey {
		t.Errorf("api key: want %v, got %v", EnvOpenAIAPIKey, chat.base.apiKey)
	}
	if chat.base.Model != AIModelCostOptimizedOpenAI {
		t.Errorf("model: want %v, got %v", AIModelCostOptimizedOpenAI, chat.base.Model)
	}
	testCases := []struct {
		role    role
		content string
	}{
		{role: RoleSystem, content: "system"},
		{role: RoleUser, content: "say test"},
		{role: RoleAssistant, content: "test"},
	}
	for i, tc := range testCases {
		if chat.messages[i].Role != tc.role {
			t.Errorf("role: want %v, got %v", tc.role, chat.messages[i].Role)
		}
		if chat.messages[i].Content != tc.content {
			t.Errorf("content: want %v, got %v", tc.content, chat.messages[i].Content)
		}
	}
	if chat.maxTokens != maxTokens {
		t.Errorf("max tokens: want %v, got %v", maxTokens, chat.maxTokens)
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
	chat, err := NewChat(
		WithProvider[*Chat](ProviderOpenAI),
		WithURL[*Chat](URLChatOpenAI),
		WithAPIKey[*Chat](EnvOpenAIAPIKey),
		WithModel[*Chat](AIModelCostOptimizedOpenAI),
	)
	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	testCases := map[string]any{
		"ctx":         chat.base.ctx,
		"logger":      chat.base.logger,
		"http client": chat.base.httpClient,
		"http method": chat.base.httpMethod,
		"max tokens":  chat.maxTokens,
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
		provider      provider
		url           url
		apiKey        apiKey
		aiModel       aiModel
		flagOpenAI    bool
		flagAnthropic bool
	}{
		{
			name:          "OpenAI",
			provider:      ProviderOpenAI,
			url:           URLChatOpenAI,
			apiKey:        EnvOpenAIAPIKey,
			aiModel:       AIModelCostOptimizedOpenAI,
			flagOpenAI:    true,
			flagAnthropic: false,
		},
		{
			name:          "Anthropic",
			provider:      ProviderAnthropic,
			url:           URLChatAnthropic,
			apiKey:        EnvAnthropicAPIKey,
			aiModel:       AIModelCostOptimizedAnthropic,
			flagOpenAI:    false,
			flagAnthropic: true,
		},
	}
	for _, tc := range testCases {
		chat, err := NewChat(
			WithProvider[*Chat](tc.provider),
			WithURL[*Chat](tc.url),
			WithAPIKey[*Chat](tc.apiKey),
			WithModel[*Chat](tc.aiModel),
		)
		if err != nil {
			t.Fatalf("want no error, got %v", err)
		}
		switch tc.provider {
		case ProviderOpenAI:
			if chat.useOpenAI != tc.flagOpenAI {
				t.Errorf("openai provider: want flag")
			}
			if chat.useAnthropic != tc.flagAnthropic {
				t.Errorf("anthropic provider: want no flag")
			}
		case ProviderAnthropic:
			if chat.useOpenAI != tc.flagOpenAI {
				t.Errorf("openai provider: want no flag")
			}
			if chat.useAnthropic != tc.flagAnthropic {
				t.Errorf("anthropic provider: want flag")
			}
		}
	}
}

func TestChatClientNew_Validate_Fields(t *testing.T) {
	testCases := []struct {
		name        string
		chatBuilder func() (*Chat, error)
		wantError   bool
	}{
		{
			name: "AllRequiredFieldsPass",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
				)
			},
			wantError: false,
		},
		{
			name: "AllRerequiredAndOptionalFieldsPass",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithContext[*Chat](context.Background()),
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
					WithTemperature(Temperature(1)),
					WithMessages(Messages{Message{Role: RoleUser, Content: "test"}}),
					WithMaxTokens(MaxTokens(4096)),
					WithHTTPClient[*Chat](http.DefaultClient),
					WithHTTPMethod[*Chat](httpMethod(http.MethodPost)),
					WithLogger[*Chat](logger.New()),
				)
			},
			wantError: false,
		},
		{
			name: "RequireProvider",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
				)
			},
			wantError: true,
		},
		{
			name: "RequireURL",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
				)
			},
			wantError: true,
		},
		{
			name: "RequireAPIKey",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
				)
			},
			wantError: true,
		},
		{
			name: "RequireModel",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
				)
			},
			wantError: true,
		},
		{
			name: "InvalidProvider",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](provider("invalid")),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
				)
			},
			wantError: true,
		},
		{
			name: "InvalidURL",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](url("invalid")),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
				)
			},
			wantError: true,
		},
		{
			name: "InvalidAPIKey",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](apiKey("NON_EXISTING_API_KEY")),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
				)
			},
			wantError: true,
		},
		{
			name: "InvalidModel",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](aiModel("invalid")),
				)
			},
			wantError: true,
		},
		{
			name: "InvalidHTTPMethod",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithHTTPMethod[*Chat](httpMethod("INVALID")), // should fail
				)
			},
			wantError: true,
		},
		{
			name: "MaxTokensLt1",
			chatBuilder: func() (*Chat, error) {
				chat, _ := NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
				)
				// override default max tokens
				chat.maxTokens = MaxTokens(-1)
				tweakErr := chat.maxTokens.Validate()
				return chat, tweakErr
			},
			wantError: true,
		},
		{
			name: "OpenAITemperatureGt2",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
					// test
					WithTemperature(3),
				)
			},
			wantError: true,
		},
		{
			name: "OpenAITemperatureLt0",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
					// test
					WithTemperature(-1),
				)
			},
			wantError: true,
		},
		{
			name: "AnthropicTemperatureGt1",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderAnthropic),
					WithURL[*Chat](URLChatAnthropic),
					WithAPIKey[*Chat](EnvAnthropicAPIKey),
					WithModel[*Chat](AIModelCostOptimizedAnthropic),
					// test
					WithTemperature(2),
				)
			},
			wantError: true,
		},
		{
			name: "AnthropicTemperatureLt0",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderAnthropic),
					WithURL[*Chat](URLChatAnthropic),
					WithAPIKey[*Chat](EnvAnthropicAPIKey),
					WithModel[*Chat](AIModelCostOptimizedAnthropic),
					// test
					WithTemperature(-1),
				)
			},
			wantError: true,
		},
		{
			name: "MessagesEmptyContent",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderAnthropic),
					WithURL[*Chat](URLChatAnthropic),
					WithAPIKey[*Chat](EnvAnthropicAPIKey),
					WithModel[*Chat](AIModelCostOptimizedAnthropic),
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
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderAnthropic),
					WithURL[*Chat](URLChatAnthropic),
					WithAPIKey[*Chat](EnvAnthropicAPIKey),
					WithModel[*Chat](AIModelCostOptimizedAnthropic),
					// test
					WithMessages(Messages{
						Message{
							Content: "system",
							Role:    RoleSystem,
						},
						Message{
							Content: "test",
							Role:    role("invalid"),
						},
					}),
				)
			},
			wantError: true,
		},
		{
			name: "NilContext",
			chatBuilder: func() (*Chat, error) {
				chat, _ := NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
				)
				// override default context
				chat.base.ctx = nil
				return chat, chat.validate()
			},
			wantError: true,
		},
		{
			name: "NilLogger",
			chatBuilder: func() (*Chat, error) {
				chat, _ := NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
				)
				// override default logger
				chat.base.logger = nil
				return chat, chat.validate()
			},
			wantError: true,
		},
		{
			name: "NilHTTPClient",
			chatBuilder: func() (*Chat, error) {
				chat, _ := NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
				)
				// override default http client
				chat.base.httpClient = nil
				return chat, chat.validate()
			},
			wantError: true,
		},
		{
			name: "EnsureOneProviderOnly",
			chatBuilder: func() (*Chat, error) {
				chat, _ := NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
				)
				// override provider flags to create conflict
				chat.useAnthropic = true
				chat.useOpenAI = true
				return chat, chat.validate()
			},
			wantError: true,
		},
		{
			name: "ProviderOpenAIUnmatchingURL",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatAnthropic), // anthropic, want openai
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
				)
			},
			wantError: true,
		},
		{
			name: "ProviderAnthropicUnmatchingURL",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderAnthropic),
					WithURL[*Chat](URLChatOpenAI), // openai, want anthropic
					WithAPIKey[*Chat](EnvAnthropicAPIKey),
					WithModel[*Chat](AIModelCostOptimizedAnthropic),
				)
			},
			wantError: true,
		},
		{
			name: "UnsupportedOpenAIModel",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedAnthropic), // anthropic, want openai
				)
			},
			wantError: true,
		},
		{
			name: "UnsupportedAnthropicModel",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderAnthropic),
					WithURL[*Chat](URLChatAnthropic),
					WithAPIKey[*Chat](EnvAnthropicAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI), // openai, want anthropic
				)
			},
			wantError: true,
		},
		{
			name: "NoRoleSystemInMessagesForAnthropic",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderAnthropic),
					WithURL[*Chat](URLChatAnthropic),
					WithAPIKey[*Chat](EnvAnthropicAPIKey),
					WithModel[*Chat](AIModelCostOptimizedAnthropic),
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
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
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
		chatBuilder     func() (*Chat, error)
		statusCode      int
		body            *bytes.Buffer
		want            string
		roundTripperErr error
		wantErr         bool
	}{
		{
			name: "SuccessAnthropic",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
				)
			},
			statusCode: 200,
			body:       bytes.NewBufferString(`{"choices":[{"message":{"content":"test"}}]}`),
			want:       "test",
			wantErr:    false,
		},
		{
			name: "SuccessOpenAI",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderAnthropic),
					WithURL[*Chat](URLChatAnthropic),
					WithAPIKey[*Chat](EnvAnthropicAPIKey),
					WithModel[*Chat](AIModelCostOptimizedAnthropic),
				)
			},
			statusCode: 200,
			body:       bytes.NewBufferString(`{"content":[{"text": "test"}]}`),
			want:       "test",
			wantErr:    false,
		},
		{
			name: "NoContentFinalCompletion",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
				)
			},
			statusCode: 200,
			body:       bytes.NewBufferString(`{"choices":[{"message":{"content":""}}]}`), // empty content
			want:       "",                                                                // empty string
			wantErr:    false,
		},
		{
			name: "MalformedURL",
			chatBuilder: func() (*Chat, error) {
				chat, _ := NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
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
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
				)
			},
			body:            bytes.NewBufferString(""),
			roundTripperErr: errors.New("network error"),
			wantErr:         true,
		},
		{
			name: "StatusNotOKOpenAI",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
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
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderAnthropic),
					WithURL[*Chat](URLChatAnthropic),
					WithAPIKey[*Chat](EnvAnthropicAPIKey),
					WithModel[*Chat](AIModelCostOptimizedAnthropic),
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
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
				)
			},
			body:    bytes.NewBufferString("{]]invalid[[}"),
			wantErr: true,
		},
		{
			name: "NoChoicesOpenAIPayload",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
				)
			},
			statusCode: 200,
			body:       bytes.NewBufferString(`{"choices":[]}`),
			wantErr:    true,
		},
		{
			name: "NoContentAnthropicResponseBody",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderAnthropic),
					WithURL[*Chat](URLChatAnthropic),
					WithAPIKey[*Chat](EnvAnthropicAPIKey),
					WithModel[*Chat](AIModelCostOptimizedAnthropic),
				)
			},
			statusCode: 200,
			body:       bytes.NewBufferString(`{"content":[]}`),
			wantErr:    true,
		},
		{
			name: "MalformedOpenAIStatusOKResponseBody",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
				)
			},
			statusCode: 200,
			body:       bytes.NewBufferString(`{"malformed:""}`),
			wantErr:    true,
		},
		{
			name: "MalformedAnthropicStatusOKResponseBody",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderAnthropic),
					WithURL[*Chat](URLChatAnthropic),
					WithAPIKey[*Chat](EnvAnthropicAPIKey),
					WithModel[*Chat](AIModelCostOptimizedAnthropic),
				)
			},
			statusCode: 200,
			body:       bytes.NewBufferString(`{"malformed:""}`),
			wantErr:    true,
		},
		{
			name: "MalformedOpenAIStatusNotOKResponseBody",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderOpenAI),
					WithURL[*Chat](URLChatOpenAI),
					WithAPIKey[*Chat](EnvOpenAIAPIKey),
					WithModel[*Chat](AIModelCostOptimizedOpenAI),
				)
			},
			statusCode: 401,
			body:       bytes.NewBufferString(`{"malformed:""}`),
			wantErr:    true,
		},
		{
			name: "MalformedAnthropicStatusNotOKResponseBody",
			chatBuilder: func() (*Chat, error) {
				return NewChat(
					WithProvider[*Chat](ProviderAnthropic),
					WithURL[*Chat](URLChatAnthropic),
					WithAPIKey[*Chat](EnvAnthropicAPIKey),
					WithModel[*Chat](AIModelCostOptimizedAnthropic),
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
