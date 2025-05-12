package tripper

import (
	"context"
	"net/http"
	"time"

	ai "github.com/alnah/fla/aiclient"
	"github.com/alnah/fla/aiclient/breaker"
	"github.com/alnah/fla/aiclient/retrier"
	"github.com/alnah/fla/logger"
)

type tripperware func(next http.RoundTripper) http.RoundTripper

type roundTripper func(*http.Request) (*http.Response, error)

func (f roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// Chain composes multiple transport middlewares around a base RoundTripper.
// This makes it easy to build reusable pipelines of behavior.
func Chain(rt http.RoundTripper, transports ...tripperware) http.RoundTripper {
	for _, m := range transports {
		rt = m(rt)
	}
	return rt
}

// AddHeader ensures every outgoing request carries the specified header.
// Useful for injecting authentication, content-type, or custom metadata.
func AddHeader(key, value string) tripperware {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripper(func(req *http.Request) (*http.Response, error) {
			clone := req.Clone(req.Context())
			clone.Header.Set(key, value)
			return next.RoundTrip(clone)
		})
	}
}

// UseStatusClassifier wraps a RoundTripper, turning 429, 409, 423 or 5xx
// into an HTTPError built by parseErrorResponse.
func UseStatusClassifier(provider ai.Provider) tripperware {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripper(func(req *http.Request) (*http.Response, error) {
			res, err := next.RoundTrip(req)
			if err != nil {
				return nil, err
			}
			if res.StatusCode == http.StatusTooManyRequests ||
				res.StatusCode == http.StatusConflict ||
				res.StatusCode == http.StatusLocked ||
				res.StatusCode >= http.StatusInternalServerError {
				return nil, ai.NewHTTPError(provider, res)
			}
			return res, nil
		})
	}
}

// UseCircuitBreaker wraps each request in a circuit breaker,
// preventing calls when the downstream service is consistently failing.
func UseCircuitBreaker(b *breaker.Breaker) tripperware {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripper(func(req *http.Request) (*http.Response, error) {
			var res *http.Response
			err := b.Execute(req.Context(), func(ctx context.Context) error {
				var e error
				clone := req.Clone(ctx)
				res, e = next.RoundTrip(clone)
				return e
			})
			return res, err
		})
	}
}

// UseRetrier applies a retry policy around each request,
// using the provided Retrier to handle transient errors according to rules.
func UseRetrier(r *retrier.Retrier, isRetryable func(error) bool) tripperware {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripper(func(req *http.Request) (*http.Response, error) {
			var res *http.Response
			err := r.Retry(req.Context(), func(ctx context.Context) error {
				var e error
				clone := req.Clone(ctx)
				res, e = next.RoundTrip(clone)
				return e
			}, isRetryable)
			return res, err
		})
	}
}

// UseLogger records timing and outcome of each HTTP request,
// aiding in metrics and debugging without altering business logic.
func UseLogger(log *logger.Logger) tripperware {
	if log == nil {
		return func(next http.RoundTripper) http.RoundTripper { return next }
	}

	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripper(func(req *http.Request) (*http.Response, error) {
			start := time.Now()
			res, err := next.RoundTrip(req)

			elapsedMs := time.Since(start).Milliseconds()

			if err != nil {
				log.Error(
					"transport",
					"method", req.Method,
					"url", req.URL.String(),
					"duration", elapsedMs,
					"error", err.Error(),
				)
				return res, err
			}

			log.Info(
				"transport",
				"method", req.Method,
				"url", req.URL.String(),
				"duration_ms", elapsedMs,
			)

			return res, nil
		})
	}
}
