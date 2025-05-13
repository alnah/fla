package aiclient

import (
	"encoding/json"
	"fmt"
)

// MarshalJSON handles provider-specific JSON fields.
func (c ChatClient) MarshalJSON() ([]byte, error) {
	v := c.MaxTokens.Int()
	switch {
	case c.UseOpenAI:
		type openaiPayload struct {
			Model       string    `json:"model"`
			Temperature float32   `json:"temperature"`
			Messages    []Message `json:"messages"`
			MaxTokens   *int      `json:"max_completion_tokens,omitempty"`
		}
		payload := openaiPayload{
			Model:       c.Model.String(),
			Temperature: c.Temperature.Float32(),
			Messages:    append([]Message{{Role: RoleSystem, Content: c.System}}, c.Messages...),
			MaxTokens:   (*int)(&c.MaxTokens),
		}
		return json.Marshal(payload)
	case c.UseAnthropic:
		type anthropicPayload struct {
			Model               string    `json:"model"`
			System              string    `json:"system"`
			Messages            []Message `json:"messages"`
			MaxCompletionTokens *int      `json:"max_tokens,omitempty"`
			Temperature         float32   `json:"temperature,omitempty"`
		}
		payload := anthropicPayload{
			Model:               c.Model.String(),
			System:              c.System,
			Messages:            c.Messages,
			MaxCompletionTokens: &v,
			Temperature:         c.Temperature.Float32(),
		}
		return json.Marshal(payload)
	default:
		return nil, fmt.Errorf("no provider configured")
	}
}
