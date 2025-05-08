package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/alnah/fla/ai"
	"github.com/alnah/fla/ai/breaker"
	"github.com/alnah/fla/ai/retrier"
	"github.com/alnah/fla/clog"
)

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

// ClassifyStatus wraps a RoundTripper, turning 429, 409, 423 or 5xx
// into an HTTPError built by parseErrorResponse.
func ClassifyStatus(provider ai.Provider) transport {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripper(func(req *http.Request) (*http.Response, error) {
			res, err := next.RoundTrip(req)
			if err != nil {
				return nil, err
			}
			if res.StatusCode == 429 || res.StatusCode == 409 ||
				res.StatusCode == 423 || res.StatusCode >= 500 {
				return nil, parseErrorResponse(provider, res)
			}
			return res, nil
		})
	}
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

// HTTPError represents an HTTP response classified as an error, carrying status,
// API error details, and any Retry-After hint.
type HTTPError struct {
	Provider   ai.Provider   // e.g. ai.ProviderOpenAI, ProviderAnthropic, ProviderElevenLabs
	StatusCode int           // HTTP status code
	Type       string        // error type from the payload
	Message    string        // error message from the payload
	RetryAfter time.Duration // parsed from Retry-After header, if any
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf(
		"HTTP %d [%s] %s: %s",
		e.StatusCode, e.Provider, e.Type, e.Message,
	)
}

// parseErrorResponse reads and restores the body, then extracts
// type/message according to the given provider.
func parseErrorResponse(provider ai.Provider, res *http.Response) *HTTPError {
	body, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	res.Body = io.NopCloser(bytes.NewReader(body))

	var typ, msg string
	switch provider {
	case ai.ProviderOpenAI, ai.ProviderAnthropic:
		// {"error": { "type": "...", "message": "..." }}
		var d struct {
			Error struct {
				Type    string `json:"type"`
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.Unmarshal(body, &d)
		typ = d.Error.Type
		msg = d.Error.Message

	case ai.ProviderElevenLabs:
		// {"detail": { "status": "...", "message": "..." }}
		var d struct {
			Detail struct {
				Status  string `json:"status"`
				Message string `json:"message"`
			} `json:"detail"`
		}
		_ = json.Unmarshal(body, &d)
		typ = d.Detail.Status
		msg = d.Detail.Message
	}

	// parse Retry-After header (seconds or HTTP-date)
	var ra time.Duration
	if h := res.Header.Get("Retry-After"); h != "" {
		if sec, err := strconv.Atoi(h); err == nil {
			ra = time.Duration(sec) * time.Second
		} else if t, err := http.ParseTime(h); err == nil {
			ra = time.Until(t)
		}
	}

	return &HTTPError{
		Provider:   provider,
		StatusCode: res.StatusCode,
		Type:       typ,
		Message:    msg,
		RetryAfter: ra,
	}
}

type ErrType string

func NewRetryClassifier(retryable map[ErrType]struct{}) func(error) bool {
	return func(err error) bool {
		// HTTP classifier
		var e *HTTPError
		if errors.As(err, &e) {
			_, ok := retryable[ErrType(e.Type)]
			return ok
		}

		// context
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return false
		}

		// network
		var n net.Error
		if errors.As(err, &n) {
			return n.Timeout()
		}
		return false
	}
}
