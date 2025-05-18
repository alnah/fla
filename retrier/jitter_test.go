package retrier

import (
	"testing"
	"time"
)

func TestNoJitter_AllDurations(t *testing.T) {
	cases := []time.Duration{
		-5 * time.Millisecond, 0,
		1 * time.Nanosecond,
		50 * time.Millisecond,
	}
	for _, d := range cases {
		got := NoJitter(d)
		if got != d {
			t.Errorf("NoJitter(%v) = %v; want %v", d, got, d)
		}
	}
}

func TestFullJitter_NonPositive(t *testing.T) {
	for _, d := range []time.Duration{-10, 0} {
		if got := FullJitter(d); got != 0 {
			t.Errorf("FullJitter(%v) = %v; want 0", d, got)
		}
	}
}

func TestFullJitter_RandomRange(t *testing.T) {
	const base = 100 * time.Millisecond
	// run enough samples to get confidence
	for i := 0; i < 100; i++ {
		got := FullJitter(base)
		if got < 0 || got > base {
			t.Errorf("FullJitter(%v) = %v; want in [0, %v]", base, got, base)
		}
	}
}

func TestEqualJitter_NonPositive(t *testing.T) {
	for _, d := range []time.Duration{-20, 0} {
		if got := EqualJitter(d); got != 0 {
			t.Errorf("EqualJitter(%v) = %v; want 0", d, got)
		}
	}
}

func TestEqualJitter_RandomRange(t *testing.T) {
	const base = 200 * time.Millisecond
	half := base / 2
	for i := 0; i < 100; i++ {
		got := EqualJitter(base)
		if got < half || got > base {
			t.Errorf("EqualJitter(%v) = %v; want in [%v, %v]", base, got, half, base)
		}
	}
}
