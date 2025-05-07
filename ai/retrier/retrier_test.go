package retrier

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/alnah/fla/ai/clock"
)

func TestRetrier_SucceedsFirstTry(t *testing.T) {
	r := New()

	if err := r.Retry(context.Background(), func(context.Context) error { return nil }, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRetrier_EventualSuccess(t *testing.T) {
	fakeClock := clock.NewFakeClock(time.Now())
	attempts := 0
	r := New(
		WithMaxAttempts(5),
		WithClock(fakeClock),
		WithBaseDelay(10*time.Millisecond),
		WithJitter(NoJitter), // deterministic
	)

	err := r.Retry(context.Background(), func(context.Context) error {
		attempts++
		if attempts < 3 {
			return errors.New("fail")
		}
		return nil
	}, nil)

	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetrier_StopsOnNonRetryable(t *testing.T) {
	fakeClock := clock.NewFakeClock(time.Now())
	r := New(
		WithMaxAttempts(5),
		WithClock(fakeClock),
	)

	nonRetryable := errors.New("non-retryable")
	attempts := 0
	err := r.Retry(context.Background(), func(context.Context) error {
		attempts++
		return nonRetryable
	}, func(err error) bool { return false })

	if !errors.Is(err, nonRetryable) {
		t.Fatalf("expected non-retryable error, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
}

func TestRetrier_ExhaustsAttempts(t *testing.T) {
	fakeClock := clock.NewFakeClock(time.Now())
	r := New(
		WithMaxAttempts(3),
		WithClock(fakeClock),
		WithBaseDelay(1*time.Millisecond),
		WithJitter(NoJitter),
	)

	err := r.Retry(context.Background(), func(context.Context) error { return errors.New("fail") }, nil)
	var re *RetrierError
	if !errors.As(err, &re) {
		t.Fatalf("expected RetrierError, got %v", err)
	}
	if re.Attempts != 3 {
		t.Fatalf("expected 3 attempts recorded, got %d", re.Attempts)
	}
}

func TestRetrier_ContextCancel(t *testing.T) {
	fakeClock := clock.NewFakeClock(time.Now())
	r := New(
		WithMaxAttempts(5),
		WithClock(fakeClock),
		WithBaseDelay(1*time.Second),
	)

	ctx, cancel := context.WithCancel(context.Background())

	// cancel after first failure before sleep returns
	go func() {
		fakeClock.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := r.Retry(ctx, func(context.Context) error { return errors.New("fail") }, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestRetry_NilOperation(t *testing.T) {
	r := New()
	err := r.Retry(context.Background(), nil, nil)
	if err == nil || err.Error() != "retrier: nil operation" {
		t.Errorf("got %v, want nil-op error", err)
	}
}

// Test New with invalid options resets to minimums
func TestNew_InvalidOptions(t *testing.T) {
	r := New(
		WithMaxAttempts(0),
		WithMultiplier(0.5),
		WithBaseDelay(-1),
		WithMaxDelay(0),
		WithJitter(nil),
		WithRand(nil),
		WithClock(nil),
	)
	// maxAttempts <1 resets to 1 (minimum), not defaultAttempts
	if r.maxAttempts != 1 {
		t.Errorf("maxAttempts=%d, want %d", r.maxAttempts, 1)
	}
	if r.multiplier != 1 {
		t.Errorf("multiplier=%.1f, want 1.0", r.multiplier)
	}
	if r.baseDelay != 0 {
		t.Errorf("baseDelay=%v, want 0", r.baseDelay)
	}
	if r.maxDelay != defaultMaxDelay {
		t.Errorf("maxDelay=%v, want %v", r.maxDelay, defaultMaxDelay)
	}
	if r.jitter == nil {
		t.Errorf("jitter should be NoJitter, got nil")
	}
	if r.clock == nil {
		t.Errorf("clock should be non-nil")
	}
	if r.rand == nil {
		t.Errorf("rand should be non-nil")
	}
}

func TestRetry_OnRetryHook(t *testing.T) {
	fc := clock.NewFakeClock(time.Now())
	var calls []struct {
		attempt   int
		errMsg    string
		nextDelay time.Duration
	}
	r := New(
		WithMaxAttempts(3),
		WithClock(fc),
		WithBaseDelay(10*time.Millisecond),
		WithMultiplier(2),
		WithMaxDelay(100*time.Millisecond),
		WithJitter(NoJitter),
		WithOnRetry(func(attempt int, err error, nextDelay time.Duration) {
			calls = append(calls, struct {
				attempt   int
				errMsg    string
				nextDelay time.Duration
			}{attempt, err.Error(), nextDelay})
		}),
	)
	ctx := context.Background()
	// operation always fails
	op := func(context.Context) error { return errors.New("E") }
	err := r.Retry(ctx, op, nil)
	var re *RetrierError
	if !errors.As(err, &re) {
		t.Fatalf("want RetrierError, got %v", err)
	}
	// onRetry called maxAttempts-1 times
	if len(calls) != r.maxAttempts-1 {
		t.Errorf("got %d calls, want %d", len(calls), r.maxAttempts-1)
	}
	// check delays: 10ms, 20ms
	expected := []time.Duration{10 * time.Millisecond, 20 * time.Millisecond}
	for i, c := range calls {
		if c.attempt != i+1 {
			t.Errorf("call %d: attempt=%d, want %d", i, c.attempt, i+1)
		}
		if c.nextDelay != expected[i] {
			t.Errorf("call %d: nextDelay=%v, want %v", i, c.nextDelay, expected[i])
		}
		if c.errMsg != "E" {
			t.Errorf("call %d: err=%q, want %q", i, c.errMsg, "E")
		}
	}
}

func TestNextDelay_TableDriven(t *testing.T) {
	r := New(WithMaxDelay(50*time.Millisecond), WithMultiplier(3))
	tests := []struct {
		prev, want time.Duration
	}{
		{prev: 10 * time.Millisecond, want: 30 * time.Millisecond},
		{prev: 30 * time.Millisecond, want: 50 * time.Millisecond}, // capped
		{prev: 100 * time.Millisecond, want: 50 * time.Millisecond},
	}
	for _, tc := range tests {
		got := r.nextDelay(tc.prev)
		if got != tc.want {
			t.Errorf("nextDelay(%v)=%v, want %v", tc.prev, got, tc.want)
		}
	}
}

func TestJitters(t *testing.T) {
	seed := int64(42)
	src := rand.NewSource(seed)
	r := New(WithRand(src), WithClock(clock.NewFakeClock(time.Now())))
	base := 100 * time.Millisecond

	// FullJitter ≤ base
	d1 := FullJitter(base, r.rand)
	if d1 < 0 || d1 > base {
		t.Errorf("FullJitter out of range: %v", d1)
	}
	// EqualJitter ∈ [base/2, base]
	d2 := EqualJitter(base, r.rand)
	if d2 < base/2 || d2 > base {
		t.Errorf("EqualJitter out of range: %v", d2)
	}
	// NoJitter == base
	d3 := NoJitter(base, r.rand)
	if d3 != base {
		t.Errorf("NoJitter=%v, want %v", d3, base)
	}
}

func TestRetry_ContextAlreadyCancelled(t *testing.T) {
	r := New(WithClock(clock.NewFakeClock(time.Now())))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := r.Retry(ctx, func(context.Context) error { return nil }, nil)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("got %v, want context.Canceled", err)
	}
}

func TestSleepCtx_CancelMidSleep(t *testing.T) {
	fc := clock.NewFakeClock(time.Now())
	r := New(WithClock(fc))
	ctx, cancel := context.WithCancel(context.Background())
	// schedule cancel after 30ms
	go func() {
		// advance clock by 30ms to trigger mid-sleep cancellation
		fc.Sleep(30 * time.Millisecond)
		cancel()
	}()
	start := fc.Now()
	err := r.sleepCtx(ctx, 100*time.Millisecond)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("sleepCtx returned %v, want context.Canceled", err)
	}
	if fc.Now().Sub(start) >= 100*time.Millisecond {
		t.Errorf("sleepCtx did not stop early, waited %v", fc.Now().Sub(start))
	}
}

func TestRetry_Concurrent(t *testing.T) {
	fc := clock.NewFakeClock(time.Now())
	r := New(
		WithMaxAttempts(5),
		WithClock(fc),
		WithBaseDelay(10*time.Millisecond),
		WithJitter(NoJitter),
	)
	var wg sync.WaitGroup
	op := func(ctx context.Context) error {
		return errors.New("fail")
	}
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = r.Retry(context.Background(), op, nil)
		}()
	}
	// advance clock enough
	fc.Sleep(100 * time.Millisecond)
	wg.Wait()
}
