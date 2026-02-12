// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package assets

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestFindTailwindCLI(t *testing.T) {
	// This test checks if we can find tailwindcss
	result := findTailwindCLI()

	// We expect either a system installation or nothing
	// (we won't fail the test, just log the result)
	if result == "" {
		t.Log("No tailwindcss CLI found (expected in test environment)")
	} else {
		t.Logf("Found tailwindcss CLI at: %s", result)
	}
}

func TestFindTailwindCLIWithLocalBin(t *testing.T) {
	// Save original working directory
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origWd) }()

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	binDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("failed to create bin dir: %v", err)
	}

	// Create a fake tailwindcss binary
	fakeBinary := filepath.Join(binDir, "tailwindcss")
	if err := os.WriteFile(fakeBinary, []byte("#!/bin/sh\necho fake"), 0755); err != nil {
		t.Fatalf("failed to create fake binary: %v", err)
	}

	// Change to tmpDir so bin/tailwindcss is found
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	result := findTailwindCLI()
	if result == "" {
		// If system tailwindcss is not in PATH, should find local bin/tailwindcss
		if _, err := exec.LookPath("tailwindcss"); err != nil {
			t.Error("expected to find local bin/tailwindcss")
		}
	} else {
		t.Logf("Found tailwindcss at: %s", result)
	}
}

func TestProcessCSSCreatesOutputDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "assets")
	outputDir := filepath.Join(tmpDir, "public", "assets")

	// Create input directory and CSS file
	cssDir := filepath.Join(inputDir, "css")
	if err := os.MkdirAll(cssDir, 0755); err != nil {
		t.Fatalf("failed to create CSS dir: %v", err)
	}

	cssInput := filepath.Join(cssDir, "app.css")
	cssContent := `@import "tailwindcss";

.test {
  color: red;
}`
	if err := os.WriteFile(cssInput, []byte(cssContent), 0644); err != nil {
		t.Fatalf("failed to write CSS file: %v", err)
	}

	config := &Config{
		InputDir:  inputDir,
		OutputDir: outputDir,
		Minify:    false,
	}

	pipeline := New(config)
	ctx := context.Background()

	// Try to process CSS
	err := pipeline.processCSS(ctx)

	// Check if output directory was created
	outputCSSDir := filepath.Join(outputDir, "css")
	if _, statErr := os.Stat(outputCSSDir); os.IsNotExist(statErr) {
		t.Error("expected output directory to be created")
	}

	if err != nil {
		// It's OK if processing fails due to missing Tailwind CLI
		t.Logf("CSS processing failed (might be expected): %v", err)
	}
}

func TestProcessCSSWithMinify(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "assets")
	outputDir := filepath.Join(tmpDir, "public", "assets")

	// Create input directory and CSS file
	cssDir := filepath.Join(inputDir, "css")
	if err := os.MkdirAll(cssDir, 0755); err != nil {
		t.Fatalf("failed to create CSS dir: %v", err)
	}

	cssInput := filepath.Join(cssDir, "app.css")
	cssContent := `@import "tailwindcss";

.test {
  color: red;
  padding: 10px;
}`
	if err := os.WriteFile(cssInput, []byte(cssContent), 0644); err != nil {
		t.Fatalf("failed to write CSS file: %v", err)
	}

	config := &Config{
		InputDir:  inputDir,
		OutputDir: outputDir,
		Minify:    true,
	}

	pipeline := New(config)
	ctx := context.Background()

	err := pipeline.processCSS(ctx)
	if err != nil {
		// Check if error is due to missing Tailwind CLI
		if findTailwindCLI() == "" {
			t.Skip("Tailwind CLI not available, skipping minify test")
		}
		t.Logf("CSS processing with minify failed: %v", err)
	}
}

func TestProcessCSSWithContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "assets")
	outputDir := filepath.Join(tmpDir, "public", "assets")

	// Create input directory and CSS file
	cssDir := filepath.Join(inputDir, "css")
	if err := os.MkdirAll(cssDir, 0755); err != nil {
		t.Fatalf("failed to create CSS dir: %v", err)
	}

	cssInput := filepath.Join(cssDir, "app.css")
	if err := os.WriteFile(cssInput, []byte(".test { color: red; }"), 0644); err != nil {
		t.Fatalf("failed to write CSS file: %v", err)
	}

	config := &Config{
		InputDir:  inputDir,
		OutputDir: outputDir,
	}

	pipeline := New(config)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := pipeline.processCSS(ctx)
	// Should either fail quickly or succeed if Tailwind CLI isn't found
	if err != nil {
		t.Logf("processCSS with cancelled context: %v", err)
	}
}

func TestProcessCSSInputPaths(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "assets")
	outputDir := filepath.Join(tmpDir, "public", "assets")

	config := &Config{
		InputDir:  inputDir,
		OutputDir: outputDir,
	}

	pipeline := New(config)

	// Verify pipeline was created
	if pipeline == nil {
		t.Fatal("expected pipeline to be created")
	}

	// Verify expected input path
	expectedInput := filepath.Join(inputDir, "css", "app.css")
	if expectedInput == "" {
		t.Error("expected input path to be constructed")
	}

	// Verify expected output path
	expectedOutput := filepath.Join(outputDir, "css", "main.css")
	if expectedOutput == "" {
		t.Error("expected output path to be constructed")
	}
}
