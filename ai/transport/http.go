package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/alnah/fla/ai"
	"github.com/alnah/fla/ai/breaker"
	"github.com/alnah/fla/ai/retrier"
	"github.com/alnah/fla/clog"
)

// HTTPError represents an HTTP response classified as an error,
// carrying status, API error details, and any Retry-After hint.
type HTTPError struct {
	Status     int           // HTTP status code
	Type       string        // API error type
	Code       string        // API error code
	Message    string        // API error message
	RetryAfter time.Duration // suggested wait before retrying
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf(
		"HTTPError: status=%d, type=%s, code=%s, message=%s",
		e.Status, e.Type, e.Code, e.Message,
	)
}

type transport func(next http.RoundTripper) http.RoundTripper

type roundTripper func(*http.Request) (*http.Response, error)

func (f roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// Chain composes multiple transport middlewares around a base RoundTripper.
// This makes it easy to build reusable pipelines of behavior.
func Chain(rt http.RoundTripper, transports ...transport) http.RoundTripper {
	for _, m := range transports {
		rt = m(rt)
	}
	return rt
}

// AddHeader ensures every outgoing request carries the specified header.
// Useful for injecting authentication, content-type, or custom metadata.
func AddHeader(key, value string) transport {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripper(func(req *http.Request) (*http.Response, error) {
			clone := req.Clone(req.Context())
			clone.Header.Set(key, value)
			return next.RoundTrip(clone)
		})
	}
}

// ClassifyStatus transforms 429 and 5xx responses into HTTPError,
// extracting any Retry-After hint to guide retry logic.
func ClassifyStatus(next http.RoundTripper) http.RoundTripper {
	return roundTripper(func(req *http.Request) (*http.Response, error) {
		res, err := next.RoundTrip(req)
		if err != nil {
			return nil, err
		}
		if res.StatusCode == 429 || res.StatusCode == 409 || res.StatusCode == 423 || res.StatusCode >= 500 {
			body, _ := io.ReadAll(res.Body)
			_ = res.Body.Close()
			res.Body = io.NopCloser(bytes.NewReader(body)) // allow higher layers to inspect

			var apiErr ai.APIError
			_ = json.Unmarshal(body, &apiErr)

			var ra time.Duration
			if h := res.Header.Get("Retry-After"); h != "" {
				// RFC 9110: either integer seconds or HTTP-date
				if sec, errConv := strconv.Atoi(h); errConv == nil {
					ra = time.Duration(sec) * time.Second
				} else if t, errTime := http.ParseTime(h); errTime == nil {
					ra = time.Until(t)
				}
			}

			return nil, &HTTPError{
				Status:     res.StatusCode,
				Type:       apiErr.Error.Type,
				Code:       apiErr.Error.Code,
				Message:    apiErr.Error.Message,
				RetryAfter: ra,
			}
		}
		return res, nil
	})
}

// UseCircuitBreaker wraps each request in a circuit breaker,
// preventing calls when the downstream service is consistently failing.
func UseCircuitBreaker(b *breaker.Breaker) transport {
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
func UseRetrier(r *retrier.Retrier, isRetryable func(error) bool) transport {
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
func UseLogger(log *clog.Logger) transport {
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
