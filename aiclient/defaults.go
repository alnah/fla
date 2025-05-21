package aiclient

import "time"

const (
	// chat completion timeout
	ChatTimeout time.Duration = 30 * time.Second
	// text-to-speech audio timeout
	TTSTimeout time.Duration = 5 * time.Minute
	// speech-to-text transcript timeout
	STTTimeout time.Duration = 30 * time.Second
)
