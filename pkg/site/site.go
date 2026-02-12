// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package site

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"

	"github.com/dmoose/templsite/pkg/assets"
	"github.com/dmoose/templsite/pkg/content"
)

// Site represents a static site with all its content and configuration
type Site struct {
	Config       *Config
	Pages        []*content.Page
	Sections     map[string]*Section
	Taxonomies   map[string]*Taxonomy
	Menus        map[string][]*MenuItem
	Data         map[string]any
	BuildTime    time.Time
	AssetVersion string // content hash for cache-busting query strings
	baseDir      string
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

// NewWithEnv creates a new Site, merging an environment-specific config override.
// For example, NewWithEnv("config.yaml", "production") loads config.yaml then
// merges config.production.yaml on top. If env is empty, behaves like New().
func NewWithEnv(configPath string, env string) (*Site, error) {
	config, err := LoadConfigWithEnv(configPath, env)
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

	// Generate templ components first
	if err := s.GenerateTemplComponents(ctx); err != nil {
		return fmt.Errorf("generating templ components: %w", err)
	}

	// Process content
	if err := s.ProcessContent(ctx); err != nil {
		return fmt.Errorf("processing content: %w", err)
	}

	// Build assets
	if err := s.BuildAssets(ctx); err != nil {
		return fmt.Errorf("building assets: %w", err)
	}

	// Generate syntax highlighting CSS if configured
	if err := s.generateHighlightCSS(); err != nil {
		return fmt.Errorf("generating highlight CSS: %w", err)
	}

	// Compute asset version hash for cache-busting
	s.computeAssetVersion()

	// Copy static files to output root (favicons, etc.)
	if err := s.copyStaticFiles(); err != nil {
		return fmt.Errorf("copying static files: %w", err)
	}

	// Write auto-generated files (robots.txt, sitemap.xml, site.webmanifest, feed.xml)
	s.writeGeneratedFiles()

	elapsed := time.Since(s.BuildTime)
	slog.Info("build complete", "duration", elapsed)

	return nil
}

// ProcessContent processes all content files and makes them available via Pages
func (s *Site) ProcessContent(ctx context.Context) error {
	contentPath := s.Config.ContentPath(s.baseDir)
	slog.Debug("processing content", "dir", contentPath)

	// Configure parser with optional syntax highlighting
	var parserOpts []content.HighlightOptions
	if s.Config.Highlight.Style != "" {
		parserOpts = append(parserOpts, content.HighlightOptions{
			Style:       s.Config.Highlight.Style,
			LineNumbers: s.Config.Highlight.LineNumbers,
		})
	}
	parser := content.NewParser(contentPath, parserOpts...)
	pages, err := parser.ParseAll(ctx)
	if err != nil {
		return fmt.Errorf("parsing content: %w", err)
	}

	// Filter pages based on build options
	s.Pages = s.filterPages(pages)

	// Organize pages into sections
	s.buildSections()

	// Build taxonomies from page frontmatter
	s.buildTaxonomies()

	// Build menus from config
	s.buildMenus()

	// Load data files
	if err := s.loadData(); err != nil {
		return fmt.Errorf("loading data: %w", err)
	}

	slog.Info("content processed",
		"pages", len(s.Pages),
		"sections", len(s.Sections),
		"taxonomies", len(s.Taxonomies),
		"menus", len(s.Menus),
		"data", len(s.Data))

	return nil
}

// filterPages filters pages based on build options (drafts, future dates)
func (s *Site) filterPages(pages []*content.Page) []*content.Page {
	var filtered []*content.Page
	now := time.Now()

	for _, page := range pages {
		// Skip drafts unless drafts are enabled
		if page.Draft && !s.Config.Build.Drafts {
			continue
		}

		// Skip future-dated pages unless future is enabled
		if !page.Date.IsZero() && page.Date.After(now) && !s.Config.Build.Future {
			continue
		}

		filtered = append(filtered, page)
	}

	return filtered
}

// buildSections organizes pages into sections and establishes relationships
func (s *Site) buildSections() {
	s.Sections = make(map[string]*Section)

	// Group pages by section
	for _, page := range s.Pages {
		sectionName := page.Section
		if sectionName == "" {
			sectionName = "_root"
		}

		section, exists := s.Sections[sectionName]
		if !exists {
			section = &Section{
				Name: sectionName,
				URL:  s.sectionURL(sectionName),
			}
			s.Sections[sectionName] = section
		}

		// Check if this is an _index.md (section index page)
		if s.isIndexPage(page) {
			section.Index = page
			section.Title = page.Title
			section.Description = page.Description
		}

		section.Pages = append(section.Pages, page)
	}

	// Set default titles for sections without _index.md
	for name, section := range s.Sections {
		if section.Title == "" {
			section.Title = s.titleFromName(name)
		}
	}

	// Sort pages and establish Prev/Next links within each section
	for _, section := range s.Sections {
		section.sortPages()
		section.linkPrevNext()
	}
}

// sectionURL generates the URL for a section
func (s *Site) sectionURL(name string) string {
	if name == "_root" {
		return "/"
	}
	return "/" + name + "/"
}

// isIndexPage checks if a page is a section index (_index.md or index.md at section root)
func (s *Site) isIndexPage(page *content.Page) bool {
	// Check for _index.md pattern in the URL
	// _index.md in content/blog/ becomes /blog/
	// index.md in content/ becomes /
	url := page.URL
	section := page.Section

	if section == "" {
		return url == "/"
	}
	return url == "/"+section+"/"
}

// titleFromName converts a section name to a title (e.g., "blog" -> "Blog")
func (s *Site) titleFromName(name string) string {
	if name == "_root" {
		return "Home"
	}
	if len(name) == 0 {
		return ""
	}
	// Capitalize first letter
	return string(name[0]-32) + name[1:]
}

// buildTaxonomies creates taxonomies from page frontmatter
func (s *Site) buildTaxonomies() {
	s.Taxonomies = make(map[string]*Taxonomy)

	// Initialize configured taxonomies
	for _, name := range s.Config.Taxonomies {
		s.Taxonomies[name] = NewTaxonomy(name)
	}

	// Process each page's frontmatter for taxonomy terms
	for _, page := range s.Pages {
		if !page.IsPublished() {
			continue
		}

		// Check each configured taxonomy
		for taxName, taxonomy := range s.Taxonomies {
			// Get terms from frontmatter (e.g., page.Frontmatter["tags"])
			terms := s.getTermsFromPage(page, taxName)
			for _, term := range terms {
				taxonomy.AddPage(term, page)
			}
		}
	}

	// Sort pages within each term by date
	for _, taxonomy := range s.Taxonomies {
		taxonomy.SortTermPages()
	}
}

// getTermsFromPage extracts taxonomy terms from a page's frontmatter
func (s *Site) getTermsFromPage(page *content.Page, taxonomy string) []string {
	// Special case: "tags" is stored in Page.Tags
	if taxonomy == "tags" {
		return page.Tags
	}

	// Otherwise, look in frontmatter
	val, ok := page.Frontmatter[taxonomy]
	if !ok {
		return nil
	}

	// Handle different types
	switch v := val.(type) {
	case []string:
		return v
	case []any:
		var terms []string
		for _, item := range v {
			if str, ok := item.(string); ok {
				terms = append(terms, str)
			}
		}
		return terms
	case string:
		// Single value as string
		return []string{v}
	}

	return nil
}

// BuildAssets builds all assets (CSS, JS, static files)
func (s *Site) BuildAssets(ctx context.Context) error {
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

// OutputDir returns the absolute path to the output directory
// This is a helper for code that needs the output directory path
func (s *Site) OutputDir() string {
	return s.Config.OutputPath(s.baseDir)
}

// GetOutputPath converts a URL path to a filesystem output path
// This is a helper for user's rendering code
func (s *Site) GetOutputPath(url string) string {
	outputDir := s.OutputDir()

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

	// Special pages that static hosts expect at the root as .html files
	if url == "404" {
		return filepath.Join(outputDir, "404.html")
	}

	// All other URLs become directory/index.html for clean URLs
	return filepath.Join(outputDir, url, "index.html")
}

// Clean removes the output directory
func (s *Site) Clean() error {
	outputDir := s.Config.OutputPath(s.baseDir)
	slog.Info("cleaning output directory", "dir", outputDir)

	// Check if output directory exists
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		slog.Debug("output directory does not exist, nothing to clean")
		return nil
	}

	// Remove the output directory and all contents
	if err := os.RemoveAll(outputDir); err != nil {
		return fmt.Errorf("removing output directory: %w", err)
	}

	slog.Debug("output directory cleaned successfully")
	return nil
}

// copyStaticFiles copies the contents of the static directory to the output root.
// This allows placing files like favicons at the site root. Skips if dir doesn't exist.
func (s *Site) copyStaticFiles() error {
	staticPath := s.Config.StaticPath(s.baseDir)
	outputDir := s.OutputDir()

	info, err := os.Stat(staticPath)
	if err != nil || !info.IsDir() {
		slog.Debug("no static directory found, skipping", "path", staticPath)
		return nil
	}

	count := 0
	err = filepath.WalkDir(staticPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(staticPath, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(outputDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		src, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("opening %s: %w", path, err)
		}
		defer func() { _ = src.Close() }()

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		dst, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("creating %s: %w", destPath, err)
		}
		defer func() { _ = dst.Close() }()

		if _, err := io.Copy(dst, src); err != nil {
			return fmt.Errorf("copying %s: %w", relPath, err)
		}

		count++
		return nil
	})

	if count > 0 {
		slog.Info("static files copied", "count", count)
	}
	return err
}

