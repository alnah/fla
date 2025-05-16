package aiclient

func (t *TTSClient) BaseClient() *baseClient { return t.base }

func WithVoice(v voice) option[*TTSClient] {
	return func(s *TTSClient) { s.voice = v }
}

func WithText(txt Text) option[*TTSClient] {
	return func(s *TTSClient) { s.text = txt }
}
func WithInstructions(i Instructions) option[*TTSClient] {
	return func(s *TTSClient) { s.instructions = i }
}
func WithSpeed(sp Speed) option[*TTSClient] {
	return func(s *TTSClient) { s.speed = sp }
}
