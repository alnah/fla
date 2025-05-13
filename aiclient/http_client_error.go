package aiclient

import "fmt"

type HTTPClientError struct {
	Provider  Provider
	Operation Operation
	Message   string
	Wrapped   error
}

func (e *HTTPClientError) Error() string {
	if e.Wrapped != nil {
		return fmt.Sprintf("%s %s: %s: %v", e.Operation.String(), e.Provider.String(), e.Message, e.Wrapped)
	}
	return fmt.Sprintf("%s %s: %s", e.Operation, e.Provider, e.Message)
}
func (e *HTTPClientError) Unwrap() error { return e.Wrapped }

func NewChatClientError(p Provider, m string, w error) *HTTPClientError {
	return &HTTPClientError{
		Provider:  p,
		Operation: OpChatCompletion,
		Message:   m,
		Wrapped:   w,
	}
}
