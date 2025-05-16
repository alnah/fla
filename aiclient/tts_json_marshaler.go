package aiclient

import (
	"encoding/json"
	"errors"
)

func (s TTSClient) MarshalJSON() ([]byte, error) {
	switch {
	case s.useOpenAI:
		type openaiPayload struct {
			Input        string `json:"input"`
			Model        string `json:"model"`
			Voice        string `json:"voice"`
			Instructions string `json:"instructions"`
		}
		payload := openaiPayload{
			Input:        s.text.String(),
			Model:        s.base.model.String(),
			Voice:        s.voice.String(),
			Instructions: s.instructions.String(),
		}
		return json.Marshal(payload)
	case s.useElevenLabs:
		type elevenlabsPayload struct {
			Text          string `json:"text"`
			ModelID       string `json:"model_id"`
			VoiceSettings struct {
				Speed Speed `json:"speed"`
			} `json:"voice_settings"`
		}
		payload := elevenlabsPayload{
			Text:    s.text.String(),
			ModelID: s.base.model.String(),
			VoiceSettings: struct {
				Speed Speed "json:\"speed\""
			}{s.speed},
		}
		return json.Marshal(payload)
	default:
		return nil, errors.New("no provider configured")
	}
}
