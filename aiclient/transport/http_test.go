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

	ai "github.com/alnah/fla/aiclient"
	"github.com/alnah/fla/aiclient/breaker"
	"github.com/alnah/fla/aiclient/clock"
	"github.com/alnah/fla/aiclient/retrier"
	"github.com/alnah/fla/logger"
)

/********* Helpers *********/

type helperRoundTripper func(*http.Request) (*http.Response, error)

func (f helperRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
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

type stubRoundTripper struct {
	resp *http.Response
	err  error
}

func (s *stubRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return s.resp, s.err
}

/********* Unit Tests *********/

func TestChain_AddHeader_One(t *testing.T) {
	next := helperRoundTripper(func(req *http.Request) (*http.Response, error) {
		if got := req.Header.Get("x-test"); got != "value" {
			t.Errorf("header x-test: want %q, got %q", "value", got)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBuffer(nil))}, nil
	})

	rt := AddHeader("x-test", "value")(next)
	rt.RoundTrip(&http.Request{Header: http.Header{}})
}

func TestChain_AddHeader_Many(t *testing.T) {
	next := helperRoundTripper(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("A") != "1" || req.Header.Get("B") != "2" {
			t.Errorf("headers not chained: A=%q, B=%q", req.Header.Get("A"), req.Header.Get("B"))
		}
		return &http.Response{StatusCode: 204, Body: io.NopCloser(bytes.NewBuffer(nil))}, nil
	})

	rt := Chain(next, AddHeader("A", "1"), AddHeader("B", "2"))
	_, _ = rt.RoundTrip(&http.Request{Header: http.Header{}})
}

func TestChain_ClassifyStatus_PassThrough(t *testing.T) {
	next := helperRoundTripper(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBufferString("ok"))}, nil
	})

	rt := ClassifyStatus(ai.ProviderAnthropic)(next)
	res, err := rt.RoundTrip(&http.Request{})
	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}

	data, _ := io.ReadAll(res.Body)
	if got := string(data); got != "ok" {
		t.Errorf("body: want \"ok\", got %q", got)
	}
}

func TestChain_ClassifyStatus_ErrorWithRetryAfter_Seconds(t *testing.T) {
	payload := []byte(`{"error":{"code":"rate_limit","message":"too many"}}`)
	next := helperRoundTripper(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 429,
			Header:     http.Header{"Retry-After": {"5"}},
			Body:       io.NopCloser(bytes.NewReader(payload)),
		}, nil
	})

	rt := ClassifyStatus(ai.ProviderOpenAI)(next)
	_, err := rt.RoundTrip(&http.Request{})

	httpErr, ok := err.(*ai.HTTPError)
	if !ok {
		t.Fatalf("want HTTPError, got %T", err)
	}
	if want := "rate_limit"; httpErr.Type != want {
		t.Errorf("type: want %q, got %q", want, httpErr.Type)
	}
	if want := "too many"; httpErr.Message != want {
		t.Errorf("message: want %q, got %q", want, httpErr.Message)
	}
	if want := 5 * time.Second; httpErr.RetryAfter != want {
		t.Errorf("retry after: want %v, got %v", want, httpErr.RetryAfter)
	}
}

func TestChain_ClassifyStatus_ErrorWithRetryAfter_HTTPDatetime(t *testing.T) {
	future := time.Now().Add(3 * time.Second).UTC()
	httpDate := future.Format(http.TimeFormat)
	payload := []byte(`{"error":{"type":"server_error","message":"oops"}}`)

	next := helperRoundTripper(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 500,
			Header:     http.Header{"Retry-After": {httpDate}},
			Body:       io.NopCloser(bytes.NewReader(payload)),
		}, nil
	})

	rt := ClassifyStatus(ai.ProviderAnthropic)(next)
	_, err := rt.RoundTrip(&http.Request{})

	httpErr, ok := err.(*ai.HTTPError)
	if !ok {
		t.Fatalf("want *ai.HTTPError, got %T", err)
	}
	if ra := httpErr.RetryAfter; ra < 2*time.Second || ra > 4*time.Second {
		t.Errorf("retry after, want roughly 3s, got %v", ra)
	}
}

