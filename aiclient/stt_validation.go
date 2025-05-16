package aiclient

import (
	"errors"
	"fmt"
)

func (t *STTClient) validate() (string, error) {
	if t.base.ctx == nil {
		return "", errors.New("context must be provided")
	}
	if t.base.logger == nil {
		return "", errors.New("logger must be set")
	}
	if t.base.httpClient == nil {
		return "", errors.New("http client must be set")
	}
	if err := t.base.provider.Validate(); err != nil {
		return "", err
	}
	if err := t.base.url.Validate(); err != nil {
		return "", err
	}
	if err := t.base.apiKey.Validate(); err != nil {
		return "", err
	}
	if err := t.base.httpMethod.Validate(); err != nil {
		return "", err
	}
	if err := t.base.model.Validate(); err != nil {
		return "", err
	}
	if err := t.language.Validate(); err != nil {
		return "", err
	}
	if t.useOpenAI == t.useElevenLabs {
		return "", errors.New("must configure exactly one provider: openai or elevenlabs")
	}
	if t.useOpenAI && t.base.provider != ProviderOpenAI {
		return "", fmt.Errorf("url indicates openai but provider is %s", t.base.provider)
	}
	if t.useElevenLabs && t.base.provider != ProviderElevenLabs {
		return "", fmt.Errorf("url indicates elevenlabs but provider is %s", t.base.provider)
	}
	allowed := []string{"flac", "mp3", "mp4", "mpeg", "mpga", "m4a", "ogg", "wav", "webm"}
	switch {
	case t.useOpenAI:
		if t.base.model != ModelSTTOpenAI {
			return "", fmt.Errorf("model %s not supported by openai", t.base.model)
		}
		return t.filePath.Validate(25, allowed...)
	case t.useElevenLabs:
		if t.base.model != ModelSTTElevenLabs {
			return "", fmt.Errorf("model %s not supported by elevenlabs", t.base.model)
		}
		return t.filePath.Validate(25, allowed...)
	}
	return "", nil
}
