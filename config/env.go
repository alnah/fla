package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/alnah/fla/locale"
	"github.com/alnah/fla/pathutil"
)

const (
	// general
	envCLIType     string = "CLI_ENV_TYPE"  // dev, test, prod
	envCLILang     string = "CLI_IETF_LANG" // fr-FR, pt-BR, en-US
	envCLILogLevel string = "CLI_LOG_LEVEL" // debug, error, warn, info

	// filename options
	envFilenameDate        string = "FILENAME_DATE"
	envFilenameLevel       string = "FILENAME_LEVEL"
	envFilenameLesson      string = "FILENAME_LESSON"
	envFilenamePreparation string = "FILENAME_PREPARATION"
	envFilenamePlan        string = "FILENAME_PLAN"
	envFilenameReading     string = "FILENAME_READING"
	envFilenameListening   string = "FILENAME_LISTENING"
	envFilenameWatching    string = "FILENAME_WATCHING"
	envFilenameCorrection  string = "FILENAME_CORRECTION"

	// working directories
	envInputDir   string = "DIR_INPUT"
	envStagingDir string = "DIR_STAGING"
	envStudentDir string = "DIR_STUDENT"
	envTeacherDir string = "DIR_TEACHER"
	envLessonsDir string = "DIR_LESSONS"

	// timeouts
	envTimeoutChat string = "TIMEOUT_CHAT"
	envTimeoutTTS  string = "TIMEOUT_TTS"
	envTimeoutSTT  string = "TIMEOUT_STT"

	// api keys
	envAPIKeyOpenAI     string = "API_KEY_OPENAI"     // #nosec G101: safe env key
	envAPIKeyAnthropic  string = "API_KEY_ANTHROPIC"  // #nosec G101: safe env key
	envAPIKeyElevenLabs string = "API_KEY_ELEVENLABS" // #nosec G101: safe env key
)

// Env defines how to read environment variables and distinguish environments.
type Env interface {
	Type() string                        // dev, test, or prod
	Get(key string) string               // retrieve raw value
	LookupEnv(key string) (string, bool) // detect presence
}

// configEnv uses the OS environment to implement Env,
// so settings can be overridden without code changes.
type configEnv struct{}

var env = &configEnv{}

// Type returns the CLI environment type to select appropriate defaults.
func (e *configEnv) Type() string {
	val := os.Getenv(envCLIType)
	if val == "" {
		return "prod"
	}
	return val
}

// Get returns the value of key or empty if unset.
func (e *configEnv) Get(key string) string {
	return os.Getenv(key)
}

// LookupEnv checks if key is present and returns its value.
func (e *configEnv) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

