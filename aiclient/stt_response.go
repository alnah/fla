package aiclient

// transcriptResponse holds the text after the transcription is done.
type transcriptResponse struct {
	Text string `json:"text,omitempty"`
}

// Content returns the text content from a transcription.
func (t transcriptResponse) Content() string { return t.Text }