// computeAssetVersion walks the output assets directory and computes a SHA256 hash
// of all file contents, storing the first 8 hex chars as the asset version.
func (s *Site) computeAssetVersion() {
	assetsOutputPath := s.Config.AssetsOutputPath(s.baseDir)

	h := sha256.New()
	_ = filepath.WalkDir(assetsOutputPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil // skip unreadable files
		}
		h.Write(data)
		return nil
	})

	s.AssetVersion = fmt.Sprintf("%x", h.Sum(nil))[:8]
	slog.Debug("asset version computed", "version", s.AssetVersion)
}

// writeGeneratedFiles writes robots.txt, sitemap.xml, site.webmanifest, and feed.xml
// to the output directory, but only if the file doesn't already exist
// (user can override via static/).
func (s *Site) writeGeneratedFiles() {
	outputDir := s.OutputDir()

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		slog.Warn("failed to create output directory for generated files", "error", err)
		return
	}

	writeIfMissing := func(name, content string) {
		path := filepath.Join(outputDir, name)
		if _, err := os.Stat(path); err == nil {
			slog.Debug("skipping generated file (already exists)", "file", name)
			return
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			slog.Warn("failed to write generated file", "file", name, "error", err)
			return
		}
		slog.Debug("generated file written", "file", name)
	}

	writeIfMissing("robots.txt", s.RobotsTxt())
	writeIfMissing("sitemap.xml", s.Sitemap())
	writeIfMissing("site.webmanifest", s.Manifest())
	writeIfMissing("feed.xml", s.Feed())
	writeIfMissing("404.html", s.Default404())
}

