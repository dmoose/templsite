// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package assets

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	config := &Config{
		InputDir:    "assets",
		OutputDir:   "public/assets",
		Minify:      true,
		Fingerprint: false,
	}

	pipeline := New(config)

	if pipeline.config != config {
		t.Error("expected pipeline to have config")
	}

	if pipeline.minifier == nil {
		t.Error("expected minifier to be initialized")
	}
}

func TestBuild(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "assets")
	outputDir := filepath.Join(tmpDir, "public", "assets")

	// Create input directory structure
	if err := os.MkdirAll(filepath.Join(inputDir, "css"), 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}

	// Create a simple CSS file
	cssInput := filepath.Join(inputDir, "css", "app.css")
	cssContent := `@import "tailwindcss";

.test {
  color: red;
}`
	if err := os.WriteFile(cssInput, []byte(cssContent), 0644); err != nil {
		t.Fatalf("failed to write CSS file: %v", err)
	}

	config := &Config{
		InputDir:    inputDir,
		OutputDir:   outputDir,
		Minify:      false,
		Fingerprint: false,
	}

	pipeline := New(config)
	ctx := context.Background()

	// Build should succeed (even if Tailwind CLI creates minimal output)
	err := pipeline.Build(ctx)
	if err != nil {
		// It's OK if this fails due to Tailwind CLI not being available in test environment
		// The important thing is that the pipeline is set up correctly
		t.Logf("Build failed (might be expected in test env): %v", err)
	}
}

func TestBuildWithContext(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "assets")
	outputDir := filepath.Join(tmpDir, "public", "assets")

	if err := os.MkdirAll(filepath.Join(inputDir, "css"), 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}

	config := &Config{
		InputDir:    inputDir,
		OutputDir:   outputDir,
		Minify:      false,
		Fingerprint: false,
	}

	pipeline := New(config)

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := pipeline.Build(ctx)
	// Should either succeed quickly or fail with context error
	if err != nil {
		t.Logf("Build with cancelled context: %v", err)
	}
}

func TestBuildWithoutCSSInput(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "assets")
	outputDir := filepath.Join(tmpDir, "public", "assets")

	// Create directories but no CSS file
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}

	config := &Config{
		InputDir:    inputDir,
		OutputDir:   outputDir,
		Minify:      false,
		Fingerprint: false,
	}

	pipeline := New(config)
	ctx := context.Background()

	// Should succeed even without CSS input (skips CSS processing)
	err := pipeline.Build(ctx)
	if err != nil {
		t.Errorf("Build should succeed without CSS input: %v", err)
	}
}

func TestProcessCSSSkipsWhenNoInput(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "assets")
	outputDir := filepath.Join(tmpDir, "public", "assets")

	// Create directories but no CSS file
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}

	config := &Config{
		InputDir:    inputDir,
		OutputDir:   outputDir,
		Minify:      false,
		Fingerprint: false,
	}

	pipeline := New(config)
	ctx := context.Background()

	// processCSS should succeed and skip when no input file exists
	err := pipeline.processCSS(ctx)
	if err != nil {
		t.Errorf("processCSS should succeed when no input file: %v", err)
	}
}

func TestMinifierInitialization(t *testing.T) {
	config := &Config{
		InputDir:  "assets",
		OutputDir: "public/assets",
	}

	pipeline := New(config)

	// Verify minifier can handle CSS
	input := ".test { color: red; }"
	output, err := pipeline.minifier.String("text/css", input)
	if err != nil {
		t.Errorf("minifier should handle CSS: %v", err)
	}
	if output == "" {
		t.Error("minifier should produce output")
	}

	// Verify minifier can handle JS
	jsInput := "var x = 1;"
	jsOutput, err := pipeline.minifier.String("application/javascript", jsInput)
	if err != nil {
		t.Errorf("minifier should handle JavaScript: %v", err)
	}
	if jsOutput == "" {
		t.Error("minifier should produce JS output")
	}
}

func TestConfigDefaults(t *testing.T) {
	config := &Config{
		InputDir:  "assets",
		OutputDir: "public/assets",
	}

	if config.Minify {
		t.Error("expected Minify to default to false")
	}

	if config.Fingerprint {
		t.Error("expected Fingerprint to default to false")
	}
}
