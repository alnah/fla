package config

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/alnah/fla/locale"
	"github.com/alnah/fla/pathutil"
	"github.com/urfave/cli-altsrc/v3"
	json "github.com/urfave/cli-altsrc/v3/json"
	"github.com/urfave/cli/v3"
)

var ErrFilenameTag = errors.New("must be <= 50 characters")

const (
	// general
	envType     string = "CLI_ENV_TYPE"  // dev, test, prod
	envLanguage string = "CLI_IETF_LANG" // fr-FR, pt-BR, en-US
	envLogLevel string = "CLI_LOG_LEVEL" // debug, error, warn, info

	// filename tags
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
	envCacheAddr         string = "REDIS_ADDRESS"
	envCachePassword     string = "REDIS_PASSWORD" // #nosec: G101: safe env key
	envCachePoolTimeout  string = "REDIS_POOL_TIMEOUT"
	envCacheDialTimeout  string = "REDIS_DIAL_TIMEOUT"
	envCacheReadTimeout  string = "REDIS_READ_TIMEOUT"
	envCacheWriteTimeout string = "REDIS_WRITE_TIMEOUT"
)

const (
	// language
	flagLanguage string = "language"

	// log level
	flagLogLevel string = "log-level"

	// filename tags
	flagFilenameDate        string = "filename-date"
	flagFilenameLesson      string = "filename-lesson"
	flagFilenamePreparation string = "filename-preparation"
	flagFilenamePlan        string = "filename-plan"
	flagFilenameReading     string = "filename-reading"
	flagFilenameListening   string = "filename-listening"
	flagFilenameWatching    string = "filename-watching"
	flagFilenameCorrection  string = "filename-correction"
	flagFilenameLevel       string = "filename-level"

	// working directories
	flagInputDir   string = "dir-input"
	flagStagingDir string = "dir-staging"
	flagStudentDir string = "dir-output-student"
	flagTeacherDir string = "dir-output-teacher"
	flagLessonsDir string = "dir-output-lessons"

	// ai client
	flagTimeoutChat string = "ai-timeout-chat"
	flagTimeoutTTS  string = "ai-timeout-tts"
	flagTimeoutSTT  string = "ai-timeout-stt"

	// cache
	flagCachePoolTimeout string = "cache-timeout-pool"

	flagCacheDialTimeout  string = "cache-timeout-dial"
	flagCacheReadTimeout  string = "cache-timeout-read"
	flagCacheWriteTimeout string = "cache-timeout-write"
)

