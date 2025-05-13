package aiclient

// Transcription holds the text after the transcription is done.
type Transcription struct {
	Text string `json:"text,omitempty"`
}

// Content returns the text content from a transcription.
func (t Transcription) Content() string { return t.Text }
