package assets

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// processCSS processes CSS files with Tailwind CLI, or copies them as-is
func (p *Pipeline) processCSS(ctx context.Context) error {
	inputDir := filepath.Join(p.config.InputDir, "css")
	outputDir := filepath.Join(p.config.OutputDir, "css")

	// Check if CSS input directory exists at all
	if _, err := os.Stat(inputDir); os.IsNotExist(err) {
		slog.Debug("no CSS directory found, skipping", "path", inputDir)
		return nil
	}

	// If app.css exists, process through Tailwind
	inputFile := filepath.Join(inputDir, "app.css")
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		// No Tailwind entry point — copy CSS files as static assets
		slog.Debug("no app.css found, copying CSS files as-is", "from", inputDir)
		return p.copyCSSStatic(ctx, inputDir, outputDir)
	}

	outputFile := filepath.Join(outputDir, "main.css")

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

// copyCSSStatic copies all CSS files from input to output without processing
func (p *Pipeline) copyCSSStatic(ctx context.Context, inputDir, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating CSS output directory: %w", err)
	}

	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return fmt.Errorf("reading CSS directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".css") {
			continue
		}
		src := filepath.Join(inputDir, entry.Name())
		dst := filepath.Join(outputDir, entry.Name())
		if err := copyFile(src, dst); err != nil {
			return fmt.Errorf("copying %s: %w", entry.Name(), err)
		}
		slog.Debug("copied CSS file", "file", entry.Name())
	}
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
