package transport

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

func TestChain_AllMiddleware_PassThrough(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://test.com", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	var gotReq *http.Request
	next := RoundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotReq = r
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString("ok")),
		}, nil
	})

	stubBreaker := &stubBreaker{open: false}
	stubRetrier := &stubRetrier{retry: false}

	rt := Chain(
		next,
		AddHeader("Authorization", "Bearer token"),
		UseStatusClassifier(
			func(code int) bool { return false },
			func(res *http.Response) error { return nil },
		),
		UseCircuitBreaker(stubBreaker),
		UseRetrier(stubRetrier, func(err error) bool { return false }),
		UseLogger(nil),
	)

	res, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("failed to execute http transaction: %v", err)
	}
	defer func() { _ = res.Body.Close }()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if got := string(data); got != "ok" {
		t.Errorf("response body: want %q, got %q", "ok", got)
	}
	if !stubBreaker.executed {
		t.Error("circuit breaker execute was not called")
	}
	if got := gotReq.Header.Get("Authorization"); got != "Bearer token" {
		t.Errorf("authorization header: want %q, got %q", "Bearer token", got)
	}
	if got := gotReq.Context().Value(key); got != "injected" {
		t.Errorf("context key: want %q, got %v", "injected", got)
	}
}

func TestChain_Yield_DefaultTransport(t *testing.T) {
	rt := Chain(nil)
	if rt != http.DefaultTransport {
		t.Errorf("round tripper: want %+v, got %=v", http.DefaultTransport, rt)
	}
}

func TestChain_Pass_CustomRoundTripper(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://test.com", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	next := RoundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString("ok")),
		}, nil
	})

	rt := Chain(next)
	res, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("failed to execute http transaction: %v", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != 200 {
		t.Errorf("status code: want %d, got %d", 200, res.StatusCode)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if string(data) != "ok" {
		t.Errorf("response body data: want %s, got %s", "ok", string(data))
	}
}

func TestChain_AddHeader_Many(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://test.com", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	next := RoundTripFunc(func(r *http.Request) (*http.Response, error) {
		if authHeader := r.Header.Get("Authorization"); authHeader != "Bearer token" {
			t.Errorf("auth header: want %s, got %s", "Bearer token", authHeader)
		}
		if contHeader := r.Header.Get("Content-Type"); contHeader != "application/json" {
			t.Errorf("content-type header: want %s, got %s", contHeader, "application/json")
		}

		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBuffer(nil)),
		}, nil
	})

	rt := Chain(
		next,
		AddHeader("Authorization", "Bearer token"),
		AddHeader("Content-Type", "application/json"),
	)

	res, err := rt.RoundTrip(req)
	defer func() { _ = res.Body.Close() }()
	if err != nil {
		t.Errorf("error executing http transaction: %v", err)
	}
}

func TestChain_UseStatusClassifier_PassThrough(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://test.com", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	next := RoundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString("ok")),
		}, nil
	})

	rt := Chain(
		next,
		UseStatusClassifier(
			func(code int) bool { return false },          // should not classify error
			func(res *http.Response) error { return nil }, // should not create error
		),
	)

	res, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("failed to execute http transaction: %v", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != 200 {
		t.Errorf("status code: want %d, got %d", 200, res.StatusCode)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if string(data) != "ok" {
		t.Errorf("response body data: want %s, got %s", "ok", string(data))
	}
}

func TestChain_UseStatusClassifier_ErrorRetryAfter(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://test.com", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	rawPayload := `{"error": {"code": "rate_limit", "message":"too many"}}`
	next := RoundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 429,
			Header:     http.Header{"Retry-After": {"5"}},
			Body:       io.NopCloser(bytes.NewBufferString(rawPayload)),
		}, nil
	})

	rt := Chain(
		next,
		UseStatusClassifier(
			func(code int) bool { return code == 429 },
			func(res *http.Response) error {
				// buildError already closes res.Body
				var parsed struct {
					Error struct {
						Code    string `json:"code"`
						Message string `json:"message"`
					} `json:"error"`
				}
				if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
					t.Fatalf("decoding payload: %v", err)
				}
				defer res.Body.Close()
				return fmt.Errorf(
					"error: code %s, message: %s, retry-after: %ss",
					parsed.Error.Code,
					parsed.Error.Message,
					res.Header.Get("Retry-After"),
				)
			},
		),
	)

	_, err = rt.RoundTrip(req)
	if err == nil {
		t.Fatal("expected an error, but got none")
	}
	for _, want := range []string{"rate_limit", "too many", "5s"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error string must contain %q; got %q", want, err.Error())
		}
	}
}

