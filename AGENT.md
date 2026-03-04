# Agent Context Document

This document provides complete context for AI agents working on the templsite project.

## Project Overview

**templsite** is a modern static site generator built entirely in Go. It eliminates Node.js dependencies while providing type-safe, component-driven web development with modern CSS tooling.

### Core Philosophy
- **Type Safety First**: Uses `templ` components that compile to native Go functions
- **Pure Go Toolchain**: Zero Node.js dependencies
- **Modern CSS**: Tailwind CSS v4 standalone CLI with DaisyUI support
- **Component-Driven**: Integration with templ-store for reusable components

## Project Status

### Current Version: v0.6.0-stage8

### Completed Stages
- ✅ Stage 1: CLI Skeleton (v0.1.0-stage1)
- ✅ Stage 2: Configuration (v0.1.0-stage2)
- ✅ Stage 3: Content Parser (v0.2.0-stage3)
- ✅ Stage 4: CSS Pipeline (v0.2.0-stage4)
- ✅ Stage 5: JS & Static Files (v0.3.0-stage5)
- ✅ Stage 6: Template System (v0.4.0-stage6)
- ✅ Stage 7: Build Command (v0.5.0-stage7)
- ✅ Stage 8: Development Server (v0.6.0-stage8)

### Next Stage
- ⏭️ Stage 9: New Site Command - Project Scaffolding

See `PLAN.md` for complete implementation roadmap.

## Project Structure

```
templsite/
├── cmd/templsite/           # CLI application
│   ├── main.go             # Entry point with signal handling
│   ├── commands/           # Command implementations
│   │   ├── build.go        # Build command (COMPLETE)
│   │   ├── serve.go        # Serve command (COMPLETE)
│   │   ├── new.go          # New site command (TODO: Stage 9)
│   │   └── components.go   # Component management (TODO: Stage 10)
│   └── templates/          # Embedded starter templates (TODO: Stage 9)
│
├── components/            # templ layout components
│   ├── layout/            # Layout components
│   │   ├── base.templ     # HTML shell (DOCTYPE, head, body)
│   │   ├── page.templ     # Page layout with header/footer
│   │   └── data.go        # Data types for components
│   └── ui/                # UI components
│       ├── header.templ   # Site header with navigation
│       └── footer.templ   # Site footer with copyright
│
├── pkg/                    # Public libraries
│   ├── site/              # Site management
│   │   ├── config.go      # Configuration with YAML support
│   │   ├── site.go        # Site struct and Build() orchestration
│   │   ├── render_test.go # Rendering tests
│   │   ├── integration_test.go # Full build integration tests
│   │   └── *_test.go      # Tests (86.1% coverage)
│   │
│   ├── content/           # Content parsing
│   │   ├── page.go        # Page struct with metadata
│   │   ├── parser.go      # Markdown parser with goldmark
│   │   └── *_test.go      # Tests (85.8% coverage)
│   │
│   ├── assets/            # Asset pipeline
│   │   ├── pipeline.go    # Pipeline orchestration
│   │   ├── css.go         # Tailwind CSS processing
│   │   ├── js.go          # JavaScript processing
│   │   ├── static.go      # Static file copying
│   │   └── *_test.go      # Tests (78.8% coverage)
│   │
│   └── components/        # Component registry (TODO: Stage 10)
│
├── internal/              # Internal packages
│   ├── server/            # Development server (COMPLETE)
│   │   ├── server.go      # HTTP server and rebuild orchestration
│   │   └── livereload.go  # WebSocket live reload
│   ├── watch/             # File watcher (COMPLETE)
│   │   ├── watcher.go     # fsnotify-based file watching
│   │   └── watcher_test.go # Tests (84.1% coverage)
│   └── build/             # Build orchestration (empty)
│
├── assets/                # Example asset files
│   ├── css/
│   │   └── app.css.example  # Tailwind v4 with theme variables
│   └── js/
│       └── app.js.example   # Common JS utilities
│
├── go.mod                 # Go module definition
├── Makefile              # Build automation
├── DESIGN.md             # Architecture & design philosophy
├── PLAN.md               # Staged implementation plan
└── README.md             # User documentation
```

