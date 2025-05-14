package aiclient

import (
	"context"
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
		WithProvider(ProviderOpenAI),
		WithURL(URLChatCompletionOpenAI),
		WithAPIKey(EnvOpenAIAPIKey),
		WithModel(AIModelCostOptimizedOpenAI),
		WithTemperature(temperature),
		WithMessages(messages),
		WithMaxTokens(maxTokens),
		WithContext(ctx),
		WithHTTPClient(httpClient),
		WithLogger(logger),
		WithHTTPMethod(httpMethod),
	)
	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if chat.provider != ProviderOpenAI {
		t.Errorf("provider: want %v, got %v", ProviderOpenAI, chat.provider)
	}
	if chat.url != URLChatCompletionOpenAI {
		t.Errorf("url: want %v, got %v", URLChatCompletionOpenAI, chat.url)
	}
	if chat.apiKey != EnvOpenAIAPIKey {
		t.Errorf("api key: want %v, got %v", EnvOpenAIAPIKey, chat.apiKey)
	}
	if chat.Model != AIModelCostOptimizedOpenAI {
		t.Errorf("model: want %v, got %v", AIModelCostOptimizedOpenAI, chat.Model)
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
	if chat.ctx != ctx {
		t.Errorf("ctx: want %v, got %v", ctx, chat.ctx)
	}
	if chat.httpClient != httpClient {
		t.Errorf("http client: want %v, got %v", httpClient, chat.httpClient)
	}
	if chat.logger != logger {
		t.Errorf("logger: want %v, got %v", logger, chat.logger)
	}
	if chat.httpMethod != httpMethod {
		t.Errorf("http method: want %v, got %v", httpMethod, chat.httpMethod)
	}
}

func TestChatClientNew_Apply_Defaults(t *testing.T) {
	chat, err := NewChatClient(
		WithContext(context.Background()),
		WithProvider(ProviderOpenAI),
		WithURL(URLChatCompletionOpenAI),
		WithAPIKey(EnvOpenAIAPIKey),
		WithModel(AIModelCostOptimizedOpenAI),
	)
	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	testCases := map[string]any{
		"logger":      chat.logger,
		"http client": chat.httpClient,
		"http method": chat.httpMethod,
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
		model         AIModel
		flagOpenAI    bool
		flagAnthropic bool
	}{
		{
			name:          "OpenAI",
			provider:      ProviderOpenAI,
			url:           URLChatCompletionOpenAI,
			apiKey:        EnvOpenAIAPIKey,
			model:         AIModelCostOptimizedOpenAI,
			flagOpenAI:    true,
			flagAnthropic: false,
		},
		{
			name:          "Anthropic",
			provider:      ProviderAnthropic,
			url:           URLChatCompletionAnthropic,
			apiKey:        EnvAnthropicAPIKey,
			model:         AIModelCostOptimizedAnthropic,
			flagOpenAI:    false,
			flagAnthropic: true,
		},
	}
	for _, tc := range testCases {
		chat, err := NewChatClient(
			WithContext(context.Background()),
			WithProvider(tc.provider),
			WithURL(tc.url),
			WithAPIKey(tc.apiKey),
			WithModel(tc.model),
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
