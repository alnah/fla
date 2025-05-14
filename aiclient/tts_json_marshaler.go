package aiclient

import (
	"encoding/json"
	"errors"
)

func (t TTSClient) MarshalJSON() ([]byte, error) {
	switch {
	case t.UseOpenAI:
		type openaiPayload struct {
			Input        string `json:"input"`
			Model        string `json:"model"`
			Voice        string `json:"string"`
			Instructions string `json:"instructions"`
		}
		payload := openaiPayload{
			Input:        t.Text.String(),
			Model:        t.base.Model.String(),
			Voice:        t.Voice.String(),
			Instructions: t.Instructions.String(),
		}
		return json.Marshal(payload)
	case t.UseElevenLabs:
		type elevenlabsPayload struct {
			Text          string `json:"text"`
			ModelID       string `json:"model_id"`
			VoiceSettings struct {
				Speed Speed `json:"speed"`
			} `json:"voice_settings"`
		}
		payload := elevenlabsPayload{
			Text:    t.Text.String(),
			ModelID: t.base.Model.String(),
			VoiceSettings: struct {
				Speed Speed "json:\"speed\""
			}{t.Speed},
		}
		return json.Marshal(payload)
	default:
		return nil, errors.New("no provider configured")
	}
}
