package breaker

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alnah/fla/clock"
)

/********* Unit Test *********/

func TestExecute_NilOperation(t *testing.T) {
	b := New()
	err := b.Execute(context.Background(), nil)
	if err == nil || err.Error() != "breaker: nil operation" {
		t.Errorf("want breaker: nil operation, got %v", err)
	}
}

func TestThresholdValidation_Defaults(t *testing.T) {
	b := New(WithFailureThreshold(0), WithSuccessThreshold(0), WithClock(nil))

	if b.s.failureThreshold != defaultFailureThreshold {
		t.Errorf("want failure threshold %d, got %d", defaultFailureThreshold, b.s.failureThreshold)
	}
	if b.s.successThreshold != defaultSuccessThreshold {
		t.Errorf("want success threshold %d, got %d", defaultSuccessThreshold, b.s.successThreshold)
	}
	if b.s.clock == nil {
		t.Error("want non-nil clock, got nil")
	}
}

func TestClosed_ResetFailuresOnSuccess(t *testing.T) {
	fc := clock.NewFakeClock(time.Now())
	b := New(WithFailureThreshold(3), WithClock(fc))

	// first call fails
	_ = b.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("fail")
	})

	// second call succeeds → failure count should reset
	if err := b.Execute(context.Background(), func(ctx context.Context) error { return nil }); err != nil {
		t.Errorf("want nil, got %v", err)
	}

	// up to threshold-1 failures again should still be closed
	for range 2 {
		_ = b.Execute(context.Background(), func(ctx context.Context) error {
			return errors.New("f")
		})
	}
	if !b.IsClosed() {
		t.Errorf("want closed, got %q", b.State())
	}
}

func TestConcurrentExecutions(t *testing.T) {
	fc := clock.NewFakeClock(time.Now())
	b := New(WithFailureThreshold(100), WithClock(fc))

	var wg sync.WaitGroup
	var mu sync.Mutex
	successes, failures := 0, 0

	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := b.Execute(context.Background(), func(ctx context.Context) error {
				time.Sleep(1 * time.Millisecond)
				return nil
			})
			mu.Lock()
			if err != nil {
				failures++
			} else {
				successes++
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	if failures != 0 {
		t.Errorf("want 0 failures, got %d", failures)
	}
	if successes != 50 {
		t.Errorf("want 50 successes, got %d", successes)
	}
}

func TestBreaker_TripsOpenAfterFailures(t *testing.T) {
	// configure breaker to open after 2 failures
	start := time.Now()
	fc := clock.NewFakeClock(start)
	b := New(
		WithFailureThreshold(2),
		WithOpenTimeout(5*time.Second),
		WithClock(fc),
	)

	err1 := b.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("failure1")
	})
	if err1 == nil || err1.Error() != "failure1" {
		t.Errorf("error first execute: want \"failure1\", got %v", err1)
	}

	err2 := b.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("failure2")
	})
	if err2 == nil || err2.Error() != "failure2" {
		t.Errorf("error first execute: want \"failure2\", got %v", err2)
	}

	err3 := b.Execute(context.Background(), func(ctx context.Context) error { return nil })
	if !errors.Is(err3, ErrOpen) {
		t.Errorf("error first execute: want ErrOpen, got %v", err3)
	}
	if !b.IsOpen() {
		t.Errorf("breaker state: want \"open\", got %q", b.State())
	}
}

func TestBreaker_TransitionsToOpenAfterFailures(t *testing.T) {
	startTime := time.Date(2025, 5, 6, 12, 0, 0, 0, time.UTC)
	fakeClock := clock.NewFakeClock(startTime)

	b := New(
		WithFailureThreshold(3),
		WithOpenTimeout(10*time.Second),
		WithClock(fakeClock),
	)
	for i := range 3 {
		err := b.Execute(context.Background(), func(ctx context.Context) error {
			return errors.New("simulated failure")
		})
		if err == nil {
			t.Fatalf("error: want failure on iteration %d, got nil", i)
		}
	}

	if !b.IsOpen() {
		t.Errorf("breaker state: want \"open\", got %q", b.State())
	}

	err := b.Execute(context.Background(), func(ctx context.Context) error { return nil })
	if !errors.Is(err, ErrOpen) {
		t.Errorf("error: want ErrOpen, got %v", err)
	}
}

func TestBreaker_BlocksWhileOpen(t *testing.T) {
	// configure breaker to open after 1 failure
	start := time.Now()
	fc := clock.NewFakeClock(start)
	b := New(
		WithFailureThreshold(1),
		WithOpenTimeout(10*time.Second),
		WithClock(fc),
	)

	// trip to open
	_ = b.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("err")
	})
	if !b.IsOpen() {
		t.Fatalf("breaker state: want \"open\", got %q", b.State())
	}

	called := false
	err := b.Execute(context.Background(), func(ctx context.Context) error {
		called = true
		return nil
	})
	if !errors.Is(err, ErrOpen) {
		t.Errorf("error first execute: want ErrOpen, got %v", err)
	}
	if called {
		t.Errorf("error: operation should not be called while open")
	}
}

