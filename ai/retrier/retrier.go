package retrier

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"runtime"
	"time"

	crand "crypto/rand"

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

func (e *RetrierError) Unwrap() error {
	return e.Err
}

// jitter defines how we add randomness to a base backoff duration.
type jitter func(time.Duration) time.Duration

var (
	// NoJitter applies a constant delay between retries.
	NoJitter jitter = func(d time.Duration) time.Duration {
		return d
	}

	// FullJitter applies a random delay up to the full backoff interval.
	FullJitter jitter = func(d time.Duration) time.Duration {
		if d <= 0 {
			return 0
		}
		// crypto-secure random in [0, d]
		return secureRandomDuration(int64(d) + 1)
	}

	// EqualJitter applies a random delay between half and the full interval.
	EqualJitter jitter = func(d time.Duration) time.Duration {
		if d <= 0 {
			return 0
		}
		half := d / 2
		return half + secureRandomDuration(int64(half)+1)
	}
)

// operation is the function to retry.
type operation func(context.Context) error

// option customizes a Retrier.
type option func(*Retrier)

// WithMaxAttempts sets how many times an operation will be tried (including the first).
func WithMaxAttempts(n int) option {
	return func(r *Retrier) { r.maxAttempts = n }
}

// WithBaseDelay sets the initial backoff interval.
func WithBaseDelay(d time.Duration) option {
	return func(r *Retrier) { r.baseDelay = d }
}

// WithMultiplier sets the exponential growth factor.
func WithMultiplier(m float64) option {
	return func(r *Retrier) { r.multiplier = m }
}

// WithMaxDelay caps the backoff interval.
func WithMaxDelay(d time.Duration) option {
	return func(r *Retrier) { r.maxDelay = d }
}

// WithJitter chooses how to randomize each interval.
func WithJitter(j jitter) option {
	return func(r *Retrier) { r.jitter = j }
}

// WithClock replaces the time source (useful for testing).
func WithClock(c clock.Clock) option {
	return func(r *Retrier) { r.clock = c }
}

// WithOnRetry registers a hook before each retry attempt.
func WithOnRetry(fn func(attempt int, err error, nextDelay time.Duration)) option {
	return func(r *Retrier) { r.onRetry = fn }
}

const (
	defaultAttempts   = 3
	defaultBaseDelay  = 100 * time.Millisecond
	defaultMultiplier = 2.0
	defaultMaxDelay   = 30 * time.Second
)

// Retrier manages retrying an operation with backoff and jitter.
type Retrier struct {
	maxAttempts int
	baseDelay   time.Duration
	multiplier  float64
	maxDelay    time.Duration
	jitter      jitter
	clock       clock.Clock
	onRetry     func(attempt int, err error, nextDelay time.Duration)
}

// New constructs a Retrier with defaults (3 attempts, 100ms base,
// 2× multiplier, 30s max, FullJitter) and applies any opts.
func New(opts ...option) *Retrier {
	r := &Retrier{
		maxAttempts: defaultAttempts,
		baseDelay:   defaultBaseDelay,
		multiplier:  defaultMultiplier,
		maxDelay:    defaultMaxDelay,
		jitter:      FullJitter,
		clock:       clock.New(),
	}
	for _, o := range opts {
		o(r)
	}
	// sanitize
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
	return r
}

// Retry invokes op until it succeeds, is non-retryable, ctx cancels, or attempts exhausted.
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
		// cancellation check
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

		// compute next delay with jitter and cap
		next := r.jitterDelay(delay)
		if next > r.maxDelay {
			next = r.jitterDelay(r.maxDelay)
		}

		// hook
		if r.onRetry != nil {
			r.onRetry(attempt, lastErr, next)
		}

		// sleep or exit early on cancel
		if err := r.sleepCtx(ctx, next); err != nil {
			return err
		}

		// prepare for next iteration
		delay = r.nextDelay(delay)
	}

	return &RetrierError{Attempts: r.maxAttempts, Err: lastErr}
}

// sleepCtx pauses up to d, but returns early if ctx is done.
func (r *Retrier) sleepCtx(ctx context.Context, d time.Duration) error {
	const step = 10 * time.Millisecond
	deadline := r.clock.Now().Add(d)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		now := r.clock.Now()
		if !now.Before(deadline) {
			return nil
		}
		rem := min(durationUntil(deadline, now), step)
		r.clock.Sleep(rem)
		runtime.Gosched()
	}
}

func durationUntil(deadline, now time.Time) time.Duration {
	return deadline.Sub(now)
}

func (r *Retrier) nextDelay(prev time.Duration) time.Duration {
	if prev >= r.maxDelay {
		return r.maxDelay
	}
	next := float64(prev) * r.multiplier
	if next > float64(r.maxDelay) || next > float64(math.MaxInt64) {
		return r.maxDelay
	}
	return time.Duration(next)
}

func (r *Retrier) jitterDelay(d time.Duration) time.Duration {
	return r.jitter(d)
}

// secureRandomDuration returns a uniform [0, max) Duration via crypto/rand.
func secureRandomDuration(max int64) time.Duration {
	if max <= 0 {
		return 0
	}
	bi := big.NewInt(max)
	ri, err := crand.Int(crand.Reader, bi)
	if err != nil {
		// on error, fall back to no jitter
		return 0
	}
	return time.Duration(ri.Int64())
}
