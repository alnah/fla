package aiclient

func (t *TTS) BaseClient() *Base { return t.base }

func WithVoice(v voice) option[*TTS] {
	return func(t *TTS) { t.Voice = v }
}

func WithText(txt Text) option[*TTS] {
	return func(t *TTS) { t.Text = txt }
}
func WithInstructions(i Instructions) option[*TTS] {
	return func(t *TTS) { t.Instructions = i }
}
func WithSpeed(s Speed) option[*TTS] {
	return func(t *TTS) { t.Speed = s }
}
