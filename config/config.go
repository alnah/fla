// Package config loads and validates fla’s settings.
// Typical initialization pipeline:
//  1. log, err := logger.New()
//  2. path, err := config.Path()
//  3. cfg, err := config.Load(log, path)
//  4. cfg.BindFlags()
//  5. flag.Parse()
//  6. err = cfg.Validate()
//  7. err = cfg.EnsureIODirs()
package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/alnah/fla/locale"
	"github.com/alnah/fla/logger"
)

const (
	// Application
	AppName string = "fla"

	// Env variables
	Env                     string = "FLA_ENV"
	EnvLang                 string = "FLA_LANG"
	EnvInput                string = "FLA_INPUT"
	EnvOutput               string = "FLA_OUTPUT"
	EnvTimeoutAudio         string = "FLA_TIMEOUT_AUDIO"
	EnvTimeoutCompletion    string = "FLA_TIMEOUT_COMPLETION"
	EnvTimeoutTranscription string = "FLA_TIMEOUT_TRANSCRIPTION"

	// Permissions
	PermUserRW  os.FileMode = 0o600
	PermUserRWX os.FileMode = 0o700
)

// Timeout groups the various AI-related deadlines so they can be managed and validated as a unit.
type Timeout struct {
	Completion    time.Duration `json:"completion"`
	Audio         time.Duration `json:"audio"`
	Transcription time.Duration `json:"transcription"`
}

// AI holds configuration for any AI client integrations.
type AI struct {
	Timeout Timeout `json:"timeout"`
}

// Config represents all mutable settings for the application.
// After Initialize and Finalize, it is guaranteed to be valid and ready for use.
type Config struct {
	Language locale.Lang `json:"language"`
	AI       AI          `json:"ai"`
	Input    string      `json:"input"`
	Output   string      `json:"output"`
	Log      *logger.Logger
	mu       sync.RWMutex
	path     string
}

// Path returns the fully resolved JSON-config filepath for the current user,
// honoring XDG on Unix and APPDATA on Windows. Fallback to ~/.config/fla.
func Path() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, AppName, filename()), nil
	}
	// windows
	if runtime.GOOS == "windows" {
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return filepath.Join(appdata, AppName, filename()), nil
		}
	}
	// fallback to ~/.config
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding home: %w", err)
	}
	return filepath.Join(home, ".config", AppName, filename()), nil
}

func env() string {
	switch e := os.Getenv(Env); e {
	case "dev", "staging":
		return e
	default:
		return "prod"
	}
}

func filename() string {
	switch env() {
	case "dev":
		return "config.dev.json"
	case "staging":
		return "config.staging.json"
	default:
		return "config.json"
	}
}

// Load loads the JSON config (file → defaults → env).
func Load(log *logger.Logger, path string) (*Config, error) {
	var (
		cfg *Config
		err error
	)

	if err = ensureConfigDir(filepath.Dir(path)); err != nil {
		log.Info("creating config directory", "config_directory", filepath.Dir(path), "error", err.Error())
		return nil, fmt.Errorf("ensure config dir: %w", err)
	}

	byt, err := readFile(log, path)
	if err != nil {
		return nil, err
	}

	cfg, err = parseJSON(log, byt, path)
	if err != nil {
		return nil, err
	}
	if err := applyDefaults(log, path, cfg); err != nil {
		return nil, err
	}
	if err := applyEnvOverrides(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// BindFlags registers command-line flags for overriding defaults.
func (c *Config) BindFlags() {
	// language
	flag.Var(&c.Language,
		"language",
		"UI language (fr|pt)",
	)

	// input/output dirs
	flag.StringVar(&c.Input,
		"input", c.Input,
		"path to input directory",
	)
	flag.StringVar(&c.Output,
		"output", c.Output,
		"path to output directory",
	)

	// ai timeouts
	flag.DurationVar(&c.AI.Timeout.Completion,
		"timeout-completion", c.AI.Timeout.Completion,
		"AI completion timeout (e.g. 30s, 1m)",
	)
	flag.DurationVar(&c.AI.Timeout.Audio,
		"timeout-audio", c.AI.Timeout.Audio,
		"AI audio timeout (e.g. 5m)",
	)
	flag.DurationVar(&c.AI.Timeout.Transcription,
		"timeout-transcription", c.AI.Timeout.Transcription,
		"AI transcription timeout (e.g. 30s)",
	)
}

// Validate checks that all Config fields are sane.
func (c *Config) Validate() error {
	// language
	if err := c.Language.Validate(); err != nil {
		return err
	}

	// input/output dirs
	dirs := map[string]string{
		"input":  c.Input,
		"output": c.Output,
	}

	for name, dir := range dirs {
		if dirname := filepath.Base(dir); len(dirname) > 255 {
			return fmt.Errorf("invalid dirname for %s: %q: exceeds 255 characters", name, dirname)
		}
	}

	// ai client timeouts
	timeouts := map[string]time.Duration{
		"completion":    c.AI.Timeout.Completion,
		"audio":         c.AI.Timeout.Audio,
		"transcription": c.AI.Timeout.Transcription,
	}
	for name, t := range timeouts {
		if t <= 0 {
			return fmt.Errorf("invalid %s timeout: %s: must be non-negative and positive", name, t)
		}
		if t > 30*time.Minute {
			return fmt.Errorf("invalid %s timeout: %s: must be less than 30 minutes", name, t)
		}
	}

	return nil
}

// EnsureIODirs creates and verifies input/output directories.
func (c *Config) EnsureIODirs() error {
	dirs := map[string]*string{
		"input":  &c.Input,
		"output": &c.Output,
	}
	for name, p := range dirs {
		if err := os.MkdirAll(*p, PermUserRWX); err != nil {
			c.Log.Info("creating dir", "name", name, "path", *p)
		}
		// verify directory is writable
		testFile := filepath.Join(*p, ".permtest")
		if err := os.WriteFile(testFile, []byte{}, PermUserRW); err != nil {
			return fmt.Errorf("writability test failed for %q: %w", *p, err)
		} else {
			_ = os.Remove(testFile)
		}
	}
	return nil
}

// Save updates the JSON file holding the config, using a temp for writing,
// and replacing the original file. Thread-safe.
func (c *Config) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	tmp := c.path + ".tmp"
	if err := os.WriteFile(tmp, data, PermUserRW); err != nil {
		return fmt.Errorf("writing temp config: %w", err)
	}

	if err := os.Rename(tmp, c.path); err != nil {
		return fmt.Errorf("replacing config file: %w", err)
	}

	return nil
}

func ensureConfigDir(dir string) error {
	return os.MkdirAll(dir, PermUserRWX)
}

// readFileSecure fixes potential file inclusion via variable for gosec G304 (CWE-22).
func readFileSecure(path string) ([]byte, error) {
	// define base directory
	baseDir := filepath.Dir(path)

	// clean and join
	cleanPath := filepath.Clean(path)
	fullPath := filepath.Join(baseDir, cleanPath)

	// derify no traversal outside baseDir
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("absolute base directory: %w", err)
	}
	absFile, err := filepath.Abs(fullPath)
	if err != nil {
		return nil, fmt.Errorf("absolute full path: %w", err)
	}
	rel, err := filepath.Rel(absBase, absFile)
	if err != nil || strings.HasPrefix(rel, "..") {
		return nil, fmt.Errorf("invalid path: %s", fullPath)
	}

	// eval symlinks on the directory
	dir := filepath.Dir(path)
	realDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return nil, fmt.Errorf("resolving config directory %q: %w", dir, err)
	}

	// sanitize the path
	file := filepath.Base(path)

	// read the file under the real directory

	return os.ReadFile(filepath.Join(realDir, file))
}

