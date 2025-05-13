package tripper

import (
	"context"
	"net/http"
	"time"

	"github.com/alnah/fla/breaker"
	"github.com/alnah/fla/logger"
	"github.com/alnah/fla/retrier"
)

type Tripperware func(next http.RoundTripper) http.RoundTripper
type Tripper func(*http.Request) (*http.Response, error)
type ShouldError func(code int) bool
type BuildError func(res *http.Response) error

func (f Tripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// Default yields a non-nil RoundTripper.
func Default(rt http.RoundTripper) http.RoundTripper {
	if rt == nil {
		return http.DefaultTransport
	}
	return rt
}

// Chain composes multiple transport middlewares around a base RoundTripper.
// This makes it easy to build reusable pipelines of behavior.
func Chain(rt http.RoundTripper, transports ...Tripperware) http.RoundTripper {
	for _, m := range transports {
		rt = m(rt)
	}
	return rt
}

// AddHeader ensures every outgoing request carries the specified header.
// Useful for injecting authentication, content-type, or custom metadata.
func AddHeader(key, value string) Tripperware {
	return func(next http.RoundTripper) http.RoundTripper {
		return Tripper(func(req *http.Request) (*http.Response, error) {
			clone := req.Clone(req.Context())
			clone.Header.Set(key, value)
			return next.RoundTrip(clone)
		})
	}
}

// UseStatusClassifier turns responses whose status code matches shouldError
// into errors using the provided buildError function.
func UseStatusClassifier(shouldError ShouldError, buildError BuildError) Tripperware {
	return func(next http.RoundTripper) http.RoundTripper {
		return Tripper(func(req *http.Request) (*http.Response, error) {
			res, err := next.RoundTrip(req)
			if err != nil {
				return nil, err
			}
			if shouldError(res.StatusCode) {
				return nil, buildError(res)
			}
			return res, nil
		})
	}
}

// UseCircuitBreaker wraps each request in a circuit breaker,
// preventing calls when the downstream service is consistently failing.
func UseCircuitBreaker(b *breaker.Breaker) Tripperware {
	return func(next http.RoundTripper) http.RoundTripper {
		return Tripper(func(req *http.Request) (*http.Response, error) {
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
func UseRetrier(r *retrier.Retrier, isRetryable func(error) bool) Tripperware {
	return func(next http.RoundTripper) http.RoundTripper {
		return Tripper(func(req *http.Request) (*http.Response, error) {
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
func UseLogger(log *logger.Logger) Tripperware {
	if log == nil {
		return func(next http.RoundTripper) http.RoundTripper { return next }
	}

	return func(next http.RoundTripper) http.RoundTripper {
		return Tripper(func(req *http.Request) (*http.Response, error) {
			start := time.Now()
			res, err := next.RoundTrip(req)

			elapsedMs := time.Since(start).Milliseconds()

			if err != nil {
				log.Error(
					"transport",
					"method", req.Method,
					"url", req.URL.String(),
					"duration_ms", elapsedMs,
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