// envOverride applies environment variables to override any Config fields,
// but only when the variable is actually present.
func (l *Manager) envOverride(cfg *Config) error {
	type overrider struct {
		key   string
		apply func(string) error
	}

	overrides := []overrider{
		// secrets only overriden in env vars -----------------------------------------
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
		{
			key:   envCacheAddr,
			apply: func(s string) error { cfg.Cache.Address = s; return nil },
		},
		{
			key:   envCachePassword,
			apply: func(s string) error { cfg.Cache.Password = s; return nil },
		},
		// overrides common between env vars and flags --------------------------------
		// log level
		{
			key: envLogLevel,
			apply: func(s string) error {
				var lvl slog.Level
				if err := lvl.UnmarshalText([]byte(s)); err != nil {
					return fmt.Errorf("parsing %s: %w", envLogLevel, err)
				}
				cfg.LogLevel = &lvl
				return nil
			},
		},
		// language tag
		{
			key: envLanguage,
			apply: func(s string) error {
				var langVal locale.Lang
				if err := langVal.Set(s); err != nil {
					return fmt.Errorf("parsing %s: %w", envLanguage, err)
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
			key: envFilenameLesson,
			apply: func(s string) error {
				if utf8.RuneCountInString(s) >= 50 {
					return fmt.Errorf("parsing %s: %w", envFilenameLesson, ErrFilenameTag)
				}
				cfg.Filename.Lesson = s
				return nil
			},
		},
		{
			key: envFilenamePreparation,
			apply: func(s string) error {
				if utf8.RuneCountInString(s) >= 50 {
					return fmt.Errorf("parsing %s: %w", envFilenamePreparation, ErrFilenameTag)
				}
				cfg.Filename.Preparation = s
				return nil
			},
		},
		{
			key: envFilenamePlan,
			apply: func(s string) error {
				if utf8.RuneCountInString(s) >= 50 {
					return fmt.Errorf("parsing %s: %w", envFilenamePlan, ErrFilenameTag)
				}
				cfg.Filename.Plan = s
				return nil
			},
		},
		{
			key: envFilenameReading,
			apply: func(s string) error {
				if utf8.RuneCountInString(s) >= 50 {
					return fmt.Errorf("parsing %s: %w", envFilenameReading, ErrFilenameTag)
				}
				cfg.Filename.Reading = s
				return nil
			},
		},
		{
			key: envFilenameListening,
			apply: func(s string) error {
				if utf8.RuneCountInString(s) >= 50 {
					return fmt.Errorf("parsing %s: %w", envFilenameListening, ErrFilenameTag)
				}
				cfg.Filename.Listening = s
				return nil
			},
		},
		{
			key: envFilenameWatching,
			apply: func(s string) error {
				if utf8.RuneCountInString(s) >= 50 {
					return fmt.Errorf("parsing %s: %w", envFilenameWatching, ErrFilenameTag)
				}
				cfg.Filename.Watching = s
				return nil
			},
		},
		{
			key: envFilenameCorrection,
			apply: func(s string) error {
				if utf8.RuneCountInString(s) >= 50 {
					return fmt.Errorf("parsing %s: %w", envFilenameCorrection, ErrFilenameTag)
				}
				cfg.Filename.Correction = s
				return nil
			},
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
		{
			key: envCachePoolTimeout,
			apply: func(s string) error {
				d, err := time.ParseDuration(s)
				if err != nil {
					return fmt.Errorf("parsing %s: %w", envCachePoolTimeout, err)
				}
				if d > 0 {
					cfg.Cache.Timeout.Pool = Duration(d)
				}
				return nil
			},
		},
		{
			key: envCacheDialTimeout,
			apply: func(s string) error {
				d, err := time.ParseDuration(s)
				if err != nil {
					return fmt.Errorf("parsing %s: %w", envCacheDialTimeout, err)
				}
				if d > 0 {
					cfg.Cache.Timeout.Dial = Duration(d)
				}
				return nil
			},
		}, {
			key: envCacheReadTimeout,
			apply: func(s string) error {
				d, err := time.ParseDuration(s)
				if err != nil {
					return fmt.Errorf("parsing %s: %w", envCacheReadTimeout, err)
				}
				if d > 0 {
					cfg.Cache.Timeout.Read = Duration(d)
				}
				return nil
			},
		}, {
			key: envCacheWriteTimeout,
			apply: func(s string) error {
				d, err := time.ParseDuration(s)
				if err != nil {
					return fmt.Errorf("parsing %s: %w", envCacheWriteTimeout, err)
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
		return errors.Join(errs...)
	}
	return nil
}

func buildFlags(m Manager) []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "path to the JSON config file",
			Action: func(ctx context.Context, c *cli.Command, b bool) error {
				if b {
					fmt.Println(m.Filepath)
					return nil
				}
				return nil
			},
		},
		&cli.StringFlag{
			Name:  "log-level",
			Usage: "log level (debug|info|warn|error)",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar(envLogLevel),
				json.JSON("loglevel", altsrc.StringSourcer(m.Filepath)),
			),
		},
		&cli.StringFlag{
			Name:  "input",
			Usage: "input directory",
			Sources: cli.NewValueSourceChain(
				json.JSON("input", altsrc.StringSourcer(m.Filepath)),
			),
		},
	}
}
