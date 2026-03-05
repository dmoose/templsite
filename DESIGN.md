# templsite: Go-native Static Site Generator

## Theory

**templsite** represents a fundamental shift from traditional static site generators toward type-safe, component-driven web development. Instead of fighting with template languages and complex build chains, we embrace Go's type system through `templ` components that compile to native Go functions. This eliminates runtime template errors and provides IDE support with autocomplete, refactoring, and compile-time validation.

The architecture deliberately avoids Node.js dependencies by leveraging Tailwind CSS v4's standalone CLI with native DaisyUI 5 plugin support. This creates a pure Go toolchain where CSS processing happens through the official Tailwind binary using `@plugin` directives, eliminating the JavaScript ecosystem's complexity while maintaining full modern CSS capabilities.

Component reusability is achieved through integration with `templ-store`, allowing developers to share and version templ components across projects. This creates a sustainable ecosystem where UI components have proper provenance tracking and can be updated safely across multiple sites.

The development experience prioritizes immediate feedback through file watching, live reload, and instant CSS compilation. Changes to templ components trigger Go compilation and server restart, while CSS modifications are processed by Tailwind's watch mode, creating a tight feedback loop without build complexity.

Asset management follows modern web standards with CSS variables for theming, automatic minification through pure Go libraries, and cache-busting for production deployments. The pipeline is designed to be simple yet complete, handling the 90% use case without configuration overhead.

Content management remains familiar through Markdown with frontmatter, but page rendering uses type-safe templ components instead of error-prone template strings. This bridges the gap between content-focused workflows and component-driven development.

The project structure follows Go conventions with `pkg/` for reusable libraries, `internal/` for implementation details, and `cmd/` for the CLI application. This makes templsite both a standalone tool and a library that other Go applications can embed.

Configuration is minimal by design, favoring convention over configuration. Default behaviors handle common use cases, while YAML configuration provides escape hatches for customization without overwhelming new users.

The build output is standard static files deployable anywhere, but the development process leverages Go's strengths: fast compilation, excellent tooling, and a robust standard library that eliminates most external dependencies.

This approach creates a sustainable alternative to JavaScript-heavy static site generators while providing modern development ergonomics and type safety that scales from personal blogs to complex documentation sites.

## Implementation

### Project Structure
```
templsite/
├── cmd/templsite/
│   ├── main.go
│   ├── commands/
│   │   ├── build.go
│   │   ├── serve.go
│   │   ├── new.go
│   │   └── components.go
│   └── templates/
│       ├── embed.go
│       ├── minimal/
│       ├── business/
│       └── blog/
├── pkg/
│   ├── site/
│   │   ├── site.go
│   │   ├── page.go
│   │   └── config.go
│   ├── content/
│   │   ├── parser.go
│   │   └── markdown.go
│   ├── assets/
│   │   ├── pipeline.go
│   │   └── css.go
│   └── components/
│       ├── registry.go
│       └── renderer.go
├── internal/
│   ├── server/
│   │   ├── server.go
│   │   └── livereload.go
│   ├── watch/
│   │   └── watcher.go
│   └── build/
│       └── builder.go
├── go.mod
├── Makefile
└── README.md
```

### Core Implementation Files

**cmd/templsite/main.go**
```go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourorg/templsite/cmd/templsite/commands"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), 
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, os.Args[1:]); err != nil {
		slog.Error("command failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: templsite <command> [args]")
	}

	switch args[0] {
	case "new":
		return commands.New(ctx, args[1:])
	case "serve":
		return commands.Serve(ctx, args[1:])
	case "build":
		return commands.Build(ctx, args[1:])
	case "components":
		return commands.Components(ctx, args[1:])
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}
```

**cmd/templsite/templates/embed.go**
```go
package templates

import (
	"embed"
	"io/fs"
)

//go:embed minimal business blog
var Templates embed.FS

func GetTemplate(name string) (fs.FS, error) {
	return fs.Sub(Templates, name)
}

func ListTemplates() []string {
	return []string{"minimal", "business", "blog"}
}
```

