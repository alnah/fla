package aiclient

import "errors"

// validate ensures required fields are set.
func (c *ChatClient) validate() error {
	if err := c.Model.Validate(); err != nil {
		return err
	}
	if err := c.Temperature.Validate(c.Model); err != nil {
		return err
	}
	if err := c.Messages.Validate(); err != nil {
		return err
	}
	if err := c.MaxTokens.Validate(); err != nil {
		return err
	}
	if err := c.provider.Validate(); err != nil {
		return err
	}
	if err := c.url.Validate(); err != nil {
		return err
	}
	if err := c.apiKey.Validate(); err != nil {
		return err
	}
	if err := c.httpMethod.Validate(); err != nil {
		return err
	}
	if c.logger == nil {
		return errors.New("logger must be set")
	}
	if c.ctx == nil {
		return errors.New("context must be provided")
	}
	if c.UseOpenAI == c.UseAnthropic {
		return errors.New("must configure exactly one provider: OpenAI or Anthropic")
	}
	return nil
}
