package aiclient

import "errors"

func (t *TTSClient) validate() error {
	if t.base.ctx == nil {
		return errors.New("context must be provided")
	}
	if t.base.logger == nil {
		return errors.New("logger must be set")
	}
	if t.base.httpClient == nil {
		return errors.New("http client must be set")
	}
	if err := t.base.provider.Validate(); err != nil {
		return err
	}
	if err := t.base.url.Validate(); err != nil {
		return err
	}
	if err := t.base.apiKey.Validate(); err != nil {
		return err
	}
	if err := t.base.httpMethod.Validate(); err != nil {
		return err
	}
	if err := t.base.Model.Validate(); err != nil {
		return err
	}
	if err := t.Voice.Validate(t.base.provider); err != nil {
		return err
	}
	if err := t.Text.Validate(); err != nil {
		return err
	}
	if err := t.Instructions.Validate(t.base.provider); err != nil {
		return err
	}
	if err := t.Speed.Validate(t.base.provider); err != nil {
		return err
	}
	return nil
}
