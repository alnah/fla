package transport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/alnah/fla/ai"
	"github.com/alnah/fla/ai/breaker"
	"github.com/alnah/fla/ai/clock"
	"github.com/alnah/fla/ai/retrier"
)

/********* Helpers *********/

type roundTripperTest func(*http.Request) (*http.Response, error)

func (f roundTripperTest) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type timeoutErr struct{}

func (timeoutErr) Error() string   { return "timeout error" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return false }

type temporaryErr struct{}

func (temporaryErr) Error() string   { return "temporary error" }
func (temporaryErr) Timeout() bool   { return false }
func (temporaryErr) Temporary() bool { return true }

/********* Tests *********/

func TestAddHeader(t *testing.T) {
	next := roundTripperTest(func(req *http.Request) (*http.Response, error) {
		if got := req.Header.Get("X-Test"); got != "value" {
			t.Errorf("expected header X-Test=value, got %q", got)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBuffer(nil))}, nil
	})

	rt := AddHeader("X-Test", "value")(next)
	_, _ = rt.RoundTrip(&http.Request{Header: http.Header{}})
}

func TestChain_AddsMultipleHeaders(t *testing.T) {
	next := roundTripperTest(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("A") != "1" || req.Header.Get("B") != "2" {
			t.Errorf("headers not chained: A=%q, B=%q", req.Header.Get("A"), req.Header.Get("B"))
		}
		return &http.Response{StatusCode: 204, Body: io.NopCloser(bytes.NewBuffer(nil))}, nil
	})

	rt := Chain(
		next,
		AddHeader("A", "1"),
		AddHeader("B", "2"),
	)
	_, _ = rt.RoundTrip(&http.Request{Header: http.Header{}})
}

func TestHTTPError_ErrorString(t *testing.T) {
	e := &HTTPError{
		Provider:   ai.ProviderOpenAI,
		StatusCode: 418,
		Type:       "teapot",
		Message:    "I'm a teapot",
	}
	expected := "HTTP 418 [openai] teapot: I'm a teapot"
	if got := e.Error(); got != expected {
		t.Errorf("Error() = %q, want %q", got, expected)
	}
}

func TestClassifyStatus_PassThrough(t *testing.T) {
	next := roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString("ok")),
		}, nil
	})

	rt := ClassifyStatus(ai.ProviderAnthropic)(next)
	res, err := rt.RoundTrip(&http.Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := io.ReadAll(res.Body)
	if got := string(data); got != "ok" {
		t.Errorf("unexpected body: %q", got)
	}
}

func TestClassifyStatus_ErrorWithRetryAfterSeconds(t *testing.T) {
	// simulate {"error":{"type":"rate_limit","message":"too many"}}
	payload := []byte(`{"error":{"type":"rate_limit","message":"too many"}}`)
	next := roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 429,
			Header:     http.Header{"Retry-After": {"5"}},
			Body:       io.NopCloser(bytes.NewReader(payload)),
		}, nil
	})

	rt := ClassifyStatus(ai.ProviderOpenAI)(next)
	_, err := rt.RoundTrip(&http.Request{})

	he, ok := err.(*HTTPError)
	if !ok {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if he.Provider != ai.ProviderOpenAI {
		t.Errorf("Provider = %v, want %v", he.Provider, ai.ProviderOpenAI)
	}
	if he.StatusCode != 429 {
		t.Errorf("StatusCode = %d, want 429", he.StatusCode)
	}
	if he.Type != "rate_limit" {
		t.Errorf("Type = %q, want %q", he.Type, "rate_limit")
	}
	if he.Message != "too many" {
		t.Errorf("Message = %q, want %q", he.Message, "too many")
	}
	if he.RetryAfter != 5*time.Second {
		t.Errorf("RetryAfter = %v, want 5s", he.RetryAfter)
	}
}

