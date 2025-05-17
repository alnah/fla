package aiclient

import (
	"errors"
	"fmt"
)

func (s *TTSClient) validate() error {
	if s.base.ctx == nil {
		return errors.New("context must be provided")
	}
	if s.base.log == nil {
		return errors.New("logger must be set")
	}
	if s.base.httpClient == nil {
		return errors.New("http client must be set")
	}
	if err := s.base.provider.Validate(); err != nil {
		return err
	}
	if err := s.base.url.Validate(); err != nil {
		return err
	}
	if err := s.base.apiKey.Validate(); err != nil {
		return err
	}
	if err := s.base.httpMethod.Validate(); err != nil {
		return err
	}
	if err := s.base.model.Validate(); err != nil {
		return err
	}
	if err := s.voice.Validate(s.base.provider); err != nil {
		return err
	}
	if err := s.text.Validate(); err != nil {
		return err
	}
	if s.useOpenAI == s.useElevenLabs {
		return errors.New("must configure exactly one provider: openai or elevenlabs")
	}
	if s.useOpenAI && s.base.provider != ProviderOpenAI {
		return fmt.Errorf("url indicates openai but provider is %s", s.base.provider)
	}
	if s.useElevenLabs && s.base.provider != ProviderElevenLabs {
		return fmt.Errorf("url indicates elevenlabs but provider is %s", s.base.provider)
	}
	switch {
	case s.useOpenAI:
		if s.base.model != ModelTTSOpenAI {
			return fmt.Errorf("model %s not supported by openai", s.base.model)
		}
		if err := s.instructions.Validate(s.base.provider); err != nil {
			return err
		}
	case s.useElevenLabs:
		if s.base.model != ModelTTSElevenLabs {
			return fmt.Errorf("model %s not supported by elevenlabs", s.base.model)
		}
		if err := s.speed.Validate(s.base.provider); err != nil {
			return err
		}
	}
	return nil
}
