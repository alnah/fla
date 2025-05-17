package breaker

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alnah/fla/clock"
)

func TestBreakerIntegration(t *testing.T) {
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
