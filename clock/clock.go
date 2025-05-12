package clock

import (
	"sync"
	"time"
)

// Clock defines operations for obtaining the current time and sleeping.
// By coding against this interface, tests can inject a fake clock
// for predictable, repeatable timing behavior.
type Clock interface {
	// Now returns the current time.
	// Use in place of time.Now to allow injection of custom clocks.
	Now() time.Time

	// Sleep pauses execution for the specified duration.
	// Use in place of time.Sleep to allow tests to advance or skip time.
	Sleep(d time.Duration)
}

type systemClock struct{}

// New returns a Clock that delegates to the real system clock.
// Use this in non-test code when deterministic timing is not required.
func New() systemClock {
	return systemClock{}
}

// Now returns the current local time from the operating system.
func (systemClock) Now() time.Time {
	return time.Now()
}

// Sleep pauses the current goroutine for at least the duration d.
func (systemClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

// FakeClock is a clock for testing purposes, now safe for concurrent use.
type FakeClock struct {
	mu          sync.Mutex
	currentTime time.Time
}

// NewFakeClock initializes a FakeClock starting at the provided time.
func NewFakeClock(start time.Time) *FakeClock {
	return &FakeClock{currentTime: start}
}

// Now returns the current fake time.
func (f *FakeClock) Now() time.Time {
	f.mu.Lock()
	t := f.currentTime
	f.mu.Unlock()
	return t
}

// Sleep advances the fake time by the specified duration in a thread-safe manner.
func (f *FakeClock) Sleep(d time.Duration) {
	f.mu.Lock()
	f.currentTime = f.currentTime.Add(d)
	f.mu.Unlock()
}
