package aiclient

import (
	"encoding/json"
	"errors"
)

func (s Speech) MarshalJSON() ([]byte, error) {
	switch {
	case s.useOpenAI:
		type openaiPayload struct {
			Input        string `json:"input"`
			Model        string `json:"model"`
			Voice        string `json:"voice"`
			Instructions string `json:"instructions"`
		}
		payload := openaiPayload{
			Input:        s.Text.String(),
			Model:        s.base.Model.String(),
			Voice:        s.Voice.String(),
			Instructions: s.Instructions.String(),
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
			Text:    s.Text.String(),
			ModelID: s.base.Model.String(),
			VoiceSettings: struct {
				Speed Speed "json:\"speed\""
			}{s.Speed},
		}
		return json.Marshal(payload)
	default:
		return nil, errors.New("no provider configured")
	}
}
