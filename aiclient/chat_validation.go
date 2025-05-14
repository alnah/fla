package aiclient

import (
	"errors"
	"fmt"
)

// validate ensures required fields are set.
func (c *ChatClient) validate() error {
	if c.base.ctx == nil {
		return errors.New("context must be provided")
	}
	if c.base.logger == nil {
		return errors.New("logger must be set")
	}
	if c.base.httpClient == nil {
		return errors.New("http client must be set")
	}
	if err := c.base.provider.Validate(); err != nil {
		return err
	}
	if err := c.base.url.Validate(); err != nil {
		return err
	}
	if err := c.base.apiKey.Validate(); err != nil {
		return err
	}
	if err := c.base.httpMethod.Validate(); err != nil {
		return err
	}
	if err := c.base.Model.Validate(); err != nil {
		return err
	}
	if err := c.MaxTokens.Validate(); err != nil {
		return err
	}
	if err := c.Temperature.Validate(c.base.Model); err != nil {
		return err
	}
	if err := c.Messages.Validate(); err != nil {
		return err
	}
	if c.UseOpenAI == c.UseAnthropic {
		return errors.New("must configure exactly one provider: openai or anthropic")
	}
	if c.UseOpenAI && c.base.provider != ProviderOpenAI {
		return fmt.Errorf("url indicates openai but provider is %s", c.base.provider)
	}
	if c.UseAnthropic && c.base.provider != ProviderAnthropic {
		return fmt.Errorf("url indicates anthropic but provider is %s", c.base.provider)
	}
	switch {
	case c.UseOpenAI:
		switch c.base.Model {
		case AIModelReasoningOpenAI,
			AIModelFlagshipOpenAI,
			AIModelCostOptimizedOpenAI,
			AIModelTTSOpenAI,
			AIModelTranscriptionOpenAI:
			// ok
		default:
			return fmt.Errorf("model %s not supported by openai", c.base.Model)
		}

	case c.UseAnthropic:
		switch c.base.Model {
		case AIModelReasoningAnthropic,
			AIModelCostOptimizedAnthropic:
			// ok
		default:
			return fmt.Errorf("model %s not supported by anthropic", c.base.Model)
		}
	}
	if c.UseAnthropic {
		for _, m := range c.Messages {
			if m.Role == RoleSystem {
				return errors.New("system message must be passed via system field, not via messages, when using anthropic")
			}
		}
	}
	if c.UseOpenAI {
		var systemCount int
		for _, m := range c.Messages {
			if m.Role == RoleSystem {
				systemCount++
			}
		}
		if systemCount > 1 {
			return errors.New("system message must be passed once, when using openai")
		}
	}
	return nil
}
