package site

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the site configuration
type Config struct {
	Title      string        `yaml:"title"`
	BaseURL    string        `yaml:"baseURL"`
	Content    ContentConfig `yaml:"content"`
	Assets     AssetsConfig  `yaml:"assets"`
	OutputDir  string        `yaml:"outputDir"`
	ThemeColor string        `yaml:"themeColor,omitempty"`
}

// ContentConfig configures content processing
type ContentConfig struct {
	Dir           string `yaml:"dir"`
	DefaultLayout string `yaml:"defaultLayout"`
}

// AssetsConfig configures asset processing
type AssetsConfig struct {
	InputDir    string `yaml:"inputDir"`
	OutputDir   string `yaml:"outputDir"`
	Minify      bool   `yaml:"minify"`
	Fingerprint bool   `yaml:"fingerprint"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Title:   "My Site",
		BaseURL: "http://localhost:8080",
		Content: ContentConfig{
			Dir:           "content",
			DefaultLayout: "page",
		},
		Assets: AssetsConfig{
			InputDir:    "assets",
			OutputDir:   "assets",
			Minify:      false,
			Fingerprint: false,
		},
		OutputDir:  "public",
		ThemeColor: "light",
	}
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	// Start with defaults
	config := DefaultConfig()

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("parsing config YAML: %w", err)
	}

	// Validate
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// LoadConfigOrDefault loads configuration or returns defaults if file doesn't exist
func LoadConfigOrDefault(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}
	return LoadConfig(path)
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Title == "" {
		return fmt.Errorf("title is required")
	}

	if c.BaseURL == "" {
		return fmt.Errorf("baseURL is required")
	}

	if c.Content.Dir == "" {
		return fmt.Errorf("content.dir is required")
	}

	if c.Content.DefaultLayout == "" {
		return fmt.Errorf("content.defaultLayout is required")
	}

	if c.Assets.InputDir == "" {
		return fmt.Errorf("assets.inputDir is required")
	}

	if c.Assets.OutputDir == "" {
		return fmt.Errorf("assets.outputDir is required")
	}

	if c.OutputDir == "" {
		return fmt.Errorf("outputDir is required")
	}

	return nil
}

// Save writes the configuration to a YAML file
func (c *Config) Save(path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// ContentPath returns the absolute path to the content directory
func (c *Config) ContentPath(base string) string {
	if filepath.IsAbs(c.Content.Dir) {
		return c.Content.Dir
	}
	return filepath.Join(base, c.Content.Dir)
}

// AssetsInputPath returns the absolute path to the assets input directory
func (c *Config) AssetsInputPath(base string) string {
	if filepath.IsAbs(c.Assets.InputDir) {
		return c.Assets.InputDir
	}
	return filepath.Join(base, c.Assets.InputDir)
}

// OutputPath returns the absolute path to the output directory
func (c *Config) OutputPath(base string) string {
	if filepath.IsAbs(c.OutputDir) {
		return c.OutputDir
	}
	return filepath.Join(base, c.OutputDir)
}

// AssetsOutputPath returns the absolute path to the assets output directory
func (c *Config) AssetsOutputPath(base string) string {
	outputBase := c.OutputPath(base)
	return filepath.Join(outputBase, c.Assets.OutputDir)
}
