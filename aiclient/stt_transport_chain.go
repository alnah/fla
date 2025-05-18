package aiclient

import (
	"net/http"

	"github.com/alnah/fla/breaker"
	"github.com/alnah/fla/retrier"
	"github.com/alnah/fla/tripper"
)

func (s *STTClient) newTransportChain() http.RoundTripper {
	return tripper.Chain(
		s.base.httpClient.Transport,
		tripper.AddHeader("Content-Type", s.contentType),
		s.addAuthHeader(),
		tripper.AddHeader("User-Agent", "Fla/1.0"),
		tripper.UseStatusClassifier(func(sc int) bool { return sc == 429 || sc >= 500 }, s.buildError()),
		tripper.UseCircuitBreaker(breaker.New()),
		tripper.UseRetrier(retrier.New(), isRetryable),
		tripper.UseLogger(s.base.log),
	)
}

func (s *STTClient) addAuthHeader() tripper.Middleware {
	if s.base.provider == ProviderOpenAI {
		return tripper.AddHeader("Authorization", "Bearer "+s.base.apiKey.GetEnv())
	}
	return tripper.AddHeader("xi-api-key", s.base.apiKey.GetEnv())

}

func (s *STTClient) buildError() tripper.ErrorFactoryFunc {
	if s.useOpenAI {
		return buildOpenaiError
	}
	return buildElevenlabsError
}
