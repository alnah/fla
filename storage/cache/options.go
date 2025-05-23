package cache

import (
	"time"

	"github.com/alnah/fla/breaker"
	"github.com/alnah/fla/logger"
	"github.com/alnah/fla/storage"
)

// WithLogger sets the Logger used by RedisCache to l.
// This overrides the default logger so that all RedisCache logs are sent to the provided logger instance.
func (b *CacheBuilder) WithLogger(l logger.Logger) *CacheBuilder {
	b.cfg.Logger = l
	return b
}

// WithAddress sets the Redis server address to s.
// The address should be in the form "host:port" (e.g. "localhost:6379") and must be non-empty.
func (b *CacheBuilder) WithAddress(s string) *CacheBuilder {
	b.cfg.Addr, _ = storage.ParseAddress(s) // don't check err because it's going be validated afterward anyway
	return b
}

// WithPassword sets the password for the Redis connection to s.
// If the server requires AUTH, this value must match the server’s requirepass setting.
func (b *CacheBuilder) WithPassword(s string) *CacheBuilder {
	b.cfg.Password = storage.Password(s)
	return b
}

// WithDatabase selects logical database n (0–15) on the Redis server.
// Calling this overrides any DB value previously set in the client options.
func (b *CacheBuilder) WithDatabase(n int) *CacheBuilder {
	b.cfg.LogicalDBs = storage.LogicalDBs(n)
	return b
}

// WithPoolSize sets the maximum number of connections in the pool to n.
// A larger pool can improve throughput under high concurrency but increases resource usage.
func (b *CacheBuilder) WithPoolSize(n int) *CacheBuilder {
	b.cfg.PoolSize = n
	return b
}

// WithMinIdleConns ensures at least n idle connections are kept open in the pool.
// This helps reduce latency on the first requests by avoiding repeated dials.
func (b *CacheBuilder) WithMinIdleConns(n int) *CacheBuilder {
	b.cfg.MinIdleConns = n
	return b
}

// WithTimeouts sets all four timeouts: PoolTimeout, DialTimeout, ReadTimeout, WriteTimeout
// to the specified values. Use this to override the package’s default safe timeouts in one call.
func (b *CacheBuilder) WithTimeouts(pool, dial, read, write time.Duration) *CacheBuilder {
	b.cfg.PoolTimeout = storage.Timeout(pool)
	b.cfg.DialTimeout = storage.Timeout(dial)
	b.cfg.ReadTimeout = storage.Timeout(read)
	b.cfg.WriteTimeout = storage.Timeout(write)
	return b
}

// WithLimiter sets the Limiter (circuit breaker or rate limiter) to l.
// If no limiter is provided, a default high-QPS breaker will be used.
func (b *CacheBuilder) WithLimiter(l breaker.Breaker) *CacheBuilder {
	b.cfg.Limiter = l
	return b
}

// WithMaxRetries configures the retry policy with max attempts, and backoff range.
// A value of 0 disables retries entirely.
func (b *CacheBuilder) WithRetries(max int, minBackoff, maxBackoff time.Duration) *CacheBuilder {
	b.cfg.MaxRetries = max
	b.cfg.MinRetryBackoff = storage.Timeout(minBackoff)
	b.cfg.MaxRetryBackoff = storage.Timeout(maxBackoff)
	return b
}

// WithTTL sets the default time-to-live for entries to t.
// All Set calls without an explicit expiration will use this TTL.
func (b *CacheBuilder) WithTTL(ttl time.Duration) *CacheBuilder {
	b.cfg.DefaultTTL = storage.Timeout(ttl)
	return b
}
