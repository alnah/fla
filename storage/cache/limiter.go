package cache

import (
	"context"

	"github.com/alnah/fla/breaker"
)

// Limiter adapts the breaker to the go-redis Limiter interface.
type Limiter struct{ br breaker.Breaker }

// NewLimiter builds a compliant go-redis Limiter.
func NewLimiter(br breaker.Breaker) *Limiter {
	return &Limiter{br}
}

var defaultRedisLimiter = NewLimiter(breaker.New(breaker.HighQPSConfig()))

// DefaultRedisLimiter complies with the Limiter interface and use the high QPS configuration
// of the circuit breaker package.
func DefaultRedisLimiter() *Limiter { return defaultRedisLimiter }

// Allow verifies if the breaker authorizes a new request.
func (r Limiter) Allow() error {
	// no-op to test admission: closed -> ok; open -> ErrOpen
	err := r.br.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	// if ErrOpen, refuse op
	return err
}

// ReportResult signals a success or a failure for an operation.
func (r Limiter) ReportResult(res error) {
	if res != nil {
		r.br.Fail()
	} else {
		r.br.Success()
	}

}