**pkg/site/site.go**
```go
package site

import (
	"context"
	"fmt"
	"time"

	"github.com/yourorg/templsite/pkg/content"
	"github.com/yourorg/templsite/pkg/assets"
)

type Site struct {
	Config    *Config
	Pages     []*content.Page
	Assets    *assets.Pipeline
	BuildTime time.Time
}

type Config struct {
	Title      string         `yaml:"title"`
	BaseURL    string         `yaml:"baseURL"`
	ContentDir string         `yaml:"contentDir"`
	OutputDir  string         `yaml:"outputDir"`
	Assets     *assets.Config `yaml:"assets"`
}

func (s *Site) Build(ctx context.Context) error {
	s.BuildTime = time.Now()
	
	if err := s.processContent(ctx); err != nil {
		return fmt.Errorf("processing content: %w", err)
	}
	
	if err := s.Assets.Build(ctx); err != nil {
		return fmt.Errorf("building assets: %w", err)
	}
	
	return s.renderPages(ctx)
}

func (s *Site) processContent(ctx context.Context) error {
	parser := content.NewParser(s.Config.ContentDir)
	pages, err := parser.ParseAll(ctx)
	if err != nil {
		return err
	}
	s.Pages = pages
	return nil
}

func (s *Site) renderPages(ctx context.Context) error {
	for _, page := range s.Pages {
		if err := s.renderPage(ctx, page); err != nil {
			return fmt.Errorf("rendering %s: %w", page.Path, err)
		}
	}
	return nil
}

func (s *Site) renderPage(ctx context.Context, page *content.Page) error {
	// Render page using templ components
	// Implementation depends on specific layout components
	return nil
}
```

**pkg/content/parser.go**
```go
package content

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v3"
)

type Parser struct {
	contentDir string
	markdown   goldmark.Markdown
}

type Page struct {
	Path        string
	Content     string
	Frontmatter map[string]any
	Layout      string
	URL         string
	Date        time.Time
	Draft       bool
}

func NewParser(contentDir string) *Parser {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithHardWraps(), html.WithXHTML()),
	)
	
	return &Parser{
		contentDir: contentDir,
		markdown:   md,
	}
}

func (p *Parser) ParseAll(ctx context.Context) ([]*Page, error) {
	var pages []*Page
	
	err := filepath.WalkDir(p.contentDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || !strings.HasSuffix(path, ".md") {
			return err
		}
		
		page, err := p.ParseFile(ctx, path)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}
		
		pages = append(pages, page)
		return nil
	})
	
	return pages, err
}

func (p *Parser) ParseFile(ctx context.Context, path string) (*Page, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	frontmatter, body, err := p.parseFrontmatter(content)
	if err != nil {
		return nil, err
	}
	
	var buf bytes.Buffer
	if err := p.markdown.Convert(body, &buf); err != nil {
		return nil, err
	}
	
	page := &Page{
		Path:        path,
		Content:     buf.String(),
		Frontmatter: frontmatter,
		Layout:      getStringDefault(frontmatter, "layout", "page"),
		Draft:       getBoolDefault(frontmatter, "draft", false),
		URL:         p.generateURL(path),
	}
	
	if dateStr, ok := frontmatter["date"].(string); ok {
		if date, err := time.Parse("2006-01-02", dateStr); err == nil {
			page.Date = date
		}
	}
	
	return page, nil
}

func (p *Parser) parseFrontmatter(content []byte) (map[string]any, []byte, error) {
	if !bytes.HasPrefix(content, []byte("---\n")) {
		return nil, content, nil
	}
	
	parts := bytes.SplitN(content[4:], []byte("\n---\n"), 2)
	if len(parts) != 2 {
		return nil, content, nil
	}
	
	var frontmatter map[string]any
	if err := yaml.Unmarshal(parts[0], &frontmatter); err != nil {
		return nil, nil, fmt.Errorf("parsing frontmatter: %w", err)
	}
	
	return frontmatter, parts[1], nil
}

func (p *Parser) generateURL(path string) string {
	rel, _ := filepath.Rel(p.contentDir, path)
	rel = strings.TrimSuffix(rel, ".md")
	if rel == "index" {
		return "/"
	}
	return "/" + rel + "/"
}

func getStringDefault(m map[string]any, key, def string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return def
}

func getBoolDefault(m map[string]any, key string, def bool) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return def
}
```

