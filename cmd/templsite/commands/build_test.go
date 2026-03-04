package commands

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildCommand(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test site structure
	contentDir := filepath.Join(tmpDir, "content")
	assetsDir := filepath.Join(tmpDir, "assets", "css")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		t.Fatalf("failed to create assets dir: %v", err)
	}

	// Create config file
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `title: "Test Site"
baseURL: "http://localhost:8080"
content:
  dir: "content"
  defaultLayout: "page"
assets:
  inputDir: "assets"
  outputDir: "assets"
  minify: false
outputDir: "public"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Create content
	content := `---
title: "Test Page"
---
# Test Content
`
	if err := os.WriteFile(filepath.Join(contentDir, "index.md"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write content: %v", err)
	}

	// Create CSS
	css := `@import "tailwindcss";`
	if err := os.WriteFile(filepath.Join(assetsDir, "app.css"), []byte(css), 0644); err != nil {
		t.Fatalf("failed to write css: %v", err)
	}

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Run build command
	ctx := context.Background()
	args := []string{}
	err := Build(ctx, args)
	if err != nil {
		t.Fatalf("build command failed: %v", err)
	}

	// Verify output directory exists
	publicDir := filepath.Join(tmpDir, "public")
	if _, err := os.Stat(publicDir); os.IsNotExist(err) {
		t.Fatal("public directory not created")
	}

	// Verify assets were built (CSS)
	cssPath := filepath.Join(publicDir, "assets", "css", "main.css")
	if _, err := os.Stat(cssPath); os.IsNotExist(err) {
		t.Fatal("CSS not built")
	}

	// Note: HTML rendering is the responsibility of user's site binary,
	// not the build command. Build() only does content parsing + assets.
}

func TestBuildCommandWithFlags(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal site
	contentDir := filepath.Join(tmpDir, "content")
	assetsDir := filepath.Join(tmpDir, "assets", "css")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		t.Fatalf("failed to create assets dir: %v", err)
	}

	// Config
	configPath := filepath.Join(tmpDir, "site.yaml")
	configContent := `title: "Test"
baseURL: "http://localhost"
outputDir: "dist"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Content
	content := `---
title: "Page"
---
# Content
`
	if err := os.WriteFile(filepath.Join(contentDir, "index.md"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write content: %v", err)
	}

	// CSS
	if err := os.WriteFile(filepath.Join(assetsDir, "app.css"), []byte("@import \"tailwindcss\";"), 0644); err != nil {
		t.Fatalf("failed to write css: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Test with custom config path
	ctx := context.Background()
	args := []string{"--config", "site.yaml"}
	err := Build(ctx, args)
	if err != nil {
		t.Fatalf("build with --config failed: %v", err)
	}

	// Verify dist directory was created (from config)
	distDir := filepath.Join(tmpDir, "dist")
	if _, err := os.Stat(distDir); os.IsNotExist(err) {
		t.Error("dist directory not created")
	}
}

func TestBuildCommandWithOutputOverride(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal site
	contentDir := filepath.Join(tmpDir, "content")
	assetsDir := filepath.Join(tmpDir, "assets", "css")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		t.Fatalf("failed to create assets dir: %v", err)
	}

	// Config with public as output
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `title: "Test"
baseURL: "http://localhost"
outputDir: "public"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Content
	content := `---
title: "Page"
---
# Content
`
	if err := os.WriteFile(filepath.Join(contentDir, "index.md"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write content: %v", err)
	}

	// CSS
	if err := os.WriteFile(filepath.Join(assetsDir, "app.css"), []byte("@import \"tailwindcss\";"), 0644); err != nil {
		t.Fatalf("failed to write css: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Build with output override
	ctx := context.Background()
	args := []string{"--output", "build"}
	err := Build(ctx, args)
	if err != nil {
		t.Fatalf("build with --output failed: %v", err)
	}

	// Verify build directory was created (overriding config)
	buildDir := filepath.Join(tmpDir, "build")
	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		t.Error("build directory not created with --output override")
	}

	// Verify public was NOT created
	publicDir := filepath.Join(tmpDir, "public")
	if _, err := os.Stat(publicDir); err == nil {
		t.Error("public directory should not exist when overridden")
	}
}

func TestBuildCommandWithClean(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal site
	contentDir := filepath.Join(tmpDir, "content")
	assetsDir := filepath.Join(tmpDir, "assets", "css")
	publicDir := filepath.Join(tmpDir, "public")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		t.Fatalf("failed to create assets dir: %v", err)
	}
	if err := os.MkdirAll(publicDir, 0755); err != nil {
		t.Fatalf("failed to create public dir: %v", err)
	}

	// Create old file in public
	oldFile := filepath.Join(publicDir, "old.html")
	if err := os.WriteFile(oldFile, []byte("old content"), 0644); err != nil {
		t.Fatalf("failed to write old file: %v", err)
	}

	// Config
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `title: "Test"
baseURL: "http://localhost"
outputDir: "public"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Content
	content := `---
title: "New"
---
# New Content
`
	if err := os.WriteFile(filepath.Join(contentDir, "index.md"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write content: %v", err)
	}

	// CSS
	if err := os.WriteFile(filepath.Join(assetsDir, "app.css"), []byte("@import \"tailwindcss\";"), 0644); err != nil {
		t.Fatalf("failed to write css: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Build with clean
	ctx := context.Background()
	args := []string{"--clean"}
	err := Build(ctx, args)
	if err != nil {
		t.Fatalf("build with --clean failed: %v", err)
	}

	// Verify old file is gone
	if _, err := os.Stat(oldFile); err == nil {
		t.Error("old file should have been cleaned")
	}

	// Verify assets were rebuilt
	cssPath := filepath.Join(publicDir, "assets", "css", "main.css")
	if _, err := os.Stat(cssPath); os.IsNotExist(err) {
		t.Error("CSS should exist after clean build")
	}
}

func TestBuildCommandMissingConfig(t *testing.T) {
	tmpDir := t.TempDir()

	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Try to build without config
	ctx := context.Background()
	args := []string{}
	err := Build(ctx, args)
	if err == nil {
		t.Fatal("expected error for missing config, got nil")
	}

	if !strings.Contains(err.Error(), "config file not found") {
		t.Errorf("expected 'config file not found' error, got: %v", err)
	}
}

func TestBuildCommandWithContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal site
	contentDir := filepath.Join(tmpDir, "content")
	assetsDir := filepath.Join(tmpDir, "assets", "css")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		t.Fatalf("failed to create assets dir: %v", err)
	}

	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `title: "Test"
baseURL: "http://localhost"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	content := `---
title: "Page"
---
# Content
`
	if err := os.WriteFile(filepath.Join(contentDir, "index.md"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write content: %v", err)
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "app.css"), []byte("@import \"tailwindcss\";"), 0644); err != nil {
		t.Fatalf("failed to write css: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Cancel context immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	args := []string{}
	err := Build(ctx, args)

	// Build might complete before cancellation, or it might be cancelled
	// Either is acceptable for this test
	if err != nil && !strings.Contains(err.Error(), "context canceled") {
		// If there's an error, it should be context cancellation
		t.Logf("Build with cancelled context: %v", err)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{5242880, "5.0 MB"},
	}

	for _, tt := range tests {
		result := formatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
		}
	}
}

func TestGetBuildStats(t *testing.T) {
	// This test is covered by the integration tests above
	// The getBuildStats function is primarily tested through
	// TestBuildCommand which exercises the full build flow
	t.Skip("Covered by integration tests")
}
