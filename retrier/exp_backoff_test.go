package retrier

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestRetrier_SucceedsFirstTry(t *testing.T) {
	r := NewExpBackoffJitter()
	if err := r.Retry(context.Background(), func(context.Context) error { return nil }, nil); err != nil {
		t.Fatalf("want no error, got %v", err)
	}
}

func TestRetrier_EventualSuccess(t *testing.T) {
	fake := NewFakeClock(time.Now())
	attempts := 0
	r := NewExpBackoffJitter(
		WithMaxAttempts(5),
		WithClock(fake),
		WithBaseDelay(10*time.Millisecond),
		WithJitter(NoJitter),
	)

	err := r.Retry(context.Background(), func(context.Context) error {
		attempts++
		if attempts < 3 {
			return errors.New("fail")
		}
		return nil
	}, nil)

	if err != nil {
		t.Fatalf("error: want success, got %v", err)
	}
	if attempts != 3 {
		t.Fatalf("error: want 3 attempts, got %d", attempts)
	}
}

func TestRetrier_StopsOnNonRetryable(t *testing.T) {
	fake := NewFakeClock(time.Now())
	nonRetryable := errors.New("nope")
	attempts := 0
	r := NewExpBackoffJitter(
		WithMaxAttempts(5),
		WithClock(fake),
	)

	err := r.Retry(context.Background(), func(context.Context) error {
		attempts++
		return nonRetryable
	}, func(err error) bool { return false })

	if !errors.Is(err, nonRetryable) {
		t.Fatalf("want non-retryable, got %v", err)
	}
	if attempts != 1 {
		t.Errorf("want 1 attempt, got %d", attempts)
	}
}

func TestRetrier_ExhaustsAttempts(t *testing.T) {
	fake := NewFakeClock(time.Now())
	r := NewExpBackoffJitter(
		WithMaxAttempts(3),
		WithClock(fake),
		WithBaseDelay(1*time.Millisecond),
		WithJitter(NoJitter),
	)

	err := r.Retry(context.Background(), func(context.Context) error {
		return errors.New("always fail")
	}, nil)
	var re *RetrierError
	if !errors.As(err, &re) {
		t.Fatalf("want RetrierError, got %v", err)
	}
	if re.attempts != 3 {
		t.Errorf("want 3 attempts, got %d", re.attempts)
	}
}

func TestRetrier_ContextCancel(t *testing.T) {
	fake := NewFakeClock(time.Now())
	r := NewExpBackoffJitter(
		WithMaxAttempts(5),
		WithClock(fake),
		WithBaseDelay(1*time.Second),
	)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// simulate time passing before cancel
		fake.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := r.Retry(ctx, func(context.Context) error {
		return errors.New("fail")
	}, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("want context.Canceled, got %v", err)
	}
}

func TestRetry_NilOperation(t *testing.T) {
	r := NewExpBackoffJitter()
	err := r.Retry(context.Background(), nil, nil)
	if err == nil || err.Error() != "retrier: nil operation" {
		t.Errorf("expected nil-op error, got %v", err)
	}
}

func TestNew_InvalidOptions(t *testing.T) {
	r := NewExpBackoffJitter(
		WithMaxAttempts(0),
		WithMultiplier(0.5),
		WithBaseDelay(-1),
		WithMaxDelay(0),
		WithJitter(nil),
		WithClock(nil),
	)
	if r.maxAttempts != 1 {
		t.Errorf("want maxAttempts=1, got %d", r.maxAttempts)
	}
	if r.multiplier != 1 {
		t.Errorf("want multiplier=1.0, got %.1f", r.multiplier)
	}
	if r.baseDelay != 0 {
		t.Errorf("want baseDelay=0, got %v", r.baseDelay)
	}
	if r.maxDelay != defMaxDelay {
		t.Errorf("want maxDelay=%v, got %v", defMaxDelay, r.maxDelay)
	}
	if r.jitter == nil {
		t.Error("jitter should not be nil")
	}
	if r.clock == nil {
		t.Error("clock should not be nil")
	}
}

