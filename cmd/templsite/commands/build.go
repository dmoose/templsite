package commands

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/dmoose/templsite/pkg/site"
	"github.com/spf13/cobra"
)

// NewBuildCmd creates the "build" command
func NewBuildCmd(ctx context.Context) *cobra.Command {
	var (
		configPath string
		env        string
		outputDir  string
		verbose    bool
		clean      bool
	)

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build the site for production",
		Long:  "Build the site for production. Processes content, assets, and renders all pages.",
		Example: `  templsite build
  templsite build --env production
  templsite build --config site.yaml
  templsite build --output dist --verbose
  templsite build --clean`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuild(ctx, configPath, env, outputDir, verbose, clean)
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "config.yaml", "path to configuration file")
	cmd.Flags().StringVar(&env, "env", "", "environment (loads config.<env>.yaml overrides)")
	cmd.Flags().StringVar(&outputDir, "output", "", "output directory (overrides config)")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "enable verbose logging")
	cmd.Flags().BoolVar(&clean, "clean", false, "clean output directory before build")

	return cmd
}

func runBuild(ctx context.Context, configPath, env, outputDir string, verbose, clean bool) error {
	// Setup logging level
	if verbose {
		logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
		slog.SetDefault(logger)
	}

	slog.Debug("build command started", "config", configPath, "verbose", verbose)

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s\n\nRun 'templsite new <sitename>' to create a new site", configPath)
	}

	// Load site configuration
	slog.Info("loading configuration", "path", configPath, "env", env)
	s, err := site.NewWithEnv(configPath, env)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override output directory if specified
	if outputDir != "" {
		slog.Debug("overriding output directory", "dir", outputDir)
		s.Config.OutputDir = outputDir
	}

	// Get absolute paths for reporting
	absConfigPath, _ := filepath.Abs(configPath)
	absOutputPath, _ := filepath.Abs(s.Config.OutputPath("."))

	slog.Info("build configuration",
		"site", s.Config.Title,
		"config", absConfigPath,
		"output", absOutputPath,
	)

	// Clean output directory if requested
	if clean {
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
	_ = filepath.Walk(outputPath, func(path string, info os.FileInfo, err error) error {
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
