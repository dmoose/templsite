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

### Current Version: v0.4.0-stage6

### Completed Stages
- ✅ Stage 1: CLI Skeleton (v0.1.0-stage1)
- ✅ Stage 2: Configuration (v0.1.0-stage2)
- ✅ Stage 3: Content Parser (v0.2.0-stage3)
- ✅ Stage 4: CSS Pipeline (v0.2.0-stage4)
- ✅ Stage 5: JS & Static Files (v0.3.0-stage5)
- ✅ Stage 6: Template System (v0.4.0-stage6)

### Next Stage
- ⏭️ Stage 7: Build Command - Complete Build Workflow

See `PLAN.md` for complete implementation roadmap.

## Project Structure

```
templsite/
├── cmd/templsite/           # CLI application
│   ├── main.go             # Entry point with signal handling
│   ├── commands/           # Command implementations (stubs)
│   │   ├── build.go        # Build command (stub)
│   │   ├── serve.go        # Serve command (stub)
│   │   ├── new.go          # New site command (stub)
│   │   └── components.go   # Component management (stub)
│   └── templates/          # Embedded starter templates (TODO)
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
│   ├── server/            # Development server (TODO: Stage 8)
│   ├── watch/             # File watcher (TODO: Stage 8)
│   └── build/             # Build orchestration (TODO: Stage 7)
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
github.com/yuin/goldmark v1.7.13         // Markdown parser
gopkg.in/yaml.v3 v3.0.1                 // YAML config
github.com/tdewolff/minify/v2 v2.24.8   // CSS/JS minification
github.com/fsnotify/fsnotify v1.9.0     // File watching (not yet used)
```

### External Tools
- **Tailwind CSS CLI**: System-installed or auto-downloaded to `bin/tailwindcss`
- **templ CLI**: For generating templ components (Stage 6+)

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
    
    // 3. Render pages (TODO: Stage 6)
    if err := s.renderPages(ctx); err != nil {
        return fmt.Errorf("rendering pages: %w", err)
    }
    
    return nil
}
```

## Makefile Targets

### Common Commands
- `make setup` - Download dependencies and Tailwind CLI
- `make build` - Build the binary
- `make test` - Run tests with race detector
- `make clean` - Clean build artifacts
- `make install` - Install to GOPATH
- `make dev` - Start development server (TODO)

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
git merge stage-N-description
git tag vX.Y.Z-stageN
```

### Commit Message Format
```
Stage N: Brief title

Detailed changes:
- Feature 1
- Feature 2
- Tests and coverage
```

## Common Development Tasks

### Adding a New Feature
1. Create branch: `git checkout -b feature-name`
2. Implement code in appropriate package (`pkg/` or `internal/`)
3. Write tests achieving >75% coverage
4. Update documentation (README.md, PLAN.md)
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

## Stage 6 Complete: Template System ✅

### Implemented Features
Stage 6 successfully implemented templ component rendering for HTML generation.

#### Completed Tasks
1. ✅ Created base templ components:
   - `components/layout/base.templ` - HTML shell (DOCTYPE, head, body)
   - `components/layout/page.templ` - Page layout with header/footer
   - `components/ui/header.templ` - Site header with navigation
   - `components/ui/footer.templ` - Site footer with copyright
2. ✅ Updated Makefile: `build` target depends on `generate`
3. ✅ Implemented `renderPages()` - orchestrates page rendering
4. ✅ Implemented `renderPage()` - renders individual pages
5. ✅ Implemented `getOutputPath()` - clean URL structure
6. ✅ Comprehensive test coverage (6 unit tests + 2 integration tests)

#### Key Implementation Details
- templ components compile to Go functions via `templ generate`
- Generated files: `*_templ.go` (gitignored via .gitignore)
- Components receive typed parameters (site config, page data)
- URL structure: `/about/` → `public/about/index.html`
- Full build pipeline: content → assets → rendering
- All errors handled at compile time (type-safe)

#### Test Results
- 42 total tests passing
- pkg/site coverage: 86.1% (increased from 85.2%)
- Integration tests verify full build pipeline works end-to-end

## Stage 7 Preparation

### What's Next: Build Command
Stage 7 will implement the complete `build` command with proper CLI interface.

#### Tasks
1. Implement `cmd/templsite/commands/build.go`
2. Add command-line flags (--config, --output, --verbose)
3. Progress reporting with structured logging
4. Error handling with helpful messages
5. Build statistics (pages built, time elapsed)
6. Integration with existing Site.Build() method

#### Key Considerations
- Use context for cancellation support
- Structured logging with log/slog
- Clear error messages for users
- Build time reporting
- Verify output directory creation

## Important Notes

### Design Decisions
1. **Convention over Configuration**: Sensible defaults minimize config
2. **Context Everywhere**: All operations support context.Context for cancellation
3. **Graceful Degradation**: Missing inputs skip processing, don't error
4. **Structured Logging**: Use `log/slog` throughout
5. **Pure Go**: No Node.js or system dependencies except Tailwind CLI

### Known Limitations
- Commands are stubs (new, serve, components) - will be implemented in later stages
- No live reload yet (Stage 8)
- No component registry yet (Stage 10)
- Build command is stub (Stage 7)
- Serve command is stub (Stage 8)
- New site scaffolding not implemented (Stage 9)

### Testing Philosophy
- Tests should be fast and isolated
- Use temp directories for file operations
- Test error paths, not just happy paths
- Context cancellation should be tested
- External dependencies (Tailwind CLI) failures should be handled gracefully

## Useful Resources

### Documentation
- [DESIGN.md](DESIGN.md) - Complete architectural design
- [PLAN.md](PLAN.md) - Staged implementation plan with timeline
- [README.md](README.md) - User-facing documentation

### External Docs
- [templ documentation](https://templ.guide/)
- [goldmark documentation](https://github.com/yuin/goldmark)
- [Tailwind CSS v4 docs](https://tailwindcss.com/)
- [tdewolff/minify docs](https://github.com/tdewolff/minify)

## Quick Reference Commands

```bash
# Build and test
make build
make test

# Check a specific package
go test -v ./pkg/site/...
go test -v -run TestSpecificTest ./pkg/content/...

# Coverage report
make test-coverage

# Format code
make fmt

# Install dependencies
go mod tidy

# Run binary
./bin/templsite --help
```

## Debugging Tips

1. **Test Failures**: Run with `-v` flag for verbose output
2. **Build Errors**: Check `go.mod` dependencies are up to date
3. **Path Issues**: Remember paths must start with project root (e.g., `templsite/...`)
4. **Context Errors**: Check if operations properly handle context cancellation
5. **Tailwind Errors**: Verify CLI is installed with `make setup-tailwind`

## Contact & Contributing

This is an active development project. When resuming work:

1. Read PLAN.md for current stage objectives
2. Review recent commits for context
3. Run `make test` to verify current state
4. Check DESIGN.md for architectural decisions
5. Follow existing patterns and conventions

---

**Last Updated**: 2025-01-20  
**Current Stage**: Stage 6 Complete (v0.4.0-stage6)  
**Next Milestone**: Stage 7 - Build Command