package aiclient

import (
	"fmt"
	"strings"
)

// httpClientError represents an error in the HTTP client layer.
// It captures which provider and operation were in flight, a human-friendly
// message, and any underlying error to facilitate debugging.
type httpClientError struct {
	Provider  provider
	Operation operation
	Message   string
	Wrapped   error
}

// Error formats a descriptive string including failed operation, provider,
// contextual message and wrapped error, so callers can quickly understand
// what went wrong.
func (e httpClientError) Error() string {
	const prefix = "http client error"

	// build the "core" piece: operation and/or provider
	var core string
	if e.Operation != "" {
		core = fmt.Sprintf("operation %s failed", e.Operation)
		if e.Provider != "" {
			core += fmt.Sprintf(" for %s", e.Provider)
		}
	} else if e.Provider != "" {
		core = fmt.Sprintf("for %s", e.Provider)
	}

	// collect remaining pieces: message and wrapped error
	parts := []string{}
	if core != "" {
		parts = append(parts, core)
	}
	if e.Message != "" {
		parts = append(parts, e.Message)
	}
	if e.Wrapped != nil {
		parts = append(parts, e.Wrapped.Error())
	}

	// assemble final string
	if len(parts) == 0 {
		return prefix
	}

	// choose separator before the first part
	first := parts[0]
	var sep string
	switch {
	case strings.HasPrefix(first, "operation"):
		sep = ": "
	case strings.HasPrefix(first, "for"):
		sep = " "
	default:
		sep = ": "
	}

	result := prefix + sep + first
	for _, p := range parts[1:] {
		result += ": " + p
	}
	return result
}

// Unwrap yields the underlying error, enabling standard Go error chaining.
func (e *httpClientError) Unwrap() error {
	return e.Wrapped
}

// NewChatClientError wraps an error arising during a chat-completion call,
// tagging it with the provider and a custom message for context.
func NewChatClientError(pvd provider, message string, wrapped error) *httpClientError {
	return &httpClientError{
		Provider:  pvd,
		Operation: opChatCompletion,
		Message:   message,
		Wrapped:   wrapped,
	}
}

// NewTTSClientError wraps an error arising during a text-to-speech call,
// tagging it with the provider and a custom message for context.
func NewTTSClientError(pvd provider, message string, wrapped error) *httpClientError {
	return &httpClientError{
		Provider:  pvd,
		Operation: opTextToSpeech,
		Message:   message,
		Wrapped:   wrapped,
	}
}

// NewSTTClientError wraps an error arising during a speech-to-text call,
// tagging it with the provider and a custom message for context.
func NewSTTClientError(pvd provider, message string, wrapped error) *httpClientError {
	return &httpClientError{
		Provider:  pvd,
		Operation: opSpeechToText,
		Message:   message,
		Wrapped:   wrapped,
	}
}
