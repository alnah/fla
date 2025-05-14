package aiclient

import (
	"fmt"
	"strings"
)

type HTTPClientError struct {
	Provider  Provider
	Operation Operation
	Message   string
	Wrapped   error
}

func (e HTTPClientError) Error() string {
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

func (e *HTTPClientError) Unwrap() error {
	return e.Wrapped
}

func NewChatClientError(p Provider, m string, w error) *HTTPClientError {
	return &HTTPClientError{
		Provider:  p,
		Operation: OpChatCompletion,
		Message:   m,
		Wrapped:   w,
	}
}
