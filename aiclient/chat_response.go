package aiclient

import (
	"encoding/json"
	"fmt"
)

// chatResponse holds the result of chat completions for chat completion providers.
type chatResponse struct {
	content string
}

func (cc chatResponse) Content() string {
	if cc.content != "" {
		return cc.content
	}
	return ""
}

// parseResponse extracts a ChatCompletion from raw JSON depending on provider.
func (c *Chat) parseResponse(byt []byte) (chatResponse, error) {
	switch {
	case c.useOpenAI:
		type openaiPayload struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		var payload openaiPayload
		if err := json.Unmarshal(byt, &payload); err != nil {
			return chatResponse{}, fmt.Errorf("failed to unmarshal response body: %w", err)
		}
		if len(payload.Choices) == 0 {
			return chatResponse{}, fmt.Errorf("no choices in OpenAI response")
		}
		return chatResponse{content: payload.Choices[0].Message.Content}, nil

	case c.useAnthropic:
		type anthropicPayload struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		}
		var payload anthropicPayload
		if err := json.Unmarshal(byt, &payload); err != nil {
			return chatResponse{}, fmt.Errorf("failed to unmarshal response body: %w", err)
		}
		if len(payload.Content) == 0 {
			return chatResponse{}, fmt.Errorf("no content in Anthropic response")
		}
		return chatResponse{content: payload.Content[0].Text}, nil

	default:
		return chatResponse{}, fmt.Errorf("no provider configured")
	}
}