**pkg/assets/pipeline.go**
```go
package assets

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
)

type Pipeline struct {
	config   *Config
	minifier *minify.M
}

type Config struct {
	InputDir    string `yaml:"inputDir"`
	OutputDir   string `yaml:"outputDir"`
	Minify      bool   `yaml:"minify"`
	Fingerprint bool   `yaml:"fingerprint"`
}

func New(config *Config) *Pipeline {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("application/javascript", js.Minify)

	return &Pipeline{
		config:   config,
		minifier: m,
	}
}

func (p *Pipeline) Build(ctx context.Context) error {
	if err := p.processCSS(ctx); err != nil {
		return fmt.Errorf("processing CSS: %w", err)
	}
	
	if err := p.processJS(ctx); err != nil {
		return fmt.Errorf("processing JS: %w", err)
	}
	
	return nil
}

func (p *Pipeline) processCSS(ctx context.Context) error {
	inputFile := filepath.Join(p.config.InputDir, "css", "app.css")
	outputFile := filepath.Join(p.config.OutputDir, "css", "main.css")

	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return err
	}

	args := []string{"-i", inputFile, "-o", outputFile}
	if p.config.Minify {
		args = append(args, "--minify")
	}

	cmd := exec.CommandContext(ctx, "tailwindcss", args...)
	return cmd.Run()
}

func (p *Pipeline) processJS(ctx context.Context) error {
	inputFile := filepath.Join(p.config.InputDir, "js", "app.js")
	outputFile := filepath.Join(p.config.OutputDir, "js", "main.js")

	input, err := os.ReadFile(inputFile)
	if err != nil {
		return err
	}

	result := string(input)
	if p.config.Minify {
		if minified, err := p.minifier.String("application/javascript", result); err == nil {
			result = minified
		}
	}

	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return err
	}

	return os.WriteFile(outputFile, []byte(result), 0644)
}
```

**internal/server/server.go**
```go
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/yourorg/templsite/pkg/site"
	"github.com/yourorg/templsite/internal/watch"
)

type Server struct {
	site    *site.Site
	watcher *watch.Watcher
	addr    string
}

func New(s *site.Site, addr string) *Server {
	return &Server{
		site: s,
		addr: addr,
	}
}

func (s *Server) Serve(ctx context.Context) error {
	watcher, err := watch.New()
	if err != nil {
		return fmt.Errorf("creating watcher: %w", err)
	}
	s.watcher = watcher

	watchPaths := []string{"content", "components", "assets", "config.yaml"}
	for _, path := range watchPaths {
		if err := watcher.Add(path); err != nil {
			slog.Warn("failed to watch path", "path", path, "error", err)
		}
	}

	go s.handleFileChanges(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRequest)
	mux.HandleFunc("/_live-reload", s.handleLiveReload)

	server := &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	slog.Info("starting development server", "addr", s.addr)

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return server.Shutdown(shutdownCtx)
}

func (s *Server) handleFileChanges(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-s.watcher.Events():
			slog.Info("file changed", "path", event.Path)
			
			if err := s.site.Build(ctx); err != nil {
				slog.Error("rebuild failed", "error", err)
				continue
			}
			
			s.notifyReload()
		}
	}
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Serve static files from output directory
	http.FileServer(http.Dir(s.site.Config.OutputDir)).ServeHTTP(w, r)
}

func (s *Server) handleLiveReload(w http.ResponseWriter, r *http.Request) {
	// WebSocket live reload implementation
}

func (s *Server) notifyReload() {
	// Notify connected browsers to reload
}
```

**Makefile**
```makefile
.PHONY: help build test clean install dev setup

GOCMD=go
BINARY_NAME=templsite
BINARY_PATH=./bin/$(BINARY_NAME)

# Tailwind CSS standalone CLI
TAILWIND_VERSION=v4.0.0
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

ifeq ($(UNAME_S),Linux)
    TAILWIND_OS=linux
endif
ifeq ($(UNAME_S),Darwin)
    TAILWIND_OS=macos
endif

ifeq ($(UNAME_M),x86_64)
    TAILWIND_ARCH=x64
endif
ifeq ($(UNAME_M),arm64)
    TAILWIND_ARCH=arm64
endif

TAILWIND_URL=https://github.com/tailwindlabs/tailwindcss/releases/download/$(TAILWIND_VERSION)/tailwindcss-$(TAILWIND_OS)-$(TAILWIND_ARCH)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

deps: ## Download Go dependencies
	$(GOCMD) mod download
	$(GOCMD) mod tidy

setup-tailwind: ## Download Tailwind CSS standalone CLI
	@echo "Downloading Tailwind CSS v4 standalone CLI..."
	mkdir -p bin
	curl -sL $(TAILWIND_URL) -o bin/tailwindcss
	chmod +x bin/tailwindcss
	@echo "Tailwind CSS CLI installed to bin/tailwindcss"

setup: deps setup-tailwind ## Complete setup

build: deps ## Build the binary
	mkdir -p bin
	$(GOCMD) build -o $(BINARY_PATH) ./cmd/templsite

test: ## Run tests
	$(GOCMD) test -v -race ./...

clean: ## Clean build artifacts
	$(GOCMD) clean
	rm -rf bin/

install: build ## Install binary
	$(GOCMD) install ./cmd/templsite

dev: build setup-tailwind ## Start development server
	./$(BINARY_PATH) serve --watch

example: build setup-tailwind ## Create example site
	mkdir -p example
	cd example && ../$(BINARY_PATH) new . --template business
	cd example && ../$(BINARY_PATH) serve
```

