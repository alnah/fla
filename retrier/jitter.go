package retrier

import "time"

// jitter defines how we add randomness to a base backoff duration.
type jitter func(time.Duration) time.Duration

// NoJitter applies a constant delay between retries.
var NoJitter jitter = func(d time.Duration) time.Duration {
	return d
}

// FullJitter applies a random delay up to the full backoff interval.
var FullJitter jitter = func(d time.Duration) time.Duration {
	if d <= 0 {
		return 0
	}
	// crypto-secure random in [0, d]
	return secureRandomDuration(int64(d) + 1)
}

// EqualJitter applies a random delay between half and the full interval.
var EqualJitter jitter = func(d time.Duration) time.Duration {
	if d <= 0 {
		return 0
	}
	half := d / 2
	return half + secureRandomDuration(int64(half)+1)
}
