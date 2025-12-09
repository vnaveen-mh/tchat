package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/fatih/color"
)

var once sync.Once
var cfg *Config
var cfgErr error

// user for internal json unmarshaling
type rawConfig struct {
	SystemPrompt string `json:"system_prompt"`
	Model        string `json:"Model"`

	// History Settings
	MaxMessages int `json:"max_messages"`

	// Logging Settings
	LogLevel string `json:"log_level"`

	// Color Settings
	Colors ColorConfig `json:"colors"`
}

// Config holds all application configuration
type Config struct {
	systemPrompt string
	model        string
	maxMessages  int
	logLevel     string

	colors ColorConfig

	appDir string

	promptColor   *color.Color `json:"-"`
	infoColor     *color.Color `json:"-"`
	errorColor    *color.Color `json:"-"`
	outputColor   *color.Color `json:"-"`
	asciiArtColor *color.Color `json:"-"`
}

// colorConfig holds color configuration for UI elements
type ColorConfig struct {
	Prompt string `json:"prompt"` // Prompt color (>)
	Info   string `json:"info"`   // Info message color
	Error  string `json:"error"`  // Error message color
	Output string `json:"output"` // AI output color
}

// New creates a new Config with the given app directory
func New() (*Config, error) {
	once.Do(func() {
		appDir := getAppDir()

		// create app dir if does not exist
		if err := os.MkdirAll(appDir, 0755); err != nil {
			cfgErr = err
			return
		}

		cfg = &Config{
			appDir: appDir,
		}

		// Set defaults
		cfg.setDefaults()

		if err := cfg.load(); err != nil {
			fmt.Printf("Error loading from config file, using defaults\n")
			return
		}
	})
	if cfgErr != nil {
		return nil, cfgErr
	}

	// Initialize color objects
	cfg.initializeColors()
	return cfg, nil
}

// SetDefaults sets default values for all configuration options
func (c *Config) setDefaults() {
	c.systemPrompt = DefaultSystemPrompt
	c.maxMessages = DefaultMaxMessages
	c.logLevel = DefaultLogLevel

	// Set default colors
	c.colors = defaultColors()
}

// DefaultColors returns default color configuration
func defaultColors() ColorConfig {
	return ColorConfig{
		Prompt: DefaultColorPrompt,
		Info:   DefaultColorInfo,
		Error:  DefaultColorError,
		Output: DefaultColorOutput,
	}
}

// GetSystemPrompt returns the current system prompt
func (c *Config) GetSystemPrompt() string {
	return c.systemPrompt
}

// GetModel returns the model from the config file
func (c *Config) GetModel() string {
	return c.model
}

// ConfigPath returns the path to the config file
func (c *Config) ConfigPath() string {
	return filepath.Join(c.appDir, "config.json")
}

func (c *Config) GetMaxMessages() int {
	return c.maxMessages
}

func (c *Config) GetLogLevel() string {
	return c.logLevel
}

func (c *Config) GetAppDir() string {
	return c.appDir
}

func (c *Config) Colors() ColorConfig {
	return c.colors
}

// String returns a human-readable representation of the config
func (c *Config) String() string {
	return fmt.Sprintf(`Configuration:
  Model: %s,
  System Prompt: %s
  Max Messages: %d
  Log Level: %s
  Config File: %s`,
		c.model,
		c.systemPrompt,
		c.maxMessages,
		c.logLevel,
		c.ConfigPath(),
	)
}

// parseColorName converts a color name string to a color.Attribute
func parseColorName(colorName string) color.Attribute {
	switch colorName {
	case "black":
		return color.FgBlack
	case "red":
		return color.FgRed
	case "green":
		return color.FgGreen
	case "yellow":
		return color.FgYellow
	case "blue":
		return color.FgBlue
	case "magenta":
		return color.FgMagenta
	case "cyan":
		return color.FgCyan
	case "white":
		return color.FgWhite
	default:
		// Default to white if invalid color name
		return color.FgWhite
	}
}

// initializeColors creates color objects from color configuration strings
// This method assumes the caller already holds the mutex lock
func (c *Config) initializeColors() {
	c.promptColor = color.New(parseColorName(c.colors.Prompt))
	c.infoColor = color.New(parseColorName(c.colors.Info))
	c.errorColor = color.New(parseColorName(c.colors.Error))
	c.outputColor = color.New(parseColorName(c.colors.Output))
	c.asciiArtColor = color.New(color.FgYellow, color.Bold) // Always yellow/bold
}

// getDefaultAppDir returns the application data directory (~/.tchat/)
// and creates it if it doesn't exist
func getAppDir() string {
	appDir := os.Getenv("TCHAT_APPDIR")
	if appDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Error while getting homedir: %s", err)
		}
		appDir = filepath.Join(homeDir, ".tchat")
	}

	return appDir
}

// load loads configuration from disk
func (c *Config) load() error {
	configPath := c.ConfigPath()

	// If config file doesn't exist, use defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil // Not an error, just use defaults
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var r rawConfig
	if err := json.Unmarshal(data, &r); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	if r.SystemPrompt != "" {
		c.systemPrompt = r.SystemPrompt
	}
	if r.Model != "" {
		c.model = r.Model
	}
	if r.MaxMessages != 0 {
		c.maxMessages = r.MaxMessages
	}
	if r.LogLevel != "" {
		c.logLevel = r.LogLevel
	}
	c.colors = r.Colors

	return nil
}
