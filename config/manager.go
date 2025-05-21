package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"

	fsys "github.com/alnah/fla/filesystem"
	"github.com/alnah/fla/logger"
)

// TODO: testing against config
// TODO: checking if prompt registry will not benefit of my custom filesystem
// TODO: now prompt registry does integration test, it should do unit tests with my own filesystem
const appName = "fla" // short for Foreign Language Acquisition

// ConfigManager defines how to load and persist application settings.
type ConfigManager interface {
	Load() (*manager, error) // read, validate, and apply defaults+env
	Save() error             // save updated settings back to disk
}

type manager struct {
	*Config
	log      logger.Logger // optional logger for visibility
	env      Env           // source of environment overrides
	userDirs UserDirs      // user directory locations
	fs       struct {
		config fsys.FileSystem // where config.json lives
		home   fsys.FileSystem // to embed lesson dirs
	}
	filename string // name of the JSON file in config FS
}

type option func(*manager)

// WithEnv lets you swap in a fake or test environment.
func WithEnv(e Env) func(*manager) { return func(m *manager) { m.env = e } }

// WithUserDirs lets you override where to find user dirs.
func WithUserDirs(u UserDirs) func(*manager) { return func(m *manager) { m.userDirs = u } }

// WithConfigFS assigns a custom filesystem for config storage.
func WithConfigFS(f fsys.FileSystem) func(*manager) {
	return func(m *manager) { m.fs.config = f }
}

// WithHomeFS assigns a filesystem for embedding lesson content.
func WithHomeFS(f fsys.FileSystem) func(*manager) {
	return func(m *manager) { m.fs.home = f }
}

// WithLogger provides visibility into loading steps.
func WithLogger(s logger.Logger) func(*manager) { return func(m *manager) { m.log = s } }

// New constructs a manager with optional overrides,
// preparing to Load or Update configuration.
func New(opts ...option) *manager {
	l := &manager{
		env:      env,
		userDirs: &user,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// Load reads the JSON config file (creating it if missing),
// applies defaults, validates, overrides with env vars,
// and ensures all required directories exist.
func (m *manager) Load() (*manager, error) {
	var err error

	// logging
	if m.log != nil {
		m.log = logger.Default()
	}
	m.log.Info("loading configuration", "env", m.env.Type())

	// prepare config FS
	if m.fs.config == nil {
		if m.fs.config, err = defaultConfigFS(); err != nil {
			return nil, NewConfigError("loading filesystem", err)
		}
	}
	if err = m.fs.config.MkdirAll(".", fsys.PermUserRWX); err != nil {
		return nil, NewConfigError("creating filesystem directory", err)
	}

	// prepare home FS
	if m.fs.home == nil {
		if m.fs.home, err = defaultHomeFS(); err != nil {
			return nil, NewConfigError("loading home filesystem", err)
		}
	}

	// determine filename
	m.filename, err = defaultConfigFilename()
	if err != nil {
		return nil, NewConfigError("building filename", err)
	}

	// read or initialize JSON
	byt, err := m.fs.config.ReadFile(m.filename)
	if errors.Is(err, fs.ErrNotExist) {
		if err := m.fs.config.WriteFile(m.filename, []byte("{}"), fsys.PermUserRW); err != nil {
			return nil, NewConfigError("creating json file", err)
		}
		byt, err = m.fs.config.ReadFile(m.filename)
	}
	if err != nil {
		return nil, NewConfigError("reading json file", err)
	}

	// decode and validate schema
	var cfg Config
	dec := json.NewDecoder(bytes.NewReader(byt))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return nil, NewConfigError("validating config schema", fmt.Errorf("%w; remove unknown fields", err))
	}
	if err := cfg.defaults(); err != nil {
		return nil, NewConfigError("applying defaults", err)
	}
	if err := cfg.Lang.Validate(); err != nil {
		return nil, NewConfigError("validating language", err)
	}

	// apply environment overrides
	if err = m.envOverride(&cfg); err != nil {
		return nil, NewConfigError("overriding environment variables", err)
	}

	// ensure each embed directory exists on the home FS
	for _, dir := range []string{
		cfg.Dir.Input,
		cfg.Dir.Staging,
		cfg.Dir.Output.Student,
		cfg.Dir.Output.Teacher,
		cfg.Dir.Output.Lessons,
	} {
		if err := m.ensureEmbedDir(m.fs.home, dir); err != nil {
			return nil, NewConfigError("creating embed directory", err)
		}
	}

	m.Config = &cfg
	return m, nil
}

// Save rewrites the JSON config file to reflect any changes,
// keeping file structure readable with indentation.
func (m *manager) Save() error {
	byt, err := json.MarshalIndent(m.Config, "", "  ")
	if err != nil {
		return NewConfigError("marshaling configuration for update", err)
	}
	if err = m.fs.config.WriteFile(m.filename, byt, fsys.PermUserRW); err != nil {
		return NewConfigError("writing configuration file", err)
	}
	return nil
}

// ensureEmbedDir makes sure the given path exists and is a directory,
// so embedded lesson directories are ready when needed.
func (m *manager) ensureEmbedDir(f fsys.FileSystem, path string) error {
	info, err := f.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return f.MkdirAll(path, fsys.PermUserRWX)
		}
		return fmt.Errorf("ensuring embed directory %s: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("embed path %s exists but is not a directory", path)
	}
	return nil
}