// readFile loads the file at path under its directory, creating a default "{}" if missing.
// It wraps all reads in securePath to mitigate gosec G304 (CWE-22).
func readFile(log *logger.Logger, path string) ([]byte, error) {
	data, err := readFileSecure(path)
	if err == nil {
		return data, nil
	}

	// if file not found, create a default "{}" and read again
	if errors.Is(err, os.ErrNotExist) {
		log.Info("config not found, creating default", "path", path)

		// ensure directory exists
		if mkErr := os.MkdirAll(filepath.Dir(path), PermUserRWX); mkErr != nil {
			return nil, fmt.Errorf("making config dir: %w", mkErr)
		}
		// write default JSON
		if wErr := os.WriteFile(path, []byte("{}"), PermUserRW); wErr != nil {
			return nil, fmt.Errorf("creating default config: %w", wErr)
		}

		return readFileSecure(path)
	}

	// any other error
	log.Error("failed reading config", "path", path, "error", err)
	return nil, fmt.Errorf("reading config %q: %w", path, err)
}

func parseJSON(log *logger.Logger, byt []byte, path string) (*Config, error) {
	var cfg Config
	dec := json.NewDecoder(bytes.NewReader(byt))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decoding json in %s: %w", path, err)
	}

	cfg.Log = log
	cfg.path = path
	return &cfg, nil
}

func appDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding user home dir: %w", err)
	}
	return filepath.Join(home, AppName), nil
}

func applyDefaults(l *logger.Logger, p string, cfg *Config) error {
	cfg.Log = l
	cfg.path = p

	appDir, err := appDir()
	if err != nil {
		return err
	}

	// apply default language
	if cfg.Language == "" {
		cfg.Language = locale.FR
	}

	// apply default input/ouput dirs
	if cfg.Input == "" {
		cfg.Input = filepath.Join(appDir, "input")
	}
	if cfg.Output == "" {
		cfg.Output = filepath.Join(appDir, "output")
	}

	// apply default ai timeouts
	if cfg.AI.Timeout.Completion == 0 {
		cfg.AI.Timeout.Completion = 30 * time.Second
	}
	if cfg.AI.Timeout.Audio == 0 {
		cfg.AI.Timeout.Audio = 5 * time.Minute
	}
	if cfg.AI.Timeout.Transcription == 0 {
		cfg.AI.Timeout.Transcription = 30 * time.Second
	}

	return nil
}

func applyEnvOverrides(cfg *Config) error {
	// language
	if v, ok := os.LookupEnv(EnvLang); ok {
		var l locale.Lang
		if err := l.Set(v); err != nil {
			return fmt.Errorf("parsing %s=%q: %w", EnvLang, v, err)
		}
		cfg.Language = l
	}

	// input/output dirs
	if v, ok := os.LookupEnv(EnvInput); ok {
		cfg.Input = v
	}
	if v, ok := os.LookupEnv(EnvOutput); ok {
		cfg.Output = v
	}

	// ai timeouts
	if v, ok := os.LookupEnv(EnvTimeoutCompletion); ok {
		d, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("parsing FLA_TIMEOUT_COMPLETION=%q: %w", v, err)
		}
		cfg.AI.Timeout.Completion = d
	}
	if v, ok := os.LookupEnv(EnvTimeoutAudio); ok {
		d, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("parsing FLA_TIMEOUT_AUDIO=%q: %w", v, err)
		}
		cfg.AI.Timeout.Audio = d
	}
	if v, ok := os.LookupEnv(EnvTimeoutTranscription); ok {
		d, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("parsing FLA_TIMEOUT_TRANSCRIPTION=%q: %w", v, err)
		}
		cfg.AI.Timeout.Transcription = d
	}

	return nil
}
