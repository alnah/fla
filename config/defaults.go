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
	if err := defaultDir(&c.Dir.Input, "input"); err != nil {
		return err
	}
	if err := defaultDir(&c.Dir.Staging, "staging"); err != nil {
		return err
	}
	if err := defaultDir(&c.Dir.Output.Student, "student"); err != nil {
		return err
	}
	if err := defaultDir(&c.Dir.Output.Teacher, "teacher"); err != nil {
		return err
	}
	if err := defaultDir(&c.Dir.Output.Lessons, "lessons"); err != nil {
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
		return nil, err
	}
	configDir := filepath.Join(userConfigDir, appName)
	return filesystem.New(configDir), nil
}

func defaultTempFS() (filesystem.FileSystem, error) {
	userCacheDir, err := user.CacheDir()
	if err != nil {
		return nil, err
	}
	envTempDir, err := defaultTempDir()
	if err != nil {
		return nil, err
	}
	tempDir := filepath.Join(userCacheDir, appName, envTempDir)
	return filesystem.New(tempDir), nil
}

func defaultHomeFS() (filesystem.FileSystem, error) {
	_, err := user.HomeDir()
	if err != nil {
		return nil, err
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

func defaultTempDir() (string, error) {
	switch env.Type() {
	case "dev":
		return "temp_dev", nil
	case "test":
		return "temp_test", nil
	default:
		return "temp", nil
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

func defaultDir(field *string, dirname string) error {
	if *field == "" {
		path, err := specDir(dirname)
		if err != nil {
			return err
		}
		*field = path
	}
	return nil
}

func specDir(dirname string) (string, error) {
	homeDir, err := user.HomeDir()
	if err != nil {
		return "", fmt.Errorf("spec directory path: failed to find user home directory: %w", err)
	}

	target := filepath.Join(homeDir, appName, dirname)
	return target, nil
}
