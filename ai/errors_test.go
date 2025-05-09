package ai

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

/********* Helpers *********/

type fakeNetErr struct {
	timeout   bool
	temporary bool
}

func (f fakeNetErr) Error() string   { return "fake network error" }
func (f fakeNetErr) Timeout() bool   { return f.timeout }
func (f fakeNetErr) Temporary() bool { return f.temporary }

/********* Tests *********/

func TestHTTPError(t *testing.T) {
	testCases := []struct {
		name       string
		provider   Provider
		body       string
		retryAfter time.Duration
	}{
		{
			name:     "OpenAI",
			provider: ProviderOpenAI,
			body:     `{"error": {"code": "test", "message": "Test"}}`,
		},
		{
			name:     "Anthropic",
			provider: ProviderAnthropic,
			body:     `{"error": {"type": "test", "message": "Test"}}`,
		},
		{
			name:     "ElevenLabs",
			provider: ProviderElevenLabs,
			body:     `{"detail": {"status": "test", "message": "Test"}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(tc.body))
			}))
			t.Cleanup(func() { srv.Close() })

			res, err := http.Get(srv.URL)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			httpErr := NewHTTPError(tc.provider, res)
			if httpErr.Status != "500 Internal Server Error" {
				t.Errorf("http error status should be %q, got %q", "500 Internal Server Error", httpErr.Status)
			}
			if httpErr.Type != "test" {
				t.Errorf("http error type should be %q, got %q", "test", httpErr.Type)
			}
			if httpErr.Message != "Test" {
				t.Errorf("http error message should be %q, got %q", "Test", httpErr.Message)
			}
			if httpErr.RetryAfter != 1*time.Second {
				t.Errorf("http error retry after value should be %v, got %v", 1*time.Second, httpErr.RetryAfter)
			}

			wantErrStr := "status=500 Internal Server Error, type=test, message=Test"
			if httpErr.Error() != wantErrStr {
				t.Errorf("http error string should be %q, got %q", wantErrStr, httpErr.Error())
			}
		})
	}
}

func TestNewRetryClassifier(t *testing.T) {
	retryableMap := map[ErrType]struct{}{
		ErrType("retryable"): {},
	}
	classify := NewRetryClassifier(retryableMap)

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "HTTPError retryable",
			err:  &HTTPError{Type: ErrType("retryable").String()},
			want: true,
		},
		{
			name: "HTTPError non-retryable",
			err:  &HTTPError{Type: ErrType("other").String()},
			want: false,
		},
		{
			name: "context.Canceled",
			err:  context.Canceled,
			want: false,
		},
		{
			name: "context.DeadlineExceeded",
			err:  context.DeadlineExceeded,
			want: false,
		},
		{
			name: "net.Error timeout",
			err:  fakeNetErr{timeout: true, temporary: false},
			want: true,
		},
		{
			name: "net.Error non-timeout",
			err:  fakeNetErr{timeout: false, temporary: true},
			want: false,
		},
		{
			name: "some other error",
			err:  errors.New("random error"),
			want: false,
		},
		{
			name: "wrapped HTTPError retryable",
			err:  fmt.Errorf("wrapper: %w", &HTTPError{Type: ErrType("retryable").String()}),
			want: true,
		},
		{
			name: "wrapped net.Error timeout",
			err:  fmt.Errorf("network wrap: %w", fakeNetErr{timeout: true}),
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := classify(tc.err)
			if got != tc.want {
				t.Errorf("classify: %v; got %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestAIError_Error(t *testing.T) {
	op := Operation("operaton")
	pvd := Provider("provider")

	tests := []struct {
		name string
		ae   *AIError
		want string
	}{
		{
			name: "no wrapped error",
			ae: &AIError{
				Operation: op,
				Provider:  pvd,
				Message:   "something went wrong",
				Wrapped:   nil,
			},
			want: "operaton provider error: something went wrong",
		},
		{
			name: "wrapped with same message",
			ae: &AIError{
				Operation: op,
				Provider:  pvd,
				Message:   "inner msg",
				Wrapped:   errors.New("inner msg"),
			},
			want: "operaton provider: inner msg",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.ae.Error()
			if got != tc.want {
				t.Fatalf("Error() = %q, want %q", got, tc.want)
			}
		})
	}

	t.Run("wrapped with different message", func(t *testing.T) {
		ae := &AIError{
			Operation: op,
			Provider:  pvd,
			Message:   "outer msg",
			Wrapped:   errors.New("inner msg"),
		}

		got := ae.Error()
		if !strings.HasPrefix(got, "operaton provider: outer msg: ") {
			t.Errorf("error string prefix = %q, want it to start with %q",
				got, "operaton provider: outer msg: ")
		}
		if !strings.Contains(got, "inner msg") {
			t.Errorf("error string %q, want it to contain inner error text", got)
		}
	})
}

func TestNewAIError_UnwrapAndConstructors(t *testing.T) {
	// prepare inputs
	opChat := OpChatCompletion
	opTTS := OpTTSAudio
	opSTT := OpSTTTranscription
	p := Provider("Provider")

	baseErr := errors.New("root cause")
	cases := []struct {
		name       string
		ctor       func() error
		wantOp     Operation
		wantPvd    Provider
		wantMsg    string
		wantUnwrap error
	}{
		{
			name:       "NewAIError",
			ctor:       func() error { return NewAIError(opChat, p, "msg1", baseErr) },
			wantOp:     opChat,
			wantPvd:    p,
			wantMsg:    "msg1",
			wantUnwrap: baseErr,
		},
		{
			name:       "NewChatError",
			ctor:       func() error { return NewChatError(p, "chat failed", baseErr) },
			wantOp:     opChat,
			wantPvd:    p,
			wantMsg:    "chat failed",
			wantUnwrap: baseErr,
		},
		{
			name:       "NewTTSError",
			ctor:       func() error { return NewTTSError(p, "tts failed", baseErr) },
			wantOp:     opTTS,
			wantPvd:    p,
			wantMsg:    "tts failed",
			wantUnwrap: baseErr,
		},
		{
			name:       "NewSTTError",
			ctor:       func() error { return NewSTTError(p, "stt failed", baseErr) },
			wantOp:     opSTT,
			wantPvd:    p,
			wantMsg:    "stt failed",
			wantUnwrap: baseErr,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.ctor()

			// should be *AIError
			ae, ok := err.(*AIError)
			if !ok {
				t.Fatalf("error type = %T, want *AIError", err)
			}

			// fields
			if ae.Operation != tc.wantOp {
				t.Errorf("operation should %q, want %q", ae.Operation, tc.wantOp)
			}
			if ae.Provider != tc.wantPvd {
				t.Errorf("provider should %q, want %q", ae.Provider, tc.wantPvd)
			}
			if ae.Message != tc.wantMsg {
				t.Errorf("message should %q, want %q", ae.Message, tc.wantMsg)
			}

			// unwrap behavior
			unwrapped := errors.Unwrap(err)
			if unwrapped != tc.wantUnwrap {
				t.Errorf("unwrap should be %v, want %v", unwrapped, tc.wantUnwrap)
			}

			// and errors.Is should work
			if !errors.Is(err, tc.wantUnwrap) {
				t.Errorf("errors is should work")
			}

			// a quick sanity on Error() containing operation, provider, and message
			out := err.Error()
			expect := fmt.Sprintf("%s %s", tc.wantOp, tc.wantPvd)
			if !strings.Contains(out, expect) || !strings.Contains(out, tc.wantMsg) {
				t.Errorf("got error string %q; want it to mention %q and %q", out, expect, tc.wantMsg)
			}
		})
	}
}
