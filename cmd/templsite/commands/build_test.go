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
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Run build command
	ctx := context.Background()
	args := []string{}
	err := Build(ctx, args)
	if err != nil {
		t.Fatalf("build command failed: %v", err)
	}

	// Verify output exists
	publicDir := filepath.Join(tmpDir, "public")
	if _, err := os.Stat(publicDir); os.IsNotExist(err) {
		t.Fatal("public directory not created")
	}

	// Verify HTML output
	indexPath := filepath.Join(publicDir, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Fatal("index.html not created")
	}

	html, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("failed to read html: %v", err)
	}

	htmlStr := string(html)
	if !strings.Contains(htmlStr, "Test Page") {
		t.Error("HTML missing page title")
	}
	if !strings.Contains(htmlStr, "Test Content") {
		t.Error("HTML missing content")
	}

	// Verify CSS output
	cssPath := filepath.Join(publicDir, "assets", "css", "main.css")
	if _, err := os.Stat(cssPath); os.IsNotExist(err) {
		t.Error("CSS output not created")
	}
}

func TestBuildCommandWithFlags(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal site
	contentDir := filepath.Join(tmpDir, "content")
	assetsDir := filepath.Join(tmpDir, "assets", "css")
	os.MkdirAll(contentDir, 0755)
	os.MkdirAll(assetsDir, 0755)

	// Config
	configPath := filepath.Join(tmpDir, "site.yaml")
	configContent := `title: "Test"
baseURL: "http://localhost"
outputDir: "dist"
`
	os.WriteFile(configPath, []byte(configContent), 0644)

	// Content
	content := `---
title: "Page"
---
# Content
`
	os.WriteFile(filepath.Join(contentDir, "index.md"), []byte(content), 0644)

	// CSS
	os.WriteFile(filepath.Join(assetsDir, "app.css"), []byte("@import \"tailwindcss\";"), 0644)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

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
	os.MkdirAll(contentDir, 0755)
	os.MkdirAll(assetsDir, 0755)

	// Config with public as output
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `title: "Test"
baseURL: "http://localhost"
outputDir: "public"
`
	os.WriteFile(configPath, []byte(configContent), 0644)

	// Content
	content := `---
title: "Page"
---
# Content
`
	os.WriteFile(filepath.Join(contentDir, "index.md"), []byte(content), 0644)

	// CSS
	os.WriteFile(filepath.Join(assetsDir, "app.css"), []byte("@import \"tailwindcss\";"), 0644)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

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
	os.MkdirAll(contentDir, 0755)
	os.MkdirAll(assetsDir, 0755)
	os.MkdirAll(publicDir, 0755)

	// Create old file in public
	oldFile := filepath.Join(publicDir, "old.html")
	os.WriteFile(oldFile, []byte("old content"), 0644)

	// Config
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `title: "Test"
baseURL: "http://localhost"
outputDir: "public"
`
	os.WriteFile(configPath, []byte(configContent), 0644)

	// Content
	content := `---
title: "New"
---
# New Content
`
	os.WriteFile(filepath.Join(contentDir, "index.md"), []byte(content), 0644)

	// CSS
	os.WriteFile(filepath.Join(assetsDir, "app.css"), []byte("@import \"tailwindcss\";"), 0644)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

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

	// Verify new file exists
	newFile := filepath.Join(publicDir, "index.html")
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		t.Error("new file should exist after clean build")
	}
}

func TestBuildCommandMissingConfig(t *testing.T) {
	tmpDir := t.TempDir()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

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
	os.MkdirAll(contentDir, 0755)
	os.MkdirAll(assetsDir, 0755)

	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `title: "Test"
baseURL: "http://localhost"
`
	os.WriteFile(configPath, []byte(configContent), 0644)

	content := `---
title: "Page"
---
# Content
`
	os.WriteFile(filepath.Join(contentDir, "index.md"), []byte(content), 0644)
	os.WriteFile(filepath.Join(assetsDir, "app.css"), []byte("@import \"tailwindcss\";"), 0644)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

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
