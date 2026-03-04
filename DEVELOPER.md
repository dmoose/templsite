# Developer Guide

This guide is for developers who want to contribute to templsite or understand its internals. For using templsite to build sites, see [README.md](README.md). For architectural decisions, see [ARCHITECTURE.md](ARCHITECTURE.md).

## Project Philosophy

### Type Safety First

Templsite uses Go's type system and templ components to catch errors at compile time, not at runtime. This means:

- **No template strings** - Components are Go functions
- **IDE support** - Autocomplete, refactoring, go-to-definition work
- **Compile-time validation** - Errors caught before deployment
- **Refactoring confidence** - Breaking changes fail at compile time

### Pure Go Toolchain

Zero Node.js dependencies by design:

- **Tailwind CSS v4 standalone CLI** - Binary downloaded by Makefile
- **Go-based minification** - tdewolff/minify for JS
- **No npm/webpack/babel** - Pure Go build pipeline
- **Simple deployment** - Single binary + static files

### Convention Over Configuration

Sensible defaults minimize configuration burden:

- Standard directory structure (`content/`, `assets/`, `components/`)
- Predictable URL generation (`about.md` → `/about/`)
- Default layouts and behaviors
- Configuration only when needed

### Progressive Enhancement

Start with semantic HTML, enhance progressively:

1. **HTML** - Semantic, accessible markup
2. **CSS** - Tailwind utilities + DaisyUI components
3. **JavaScript** - Minimal, purposeful interactivity
4. **Frameworks** - Only if simpler approaches fail

## Architecture Overview

### Component Override Pattern

**Key Decision**: Each user site is a standalone Go project that compiles its own binary.

Why this matters:
- Templ components are Go functions that compile at build time
- User's binary imports **their** components, not templsite's
- Full customization without runtime hacks
- Type safety maintained end-to-end

See [ARCHITECTURE.md](ARCHITECTURE.md) for complete explanation.

### Library vs Application

**templsite** has two roles:

1. **Library** (`pkg/*`): Infrastructure for building sites
   - Content parsing (Markdown → Page structs)
   - Asset pipeline (CSS, JS, static files)
   - Dev server (HTTP, WebSocket, file watching)
   - Configuration management

2. **Application** (`cmd/templsite/*`): CLI tool for site management
   - `new` command - Project scaffolding
   - `components` command - Registry integration
   - Template embedding and extraction

User sites import the library and provide their own rendering logic.

## Project Structure

```
templsite/
├── cmd/templsite/              # CLI application
│   ├── main.go                # Entry point, command router
│   ├── commands/              # Command implementations
│   │   ├── build.go           # Build command (stub - user's binary does this)
│   │   ├── serve.go           # Serve command (stub)
│   │   ├── new.go             # New site scaffolding
│   │   └── components.go      # Component registry (Stage 10)
│   └── templates/             # Embedded site templates
│       ├── embed.go           # go:embed directives
│       └── minimal/           # Minimal template
│
├── pkg/                       # Public library API
│   ├── site/                  # Site management
│   │   ├── site.go           # Site struct, Build() orchestration
│   │   └── config.go         # Configuration loading
│   ├── content/              # Content parsing
│   │   ├── parser.go         # Markdown parser
│   │   └── page.go           # Page struct
│   ├── assets/               # Asset pipeline
│   │   ├── pipeline.go       # Orchestration
│   │   ├── css.go            # Tailwind CSS processing
│   │   ├── js.go             # JavaScript minification
│   │   └── static.go         # Static file copying
│   ├── server/               # Development server
│   │   ├── server.go         # HTTP server
│   │   └── livereload.go     # WebSocket live reload
│   └── components/           # Component registry (Stage 10)
│
├── internal/                  # Internal packages
│   ├── watch/                # File watching
│   │   └── watcher.go        # fsnotify wrapper
│   └── build/                # Build helpers (empty)
│
├── components/                # Example components (for templsite repo itself)
│   ├── layout/               # Layouts
│   └── ui/                   # UI components
│
├── assets/                    # Example assets
│   ├── css/
│   │   └── app.css.example  # Tailwind v4 template
│   └── js/
│       └── app.js.example   # JS utilities template
│
├── go.mod                     # Module definition
├── Makefile                   # Build automation
└── *.md                       # Documentation
```

## Development Setup

### Prerequisites

```bash
# Required
go 1.21+
make
curl (for downloading Tailwind CLI)

# Optional but recommended
git
templ CLI: go install github.com/a-h/templ/cmd/templ@latest
```

### Initial Setup

