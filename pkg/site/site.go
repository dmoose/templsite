package site

import (
	"context"
	"fmt"
	"log/slog"
	"time"

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
	// TODO: Implement in Stage 4-5
	slog.Debug("building assets", "inputDir", s.Config.Assets.InputDir)
	return nil
}

// renderPages renders all pages (Stage 6)
func (s *Site) renderPages(ctx context.Context) error {
	// TODO: Implement in Stage 6
	slog.Debug("rendering pages")
	return nil
}

// Clean removes the output directory
func (s *Site) Clean() error {
	// TODO: Implement when needed
	slog.Info("cleaning output directory", "dir", s.Config.OutputDir)
	return nil
}
