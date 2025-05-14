package aiclient

func (c *ChatClient) BaseClient() *BaseClient { return c.base }

func WithTemperature(t Temperature) Option[*ChatClient] {
	return func(c *ChatClient) { c.Temperature = t }
}
func WithMessages(ms Messages) Option[*ChatClient] {
	return func(c *ChatClient) { c.Messages = ms }
}
func WithMaxTokens(mt MaxTokens) Option[*ChatClient] {
	return func(c *ChatClient) { c.MaxTokens = mt }
}
