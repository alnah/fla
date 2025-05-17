package aiclient

import (
	"net/http"

	"github.com/alnah/fla/breaker"
	"github.com/alnah/fla/retrier"
	"github.com/alnah/fla/tripper"
)

const anthropicVersion string = "2023-06-01"

func (c *ChatClient) newTransportChain() http.RoundTripper {
	return tripper.Chain(
		tripper.Default(c.base.httpClient.Transport),
		tripper.AddHeader("Content-Type", "application/json"),
		c.addAuthHeader(),
		c.addSpecHeader(),
		tripper.AddHeader("User-Agent", "Fla/1.0"),
		tripper.UseStatusClassifier(func(sc int) bool { return sc == 429 || sc >= 500 }, c.buildError()),
		tripper.UseCircuitBreaker(breaker.New()),
		tripper.UseRetrier(retrier.New(), isRetryable),
		tripper.UseLogger(c.base.log),
	)
}

func (c *ChatClient) addAuthHeader() tripper.Tripperware {
	if c.base.provider == ProviderOpenAI {
		return tripper.AddHeader("Authorization", "Bearer "+c.base.apiKey.GetEnv())
	}
	return tripper.AddHeader("x-api-key", c.base.apiKey.GetEnv())
}

func (c *ChatClient) addSpecHeader() tripper.Tripperware {
	if c.useAnthropic {
		return tripper.AddHeader("anthropic-version", anthropicVersion)
	}
	return tripper.AddHeader("key", "")
}

func (c *ChatClient) buildError() tripper.BuildError {
	if c.useOpenAI {
		return buildOpenaiError
	}
	return buildAnthropicError
}