// envOverride applies environment variables to override any Config fields,
// but only when the variable is actually present.
func (l *manager) envOverride(cfg *Config) error {
	var accErrors []error

	// override log level if provided
	if v, ok := l.env.LookupEnv(envCLILogLevel); ok {
		var lvl slog.Level
		if err := lvl.UnmarshalText([]byte(v)); err != nil {
			accErrors = append(accErrors, fmt.Errorf("parsing %s: %w", envCLILogLevel, err))
		} else {
			cfg.LogLevel = &lvl
		}
	}

	// override language tag if provided
	if v, ok := l.env.LookupEnv(envCLILang); ok {
		var langVal locale.Lang
		if err := langVal.Set(v); err != nil {
			accErrors = append(accErrors, fmt.Errorf("parsing %s: %w", envCLILang, err))
		} else {
			cfg.Lang = &langVal
		}
	}

	// parse boolean filename flags if provided
	if s, ok := l.env.LookupEnv(envFilenameDate); ok {
		b, err := strconv.ParseBool(s)
		if err != nil {
			accErrors = append(accErrors, fmt.Errorf("parsing %s: %w", envFilenameDate, err))
		} else {
			cfg.Filename.Date = b
		}
	}
	if s, ok := l.env.LookupEnv(envFilenameLevel); ok {
		b, err := strconv.ParseBool(s)
		if err != nil {
			accErrors = append(accErrors, fmt.Errorf("parsing %s: %w", envFilenameLevel, err))
		} else {
			cfg.Filename.Level = b
		}
	}

	// parse string filename tags if provided
	if v, ok := l.env.LookupEnv(envFilenameLesson); ok {
		cfg.Filename.Lesson = v
	}
	if v, ok := l.env.LookupEnv(envFilenamePreparation); ok {
		cfg.Filename.Preparation = v
	}
	if v, ok := l.env.LookupEnv(envFilenamePlan); ok {
		cfg.Filename.Plan = v
	}
	if v, ok := l.env.LookupEnv(envFilenameReading); ok {
		cfg.Filename.Reading = v
	}
	if v, ok := l.env.LookupEnv(envFilenameListening); ok {
		cfg.Filename.Listening = v
	}
	if v, ok := l.env.LookupEnv(envFilenameWatching); ok {
		cfg.Filename.Watching = v
	}
	if v, ok := l.env.LookupEnv(envFilenameCorrection); ok {
		cfg.Filename.Correction = v
	}

	// parse directory overrides if provided
	if d, ok := l.env.LookupEnv(envInputDir); ok {
		if path, err := pathutil.DirPath(d).Secure(); err != nil {
			accErrors = append(accErrors, fmt.Errorf("parsing %s: %w", envInputDir, err))
		} else {
			cfg.Dir.Input = path
		}
	}
	if d, ok := l.env.LookupEnv(envStagingDir); ok {
		if path, err := pathutil.DirPath(d).Secure(); err != nil {
			accErrors = append(accErrors, fmt.Errorf("parsing %s: %w", envStagingDir, err))
		} else {
			cfg.Dir.Staging = path
		}
	}
	if d, ok := l.env.LookupEnv(envStudentDir); ok {
		if path, err := pathutil.DirPath(d).Secure(); err != nil {
			accErrors = append(accErrors, fmt.Errorf("parsing %s: %w", envStudentDir, err))
		} else {
			cfg.Dir.Output.Student = path
		}
	}
	if d, ok := l.env.LookupEnv(envTeacherDir); ok {
		if path, err := pathutil.DirPath(d).Secure(); err != nil {
			accErrors = append(accErrors, fmt.Errorf("parsing %s: %w", envTeacherDir, err))
		} else {
			cfg.Dir.Output.Teacher = path
		}
	}
	if d, ok := l.env.LookupEnv(envLessonsDir); ok {
		if path, err := pathutil.DirPath(d).Secure(); err != nil {
			accErrors = append(accErrors, fmt.Errorf("parsing %s: %w", envLessonsDir, err))
		} else {
			cfg.Dir.Output.Lessons = path
		}
	}

	// parse timeouts if provided
	if s, ok := l.env.LookupEnv(envTimeoutChat); ok {
		if dur, err := time.ParseDuration(s); err != nil {
			accErrors = append(accErrors, fmt.Errorf("parsing %s: %w", envTimeoutChat, err))
		} else if dur > 0 {
			cfg.Timeout.Chat = Duration(dur)
		}
	}
	if s, ok := l.env.LookupEnv(envTimeoutTTS); ok {
		if dur, err := time.ParseDuration(s); err != nil {
			accErrors = append(accErrors, fmt.Errorf("parsing %s: %w", envTimeoutTTS, err))
		} else if dur > 0 {
			cfg.Timeout.TTS = Duration(dur)
		}
	}
	if s, ok := l.env.LookupEnv(envTimeoutSTT); ok {
		if dur, err := time.ParseDuration(s); err != nil {
			accErrors = append(accErrors, fmt.Errorf("parsing %s: %w", envTimeoutSTT, err))
		} else if dur > 0 {
			cfg.Timeout.STT = Duration(dur)
		}
	}

	// pull in API keys if provided
	if v, ok := l.env.LookupEnv(envAPIKeyOpenAI); ok {
		cfg.APIKey.OpenAI = v
	}
	if v, ok := l.env.LookupEnv(envAPIKeyAnthropic); ok {
		cfg.APIKey.Anthropic = v
	}
	if v, ok := l.env.LookupEnv(envAPIKeyElevenLabs); ok {
		cfg.APIKey.ElevenLabs = v
	}

	if len(accErrors) > 0 {
		return fmt.Errorf("environment override errors: %w", errors.Join(accErrors...))
	}
	return nil
}
