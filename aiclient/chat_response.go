package aiclient

import (
	"encoding/json"
	"fmt"
)

// ChatCompletion holds the result of chat completions for chat completion providers.
type ChatCompletion struct {
	Content string
}

func (cc ChatCompletion) String() string {
	if cc.Content != "" {
		return cc.Content
	}
	return ""
}

// ParseResponse extracts a ChatCompletion from raw JSON depending on provider.
func (c *ChatClient) ParseResponse(byt []byte) (ChatCompletion, error) {
	switch {
	case c.UseOpenAI:
		type openaiPayload struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		var payload openaiPayload
		if err := json.Unmarshal(byt, &payload); err != nil {
			return ChatCompletion{}, fmt.Errorf("failed to unmarshal response body: %w", err)
		}
		if len(payload.Choices) == 0 {
			return ChatCompletion{}, fmt.Errorf("no choices in OpenAI response")
		}
		return ChatCompletion{Content: payload.Choices[0].Message.Content}, nil

	case c.UseAnthropic:
		type anthropicPayload struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		}
		var payload anthropicPayload
		if err := json.Unmarshal(byt, &payload); err != nil {
			return ChatCompletion{}, fmt.Errorf("failed to unmarshal response body: %w", err)
		}
		if len(payload.Content) == 0 {
			return ChatCompletion{}, fmt.Errorf("no content in Anthropic response")
		}
		return ChatCompletion{Content: payload.Content[0].Text}, nil

	default:
		return ChatCompletion{}, fmt.Errorf("no provider configured")
	}
}
