package cache

import (
	"context"
	"time"
)

// Cache defines a generic key/value store interface.
// It abstracts backing implementations (e.g., Redis) and
// supports basic operations, optional TTLs, and graceful shutdown.
type Cache interface {
	// Set stores a value under key with an optional TTL.
	// A zero TTL uses the client's default expiration.
	Set(ctx context.Context, key string, value any, ttl time.Duration) error

	// Get retrieves a value by key, returning ErrCacheMiss if absent.
	Get(ctx context.Context, key string) (Result, error)

	// Delete removes the key; returns true if the key existed.
	Delete(ctx context.Context, key string) (bool, error)

	// Exists reports whether a key is present without side effects.
	Exists(ctx context.Context, key string) (bool, error)

	// Shutdown releases resources and closes any open connections.
	Shutdown() error
}

// Result wraps raw bytes from Cache.Get and provides convenience methods.
type Result []byte

// String returns the result as a UTF-8 string.
func (r Result) String() string { return string(r) }

// Bytes returns the raw byte slice.
func (r Result) Bytes() []byte { return r }
