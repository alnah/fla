package aiclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
)

// isRetryable determines whether an error should be retried,
// based on HTTP status (e.g. 429, 5xx), network timeouts,
// or context cancellation semantics.
func isRetryable(err error) bool {
	var (
		oe     *openaiError
		ae     *anthropicError
		ee     *elevenlabsError
		netErr net.Error
	)

	switch {
	// retry on HTTP 429 or any 5xx
	case errors.As(err, &oe):
		return oe.StatusCode == 429 || oe.StatusCode >= 500

	case errors.As(err, &ae):
		return ae.StatusCode == 429 || ae.StatusCode >= 500

	case errors.As(err, &ee):
		return ee.StatusCode == 429 || ee.StatusCode >= 500

	// user cancelled or deadline exceeded are not retryable
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return false

	// network timeouts are retryable
	case errors.As(err, &netErr):
		return netErr.Timeout()

	default:
		return false
	}
}

// buildProviderError inspects an HTTP response’s status code and delegates
// to the appropriate provider-specific builder to produce a rich error.
func buildProviderError(provider provider, res *http.Response) error {
	switch provider {
	case ProviderOpenAI:
		return buildOpenaiError(res)
	case ProviderAnthropic:
		return buildAnthropicError(res)
	case ProviderElevenLabs:
		return buildElevenlabsError(res)
	default:
		return fmt.Errorf("%s error: status %d", provider.String(), res.StatusCode)
	}
}

// openaiError models an error response from OpenAI’s API,
// preserving status, type, message and code for diagnostics.
type openaiError struct {
	StatusCode int
	Message    string
	Type       string
	Param      string
	Code       string
}

func (e *openaiError) Error() string {
	return fmt.Sprintf("openai error %d %s: %s", e.StatusCode, e.Type, e.Message)
}

// buildOpenaiError decodes OpenAI’s error payload into openaiError,
// falling back to a generic status error if decoding fails.
func buildOpenaiError(res *http.Response) error {
	defer func() { _ = res.Body.Close() }()

	// openai error shape
	var payload struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Param   string `json:"param,omitempty"`
			Code    string `json:"code,omitempty"`
		} `json:"error"`
	}

	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		// fallback
		return fmt.Errorf("openai api error: status %d", res.StatusCode)
	}

	// build error
	return &openaiError{
		StatusCode: res.StatusCode,
		Message:    payload.Error.Message,
		Type:       payload.Error.Type,
		Param:      payload.Error.Param,
		Code:       payload.Error.Code,
	}
}

// anthropicError models an error response from Anthropic’s API,
// preserving status and message for diagnostics.
type anthropicError struct {
	StatusCode int
	ErrType    string
	Message    string
}

func (e *anthropicError) Error() string {
	return fmt.Sprintf("anthropic api error %d %s: %s", e.StatusCode, e.ErrType, e.Message)
}

// buildAnthropicError decodes Anthropic’s error payload into anthropicError,
// falling back to a generic status error if decoding fails.
func buildAnthropicError(res *http.Response) error {
	defer func() { _ = res.Body.Close() }()

	// anthropic error shape
	var payload struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		// fallback
		return fmt.Errorf("anthropic api error: status %d", res.StatusCode)
	}

	// build error
	return &anthropicError{
		StatusCode: res.StatusCode,
		ErrType:    payload.Error.Type,
		Message:    payload.Error.Message,
	}
}

// elevenlabsError models an error response from ElevenLabs’ API,
// preserving status and detail message for diagnostics.
type elevenlabsError struct {
	StatusCode int
	Status     string
	Message    string
}

func (e *elevenlabsError) Error() string {
	return fmt.Sprintf("elevenlabs api error %d %s: %s", e.StatusCode, e.Status, e.Message)
}

// buildElevenlabsError decodes ElevenLabs’ error payload into elevenlabsError,
// falling back to a generic status error if decoding fails.
func buildElevenlabsError(res *http.Response) error {
	defer func() { _ = res.Body.Close() }()

	// eleven labs error shape
	var payload struct {
		Detail struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		} `json:"detail"`
	}

	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		// fallback
		return fmt.Errorf("elevenlabs api error: status %d", res.StatusCode)
	}

	// build error
	return &elevenlabsError{
		StatusCode: res.StatusCode,
		Status:     payload.Detail.Status,
		Message:    payload.Detail.Message,
	}
}
