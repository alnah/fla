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
		tripper.Default(c.hc.Transport),
		tripper.AddHeader("Content-Type", "application/json"),
		c.addAuthHeaders(),
		c.addProviderSpecHeaders(),
		tripper.UseStatusClassifier(func(sc int) bool { return sc == 429 || sc >= 500 }, c.buildError()),
		tripper.UseCircuitBreaker(breaker.New()),
		tripper.UseRetrier(retrier.New(), IsRetryable),
		tripper.UseLogger(c.logger),
	)
}

func (c *ChatClient) addAuthHeaders() tripper.Tripperware {
	switch c.provider {
	case ProviderOpenAI:
		return tripper.AddHeader("Authorization", fmt.Sprintf("Bearer %s", c.apiKey.GetEnv()))
	case ProviderAnthropic:
		return tripper.AddHeader("x-api-key", c.apiKey.GetEnv())
	}
	return nil
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
	case c.UseAnthropic:
		return BuildAnthropicError
	default:
		return tripper.BuildError(nil)
	}
}
