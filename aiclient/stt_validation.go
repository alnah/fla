package aiclient

import (
	"errors"
	"fmt"
)

func (s *STTClient) validate() (string, error) {
	if s.base.ctx == nil {
		return "", errors.New("context must be provided")
	}
	if s.base.log == nil {
		return "", errors.New("logger must be set")
	}
	if s.base.httpClient == nil {
		return "", errors.New("http client must be set")
	}
	if err := s.base.provider.Validate(); err != nil {
		return "", err
	}
	if err := s.base.url.Validate(); err != nil {
		return "", err
	}
	if err := s.base.apiKey.Validate(); err != nil {
		return "", err
	}
	if err := s.base.httpMethod.Validate(); err != nil {
		return "", err
	}
	if err := s.base.model.Validate(); err != nil {
		return "", err
	}
	if err := s.language.Validate(); err != nil {
		return "", err
	}
	if s.useOpenAI == s.useElevenLabs {
		return "", errors.New("must configure exactly one provider: openai or elevenlabs")
	}
	if s.useOpenAI && s.base.provider != ProviderOpenAI {
		return "", fmt.Errorf("url indicates openai but provider is %s", s.base.provider)
	}
	if s.useElevenLabs && s.base.provider != ProviderElevenLabs {
		return "", fmt.Errorf("url indicates elevenlabs but provider is %s", s.base.provider)
	}
	allowed := []string{"flac", "mp3", "mp4", "mpeg", "mpga", "m4a", "ogg", "wav", "webm"}
	switch {
	case s.useOpenAI:
		if s.base.model != ModelSTTOpenAI {
			return "", fmt.Errorf("model %s not supported by openai", s.base.model)
		}
		return s.filePath.Validate(25, allowed...)
	case s.useElevenLabs:
		if s.base.model != ModelSTTElevenLabs {
			return "", fmt.Errorf("model %s not supported by elevenlabs", s.base.model)
		}
		return s.filePath.Validate(25, allowed...)
	}
	return "", nil
}