## Key Technologies

### Dependencies
```go
github.com/a-h/templ v0.3.960           // templ components
github.com/yuin/goldmark v1.7.13        // Markdown parser
gopkg.in/yaml.v3 v3.0.1                 // YAML config
github.com/tdewolff/minify/v2 v2.24.8   // CSS/JS minification
github.com/fsnotify/fsnotify v1.9.0     // File watching
github.com/gorilla/websocket v1.5.3     // WebSocket for live reload
```

### External Tools
- **Tailwind CSS CLI**: System-installed or auto-downloaded to `bin/tailwindcss`
- **templ CLI**: For generating templ components (used in build process)

## Configuration System

### Config Structure (`pkg/site/config.go`)
```go
type Config struct {
    Title      string         `yaml:"title"`
    BaseURL    string         `yaml:"baseURL"`
    Content    ContentConfig  `yaml:"content"`
    Assets     AssetsConfig   `yaml:"assets"`
    OutputDir  string         `yaml:"outputDir"`
    ThemeColor string         `yaml:"themeColor,omitempty"`
}

type ContentConfig struct {
    Dir           string `yaml:"dir"`
    DefaultLayout string `yaml:"defaultLayout"`
}

type AssetsConfig struct {
    InputDir    string `yaml:"inputDir"`
    OutputDir   string `yaml:"outputDir"`
    Minify      bool   `yaml:"minify"`
    Fingerprint bool   `yaml:"fingerprint"`
}
```

### Default Values (Convention over Configuration)
- Title: "My Site"
- BaseURL: "http://localhost:8080"
- Content.Dir: "content"
- Content.DefaultLayout: "page"
- Assets.InputDir: "assets"
- Assets.OutputDir: "assets"
- OutputDir: "public"

### Path Helpers
- `ContentPath(base string)` - Absolute path to content directory
- `AssetsInputPath(base string)` - Absolute path to asset input
- `OutputPath(base string)` - Absolute path to output directory
- `AssetsOutputPath(base string)` - Absolute path to asset output

## Content Parser

### Page Structure (`pkg/content/page.go`)
```go
type Page struct {
    Path        string              // Original file path
    Content     string              // Rendered HTML
    Frontmatter map[string]any      // Parsed YAML frontmatter
    Layout      string              // Layout template name
    URL         string              // Generated URL path
    Date        time.Time           // Publication date
    Draft       bool                // Draft status
    Title       string              // Page title
    Description string              // Meta description
    Tags        []string            // Tags
    Author      string              // Author name
}
```

### Frontmatter Parsing
- Supports YAML between `---` delimiters
- Handles empty, partial, and full frontmatter
- Multiple date formats: YYYY-MM-DD, RFC3339, etc.
- YAML may parse dates as time.Time or string

### Markdown Processing
- goldmark with GitHub Flavored Markdown
- Extensions: GFM, Tables, Strikethrough, Linkify, Task Lists
- Auto heading IDs
- XHTML output with hard wraps

### URL Generation
- `index.md` → `/`
- `about.md` → `/about/`
- `blog/post.md` → `/blog/post/`
- `blog/index.md` → `/blog/`

### Helper Methods
- `IsPublished()` - Checks draft status and date
- `HasTag(tag string)` - Tag filtering

## Asset Pipeline

### Pipeline Flow
1. **CSS Processing** (`processCSS`)
   - Input: `assets/css/app.css`
   - Output: `public/assets/css/main.css`
   - Uses Tailwind CLI (system or local)
   - Supports `--minify` flag

2. **JavaScript Processing** (`processJS`)
   - Input: `assets/js/app.js`
   - Output: `public/assets/js/main.js`
   - Pure Go minification with tdewolff/minify
   - Gracefully skips if no input

3. **Static File Copying** (`copyStatic`)
   - Copies images, fonts, videos, PDFs, etc.
   - Skips `css/` and `js/` directories
   - Preserves directory structure and permissions
   - Supports nested directories

