package assets

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

// processCSS processes CSS files with Tailwind CLI
func (p *Pipeline) processCSS(ctx context.Context) error {
	// Define input and output paths
	inputFile := filepath.Join(p.config.InputDir, "css", "app.css")
	outputDir := filepath.Join(p.config.OutputDir, "css")
	outputFile := filepath.Join(outputDir, "main.css")

	// Check if input file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		slog.Debug("no CSS input file found, skipping", "path", inputFile)
		return nil
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Determine which tailwindcss binary to use
	tailwindCmd := findTailwindCLI()
	if tailwindCmd == "" {
		return fmt.Errorf("tailwindcss CLI not found in PATH or bin/tailwindcss")
	}

	slog.Info("processing CSS", "input", inputFile, "output", outputFile, "minify", p.config.Minify)

	// Build command arguments
	args := []string{
		"-i", inputFile,
		"-o", outputFile,
	}

	// Add minify flag if enabled
	if p.config.Minify {
		args = append(args, "--minify")
	}

	// Execute Tailwind CLI
	cmd := exec.CommandContext(ctx, tailwindCmd, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running tailwindcss: %w", err)
	}

	slog.Debug("CSS processed successfully", "output", outputFile)
	return nil
}

// findTailwindCLI finds the tailwindcss CLI binary
// First checks system PATH, then checks local bin/tailwindcss
func findTailwindCLI() string {
	// Check if tailwindcss is in PATH
	if path, err := exec.LookPath("tailwindcss"); err == nil {
		return path
	}

	// Check for local bin/tailwindcss
	localPath := "bin/tailwindcss"
	if _, err := os.Stat(localPath); err == nil {
		absPath, err := filepath.Abs(localPath)
		if err == nil {
			return absPath
		}
		return localPath
	}

	return ""
}
