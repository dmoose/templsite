package site

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Title == "" {
		t.Error("default config should have a title")
	}

	if config.BaseURL == "" {
		t.Error("default config should have a baseURL")
	}

	if config.Content.Dir == "" {
		t.Error("default config should have content.dir")
	}

	if config.Content.DefaultLayout == "" {
		t.Error("default config should have content.defaultLayout")
	}

	if config.Assets.InputDir == "" {
		t.Error("default config should have assets.inputDir")
	}

	if config.OutputDir == "" {
		t.Error("default config should have outputDir")
	}

	// Validate default config
	if err := config.Validate(); err != nil {
		t.Errorf("default config should be valid: %v", err)
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configYAML := `
title: "Test Site"
baseURL: "https://example.com"
content:
  dir: "content"
  defaultLayout: "page"
assets:
  inputDir: "assets"
  outputDir: "assets"
  minify: true
  fingerprint: true
outputDir: "public"
themeColor: "dark"
`

	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if config.Title != "Test Site" {
		t.Errorf("expected title 'Test Site', got '%s'", config.Title)
	}

	if config.BaseURL != "https://example.com" {
		t.Errorf("expected baseURL 'https://example.com', got '%s'", config.BaseURL)
	}

	if config.Content.Dir != "content" {
		t.Errorf("expected content.dir 'content', got '%s'", config.Content.Dir)
	}

	if config.Content.DefaultLayout != "page" {
		t.Errorf("expected defaultLayout 'page', got '%s'", config.Content.DefaultLayout)
	}

	if !config.Assets.Minify {
		t.Error("expected assets.minify to be true")
	}

	if !config.Assets.Fingerprint {
		t.Error("expected assets.fingerprint to be true")
	}

	if config.ThemeColor != "dark" {
		t.Errorf("expected themeColor 'dark', got '%s'", config.ThemeColor)
	}
}

func TestLoadConfigNotFound(t *testing.T) {
	_, err := LoadConfig("nonexistent.yaml")
	if err == nil {
		t.Error("expected error when loading nonexistent config")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	invalidYAML := `
title: "Test"
baseURL: [this is not valid yaml
`

	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("expected error when parsing invalid YAML")
	}
}

