package assets

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// processJS processes JavaScript files with optional minification
func (p *Pipeline) processJS(ctx context.Context) error {
	// Define input and output paths
	inputFile := filepath.Join(p.config.InputDir, "js", "app.js")
	outputDir := filepath.Join(p.config.OutputDir, "js")
	outputFile := filepath.Join(outputDir, "main.js")

	// Check if input file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		slog.Debug("no JS input file found, skipping", "path", inputFile)
		return nil
	}

	// Read input file
	input, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("reading JS file: %w", err)
	}

	slog.Info("processing JavaScript", "input", inputFile, "output", outputFile, "minify", p.config.Minify)

	// Process the JavaScript
	output := string(input)

	// Minify if enabled
	if p.config.Minify {
		minified, err := p.minifier.String("application/javascript", output)
		if err != nil {
			slog.Warn("failed to minify JavaScript, using unminified version", "error", err)
		} else {
			output = minified
			slog.Debug("JavaScript minified successfully")
		}
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Write output file
	if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
		return fmt.Errorf("writing JS file: %w", err)
	}

	slog.Debug("JavaScript processed successfully", "output", outputFile, "size", len(output))
	return nil
}

// processJSDirectory processes all JavaScript files in a directory
func (p *Pipeline) processJSDirectory(ctx context.Context, inputDir, outputDir string) error {
	// Walk the input directory
	return filepath.WalkDir(inputDir, func(path string, d os.DirEntry, err error) error {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process .js files
		if !strings.HasSuffix(path, ".js") {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(inputDir, path)
		if err != nil {
			return fmt.Errorf("calculating relative path: %w", err)
		}

		// Calculate output path
		outPath := filepath.Join(outputDir, relPath)

		// Read input file
		input, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		output := string(input)

		// Minify if enabled
		if p.config.Minify {
			minified, err := p.minifier.String("application/javascript", output)
			if err != nil {
				slog.Warn("failed to minify JavaScript", "file", path, "error", err)
			} else {
				output = minified
			}
		}

		// Create output directory
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}

		// Write output file
		if err := os.WriteFile(outPath, []byte(output), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", outPath, err)
		}

		slog.Debug("processed JavaScript file", "input", path, "output", outPath)
		return nil
	})
}
