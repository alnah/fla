package breaker

import (
	"time"

	"github.com/alnah/fla/clock"
)

// Config collects all tunables; zero values pick sane defaults.
type Config struct {
	FailureThreshold int           // consecutive failures to trip
	FailureWindow    time.Duration // look-back window for failures
	SuccessThreshold int           // consecutive successes to recover
	OpenTimeout      time.Duration // base open time
	MaxOpenTimeout   time.Duration // upper cap for back-off
	BackoffMaxExp    uint          // 0: no back-off, 1:×2, 2:×4 …
	Clock            clock.Clock
	OnStateChange    func(prev, next state)
	Metrics          Metrics
}

// ThirdPartyConfig is tuned for paid, external providers.
func ThirdPartyConfig() Config {
	return Config{
		FailureThreshold: 3,
		SuccessThreshold: 1,
		OpenTimeout:      60 * time.Second,
		BackoffMaxExp:    4, // 60 s → 2 m → 4 m → 5 m (cap)
		MaxOpenTimeout:   5 * time.Minute,
		Clock:            clock.New(),
	}
}

// WebAPIConfig targets steady public APIs around 100 queries per second.
func WebAPIConfig() Config {
	return Config{
		FailureThreshold: 5,
		FailureWindow:    10 * time.Second,
		SuccessThreshold: 2,
		OpenTimeout:      20 * time.Second,
		BackoffMaxExp:    2, // 20 s → 40 s → 60 s (cap)
		MaxOpenTimeout:   60 * time.Second,
		Clock:            clock.New(),
	}
}

// LowQPSConfig returns a conservative setup for sporadic internal jobs.
func LowQPSConfig() Config {
	return Config{
		FailureThreshold: 5,
		SuccessThreshold: 1,
		OpenTimeout:      30 * time.Second,
		// consecutive counting ⇒ leave FailureWindow = 0
		BackoffMaxExp: 0, // disabled
		Clock:         clock.New(),
	}
}

// HighQPSConfig suits 1000-10000 queries per second for back-end services.
func HighQPSConfig() Config {
	return Config{
		FailureThreshold: 20,
		FailureWindow:    5 * time.Second,
		SuccessThreshold: 3,
		OpenTimeout:      10 * time.Second,
		BackoffMaxExp:    3, // 10 s → 20 s → 40 s → 80 s (cap)
		MaxOpenTimeout:   90 * time.Second,
		Clock:            clock.New(),
	}
}

// CriticalPathConfig minimises false trips for payments / checkout flows.
func CriticalPathConfig() Config {
	return Config{
		FailureThreshold: 10,
		FailureWindow:    5 * time.Second,
		SuccessThreshold: 2,
		OpenTimeout:      5 * time.Second,
		BackoffMaxExp:    1, // 5 s → 10 s (cap)
		MaxOpenTimeout:   20 * time.Second,
		Clock:            clock.New(),
	}
}
