package aiclient

import (
	"net/http"

	"github.com/alnah/fla/breaker"
	"github.com/alnah/fla/retrier"
	"github.com/alnah/fla/transport"
)

func (s *STTClient) newTransportChain() http.RoundTripper {
	return transport.Chain(
		s.base.httpClient.Transport,
		transport.AddHeader("Content-Type", s.contentType),
		s.addAuthHeader(),
		transport.AddHeader("User-Agent", "Fla/1.0"),
		transport.UseStatusClassifier(func(sc int) bool { return sc == 429 || sc >= 500 }, s.buildError()),
		transport.UseCircuitBreaker(breaker.New()),
		transport.UseRetrier(retrier.New(), isRetryable),
		transport.UseLogger(s.base.log),
	)
}

func (s *STTClient) addAuthHeader() transport.Middleware {
	if s.base.provider == ProviderOpenAI {
		return transport.AddHeader("Authorization", "Bearer "+s.base.apiKey.GetEnv())
	}
	return transport.AddHeader("xi-api-key", s.base.apiKey.GetEnv())

}

func (s *STTClient) buildError() transport.ErrorFactoryFunc {
	if s.useOpenAI {
		return buildOpenaiError
	}
	return buildElevenlabsError
}
