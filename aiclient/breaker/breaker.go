package breaker

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/alnah/fla/aiclient/clock"
)

// ErrOpen is returned by Execute when the breaker is open and the operation is blocked.
var ErrOpen = errors.New("circuit breaker: circuit is open")

type operation func(ctx context.Context) error

type state struct{ slug string }

func (s state) String() string { return s.slug }

var (
	closed   = state{"closed"}    // normal operation; everything flows
	open     = state{"opened"}    // short‑circuiting; no calls allowed
	halfOpen = state{"half‑open"} // probing; limited calls allowed
)

type settings struct {
	failureThreshold int           // consecutive failures to trip from closed → open
	successThreshold int           // consecutive successes to recover from half‑open → closed
	openTimeout      time.Duration // minimum open period before probing
	clock            clock.Clock   // dependency‑injected time source
	onStateChange    func(prev, next state)
}

// Breaker wraps an operation with failure tracking and state transitions.
// It stops forwarding calls after repeated failures, waits for a configurable timeout,
// then allows limited probes to detect recovery.
type Breaker struct {
	s         settings
	mu        sync.Mutex
	state     state
	failures  int
	successes int
	allowed   int       // remaining probes allowed in half‑open
	openUntil time.Time // gate to leave open state
}

const (
	defaultFailureThreshold = 5
	defaultSuccessThreshold = 1
	defaultOpenTimeout      = 30 * time.Second
)

type option func(*Breaker)

// WithFailureThreshold sets how many consecutive failures trigger
// the breaker to open, blocking further calls.
func WithFailureThreshold(n int) option { return func(b *Breaker) { b.s.failureThreshold = n } }

// WithSuccessThreshold sets how many consecutive successes in half‑open
// state are needed to close the breaker again.
func WithSuccessThreshold(n int) option { return func(b *Breaker) { b.s.successThreshold = n } }

// WithOpenTimeout sets how long the breaker remains open before
// allowing a probe to check for recovery.
func WithOpenTimeout(t time.Duration) option { return func(b *Breaker) { b.s.openTimeout = t } }

// WithClock injects a custom clock for testing or alternative time sources.
func WithClock(c clock.Clock) option { return func(b *Breaker) { b.s.clock = c } }

// WithOnStateChange registers a hook that is called asynchronously whenever
// the breaker transitions to a different state.
func WithOnStateChange(fn func(prev, next state)) option {
	return func(b *Breaker) { b.s.onStateChange = fn }
}

// New constructs a Breaker with sensible defaults and applies any provided options to tune thresholds, timeouts, or the clock.
func New(opts ...option) *Breaker {
	b := &Breaker{
		s: settings{
			failureThreshold: defaultFailureThreshold,
			successThreshold: defaultSuccessThreshold,
			openTimeout:      defaultOpenTimeout,
			clock:            clock.New(),
		},
	}
	for _, o := range opts {
		o(b)
	}
	if b.s.failureThreshold < 1 {
		b.s.failureThreshold = defaultFailureThreshold
	}
	if b.s.successThreshold < 1 {
		b.s.successThreshold = defaultSuccessThreshold
	}
	if b.s.clock == nil {
		b.s.clock = clock.New()
	}

	b.state = closed
	return b
}

// State returns the breaker’s current state (closed, opened, or half‑open) in a thread‑safe manner.
func (b *Breaker) State() state {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

// IsOpen reports whether the breaker is currently in the open state.
func (b *Breaker) IsOpen() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state == open
}

// IsHalfOpen reports whether the breaker is currently in the half‑open state.
func (b *Breaker) IsHalfOpen() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state == halfOpen
}

// IsClosed reports whether the breaker is currently in the closed state.
func (b *Breaker) IsClosed() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state == closed
}

// Execute runs the provided operation if the breaker allows it.
// When the breaker is open, Execute returns ErrOpen immediately.
// Successes reset failure count in closed state; failures count toward tripping the breaker open.
func (b *Breaker) Execute(ctx context.Context, op operation) error {
	if op == nil {
		return errors.New("breaker: nil operation")
	}

	// admission phase
	b.mu.Lock()
	now := b.s.clock.Now()

	switch b.state {
	case open:
		if now.Before(b.openUntil) {
			b.mu.Unlock()
			return ErrOpen
		}
		// first caller after the timeout transitions to half‑open *without* consuming a probe slot
		b.transitionLocked(halfOpen)
		b.failures, b.successes = 0, 0
		b.allowed = b.s.successThreshold
	case halfOpen:
		if b.allowed <= 0 {
			b.mu.Unlock()
			return ErrOpen
		}
		b.allowed-- // consume one probe slot
	case closed:
		// nothing special
	}
	b.mu.Unlock()

	// run operation outside the lock
	err := op(ctx)

	// feedback phase
	b.mu.Lock()
	defer b.mu.Unlock()

	if err == nil { // success path
		switch b.state {
		case halfOpen:
			b.successes++
			if b.successes >= b.s.successThreshold {
				b.transitionLocked(closed)
				b.failures = 0
			}
		case closed:
			b.failures = 0
		}
		return nil
	}

	// failure path
	switch b.state {
	case halfOpen:
		b.tripLocked()
	case closed:
		b.failures++
		if b.failures >= b.s.failureThreshold {
			b.tripLocked()
		}
	}

	return err
}

func (b *Breaker) tripLocked() {
	b.openUntil = b.s.clock.Now().Add(b.s.openTimeout)
	b.transitionLocked(open)
	b.failures, b.successes, b.allowed = 0, 0, 0
}

func (b *Breaker) transitionLocked(next state) {
	if b.state == next {
		return
	}
	prev := b.state
	b.state = next
	if b.s.onStateChange != nil {
		go func() {
			defer func() { _ = recover() }()
			b.s.onStateChange(prev, next)
		}()
	}
}
