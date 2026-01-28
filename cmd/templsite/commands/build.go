package commands

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"git.catapulsion.com/templsite/pkg/site"
)

// Build builds the site for production
func Build(ctx context.Context, args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	configPath := fs.String("config", "config.yaml", "path to configuration file")
	outputDir := fs.String("output", "", "output directory (overrides config)")
	verbose := fs.Bool("verbose", false, "enable verbose logging")
	clean := fs.Bool("clean", false, "clean output directory before build")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: templsite build [options]

Build the site for production. Processes content, assets, and renders all pages.

Options:
`)
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  templsite build
  templsite build --config site.yaml
  templsite build --output dist --verbose
  templsite build --clean

`)
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Setup logging level
	logLevel := slog.LevelInfo
	if *verbose {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	slog.Debug("build command started", "config", *configPath, "verbose", *verbose)

	// Check if config file exists
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s\n\nRun 'templsite new <sitename>' to create a new site", *configPath)
	}

	// Load site configuration
	slog.Info("loading configuration", "path", *configPath)
	s, err := site.New(*configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override output directory if specified
	if *outputDir != "" {
		slog.Debug("overriding output directory", "dir", *outputDir)
		s.Config.OutputDir = *outputDir
	}

	// Get absolute paths for reporting
	absConfigPath, _ := filepath.Abs(*configPath)
	absOutputPath, _ := filepath.Abs(s.Config.OutputPath("."))

	slog.Info("build configuration",
		"site", s.Config.Title,
		"config", absConfigPath,
		"output", absOutputPath,
	)

	// Clean output directory if requested
	if *clean {
		slog.Info("cleaning output directory", "dir", absOutputPath)
		if err := os.RemoveAll(absOutputPath); err != nil {
			slog.Warn("failed to clean output directory", "error", err)
		}
	}

	// Build the site
	startTime := time.Now()
	slog.Info("starting build", "time", startTime.Format(time.RFC3339))

	if err := s.Build(ctx); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	elapsed := time.Since(startTime)

	// Report build statistics
	stats := getBuildStats(s, absOutputPath)

	slog.Info("build complete",
		"duration", elapsed,
		"pages", stats.Pages,
		"assets", stats.Assets,
		"totalFiles", stats.TotalFiles,
		"totalSize", formatBytes(stats.TotalSize),
	)

	// Print success message
	fmt.Println()
	fmt.Println("✓ Build successful!")
	fmt.Printf("  Duration: %v\n", elapsed.Round(time.Millisecond))
	fmt.Printf("  Pages: %d\n", stats.Pages)
	fmt.Printf("  Assets: %d\n", stats.Assets)
	fmt.Printf("  Total files: %d\n", stats.TotalFiles)
	fmt.Printf("  Total size: %s\n", formatBytes(stats.TotalSize))
	fmt.Printf("  Output: %s\n", absOutputPath)
	fmt.Println()

	return nil
}

// BuildStats holds build statistics
type BuildStats struct {
	Pages      int
	Assets     int
	TotalFiles int
	TotalSize  int64
}

// getBuildStats collects statistics about the build output
func getBuildStats(s *site.Site, outputPath string) BuildStats {
	stats := BuildStats{
		Pages: len(s.Pages),
	}

	// Walk output directory to count files and sizes
	filepath.Walk(outputPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			return nil
		}

		stats.TotalFiles++
		stats.TotalSize += info.Size()

		// Count assets (anything not .html)
		if filepath.Ext(path) != ".html" {
			stats.Assets++
		}

		return nil
	})

	return stats
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
