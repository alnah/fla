package breaker

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// fakeClock implements a controllable clock for testing.
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func newFakeClock(start time.Time) *fakeClock {
	return &fakeClock{now: start}
}

func (f *fakeClock) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

func (f *fakeClock) Add(d time.Duration) {
	f.mu.Lock()
	f.now = f.now.Add(d)
	f.mu.Unlock()
}

func (f *fakeClock) Sleep(d time.Duration) {}

// fakeMetrics collects counts of calls
type fakeMetrics struct {
	tripped  int
	rejected int
	success  int
	failure  int
}

func (f *fakeMetrics) IncTripped()  { f.tripped++ }
func (f *fakeMetrics) IncRejected() { f.rejected++ }
func (f *fakeMetrics) IncSuccess()  { f.success++ }
func (f *fakeMetrics) IncFailure()  { f.failure++ }

func newTestBreaker(clock *fakeClock, metrics Metrics) *breaker {
	return New(Config{
		FailureThreshold: 2,
		SuccessThreshold: 1,
		OpenTimeout:      10 * time.Second,
		FailureWindow:    1 * time.Minute,
		BackoffMaxExp:    2,
		MaxOpenTimeout:   1 * time.Minute,
		Clock:            clock,
		Metrics:          metrics,
	}).(*breaker)
}

func TestExecute_NilOp(t *testing.T) {
	b := newTestBreaker(newFakeClock(time.Now()), nil)
	err := b.Execute(context.Background(), nil)
	if err == nil || err.Error() != "breaker: nil operation" {
		t.Fatalf("want error 'breaker: nil operation', got %v", err)
	}
}

func TestBreaker_TripAndReset(t *testing.T) {
	start := time.Now()
	clk := newFakeClock(start)
	metrics := &fakeMetrics{}
	b := newTestBreaker(clk, metrics)

	// First failure
	err1 := b.Execute(context.Background(), func(_ context.Context) error { return errors.New("fail1") })
	if err1 == nil || err1.Error() != "fail1" {
		t.Fatalf("want fail1, got %v", err1)
	}
	if metrics.failure != 1 {
		t.Errorf("want 1 failure, got %d", metrics.failure)
	}

	// Second failure should trip
	err2 := b.Execute(context.Background(), func(_ context.Context) error { return errors.New("fail2") })
	if err2 == nil || err2.Error() != "fail2" {
		t.Fatalf("want fail2, got %v", err2)
	}
	if metrics.failure != 2 {
		t.Errorf("want 2 failures, got %d", metrics.failure)
	}
	if metrics.tripped != 1 {
		t.Errorf("want 1 trip, got %d", metrics.tripped)
	}

	// Now breaker is open, next call rejected
	err3 := b.Execute(context.Background(), func(_ context.Context) error { return nil })
	if !errors.Is(err3, ErrOpen) {
		t.Errorf("want ErrOpen, got %v", err3)
	}
	if metrics.rejected != 1 {
		t.Errorf("want 1 rejection, got %d", metrics.rejected)
	}

	// Advance time past open timeout
	clk.Add(11 * time.Second)

	// Probe allowed in half-open
	err4 := b.Execute(context.Background(), func(_ context.Context) error { return nil })
	if err4 != nil {
		t.Fatalf("want no error, got %v", err4)
	}
	if metrics.success != 1 {
		t.Errorf("want 1 success, got %d", metrics.success)
	}

	// After success threshold, breaker resets to closed
	err5 := b.Execute(context.Background(), func(_ context.Context) error { return nil })
	if err5 != nil {
		t.Fatalf("want no error, got %v", err5)
	}
	if metrics.success != 2 {
		t.Errorf("want 2 successes, got %d", metrics.success)
	}
}

func TestBreaker_BackoffMaxTimeout(t *testing.T) {
	clk := newFakeClock(time.Now())
	metrics := &fakeMetrics{}
	b := newTestBreaker(clk, metrics)

	// Trip to exceed backoff exp threshold
	// First trip
	_ = b.Execute(context.Background(), func(_ context.Context) error { return errors.New("fail") })
	clk.Add(11 * time.Second)
	_ = b.Execute(context.Background(), func(_ context.Context) error { return errors.New("fail") })
	// Second trip
	clk.Add(11 * time.Second)
	_ = b.Execute(context.Background(), func(_ context.Context) error { return errors.New("fail") })

	// Now in open state; compute current untilOpen duration relative to now
	dur := b.untilOpen.Sub(clk.Now())
	if dur > b.cfg.MaxOpenTimeout {
		t.Errorf("want open duration <= %v, got %v", b.cfg.MaxOpenTimeout, dur)
	}
}

func TestFailureWindowEviction(t *testing.T) {
	start := time.Now()
	clk := newFakeClock(start)
	metrics := &fakeMetrics{}
	b := newTestBreaker(clk, metrics)

	// First failure
	_ = b.Execute(context.Background(), func(_ context.Context) error { return errors.New("fail1") })
	// Advance beyond failure window so it's evicted
	clk.Add(2 * time.Minute)

	// Next failures count as first again
	_ = b.Execute(context.Background(), func(_ context.Context) error { return errors.New("fail2") })
	_ = b.Execute(context.Background(), func(_ context.Context) error { return errors.New("fail3") })

	// Should trip on second since window reset
	if metrics.tripped != 1 {
		t.Errorf("want 1 trip, got %d", metrics.tripped)
	}
}

func TestContextCancelNeutral(t *testing.T) {
	clk := newFakeClock(time.Now())
	b := newTestBreaker(clk, nil)

	// Cancelled context should not count as failure in closed
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := b.Execute(ctx, func(_ context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("want context.Canceled, got %v", err)
	}

	// Should still allow subsequent calls
	err2 := b.Execute(context.Background(), func(_ context.Context) error { return nil })
	if err2 != nil {
		t.Errorf("want no error, got %v", err2)
	}
}

func TestOnStateChangeCalled(t *testing.T) {
	clk := newFakeClock(time.Now())
	// Channel to capture state changes
	ch := make(chan struct{ from, to state }, 3)
	b := New(Config{
		FailureThreshold: 1,
		OpenTimeout:      10 * time.Second,
		FailureWindow:    time.Minute,
		Clock:            clk,
		OnStateChange: func(p, n state) {
			ch <- struct{ from, to state }{p, n}
		},
	}).(*breaker)

	// Trip to open
	_ = b.Execute(context.Background(), func(_ context.Context) error { return errors.New("fail") })
	// Expect Closed->Open
	select {
	case evt := <-ch:
		if evt.from != Closed || evt.to != Open {
			t.Errorf("want (closed->open), got (%v->%v)", evt.from, evt.to)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for open event")
	}

	// Advance to half-open and immediate reset on success
	clk.Add(11 * time.Second)
	_ = b.Execute(context.Background(), func(_ context.Context) error { return nil })
	// Expect Open->HalfOpen then HalfOpen->Closed
	exp := []struct{ from, to state }{{Open, HalfOpen}, {HalfOpen, Closed}}
	for i, wantEvt := range exp {
		select {
		case evt := <-ch:
			if evt.from != wantEvt.from || evt.to != wantEvt.to {
				t.Errorf("[%d] want (%v->%v), got (%v->%v)", i, wantEvt.from, wantEvt.to, evt.from, evt.to)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("timeout waiting for event %d", i)
		}
	}
}
