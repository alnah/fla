package cache

import (
	"time"

	"github.com/alnah/fla/storage"
)

// Redis cache
const (
	// global ttl
	RedisDefaultTTL = storage.Timeout(15 * time.Second)
	// logical databases
	RedisLogicalDBs storage.LogicalDBs = 15
	// timeouts
	RedisPoolTimeout  = storage.Timeout(30 * time.Second)
	RedisDialTimeout  = storage.Timeout(5 * time.Second)
	RedisWriteTimeout = storage.Timeout(3 * time.Second)
	RedisReadTimeout  = storage.Timeout(3 * time.Second)
	// retry mechanism
	RedisMaxRetries      int = 5
	RedisMinRetryBackoff     = storage.Timeout(100 * time.Millisecond)
	RedisMaxRetryBackoff     = storage.Timeout(500 * time.Millisecond)
)