## README.md

# templsite

A modern static site generator built with Go and templ components. Zero Node.js dependencies, type-safe templates, and modern CSS with Tailwind + DaisyUI.

## Features

- **Type-safe templates**: Use Go's templ components instead of error-prone template strings
- **Zero Node.js**: Pure Go toolchain with Tailwind CSS standalone CLI
- **Live reload**: Instant feedback during development
- **Modern CSS**: Tailwind CSS v4 + DaisyUI 5 with CSS variables
- **Component library**: Integration with templ-store for reusable components

## Installation

### From releases
```bash
curl -L https://github.com/yourorg/templsite/releases/latest/download/templsite_linux_amd64.tar.gz | tar xz
sudo mv templsite /usr/local/bin/
```

### From source (requires Go 1.25+)
```bash
git clone https://github.com/yourorg/templsite
cd templsite
make setup
make install
```

## Quick start

### Create a new site
```bash
templsite new mysite --template business
cd mysite
templsite serve
```

Your site will be available at http://localhost:8080 with live reload.

### Project structure
```
mysite/
├── content/           # Markdown content
├── components/        # templ components
├── assets/           # CSS, JS, images
├── config.yaml       # Site configuration
└── public/           # Generated output
```

## Usage

### Commands

**Create new site**
```bash
templsite new <path> [--template <name>]
```

Available templates: `minimal`, `business`, `blog`

**Development server**
```bash
templsite serve [--port 8080] [--watch]
```

**Build for production**
```bash
templsite build [--output public]
```

**Component management**
```bash
templsite components add cc:ui/hero
templsite components list
templsite components update
```

### Configuration

**config.yaml**
```yaml
title: "My Site"
baseURL: "https://example.com"

content:
  dir: "content"
  defaultLayout: "page"

assets:
  inputDir: "assets"
  outputDir: "public/assets"
  minify: true
  fingerprint: true
```

### Content

**Frontmatter example**
```markdown
---
title: "My Post"
date: 2025-01-15
layout: "post"
---

# My Post Content

Regular markdown content here.
```

### Components

**Layout component (components/layout/base.templ)**
```go
package layout

templ Base(title string, content templ.Component) {
    <!DOCTYPE html>
    <html data-theme="light">
    <head>
        <title>{title}</title>
        <link rel="stylesheet" href="/assets/css/main.css">
    </head>
    <body class="min-h-screen bg-base-100">
        {content}
    </body>
    </html>
}
```

**UI component (components/ui/hero.templ)**
```go
package ui

templ Hero(title, subtitle string) {
    <div class="hero min-h-screen bg-base-200">
        <div class="hero-content text-center">
            <h1 class="text-5xl font-bold text-primary">{title}</h1>
            <p class="py-6">{subtitle}</p>
            <button class="btn btn-primary">Get Started</button>
        </div>
    </div>
}
```

### Theming

**assets/css/app.css**
```css
@import "tailwindcss";
@plugin "daisyui";

@theme {
  --color-brand-500: oklch(0.6 0.15 210);
}

:root {
  --color-primary: var(--color-brand-500);
}
```

## Development

### Prerequisites
- Go 1.25+
- No Node.js required

### Setup
```bash
make setup  # Downloads Go deps + Tailwind CLI
make dev    # Start development server
```

### Asset processing
- **CSS**: Tailwind CSS v4 standalone CLI with DaisyUI plugin
- **JS**: Pure Go minification with tdewolff/minify
- **Images**: Automatic optimization and fingerprinting

## Deployment

### Static hosting
```bash
templsite build
# Upload public/ directory
```

### Docker
```dockerfile
FROM alpine:latest
COPY templsite /usr/local/bin/
COPY public/ /var/www/
EXPOSE 8080
CMD ["templsite", "serve", "--addr", ":8080", "--static", "/var/www"]
```

## License

MIT License
