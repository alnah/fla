package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

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
	e := &HTTPError{Status: 418, Type: "t", Code: "c", Message: "m"}
	exp := "HTTPError: status=418, type=t, code=c, message=m"
	if got := e.Error(); got != exp {
		t.Errorf("Error() = %q, want %q", got, exp)
	}
}

func TestClassifyStatus_PassThrough(t *testing.T) {
	next := roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBufferString("ok"))}, nil
	})

	rt := ClassifyStatus(next)
	res, err := rt.RoundTrip(&http.Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body, _ := io.ReadAll(res.Body)
	if string(body) != "ok" {
		t.Errorf("unexpected body: %q", body)
	}
}

func TestClassifyStatus_ErrorWithRetryAfterSeconds(t *testing.T) {
	apiErr := APIError{}
	apiErr.Error.Type = "type"
	apiErr.Error.Code = "code"
	apiErr.Error.Message = "msg"
	b, _ := json.Marshal(apiErr)

	next := roundTripperTest(func(req *http.Request) (*http.Response, error) {
		headers := http.Header{"Retry-After": {"5"}}
		return &http.Response{StatusCode: 429, Header: headers, Body: io.NopCloser(bytes.NewReader(b))}, nil
	})

	rt := ClassifyStatus(next)
	_, err := rt.RoundTrip(&http.Request{})

	he, ok := err.(*HTTPError)
	if !ok {
		t.Fatalf("error is not HTTPError: %v", err)
	}
	if he.RetryAfter != 5*time.Second {
		t.Errorf("expected RetryAfter=5s, got %v", he.RetryAfter)
	}
	if he.Status != 429 || he.Type != "type" || he.Code != "code" || he.Message != "msg" {
		t.Errorf("unexpected HTTPError fields: %+v", he)
	}
}

func TestClassifyStatus_ErrorWithRetryAfterHTTPDate(t *testing.T) {
	future := time.Now().Add(3 * time.Second).UTC()
	datestr := future.Format(http.TimeFormat)
	apiErr := APIError{}
	apiErr.Error.Type = "t"
	apiErr.Error.Code = "c"
	apiErr.Error.Message = "m"
	b, _ := json.Marshal(apiErr)

	next := roundTripperTest(func(req *http.Request) (*http.Response, error) {
		headers := http.Header{"Retry-After": {datestr}}
		return &http.Response{StatusCode: 500, Header: headers, Body: io.NopCloser(bytes.NewReader(b))}, nil
	})

	rt := ClassifyStatus(next)
	_, err := rt.RoundTrip(&http.Request{})

	he, ok := err.(*HTTPError)
	if !ok {
		t.Fatalf("error is not HTTPError: %v", err)
	}
	d := he.RetryAfter
	if d < 2*time.Second || d > 4*time.Second {
		t.Errorf("RetryAfter out of expected range: %v", d)
	}
}

func TestUseCircuitBreaker_TripsAndBlocks(t *testing.T) {
	// underlying always errors
	next := roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("fail")
	})
	b := breaker.New(
		breaker.WithFailureThreshold(1),
		breaker.WithOpenTimeout(time.Minute),
		breaker.WithClock(clock.NewFakeClock(time.Now())),
	)

	rt := UseCircuitBreaker(b)(next)
	// failure
	_, err1 := rt.RoundTrip(&http.Request{})
	if err1 == nil || err1.Error() != "fail" {
		t.Errorf("expected underlying error, got %v", err1)
	}
	// blocked by circuit
	_, err2 := rt.RoundTrip(&http.Request{})
	if err2 != breaker.ErrOpen {
		t.Errorf("expected ErrOpen, got %v", err2)
	}
}

func TestUseRetrier_RetriesOnError(t *testing.T) {
	count := 0

	next := roundTripperTest(func(req *http.Request) (*http.Response, error) {
		count++
		if count < 3 {
			return nil, errors.New("retry")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBufferString("ok"))}, nil
	})
	fakeClock := clock.NewFakeClock(time.Now())

	r := retrier.New(
		retrier.WithMaxAttempts(5),
		retrier.WithBaseDelay(0),
		retrier.WithJitter(retrier.NoJitter),
		retrier.WithClock(fakeClock),
	)

	rt := UseRetrier(r, func(error) bool { return true })(next)
	res, err := rt.RoundTrip(&http.Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body, _ := io.ReadAll(res.Body)
	if string(body) != "ok" || count != 3 {
		t.Errorf("expected 3 attempts and ok body, got %d and %q", count, body)
	}
}

func TestUseLogger_NilLogger_PassesThrough(t *testing.T) {
	next := http.DefaultTransport
	rt := UseLogger(nil)(next)
	if rt != next {
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
		{
			name:      "HTTPError retryable",
			retryable: map[ErrType]struct{}{"foo": {}},
			err:       &HTTPError{Type: "foo"},
			want:      true,
		},
		{
			name:      "HTTPError not retryable",
			retryable: map[ErrType]struct{}{"other": {}},
			err:       &HTTPError{Type: "foo"},
			want:      false,
		},
		{
			name:      "wrapped HTTPError retryable",
			retryable: map[ErrType]struct{}{"bar": {}},
			err:       fmt.Errorf("wrapped: %w", &HTTPError{Type: "bar"}),
			want:      true,
		},
		{
			name:      "context canceled",
			retryable: map[ErrType]struct{}{"foo": {}},
			err:       context.Canceled,
			want:      false,
		},
		{
			name:      "context deadline",
			retryable: map[ErrType]struct{}{"foo": {}},
			err:       context.DeadlineExceeded,
			want:      false,
		},
		{
			name:      "net timeout error",
			retryable: map[ErrType]struct{}{"irrelevant": {}},
			err:       timeoutErr{},
			want:      true,
		},
		{
			name:      "net non-timeout error",
			retryable: nil,
			err:       temporaryErr{},
			want:      false,
		},
		{
			name:      "other random error",
			retryable: map[ErrType]struct{}{"foo": {}},
			err:       errors.New("random"),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classifier := NewRetryClassifier(tt.retryable)
			got := classifier(tt.err)
			if got != tt.want {
				t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

/********* Benchmark *********/

func BenchmarkChain_NoMiddleware(b *testing.B) {
	rt := Chain(http.DefaultTransport)
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rt.RoundTrip(req)
	}
}
