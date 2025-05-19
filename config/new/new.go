package new

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/alnah/fla/locale"
	"github.com/alnah/fla/pathutil"
)

// TODO: Decouple FS with fs.FS interface for testing.
// TODO: Atomic file creation and permission.
// TODO: Store the timeouts, and expose methods returning context.Context with :deadlines.
// TODO: Slog errors:

const (
	appName string = "fla"      // means "foreign language acquisition"
	envType string = "ENV_TYPE" // dev, test, prod
)

type Config struct {
	// logging
	LogLevel *slog.Level `json:"log_level"` // debug, error, warn, info

	// language
	Lang *locale.Lang `json:"ietf_lang"` // fr-FR, pt-BR, en-US

	// filename options
	Filename struct {
		Date        bool   `json:"date"`        // prepend current date (YYYY-MM-DD)
		Level       bool   `json:"level"`       // append CEFR level (e.g. A1, B2)
		Lesson      string `json:"lesson"`      // prefix for lesson content
		Preparation string `json:"preparation"` // prefix for preparation materials
		Plan        string `json:"plan"`        // prefix for teacher’s plan
		Reading     string `json:"reading"`     // prefix for reading exercises
		Listening   string `json:"listening"`   // prefix for listening exercises
		Watching    string `json:"watching"`    // prefix for watching exercises
		Correction  string `json:"correction"`  // prefix for correction
	} `json:"filename"`

	// working directories
	InputDir   string `json:"input_dir"`
	StagingDir string `json:"staging_dir"`
	Output     struct {
		StudentDir string `json:"student_dir"` // for preparation, reading, listening, watching
		TeacherDir string `json:"teacher_dir"` // for teacher's lesson, and plan
		LessonsDir string `json:"lessons_dir"` // for everything
	} `json:"output"`

	// infrastructure
	Timeout struct {
		Chat time.Duration `json:"chat"` // chat completion timeout
		TTS  time.Duration `json:"tts"`  // text-to-speech audio timeout
		STT  time.Duration `json:"stt"`  // speech-to-text transcript timeout
	} `json:"timeout"`

	// internal
	tempDirPath    string
	configFilePath string
}

func Load() (*Config, error) {
	return load(configFilename(), tempDirname())
}

func load(configFilename, tempDirname string) (*Config, error) {
	c := &Config{}

	if err := c.buildConfigFilePath(configFilename); err != nil {
		return nil, fmt.Errorf("failed to build configuration file path: %w", err)
	}
	if err := c.buildTempDirPath(tempDirname); err != nil {
		return nil, fmt.Errorf("failed to build temporary directory path: %w", err)
	}

	byt, err := c.readConfigFile()
	if err != nil {
		return nil, fmt.Errorf("failed to read json configuration file %s: %w", c.configFilePath, err)
	}

	dec := json.NewDecoder(bytes.NewReader(byt))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("failed to decode json configuration file %s: %w", c.configFilePath, err)
	}

	if err := c.applyDefaults(); err != nil {
		return nil, fmt.Errorf("failed to apply defaults: %w", err)
	}

	return c, nil
}

func (c *Config) ConfigPath() string {
	return c.configFilePath
}

func (c *Config) TempDir() string {
	return c.tempDirPath
}

func environment() string {
	switch value := os.Getenv(envType); value {
	case "dev", "test":
		return value
	default:
		return "prod"
	}
}

func configFilename() string {
	switch environment() {
	case "dev":
		return "config.dev.json"
	case "test":
		return "config.test.json"
	default:
		return "config.json"
	}
}

func tempDirname() string {
	switch environment() {
	case "dev":
		return "temp_dev"
	case "test":
		return "temp_test"
	default:
		return "temp"
	}
}

func ensureDir(path string) error {
	return os.MkdirAll(path, pathutil.PermUserRWX)
}

func (c *Config) buildConfigFilePath(filename string) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to find user config directory: %w", err)
	}

	fullDir := filepath.Join(configDir, ".config", appName)
	secureDir, err := pathutil.DirPath(fullDir).Validate()
	if err != nil {
		return err
	}
	// ensure config directory exists
	if err := ensureDir(secureDir); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	c.configFilePath = filepath.Join(secureDir, filename)
	return nil
}

func (c *Config) buildTempDirPath(dirname string) error {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return fmt.Errorf("failed to find user cache directory: %w", err)
	}

	fullDir := filepath.Join(cacheDir, appName, dirname)
	secureDir, err := pathutil.DirPath(fullDir).Validate()
	if err != nil {
		return err
	}
	// ensure temp directory exists
	if err := ensureDir(secureDir); err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}

	c.tempDirPath = secureDir
	return nil
}

func (c *Config) readConfigFile() ([]byte, error) {
	dir := filepath.Dir(c.configFilePath)
	file := filepath.Base(c.configFilePath)
	fsys := os.DirFS(dir)

	data, err := fs.ReadFile(fsys, file)
	if err == nil {
		return data, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		if mkErr := ensureDir(dir); mkErr != nil {
			return nil, fmt.Errorf("making configuration directory: %w", mkErr)
		}
		if wErr := os.WriteFile(c.configFilePath, []byte("{}"), pathutil.PermUserRW); wErr != nil {
			return nil, fmt.Errorf("creating default configuration file: %w", wErr)
		}
		// read again via fs.ReadFile
		return fs.ReadFile(fsys, file)
	}
	return nil, err
}

func (c *Config) applyDefaults() error {
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
	if err := defaultDir(&c.InputDir, "input"); err != nil {
		return err
	}
	if err := defaultDir(&c.StagingDir, "staging"); err != nil {
		return err
	}
	if err := defaultDir(&c.Output.StudentDir, "student"); err != nil {
		return err
	}
	if err := defaultDir(&c.Output.TeacherDir, "teacher"); err != nil {
		return err
	}
	if err := defaultDir(&c.Output.LessonsDir, "lessons"); err != nil {
		return err
	}

	// timeouts
	defaultVal(&c.Timeout.Chat, 30*time.Second)
	defaultVal(&c.Timeout.TTS, 5*time.Minute)
	defaultVal(&c.Timeout.STT, 30*time.Second)

	return nil
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
		path, err := buildAppDirPath(dirname)
		if err != nil {
			return err
		}
		*field = path
	}
	return nil
}

func buildAppDirPath(dirname string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to find user home directory: %w", err)
	}

	securePath, err := pathutil.DirPath(filepath.Join(homeDir, appName, dirname)).Validate()
	if err != nil {
		return "", fmt.Errorf("failed to build %s directory path: %w", dirname, err)
	}

	return securePath, nil
}
