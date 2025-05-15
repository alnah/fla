package aiclient

import (
	"net/http"

	"github.com/alnah/fla/breaker"
	"github.com/alnah/fla/retrier"
	"github.com/alnah/fla/tripper"
)

const anthropicVersion string = "2023-06-01"

func (c *Chat) newTransportChain() http.RoundTripper {
	return tripper.Chain(
		tripper.Default(c.base.httpClient.Transport),
		tripper.AddHeader("Content-Type", "application/json"),
		func(next http.RoundTripper) http.RoundTripper {
			if c.base.provider == ProviderOpenAI {
				return tripper.AddHeader("Authorization", "Bearer "+c.base.apiKey.GetEnv())(next)
			}
			return tripper.AddHeader("x-api-key", c.base.apiKey.GetEnv())(next)
		},
		func(next http.RoundTripper) http.RoundTripper {
			if c.useAnthropic {
				return tripper.AddHeader("anthropic-version", anthropicVersion)(next)
			}
			return tripper.AddHeader("key", "")(next)
		},
		tripper.AddHeader("User-Agent", "Fla/1.0"),
		tripper.UseStatusClassifier(
			func(code int) bool { return code == 429 || code >= 500 },
			func(res *http.Response) error {
				if c.useOpenAI {
					return buildOpenaiError(res)
				}
				return buildAnthropicError(res)
			},
		),
		tripper.UseCircuitBreaker(breaker.New()),
		tripper.UseRetrier(retrier.New(), isRetryable),
		tripper.UseLogger(c.base.logger),
	)
}
