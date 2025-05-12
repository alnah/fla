package clock

import (
	"testing"
	"time"
)

func TestSystemClockNow(t *testing.T) {
	sys := New()
	now := time.Now()
	diff := sys.Now().Sub(now)

	if diff < 0 || diff > time.Second {
		t.Errorf("expected system clock close to now, got diff %v", diff)
	}
}

func TestFakeClockNow(t *testing.T) {
	start := time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC)
	fake := NewFakeClock(start)

	if !fake.Now().Equal(start) {
		t.Errorf("expected %v, got %v", start, fake.Now())
	}
}

func TestFakeClockSleep(t *testing.T) {
	start := time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC)
	fake := NewFakeClock(start)

	fake.Sleep(2 * time.Hour)
	expected := start.Add(2 * time.Hour)

	if !fake.Now().Equal(expected) {
		t.Errorf("expected %v after sleep, got %v", expected, fake.Now())
	}
}
