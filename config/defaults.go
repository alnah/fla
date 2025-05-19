package config

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/alnah/fla/filesystem"
	"github.com/alnah/fla/locale"
)

func (c *Config) defaults() error {
	// log level
	defaultPtr(&c.LogLevel, slog.LevelError)

	// language
	defaultPtr(&c.Lang, locale.LangFrFR)

	// filenames
	defaultVal(&c.Filename.Date, true)
	defaultVal(&c.Filename.Level, true)
	defaultVal(&c.Filename.Lesson, "Leçon")
	defaultVal(&c.Filename.Preparation, "Préparation")
	defaultVal(&c.Filename.Plan, "Plan")
	defaultVal(&c.Filename.Reading, "Texte")
	defaultVal(&c.Filename.Listening, "Audio")
	defaultVal(&c.Filename.Watching, "Vidéo")
	defaultVal(&c.Filename.Correction, "Correction")

	// working directories
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

	// timeouts
	defaultVal(&c.Timeout.Chat, 30*time.Second)
	defaultVal(&c.Timeout.TTS, 5*time.Minute)
	defaultVal(&c.Timeout.STT, 30*time.Second)

	return nil
}

func defaultConfigFS() (filesystem.FileSystem, error) {
	userConfigDir, err := user.ConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to create the config file system: %w", err)
	}
	configDir := filepath.Join(userConfigDir, appName)
	return filesystem.New(configDir), nil
}

func defaultTempFS() (filesystem.FileSystem, error) {
	userCacheDir, err := user.CacheDir()
	if err != nil {
		return nil, fmt.Errorf("failed to create the temp filesystem: %w", err)
	}
	tempDir := filepath.Join(userCacheDir, appName, defaultTempDir())
	return filesystem.New(tempDir), nil
}

func defaultHomeFS() (filesystem.FileSystem, error) {
	_, err := user.HomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to create the home filesystem: %w", err)
	}
	return filesystem.New("/"), nil
}

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

func defaultPtr[T any](field **T, def T) {
	if *field == nil {
		*field = &def
	}
}

func defaultVal[T comparable](field *T, def T) {
	var zero T
	if *field == zero {
		*field = def
	}
}

func defaultEmbedDir(field *string, dirname string) error {
	if *field == "" {
		path, err := embedDir(dirname)
		if err != nil {
			return fmt.Errorf("failed to create embed cli directory: %w", err)
		}
		*field = path
	}
	return nil
}

func embedDir(dirname string) (string, error) {
	homeDir, err := user.HomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to find user home directory: %w", err)
	}

	target := filepath.Join(homeDir, appName, dirname)
	return target, nil
}
