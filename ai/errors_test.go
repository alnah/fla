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

func assertNoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
}

func assertValue(t testing.TB, name string, got, want any) {
	t.Helper()
	if got != want {
		t.Errorf("%s: want %q, got %q", name, want, got)
	}
}

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
			assertNoError(t, err)

			httpErr := NewHTTPError(tc.provider, res)
			assertValue(t, "status", httpErr.Status, tc.wantStatus)
			assertValue(t, "type", httpErr.Type, tc.wantType)
			assertValue(t, "message", httpErr.Message, tc.wantMessage)
			assertValue(t, "retry after", httpErr.RetryAfter.String(), tc.wantRetryAfter.String())

			wantErr := fmt.Sprintf("status=%s, type=%s, message=%s", tc.wantStatus, tc.wantType, tc.wantMessage)
			assertValue(t, "error string", httpErr.Error(), wantErr)
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
		t.Run(tc.name, func(t *testing.T) { assertValue(t, "classify", classify(tc.err), tc.want) })
	}
}

func TestAIErrorString(t *testing.T) {
	op := Operation("operation")
	pvd := Provider("provider")

	tests := []struct {
		name string
		err  *AIError
		want string
	}{
		{
			name: "no wrapped error",
			err:  &AIError{Operation: op, Provider: pvd, Message: "something went wrong"},
			want: "operation provider error: something went wrong",
		},
		{
			name: "wrapped same msg",
			err:  &AIError{Operation: op, Provider: pvd, Message: "inner msg", Wrapped: errors.New("inner msg")},
			want: "operation provider: inner msg",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) { assertValue(t, "error string", tc.err.Error(), tc.want) })
	}

	t.Run("wrapped different msg", func(t *testing.T) {
		ae := &AIError{
			Operation: op,
			Provider:  pvd,
			Message:   "outer msg",
			Wrapped:   errors.New("inner msg"),
		}

		got := ae.Error()
		prefix := "operation provider: outer msg: "
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

			aiErr, ok := err.(*AIError)
			if !ok {
				t.Fatalf("type: got %T, want *AIError", err)
			}
			assertValue(t, "operation", aiErr.Operation, tc.wantOp)
			assertValue(t, "provider", aiErr.Provider, tc.wantPvd)
			assertValue(t, "message", aiErr.Message, tc.wantMsg)
			assertValue(t, "unwrap", errors.Unwrap(err), tc.wantUnwrap)
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
