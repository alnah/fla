package aiclient

func (t *TTSClient) BaseClient() *BaseClient { return t.base }

func WithVoice(v Voice) Option[*TTSClient] {
	return func(t *TTSClient) { t.Voice = v }
}

func WithText(txt Text) Option[*TTSClient] {
	return func(t *TTSClient) { t.Text = txt }
}
func WithInstructions(i Instructions) Option[*TTSClient] {
	return func(t *TTSClient) { t.Instructions = i }
}
func WithSpeed(s Speed) Option[*TTSClient] {
	return func(t *TTSClient) { t.Speed = s }
}
