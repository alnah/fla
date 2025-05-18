package transport

import (
	"context"
	"net/http"
	"time"

	"github.com/alnah/fla/breaker"
	"github.com/alnah/fla/logger"
	"github.com/alnah/fla/retrier"
)

// marker is a private type used solely for storing values in Go contexts.
// Defining a unique type prevents collisions with context keys from other packages.
type marker string

// key is the context key under which we inject a test marker in the circuit‐breaker middleware.
// By using a package‐local marker type, we ensure no other context user can accidentally overwrite or read this value.
const key marker = "key"

// Middleware represents a reusable piece of HTTP-client behavior.
// It lets you layer concerns—such as retries, logging, or headers—around any transport.
type Middleware func(next http.RoundTripper) http.RoundTripper

// RoundTripFunc adapts a plain function into an http.RoundTripper.
// Use it when you want a lightweight transport without defining a new struct.
type RoundTripFunc func(*http.Request) (*http.Response, error)

// ErrorClassifierFunc decides which HTTP status codes should be treated as errors.
// Implement this to define your own failure conditions.
type ErrorClassifierFunc func(code int) bool

// ErrorFactoryFunc builds an error from an HTTP response when a status code is classified as a failure.
// Use this to parse error payloads or include headers like Retry-After.
type ErrorFactoryFunc func(res *http.Response) error

func (t RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return t(req)
}

// Chain builds a layered HTTP transport by applying each Middleware in order.
// Use this to assemble a customized client pipeline—inject headers, retry failures,
// open circuit breakers, or log every call—without ever touching your business logic.
func Chain(rt http.RoundTripper, mws ...Middleware) http.RoundTripper {
	base := applyDefault(rt)
	for _, mw := range mws {
		base = mw(base)
	}
	return base
}

// applyDefault guarantees that you always have a non-nil transport.
// Pass nil to get the Go standard library’s http.DefaultTransport.
func applyDefault(rt http.RoundTripper) http.RoundTripper {
	if rt == nil {
		return http.DefaultTransport
	}
	return rt
}

// AddHeader ensures that every request passing through carries the given header.
// Handy for transparent authentication tokens, content-types, or any custom metadata.
func AddHeader(key, value string) Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			clone := req.Clone(req.Context())
			clone.Header.Set(key, value)
			return next.RoundTrip(clone)
		})
	}
}

// UseStatusClassifier converts selected HTTP status codes into errors.
// Ideal for turning 4xx/5xx payloads into typed errors by parsing the response body.
func UseStatusClassifier(shouldError ErrorClassifierFunc, buildError ErrorFactoryFunc) Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (*http.Response, error) {
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

// UseCircuitBreaker surrounds each request with a circuit breaker.
// It prevents calls to a failing service until it has recovered, protecting downstream stability.
func UseCircuitBreaker(b breaker.Breaker) Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			// inject context for testing
			ctx := context.WithValue(req.Context(), key, "injected")

			var res *http.Response
			err := b.Execute(ctx, func(innerCtx context.Context) error {
				// clone the request with the new context
				clone := req.Clone(innerCtx)
				var opErr error
				res, opErr = next.RoundTrip(clone)
				return opErr
			})

			return res, err
		})
	}
}

// UseRetrier retries transient errors according to your policy.
// Supply a Retrier implementation and a function to detect retryable errors.
func UseRetrier(r retrier.Retrier, isRetryable func(error) bool) Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (*http.Response, error) {
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

// UseLogger records timing and success or failure of each request.
// It’s designed to add minimal overhead while giving visibility into your HTTP traffic.
func UseLogger(log *logger.Logger) Middleware {
	if log == nil {
		return func(next http.RoundTripper) http.RoundTripper { return next }
	}

	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (*http.Response, error) {
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
