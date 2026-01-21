package site

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
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
outputDir: "public"
`

	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	site, err := New(configPath)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if site.Config == nil {
		t.Error("expected site to have config")
	}

	if site.Config.Title != "Test Site" {
		t.Errorf("expected title 'Test Site', got '%s'", site.Config.Title)
	}
}

func TestNewWithNonexistentConfig(t *testing.T) {
	_, err := New("nonexistent.yaml")
	if err == nil {
		t.Error("expected error when loading nonexistent config")
	}
}

func TestNewWithConfig(t *testing.T) {
	config := DefaultConfig()
	config.Title = "Direct Config"

	site := NewWithConfig(config)
	if site.Config == nil {
		t.Error("expected site to have config")
	}

	if site.Config.Title != "Direct Config" {
		t.Errorf("expected title 'Direct Config', got '%s'", site.Config.Title)
	}
}

func TestSetBaseDir(t *testing.T) {
	site := NewWithConfig(DefaultConfig())

	site.SetBaseDir("/test/base")

	if site.baseDir != "/test/base" {
		t.Errorf("expected baseDir '/test/base', got '%s'", site.baseDir)
	}
}

func TestBuild(t *testing.T) {
	tmpDir := t.TempDir()

	// Create content directory
	contentDir := filepath.Join(tmpDir, "content")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	site := NewWithConfig(DefaultConfig())
	site.SetBaseDir(tmpDir)

	ctx := context.Background()
	err := site.Build(ctx)

	// Build should succeed
	if err != nil {
		t.Errorf("Build failed: %v", err)
	}

	// BuildTime should be set
	if site.BuildTime.IsZero() {
		t.Error("expected BuildTime to be set after build")
	}

	// BuildTime should be recent
	elapsed := time.Since(site.BuildTime)
	if elapsed > time.Second {
		t.Errorf("BuildTime seems too old: %v ago", elapsed)
	}
}

func TestBuildWithContext(t *testing.T) {
	tmpDir := t.TempDir()

	// Create content directory
	contentDir := filepath.Join(tmpDir, "content")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	site := NewWithConfig(DefaultConfig())
	site.SetBaseDir(tmpDir)

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Build should fail with context error
	err := site.Build(ctx)
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestBuildSetsTimestamp(t *testing.T) {
	tmpDir := t.TempDir()

	// Create content directory
	contentDir := filepath.Join(tmpDir, "content")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	site := NewWithConfig(DefaultConfig())
	site.SetBaseDir(tmpDir)

	before := time.Now()
	ctx := context.Background()

	if err := site.Build(ctx); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	after := time.Now()

	// Verify BuildTime is between before and after
	if site.BuildTime.Before(before) {
		t.Error("BuildTime is before build started")
	}

	if site.BuildTime.After(after) {
		t.Error("BuildTime is after build finished")
	}
}

func TestClean(t *testing.T) {
	site := NewWithConfig(DefaultConfig())

	// Clean should not error (even though it's a stub)
	err := site.Clean()
	if err != nil {
		t.Errorf("Clean failed: %v", err)
	}
}
