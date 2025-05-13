package aiclient

import (
	"errors"
	"fmt"
)

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
	if c.UseOpenAI && c.provider != ProviderOpenAI {
		return fmt.Errorf("URL indicates OpenAI but provider is %s", c.provider)
	}
	if c.UseAnthropic && c.provider != ProviderAnthropic {
		return fmt.Errorf("URL indicates Anthropic but provider is %s", c.provider)
	}
	switch {
	case c.UseOpenAI:
		switch c.Model {
		case AIModelReasoningOpenAI,
			AIModelFlagshipOpenAI,
			AIModelCostOptimizedOpenAI,
			AIModelTTSOpenAI,
			AIModelTranscriptionOpenAI:
			// ok
		default:
			return fmt.Errorf("model %s not supported by OpenAI", c.Model)
		}

	case c.UseAnthropic:
		switch c.Model {
		case AIModelReasoningAnthropic,
			AIModelCostOptimizedAnthropic:
			// ok
		default:
			return fmt.Errorf("model %s not supported by Anthropic", c.Model)
		}
	}
	if c.UseAnthropic {
		for _, m := range c.Messages {
			if m.Role == RoleSystem {
				return errors.New("system messages must be passed via c.System, not in c.Messages, when using Anthropic")
			}
		}
	}
	return nil
}
