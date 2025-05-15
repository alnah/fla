package aiclient

func (t *Speech) BaseClient() *Base { return t.base }

func WithVoice(v voice) option[*Speech] {
	return func(s *Speech) { s.Voice = v }
}

func WithText(txt Text) option[*Speech] {
	return func(s *Speech) { s.Text = txt }
}
func WithInstructions(i Instructions) option[*Speech] {
	return func(s *Speech) { s.Instructions = i }
}
func WithSpeed(sp Speed) option[*Speech] {
	return func(s *Speech) { s.Speed = sp }
}
