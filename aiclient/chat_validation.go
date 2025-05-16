package aiclient

import (
	"errors"
	"fmt"
)

// validate ensures required fields are set.
func (c *Chat) validate() error {
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
	if err := c.maxTokens.Validate(); err != nil {
		return err
	}
	if err := c.temperature.Validate(c.base.Model); err != nil {
		return err
	}
	if err := c.messages.Validate(); err != nil {
		return err
	}
	if c.useOpenAI == c.useAnthropic {
		return errors.New("must configure exactly one provider: openai or anthropic")
	}
	if c.useOpenAI && c.base.provider != ProviderOpenAI {
		return fmt.Errorf("url indicates openai but provider is %s", c.base.provider)
	}
	if c.useAnthropic && c.base.provider != ProviderAnthropic {
		return fmt.Errorf("url indicates anthropic but provider is %s", c.base.provider)
	}
	switch {
	case c.useOpenAI:
		switch c.base.Model {
		case AIModelReasoningOpenAI, AIModelFlagshipOpenAI, AIModelCostOptimizedOpenAI:
			// ok
		default:
			return fmt.Errorf("model %s not supported by openai", c.base.Model)
		}

	case c.useAnthropic:
		switch c.base.Model {
		case AIModelReasoningAnthropic, AIModelCostOptimizedAnthropic:
			// ok
		default:
			return fmt.Errorf("model %s not supported by anthropic", c.base.Model)
		}
	}
	if c.useAnthropic {
		for _, m := range c.messages {
			if m.Role == RoleSystem {
				return errors.New("system message must be passed via system field, not via messages, when using anthropic")
			}
		}
	}
	if c.useOpenAI {
		var systemCount int
		for _, m := range c.messages {
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
