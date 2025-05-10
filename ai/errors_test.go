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

/********* Unit Tests *********/

func TestHTTPError(t *testing.T) {
	testCases := []struct {
		name           string
		provider       Provider
		body           string
		wantStatus     string
		wantType       string
		wantMessage    string
		wantRetryAfter time.Duration
	}{
		{
			name:           "openai",
			provider:       ProviderOpenAI,
			body:           `{"error": {"code": "test", "message": "Test"}}`,
			wantStatus:     "500 Internal Server Error",
			wantType:       "test",
			wantMessage:    "Test",
			wantRetryAfter: 1 * time.Second,
		},
		{
			name:           "anthropic",
			provider:       ProviderAnthropic,
			body:           `{"error": {"type": "test", "message": "Test"}}`,
			wantStatus:     "500 Internal Server Error",
			wantType:       "test",
			wantMessage:    "Test",
			wantRetryAfter: 1 * time.Second,
		},
		{
			name:           "elevenlabs",
			provider:       ProviderElevenLabs,
			body:           `{"detail": {"status": "test", "message": "Test"}}`,
			wantStatus:     "500 Internal Server Error",
			wantType:       "test",
			wantMessage:    "Test",
			wantRetryAfter: 1 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(tc.body))
			}))
			t.Cleanup(srv.Close)

			res, err := http.Get(srv.URL)
			if err != nil {
				t.Fatalf("http get: got %v, want no error", err)
			}

			httpErr := NewHTTPError(tc.provider, res)

			if got, want := httpErr.Status, tc.wantStatus; got != want {
				t.Errorf("status: got %q, want %q", got, want)
			}
			if got, want := httpErr.Type, tc.wantType; got != want {
				t.Errorf("type: got %q, want %q", got, want)
			}
			if got, want := httpErr.Message, tc.wantMessage; got != want {
				t.Errorf("message: got %q, want %q", got, want)
			}
			if got, want := httpErr.RetryAfter, tc.wantRetryAfter; got != want {
				t.Errorf("retryafter: got %v, want %v", got, want)
			}

			wantErr := fmt.Sprintf("status=%s, type=%s, message=%s", tc.wantStatus, tc.wantType, tc.wantMessage)
			if got := httpErr.Error(); got != wantErr {
				t.Errorf("error string: got %q, want %q", got, wantErr)
			}
		})
	}
}

