package cache

import (
	"errors"
	"strings"
)

// ErrCacheMiss represents a key not found error.
var ErrCacheMiss = errors.New("key not found")

// CacheError represents a generic cache holding the component e.g. "redis", the
// failed operation e.g. "initializing", and the error cause.
type CacheError struct {
	Component string
	Operation string
	Cause     error
}

// Error returns the error string.
func (e *CacheError) Error() string {
	parts := []string{}

	if e.Component != "" {
		parts = append(parts, e.Component+" cache")
	}

	if e.Operation != "" {
		if len(parts) > 0 {
			parts[len(parts)-1] += ": " + e.Operation
		} else {
			parts = append(parts, e.Operation)
		}
	}

	if e.Cause != nil && e.Cause.Error() != "" {
		parts = append(parts, e.Cause.Error())
	}

	if len(parts) == 0 {
		return "cache error"
	}

	return strings.Join(parts, ": ")
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
