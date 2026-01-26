package site

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"git.catapulsion.com/templsite/components/layout"
	"git.catapulsion.com/templsite/pkg/assets"
	"git.catapulsion.com/templsite/pkg/content"
)

// Site represents a static site with all its content and configuration
type Site struct {
	Config    *Config
	Pages     []*content.Page
	BuildTime time.Time
	baseDir   string
}

// New creates a new Site instance
func New(configPath string) (*Site, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	return &Site{
		Config:  config,
		baseDir: ".",
	}, nil
}

// NewWithConfig creates a new Site with an existing configuration
func NewWithConfig(config *Config) *Site {
	return &Site{
		Config:  config,
		baseDir: ".",
	}
}

// SetBaseDir sets the base directory for the site
func (s *Site) SetBaseDir(dir string) {
	s.baseDir = dir
}

// Build builds the entire site
func (s *Site) Build(ctx context.Context) error {
	s.BuildTime = time.Now()

	slog.Info("building site", "title", s.Config.Title)

	// TODO: Stage 3 - Process content
	if err := s.processContent(ctx); err != nil {
		return fmt.Errorf("processing content: %w", err)
	}

	// TODO: Stage 4-5 - Build assets
	if err := s.buildAssets(ctx); err != nil {
		return fmt.Errorf("building assets: %w", err)
	}

	// TODO: Stage 6 - Render pages
	if err := s.renderPages(ctx); err != nil {
		return fmt.Errorf("rendering pages: %w", err)
	}

	elapsed := time.Since(s.BuildTime)
	slog.Info("build complete", "duration", elapsed)

	return nil
}

// processContent processes all content files (Stage 3)
func (s *Site) processContent(ctx context.Context) error {
	contentPath := s.Config.ContentPath(s.baseDir)
	slog.Debug("processing content", "dir", contentPath)

	parser := content.NewParser(contentPath)
	pages, err := parser.ParseAll(ctx)
	if err != nil {
		return fmt.Errorf("parsing content: %w", err)
	}

	s.Pages = pages
	slog.Info("content processed", "pages", len(pages))

	return nil
}

// buildAssets builds all assets (Stage 4-5)
func (s *Site) buildAssets(ctx context.Context) error {
	assetsInputPath := s.Config.AssetsInputPath(s.baseDir)
	assetsOutputPath := s.Config.AssetsOutputPath(s.baseDir)

	slog.Debug("building assets", "inputDir", assetsInputPath, "outputDir", assetsOutputPath)

	// Create asset pipeline config
	pipelineConfig := &assets.Config{
		InputDir:    assetsInputPath,
		OutputDir:   assetsOutputPath,
		Minify:      s.Config.Assets.Minify,
		Fingerprint: s.Config.Assets.Fingerprint,
	}

	// Create and run pipeline
	pipeline := assets.New(pipelineConfig)
	if err := pipeline.Build(ctx); err != nil {
		return fmt.Errorf("building assets: %w", err)
	}

	return nil
}

// renderPages renders all pages (Stage 6)
func (s *Site) renderPages(ctx context.Context) error {
	slog.Debug("rendering pages", "count", len(s.Pages))

	for _, page := range s.Pages {
		if err := s.renderPage(ctx, page); err != nil {
			return fmt.Errorf("rendering %s: %w", page.Path, err)
		}
	}

	slog.Info("pages rendered", "count", len(s.Pages))
	return nil
}

// renderPage renders a single page using the appropriate templ component
func (s *Site) renderPage(ctx context.Context, page *content.Page) error {
	slog.Debug("rendering page", "path", page.Path, "url", page.URL, "layout", page.Layout)

	// Determine output path from URL
	outputPath := s.getOutputPath(page.URL)
	slog.Debug("output path determined", "url", page.URL, "outputPath", outputPath)

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory %s: %w", outputDir, err)
	}

	// Create output file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output file %s: %w", outputPath, err)
	}
	defer f.Close()

	// Select and render layout
	component := layout.Page(s.Config.Title, s.Config.BaseURL, page)
	if err := component.Render(ctx, f); err != nil {
		return fmt.Errorf("rendering component: %w", err)
	}

	slog.Debug("page rendered successfully", "path", page.Path, "output", outputPath)
	return nil
}

// getOutputPath converts a URL path to a filesystem output path
func (s *Site) getOutputPath(url string) string {
	outputDir := s.Config.OutputPath(s.baseDir)

	// Remove leading slash
	if len(url) > 0 && url[0] == '/' {
		url = url[1:]
	}

	// Remove trailing slash
	if len(url) > 0 && url[len(url)-1] == '/' {
		url = url[:len(url)-1]
	}

	// Root URL becomes index.html
	if url == "" {
		return filepath.Join(outputDir, "index.html")
	}

	// All other URLs become directory/index.html for clean URLs
	return filepath.Join(outputDir, url, "index.html")
}

// Clean removes the output directory
func (s *Site) Clean() error {
	// TODO: Implement when needed
	slog.Info("cleaning output directory", "dir", s.Config.OutputDir)
	return nil
}
