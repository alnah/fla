package aiclient

import (
	"net/http"

	"github.com/alnah/fla/breaker"
	"github.com/alnah/fla/retrier"
	"github.com/alnah/fla/transport"
)

const anthropicVersion string = "2023-06-01"

func (c *ChatClient) newTransportChain() http.RoundTripper {
	return transport.Chain(
		c.base.httpClient.Transport,
		transport.AddHeader("Content-Type", "application/json"),
		c.addAuthHeader(),
		c.addSpecHeader(),
		transport.AddHeader("User-Agent", "Fla/1.0"),
		transport.UseStatusClassifier(func(sc int) bool { return sc == 429 || sc >= 500 }, c.buildError()),
		transport.UseCircuitBreaker(breaker.New(breaker.ThirdPartyConfig())),
		transport.UseRetrier(retrier.New(), isRetryable),
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
