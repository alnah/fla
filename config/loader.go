package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"

	"github.com/alnah/fla/filesystem"
	fsys "github.com/alnah/fla/filesystem"
)

const appName string = "fla" // means "foreign language acquisition"

type Handler interface {
	Load() (*Config, error)
	Save() error
}

type loader struct {
	log      *slog.Logger
	env      Env
	userDirs UserDirs
	fs       struct {
		config fsys.FileSystem
		temp   fsys.FileSystem
		home   fsys.FileSystem
	}
	filename string
}

type option func(*loader)

func WithEnv(e Env) func(*loader)                  { return func(l *loader) { l.env = e } }
func WithUserDirs(u UserDirs) func(*loader)        { return func(l *loader) { l.userDirs = u } }
func WithConfigFS(f fsys.FileSystem) func(*loader) { return func(l *loader) { l.fs.config = f } }
func WithTempFS(f fsys.FileSystem) func(*loader)   { return func(l *loader) { l.fs.temp = f } }
func WithHomeFS(f fsys.FileSystem) func(*loader)   { return func(l *loader) { l.fs.home = f } }
func WithLogger(s *slog.Logger) func(*loader)      { return func(l *loader) { l.log = s } }

func NewLoader(opts ...option) *loader {
	l := &loader{
		env:      env,
		userDirs: &user,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

func (l *loader) Load() (*Config, error) {
	var err error

	// display info
	if l.log != nil {
		l.log.Info("loading configuration", "env", l.env.Type())
	}

	// default config FS
	if l.fs.config == nil {
		if l.fs.config, err = defaultConfigFS(); err != nil {
			return nil, err
		}
	}
	// ensure config dir
	if err = l.fs.config.MkdirAll(".", fsys.PermUserRWX); err != nil {
		return nil, err
	}

	// default temp FS
	if l.fs.temp == nil {
		if l.fs.temp, err = defaultTempFS(); err != nil {
			return nil, err
		}
	}
	// ensure temp dir
	if err = l.fs.temp.MkdirAll(".", fsys.PermUserRWX); err != nil {
		return nil, err
	}

	// default home FS
	if l.fs.home == nil {
		if l.fs.home, err = defaultHomeFS(); err != nil {
			return nil, err
		}
	}
	// no need to ensure home dir, it should exist in the os

	// default confif fime depending on environment type
	l.filename, err = defaultConfigFilename()
	if err != nil {
		return nil, err
	}

	// reading/initializing config file
	byt, err := l.fs.config.ReadFile(l.filename)
	if errors.Is(err, fs.ErrNotExist) {
		if err := l.fs.config.WriteFile(l.filename, []byte("{}"), fsys.PermUserRW); err != nil {
			return nil, err
		}
		byt, err = l.fs.config.ReadFile(l.filename)
	}
	if err != nil {
		return nil, err
	}

	// decoding config
	var cfg Config
	dec := json.NewDecoder(bytes.NewReader(byt))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}
	if err := cfg.defaults(); err != nil {
		return nil, err
	}
	if err := cfg.Lang.Validate(); err != nil {
		return nil, err
	}

	// env overrides
	if err = l.envOverride(&cfg); err != nil {
		return nil, err
	}

	// initializing home dirs
	for _, dir := range []string{
		cfg.Dir.Input,
		cfg.Dir.Staging,
		cfg.Dir.Output.Student,
		cfg.Dir.Output.Teacher,
		cfg.Dir.Output.Lessons,
	} {
		if err := l.ensureSpecDir(l.fs.home, dir); err != nil {
			return nil, err
		}
	}

	return &cfg, nil
}

func (l *loader) UpdateConfigFile(cfg *Config) error {
	byt, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}
	if err = l.fs.config.WriteFile(l.filename, byt, filesystem.PermUserRW); err != nil {
		return err
	}
	return nil
}

func (l *loader) ensureSpecDir(f fsys.FileSystem, path string) error {
	info, err := f.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return f.MkdirAll(path, fsys.PermUserRWX)
		}
		return fmt.Errorf("ensure spec directory path %s: stat failed: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("ensure spec directory failed: %s exists, but is not a directory", path)
	}
	return nil
}
