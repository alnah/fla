package aiclient

import (
	"fmt"
	"strings"
)

type httpClientError struct {
	Provider  provider
	Operation operation
	Message   string
	Wrapped   error
}

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

func (e *httpClientError) Unwrap() error {
	return e.Wrapped
}

func NewChatError(provider provider, message string, wrapped error) *httpClientError {
	return &httpClientError{
		Provider:  provider,
		Operation: OpChatCompletion,
		Message:   message,
		Wrapped:   wrapped,
	}
}

func NewTTSError(provider provider, message string, wrapped error) *httpClientError {
	return &httpClientError{
		Provider:  provider,
		Operation: OpTranscript,
		Message:   message,
		Wrapped:   wrapped,
	}
}