func TestRetry_OnRetryHook(t *testing.T) {
	fake := NewFakeClock(time.Now())
	var calls []struct {
		attempt   int
		errMsg    string
		nextDelay time.Duration
	}
	r := NewExpBackoffJitter(
		WithMaxAttempts(3),
		WithClock(fake),
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
	op := func(context.Context) error { return errors.New("E") }
	err := r.Retry(context.Background(), op, nil)
	var re *RetrierError
	if !errors.As(err, &re) {
		t.Fatalf("want RetrierError, got %v", err)
	}
	want := []time.Duration{10 * time.Millisecond, 20 * time.Millisecond}
	if len(calls) != len(want) {
		t.Fatalf("want %d OnRetry calls, got %d", len(want), len(calls))
	}
	for i, c := range calls {
		if c.attempt != i+1 {
			t.Errorf("call %d: want attempt %d, got %d", i, i+1, c.attempt)
		}
		if c.nextDelay != want[i] {
			t.Errorf("call %d: want delay %v, got %v", i, want[i], c.nextDelay)
		}
		if c.errMsg != "E" {
			t.Errorf("call %d: want errMsg=E, got %q", i, c.errMsg)
		}
	}
}

func TestNextDelay_TableDriven(t *testing.T) {
	r := NewExpBackoffJitter(WithMaxDelay(50*time.Millisecond), WithMultiplier(3))
	tests := []struct {
		prev, want time.Duration
	}{
		{10 * time.Millisecond, 30 * time.Millisecond},
		{30 * time.Millisecond, 50 * time.Millisecond}, // cap at maxDelay
		{100 * time.Millisecond, 50 * time.Millisecond},
	}
	for _, tc := range tests {
		if got := r.nextDelay(tc.prev); got != tc.want {
			t.Errorf("nextDelay(%v): want %v, got %v", tc.prev, tc.want, got)
		}
	}
}

func TestJitterFunctions(t *testing.T) {
	base := 100 * time.Millisecond

	if d := NoJitter(base); d != base {
		t.Errorf("NoJitter: want %v, got %v", base, d)
	}
	for i := range 10 {
		d := FullJitter(base)
		if d < 0 || d > base {
			t.Errorf("FullJitter #%d out of range: %v", i, d)
		}
	}
	half := base / 2
	for i := range 10 {
		d := EqualJitter(base)
		if d < half || d > base {
			t.Errorf("EqualJitter #%d out of range: %v", i, d)
		}
	}
}

func TestRetry_ContextAlreadyCancelled(t *testing.T) {
	fake := NewFakeClock(time.Now())
	r := NewExpBackoffJitter(WithClock(fake))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := r.Retry(ctx, func(context.Context) error { return nil }, nil)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("want context.Canceled, got %v", err)
	}
}

func TestSleepCtx_CancelMidSleep(t *testing.T) {
	r := NewExpBackoffJitter()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	err := r.sleepCtx(ctx, 100*time.Millisecond)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("sleepCtx: want Canceled, got %v", err)
	}
	if d := time.Since(start); d >= 100*time.Millisecond {
		t.Fatalf("sleepCtx: did not stop early, waited %v", d)
	}
}

func TestRetry_Concurrent(t *testing.T) {
	fake := NewFakeClock(time.Now())
	r := NewExpBackoffJitter(
		WithMaxAttempts(5),
		WithClock(fake),
		WithBaseDelay(10*time.Millisecond),
		WithJitter(NoJitter),
	)
	var wg sync.WaitGroup
	op := func(ctx context.Context) error { return errors.New("fail") }

	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = r.Retry(context.Background(), op, nil)
		}()
	}

	// advance through all retries
	fake.Sleep(200 * time.Millisecond)
	wg.Wait()
}

type FakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func NewFakeClock(t time.Time) *FakeClock {
	return &FakeClock{now: t}
}

func (f *FakeClock) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

func (f *FakeClock) Sleep(d time.Duration) {
	f.mu.Lock()
	f.now = f.now.Add(d)
	f.mu.Unlock()
}
