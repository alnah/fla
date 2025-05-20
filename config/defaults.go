package config

import (
	"log/slog"
	"path/filepath"
	"time"

	"github.com/alnah/fla/filesystem"
	"github.com/alnah/fla/locale"
)

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

	// timeouts guard external dependencies
	defaultVal(&c.Timeout.Chat, Duration(30*time.Second))
	defaultVal(&c.Timeout.TTS, Duration(5*time.Minute))
	defaultVal(&c.Timeout.STT, Duration(30*time.Second))

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

// defaultTempFS returns a FileSystem for transient data to avoid clutter.
func defaultTempFS() (filesystem.FileSystem, error) {
	userCacheDir, err := user.CacheDir()
	if err != nil {
		return nil, err
	}
	tempDir := filepath.Join(userCacheDir, appName, defaultTempDir())
	return filesystem.New(tempDir), nil
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

// defaultTempDir picks a temp folder name reflecting the environment,
// preventing accidental data mix-up.
func defaultTempDir() string {
	switch env.Type() {
	case "dev":
		return "temp_dev"
	case "test":
		return "temp_test"
	default:
		return "temp"
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
