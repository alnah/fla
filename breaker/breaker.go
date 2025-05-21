package breaker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/alnah/fla/clock"
)

type state string

// Public breaker states.
const (
	Closed   state = "closed"    // everything flows
	Open     state = "open"      // no calls allowed
	HalfOpen state = "half-open" // limited calls allowed
)

// ErrOpen is returned when the breaker rejects a call while open.
var ErrOpen = errors.New("circuit breaker: is open")

// Metrics exposes optional counters.
type Metrics interface {
	IncTripped()
	IncRejected()
	IncSuccess()
	IncFailure()
}

const (
	defFailureThreshold = 5
	defSuccessThreshold = 1
	defOpenTimeout      = 30 * time.Second
)

type Breaker interface {
	Execute(ctx context.Context, opCtx func(context.Context) error) error
	Success()
	Fail()
}

// breaker wraps an operation with failure tracking and state transitions.
// It stops forwarding calls after repeated failures, waits for a configurable timeout,
// then allows limited probes to detect recovery.
type breaker struct {
	cfg            Config
	mu             sync.RWMutex
	state          state
	failures       []time.Time // failures inside FailureWindow
	successes      int         // successes in HalfOpen
	allowed        int         // remaining probes
	untilOpen      time.Time   // stay-open deadline
	backoffPenalty uint        // current 2^back penalty
}

// New constructs a Breaker with sensible defaults and applies any provided options to tune thresholds, timeouts, or the clock.
func New(cfg Config) Breaker {
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = defFailureThreshold
	}
	if cfg.SuccessThreshold <= 0 {
		cfg.SuccessThreshold = defSuccessThreshold
	}
	if cfg.OpenTimeout <= 0 {
		cfg.OpenTimeout = defOpenTimeout
	}
	if cfg.Clock == nil {
		cfg.Clock = clock.New()
	}
	return &breaker{
		cfg:   cfg,
		state: Closed,
	}
}

func (b *breaker) now() time.Time { return b.cfg.Clock.Now() }

func (b *breaker) emit(prev, next state) {
	if fn := b.cfg.OnStateChange; fn != nil {
		defer func() { _ = recover() }()
		fn(prev, next)
	}
}

func (b *breaker) to(next state) {
	prev := b.state
	if prev == next {
		return
	}
	b.state = next
	b.emit(prev, next)
}

func (b *breaker) evictOld(now time.Time) {
	if b.cfg.FailureWindow == 0 || len(b.failures) == 0 {
		return
	}
	cut := now.Add(-b.cfg.FailureWindow)
	i := 0
	for i < len(b.failures) && b.failures[i].Before(cut) {
		i++
	}
	b.failures = b.failures[i:]
}

// Success records a successful call.
func (b *breaker) Success() {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case HalfOpen:
		b.successes++
		if b.successes >= b.cfg.SuccessThreshold {
			b.reset()
		}
	case Closed:
		b.failures = b.failures[:0]
	}
}

// Fail records a failed call.
func (b *breaker) Fail() {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := b.now()
	switch b.state {
	case HalfOpen:
		b.trip(now)
	case Closed:
		b.evictOld(now)
		b.failures = append(b.failures, now)
		if len(b.failures) >= b.cfg.FailureThreshold {
			b.trip(now)
		}
	}
}

// Execute runs the provided operation if the breaker allows it.
// When the breaker is open, Execute returns ErrOpen immediately.
// Successes reset failure count in closed state; failures count toward tripping the breaker open.
func (b *breaker) Execute(ctx context.Context, op func(context.Context) error) error {
	if op == nil {
		return errors.New("circuit breaker: nil operation")
	}

	// admission
	b.mu.Lock()
	now := b.now()

	switch b.state {
	case Open:
		if now.Before(b.untilOpen) {
			b.mu.Unlock()
			if m := b.cfg.Metrics; m != nil {
				m.IncRejected()
			}
			return ErrOpen
		}
		// first probe
		b.to(HalfOpen)
		b.successes = 0
		b.allowed = b.cfg.SuccessThreshold
	case HalfOpen:
		if b.allowed == 0 {
			b.mu.Unlock()
			if m := b.cfg.Metrics; m != nil {
				m.IncRejected()
			}
			return ErrOpen
		}
		b.allowed--
	}
	b.mu.Unlock()

	// run outside lock
	err := op(ctx)

	// feedback
	b.mu.Lock()
	defer b.mu.Unlock()

	// treat ctx cancellation as neutral unless in HalfOpen
	cancelled := errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)

	switch {
	case err == nil:
		if m := b.cfg.Metrics; m != nil {
			m.IncSuccess()
		}
		switch b.state {
		case HalfOpen:
			b.successes++
			if b.successes >= b.cfg.SuccessThreshold {
				b.reset()
			}
		case Closed:
			b.failures = b.failures[:0]
		}
		return nil

	case cancelled && b.state != HalfOpen:
		return fmt.Errorf("circuit breaker: %w", err)

	default: // failure
		if m := b.cfg.Metrics; m != nil {
			m.IncFailure()
		}
		n := b.now()

		switch b.state {
		case HalfOpen:
			b.trip(n)
		case Closed:
			b.evictOld(n)
			b.failures = append(b.failures, n)
			if len(b.failures) >= b.cfg.FailureThreshold {
				b.trip(n)
			}
		}
		return fmt.Errorf("circuit breaker: %w", err)
	}
}

func (b *breaker) trip(now time.Time) {
	dur := b.cfg.OpenTimeout
	if b.backoffPenalty < b.cfg.BackoffMaxExp {
		dur <<= b.backoffPenalty
	}
	if b.cfg.MaxOpenTimeout > 0 && dur > b.cfg.MaxOpenTimeout {
		dur = b.cfg.MaxOpenTimeout
	}
	b.backoffPenalty++
	b.untilOpen = now.Add(dur)
	b.failures = b.failures[:0]
	b.successes = 0
	b.allowed = 0
	b.to(Open)
	if m := b.cfg.Metrics; m != nil {
		m.IncTripped()
	}
}

func (b *breaker) reset() {
	b.backoffPenalty = 0
	b.failures = b.failures[:0]
	b.successes = 0
	b.allowed = 0
	b.to(Closed)
}
