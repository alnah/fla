package aiclient

import "errors"

func (s *Speech) validate() error {
	if s.base.ctx == nil {
		return errors.New("context must be provided")
	}
	if s.base.logger == nil {
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
	if err := s.base.Model.Validate(); err != nil {
		return err
	}
	if err := s.Voice.Validate(s.base.provider); err != nil {
		return err
	}
	if err := s.Text.Validate(); err != nil {
		return err
	}
	if err := s.Instructions.Validate(s.base.provider); err != nil {
		return err
	}
	if err := s.Speed.Validate(s.base.provider); err != nil {
		return err
	}
	return nil
}
