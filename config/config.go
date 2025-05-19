package config

import (
	"context"
	"log/slog"
	"time"

	"github.com/alnah/fla/locale"
)

type Config struct {
	LogLevel *slog.Level  `json:"loglevel,omitempty"` // debug, error, warn, info
	Lang     *locale.Lang `json:"lang,omitempty"`     // fr-FR, pt-BR, en-US

	Filename struct {
		Date        bool   `json:"date,omitempty"`        // prepend current date (YYYY-MM-DD)
		Level       bool   `json:"level,omitempty"`       // append CEFR level (e.g. A1, B2)
		Lesson      string `json:"lesson,omitempty"`      // prefix for lesson content
		Preparation string `json:"preparation,omitempty"` // prefix for preparation materials
		Plan        string `json:"plan,omitempty"`        // prefix for teacher’s plan
		Reading     string `json:"reading,omitempty"`     // prefix for reading exercises
		Listening   string `json:"listening,omitempty"`   // prefix for listening exercises
		Watching    string `json:"watching,omitempty"`    // prefix for watching exercises
		Correction  string `json:"correction,omitempty"`  // prefix for correction
	} `json:"filename"`

	Dir struct {
		Input   string `json:"input,omitempty"`
		Staging string `json:"staging,omitempty"`
		Output  struct {
			Student string `json:"student,omitempty"` // for preparation, reading, listening, watching
			Teacher string `json:"teacher,omitempty"` // for teacher's lesson, and plan
			Lessons string `json:"lessons,omitempty"` // for everything
		} `json:"output"`
	}

	// infrastructure
	Timeout struct {
		Chat time.Duration `json:"chat,string,omitempty"` // chat completion timeout
		TTS  time.Duration `json:"tts,string,omitempty"`  // text-to-speech audio timeout
		STT  time.Duration `json:"stt,string,omitempty"`  // speech-to-text transcript timeout
	} `json:"timeout"`

	apiKey struct {
		openai     string
		anthropic  string
		elevenlabs string
	}
}

func (c *Config) APIKeyOpenAI() string     { return c.apiKey.openai }
func (c *Config) APIKeyAnthropic() string  { return c.apiKey.anthropic }
func (c *Config) APIKeyElevenLabs() string { return c.apiKey.elevenlabs }

func (c *Config) ChatContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, c.Timeout.Chat)
}

func (c *Config) TTSContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, c.Timeout.TTS)
}

func (c *Config) STTContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, c.Timeout.STT)
}
