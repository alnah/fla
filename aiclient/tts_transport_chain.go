package aiclient

import (
	"net/http"

	"github.com/alnah/fla/breaker"
	"github.com/alnah/fla/retrier"
	"github.com/alnah/fla/tripper"
)

func (t *TTSClient) newTransportChain() http.RoundTripper {
	return tripper.Chain(
		t.base.httpClient.Transport,
		tripper.AddHeader("Content-Type", "application/json"),
		t.addAuthHeader(),
		tripper.AddHeader("User-Agent", "Fla/1.0"),
		tripper.UseStatusClassifier(func(sc int) bool { return sc == 429 || sc >= 500 }, t.buildError()),
		tripper.UseCircuitBreaker(breaker.New()),
		tripper.UseRetrier(retrier.New(), isRetryable),
		tripper.UseLogger(t.base.log),
	)
}

func (t *TTSClient) addAuthHeader() tripper.Middleware {
	if t.base.provider == ProviderOpenAI {
		return tripper.AddHeader("Authorization", "Bearer "+t.base.apiKey.GetEnv())
	}
	return tripper.AddHeader("xi-api-key", t.base.apiKey.GetEnv())
}

func (t *TTSClient) buildError() tripper.ErrorFactoryFunc {
	if t.useOpenAI {
		return buildOpenaiError
	}
	return buildElevenlabsError

}
