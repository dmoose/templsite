package site

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the site configuration
type Config struct {
	Title       string        `yaml:"title"`
	BaseURL     string        `yaml:"baseURL"`
	Description string        `yaml:"description,omitempty"`
	Language    string        `yaml:"language,omitempty"`
	Content     ContentConfig `yaml:"content"`
	Assets      AssetsConfig  `yaml:"assets"`
	OutputDir   string        `yaml:"outputDir"`
	ThemeColor  string        `yaml:"themeColor,omitempty"`

	// Taxonomies lists the taxonomy names to build (e.g., ["tags", "categories"])
	// Each taxonomy extracts terms from page frontmatter with the same key
	Taxonomies []string `yaml:"taxonomies,omitempty"`

	// Menus defines navigation menus (e.g., "main", "footer")
	Menus map[string][]MenuItemConfig `yaml:"menus,omitempty"`

	// DataDir is the directory containing data files (YAML/JSON)
	DataDir string `yaml:"dataDir,omitempty"`

	// StaticDir is the directory containing files to copy to the output root
	// (favicons, etc.). Defaults to "static".
	StaticDir string `yaml:"staticDir,omitempty"`

	// Build contains build-time options
	Build BuildConfig `yaml:"build,omitempty"`

	// Highlight configures syntax highlighting for code blocks
	Highlight HighlightConfig `yaml:"highlight,omitempty"`

	// Params holds arbitrary user-defined parameters accessible in templates
	// via s.Config.Params["key"]. Environment overrides merge into this map.
	Params map[string]any `yaml:"params,omitempty"`
}

// MenuItemConfig represents a menu item in configuration
type MenuItemConfig struct {
	Name   string `yaml:"name"`
	URL    string `yaml:"url"`
	Weight int    `yaml:"weight,omitempty"`
}

// BuildConfig contains build-time options
type BuildConfig struct {
	// Drafts includes draft pages in the build when true
	Drafts bool `yaml:"drafts,omitempty"`
	// Future includes pages with future dates when true
	Future bool `yaml:"future,omitempty"`
}

// HighlightConfig configures syntax highlighting for code blocks
type HighlightConfig struct {
	// Style is the Chroma style name (e.g., "monokai", "github", "dracula").
	// When set, code blocks are syntax-highlighted and a chroma.css file is generated.
	// Empty string disables syntax highlighting.
	Style string `yaml:"style,omitempty"`
	// LineNumbers adds line numbers to code blocks when true
	LineNumbers bool `yaml:"lineNumbers,omitempty"`
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
		Taxonomies: []string{"tags"},
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

	// Allow environment variable to override BaseURL for production deploys
	if envURL := os.Getenv("SITE_BASE_URL"); envURL != "" {
		config.BaseURL = envURL
	}

	// Validate
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// LoadConfigWithEnv loads the base config and merges an environment-specific
// override file on top of it. For example, with basePath "config.yaml" and
// env "production", it looks for "config.production.yaml" in the same directory.
// If env is empty, behaves identically to LoadConfig.
// If the env override file does not exist, the base config is used without error.
func LoadConfigWithEnv(basePath string, env string) (*Config, error) {
	config, err := LoadConfig(basePath)
	if err != nil {
		return nil, err
	}

	if env == "" {
		return config, nil
	}

	// Derive env config path: config.yaml → config.production.yaml
	ext := filepath.Ext(basePath)
	envPath := basePath[:len(basePath)-len(ext)] + "." + env + ext

	// If env config file doesn't exist, use base config as-is
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return config, nil
	}

	// Read and parse env override
	data, err := os.ReadFile(envPath)
	if err != nil {
		return nil, fmt.Errorf("reading env config file %s: %w", envPath, err)
	}

	// Unmarshal into a fresh struct (not DefaultConfig) so we can detect
	// which fields were actually set in the override file
	var override Config
	if err := yaml.Unmarshal(data, &override); err != nil {
		return nil, fmt.Errorf("parsing env config YAML %s: %w", envPath, err)
	}

	mergeConfig(config, &override)

	// Re-apply env var override (takes highest precedence)
	if envURL := os.Getenv("SITE_BASE_URL"); envURL != "" {
		config.BaseURL = envURL
	}

	// Re-validate after merge
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration after env merge: %w", err)
	}

	return config, nil
}

// mergeConfig merges non-zero fields from override into base.
// Scalar fields are overwritten if the override value is non-zero.
// Map fields (Params, Menus) are key-merged so base keys not in override are preserved.
func mergeConfig(base, override *Config) {
	if override.Title != "" {
		base.Title = override.Title
	}
	if override.BaseURL != "" {
		base.BaseURL = override.BaseURL
	}
	if override.Description != "" {
		base.Description = override.Description
	}
	if override.Language != "" {
		base.Language = override.Language
	}
	if override.ThemeColor != "" {
		base.ThemeColor = override.ThemeColor
	}
	if override.OutputDir != "" {
		base.OutputDir = override.OutputDir
	}
	if override.DataDir != "" {
		base.DataDir = override.DataDir
	}
	if override.StaticDir != "" {
		base.StaticDir = override.StaticDir
	}

	// Content config
	if override.Content.Dir != "" {
		base.Content.Dir = override.Content.Dir
	}
	if override.Content.DefaultLayout != "" {
		base.Content.DefaultLayout = override.Content.DefaultLayout
	}

	// Assets config
	if override.Assets.InputDir != "" {
		base.Assets.InputDir = override.Assets.InputDir
	}
	if override.Assets.OutputDir != "" {
		base.Assets.OutputDir = override.Assets.OutputDir
	}
	if override.Assets.Minify {
		base.Assets.Minify = true
	}
	if override.Assets.Fingerprint {
		base.Assets.Fingerprint = true
	}

	// Build config
	if override.Build.Drafts {
		base.Build.Drafts = true
	}
	if override.Build.Future {
		base.Build.Future = true
	}

	// Highlight config
	if override.Highlight.Style != "" {
		base.Highlight.Style = override.Highlight.Style
	}
	if override.Highlight.LineNumbers {
		base.Highlight.LineNumbers = true
	}

	// Taxonomies — replace entirely if override specifies any
	if len(override.Taxonomies) > 0 {
		base.Taxonomies = override.Taxonomies
	}

	// Menus — key-merge
	if len(override.Menus) > 0 {
		if base.Menus == nil {
			base.Menus = make(map[string][]MenuItemConfig)
		}
		for k, v := range override.Menus {
			base.Menus[k] = v
		}
	}

	// Params — key-merge
	if len(override.Params) > 0 {
		if base.Params == nil {
			base.Params = make(map[string]any)
		}
		for k, v := range override.Params {
			base.Params[k] = v
		}
	}
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

// StaticPath returns the absolute path to the static directory
func (c *Config) StaticPath(base string) string {
	dir := c.StaticDir
	if dir == "" {
		dir = "static"
	}
	if filepath.IsAbs(dir) {
		return dir
	}
	return filepath.Join(base, dir)
}

// DataPath returns the absolute path to the data directory
func (c *Config) DataPath(base string) string {
	if c.DataDir == "" {
		return filepath.Join(base, "data")
	}
	if filepath.IsAbs(c.DataDir) {
		return c.DataDir
	}
	return filepath.Join(base, c.DataDir)
}