### Tailwind CLI Detection
```go
func findTailwindCLI() string
```
1. Checks system PATH for `tailwindcss`
2. Checks local `bin/tailwindcss`
3. Returns empty string if not found

## Build Workflow

### Site.Build() Orchestration
```go
func (s *Site) Build(ctx context.Context) error {
    s.BuildTime = time.Now()
    
    // 1. Process content (parse Markdown)
    if err := s.processContent(ctx); err != nil {
        return fmt.Errorf("processing content: %w", err)
    }
    
    // 2. Build assets (CSS, JS, static)
    if err := s.buildAssets(ctx); err != nil {
        return fmt.Errorf("building assets: %w", err)
    }
    
    // 3. Render pages
    if err := s.renderPages(ctx); err != nil {
        return fmt.Errorf("rendering pages: %w", err)
    }
    
    return nil
}
```

## Makefile Targets

### Common Commands
- `make setup` - Download dependencies and Tailwind CLI
- `make build` - Build the binary (includes `templ generate`)
- `make test` - Run tests with race detector
- `make clean` - Clean build artifacts (including generated templ files)
- `make install` - Install to GOPATH
- `make generate` - Generate templ components

### Tailwind CLI Management
- Detects system-installed Tailwind first
- Falls back to downloading latest from GitHub releases
- OS/architecture detection (Linux/macOS, x64/arm64)
- Stores in `bin/tailwindcss`

## Testing Strategy

### Test Coverage
- `pkg/site`: 86.1%
- `pkg/content`: 85.8%
- `pkg/assets`: 78.8%
- `internal/watch`: 84.1%
- `cmd/templsite/commands`: 50.4%

### Testing Patterns
- Use `t.TempDir()` for isolated file operations
- Test both success and error cases
- Test context cancellation
- Mock external dependencies when needed
- Table-driven tests for multiple scenarios

### Example Test Structure
```go
func TestFeature(t *testing.T) {
    tmpDir := t.TempDir()
    // Setup
    // Execute
    // Assert
}
```

## Git Workflow

### Branch Strategy
```bash
git checkout -b stage-N-description
# Implement feature
git add .
git commit -m "Stage N: Description"
git checkout main
git merge stage-N-description --no-ff
git tag vX.Y.Z-stageN
```

### Commit Message Format
```
Stage N: Brief title

Detailed changes:
- Feature 1
- Feature 2
- Tests and coverage

Deliverables:
✅ Item 1
✅ Item 2
```

## Common Development Tasks

### Adding a New Feature
1. Create branch: `git checkout -b feature-name`
2. Implement code in appropriate package (`pkg/` or `internal/`)
3. Write tests achieving >75% coverage
4. Update documentation (AGENT.md, PLAN.md, README.md if user-facing)
5. Run `make test` to verify
6. Commit with descriptive message

### Adding a New Package
1. Create directory in `pkg/` (public) or `internal/` (private)
2. Add package documentation comment
3. Implement functionality
4. Create `*_test.go` file with tests
5. Update imports in dependent packages

### Fixing a Bug
1. Write failing test that reproduces bug
2. Fix the bug
3. Verify test now passes
4. Check for similar issues in related code

## Completed Stages

### Stage 6: Template System ✅

#### Implemented Features
- Created base templ components:
  - `components/layout/base.templ` - HTML shell (DOCTYPE, head, body)
  - `components/layout/page.templ` - Page layout with header/footer
  - `components/ui/header.templ` - Site header with navigation
  - `components/ui/footer.templ` - Site footer with copyright
- Updated Makefile: `build` target depends on `generate`
- Implemented `renderPages()` - orchestrates page rendering
- Implemented `renderPage()` - renders individual pages
- Implemented `getOutputPath()` - clean URL structure

#### Key Implementation Details
- templ components compile to Go functions via `templ generate`
- Generated files: `*_templ.go` (gitignored)
- Components receive typed parameters (site config, page data)
- URL structure: `/about/` → `public/about/index.html`
- Full build pipeline: content → assets → rendering
- All errors handled at compile time (type-safe)

