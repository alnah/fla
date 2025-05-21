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

	// ai clients
	envTimeoutChat string = "TIMEOUT_CHAT"
	envTimeoutTTS  string = "TIMEOUT_TTS"
	envTimeoutSTT  string = "TIMEOUT_STT"

	// api keys
	envAPIKeyOpenAI     string = "API_KEY_OPENAI"     // #nosec G101: safe env key
	envAPIKeyAnthropic  string = "API_KEY_ANTHROPIC"  // #nosec G101: safe env key
	envAPIKeyElevenLabs string = "API_KEY_ELEVENLABS" // #nosec G101: safe env key

	// redis
	envRedisAddr         string = "REDIS_ADDRESS"
	envRedisPassword     string = "REDIS_PASSWORD" // #nosec: G101: safe env key
	envRedisPoolTimeout  string = "REDIS_POOL_TIMEOUT"
	envRedisDialTimeout  string = "REDIS_DIAL_TIMEOUT"
	envRedisReadTimeout  string = "REDIS_READ_TIMEOUT"
	envRedisWriteTimeout string = "REDIS_WRITE_TIMEOUT"
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
	type overrider struct {
		key   string
		apply func(string) error
	}

	overrides := []overrider{
		// log level
		{
			key: envCLILogLevel,
			apply: func(s string) error {
				var lvl slog.Level
				if err := lvl.UnmarshalText([]byte(s)); err != nil {
					return fmt.Errorf("parsing %s: %w", envCLILogLevel, err)
				}
				cfg.LogLevel = &lvl
				return nil
			},
		},
		// language tag
		{
			key: envCLILang,
			apply: func(s string) error {
				var langVal locale.Lang
				if err := langVal.Set(s); err != nil {
					return fmt.Errorf("parsing %s: %w", envCLILang, err)
				}
				cfg.Lang = &langVal
				return nil
			},
		},
		// filename boolean flags
		{
			key: envFilenameDate,
			apply: func(s string) error {
				b, err := strconv.ParseBool(s)
				if err != nil {
					return fmt.Errorf("parsing %s: %w", envFilenameDate, err)
				}
				cfg.Filename.Date = b
				return nil
			},
		},
		{
			key: envFilenameLevel,
			apply: func(s string) error {
				b, err := strconv.ParseBool(s)
				if err != nil {
					return fmt.Errorf("parsing %s: %w", envFilenameLevel, err)
				}
				cfg.Filename.Level = b
				return nil
			},
		},
		// filename string tags
		{
			key:   envFilenameLesson,
			apply: func(s string) error { cfg.Filename.Lesson = s; return nil },
		},
		{
			key:   envFilenamePreparation,
			apply: func(s string) error { cfg.Filename.Preparation = s; return nil },
		},
		{
			key:   envFilenamePlan,
			apply: func(s string) error { cfg.Filename.Plan = s; return nil },
		},
		{
			key:   envFilenameReading,
			apply: func(s string) error { cfg.Filename.Reading = s; return nil },
		},
		{
			key:   envFilenameListening,
			apply: func(s string) error { cfg.Filename.Listening = s; return nil },
		},
		{
			key:   envFilenameWatching,
			apply: func(s string) error { cfg.Filename.Watching = s; return nil },
		},
		{
			key:   envFilenameCorrection,
			apply: func(s string) error { cfg.Filename.Correction = s; return nil },
		},
		// directory overrides (with Secure)
		{
			key: envInputDir,
			apply: func(s string) error {
				p, err := pathutil.DirPath(s).Secure()
				if err != nil {
					return fmt.Errorf("parsing %s: %w", envInputDir, err)
				}
				cfg.Dir.Input = p
				return nil
			},
		},
		{
			key: envStagingDir,
			apply: func(s string) error {
				p, err := pathutil.DirPath(s).Secure()
				if err != nil {
					return fmt.Errorf("parsing %s: %w", envStagingDir, err)
				}
				cfg.Dir.Staging = p
				return nil
			},
		},
		{
			key: envStudentDir,
			apply: func(s string) error {
				p, err := pathutil.DirPath(s).Secure()
				if err != nil {
					return fmt.Errorf("parsing %s: %w", envStudentDir, err)
				}
				cfg.Dir.Output.Student = p
				return nil
			},
		},
		{
			key: envTeacherDir,
			apply: func(s string) error {
				p, err := pathutil.DirPath(s).Secure()
				if err != nil {
					return fmt.Errorf("parsing %s: %w", envTeacherDir, err)
				}
				cfg.Dir.Output.Teacher = p
				return nil
			},
		},
		{
			key: envLessonsDir,
			apply: func(s string) error {
				p, err := pathutil.DirPath(s).Secure()
				if err != nil {
					return fmt.Errorf("parsing %s: %w", envLessonsDir, err)
				}
				cfg.Dir.Output.Lessons = p
				return nil
			},
		},
		// timeout overrides
		{
			key: envTimeoutChat,
			apply: func(s string) error {
				d, err := time.ParseDuration(s)
				if err != nil {
					return fmt.Errorf("parsing %s: %w", envTimeoutChat, err)
				}
				if d > 0 {
					cfg.AI.Timeout.Chat = Duration(d)
				}
				return nil
			},
		},
		{
			key: envTimeoutTTS,
			apply: func(v string) error {
				d, err := time.ParseDuration(v)
				if err != nil {
					return fmt.Errorf("parsing %s: %w", envTimeoutTTS, err)
				}
				if d > 0 {
					cfg.AI.Timeout.TTS = Duration(d)
				}
				return nil
			},
		},
		{
			key: envTimeoutSTT,
			apply: func(s string) error {
				d, err := time.ParseDuration(s)
				if err != nil {
					return fmt.Errorf("parsing %s: %w", envTimeoutSTT, err)
				}
				if d > 0 {
					cfg.AI.Timeout.STT = Duration(d)
				}
				return nil
			},
		},
		// API keys
		{
			key:   envAPIKeyOpenAI,
			apply: func(s string) error { cfg.AI.APIKey.OpenAI = s; return nil },
		},
		{
			key:   envAPIKeyAnthropic,
			apply: func(s string) error { cfg.AI.APIKey.Anthropic = s; return nil },
		},
		{
			key:   envAPIKeyElevenLabs,
			apply: func(s string) error { cfg.AI.APIKey.ElevenLabs = s; return nil },
		},
		// redis
		{
			key:   envRedisAddr,
			apply: func(s string) error { cfg.Cache.Address = s; return nil },
		},
		{
			key:   envRedisPassword,
			apply: func(s string) error { cfg.Cache.Password = s; return nil },
		},
		{
			key: envRedisPoolTimeout,
			apply: func(s string) error {
				d, err := time.ParseDuration(s)
				if err != nil {
					return fmt.Errorf("parsing %s: %w", envRedisPoolTimeout, err)
				}
				if d > 0 {
					cfg.Cache.Timeout.Pool = Duration(d)
				}
				return nil
			},
		},
		{
			key: envRedisDialTimeout,
			apply: func(s string) error {
				d, err := time.ParseDuration(s)
				if err != nil {
					return fmt.Errorf("parsing %s: %w", envRedisDialTimeout, err)
				}
				if d > 0 {
					cfg.Cache.Timeout.Dial = Duration(d)
				}
				return nil
			},
		}, {
			key: envRedisReadTimeout,
			apply: func(s string) error {
				d, err := time.ParseDuration(s)
				if err != nil {
					return fmt.Errorf("parsing %s: %w", envRedisReadTimeout, err)
				}
				if d > 0 {
					cfg.Cache.Timeout.Read = Duration(d)
				}
				return nil
			},
		}, {
			key: envRedisWriteTimeout,
			apply: func(s string) error {
				d, err := time.ParseDuration(s)
				if err != nil {
					return fmt.Errorf("parsing %s: %w", envRedisWriteTimeout, err)
				}
				if d > 0 {
					cfg.Cache.Timeout.Write = Duration(d)
				}
				return nil
			},
		},
	}

	var errs []error
	for _, o := range overrides {
		if v, ok := l.env.LookupEnv(o.key); ok {
			if err := o.apply(v); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("environment override errors: %w", errors.Join(errs...))
	}
	return nil
}
