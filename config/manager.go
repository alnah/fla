package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	fsys "github.com/alnah/fla/filesystem"
	"github.com/alnah/fla/logger"
)

const appName = "fla" // short for Foreign Language Acquisition

// ConfigManager defines how to load and persist application settings.
type ConfigManager interface {
	Load() (*Manager, error) // read, validate, and apply defaults+env
	Save() error             // save updated settings back to disk
}

type Manager struct {
	// in-memory config
	Config *Config

	// config filepath
	Filepath string

	// file systems
	FS struct {
		Config fsys.FileSystem // config file
		Home   fsys.FileSystem // embed directories
	}

	// logger
	log logger.Logger // logger

	// testing purpose
	env      Env      // for testing purpose
	userDirs UserDirs // for testing purpose

	// internal
	filename   string // name of the JSON file
	fieldsMeta []fieldMeta
}

type option func(*Manager)

// WithEnv lets you swap in a fake or test environment.
func WithEnv(e Env) func(*Manager) { return func(m *Manager) { m.env = e } }

// WithUserDirs lets you override where to find user dirs.
func WithUserDirs(u UserDirs) func(*Manager) { return func(m *Manager) { m.userDirs = u } }

// WithConfigFS assigns a custom filesystem for config storage.
func WithConfigFS(f fsys.FileSystem) func(*Manager) {
	return func(m *Manager) { m.FS.Config = f }
}

// WithHomeFS assigns a filesystem for embedding lesson content.
func WithHomeFS(f fsys.FileSystem) func(*Manager) {
	return func(m *Manager) { m.FS.Home = f }
}

// WithLogger provides visibility into loading steps.
func WithLogger(s logger.Logger) func(*Manager) { return func(m *Manager) { m.log = s } }

// NewConfigManager constructs a manager with optional overrides,
// preparing to Load or Update configuration.
func NewConfigManager(opts ...option) *Manager {
	l := &Manager{
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
func (m *Manager) Load() (*Manager, error) {
	var err error

	// logging
	if m.log != nil {
		m.log = logger.Default()
	}
	m.log.Info("loading configuration", "env", m.env.Type())

	// prepare config FS
	if m.FS.Config == nil {
		if m.FS.Config, err = defaultConfigFS(); err != nil {
			return nil, newConfigError("loading filesystem", err)
		}
	}
	if err = m.FS.Config.MkdirAll(".", fsys.PermUserRWX); err != nil {
		return nil, newConfigError("creating filesystem directory", err)
	}

	// prepare home FS
	if m.FS.Home == nil {
		if m.FS.Home, err = defaultHomeFS(); err != nil {
			return nil, newConfigError("loading home filesystem", err)
		}
	}

	// determine filename
	m.filename, err = defaultConfigFilename()
	if err != nil {
		return nil, newConfigError("building filename", err)
	}
	// determine filepath
	m.Filepath = filepath.Join(m.FS.Config.Root(), m.filename)

	// read or initialize JSON
	byt, err := m.FS.Config.ReadFile(m.filename)
	if errors.Is(err, fs.ErrNotExist) {
		if err := m.FS.Config.WriteFile(m.filename, []byte("{}"), fsys.PermUserRW); err != nil {
			return nil, newConfigError("creating json file", err)
		}
		byt, err = m.FS.Config.ReadFile(m.filename)
	}
	if err != nil {
		return nil, newConfigError("reading json file", err)
	}

	// decode and validate schema
	var cfg Config
	dec := json.NewDecoder(bytes.NewReader(byt))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return nil, newConfigError("validating config schema", fmt.Errorf("%w; remove unknown fields", err))
	}
	if err := cfg.defaults(); err != nil {
		return nil, newConfigError("applying defaults", err)
	}
	if err := cfg.Lang.Validate(); err != nil {
		return nil, newConfigError("validating language", err)
	}

	// builds field metas
	m.fieldsMeta = fields(&cfg)

	// env vars overrides
	if err := m.OverrideEnvBak(&cfg); err != nil {
		return nil, newConfigError("overriding environment variables", err)
	}

	// ensure each embed directory exists on the home FS
	for _, dir := range []string{
		cfg.Dir.Input,
		cfg.Dir.Staging,
		cfg.Dir.Output.Student,
		cfg.Dir.Output.Teacher,
		cfg.Dir.Output.Lessons,
	} {
		if err := m.ensureEmbedDir(m.FS.Home, dir); err != nil {
			return nil, newConfigError("creating embed directory", err)
		}
	}

	m.Config = &cfg
	return m, nil
}

// Save rewrites the JSON config file to reflect any changes,
// keeping file structure readable with indentation.
func (m *Manager) Save() error {
	byt, err := json.MarshalIndent(m.Config, "", "  ")
	if err != nil {
		return newConfigError("marshaling configuration for update", err)
	}
	if err = m.FS.Config.WriteFile(m.filename, byt, fsys.PermUserRW); err != nil {
		return newConfigError("writing configuration file", err)
	}
	return nil
}

// ensureEmbedDir makes sure the given path exists and is a directory,
// so embedded lesson directories are ready when needed.
func (m *Manager) ensureEmbedDir(f fsys.FileSystem, path string) error {
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
