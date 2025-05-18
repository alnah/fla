package retrier

import (
	"testing"
	"time"
)

func TestNoJitter_AllDurations(t *testing.T) {
	cases := []time.Duration{
		-5 * time.Millisecond,
		0,
		1 * time.Nanosecond,
		50 * time.Millisecond,
	}
	for _, d := range cases {
		got := NoJitter(d)
		if got != d {
			t.Errorf("no jitter: want %v, got %v", d, got)
		}
	}
}

func TestFullJitter_NonPositive(t *testing.T) {
	for _, d := range []time.Duration{-10, 0} {
		got := FullJitter(d)
		if got != 0 {
			t.Errorf("full jitter non-positive: want 0, got %v", got)
		}
	}
}

func TestFullJitter_RandomRange(t *testing.T) {
	const base = 100 * time.Millisecond
	for range 100 {
		got := FullJitter(base)
		if got < 0 || got > base {
			t.Errorf("full jitter: want in [0, %v], got %v", base, got)
		}
	}
}

func TestEqualJitter_NonPositive(t *testing.T) {
	for _, d := range []time.Duration{-20, 0} {
		got := EqualJitter(d)
		if got != 0 {
			t.Errorf("equal jitter non-positive: want 0, got %v", got)
		}
	}
}

func TestEqualJitter_RandomRange(t *testing.T) {
	const base = 200 * time.Millisecond
	half := base / 2
	for range 100 {
		got := EqualJitter(base)
		if got < half || got > base {
			t.Errorf("equal jitter: want in [%v, %v], got %v", half, base, got)
		}
	}
}