```bash
# Clone repository
git clone https://github.com/yourorg/templsite
cd templsite

# Download dependencies and Tailwind CLI
make setup

# Build templsite binary
make build

# Run tests
make test

# Generate templ components (if developing components)
make generate
```

### Makefile Targets

```bash
make help           # Show all available targets
make setup          # Complete setup (deps + Tailwind CLI)
make build          # Build the templsite binary
make test           # Run tests with race detector
make generate       # Generate templ components
make clean          # Clean build artifacts
make install        # Install to GOPATH
make fmt            # Format code
make vet            # Run go vet
make lint           # Run golangci-lint (if installed)
```

## Core Concepts

### Site.Build() Pipeline

The build process has three phases:

```go
func (s *Site) Build(ctx context.Context) error {
    // 1. Generate templ components
    s.GenerateTemplComponents(ctx)
    
    // 2. Process content (Markdown → Page structs)
    s.ProcessContent(ctx)
    
    // 3. Build assets (CSS, JS, static)
    s.BuildAssets(ctx)
    
    // Note: Rendering happens in user's binary, not here
    return nil
}
```

**Important**: `Site.Build()` does NOT render HTML. That happens in the user's `main.go` using their components.

### Content Parsing

**Markdown → Page struct**:

```go
// Parse all .md files in content directory
parser := content.NewParser("content")
pages, err := parser.ParseAll(ctx)

// Each page has:
type Page struct {
    Path        string              // File path
    Content     string              // Rendered HTML
    Frontmatter map[string]any      // YAML metadata
    Layout      string              // Layout name
    URL         string              // Generated URL
    Date        time.Time           // Publication date
    Draft       bool                // Draft status
    Title       string              // Page title
    Description string              // Meta description
    Tags        []string            // Tags
    Author      string              // Author
}
```

**URL generation**:
- `index.md` → `/`
- `about.md` → `/about/`
- `blog/post.md` → `/blog/post/`

### Asset Pipeline

**CSS Processing**:
```go
// Runs: tailwindcss -i input.css -o output.css [--minify]
// Tailwind CLI scans all files for class usage
// Generates minimal CSS with only used classes
```

**JavaScript Processing**:
```go
// Uses tdewolff/minify (pure Go)
// Concatenates multiple JS files
// Minifies if config.Assets.Minify = true
```

**Static Files**:
```go
// Copies images, fonts, etc.
// Preserves directory structure
// Skips css/ and js/ directories
```

### Development Server

**File watching**:
```go
// Uses fsnotify to watch:
// - content/
// - assets/
// - components/
// - config.yaml

// Debouncing prevents rapid rebuilds
// Filters relevant extensions (.md, .templ, .css, .js, .yaml, .go)
```

**Live reload**:
```go
// WebSocket connection at /_live-reload
// Broadcasts reload message to all connected browsers
// Automatic reconnection with exponential backoff
// Script injected into HTML automatically
```

## Adding New Features

### Adding a Package

1. Create directory in `pkg/` (public) or `internal/` (private)
2. Add package documentation comment
3. Implement functionality
4. Create `*_test.go` file
5. Update imports in dependent packages

Example:
```go
// pkg/newfeature/feature.go

// Package newfeature provides ...
package newfeature

// New creates a new Feature instance
func New(config *Config) *Feature {
    return &Feature{config: config}
}
```

### Adding a CLI Command

1. Create `cmd/templsite/commands/mycommand.go`
2. Implement function with signature: `func MyCommand(ctx context.Context, args []string) error`
3. Add to router in `cmd/templsite/main.go`
4. Add help text and flag parsing
5. Write integration tests in `*_test.go`

Example:
```go
// cmd/templsite/commands/mycommand.go
package commands

import (
    "context"
    "flag"
    "fmt"
)

func MyCommand(ctx context.Context, args []string) error {
    flags := flag.NewFlagSet("mycommand", flag.ExitOnError)
    verbose := flags.Bool("verbose", false, "Enable verbose output")
    
    if err := flags.Parse(args); err != nil {
        return err
    }
    
    // Implementation here
    return nil
}
```

### Adding a Template

1. Create directory in `cmd/templsite/templates/`
2. Add complete site structure (main.go, components/, content/, etc.)
3. Use `{{.ModulePath}}` for import paths
4. Update `embed.go` with new template
5. Add to `ListTemplates()` function

Example structure:
```
templates/
└── mytemplate/
    ├── main.go              # Uses {{.ModulePath}}
    ├── go.mod.tmpl         # Template for go.mod
    ├── config.yaml
    ├── Makefile
    ├── components/
    ├── content/
    └── assets/
```

## Testing

### Test Organization

