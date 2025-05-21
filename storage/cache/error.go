package cache

import (
	"errors"
	"fmt"
)

// CacheError represents a generic cache holding the component e.g. "redis", the
// failed operation e.g. "initializing", and the error cause.
type CacheError struct {
	Component string
	Operation string
	Cause     error
}

// Error returns the error string.
func (e *CacheError) Error() string {
	return fmt.Sprintf("%s cache: %s: %v", e.Component, e.Operation, e.Cause)
}

// Unwrap returns the error cause.
func (e *CacheError) Unwrap() error { return e.Cause }

// NewCacheError is a helper to build a cache error using the component e.g. "redis",
// the failed operation e.g. "initializing", and the err cause.
func NewCacheError(component, operation string, err error) *CacheError {
	return &CacheError{Component: component, Operation: operation, Cause: err}
}

// NewRedisCacheError is a helper to build a redis cache error, using the operation, e.g.
// "initiliazing", and the err cause.
func NewRedisCacheError(operation string, err error) *CacheError {
	return NewCacheError("redis", operation, err)
}

// ErrCacheMiss represents a key not found error.
var ErrCacheMiss = errors.New("key not found")
