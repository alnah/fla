package config

import (
	"log/slog"
	"path/filepath"

	ai "github.com/alnah/fla/aiclient"
	"github.com/alnah/fla/filesystem"
	"github.com/alnah/fla/locale"
	"github.com/alnah/fla/storage/cache"
)

// RedisStoreLogicalDBs defines the number of logical databases used in the Redis store to prevent key namespace collisions.
// All databases share the same memory, resources, and threads. With 16 DBs, indices range from 0 to 15.
const RedisStoreLogicalDBs int = 16

// defaults assigns sensible fallbacks so users need only override what matters.
func (c *Config) defaults() error {
	// ensure there's always a log level
	defaultPtr(&c.LogLevel, slog.LevelError)

	// default to French lessons unless overridden
	defaultPtr(&c.Lang, locale.LangFrFR)

	// embed date and level in filenames by default
	defaultVal(&c.Filename.Date, true)
	defaultVal(&c.Filename.Level, true)
	defaultVal(&c.Filename.Lesson, "Leçon")
	defaultVal(&c.Filename.Preparation, "Préparation")
	defaultVal(&c.Filename.Plan, "Plan")
	defaultVal(&c.Filename.Reading, "Texte")
	defaultVal(&c.Filename.Listening, "Audio")
	defaultVal(&c.Filename.Watching, "Vidéo")
	defaultVal(&c.Filename.Correction, "Correction")

	// set up working directories under user’s home/config areas
	if err := defaultEmbedDir(&c.Dir.Input, "input"); err != nil {
		return err
	}
	if err := defaultEmbedDir(&c.Dir.Staging, "staging"); err != nil {
		return err
	}
	if err := defaultEmbedDir(&c.Dir.Output.Student, "student"); err != nil {
		return err
	}
	if err := defaultEmbedDir(&c.Dir.Output.Teacher, "teacher"); err != nil {
		return err
	}
	if err := defaultEmbedDir(&c.Dir.Output.Lessons, "lessons"); err != nil {
		return err
	}

	// ai clients' timeouts
	defaultVal(&c.AI.Timeout.Chat, Duration(ai.ChatTimeout))
	defaultVal(&c.AI.Timeout.TTS, Duration(ai.TTSTimeout))
	defaultVal(&c.AI.Timeout.STT, Duration(ai.STTTimeout))

	// redis cache store
	defaultVal(&c.Cache.Timeout.Pool, Duration(cache.RedisPoolTimeout))
	defaultVal(&c.Cache.Timeout.Dial, Duration(cache.RedisDialTimeout))
	defaultVal(&c.Cache.Timeout.Write, Duration(cache.RedisWriteTimeout))
	defaultVal(&c.Cache.Timeout.Write, Duration(cache.RedisWriteTimeout))

	return nil
}

// defaultConfigFS returns a FileSystem rooted in the user’s config directory
// so that settings persist across runs.
func defaultConfigFS() (filesystem.FileSystem, error) {
	userConfigDir, err := user.ConfigDir()
	if err != nil {
		return nil, err
	}
	configDir := filepath.Join(userConfigDir, appName)
	return filesystem.New(configDir), nil
}

// defaultHomeFS gives full filesystem access for embedding lesson directories.
func defaultHomeFS() (filesystem.FileSystem, error) {
	_, err := user.HomeDir()
	if err != nil {
		return nil, err
	}
	return filesystem.New("/"), nil
}

// defaultConfigFilename chooses a filename variant based on environment,
// enabling separate dev/test/prod settings.
func defaultConfigFilename() (string, error) {
	switch env.Type() {
	case "dev":
		return "config.dev.json", nil
	case "test":
		return "config.test.json", nil
	default:
		return "config.json", nil
	}
}

// defaultPtr sets a pointer field only when it’s nil,
// so user overrides are respected.
func defaultPtr[T any](field **T, def T) {
	if *field == nil {
		*field = &def
	}
}

// defaultVal sets a value when it’s zero,
// ensuring flags default to true or meaningful strings.
func defaultVal[T comparable](field *T, def T) {
	var zero T
	if *field == zero {
		*field = def
	}
}

// defaultEmbedDir auto-generates a directory under the app’s home area
// to simplify setup.
func defaultEmbedDir(field *string, dirname string) error {
	if *field == "" {
		path, err := embedDir(dirname)
		if err != nil {
			return err
		}
		*field = path
	}
	return nil
}

// embedDir computes the full path for embedding data,
// centralizing layout logic.
func embedDir(dirname string) (string, error) {
	homeDir, err := user.HomeDir()
	if err != nil {
		return "", err
	}
	target := filepath.Join(homeDir, appName, dirname)
	return target, nil
}
