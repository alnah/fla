package config

import (
	"context"
	"log/slog"

	"github.com/alnah/fla/locale"
)

// Config groups all user-tweakable settings so the application
// can drive file naming, directories, timeouts, and API keys from one source.
type Config struct {
	LogLevel *slog.Level  `json:"loglevel,omitempty"` // control verbosity for diagnosing issues
	Lang     *locale.Lang `json:"lang,omitempty"`     // set locale to adapt content language

	// Filename prefixes and flags ensure outputs follow a consistent naming scheme.
	Filename struct {
		Date        bool   `json:"date,omitempty"`        // include date for versioning
		Level       bool   `json:"level,omitempty"`       // append CEFR level for clarity
		Lesson      string `json:"lesson,omitempty"`      // tag lesson documents
		Preparation string `json:"preparation,omitempty"` // tag prep materials
		Plan        string `json:"plan,omitempty"`        // tag teacher plans
		Reading     string `json:"reading,omitempty"`     // tag reading exercises
		Listening   string `json:"listening,omitempty"`   // tag audio exercises
		Watching    string `json:"watching,omitempty"`    // tag video exercises
		Correction  string `json:"correction,omitempty"`  // tag correction sheets
	} `json:"filename"`

	// Dir holds input, staging, and output locations to separate workflows.
	Dir struct {
		Input   string `json:"input,omitempty"`
		Staging string `json:"staging,omitempty"`
		Output  struct {
			Student string `json:"student,omitempty"` // store student-facing content
			Teacher string `json:"teacher,omitempty"` // store teacher content
			Lessons string `json:"lessons,omitempty"` // catch-all directory
		} `json:"output"`
	} `json:"directories"`

	// Timeout values protect against hung external calls.
	Timeout struct {
		Chat Duration `json:"chat,string,omitempty"` // bound AI chat responses
		TTS  Duration `json:"tts,string,omitempty"`  // bound audio synthesis
		STT  Duration `json:"stt,string,omitempty"`  // bound speech recognition
	} `json:"timeouts"`

	// APIKey holds credentials so components can authenticate to services.
	APIKey struct {
		OpenAI     string `json:"-"`
		Anthropic  string `json:"-"`
		ElevenLabs string `json:"-"`
	} `json:"-"`
}

// ChatContext derives a context with the configured chat timeout
// so that AI requests don’t block indefinitely.
func (c *Config) ChatContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, c.Timeout.Chat.ToTimeDuration())
}

// TTSContext derives a context with the configured TTS timeout.
func (c *Config) TTSContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, c.Timeout.TTS.ToTimeDuration())
}

// STTContext derives a context with the configured STT timeout.
func (c *Config) STTContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, c.Timeout.STT.ToTimeDuration())
}
