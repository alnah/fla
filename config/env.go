package config

import (
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

type Env interface {
	Type() string
	Get(key string) string
	LookupEnv(key string) (string, bool)
}

type configEnv struct{}

var env = &configEnv{}

func (e *configEnv) Type() string {
	val := os.Getenv(envCLIType)
	if val == "" {
		return "prod"
	}
	return val
}

func (e *configEnv) Get(key string) string {
	return os.Getenv(key)
}

func (e *configEnv) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

func (l *loader) envOverride(cfg *Config) error {
	// log level
	if v, ok := l.env.LookupEnv(envCLILogLevel); ok {
		var lvl slog.Level
		err := lvl.UnmarshalText([]byte(v))
		if err != nil {
			return fmt.Errorf("parsing %s: %w", envCLILogLevel, err)
		}
		cfg.LogLevel = &lvl
	}

	// language
	if v, ok := l.env.LookupEnv(envCLILang); ok {
		var langVal locale.Lang
		if err := langVal.Set(v); err != nil {
			return fmt.Errorf("parsing %s: %w", envCLILang, err)
		}
		cfg.Lang = &langVal
	}

	// filename options: bool flags
	if b, err := l.parseBool(envFilenameDate); err != nil {
		return fmt.Errorf("parsing %s: %w", envFilenameDate, err)
	} else {
		cfg.Filename.Date = b
	}
	if b, err := l.parseBool(envFilenameLevel); err != nil {
		return fmt.Errorf("parsing %s: %w", envFilenameLevel, err)
	} else {
		cfg.Filename.Level = b
	}

	// filename options: string flags
	cfg.Filename.Lesson = l.parseString(envFilenameLesson)
	cfg.Filename.Preparation = l.parseString(envFilenamePreparation)
	cfg.Filename.Plan = l.parseString(envFilenamePlan)
	cfg.Filename.Reading = l.parseString(envFilenameReading)
	cfg.Filename.Listening = l.parseString(envFilenameListening)
	cfg.Filename.Watching = l.parseString(envFilenameWatching)
	cfg.Filename.Correction = l.parseString(envFilenameCorrection)

	// working directories
	if d, err := l.parseDir(envInputDir); err != nil {
		return fmt.Errorf("parsing %s: %w", envInputDir, err)
	} else if d != "" {
		cfg.Dir.Input = d
	}
	if d, err := l.parseDir(envStagingDir); err != nil {
		return fmt.Errorf("parsing %s: %w", envStagingDir, err)
	} else if d != "" {
		cfg.Dir.Staging = d
	}
	if d, err := l.parseDir(envStudentDir); err != nil {
		return fmt.Errorf("parsing %s: %w", envStudentDir, err)
	} else if d != "" {

		cfg.Dir.Output.Student = d
	}
	if d, err := l.parseDir(envTeacherDir); err != nil {
		return fmt.Errorf("parsing %s: %w", envTeacherDir, err)
	} else if d != "" {

		cfg.Dir.Output.Teacher = d
	}
	if d, err := l.parseDir(envLessonsDir); err != nil {
		return fmt.Errorf("parsing %s: %w", envLessonsDir, err)
	} else if d != "" {
		cfg.Dir.Output.Lessons = d
	}

	// timeouts
	if d, err := l.parseDuration(envTimeoutChat); err != nil {
		return fmt.Errorf("parsing %s: %w", envTimeoutChat, err)
	} else if d > 0 {
		cfg.Timeout.Chat = d
	}
	if d, err := l.parseDuration(envTimeoutTTS); err != nil {
		return fmt.Errorf("parsing %s: %w", envTimeoutTTS, err)
	} else if d > 0 {
		cfg.Timeout.TTS = d
	}
	if d, err := l.parseDuration(envTimeoutSTT); err != nil {
		return fmt.Errorf("parsing %s: %w", envTimeoutSTT, err)
	} else if d > 0 {
		cfg.Timeout.STT = d
	}

	// api keys
	cfg.APIKey.OpenAI = l.parseString(envAPIKeyOpenAI)
	cfg.APIKey.Anthropic = l.parseString(envAPIKeyAnthropic)
	cfg.APIKey.ElevenLabs = l.parseString(envAPIKeyElevenLabs)

	return nil
}

func (l *loader) parseBool(key string) (bool, error) {
	if s, ok := l.env.LookupEnv(key); ok {
		return strconv.ParseBool(s)
	}
	return false, nil
}

func (l *loader) parseString(key string) string {
	if s, ok := l.env.LookupEnv(key); ok {
		return s
	}
	return ""
}

func (l *loader) parseDir(key string) (string, error) {
	if d, ok := l.env.LookupEnv(key); ok {
		return pathutil.DirPath(d).Secure()
	}
	return "", nil
}

func (l *loader) parseDuration(key string) (time.Duration, error) {
	if s, ok := l.env.LookupEnv(key); ok {
		return time.ParseDuration(s)
	}
	return 0, nil
}