#### Test Results
- 48 total tests passing
- pkg/site coverage: 86.1%
- Integration tests verify full build pipeline

### Stage 7: Build Command ✅

#### Implemented Features
- Implemented `cmd/templsite/commands/build.go` with full CLI
- Command-line flags:
  - `--config` - Specify configuration file path
  - `--output` - Override output directory
  - `--verbose` - Enable debug-level logging
  - `--clean` - Clean output directory before build
- Progress reporting with structured logging (log/slog)
- Build statistics reporting (pages, assets, files, size, duration)
- Error handling with helpful messages
- Context cancellation support

#### Key Implementation Details
- Flag parsing with standard `flag` package
- Structured logging levels: Info (default), Debug (--verbose)
- Build stats collection via directory walking
- Human-readable size formatting (B, KB, MB, GB)
- Absolute path reporting for clarity

#### Test Results
- 7 integration tests for build command
- Coverage: 50.4% for commands package (diluted by stub commands)
- Tests cover: basic build, flags, output override, clean, missing config, cancellation

#### User Experience
```bash
$ templsite build
✓ Build successful!
  Duration: 258ms
  Pages: 1
  Assets: 1
  Total files: 2
  Total size: 6.0 KB
  Output: /path/to/public
```

### Stage 8: Development Server ✅

#### Implemented Features
- File watching with fsnotify (`internal/watch/watcher.go`):
  - Watches content, assets, components, config.yaml
  - Debouncing (100ms default) to prevent rebuild spam
  - Filters relevant file types (.md, .templ, .css, .js, .yaml, .go)
  - Ignores hidden files, build output, generated files
  
- WebSocket live reload (`internal/server/livereload.go`):
  - Connection pool for multiple browsers
  - Ping/pong keep-alive
  - Graceful disconnect handling
  - Automatic reconnection with exponential backoff
  - Client-side JavaScript automatically injected into HTML

- HTTP development server (`internal/server/server.go`):
  - Serves static files from public directory
  - Auto-injects live reload script
  - Automatic rebuild on file changes
  - Clean URLs (serves index.html for directories)
  - Proper content types for all file formats
  - No-cache headers for development
  - Recursive directory watching

- Serve command (`cmd/templsite/commands/serve.go`):
  - Flag support (--port, --addr, --config, --verbose)
  - Graceful shutdown on Ctrl+C
  - Initial build on startup
  - Progress reporting

#### Key Implementation Details
- Debouncing prevents multiple rapid rebuilds (100ms delay, configurable)
- WebSocket route: `/_live-reload`
- File watcher recursively adds subdirectories
- Minimum rebuild interval: 500ms (prevents overlapping builds)
- JavaScript reconnection: 1s → 30s exponential backoff

#### Test Results
- 8 unit tests for watcher (84.1% coverage)
- Tests cover: file changes, debouncing, context cancellation, file filtering
- 56 total tests passing across project

#### User Experience
```bash
$ templsite serve

✓ Development server running at http://localhost:8080
  Press Ctrl+C to stop

[Edit file] → Automatic rebuild → Browser refreshes
```

## Stage 9 Preparation

### What's Next: New Site Command
Stage 9 will implement project scaffolding from templates.

#### Tasks
1. Create three starter templates (minimal, business, blog)
2. Implement `cmd/templsite/templates/embed.go` with `go:embed`
3. Implement `cmd/templsite/commands/new.go` with template extraction
4. Create sample content for each template
5. Add template-specific documentation

#### Key Considerations
- Templates embedded with `//go:embed` directive
- Must copy recursively preserving structure
- Initialize config.yaml with user-provided values
- Each template should be immediately buildable
- Include .gitignore in templates

## Important Notes

### Design Decisions
1. **Convention over Configuration**: Sensible defaults minimize config
2. **Context Everywhere**: All operations support context.Context for cancellation
3. **Graceful Degradation**: Missing inputs skip processing, don't error
4. **Structured Logging**: Use `log/slog` throughout
5. **Pure Go**: No Node.js or system dependencies except Tailwind CLI

