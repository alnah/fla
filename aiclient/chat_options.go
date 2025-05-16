package aiclient

func (c *ChatClient) BaseClient() *baseClient { return c.base }

func WithTemperature(t Temperature) option[*ChatClient] {
	return func(c *ChatClient) { c.temperature = t }
}
func WithMessages(ms Messages) option[*ChatClient] {
	return func(c *ChatClient) { c.messages = ms }
}
func WithMaxTokens(mt MaxTokens) option[*ChatClient] {
	return func(c *ChatClient) { c.maxTokens = mt }
}