// GenerateTemplComponents runs templ generate on the components directory
func (s *Site) GenerateTemplComponents(ctx context.Context) error {
	componentsPath := filepath.Join(s.baseDir, "components")

	// Check if components directory exists
	if _, err := os.Stat(componentsPath); os.IsNotExist(err) {
		slog.Debug("no components directory found, skipping templ generate")
		return nil
	}

	slog.Debug("generating templ components", "path", componentsPath)

	cmd := exec.CommandContext(ctx, "templ", "generate", "-path", componentsPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = s.baseDir

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running templ generate: %w", err)
	}

	slog.Debug("templ components generated successfully")
	return nil
}

// generateHighlightCSS writes a chroma.css file to the assets output directory
// when syntax highlighting is configured. The CSS provides colors for the
// Chroma token classes emitted by goldmark-highlighting.
func (s *Site) generateHighlightCSS() error {
	if s.Config.Highlight.Style == "" {
		return nil
	}

	style := styles.Get(s.Config.Highlight.Style)

	assetsOutputPath := s.Config.AssetsOutputPath(s.baseDir)
	cssDir := filepath.Join(assetsOutputPath, "css")
	if err := os.MkdirAll(cssDir, 0755); err != nil {
		return fmt.Errorf("creating CSS directory: %w", err)
	}

	cssPath := filepath.Join(cssDir, "chroma.css")
	f, err := os.Create(cssPath)
	if err != nil {
		return fmt.Errorf("creating chroma.css: %w", err)
	}
	defer func() { _ = f.Close() }()

	formatter := chromahtml.New(chromahtml.WithClasses(true))
	if err := formatter.WriteCSS(f, style); err != nil {
		return fmt.Errorf("writing chroma CSS: %w", err)
	}

	slog.Info("syntax highlighting CSS generated", "style", s.Config.Highlight.Style, "path", cssPath)
	return nil
}
