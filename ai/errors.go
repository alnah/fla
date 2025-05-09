package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

// OpenAIError represents an error response from OpenAI API.
type OpenAIError struct {
	Error struct {
		Type    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// AnthropicError represents an error response from Anthropic API.
type AnthropicError struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// ElevenLabsError rerpesents an error response from ElevenLabs API.
type ElevenLabsError struct {
	Error struct {
		Type    string `json:"status"`
		Message string `json:"message"`
	} `json:"detail"`
}

// HTTPError represents an HTTP response classified as an error, carrying status,
// API error details, and any Retry-After hint.
type HTTPError struct {
	Status     string
	Type       string
	Message    string
	RetryAfter time.Duration // parsed from Retry-After header, if any
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("status=%s, type=%s, message=%s", e.Status, e.Type, e.Message)
}

// NewHTTPError reads and restores the body, extracts type, and message according to the given provider.
func NewHTTPError(provider Provider, res *http.Response) *HTTPError {
	// read and restore body
	body, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	res.Body = io.NopCloser(bytes.NewReader(body))

	// handle different types of error
	var typ, msg string
	switch provider {
	case ProviderOpenAI:
		var e OpenAIError
		_ = json.Unmarshal(body, &e)
		typ, msg = e.Error.Type, e.Error.Message

	case ProviderAnthropic:
		var e AnthropicError
		_ = json.Unmarshal(body, &e)
		typ, msg = e.Error.Type, e.Error.Message

	case ProviderElevenLabs:
		var e ElevenLabsError
		_ = json.Unmarshal(body, &e)
		typ, msg = e.Error.Type, e.Error.Message
	}

	// parse Retry-After header (seconds or HTTP-date)
	var ra time.Duration
	if h := res.Header.Get("Retry-After"); h != "" {
		if sec, err := strconv.Atoi(h); err == nil {
			ra = time.Duration(sec) * time.Second
		} else if t, err := http.ParseTime(h); err == nil {
			ra = time.Until(t)
		}
	}

	return &HTTPError{Status: res.Status, Type: typ, Message: msg, RetryAfter: ra}
}

type ErrType string

func NewRetryClassifier(retryable map[ErrType]struct{}) func(error) bool {
	return func(err error) bool {
		// HTTP classifier
		var e *HTTPError
		if errors.As(err, &e) {
			_, ok := retryable[ErrType(e.Type)]
			return ok
		}

		// context
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return false
		}

		// network
		var n net.Error
		if errors.As(err, &n) {
			return n.Timeout()
		}
		return false
	}
}

// AIError reports an unsucessful AI operation, for a provider, with a useful message,
// and the wrapped error cause.
type AIError struct {
	Operation Operation
	Provider  Provider
	Message   string
	Wrapped   error
}

func (e *AIError) Error() string {
	if e.Wrapped != nil {
		wrappedMsg := e.Wrapped.Error()
		if e.Message == wrappedMsg {
			return fmt.Sprintf("%s %s: %s", e.Operation, e.Provider, wrappedMsg)
		}
		return fmt.Sprintf("%s %s: %s: %v", e.Operation, e.Provider, e.Message, e.Wrapped)
	}
	return fmt.Sprintf("%s %s error: %s", e.Operation, e.Provider, e.Message)
}

func (e *AIError) Unwrap() error { return e.Wrapped }

func NewAIError(op Operation, pvd Provider, msg string, err error) error {
	var httpErr *HTTPError

	if errors.As(err, &httpErr) {
		return &AIError{Operation: op, Provider: pvd, Message: msg, Wrapped: httpErr}
	}

	return &AIError{Operation: op, Provider: pvd, Message: msg, Wrapped: err}
}

// /NewChatError returns an unsuccessful chat completion.
func NewChatError(p Provider, msg string, err error) error {
	return NewAIError(OpChatCompletion, p, msg, err)
}

// NewTTSError returns an unsuccessful text-to-speech synthesis.
func NewTTSError(p Provider, msg string, err error) error {
	return NewAIError(OpTTSAudio, p, msg, err)
}

// NewSTTError returns an unsuccessful speech-to-text transcription.
func NewSTTError(p Provider, msg string, err error) error {
	return NewAIError(OpSTTTranscription, p, msg, err)
}
