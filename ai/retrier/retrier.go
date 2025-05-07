package retrier

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/alnah/fla/ai/clock"
)

// RetrierError is returned when an operation fails after exhausting all retry attempts.
type RetrierError struct {
	Attempts int   // total attempts made (≥ maxAttempts)
	Err      error // last error encountered
}

func (e *RetrierError) Error() string {
	return fmt.Sprintf("retrier: after %d attempt(s): %v", e.Attempts, e.Err)
}

func (e *RetrierError) Unwrap() error { return e.Err }

type jitter func(d time.Duration, rnd *rand.Rand) time.Duration

var (
	// NoJitter applies a constant delay between retries.
	NoJitter jitter = func(d time.Duration, _ *rand.Rand) time.Duration { return d }

	// FullJitter applies a random delay up to the full backoff interval.
	FullJitter jitter = func(d time.Duration, rnd *rand.Rand) time.Duration {
		if d <= 0 {
			return 0
		}
		return time.Duration(rnd.Int63n(int64(d) + 1))
	}

	// EqualJitter applies a random delay between half and the full backoff interval.
	EqualJitter jitter = func(d time.Duration, rnd *rand.Rand) time.Duration {
		if d <= 0 {
			return 0
		}
		half := d / 2
		return half + time.Duration(rnd.Int63n(int64(half)+1))
	}
)

type operation func(context.Context) error

type option func(*Retrier)

// Retrier manages retrying an operation with configurable backoff, jitter,
// maximum attempts, and an optional hook before each retry.
type Retrier struct {
	maxAttempts int           // maximum number of attempts (including the first)
	baseDelay   time.Duration // initial backoff interval
	multiplier  float64       // factor by which the delay grows after each failure
	maxDelay    time.Duration // upper bound on backoff interval
	jitter      jitter
	rand        *rand.Rand
	randMu      sync.Mutex
	clock       clock.Clock
	onRetry     func(attempt int, err error, nextDelay time.Duration)
}

// WithMaxAttempts sets how many times an operation will be tried before giving up.
func WithMaxAttempts(n int) option { return func(r *Retrier) { r.maxAttempts = n } }

// WithBaseDelay sets the initial delay before the first retry.
func WithBaseDelay(d time.Duration) option { return func(r *Retrier) { r.baseDelay = d } }

// WithMultiplier sets the growth factor for exponential backoff.
func WithMultiplier(m float64) option { return func(r *Retrier) { r.multiplier = m } }

// WithMaxDelay caps the backoff delay to avoid excessively long waits.
func WithMaxDelay(d time.Duration) option { return func(r *Retrier) { r.maxDelay = d } }

// WithJitter injects randomness into backoff intervals to spread retries over time.
func WithJitter(j jitter) option { return func(r *Retrier) { r.jitter = j } }

// WithRand provides a custom random source for deterministic jitter in tests.
func WithRand(src rand.Source) option { return func(r *Retrier) { r.rand = rand.New(src) } }

// WithClock replaces the time source and sleep function, enabling controlled testing.
func WithClock(c clock.Clock) option { return func(r *Retrier) { r.clock = c } }

// WithOnRetry registers a callback that is invoked before each retry attempt.
func WithOnRetry(fn func(attempt int, err error, nextDelay time.Duration)) option {
	return func(r *Retrier) { r.onRetry = fn }
}

const (
	defaultAttempts   = 3
	defaultBaseDelay  = 100 * time.Millisecond
	defaultMultiplier = 2.0
	defaultMaxDelay   = 30 * time.Second
)

// New creates a Retrier with default settings and applies any provided options.
// Defaults: 3 attempts, 100ms base delay, 2× multiplier, 30s max delay, full jitter.
func New(opts ...option) *Retrier {
	r := &Retrier{
		maxAttempts: defaultAttempts,
		baseDelay:   defaultBaseDelay,
		multiplier:  defaultMultiplier,
		maxDelay:    defaultMaxDelay,
		jitter:      FullJitter,
		rand:        rand.New(rand.NewSource(time.Now().UnixNano())),
		clock:       clock.New(),
	}

	for _, o := range opts {
		o(r)
	}

	if r.maxAttempts < 1 {
		r.maxAttempts = 1
	}
	if r.multiplier < 1 {
		r.multiplier = 1
	}
	if r.baseDelay < 0 {
		r.baseDelay = 0
	}
	if r.maxDelay <= 0 {
		r.maxDelay = defaultMaxDelay
	}
	if r.jitter == nil {
		r.jitter = NoJitter
	}
	if r.clock == nil {
		r.clock = clock.New()
	}
	if r.rand == nil {
		r.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	return r
}

// Retry retries the operation until it succeeds, a non-retryable error occurs,
// context is cancelled, or the attempt limit is reached.
func (r *Retrier) Retry(ctx context.Context, op operation, isRetryable func(error) bool) error {
	if op == nil {
		return errors.New("retrier: nil operation")
	}
	if isRetryable == nil {
		isRetryable = func(error) bool { return true }
	}

	delay := r.baseDelay
	var lastErr error

	for attempt := 1; attempt <= r.maxAttempts; attempt++ {
		// honour cancellation before each attempt
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = op(ctx)
		if lastErr == nil {
			return nil
		}
		if !isRetryable(lastErr) {
			return lastErr
		}
		if attempt == r.maxAttempts {
			return &RetrierError{Attempts: attempt, Err: lastErr}
		}

		nextDelay := r.jitterDelay(delay)
		if nextDelay > r.maxDelay {
			nextDelay = r.jitterDelay(r.maxDelay)
		}

		if r.onRetry != nil {
			r.onRetry(attempt, lastErr, nextDelay)
		}

		if err := r.sleepCtx(ctx, nextDelay); err != nil {
			return err
		}

		// exponential backoff for next iteration, capped
		delay = r.nextDelay(delay)
	}

	return &RetrierError{Attempts: r.maxAttempts, Err: lastErr}
}

// sleepCtx pauses for d or returns early if ctx is cancelled.
func (r *Retrier) sleepCtx(ctx context.Context, d time.Duration) error {
	const step = 10 * time.Millisecond // pas de 10 ms pour FakeClock
	deadline := r.clock.Now().Add(d)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		now := r.clock.Now()
		if !now.Before(deadline) {
			return nil
		}
		rem := min(deadline.Sub(now), step)
		r.clock.Sleep(rem)
		runtime.Gosched()
	}
}

func (r *Retrier) nextDelay(prev time.Duration) time.Duration {
	if prev >= r.maxDelay {
		return r.maxDelay
	}
	nextF := float64(prev) * r.multiplier
	if nextF > float64(r.maxDelay) || nextF > float64(math.MaxInt64) {
		return r.maxDelay
	}
	return time.Duration(nextF)
}

func (r *Retrier) jitterDelay(d time.Duration) time.Duration {
	r.randMu.Lock()
	j := r.jitter(d, r.rand)
	r.randMu.Unlock()
	return j
}
