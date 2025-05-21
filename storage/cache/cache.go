package cache

import (
	"context"
	"time"
)

// TODO: perform health check peridically (perhaps once every few seconds)
// TODO: implement OpenTelemetry
type CacheClient interface {
	// Set stores a value under key, with optional TTL.
	// If ttl==0, use the default TTL configured on the client.
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	// Get returns the stored string or a CacheError.
	Get(ctx context.Context, key string) (string, error)
	// Delete removes the key; returns true if it was present.
	Delete(ctx context.Context, key string) (bool, error)
	// Exists reports whether the key exists.
	Exists(ctx context.Context, key string) (bool, error)
	// Shutdown closes the cache client.
	Shutdown() error
}
