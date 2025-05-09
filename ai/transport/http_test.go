package transport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/alnah/fla/ai"
	"github.com/alnah/fla/ai/breaker"
	"github.com/alnah/fla/ai/clock"
	"github.com/alnah/fla/ai/retrier"
	"github.com/alnah/fla/clog"
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

type stubRT struct {
	resp *http.Response
	err  error
}

func (s *stubRT) RoundTrip(req *http.Request) (*http.Response, error) {
	return s.resp, s.err
}

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
	payload := []byte(`{"error":{"code":"rate_limit","message":"too many"}}`)
	next := roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 429,
			Header:     http.Header{"Retry-After": {"5"}},
			Body:       io.NopCloser(bytes.NewReader(payload)),
		}, nil
	})

	rt := ClassifyStatus(ai.ProviderOpenAI)(next)
	_, err := rt.RoundTrip(&http.Request{})

	he, ok := err.(*ai.HTTPError)
	if !ok {
		t.Fatalf("expected *ai.HTTPError, got %T: %v", err, err)
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

	he, ok := err.(*ai.HTTPError)
	if !ok {
		t.Fatalf("expected *ai.HTTPError, got %T: %v", err, err)
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
		retryable map[ai.ErrType]struct{}
		err       error
		want      bool
	}{
		{"ai.HTTPError retryable", map[ai.ErrType]struct{}{"foo": {}}, &ai.HTTPError{Type: "foo"}, true},
		{"ai.HTTPError not retryable", map[ai.ErrType]struct{}{"bar": {}}, &ai.HTTPError{Type: "foo"}, false},
		{"wrapped ai.HTTPError", map[ai.ErrType]struct{}{"baz": {}}, fmt.Errorf("wrap: %w", &ai.HTTPError{Type: "baz"}), true},
		{"context.Canceled", map[ai.ErrType]struct{}{"foo": {}}, context.Canceled, false},
		{"context.DeadlineExceeded", map[ai.ErrType]struct{}{"foo": {}}, context.DeadlineExceeded, false},
		{"net timeout error", nil, timeoutErr{}, true},
		{"net non-timeout", nil, temporaryErr{}, false},
		{"other error", map[ai.ErrType]struct{}{"x": {}}, errors.New("x"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ai.NewRetryClassifier(tt.retryable)(tt.err)
			if got != tt.want {
				t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestUseLogger(t *testing.T) {
	tests := []struct {
		name        string
		stubResp    *http.Response
		stubErr     error
		wantPattern string // regex to match in log output
	}{
		{
			name:        "success path logs Info with duration_ms",
			stubResp:    &http.Response{StatusCode: 200},
			stubErr:     nil,
			wantPattern: `transport.*method=GET.*url=http://example\.com/.*duration_ms=\d+`,
		},
		{
			name:        "error path logs Error with error message",
			stubResp:    nil,
			stubErr:     errors.New("network fail"),
			wantPattern: `transport.*method=GET.*url=http://example\.com/.*duration=\d+.*error="network fail"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare a buffer to capture log output
			var buf bytes.Buffer
			handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
				AddSource: false,
				Level:     slog.LevelDebug,
			})
			logger := clog.NewWithHandler(handler)

			// wrap the stub RoundTripper
			stub := &stubRT{resp: tt.stubResp, err: tt.stubErr}
			rt := UseLogger(logger)(stub)

			// issue a single GET request
			req, _ := http.NewRequest("GET", "http://example.com/", nil)
			start := time.Now()
			resp, err := rt.RoundTrip(req)
			elapsed := time.Since(start)

			// ensure behavior is preserved
			if resp != tt.stubResp {
				t.Errorf("got resp %v; want %v", resp, tt.stubResp)
			}
			if err != tt.stubErr {
				t.Errorf("got err %v; want %v", err, tt.stubErr)
			}

			out := buf.String()
			// verify timestamp-derived duration is sensible
			if elapsed < 0 {
				t.Errorf("elapsed time should be non-negative, got %v", elapsed)
			}

			// match against expected log pattern
			re := regexp.MustCompile(tt.wantPattern)
			if !re.MatchString(out) {
				t.Errorf("log output did not match.\nOutput: %s\nPattern: %s", out, tt.wantPattern)
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
