package retrier

import (
	"errors"
	"testing"
)

func TestErrorString(t *testing.T) {
	wrapped := errors.New("boom")
	re := &RetrierError{attempts: 3, wrapped: wrapped}

	want := "retrier: after 3 attempt(s): boom"
	if got := re.Error(); got != want {
		t.Errorf("error string: want %q, got %q", want, got)
	}
}

func TestUnwrapAndIs(t *testing.T) {
	wrapped := errors.New("boom")
	re := &RetrierError{attempts: 3, wrapped: wrapped}

	if got := re.Unwrap(); got != wrapped {
		t.Errorf("unwrap: want %v, got %v", wrapped, got)
	}
	if !errors.Is(re, wrapped) {
		t.Errorf("errors.Is: want true, got false")
	}
}

func TestNilWrapped(t *testing.T) {
	re := &RetrierError{attempts: 0, wrapped: nil}

	want := "retrier: after 0 attempt(s): <nil>"
	if got := re.Error(); got != want {
		t.Errorf("error string nil: want %q, got %q", want, got)
	}
	if got := re.Unwrap(); got != nil {
		t.Errorf("unwrap nil: want nil, got %v", got)
	}
}
