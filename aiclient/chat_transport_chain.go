package aiclient

import (
	"fmt"
	"net/http"

	"github.com/alnah/fla/breaker"
	"github.com/alnah/fla/retrier"
	"github.com/alnah/fla/tripper"
)

const AnthropicVersion string = "2023-06-01"

func (c *ChatClient) configureTransportChain() http.RoundTripper {
	return tripper.Chain(
		tripper.Default(c.base.httpClient.Transport),
		tripper.AddHeader("Content-Type", "application/json"),
		c.addAuthHeaders(),
		c.addProviderSpecHeaders(),
		tripper.UseStatusClassifier(func(sc int) bool { return sc == 429 || sc >= 500 }, c.buildError()),
		tripper.UseCircuitBreaker(breaker.New()),
		tripper.UseRetrier(retrier.New(), IsRetryable),
		tripper.UseLogger(c.base.logger),
	)
}

func (c *ChatClient) addAuthHeaders() tripper.Tripperware {
	switch c.base.provider {
	case ProviderOpenAI:
		return tripper.AddHeader("Authorization", fmt.Sprintf("Bearer %s", c.base.apiKey.GetEnv()))
	default:
		return tripper.AddHeader("x-api-key", c.base.apiKey.GetEnv())
	}
}

func (c *ChatClient) addProviderSpecHeaders() tripper.Tripperware {
	if c.UseAnthropic {
		return tripper.AddHeader("anthropic-version", AnthropicVersion)
	}
	return tripper.AddHeader("key", "")
}

func (c *ChatClient) buildError() tripper.BuildError {
	switch {
	case c.UseOpenAI:
		return BuildOpenAIError
	default:
		return BuildAnthropicError
	}
}
