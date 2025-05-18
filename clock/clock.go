package clock

import (
	"time"
)

// Clock defines operations for obtaining the current time and sleeping.
// By coding against this interface, tests can inject a fake clock
// for predictable, repeatable timing behavior.
type Clock interface {
	Now() time.Time
	Sleep(d time.Duration)
}

type clock struct{}

// New returns a Clock that delegates to the real system clock.
// Use this in non-test code when deterministic timing is not required.
func New() clock {
	return clock{}
}

// Now returns the current local time from the operating system.
func (clock) Now() time.Time {
	return time.Now()
}

// Sleep pauses the current goroutine for at least the duration d.
func (clock) Sleep(d time.Duration) {
	time.Sleep(d)
}
