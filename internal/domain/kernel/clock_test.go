package kernel_test

import (
	"testing"
	"time"

	"github.com/alnah/fla/internal/domain/kernel"
)

// StubClock implements kernel.Clock for testing
type StubClock struct {
	t time.Time
}

func NewStubClock(t time.Time) *StubClock {
	return &StubClock{t: t}
}

func (s *StubClock) Now() time.Time {
	return s.t
}

func TestStubClock(t *testing.T) {
	t.Run("returns configured time", func(t *testing.T) {
		want := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		clock := NewStubClock(want)

		got := clock.Now()

		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("implements Clock interface", func(t *testing.T) {
		// This test ensures StubClock implements kernel.Clock
		var _ kernel.Clock = &StubClock{}
	})
}