```
pkg/site/
├── site.go
├── site_test.go           # Unit tests for site.go
├── config.go
├── config_test.go         # Unit tests for config.go
└── integration_test.go    # Full pipeline tests
```

### Testing Patterns

**Use temp directories**:
```go
func TestFeature(t *testing.T) {
    tmpDir := t.TempDir() // Automatically cleaned up
    
    // Create test files in tmpDir
    // Run test
    // Assert results
}
```

**Table-driven tests**:
```go
func TestParser(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected Page
        wantErr  bool
    }{
        {"valid", "# Title", Page{...}, false},
        {"empty", "", Page{}, false},
        {"invalid", "---\nbad yaml\n", Page{}, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

**Context cancellation**:
```go
func TestCancellation(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    cancel() // Cancel immediately
    
    err := operation(ctx)
    if err != context.Canceled {
        t.Errorf("expected context.Canceled, got %v", err)
    }
}
```

**Mock external dependencies**:
```go
// For Tailwind CLI, test with missing binary
// For file watching, use temp directories
// For HTTP, use httptest.Server
```

### Running Tests

```bash
# All tests
make test
go test ./...

# Specific package
go test ./pkg/site/...

# With coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Verbose output
go test -v ./pkg/site/

# Run specific test
go test -v -run TestBuild ./pkg/site/

# Race detector (always use for concurrent code)
go test -race ./...

# Benchmarks
go test -bench=. ./pkg/content/
```

### Coverage Goals

- New packages: >75% coverage
- Critical paths: >90% coverage
- Overall project: >80% coverage

Current status:
- `pkg/content`: 85.8% ✅
- `pkg/assets`: 78.8% ✅
- `internal/watch`: 84.1% ✅
- `pkg/site`: 72% (needs improvement)
- `cmd/templsite/commands`: 23% (needs improvement)

## Code Style

### General Guidelines

```go
// Good: Use structured logging
slog.Info("building site", "title", site.Config.Title, "pages", len(site.Pages))

// Bad: Use fmt.Printf
fmt.Printf("Building %s with %d pages\n", site.Config.Title, len(site.Pages))

// Good: Wrap errors with context
return fmt.Errorf("processing content: %w", err)

// Bad: Return unwrapped errors
return err

// Good: Check context cancellation in loops
for _, page := range pages {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    // Process page
}

// Good: Meaningful variable names
contentParser := content.NewParser(dir)

// Bad: Single letters (except i, j in loops)
p := content.NewParser(dir)
```

### Package Documentation

```go
// Package site provides core site management functionality for templsite.
//
// The Site type orchestrates content parsing, asset building, and provides
// helpers for page rendering. It does not perform rendering itself - that
// happens in the user's binary using their custom components.
//
// Example usage:
//
//     site, err := site.New("config.yaml")
//     if err != nil {
//         return err
//     }
//     
//     if err := site.Build(ctx); err != nil {
//         return err
//     }
//     
//     // User's rendering code here
//
package site
```

### Function Documentation

```go
// GetOutputPath converts a URL path to a filesystem output path.
// It handles clean URL structure by generating directory/index.html paths.
//
// Examples:
//   - "/" → "public/index.html"
//   - "/about/" → "public/about/index.html"
//   - "/blog/post/" → "public/blog/post/index.html"
func (s *Site) GetOutputPath(url string) string {
    // Implementation
}
```

### Error Messages

```go
// Good: Helpful, actionable
return fmt.Errorf("config file not found: %s\nCreate one with: templsite new mysite", path)

// Bad: Vague
return fmt.Errorf("file not found")

// Good: Include context
return fmt.Errorf("parsing frontmatter in %s: %w", page.Path, err)

// Bad: No context
return err
```

## CSS & JavaScript Guidelines

### CSS Philosophy

**Minimal CSS files** - Let Tailwind and DaisyUI do the work:

```css
/* assets/css/app.css - This is ALL you need */
@import "tailwindcss";
@plugin "daisyui";

/* Optional: Brand colors */
@theme {
  --color-brand-500: oklch(0.6 0.15 210);
}
```

**Use DaisyUI components directly**:
```templ
// Good: Use DaisyUI classes
<button class="btn btn-primary">Click Me</button>

// Bad: Custom CSS classes
<button class="my-custom-button">Click Me</button>
```

**Avoid @layer components**:
```css
/* Bad: Custom component classes */
@layer components {
  .my-card {
    @apply rounded-lg shadow-xl p-4;
  }
}