### Known Limitations
- New site command not implemented (Stage 9)
- Component registry not implemented (Stage 10)
- No incremental builds (rebuilds everything)
- No advanced features (RSS, sitemap, search, etc.) - Post-1.0

### Testing Philosophy
- Tests should be fast and isolated
- Use temp directories for file operations
- Test error paths, not just happy paths
- Context cancellation should be tested
- External dependencies (Tailwind CLI) failures should be handled gracefully
- Prefer table-driven tests for multiple scenarios

### Code Style Guidelines
- Use structured logging (`log/slog`) not `fmt.Printf`
- Wrap errors with context: `fmt.Errorf("action failed: %w", err)`
- Check context cancellation in loops: `select { case <-ctx.Done(): return ctx.Err() }`
- Use meaningful variable names (avoid single letters except loop counters)
- Add package-level documentation comments
- Export only what's necessary from packages

## Useful Resources

### Documentation
- [DESIGN.md](DESIGN.md) - Complete architectural design and theory
- [PLAN.md](PLAN.md) - Staged implementation plan with timeline
- [README.md](README.md) - User-facing documentation

### External Docs
- [templ documentation](https://templ.guide/)
- [goldmark documentation](https://github.com/yuin/goldmark)
- [Tailwind CSS v4 docs](https://tailwindcss.com/)
- [tdewolff/minify docs](https://github.com/tdewolff/minify)
- [fsnotify documentation](https://github.com/fsnotify/fsnotify)
- [gorilla/websocket documentation](https://github.com/gorilla/websocket)

## Quick Reference Commands

```bash
# Build and test
make build
make test

# Check a specific package
go test -v ./pkg/site/...
go test -v -run TestSpecificTest ./pkg/content/...

# Coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Format code
make fmt

# Install dependencies
go mod tidy

# Generate templ components
make generate

# Run binary
./bin/templsite --help
./bin/templsite build
./bin/templsite serve
```

## Debugging Tips

1. **Test Failures**: Run with `-v` flag for verbose output
2. **Build Errors**: Check `go.mod` dependencies are up to date
3. **Path Issues**: Remember paths must start with project root (e.g., `templsite/...`)
4. **Context Errors**: Check if operations properly handle context cancellation
5. **Tailwind Errors**: Verify CLI is installed with `make setup-tailwind`
6. **WebSocket Issues**: Check browser console for connection errors
7. **File Watch Issues**: Use `--verbose` flag to see debug logs

## Current Capabilities

### What Works
- ✅ Parse Markdown with YAML frontmatter
- ✅ Process Tailwind CSS (auto-download CLI if needed)
- ✅ Minify JavaScript
- ✅ Copy static assets
- ✅ Render pages with type-safe templ components
- ✅ Build complete static sites
- ✅ Development server with live reload
- ✅ File watching with automatic rebuilds
- ✅ Multi-browser support for live reload

### What Doesn't Work Yet
- ❌ Project scaffolding (`templsite new`)
- ❌ Component registry integration
- ❌ Advanced features (RSS, sitemap, search)
- ❌ Incremental builds

### How to Use
```bash
# Setup a project manually
mkdir mysite && cd mysite
mkdir -p content assets/css components
echo "title: My Site\nbaseURL: http://localhost:8080" > config.yaml
echo "# Hello\nContent here" > content/index.md
echo "@import 'tailwindcss';" > assets/css/app.css

# Create templ components (copy from examples)
# Then build
templsite build

# Or develop
templsite serve
```

## Contact & Contributing

This is an active development project. When resuming work:

1. Read PLAN.md for current stage objectives
2. Review recent commits for context (`git log --oneline -10`)
3. Run `make test` to verify current state
4. Check DESIGN.md for architectural decisions
5. Follow existing patterns and conventions
6. Always add debug logging for new features
7. Update this document (AGENT.md) when completing stages

---

**Last Updated**: 2025-12-20  
**Current Stage**: Stage 8 Complete (v0.6.0-stage8)  
**Next Milestone**: Stage 9 - New Site Command  
**Total Tests**: 56 passing  
**Average Coverage**: ~77%