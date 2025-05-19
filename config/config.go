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
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/alnah/fla/locale"
)

const (
	// Application
	AppName string = "fla"

	// Env variables
	Env            string = "FLA_ENV"
	EnvLang        string = "FLA_LANG"
	EnvInput       string = "FLA_INPUT"
	EnvOutput      string = "FLA_OUTPUT"
	EnvTimeoutTTS  string = "FLA_TIMEOUT_TTS"
	EnvTimeoutChat string = "FLA_TIMEOUT_CHAT"
	EnvTimeoutSTT  string = "FLA_TIMEOUT_STT"

	// Permissions
	PermUserRW  os.FileMode = 0o600
	PermUserRWX os.FileMode = 0o700
)

// Timeout groups the various AI-related deadlines so they can be managed and validated as a unit.
type Timeout struct {
	Chat time.Duration `json:"chat_completion"`
	TTS  time.Duration `json:"tts_audio"`
	STT  time.Duration `json:"stt_transcript"`
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
	log      *slog.Logger
	mu       sync.RWMutex
	dirpath  string
	filename string
}

type ConfigPath struct {
	DirPath  string
	FileName string
}

// Path returns the fully resolved JSON-config filepath for the current user,
// honoring XDG on Unix and APPDATA on Windows. Fallback to ~/.config/fla.
// Returns the dirpath, and the filename.
func Path() (ConfigPath, error) {
	// unix
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return ConfigPath{DirPath: filepath.Join(xdg, AppName), FileName: filename()}, nil
	}
	// windows
	if runtime.GOOS == "windows" {
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return ConfigPath{DirPath: filepath.Join(appdata, appdata), FileName: filename()}, nil
		}
	}
	// fallback to ~/.config
	home, err := os.UserHomeDir()
	if err != nil {
		return ConfigPath{}, fmt.Errorf("finding home: %w", err)
	}
	return ConfigPath{DirPath: filepath.Join(home, ".config", AppName), FileName: filename()}, nil
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
func Load(log *slog.Logger, dirpath, filename string) (*Config, error) {
	var (
		cfg *Config
		err error
	)

	if err = ensureConfigDir(filepath.Dir(dirpath)); err != nil {
		log.Info("creating config directory", "config_directory", filepath.Dir(dirpath), "error", err.Error())
		return nil, fmt.Errorf("ensure config dir: %w", err)
	}

	byt, err := readFile(log, dirpath, filename)
	if err != nil {
		return nil, err
	}

	cfg, err = parseJSON(log, byt, dirpath, filename)
	if err != nil {
		return nil, err
	}
	if err := cfg.applyDefaults(log, dirpath, filename); err != nil {
		return nil, err
	}
	if err := cfg.applyEnvOverrides(); err != nil {
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
	flag.DurationVar(&c.AI.Timeout.Chat,
		"timeout-chat", c.AI.Timeout.Chat,
		"chat timeout (e.g. 30s, 1m)",
	)
	flag.DurationVar(&c.AI.Timeout.TTS,
		"timeout-tts", c.AI.Timeout.TTS,
		"text-to-speech timeout (e.g. 5m)",
	)
	flag.DurationVar(&c.AI.Timeout.STT,
		"timeout-stt", c.AI.Timeout.STT,
		"speech-to-text timeout (e.g. 30s)",
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
		"chat": c.AI.Timeout.Chat,
		"tts":  c.AI.Timeout.TTS,
		"stt":  c.AI.Timeout.STT,
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
			c.log.Info("creating dir", "name", name, "path", *p)
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

	tmp := filepath.Join(c.dirpath, c.filename, ".tmp")
	if err := os.WriteFile(tmp, data, PermUserRW); err != nil {
		return fmt.Errorf("writing temp config: %w", err)
	}

	if err := os.Rename(tmp, filepath.Join(c.dirpath, c.filename)); err != nil {
		return fmt.Errorf("replacing config file: %w", err)
	}

	return nil
}

func ensureConfigDir(dir string) error {
	return os.MkdirAll(dir, PermUserRWX)
}

// readFileSecure fixes potential file inclusion via variable for gosec G304 (CWE-22).
func readFileSecure(dirpath, filename string) ([]byte, error) {
	// 1 — filename must not contain any separators
	if filename != filepath.Base(filename) {
		return nil, fmt.Errorf("invalid filename %q", filename)
	}

	base := filepath.Clean(dirpath)                         // canonical base
	target := filepath.Clean(filepath.Join(base, filename)) // canonical file

	// 2 — confinement: final path must stay under base dir
	if !strings.HasPrefix(target, base+string(os.PathSeparator)) {
		return nil, fmt.Errorf("path %q escapes base dir %q", target, base)
	}

	// 3 — safe read
	return os.ReadFile(target) // gosec: OK after the two guarantees above
}

// readFile loads the file at path under its directory, creating a default "{}" if missing.
// It wraps all reads in securePath to mitigate gosec G304 (CWE-22).
func readFile(log *slog.Logger, dirpath, filename string) ([]byte, error) {
	data, err := readFileSecure(dirpath, filename)
	if err == nil {
		return data, nil
	}

	// if file not found, create a default "{}" and read again
	if errors.Is(err, os.ErrNotExist) {
		log.Info("config not found, creating default", "filepath", filepath.Join(dirpath, filename))

		// ensure directory exists
		if mkErr := os.MkdirAll(filepath.Dir(dirpath), PermUserRWX); mkErr != nil {
			return nil, fmt.Errorf("making config dir: %w", mkErr)
		}
		// write default JSON
		if wErr := os.WriteFile(filepath.Join(dirpath, filename), []byte("{}"), PermUserRW); wErr != nil {
			return nil, fmt.Errorf("creating default config: %w", wErr)
		}

		return readFileSecure(dirpath, filename)
	}

	// any other error
	log.Error("failed reading config", "filepath", filepath.Join(dirpath, filename), "error", err)
	return nil, fmt.Errorf("reading config %q: %w", filepath.Join(dirpath, filename), err)
}

func parseJSON(log *slog.Logger, byt []byte, dirpath, filename string) (*Config, error) {
	var cfg Config
	dec := json.NewDecoder(bytes.NewReader(byt))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decoding json in %s: %w", filepath.Join(dirpath, filename), err)
	}

	cfg.log = log
	cfg.dirpath = dirpath
	cfg.filename = filename
	return &cfg, nil
}

func appDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding user home dir: %w", err)
	}
	return filepath.Join(home, AppName), nil
}

