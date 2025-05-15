package aiclient

// transcriptionResponse holds the text after the transcription is done.
type transcriptionResponse struct {
	Text string `json:"text,omitempty"`
}

// Content returns the text content from a transcription.
func (t transcriptionResponse) Content() string { return t.Text }
