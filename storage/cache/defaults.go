package cache

import "time"

// Redis cache
const (
	// global ttl
	RedisDefaultTTL time.Duration = 15 * time.Second
	// logical databases
	RedisLogicalDBs int = 15
	// timeouts
	RedisPoolTimeout  time.Duration = 30 * time.Second
	RedisDialTimeout  time.Duration = 5 * time.Second
	RedisWriteTimeout time.Duration = 3 * time.Second
	RedisReadTimeout  time.Duration = 3 * time.Second
	// retry mechanism
	RedisMaxRetries      int           = 5
	RedisMinRetryBackoff time.Duration = 100 * time.Millisecond
	RedisMaxRetryBackoff time.Duration = 500 * time.Millisecond
)
