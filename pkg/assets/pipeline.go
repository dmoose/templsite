// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package assets

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
)

// Pipeline handles asset processing (CSS, JS, static files)
type Pipeline struct {
	config   *Config
	minifier *minify.M
}

// Config configures the asset pipeline
type Config struct {
	InputDir    string
	OutputDir   string
	Minify      bool
	Fingerprint bool
}

// New creates a new asset pipeline
func New(config *Config) *Pipeline {
	// Initialize minifier
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("application/javascript", js.Minify)

	return &Pipeline{
		config:   config,
		minifier: m,
	}
}

// Build processes all assets
func (p *Pipeline) Build(ctx context.Context) error {
	slog.Debug("building assets", "inputDir", p.config.InputDir, "outputDir", p.config.OutputDir)

	// Process CSS
	if err := p.processCSS(ctx); err != nil {
		return fmt.Errorf("processing CSS: %w", err)
	}

	// Process JS
	if err := p.processJS(ctx); err != nil {
		return fmt.Errorf("processing JS: %w", err)
	}

	// Copy static files
	if err := p.copyStatic(ctx); err != nil {
		return fmt.Errorf("copying static files: %w", err)
	}

	slog.Info("assets built successfully")
	return nil
}