func TestLoadConfigOrDefault(t *testing.T) {
	// Test with nonexistent file - should return defaults
	config, err := LoadConfigOrDefault("nonexistent.yaml")
	if err != nil {
		t.Fatalf("LoadConfigOrDefault should not error on missing file: %v", err)
	}

	if config.Title != "My Site" {
		t.Errorf("expected default title, got '%s'", config.Title)
	}

	// Test with existing file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configYAML := `
title: "Custom Site"
baseURL: "https://example.com"
content:
  dir: "content"
  defaultLayout: "page"
assets:
  inputDir: "assets"
  outputDir: "assets"
outputDir: "public"
`

	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	config, err = LoadConfigOrDefault(configPath)
	if err != nil {
		t.Fatalf("LoadConfigOrDefault failed: %v", err)
	}

	if config.Title != "Custom Site" {
		t.Errorf("expected title 'Custom Site', got '%s'", config.Title)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		shouldErr bool
	}{
		{
			name:      "valid config",
			config:    DefaultConfig(),
			shouldErr: false,
		},
		{
			name: "missing title",
			config: &Config{
				BaseURL: "https://example.com",
				Content: ContentConfig{
					Dir:           "content",
					DefaultLayout: "page",
				},
				Assets: AssetsConfig{
					InputDir:  "assets",
					OutputDir: "assets",
				},
				OutputDir: "public",
			},
			shouldErr: true,
		},
		{
			name: "missing baseURL",
			config: &Config{
				Title: "Test",
				Content: ContentConfig{
					Dir:           "content",
					DefaultLayout: "page",
				},
				Assets: AssetsConfig{
					InputDir:  "assets",
					OutputDir: "assets",
				},
				OutputDir: "public",
			},
			shouldErr: true,
		},
		{
			name: "missing content.dir",
			config: &Config{
				Title:   "Test",
				BaseURL: "https://example.com",
				Content: ContentConfig{
					DefaultLayout: "page",
				},
				Assets: AssetsConfig{
					InputDir:  "assets",
					OutputDir: "assets",
				},
				OutputDir: "public",
			},
			shouldErr: true,
		},
		{
			name: "missing outputDir",
			config: &Config{
				Title:   "Test",
				BaseURL: "https://example.com",
				Content: ContentConfig{
					Dir:           "content",
					DefaultLayout: "page",
				},
				Assets: AssetsConfig{
					InputDir:  "assets",
					OutputDir: "assets",
				},
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.shouldErr && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("expected no validation error, got: %v", err)
			}
		})
	}
}

func TestConfigSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	config := DefaultConfig()
	config.Title = "Saved Site"

	if err := config.Save(configPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	// Load it back and verify
	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}

	if loaded.Title != "Saved Site" {
		t.Errorf("expected title 'Saved Site', got '%s'", loaded.Title)
	}
}

func TestConfigPaths(t *testing.T) {
	config := DefaultConfig()
	baseDir := "/test/base"

	// Test relative paths
	contentPath := config.ContentPath(baseDir)
	expected := filepath.Join(baseDir, "content")
	if contentPath != expected {
		t.Errorf("expected content path '%s', got '%s'", expected, contentPath)
	}

	assetsInputPath := config.AssetsInputPath(baseDir)
	expected = filepath.Join(baseDir, "assets")
	if assetsInputPath != expected {
		t.Errorf("expected assets input path '%s', got '%s'", expected, assetsInputPath)
	}

	outputPath := config.OutputPath(baseDir)
	expected = filepath.Join(baseDir, "public")
	if outputPath != expected {
		t.Errorf("expected output path '%s', got '%s'", expected, outputPath)
	}

	assetsOutputPath := config.AssetsOutputPath(baseDir)
	expected = filepath.Join(baseDir, "public", "assets")
	if assetsOutputPath != expected {
		t.Errorf("expected assets output path '%s', got '%s'", expected, assetsOutputPath)
	}

	// Test absolute paths
	config.Content.Dir = "/absolute/content"
	contentPath = config.ContentPath(baseDir)
	if contentPath != "/absolute/content" {
		t.Errorf("expected absolute path '/absolute/content', got '%s'", contentPath)
	}
}

func TestConfigPartialOverride(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Only override some fields, others should use defaults
	configYAML := `
title: "Partial Config"
baseURL: "https://example.com"
`

	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Custom values
	if config.Title != "Partial Config" {
		t.Errorf("expected title 'Partial Config', got '%s'", config.Title)
	}

	if config.BaseURL != "https://example.com" {
		t.Errorf("expected baseURL 'https://example.com', got '%s'", config.BaseURL)
	}

	// Default values should still be present
	if config.Content.Dir != "content" {
		t.Errorf("expected default content.dir 'content', got '%s'", config.Content.Dir)
	}

	if config.OutputDir != "public" {
		t.Errorf("expected default outputDir 'public', got '%s'", config.OutputDir)
	}
}

func TestLoadConfigWithEnv(t *testing.T) {
	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, "config.yaml")
	envPath := filepath.Join(tmpDir, "config.production.yaml")

	baseYAML := `
title: "My Site"
baseURL: "http://localhost:8080"
description: "Dev description"
content:
  dir: "content"
  defaultLayout: "page"
assets:
  inputDir: "assets"
  outputDir: "assets"
outputDir: "public"
`
	envYAML := `
baseURL: "https://mysite.com"
description: "Production description"
`

	if err := os.WriteFile(basePath, []byte(baseYAML), 0644); err != nil {
		t.Fatalf("failed to write base config: %v", err)
	}
	if err := os.WriteFile(envPath, []byte(envYAML), 0644); err != nil {
		t.Fatalf("failed to write env config: %v", err)
	}

	config, err := LoadConfigWithEnv(basePath, "production")
	if err != nil {
		t.Fatalf("LoadConfigWithEnv failed: %v", err)
	}

	// Overridden values
	if config.BaseURL != "https://mysite.com" {
		t.Errorf("expected baseURL 'https://mysite.com', got '%s'", config.BaseURL)
	}
	if config.Description != "Production description" {
		t.Errorf("expected description 'Production description', got '%s'", config.Description)
	}

	// Preserved base values
	if config.Title != "My Site" {
		t.Errorf("expected title 'My Site', got '%s'", config.Title)
	}
	if config.OutputDir != "public" {
		t.Errorf("expected outputDir 'public', got '%s'", config.OutputDir)
	}
}

func TestLoadConfigWithEnvParams(t *testing.T) {
	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, "config.yaml")
	envPath := filepath.Join(tmpDir, "config.production.yaml")

	baseYAML := `
title: "My Site"
baseURL: "http://localhost:8080"
content:
  dir: "content"
  defaultLayout: "page"
assets:
  inputDir: "assets"
  outputDir: "assets"
outputDir: "public"
params:
  captchaSiteKey: "test-key-123"
  siteName: "dev"
`
	envYAML := `
params:
  captchaSiteKey: "real-key-456"
  analyticsId: "GA-PROD"
`

	if err := os.WriteFile(basePath, []byte(baseYAML), 0644); err != nil {
		t.Fatalf("failed to write base config: %v", err)
	}
	if err := os.WriteFile(envPath, []byte(envYAML), 0644); err != nil {
		t.Fatalf("failed to write env config: %v", err)
	}

	config, err := LoadConfigWithEnv(basePath, "production")
	if err != nil {
		t.Fatalf("LoadConfigWithEnv failed: %v", err)
	}

	// Overridden param
	if config.Params["captchaSiteKey"] != "real-key-456" {
		t.Errorf("expected captchaSiteKey 'real-key-456', got '%v'", config.Params["captchaSiteKey"])
	}

	// New param from env
	if config.Params["analyticsId"] != "GA-PROD" {
		t.Errorf("expected analyticsId 'GA-PROD', got '%v'", config.Params["analyticsId"])
	}

	// Preserved base param
	if config.Params["siteName"] != "dev" {
		t.Errorf("expected siteName 'dev', got '%v'", config.Params["siteName"])
	}
}

func TestLoadConfigWithEnvMissing(t *testing.T) {
	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, "config.yaml")

	baseYAML := `
title: "My Site"
baseURL: "http://localhost:8080"
content:
  dir: "content"
  defaultLayout: "page"
assets:
  inputDir: "assets"
  outputDir: "assets"
outputDir: "public"
`

	if err := os.WriteFile(basePath, []byte(baseYAML), 0644); err != nil {
		t.Fatalf("failed to write base config: %v", err)
	}

	// No config.staging.yaml exists — should succeed with base config
	config, err := LoadConfigWithEnv(basePath, "staging")
	if err != nil {
		t.Fatalf("LoadConfigWithEnv should not error on missing env file: %v", err)
	}

	if config.BaseURL != "http://localhost:8080" {
		t.Errorf("expected base baseURL, got '%s'", config.BaseURL)
	}
}

func TestLoadConfigWithEnvEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, "config.yaml")

	baseYAML := `
title: "My Site"
baseURL: "http://localhost:8080"
content:
  dir: "content"
  defaultLayout: "page"
assets:
  inputDir: "assets"
  outputDir: "assets"
outputDir: "public"
`

	if err := os.WriteFile(basePath, []byte(baseYAML), 0644); err != nil {
		t.Fatalf("failed to write base config: %v", err)
	}

	// Empty env string — behaves like LoadConfig
	config, err := LoadConfigWithEnv(basePath, "")
	if err != nil {
		t.Fatalf("LoadConfigWithEnv with empty env failed: %v", err)
	}

	if config.Title != "My Site" {
		t.Errorf("expected title 'My Site', got '%s'", config.Title)
	}
}