func TestChain_UseCircuitBreaker_TripsBlocks(t *testing.T) {
	next := helperRoundTripper(func(req *http.Request) (*http.Response, error) {
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
		t.Errorf("want underlying error, got %v", err1)
	}
	// circuit now open
	_, err2 := rt.RoundTrip(&http.Request{})
	if err2 != breaker.ErrOpen {
		t.Errorf("want ErrOpen, got %v", err2)
	}
}

func TestChain_UseRetrier_RetryOnError(t *testing.T) {
	attempts := 0
	next := helperRoundTripper(func(req *http.Request) (*http.Response, error) {
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
		t.Fatalf("want not error, got %v", err)
	}
	data, _ := io.ReadAll(resp.Body)
	if string(data) != "done" || attempts != 3 {
		t.Errorf("use retrier: got %q after %d attempts, want \"done\", and 3 attempts", data, attempts)
	}
}

func TestChain_UseLogger_PassThroughNilLogger(t *testing.T) {
	orig := http.DefaultTransport
	rt := UseLogger(nil)(orig)
	if rt != orig {
		t.Error("use logger with nil: want original round tripper")
	}
}

func TestChain_UseLogger(t *testing.T) {
	tests := []struct {
		name        string
		stubResp    *http.Response
		stubErr     error
		wantPattern string // regex to match in log output
	}{
		{
			name:        "LogInfoWithDurationMS",
			stubResp:    &http.Response{StatusCode: 200},
			stubErr:     nil,
			wantPattern: `transport.*method=GET.*url=http://example\.com/.*duration_ms=\d+`,
		},
		{
			name:        "LogErrorWithErrorMessage",
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
			logger := logger.NewWithHandler(handler)

			// wrap the stub RoundTripper
			stub := &stubRoundTripper{resp: tt.stubResp, err: tt.stubErr}
			rt := UseLogger(logger)(stub)

			// issue a single GET request
			req, _ := http.NewRequest("GET", "http://example.com/", nil)
			start := time.Now()
			res, err := rt.RoundTrip(req)
			elapsed := time.Since(start)

			// ensure behavior is preserved
			if res != tt.stubResp {
				t.Errorf("response: want %v; got %v", tt.stubResp, res)
			}
			if err != tt.stubErr {
				t.Errorf("error: want %v, got %v", tt.stubErr, err)
			}

			out := buf.String()
			// verify timestamp-derived duration is sensible
			if elapsed < 0 {
				t.Errorf("elapsed time: want non-negative, got %v", elapsed)
			}

			// match against expected log pattern
			re := regexp.MustCompile(tt.wantPattern)
			if !re.MatchString(out) {
				t.Errorf("log output did not match.\ngot: %s\nwant: %s", out, tt.wantPattern)
			}
		})
	}
}

func TestFactory_NewRetryClassifier(t *testing.T) {
	tests := []struct {
		name      string
		retryable map[ai.ErrType]struct{}
		err       error
		want      bool
	}{
		{
			name:      "HTTPErrorRetryable",
			retryable: map[ai.ErrType]struct{}{"foo": {}},
			err:       &ai.HTTPError{Type: "foo"},
			want:      true,
		},

		{
			name:      "HTTPErrorNotRetryable",
			retryable: map[ai.ErrType]struct{}{"bar": {}},
			err:       &ai.HTTPError{Type: "foo"},
			want:      false,
		},

		{
			name:      "WrappedHTTPError",
			retryable: map[ai.ErrType]struct{}{"baz": {}},
			err:       fmt.Errorf("wrap: %w", &ai.HTTPError{Type: "baz"}),
			want:      true,
		},

		{
			name:      "ContextCanceled",
			retryable: map[ai.ErrType]struct{}{"foo": {}},
			err:       context.Canceled,
			want:      false,
		},

		{
			name:      "ContextDeadlineExceeded",
			retryable: map[ai.ErrType]struct{}{"foo": {}},
			err:       context.DeadlineExceeded,
			want:      false,
		},

		{
			name:      "NetTimeout",
			retryable: nil,
			err:       timeoutErr{},
			want:      true,
		},

		{
			name:      "NetNonTimeout",
			retryable: nil,
			err:       temporaryErr{},
			want:      false,
		},

		{
			name:      "OtherERror",
			retryable: map[ai.ErrType]struct{}{"x": {}},
			err:       errors.New("x"),
			want:      false,
		},
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
