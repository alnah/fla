package tripper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/alnah/fla/logger"
)

type stubBreaker struct {
	open     bool
	executed bool
	sawCtx   context.Context
}

func (s *stubBreaker) Execute(ctx context.Context, op func(context.Context) error) error {
	s.executed = true
	if s.open {
		return errors.New("breaker is open")
	}
	s.sawCtx = ctx
	return op(ctx)
}

type stubRetrier struct {
	retry bool
}

func (s *stubRetrier) Retry(
	ctx context.Context,
	op func(_ context.Context) error,
	isRetryable func(err error) bool,
) error {
	var err error
	for range 3 {
		err = op(ctx)
		if err == nil || !isRetryable(err) || !s.retry {
			break
		}
	}
	return err
}

type stubRoundTripper struct {
	res *http.Response
	err error
}

func (s *stubRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return s.res, s.err
}

func TestChain_Default(t *testing.T) {
	// build request because default is http.DefaultTransport and it will panic
	// if the header is not properly defined
	req, err := http.NewRequest("GET", "http://test.com", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	rt := Chain(nil)
	_, err = rt.RoundTrip(req)
	if err != nil {
		t.Errorf("round trip: want no error, got %v", err)
	}
}

func TestChain_WithCustomTripper(t *testing.T) {
	next := Tripper(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBuffer(nil)),
		}, nil
	})

	// build request because default is http.DefaultTransport and it will panic
	// if the header is not properly defined
	req, err := http.NewRequest("GET", "http://test.com", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	rt := Chain(Default(next))
	_, err = rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("round trip: want no error, got %v", err)
	}
}

func TestTripperChain_AddHeader(t *testing.T) {
	next := Tripper(func(r *http.Request) (*http.Response, error) {
		if got1 := r.Header.Get("Authorization"); got1 != "Bearer token" {
			t.Errorf("header: want %s, got %s", "Bearer token", got1)
		}
		if got2 := r.Header.Get("Content-Type"); got2 != "application/json" {
			t.Errorf("header: want %s, got %s", "application/json", got2)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBuffer(nil))}, nil
	})

	tripper := Chain(
		next,
		AddHeader("Authorization", "Bearer token"),
		AddHeader("Content-Type", "application/json"),
	)

	_, _ = tripper.RoundTrip(&http.Request{Header: http.Header{}})
}

func TestTripperChain_UseStatusClassifier_PassThrough(t *testing.T) {
	next := Tripper(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString("ok")),
		}, nil
	})

	tripper := UseStatusClassifier(
		func(code int) bool { return false },          // should not classify error
		func(res *http.Response) error { return nil }, // should not build error
	)(next)

	res, err := tripper.RoundTrip(&http.Request{})
	if err != nil {
		t.Fatalf("round trip: want no error, got %v", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("reading response body: want no error, got %v", err)
	}

	if string(body) != "ok" {
		t.Errorf("body: want %s, got %s", "ok", string(body))
	}
}

func TestTripperChain_UseStatusClassifier_ErrorRetryAfter(t *testing.T) {
	rawPayload := `{"error": {"code": "rate_limit", "message":"too many"}}`
	next := Tripper(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 429,
			Header:     http.Header{"Retry-After": {"5"}},
			Body:       io.NopCloser(bytes.NewBufferString(rawPayload)),
		}, nil
	})

	tripper := UseStatusClassifier(
		func(code int) bool { return code == 429 },
		func(res *http.Response) error {
			defer func() { _ = res.Body.Close() }()

			var parsedPayload struct {
				Error struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}

			if err := json.NewDecoder(res.Body).Decode(&parsedPayload); err != nil {
				t.Fatalf("want no error, got %v", err)
			}

			return fmt.Errorf(
				"error: code %s, message: %s, retry-after: %ss",
				parsedPayload.Error.Code,
				parsedPayload.Error.Message,
				res.Header.Get("Retry-After"),
			)
		},
	)(next)

	_, err := tripper.RoundTrip(&http.Request{})
	if err == nil {
		t.Fatal("want error, got nil")
	}

	for _, want := range []string{"rate_limit", "too many", "5s"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error string must contain: %s", want)
		}
	}
}

func TestUseCircuitBreaker(t *testing.T) {
	cases := []struct {
		name         string
		open         bool
		wantExec     bool
		wantErrPart  string
		wantCtxValue any
	}{
		{
			name:         "ClosedCallThroughWithNewCtx",
			open:         false,
			wantExec:     true,
			wantCtxValue: "injected",
		},
		{
			name:        "OpenSkipCallAndError",
			open:        true,
			wantExec:    true,
			wantErrPart: "breaker is open",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stubBreaker := &stubBreaker{open: tc.open}

			var gotReq *http.Request
			next := Tripper(func(r *http.Request) (*http.Response, error) {
				gotReq = r
				return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBuffer(nil))}, nil
			})

			rt := UseCircuitBreaker(stubBreaker)(next)

			req, _ := http.NewRequest(http.MethodGet, "http://test.com", nil)
			res, err := rt.RoundTrip(req)

			if !stubBreaker.executed {
				t.Error("execute was not called")
			}
			if tc.wantErrPart != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErrPart) {
					t.Fatalf("want error containing %q, got %v", tc.wantErrPart, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("want no error: got %v", err)
			}
			if gotReq == nil {
				t.Fatal("underlying round trip was not called")
			}
			// make sure injected context got through
			if gotReq.Context().Value(key) != tc.wantCtxValue {
				t.Errorf("ctx value: want %v, got %v", tc.wantCtxValue, gotReq.Context().Value(key))
			}
			if res == nil {
				t.Error("want res, got nil")
			}
		})
	}
}

func TestChain_UseRetrier_RetryOnError(t *testing.T) {
	attempts := 0
	next := Tripper(func(req *http.Request) (*http.Response, error) {
		attempts++
		if attempts < 3 {
			return nil, errors.New("transient")
		}
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString("done")),
		}, nil
	})

	r := &stubRetrier{retry: true}
	rt := UseRetrier(r, func(err error) bool { return true })(next)

	res, err := rt.RoundTrip(&http.Request{})
	if err != nil {
		t.Fatalf("round trip: want no error, got %v", err)
	}
	defer func() { _ = res.Body.Close() }()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("reading body: want no error, got %v", err)
	}

	if string(data) != "done" || attempts != 3 {
		t.Errorf("using retrier: got %s after %d attempts, want done and 3 attempts", data, attempts)
	}
}

func TestChain_UseLogger_NilLogger(t *testing.T) {
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
			name:     "LogInfoWithDurationMS",
			stubResp: &http.Response{StatusCode: 200},
			stubErr:  nil,
		},
		{
			name:     "LogErrorWithErrorMessage",
			stubResp: nil,
			stubErr:  errors.New("network fail"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare a buffer to capture log output
			logger := logger.Test()

			// wrap the stub RoundTripper
			stub := &stubRoundTripper{res: tt.stubResp, err: tt.stubErr}
			rt := UseLogger(logger)(stub)

			// issue a single GET request
			req, _ := http.NewRequest(http.MethodGet, "http://test.com/", nil)
			res, err := rt.RoundTrip(req)

			// ensure behavior is preserved
			if res != tt.stubResp {
				t.Errorf("response: want %v; got %v", tt.stubResp, res)
			}
			if err != tt.stubErr {
				t.Errorf("error: want %v, got %v", tt.stubErr, err)
			}

		})
	}
}
