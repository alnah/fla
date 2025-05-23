package cache

import (
	"context"
	"errors"
	"runtime"
	"time"

	"github.com/alnah/fla/breaker"
	"github.com/alnah/fla/logger"
	"github.com/alnah/fla/storage"
	"github.com/redis/go-redis/v9"
)

// CacheConfig holds all settings for a RedisCache instance.
type CacheConfig struct {
	Addr            storage.Address
	Password        storage.Password
	LogicalDBs      storage.LogicalDBs
	PoolSize        int
	MinIdleConns    int
	PoolTimeout     storage.Timeout
	DialTimeout     storage.Timeout
	ReadTimeout     storage.Timeout
	WriteTimeout    storage.Timeout
	MaxRetries      int
	MinRetryBackoff storage.Timeout
	MaxRetryBackoff storage.Timeout
	DefaultTTL      storage.Timeout
	Limiter         breaker.Breaker
	Logger          logger.Logger
}

// CacheBuilder builds a RedisCache through chained configuration.
// The builder pattern is used because the *redis.Options can't be mutated in place to implement a functional pattern.
// It provides key/value cache backed by Redis.
type CacheBuilder struct{ cfg CacheConfig }

// NewCache returns a Redis cache initialized with sensible defaults.
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
func NewCache() *CacheBuilder {
	return &CacheBuilder{
		cfg: CacheConfig{
			LogicalDBs:      RedisLogicalDBs,
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
	defaultTTL storage.Timeout
}

// Build validates the configuration, initializes a Redis client,
// and returns a ready-to-use Cache or an error.
func (b *CacheBuilder) Build() (*RedisCache, error) {
	cfg := b.cfg

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	rc := redis.NewClient(&redis.Options{
		Addr:            cfg.Addr.String(),
		Password:        cfg.Password.String(),
		DB:              cfg.LogicalDBs.Int(),
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		PoolTimeout:     cfg.PoolTimeout.Duration(),
		DialTimeout:     cfg.ReadTimeout.Duration(),
		ReadTimeout:     cfg.ReadTimeout.Duration(),
		WriteTimeout:    cfg.WriteTimeout.Duration(),
		MaxRetries:      cfg.MaxRetries,
		MinRetryBackoff: cfg.MinRetryBackoff.Duration(),
		MaxRetryBackoff: cfg.MaxRetryBackoff.Duration(),
		Limiter:         NewLimiter(cfg.Limiter),
	})
	// starting health check
	startHealthCheck(context.Background(), rc, logger.Default(), 1*time.Minute)

	return &RedisCache{
		log:        cfg.Logger,
		client:     rc,
		defaultTTL: cfg.DefaultTTL,
	}, nil
}

// Shutdown closes the client, releasing any open resources.
// It is rare to Close a Client, as the Client is meant to be long-lived and shared between many goroutines.
func (rc *RedisCache) Shutdown() error {
	return rc.client.Close()
}

// Set stores a value with optional TTL, defaulting if ttl<=0.
// It uses a circuit-breaker to guard Redis traffic.
func (rc *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	// choose expiration
	exp := ttl
	if exp <= 0 {
		exp = rc.defaultTTL.Duration()
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

// Get retrieves bytes for key or ErrCacheMiss if not found.
func (rc *RedisCache) Get(ctx context.Context, key string) (Result, error) {
	if err := rc.client.Options().Limiter.Allow(); err != nil {
		return nil, NewRedisCacheError("getting", err)
	}
	val, err := rc.client.Get(ctx, key).Bytes()
	rc.client.Options().Limiter.ReportResult(err)
	if err == redis.Nil {
		return nil, NewRedisCacheError("getting", ErrCacheMiss)
	}
	if err != nil {
		return nil, NewRedisCacheError("getting", err)
	}
	return val, nil
}

// Delete removes a key, returning true if it existed.
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

// Exists checks presence of key without modifying state.
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

// validate checks all fields of CacheConfig without using reflection.
func (c CacheConfig) validate() error {
	var accError []error

	if err := c.Addr.Validate(); err != nil {
		accError = append(accError, NewRedisCacheError("validating: address", err))
	}
	if err := c.Password.Validate(); err != nil {
		accError = append(accError, NewRedisCacheError("validating: password", err))
	}
	if err := c.LogicalDBs.Validate(); err != nil {
		accError = append(accError, NewRedisCacheError("validating: number of logical databses", err))
	}
	if err := c.PoolTimeout.Validate(); err != nil {
		accError = append(accError, NewRedisCacheError("validating: pool timeout", err))
	}
	if err := c.DialTimeout.Validate(); err != nil {
		accError = append(accError, NewRedisCacheError("validating: dial timeout", err))
	}
	if err := c.ReadTimeout.Validate(); err != nil {
		accError = append(accError, NewRedisCacheError("validating: read timeout", err))
	}
	if err := c.WriteTimeout.Validate(); err != nil {
		accError = append(accError, NewRedisCacheError("validating: write timeout", err))
	}
	if err := c.MinRetryBackoff.Validate(); err != nil {
		accError = append(accError, NewRedisCacheError("validating: min retry backoff", err))
	}
	if err := c.MaxRetryBackoff.Validate(); err != nil {
		accError = append(accError, NewRedisCacheError("validating: max retry backoff", err))
	}
	if err := c.DefaultTTL.Validate(); err != nil {
		accError = append(accError, NewRedisCacheError("validating: default ttl", err))
	}
	if c.MaxRetries < 0 {
		accError = append(accError, NewRedisCacheError("validating: max retries: must be greater than 0", nil))
	}
	if c.PoolSize <= 0 {
		accError = append(accError, NewRedisCacheError("validating: pool size: must be greater than or equal to 0", nil))
	}
	if c.MinIdleConns < 0 {
		accError = append(accError, NewRedisCacheError("validating: min idle connections: must be greater than or equal to 0", nil))
	}
	if c.MinIdleConns > c.PoolSize {
		accError = append(accError, NewRedisCacheError("validating: min idle connections can't be greater than pool size", nil))
	}
	if c.Limiter == nil {
		accError = append(accError, NewRedisCacheError("validating: limiter can't be nil", nil))
	}
	if c.Logger == nil {
		accError = append(accError, NewRedisCacheError("validating: logger can't be nil", nil))
	}

	return errors.Join(accError...)
}

// startHealthCheck spawns a goroutine that pings Redis every interval.
// On failure, it logs and triggers reconnection via Process.
func startHealthCheck(ctx context.Context, client *redis.Client, logger logger.Logger, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// prepare the PING command
				cmd := redis.NewStringCmd(ctx, "PING")

				// send it (Process returns error directly)
				if err := client.Process(ctx, cmd); err != nil {
					logger.Warn("redis health-check", "ping failed", err)
					continue
				}

				// inspect the result of the command
				if res, err := cmd.Result(); err != nil {
					logger.Error("redis health-check", "reconnect failed", err)
				} else {
					logger.Info("redis health-check", "reconnected successfully, PING=", res)
				}

			case <-ctx.Done():
				return
			}
		}
	}()
}