func TestHalfOpen_FailedProbeTripsBackToOpen(t *testing.T) {
	start := time.Now()
	fc := clock.NewFakeClock(start)
	b := New(
		WithFailureThreshold(1),
		WithSuccessThreshold(1),
		WithOpenTimeout(1*time.Second),
		WithClock(fc),
	)

	// trip to open
	_ = b.Execute(context.Background(), func(ctx context.Context) error { return errors.New("boom1") })
	if !b.IsOpen() {
		t.Fatalf("want open, got %q", b.State())
	}

	// advance past timeout into half-open
	fc.Sleep(2 * time.Second)

	// failed probe should return the error and trip back to open
	err := b.Execute(context.Background(), func(ctx context.Context) error { return errors.New("boom2") })
	if err == nil || err.Error() != "boom2" {
		t.Errorf("want \"boom2\", got %v", err)
	}
	if !b.IsOpen() {
		t.Errorf("want open, got %q", b.State())
	}
}

func TestBreaker_HalfOpenAfterTimeout(t *testing.T) {
	// configure breaker to open after 1 failure and close after 1 success in half-open
	start := time.Now()
	fc := clock.NewFakeClock(start)
	b := New(
		WithFailureThreshold(1),
		WithSuccessThreshold(1),
		WithOpenTimeout(3*time.Second),
		WithClock(fc),
	)

	// trip to open
	_ = b.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("boom")
	})
	if !b.IsOpen() {
		t.Fatalf("breaker state: want \"open\", got %q", b.State())
	}

	// move past timeout
	fc.Sleep(4 * time.Second)

	// execute in half-open
	err := b.Execute(context.Background(), func(ctx context.Context) error { return nil })
	if err != nil {
		t.Errorf("error: Execute in half-open: want nil, got %v", err)
	}
	if !b.IsClosed() {
		t.Errorf("breaker state: want \"closed\", got %q", b.State())
	}
}

func TestBreaker_TransitionsToHalfOpenAndBackToClosed(t *testing.T) {
	startTime := time.Date(2025, 5, 6, 12, 0, 0, 0, time.UTC)
	fakeClock := clock.NewFakeClock(startTime)

	b := New(
		WithFailureThreshold(1),
		WithSuccessThreshold(2),
		WithOpenTimeout(5*time.Second),
		WithClock(fakeClock),
	)

	// trip to open
	_ = b.Execute(context.Background(), func(ctx context.Context) error { return errors.New("boom") })

	if !b.IsOpen() {
		t.Fatalf("breaker state: want \"open\", got %q", b.State())
	}

	// advance past open timeout
	fakeClock.Sleep(6 * time.Second)

	// first successful probe
	if err := b.Execute(context.Background(), func(ctx context.Context) error { return nil }); err != nil {
		t.Fatalf("error: first probe: want no error, got %v", err)
	}
	if !b.IsHalfOpen() {
		t.Errorf("breaker state: want \"half-open\", got %q", b.State())
	}

	// second successful probe -> closed
	if err := b.Execute(context.Background(), func(ctx context.Context) error { return nil }); err != nil {
		t.Fatalf("error: second probe: want no error, got %v", err)
	}
	if !b.IsClosed() {
		t.Errorf("breaker state: want \"closed\", got %q", b.State())
	}
}

func TestOnStateChange_Called(t *testing.T) {
	fc := clock.NewFakeClock(time.Now())
	var mu sync.Mutex
	calls := []string{}
	onChange := func(prev, next state) {
		mu.Lock()
		calls = append(calls, prev.String()+"->"+next.String())
		mu.Unlock()
	}

	b := New(
		WithFailureThreshold(1),
		WithSuccessThreshold(1),
		WithOpenTimeout(1*time.Second),
		WithClock(fc),
		WithOnStateChange(onChange),
	)

	// closed -> opened
	_ = b.Execute(context.Background(), func(ctx context.Context) error { return errors.New("fail") })
	// advance past open timeout
	fc.Sleep(2 * time.Second)
	// opened -> half-open
	_ = b.Execute(context.Background(), func(ctx context.Context) error { return nil })
	// half-open -> closed
	_ = b.Execute(context.Background(), func(ctx context.Context) error { return nil })

	// give the async hooks a moment
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	want := map[string]bool{
		closed.String() + "->" + open.String():     true,
		open.String() + "->" + halfOpen.String():   true,
		halfOpen.String() + "->" + closed.String(): true,
	}
	got := make(map[string]bool)
	for _, c := range calls {
		got[c] = true
	}
	for tran := range want {
		if !got[tran] {
			t.Errorf("want transition %q called, but it was not", tran)
		}
	}
}

/********* Integration Tests *********/

func TestBreakerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	var failMode atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if failMode.Load() {
			http.Error(w, "simulated failure", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	startTime := time.Date(2025, 5, 6, 12, 0, 0, 0, time.UTC)
	fakeClock := clock.NewFakeClock(startTime)

	br := New(
		WithFailureThreshold(3),
		WithSuccessThreshold(2),
		WithOpenTimeout(10*time.Second),
		WithClock(fakeClock),
	)

	op := func(ctx context.Context) error {
		resp, err := http.Get(server.URL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 500 {
			return errors.New("server error")
		}
		return nil
	}

	// trip to open
	failMode.Store(true)
	for range 3 {
		_ = br.Execute(context.Background(), op)
	}

	if !br.IsOpen() {
		t.Fatalf("breaker state: want \"open\", got %q", br.State())
	}

	// recover and close
	fakeClock.Sleep(11 * time.Second)
	failMode.Store(false)

	for i := range 2 {
		if err := br.Execute(context.Background(), op); err != nil {
			t.Fatalf("error: iteration %d: want no error, got %v", i, err)
		}
	}

	if br.IsOpen() || br.IsHalfOpen() {
		t.Fatalf("breaker state: want \"closed\", got %q", br.State())
	}
}
