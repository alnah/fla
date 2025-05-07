package breaker

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alnah/fla/ai/clock"
)

func TestBreaker_TransitionsToOpenAfterFailures(t *testing.T) {
	startTime := time.Date(2025, 5, 6, 12, 0, 0, 0, time.UTC)
	fakeClock := clock.NewFakeClock(startTime)

	b := New(WithFailureThreshold(3), WithOpenTimeout(10*time.Second), WithClock(fakeClock))

	// simulate consecutive failures
	for i := range 3 {
		err := b.Execute(context.Background(), func(ctx context.Context) error {
			return errors.New("simulated failure")
		})
		if err == nil {
			t.Fatalf("expected error on iteration %d", i)
		}
	}

	if !b.IsOpen() {
		t.Errorf("expected state to be 'opened', got '%s'", b.State())
	}

	// ensure Execute returns ErrOpen while still within open window
	err := b.Execute(context.Background(), func(ctx context.Context) error { return nil })
	if !errors.Is(err, ErrOpen) {
		t.Errorf("expected ErrOpen, got %v", err)
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
		t.Fatalf("expected breaker to be open, got %s", b.State())
	}

	// advance beyond open timeout; breaker should move to half‑open on next call
	fakeClock.Sleep(6 * time.Second)

	// first successful probe (allowed untouched at transition)
	if err := b.Execute(context.Background(), func(ctx context.Context) error { return nil }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !b.IsHalfOpen() {
		t.Fatalf("expected breaker to be half‑open after first probe, got %s", b.State())
	}

	// second successful probe → closed
	if err := b.Execute(context.Background(), func(ctx context.Context) error { return nil }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := b.State().String(); got != "closed" {
		t.Errorf("expected state 'closed', got %s", got)
	}
}

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

	// fail to open breaker
	failMode.Store(true)
	for range 3 {
		_ = br.Execute(context.Background(), op)
	}

	if !br.IsOpen() {
		t.Fatalf("expected breaker to be open after failures")
	}

	// wait and switch to success path
	fakeClock.Sleep(11 * time.Second)
	failMode.Store(false)

	// two successes close the breaker
	for i := range 2 {
		if err := br.Execute(context.Background(), op); err != nil {
			t.Fatalf("iteration %d: unexpected error %v", i, err)
		}
	}

	if br.IsOpen() || br.IsHalfOpen() {
		t.Fatalf("expected breaker to be closed, got %s", br.State())
	}
}

func TestBreaker_TripsOpenAfterFailures(t *testing.T) {
	// configure le breaker pour qu'il bascule en open après 2 échecs
	start := time.Now()
	fc := clock.NewFakeClock(start)
	b := New(
		WithFailureThreshold(2),
		WithOpenTimeout(5*time.Second),
		WithClock(fc),
	)

	// failure returned
	err1 := b.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("failure1")
	})
	if err1 == nil || err1.Error() != "failure1" {
		t.Errorf("first Execute: got %v, want failure1 error", err1)
	}

	// second failure returned
	err2 := b.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("failure2")
	})
	if err2 == nil || err2.Error() != "failure2" {
		t.Errorf("second Execute: got %v, want failure2 error", err2)
	}

	// breaker is open => ErrOpen
	err3 := b.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	if !errors.Is(err3, ErrOpen) {
		t.Errorf("third Execute: got %v, want ErrOpen", err3)
	}
	if !b.IsOpen() {
		t.Errorf("breaker state: got %s, want open", b.State())
	}
}

func TestBreaker_HalfOpenAfterTimeout(t *testing.T) {
	// configure pour qu'il bascule après 1 échec, puis redevienne closed après 1 succès en half-open
	start := time.Now()
	fc := clock.NewFakeClock(start)
	b := New(
		WithFailureThreshold(1),
		WithSuccessThreshold(1),
		WithOpenTimeout(3*time.Second),
		WithClock(fc),
	)

	// trip open
	_ = b.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("boom")
	})
	if !b.IsOpen() {
		t.Fatal("breaker should be open after failure")
	}

	// move forward in time beyond the timeout
	fc.Sleep(4 * time.Second)

	// first execution in half-open state
	err := b.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})

	// must execute, then breaker is closed
	if err != nil {
		t.Errorf("Execute in half-open: got %v, want nil", err)
	}
	if !b.IsClosed() {
		t.Errorf("breaker state: got %s, want closed", b.State())
	}
}

func TestBreaker_BlocksWhileOpen(t *testing.T) {
	// breaker is open
	start := time.Now()
	fc := clock.NewFakeClock(start)
	b := New(
		WithFailureThreshold(1),
		WithOpenTimeout(10*time.Second),
		WithClock(fc),
	)

	_ = b.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("err")
	})
	if !b.IsOpen() {
		t.Fatal("breaker should be open")
	}

	// no-op, error
	called := false
	err := b.Execute(context.Background(), func(ctx context.Context) error {
		called = true
		return nil
	})
	if !errors.Is(err, ErrOpen) {
		t.Errorf("blocked Execute: got %v, want ErrOpen", err)
	}
	if called {
		t.Error("operation should not have been called while open")
	}
}
