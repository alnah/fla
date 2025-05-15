package aiclient

func (c *Chat) BaseClient() *Base { return c.base }

func WithTemperature(t Temperature) option[*Chat] {
	return func(c *Chat) { c.temperature = t }
}
func WithMessages(ms Messages) option[*Chat] {
	return func(c *Chat) { c.messages = ms }
}
func WithMaxTokens(mt MaxTokens) option[*Chat] {
	return func(c *Chat) { c.maxTokens = mt }
}