func TestChain_UseCircuitBreaker(t *testing.T) {
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

			req, err := http.NewRequest(http.MethodGet, "http://test.com", nil)
			if err != nil {
				t.Fatalf("failed to built http request: %v", err)
			}

			var gotReq *http.Request
			next := RoundTripFunc(func(r *http.Request) (*http.Response, error) {
				gotReq = r
				return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBuffer(nil))}, nil
			})

			rt := Chain(
				next,
				UseCircuitBreaker(stubBreaker),
			)

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
				t.Fatalf("failed to execute http transaction: %v", err)
			}
			defer func() { _ = res.Body.Close() }()

			if gotReq == nil {
				t.Error("underlying round trip was not called")
			}
			if gotReq.Context().Value(key) != tc.wantCtxValue {
				t.Errorf("ctx value: want %v, got %v", tc.wantCtxValue, gotReq.Context().Value(key))
			}
		})
	}
}

func TestChain_UseRetrier_RetryOnError(t *testing.T) {
	attempts := 0
	next := RoundTripFunc(func(req *http.Request) (*http.Response, error) {
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
		t.Fatalf("failed to execute http transaction: %v", err)
	}
	defer func() { _ = res.Body.Close() }()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if string(data) != "done" || attempts != 3 {
		t.Errorf("using retrier: got %s after %d attempts, want done and 3 attempts", data, attempts)
	}
}

func TestChain_UseLogger_NilLogger(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://test.com/", nil)
	if err != nil {
		t.Fatalf("failed to build http request: %v", err)
	}

	stub := &stubRoundTripper{
		res: &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBuffer(nil))},
	}
	rt := Chain(
		stub,
		UseLogger(nil),
	)
	res, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("failed to execute http transaction: %v", err)
	}
	defer func() { _ = res.Body.Close() }()

	if rt != stub {
		t.Error("use logger with nil: want original round tripper, but didn't get it")
	}
}

func TestChain_UseLogger(t *testing.T) {
	tests := []struct {
		name     string
		stubResp *http.Response
		stubErr  error
	}{
		{
			name:     "Response",
			stubResp: &http.Response{StatusCode: 200},
			stubErr:  nil,
		},
		{
			name:     "Error",
			stubResp: nil,
			stubErr:  errors.New("network fail"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "http://test.com/", nil)
			if err != nil {
				t.Fatalf("failed to build http request: %v", err)
			}
			testLogger := logger.NewTestLogger()

			rt := Chain(
				&stubRoundTripper{res: tt.stubResp, err: tt.stubErr},
				UseLogger(testLogger),
			)

			res, err := rt.RoundTrip(req)
			// no need to close the body, it's nil
			if res != tt.stubResp {
				t.Errorf("response: want %v; got %v", tt.stubResp, res)
			}
			if err != tt.stubErr {
				t.Errorf("error: want %v, got %v", tt.stubErr, err)
			}
		})
	}
}

type stubBreaker struct {
	open     bool
	executed bool
	sawCtx   context.Context
}

func (s *stubBreaker) Success()

func (s *stubBreaker) Fail()

func (s *stubBreaker) Execute(
	ctx context.Context,
	op func(context.Context) error,
) error {
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
