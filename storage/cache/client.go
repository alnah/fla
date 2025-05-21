package cache

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/alnah/fla/breaker"
	"github.com/alnah/fla/logger"
	"github.com/redis/go-redis/v9"
)

// CacheConfig holds all settings for a RedisCache instance.
type CacheConfig struct {
	Addr            string
	Password        string
	Databases       int
	PoolSize        int
	MinIdleConns    int
	PoolTimeout     time.Duration
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	MaxRetries      int
	MinRetryBackoff time.Duration
	MaxRetryBackoff time.Duration
	DefaultTTL      time.Duration
	Limiter         breaker.Breaker
	Logger          logger.Logger
}

// CacheBuilder builds a RedisCache through chained configuration.
// The builder pattern is used because the *redis.Options can't be mutated in place to implement a functional pattern.
// It provides key/value cache backed by Redis.
type CacheBuilder struct{ cfg CacheConfig }

// NewRedisCache returns a RedisCache initialized with sensible defaults.
//
// It embeds a default logger and initializes a *redis.Client with sensible defaults:
//   - DB is set to 16 if unset
//   - Connection pool size is 10 per proc
//   - Min idle connections is 1 per proc
//   - Timeouts (dial, read, write, pool) default to safe values
//   - Retries are enabled with 5 attempts and backoff (100ms–500ms)
//   - A circuit-breaker-based limiter is set to a high-QPS profile if none is provided
//
// Use With* options to override any settings.
// Returns an error if address and passwords are missing.
func NewRedisCache() *CacheBuilder {
	return &CacheBuilder{
		cfg: CacheConfig{
			Databases:       RedisLogicalDBs,
			PoolSize:        10 * runtime.NumCPU(),
			MinIdleConns:    runtime.NumCPU(),
			PoolTimeout:     RedisPoolTimeout,
			DialTimeout:     RedisDialTimeout,
			ReadTimeout:     RedisReadTimeout,
			WriteTimeout:    RedisWriteTimeout,
			MaxRetries:      RedisMaxRetries,
			MinRetryBackoff: RedisMinRetryBackoff,
			MaxRetryBackoff: RedisMaxRetryBackoff,
			DefaultTTL:      RedisDefaultTTL,
			Limiter:         breaker.New(breaker.HighQPSConfig()),
			Logger:          logger.Default(),
		},
	}
}

// RedisCache provides a key/value cache backed by Redis.
type RedisCache struct {
	log        logger.Logger
	client     *redis.Client
	defaultTTL time.Duration
}

// Build constructs the RedisCache from the accumulated configuration.
// Returns an error if mandatory fields (Addr) are missing.
func (b *CacheBuilder) Build() (*RedisCache, error) {
	cfg := b.cfg

	if cfg.Addr == "" {
		return nil, NewRedisCacheError("build", errors.New("address is missing"))
	}

	if cfg.Password == "" {
		cfg.Logger.Warn("redis cache", "initialization", "password is missing")
	}

	// Create underlying Redis client
	client := redis.NewClient(&redis.Options{
		Addr:            cfg.Addr,
		Password:        cfg.Password,
		DB:              cfg.Databases,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		PoolTimeout:     cfg.PoolTimeout,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		MaxRetries:      cfg.MaxRetries,
		MinRetryBackoff: cfg.MinRetryBackoff,
		MaxRetryBackoff: cfg.MaxRetryBackoff,
		Limiter:         NewLimiter(cfg.Limiter),
	})

	// test connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, NewRedisCacheError("build", fmt.Errorf("ping failed: %w", err))
	}

	// wrap in our RedisCache
	return &RedisCache{
		log:        cfg.Logger,
		client:     client,
		defaultTTL: cfg.DefaultTTL,
	}, nil
}

// Shutdown closes the client, releasing any open resources.
// It is rare to Close a Client, as the Client is meant to be long-lived and shared between many goroutines.
func (rc *RedisCache) Shutdown() error {
	return rc.client.Close()
}

// Set implements CacheClient.Set.
// If ttl <= 0, falls back to rc.defaultTTL.
// Wraps every call with the circuit-breaker.
func (rc *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	// choose expiration
	exp := ttl
	if exp <= 0 {
		exp = rc.defaultTTL
	}
	// admission
	if err := rc.client.Options().Limiter.Allow(); err != nil {
		return NewRedisCacheError("setting", err)
	}
	// execute
	err := rc.client.Set(ctx, key, value, exp).Err()
	// report
	rc.client.Options().Limiter.ReportResult(err)
	if err != nil {
		return NewRedisCacheError("setting", err)
	}
	return nil
}

// Get implements CacheClient.Get.
// Returns ErrCacheMiss if key does not exist.
func (rc *RedisCache) Get(ctx context.Context, key string) (string, error) {
	if err := rc.client.Options().Limiter.Allow(); err != nil {
		return "", NewRedisCacheError("getting", err)
	}
	val, err := rc.client.Get(ctx, key).Result()
	rc.client.Options().Limiter.ReportResult(err)
	if err == redis.Nil {
		return "", NewRedisCacheError("getting", ErrCacheMiss)
	}
	if err != nil {
		return "", NewRedisCacheError("getting", err)
	}
	return val, nil
}

// Delete implements CacheClient.Delete.
// Returns true if the key was present.
func (rc *RedisCache) Delete(ctx context.Context, key string) (bool, error) {
	if err := rc.client.Options().Limiter.Allow(); err != nil {
		return false, NewRedisCacheError("deleting", err)
	}
	n, err := rc.client.Del(ctx, key).Result()
	rc.client.Options().Limiter.ReportResult(err)
	if err != nil {
		return false, NewRedisCacheError("deleting", err)
	}
	return n > 0, nil
}

// Exists implements CacheClient.Exists.
// Returns true if the key exists.
func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	if err := rc.client.Options().Limiter.Allow(); err != nil {
		return false, NewRedisCacheError("checking existence", err)
	}
	n, err := rc.client.Exists(ctx, key).Result()
	rc.client.Options().Limiter.ReportResult(err)
	if err != nil {
		return false, NewRedisCacheError("checking existence", err)
	}
	return n > 0, NewRedisCacheError("checking existence", err)
}