func TestNewRetryClassifier(t *testing.T) {
	retryMap := map[ErrType]struct{}{ErrType("retryable"): {}}
	classify := NewRetryClassifier(retryMap)

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "httperror retryable",
			err:  &HTTPError{Type: ErrType("retryable").String()},
			want: true,
		},
		{
			name: "httperror non-retryable",
			err:  &HTTPError{Type: ErrType("other").String()},
			want: false,
		},
		{
			name: "context.canceled",
			err:  context.Canceled,
			want: false,
		},
		{
			name: "context.deadlineexceeded",
			err:  context.DeadlineExceeded,
			want: false,
		},
		{
			name: "net.error timeout",
			err:  fakeNetErr{timeout: true},
			want: true,
		},
		{
			name: "net.error non-timeout",
			err:  fakeNetErr{temporary: true},
			want: false,
		},
		{
			name: "other error",
			err:  errors.New("err"),
			want: false,
		},
		{
			name: "wrapped httperror",
			err:  fmt.Errorf("wrap: %w", &HTTPError{Type: ErrType("retryable").String()}),
			want: true,
		},
		{
			name: "wrapped net.error timeout",
			err:  fmt.Errorf("wrap: %w", fakeNetErr{timeout: true}),
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			got := classify(tc.err)
			if got != tc.want {
				t.Errorf("classify: got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAIErrorErrorString(t *testing.T) {
	op := Operation("operaton")
	pvd := Provider("provider")

	tests := []struct {
		name string
		err  *AIError
		want string
	}{
		{
			name: "no wrapped error",
			err:  &AIError{Operation: op, Provider: pvd, Message: "something went wrong"},
			want: "operaton provider error: something went wrong",
		},
		{
			name: "wrapped same msg",
			err:  &AIError{Operation: op, Provider: pvd, Message: "inner msg", Wrapped: errors.New("inner msg")},
			want: "operaton provider: inner msg",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			if got := tc.err.Error(); got != tc.want {
				t.Errorf("error string: got %q, want %q", got, tc.want)
			}
		})
	}

	t.Run("wrapped different msg", func(t *testing.T) {
		t.Helper()

		ae := &AIError{
			Operation: op,
			Provider:  pvd,
			Message:   "outer msg",
			Wrapped:   errors.New("inner msg"),
		}

		got := ae.Error()
		prefix := "operaton provider: outer msg: "
		if !strings.HasPrefix(got, prefix) {
			t.Errorf("error string prefix: got %q, want prefix %q", got, prefix)
		}
		if !strings.Contains(got, "inner msg") {
			t.Errorf("error string: %q should contain %q", got, "inner msg")
		}
	})
}

func TestNewAIErrorUnwrapAndConstructors(t *testing.T) {
	opChat := OpChatCompletion
	opTTS := OpTTSAudio
	opSTT := OpSTTTranscription
	p := Provider("provider")
	base := errors.New("root cause")

	cases := []struct {
		name       string
		op         Operation
		pvd        Provider
		msg        string
		wrapped    error
		wantOp     Operation
		wantPvd    Provider
		wantMsg    string
		wantUnwrap error
	}{
		{
			name:       "newaierror",
			op:         opChat,
			pvd:        p,
			msg:        "msg1",
			wrapped:    base,
			wantOp:     opChat,
			wantPvd:    p,
			wantMsg:    "msg1",
			wantUnwrap: base,
		},
		{
			name:       "newchaterror",
			op:         opChat,
			pvd:        p,
			msg:        "chat failed",
			wrapped:    base,
			wantOp:     opChat,
			wantPvd:    p,
			wantMsg:    "chat failed",
			wantUnwrap: base,
		},
		{
			name:       "newttserror",
			op:         opTTS,
			pvd:        p,
			msg:        "tts failed",
			wrapped:    base,
			wantOp:     opTTS,
			wantPvd:    p,
			wantMsg:    "tts failed",
			wantUnwrap: base,
		},
		{
			name:       "newstterror",
			op:         opSTT,
			pvd:        p,
			msg:        "stt failed",
			wrapped:    base,
			wantOp:     opSTT,
			wantPvd:    p,
			wantMsg:    "stt failed",
			wantUnwrap: base,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()

			var err error
			switch tc.name {
			case "newaierror":
				err = NewAIError(tc.op, tc.pvd, tc.msg, tc.wrapped)
			case "newchaterror":
				err = NewChatError(tc.pvd, tc.msg, tc.wrapped)
			case "newttserror":
				err = NewTTSError(tc.pvd, tc.msg, tc.wrapped)
			case "newstterror":
				err = NewSTTError(tc.pvd, tc.msg, tc.wrapped)
			}

			ae, ok := err.(*AIError)
			if !ok {
				t.Fatalf("type: got %T, want *AIError", err)
			}
			if got, want := ae.Operation, tc.wantOp; got != want {
				t.Errorf("operation: got %q, want %q", got, want)
			}
			if got, want := ae.Provider, tc.wantPvd; got != want {
				t.Errorf("provider: got %q, want %q", got, want)
			}
			if got, want := ae.Message, tc.wantMsg; got != want {
				t.Errorf("message: got %q, want %q", got, want)
			}
			if got, want := errors.Unwrap(err), tc.wantUnwrap; got != want {
				t.Errorf("unwrap: got %v, want %v", got, want)
			}
			if !errors.Is(err, tc.wantUnwrap) {
				t.Errorf("errors.is: %v should be recognized", tc.wantUnwrap)
			}

			out := err.Error()
			want := fmt.Sprintf("%s %s", tc.wantOp, tc.wantPvd)
			if !strings.Contains(out, want) || !strings.Contains(out, tc.wantMsg) {
				t.Errorf("error string: got %q, want it to contain %q and %q", out, want, tc.wantMsg)
			}
		})
	}
}
