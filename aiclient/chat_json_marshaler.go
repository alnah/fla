package aiclient

import (
	"encoding/json"
	"fmt"
)

// MarshalJSON handles provider-specific JSON fields.
func (c ChatClient) MarshalJSON() ([]byte, error) {
	v := c.maxTokens.Int()
	switch {
	case c.useOpenAI:
		type openaiPayload struct {
			Model       string    `json:"model"`
			Temperature float32   `json:"temperature"`
			Messages    []Message `json:"messages"`
			MaxTokens   *int      `json:"max_completion_tokens,omitempty"`
		}
		payload := openaiPayload{
			Model:       c.base.model.String(),
			Temperature: c.temperature.Float32(),
			Messages:    append([]Message{{Role: RoleSystem, Content: c.system}}, c.messages...),
			MaxTokens:   (*int)(&c.maxTokens),
		}
		return json.Marshal(payload)
	case c.useAnthropic:
		type anthropicPayload struct {
			Model               string    `json:"model"`
			System              string    `json:"system"`
			Messages            []Message `json:"messages"`
			MaxCompletionTokens *int      `json:"max_tokens,omitempty"`
			Temperature         float32   `json:"temperature,omitempty"`
		}
		payload := anthropicPayload{
			Model:               c.base.model.String(),
			System:              c.system,
			Messages:            c.messages,
			MaxCompletionTokens: &v,
			Temperature:         c.temperature.Float32(),
		}
		return json.Marshal(payload)
	default:
		return nil, fmt.Errorf("no provider configured")
	}
}
