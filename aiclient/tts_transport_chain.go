package aiclient

import (
	"net/http"

	"github.com/alnah/fla/breaker"
	"github.com/alnah/fla/retrier"
	"github.com/alnah/fla/transport"
)

func (t *TTSClient) newTransportChain() http.RoundTripper {
	return transport.Chain(
		t.base.httpClient.Transport,
		transport.AddHeader("Content-Type", "application/json"),
		t.addAuthHeader(),
		transport.AddHeader("User-Agent", "Fla/1.0"),
		transport.UseStatusClassifier(func(sc int) bool { return sc == 429 || sc >= 500 }, t.buildError()),
		transport.UseCircuitBreaker(breaker.New(breaker.ThirdPartyConfig())),
		transport.UseRetrier(retrier.NewExpBackoffJitter(), isRetryable),
		transport.UseLogger(t.base.log),
	)
}

func (t *TTSClient) addAuthHeader() transport.Middleware {
	if t.base.provider == ProviderOpenAI {
		return transport.AddHeader("Authorization", "Bearer "+t.base.apiKey.GetEnv())
	}
	return transport.AddHeader("xi-api-key", t.base.apiKey.GetEnv())
}

func (t *TTSClient) buildError() transport.ErrorFactoryFunc {
	if t.useOpenAI {
		return buildOpenaiError
	}
	return buildElevenlabsError

}
