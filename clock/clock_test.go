package clock

import (
	"testing"
	"time"
)

func TestClock_Now(t *testing.T) {
	c := New()
	before := time.Now()
	got := c.Now()
	if got.Before(before) {
		t.Fatalf("Now() = %v; want ≥ %v", got, before)
	}
}

func TestClock_SleepAtLeast(t *testing.T) {
	c := New()
	d := 10 * time.Millisecond
	start := time.Now()
	c.Sleep(d)
	elapsed := time.Since(start)
	if elapsed < d {
		t.Fatalf("slept %v; want ≥ %v", elapsed, d)
	}
}

func TestClock_SleepZero(t *testing.T) {
	c := New()
	start := time.Now()
	c.Sleep(0)
	if since := time.Since(start); since > time.Millisecond {
		t.Fatalf("took %v; want ≲ 1ms", since)
	}
}
