package aiclient

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

type fakeNetErr struct{ timeout, temporary bool }

func (f fakeNetErr) Error() string   { return "fake net error" }
func (f fakeNetErr) Timeout() bool   { return f.timeout }
func (f fakeNetErr) Temporary() bool { return f.temporary }

func wrap(err error) error { return fmt.Errorf("wrapped: %w", err) }

func TestProvider_RetryPolicy(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"OpenAI429", &openaiError{StatusCode: 429}, true},
		{"OpenAI500", &openaiError{StatusCode: 502}, true},
		{"OpenAI400", &openaiError{StatusCode: 400}, false},
		{"Anthropic429", &anthropicError{StatusCode: 429}, true},
		{"Anthropic502", &anthropicError{StatusCode: 502}, true},
		{"Anthropic400", &anthropicError{StatusCode: 400}, false},
		{"ElevenLabs429", &elevenlabsError{StatusCode: 429}, true},
		{"ElevenLabs500", &elevenlabsError{StatusCode: 500}, true},
		{"ElevenLabs400", &elevenlabsError{StatusCode: 400}, false},
		{"Wrapped", wrap(&openaiError{StatusCode: 503}), true},
		{"NetTimeoutTrue", fakeNetErr{timeout: true}, true},
		{"NetTimeoutFalse", fakeNetErr{timeout: false}, false},
		{"ContextCanceled", context.Canceled, false},
		{"DeadlineExceeded", context.DeadlineExceeded, false},
		{"SomeOtherErr", errors.New("some oher error"), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isRetryable(tc.err)
			if got != tc.want {
				t.Errorf("%#v: want %v, got %v", tc.err, tc.want, got)
			}
		})
	}
}
