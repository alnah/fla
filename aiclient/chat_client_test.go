package aiclient

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/alnah/fla/logger"
)

func TestChatClient_New_WithOptions(t *testing.T) {
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
