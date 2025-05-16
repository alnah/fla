package aiclient

import (
	"net/http"

	"github.com/alnah/fla/breaker"
	"github.com/alnah/fla/retrier"
	"github.com/alnah/fla/tripper"
)

func (t *STTClient) newTransportChain() http.RoundTripper {
	return tripper.Chain(
		tripper.Default(t.base.httpClient.Transport),
		tripper.AddHeader("Content-Type", t.contentType),
		func(next http.RoundTripper) http.RoundTripper {
			if t.base.provider == ProviderOpenAI {
				return tripper.AddHeader("Authorization", "Bearer "+t.base.apiKey.GetEnv())(next)
			}
			return tripper.AddHeader("xi-api-key", t.base.apiKey.GetEnv())(next)
		},
		tripper.AddHeader("User-Agent", "Fla/1.0"),
		tripper.UseStatusClassifier(
			func(code int) bool { return code == 429 || code >= 500 },
			func(res *http.Response) error {
				if t.useOpenAI {
					return buildOpenaiError(res)
				}
				return buildElevenlabsError(res)
			},
		),
		tripper.UseCircuitBreaker(breaker.New()),
		tripper.UseRetrier(retrier.New(), isRetryable),
		tripper.UseLogger(t.base.logger),
	)
}
