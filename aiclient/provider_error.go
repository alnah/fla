package aiclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
)

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

type anthropicError struct {
	StatusCode int
	ErrType    string
	Message    string
}

func (e *anthropicError) Error() string {
	return fmt.Sprintf("anthropic api error %d %s: %s", e.StatusCode, e.ErrType, e.Message)
}

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

type elevenlabsError struct {
	StatusCode int
	Status     string
	Message    string
}

func (e *elevenlabsError) Error() string {
	return fmt.Sprintf("elevenlabs api error %d %s: %s", e.StatusCode, e.Status, e.Message)
}

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

func isRetryable(err error) bool {
	var (
		openaiError     *openaiError
		anthropicError  *anthropicError
		elevenlabsError *elevenlabsError
		netError        net.Error
	)

	switch {
	case errors.As(err, &openaiError):
		return openaiError.StatusCode == 429 || openaiError.StatusCode >= 500
	case errors.As(err, &anthropicError):
		return anthropicError.StatusCode == 429 || anthropicError.StatusCode >= 500
	case errors.As(err, &elevenlabsError):
		return elevenlabsError.StatusCode == 429 || elevenlabsError.StatusCode >= 500
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return false
	case errors.As(err, &netError):
		return netError.Timeout()
	default:
		return false
	}
}