func TestClassifyStatus_ErrorWithRetryAfterHTTPDate(t *testing.T) {
	// future HTTP-date
	future := time.Now().Add(3 * time.Second).UTC()
	httpDate := future.Format(http.TimeFormat)
	payload := []byte(`{"error":{"type":"server_error","message":"oops"}}`)

	next := roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 500,
			Header:     http.Header{"Retry-After": {httpDate}},
			Body:       io.NopCloser(bytes.NewReader(payload)),
		}, nil
	})

	rt := ClassifyStatus(ai.ProviderAnthropic)(next)
	_, err := rt.RoundTrip(&http.Request{})

	he, ok := err.(*HTTPError)
	if !ok {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if he.Provider != ai.ProviderAnthropic {
		t.Errorf("Provider = %v, want %v", he.Provider, ai.ProviderAnthropic)
	}
	ra := he.RetryAfter
	if ra < 2*time.Second || ra > 4*time.Second {
		t.Errorf("RetryAfter = %v, want roughly 3s", ra)
	}
}

func TestUseCircuitBreaker_TripsAndBlocks(t *testing.T) {
	next := roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("downstream fail")
	})
	br := breaker.New(
		breaker.WithFailureThreshold(1),
		breaker.WithOpenTimeout(time.Minute),
		breaker.WithClock(clock.NewFakeClock(time.Now())),
	)

	rt := UseCircuitBreaker(br)(next)
	// first call fails
	_, err1 := rt.RoundTrip(&http.Request{})
	if err1 == nil || err1.Error() != "downstream fail" {
		t.Errorf("expected underlying error, got %v", err1)
	}
	// circuit now open
	_, err2 := rt.RoundTrip(&http.Request{})
	if err2 != breaker.ErrOpen {
		t.Errorf("expected ErrOpen, got %v", err2)
	}
}

func TestUseRetrier_RetriesOnError(t *testing.T) {
	attempts := 0
	next := roundTripperTest(func(req *http.Request) (*http.Response, error) {
		attempts++
		if attempts < 3 {
			return nil, errors.New("transient")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBufferString("done"))}, nil
	})
	fc := clock.NewFakeClock(time.Now())

	r := retrier.New(
		retrier.WithMaxAttempts(5),
		retrier.WithBaseDelay(0),
		retrier.WithJitter(retrier.NoJitter),
		retrier.WithClock(fc),
	)

	rt := UseRetrier(r, func(err error) bool { return true })(next)
	resp, err := rt.RoundTrip(&http.Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := io.ReadAll(resp.Body)
	if string(data) != "done" || attempts != 3 {
		t.Errorf("got %q after %d attempts, want \"done\" and 3", data, attempts)
	}
}

func TestUseLogger_NilLogger_PassesThrough(t *testing.T) {
	orig := http.DefaultTransport
	rt := UseLogger(nil)(orig)
	if rt != orig {
		t.Error("UseLogger(nil) should return the original RoundTripper")
	}
}

func TestNewRetryClassifier(t *testing.T) {
	tests := []struct {
		name      string
		retryable map[ErrType]struct{}
		err       error
		want      bool
	}{
		{"HTTPError retryable", map[ErrType]struct{}{"foo": {}}, &HTTPError{Type: "foo"}, true},
		{"HTTPError not retryable", map[ErrType]struct{}{"bar": {}}, &HTTPError{Type: "foo"}, false},
		{"wrapped HTTPError", map[ErrType]struct{}{"baz": {}}, fmt.Errorf("wrap: %w", &HTTPError{Type: "baz"}), true},
		{"context.Canceled", map[ErrType]struct{}{"foo": {}}, context.Canceled, false},
		{"context.DeadlineExceeded", map[ErrType]struct{}{"foo": {}}, context.DeadlineExceeded, false},
		{"net timeout error", nil, timeoutErr{}, true},
		{"net non-timeout", nil, temporaryErr{}, false},
		{"other error", map[ErrType]struct{}{"x": {}}, errors.New("x"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRetryClassifier(tt.retryable)(tt.err)
			if got != tt.want {
				t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

/********* Benchmark *********/

func BenchmarkChain_NoMiddleware(b *testing.B) {
	rt := Chain(http.DefaultTransport)
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rt.RoundTrip(req)
	}
}
