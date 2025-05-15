package aiclient

func (t *Speech) BaseClient() *Base { return t.base }

func WithVoice(v voice) option[*Speech] {
	return func(t *Speech) { t.Voice = v }
}

func WithText(txt Text) option[*Speech] {
	return func(t *Speech) { t.Text = txt }
}
func WithInstructions(i Instructions) option[*Speech] {
	return func(t *Speech) { t.Instructions = i }
}
func WithSpeed(s Speed) option[*Speech] {
	return func(t *Speech) { t.Speed = s }
}