func (c *Config) applyDefaults(log *slog.Logger, dirpath, filename string) error {
	c.log = log
	c.dirpath = dirpath
	c.filename = filename

	appDir, err := appDir()
	if err != nil {
		return err
	}

	// apply default language
	if c.Language == "" {
		c.Language = locale.LangFrFR
	}

	// apply default input/ouput dirs
	if c.Input == "" {
		c.Input = filepath.Join(appDir, "input")
	}
	if c.Output == "" {
		c.Output = filepath.Join(appDir, "output")
	}

	// apply default ai timeouts
	if c.AI.Timeout.Chat == 0 {
		c.AI.Timeout.Chat = 30 * time.Second
	}
	if c.AI.Timeout.TTS == 0 {
		c.AI.Timeout.TTS = 5 * time.Minute
	}
	if c.AI.Timeout.STT == 0 {
		c.AI.Timeout.STT = 30 * time.Second
	}

	return nil
}

func (c *Config) applyEnvOverrides() error {
	// language
	if v, ok := os.LookupEnv(EnvLang); ok {
		var l locale.Lang
		if err := l.Set(v); err != nil {
			return fmt.Errorf("parsing %s=%q: %w", EnvLang, v, err)
		}
		c.Language = l
	}

	// input/output dirs
	if v, ok := os.LookupEnv(EnvInput); ok {
		c.Input = v
	}
	if v, ok := os.LookupEnv(EnvOutput); ok {
		c.Output = v
	}

	// ai timeouts
	if v, ok := os.LookupEnv(EnvTimeoutChat); ok {
		d, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("parsing FLA_TIMEOUT_CHAT=%q: %w", v, err)
		}
		c.AI.Timeout.Chat = d
	}
	if v, ok := os.LookupEnv(EnvTimeoutTTS); ok {
		d, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("parsing FLA_TIMEOUT_TTS=%q: %w", v, err)
		}
		c.AI.Timeout.TTS = d
	}
	if v, ok := os.LookupEnv(EnvTimeoutSTT); ok {
		d, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("parsing FLA_TIMEOUT_STT=%q: %w", v, err)
		}
		c.AI.Timeout.STT = d
	}

	return nil
}
