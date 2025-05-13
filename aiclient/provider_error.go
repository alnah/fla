package aiclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
)

type OpenAIError struct {
	StatusCode int
	Message    string
	Type       string
	Param      string
	Code       string
}

func (e *OpenAIError) Error() string {
	return fmt.Sprintf("OpenAI error %d %s: %s", e.StatusCode, e.Type, e.Message)
}

func BuildOpenAIError(res *http.Response) error {
	defer res.Body.Close()

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
		return fmt.Errorf("OpenAI API error: status %d", res.StatusCode)
	}

	// build error
	return &OpenAIError{
		StatusCode: res.StatusCode,
		Message:    payload.Error.Message,
		Type:       payload.Error.Type,
		Param:      payload.Error.Param,
		Code:       payload.Error.Code,
	}
}

type AnthropicError struct {
	StatusCode int
	ErrType    string
	Message    string
}

func (e *AnthropicError) Error() string {
	return fmt.Sprintf("Anthropic API error %d %s: %s", e.StatusCode, e.ErrType, e.Message)
}

func BuildAnthropicError(res *http.Response) error {
	defer res.Body.Close()

	// anthropic error shape
	var payload struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		// fallback
		return fmt.Errorf("Anthropic API error: status %d", res.StatusCode)
	}

	// build error
	return &AnthropicError{
		StatusCode: res.StatusCode,
		ErrType:    payload.Error.Type,
		Message:    payload.Error.Message,
	}
}

type ElevenLabsError struct {
	StatusCode int
	Status     string
	Message    string
}

func (e *ElevenLabsError) Error() string {
	return fmt.Sprintf("ElevenLabs API error %d %s: %s", e.StatusCode, e.Status, e.Message)
}

func BuildElevenLabsError(res *http.Response) error {
	defer res.Body.Close()

	// eleven labs error shape
	var payload struct {
		Details struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		} `json:"detail"`
	}

	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		// fallback
		return fmt.Errorf("ElevenLabs API error: status %d", res.StatusCode)
	}

	// build error
	return &ElevenLabsError{
		StatusCode: res.StatusCode,
		Status:     payload.Details.Status,
		Message:    payload.Details.Message,
	}
}

func IsRetryable(err error) bool {
	var (
		openaiError     OpenAIError
		anthropicError  AnthropicError
		elevenlabsError ElevenLabsError
		netError        net.Error
	)

	switch {
	case errors.As(err, &openaiError):
		return openaiError.StatusCode == 429 || openaiError.StatusCode >= 500
	case errors.As(err, &anthropicError):
		return anthropicError.StatusCode == 429 || anthropicError.StatusCode >= 500
	case errors.As(err, &elevenlabsError):
		return elevenlabsError.StatusCode == 429 || elevenlabsError.StatusCode >= 500
	case errors.As(err, &netError):
		return netError.Timeout()
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return false
	default:
		return false
	}
}
