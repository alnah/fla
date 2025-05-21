package config

import (
	"context"
	"log/slog"
	"time"

	"github.com/alnah/fla/locale"
	"github.com/alnah/fla/storage/cache"
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
	// AI holds the timeouts, and the api keys
	AI struct {
		// Timeout values protect against hung external calls.
		Timeout struct {
			Chat Duration `json:"chat,omitempty"` // bound AI chat responses
			TTS  Duration `json:"tts,omitempty"`  // bound audio synthesis
			STT  Duration `json:"stt,omitempty"`  // bound speech recognition
		} `json:"timeout"`
		// APIKey holds credentials so components can authenticate to services.
		APIKey struct {
			OpenAI     string `json:"-"`
			Anthropic  string `json:"-"`
			ElevenLabs string `json:"-"`
		} `json:"-"`
	} `json:"ai"`
	// Cache holds configuration parameters for the Cache client.
	Cache struct {
		Address  string `json:"-"`
		Password string `json:"-"`
		Timeout  struct {
			Pool  Duration `json:"pool,,omitempty"`
			Dial  Duration `json:"dial,,omitempty"`
			Read  Duration `json:"read,,omitempty"`
			Write Duration `json:"write,,omitempty"`
		} `json:"timeouts"`
	} `json:"cache"`
}

// ChatContext derives a context with the configured chat timeout
// so that AI requests don’t block indefinitely.
func (c *Config) ChatContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, c.AI.Timeout.Chat.ToTimeDuration())
}

// TTSContext derives a context with the configured TTS timeout.
func (c *Config) TTSContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, c.AI.Timeout.TTS.ToTimeDuration())
}

// STTContext derives a context with the configured STT timeout.
func (c *Config) STTContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, c.AI.Timeout.STT.ToTimeDuration())
}

// CacheTimeout derives a context with the configured Cache timeout.
// Should be used for blocking operations.
func (c *Config) CacheTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, cache.RedisWriteTimeout+cache.RedisReadTimeout+100*time.Millisecond)
}