/* Good: Use DaisyUI's card component */
<div class="card bg-base-100 shadow-xl">...</div>
```

### JavaScript Philosophy

**Keep it minimal**:

```javascript
// Good: Small, focused utilities
// assets/js/app.js
window.utils = {
    formatDate(date) {
        return new Date(date).toLocaleDateString();
    },
    
    debounce(func, wait) {
        let timeout;
        return function(...args) {
            clearTimeout(timeout);
            timeout = setTimeout(() => func.apply(this, args), wait);
        };
    }
};
```

**Component-specific scripts inline**:
```templ
templ MyComponent() {
    <div id="counter">0</div>
    <button id="increment">+</button>
    <script>
        {
            let count = 0;
            const display = document.getElementById('counter');
            document.getElementById('increment').onclick = () => {
                count++;
                display.textContent = count;
            };
        }
    </script>
}
```

**Progressive enhancement**:
1. Works without JavaScript
2. Enhanced with JavaScript
3. No framework required for simple interactions

## Git Workflow

### Branching

```bash
# Create feature branch
git checkout -b feature-name

# Or for stages
git checkout -b stage-10-component-registry
```

### Commits

```bash
# Good commit messages
git commit -m "Add component registry client with caching"
git commit -m "Fix: Handle missing Tailwind CLI gracefully"
git commit -m "Test: Add integration tests for build command"

# Commit frequently with logical units
git add pkg/components/registry.go
git commit -m "Add registry client structure"

git add pkg/components/registry_test.go
git commit -m "Add registry client tests"
```

### Merging

```bash
# Merge with no-ff to preserve branch history
git checkout main
git merge feature-name --no-ff

# For stages, tag after merge
git tag v0.8.0-stage10
```

### Pull Requests

1. Create feature branch
2. Make changes with frequent commits
3. Write/update tests
4. Update documentation
5. Push and create PR
6. Address review comments
7. Squash if needed before merge

## Debugging

### Common Issues

**Tests failing after architectural changes**:
- Check if `Site.Build()` behavior changed
- Update mocks/stubs to match new interfaces
- Review test assumptions about rendering

**Tailwind CLI not found**:
- Run `make setup-tailwind`
- Check `bin/tailwindcss` exists and is executable
- Verify OS/architecture detection in Makefile

**Live reload not working**:
- Check browser console for WebSocket errors
- Verify `/_live-reload` route is accessible
- Check file watcher is detecting changes (`--verbose`)

**templ components not updating**:
- Run `templ generate` manually
- Check `*_templ.go` files are generated
- Rebuild binary after component changes

### Debug Logging

```bash
# Enable verbose logging
./templsite build --verbose
./templsite serve --verbose

# Or set in code
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))
slog.SetDefault(logger)
```

### Profiling

```go
import _ "net/http/pprof"

// In serve command, pprof automatically available at:
// http://localhost:8080/debug/pprof/

// CPU profile
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

// Memory profile
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

## Release Process

### Version Numbering

Follow semantic versioning:
- **Major** (v2.0.0): Breaking changes
- **Minor** (v1.1.0): New features, backward compatible
- **Patch** (v1.0.1): Bug fixes

### Creating a Release

```bash
# 1. Update version in code
# 2. Update CHANGELOG.md
# 3. Run full test suite
make test

# 4. Create tag
git tag v1.0.0

# 5. Push tag
git push origin v1.0.0

# 6. GitHub Actions builds binaries (when configured)
# 7. Create GitHub release with notes
```

## Contributing

### Before Contributing

1. Read this document
2. Read [ARCHITECTURE.md](ARCHITECTURE.md)
3. Check [PLAN.md](PLAN.md) for current priorities
4. Look for existing issues or create one
5. Discuss approach before large changes

### Contribution Checklist

- [ ] Tests written and passing
- [ ] Documentation updated
- [ ] Code formatted (`make fmt`)
- [ ] No lint errors (`make vet`)
- [ ] Coverage maintained or improved
- [ ] Examples added if applicable
- [ ] CHANGELOG.md updated (for significant changes)

### Code Review

Expect feedback on:
- Test coverage and quality
- Error handling
- Documentation clarity
- Performance implications
- Breaking changes
- Alignment with project philosophy

## Resources

### Go Resources
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Proverbs](https://go-proverbs.github.io/)

### Project Resources
- [templ documentation](https://templ.guide/)
- [goldmark](https://github.com/yuin/goldmark)
- [Tailwind CSS v4](https://tailwindcss.com/)
- [DaisyUI](https://daisyui.com/)
- [fsnotify](https://github.com/fsnotify/fsnotify)

### Tools
- [golangci-lint](https://golangci-lint.run/)
- [goreleaser](https://goreleaser.com/)
- [Air](https://github.com/cosmtrek/air) - Live reload for Go apps

---

**Questions?** Open an issue or discussion in the repository.