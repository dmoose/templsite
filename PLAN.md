# templsite Implementation Plan

## Overview
This document outlines the staged implementation plan for building templsite, a modern static site generator built with Go and templ components. The plan is divided into 13 stages, each with clear deliverables and time estimates.

## Current Status
- ✓ Stage 0: Foundation Setup (Complete)
- ✓ Stage 1: Project Structure & Basic CLI (Complete - v0.1.0-stage1)
- ✓ Stage 2: Site Configuration & Core Types (Complete - v0.1.0-stage2)
- ✓ Stage 3: Content Parser (Complete - v0.2.0-stage3)
- ✓ Stage 4: Asset Pipeline - CSS (Complete - v0.2.0-stage4)
- ⏳ Stage 5: Asset Pipeline - JS & Static Files (Next)

---

## Stage 0: Foundation Setup ✓
**Status**: Complete  
**Duration**: N/A

### Completed:
- ✓ Go module initialized with `go.mod`
- ✓ Git repository enabled
- ✓ Design document (`DESIGN.md`) created

---

## Stage 1: Project Structure & Basic CLI ✓
**Status**: Complete  
**Completed**: 2025-01-15  
**Tag**: v0.1.0-stage1  
**Duration**: Week 1  
**Goal**: Get the basic scaffolding and CLI working

### Tasks:
1. **Create directory structure**
   ```
   cmd/templsite/
   cmd/templsite/commands/
   cmd/templsite/templates/
   pkg/site/
   pkg/content/
   pkg/assets/
   pkg/components/
   internal/server/
   internal/watch/
   internal/build/
   ```

2. **Add core dependencies to go.mod**
   ```bash
   go get github.com/a-h/templ@latest
   go get github.com/yuin/goldmark@latest
   go get gopkg.in/yaml.v3@latest
   go get github.com/tdewolff/minify/v2@latest
   go get github.com/fsnotify/fsnotify@latest
   ```

3. **Build basic CLI skeleton** (`cmd/templsite/main.go`)
   - Command parser
   - Context handling with signals (SIGINT, SIGTERM)
   - Stub commands: `new`, `build`, `serve`, `components`
   - Proper error handling and exit codes

4. **Add Makefile**
   - `help` - Show available targets
   - `deps` - Download Go dependencies
   - `setup-tailwind` - Download Tailwind CSS CLI
   - `setup` - Complete setup (deps + tailwind)
   - `build` - Build the binary
   - `test` - Run tests
   - `clean` - Clean build artifacts
   - `install` - Install binary to GOPATH
   - `dev` - Start development server
   - OS/architecture detection for Tailwind CLI

### Deliverables:
- ✅ Directory structure created
- ✅ Dependencies added to `go.mod`
- ✅ `templsite --help` shows commands
- ✅ Binary builds successfully with `make build`
- ✅ Makefile targets work (detects system Tailwind or downloads latest)

### Git:
```bash
git checkout -b stage-1-cli-skeleton
# ... implement ...
git commit -m "Stage 1: CLI skeleton and project structure"
git checkout main
git merge stage-1-cli-skeleton
git tag v0.1.0-stage1
```

---

## Stage 2: Site Configuration & Core Types ✓
**Status**: Complete  
**Completed**: 2025-01-15  
**Tag**: v0.1.0-stage2  
**Duration**: Week 1  
**Goal**: Define data structures and configuration loading

### Tasks:
1. **Implement `pkg/site/config.go`**
   - `Config` struct with YAML tags
   - Fields: Title, BaseURL, ContentDir, OutputDir, Assets
   - Default values (convention over configuration)
   - `LoadConfig(path string)` function
   - Config file validation

2. **Implement `pkg/site/site.go`**
   - `Site` struct (Config, Pages, Assets, BuildTime)
   - `NewSite(configPath string)` constructor
   - `Build(ctx context.Context)` method (stub implementation)
   - Helper methods

3. **Create default `config.yaml` template**
   - Sensible defaults
   - Inline comments explaining each option
   - Example values

4. **Write tests**
   - Test config loading
   - Test default values
   - Test validation

### Deliverables:
- ✅ Config loading works
- ✅ Default values applied correctly
- ✅ Tests passing (82.7% coverage)
- ✅ Example `config.yaml` created

### Git:
```bash
git checkout -b stage-2-configuration
# ... implement ...
git commit -m "Stage 2: Configuration and core types"
git tag v0.1.0-stage2
```

---

## Stage 3: Content Parser ✓
**Status**: Complete  
**Completed**: 2025-01-15  
**Tag**: v0.2.0-stage3  
**Duration**: Week 2  
**Goal**: Parse Markdown files with frontmatter

### Tasks:
1. **Implement `pkg/content/parser.go`**
   - `Parser` struct with goldmark integration
   - `NewParser(contentDir string)` constructor
   - `ParseAll(ctx context.Context)` - walk directory and parse all .md files
   - `ParseFile(ctx context.Context, path string)` - parse single file
   - `parseFrontmatter(content []byte)` - extract YAML frontmatter
   - `generateURL(path string)` - create clean URLs

2. **Implement `pkg/content/page.go`**
   - `Page` struct: Path, Content, Frontmatter, Layout, URL, Date, Draft
   - Frontmatter helper functions: `getStringDefault`, `getBoolDefault`
   - Date parsing from frontmatter

3. **Configure goldmark**
   - GitHub Flavored Markdown extension
   - Auto heading IDs
   - XHTML output
   - Hard wraps support

4. **Write comprehensive tests**
   - Test frontmatter extraction
   - Test Markdown to HTML conversion
   - Test URL generation (index.md → /, post.md → /post/)
   - Test date parsing
   - Test draft handling

### Deliverables:
- ✅ Can parse `.md` files with frontmatter
- ✅ HTML generation works (goldmark with GFM)
- ✅ URL generation correct
- ✅ Tests passing with 85.8% coverage

### Git:
```bash
git checkout -b stage-3-content-parser
# ... implement ...
git commit -m "Stage 3: Markdown content parser"
git tag v0.2.0-stage3
```

---

## Stage 4: Asset Pipeline - CSS ✓
**Status**: Complete  
**Completed**: 2025-01-19  
**Tag**: v0.2.0-stage4  
**Duration**: Week 2  
**Goal**: Process CSS with Tailwind standalone CLI

### Tasks:
1. **Implement `pkg/assets/pipeline.go`**
   - `Pipeline` struct
   - `Config` struct: InputDir, OutputDir, Minify, Fingerprint
   - `New(config *Config)` constructor
   - `Build(ctx context.Context)` - orchestrate all asset processing
   - Initialize tdewolff minifier

2. **Implement `pkg/assets/css.go`**
   - `processCSS(ctx context.Context)` method
   - Execute Tailwind CLI via `exec.CommandContext`
   - Pass input/output paths
   - Support `--minify` flag
   - Handle errors with context

3. **Update Makefile**
   - Auto-download Tailwind CSS v4 standalone CLI
   - OS detection (Linux/macOS)
   - Architecture detection (x64/arm64)
   - Build correct download URL
   - Make binary executable
   - Store in `bin/tailwindcss`

4. **Create example CSS template**
   - `assets/css/app.css` with:
     - `@import "tailwindcss"`
     - `@plugin "daisyui"`
     - `@theme` section with custom colors
     - CSS variables example

### Deliverables:
- ✅ `make setup` detects system Tailwind or downloads latest
- ✅ CSS processing works (Tailwind → compiled CSS)
- ✅ Minification flag works
- ✅ Example CSS created with Tailwind v4 syntax
- ✅ Tests passing with 75.5% coverage

### Git:
```bash
git checkout -b stage-4-css-pipeline
# ... implement ...
git commit -m "Stage 4: CSS asset pipeline with Tailwind"
git tag v0.2.0-stage4
```

---

## Stage 5: Asset Pipeline - JS & Static Files
**Duration**: Week 3  
**Goal**: Complete asset pipeline

### Tasks:
1. **Implement `pkg/assets/js.go`**
   - `processJS(ctx context.Context)` method
   - Read input JS file
   - Minify using tdewolff/minify
   - Write output JS file
   - Preserve source maps (optional)

2. **Implement `pkg/assets/static.go`**
   - `copyStatic(ctx context.Context)` method
   - Walk static asset directory (images, fonts, etc.)
   - Copy files preserving directory structure
   - Skip processed files (CSS, JS)

3. **Add cache-busting** (optional for v1)
   - Content-based hashing for filenames
   - Generate asset manifest
   - Helper to get hashed filenames

4. **Write tests**
   - Test JS minification
   - Test static file copying
   - Test fingerprinting

### Deliverables:
- [ ] JS processing works
- [ ] Static files copied correctly
- [ ] Full asset pipeline functional
- [ ] Tests passing

### Git:
```bash
git checkout -b stage-5-full-asset-pipeline
# ... implement ...
git commit -m "Stage 5: Complete asset pipeline (JS + static)"
git tag v0.3.0-stage5
```

---

## Stage 6: Template System - Basic Rendering
**Duration**: Week 3  
**Goal**: Render pages using templ components

### Tasks:
1. **Install templ CLI**
   ```bash
   go install github.com/a-h/templ/cmd/templ@latest
   ```

2. **Create base templ components**
   - `components/layout/base.templ` - HTML shell (head, body)
   - `components/layout/page.templ` - Page layout wrapper
   - `components/ui/header.templ` - Site header
   - `components/ui/footer.templ` - Site footer

3. **Update Makefile**
   - Add `generate` target: `templ generate`
   - Run before build
   - Add `.templ_generated.go` to .gitignore

4. **Implement rendering in `pkg/site/site.go`**
   - `renderPages(ctx context.Context)` method
   - `renderPage(ctx context.Context, page *Page)` method
   - Load compiled templ components
   - Pass page data to templates
   - Write output HTML files to OutputDir
   - Preserve URL structure

5. **Test rendering**
   - Single page render
   - Multiple pages
   - Different layouts
   - URL structure correct

### Deliverables:
- [ ] templ components compile
- [ ] Can render Markdown → HTML using templ
- [ ] Output structure correct
- [ ] Tests passing

### Git:
```bash
git checkout -b stage-6-templ-rendering
# ... implement ...
git commit -m "Stage 6: templ component rendering"
git tag v0.4.0-stage6
```

---

## Stage 7: Build Command
**Duration**: Week 4  
**Goal**: Complete build workflow

### Tasks:
1. **Implement `cmd/templsite/commands/build.go`**
   - Parse command flags (--config, --output, --verbose)
   - Load configuration
   - Create Site instance
   - Execute build pipeline
   - Progress reporting with slog
   - Error handling with context

2. **Enhance `Site.Build()`**
   - Load configuration
   - Process all content (parse Markdown)
   - Build all assets (CSS, JS, static)
   - Render all pages to HTML
   - Generate clean URL structure
   - Copy static files
   - Report timing and stats

3. **Add structured logging**
   - Use `log/slog` throughout
   - Log levels: Debug, Info, Warn, Error
   - Show build progress
   - Report errors clearly
   - --verbose flag for detailed output

4. **Test full build**
   - Create test site structure
   - Run `templsite build`
   - Verify output directory structure
   - Verify HTML output
   - Verify assets compiled
   - Check error handling

### Deliverables:
- [ ] `templsite build` creates complete static site
- [ ] Progress reporting works
- [ ] Error messages helpful
- [ ] Tests passing

### Git:
```bash
git checkout -b stage-7-build-command
# ... implement ...
git commit -m "Stage 7: Complete build command"
git tag v0.5.0-stage7
```

---

## Stage 8: Development Server
**Duration**: Week 4-5  
**Goal**: Local dev server with hot reload

### Tasks:
1. **Implement `internal/watch/watcher.go`**
   - Use `fsnotify` for file watching
   - `Watcher` struct with channels
   - `Add(path string)` to watch directories
   - Filter relevant file changes (.md, .templ, .css, .js, config.yaml)
   - Debounce rapid changes (avoid rebuild spam)
   - Handle watcher errors

2. **Implement `internal/server/livereload.go`**
   - WebSocket connection management
   - Connection pool for multiple browsers
   - `notifyReload()` - broadcast to all connections
   - Inject live reload script into HTML
   - Handle connection errors

3. **Implement `internal/server/server.go`**
   - `Server` struct (site, watcher, addr)
   - `New(site *Site, addr string)` constructor
   - `Serve(ctx context.Context)` method
   - HTTP server for static files
   - Special route: `/_live-reload` for WebSocket
   - `handleRequest(w, r)` - serve from output directory
   - `handleFileChanges(ctx)` - watch loop
   - Trigger rebuild on file changes
   - Notify connected browsers
   - Graceful shutdown

4. **Implement `cmd/templsite/commands/serve.go`**
   - Parse flags (--port, --addr, --watch)
   - Create server
   - Start file watcher
   - Start HTTP server
   - Handle Ctrl+C gracefully
   - Show server URL

5. **Live reload client script**
   - Small JavaScript snippet
   - Connect to WebSocket
   - Listen for reload message
   - Reload page on notification

### Deliverables:
- [ ] `templsite serve` starts server
- [ ] File watching triggers rebuilds
- [ ] Live reload works in browser
- [ ] Graceful shutdown on Ctrl+C
- [ ] Tests passing

### Git:
```bash
git checkout -b stage-8-dev-server
# ... implement ...
git commit -m "Stage 8: Development server with live reload"
git tag v0.6.0-stage8
```

---

## Stage 9: New Site Command
**Duration**: Week 5  
**Goal**: Bootstrap new sites from templates

### Tasks:
1. **Create starter templates**
   
   **Minimal template** (`cmd/templsite/templates/minimal/`)
   - Basic directory structure
   - Simple config.yaml
   - One example page (index.md)
   - Basic templ components
   - Minimal CSS
   
   **Business template** (`cmd/templsite/templates/business/`)
   - Landing page layout
   - Hero section
   - Features grid
   - CTA sections
   - Contact form
   - Professional styling
   
   **Blog template** (`cmd/templsite/templates/blog/`)
   - Blog post layout
   - Post listing page
   - Archive by date
   - Tags/categories
   - RSS-ready structure

2. **Implement `cmd/templsite/templates/embed.go`**
   - `//go:embed minimal business blog` directive
   - `Templates` embed.FS variable
   - `GetTemplate(name string)` - extract template FS
   - `ListTemplates()` - return available templates

3. **Implement `cmd/templsite/commands/new.go`**
   - Parse args (site path, --template flag)
   - Validate path doesn't exist
   - Get template from embed
   - Create target directory
   - Copy template files recursively
   - Initialize config with site-specific values
   - Print success message with next steps

4. **Add sample content to each template**
   - Example pages with frontmatter
   - Sample components
   - Starter CSS with theme examples
   - README.md with template-specific instructions
   - .gitignore file

### Deliverables:
- [ ] Three templates created and embedded
- [ ] `templsite new mysite --template blog` works
- [ ] Created site builds successfully
- [ ] Each template documented
- [ ] Tests passing

### Git:
```bash
git checkout -b stage-9-new-command
# ... implement ...
git commit -m "Stage 9: New site command with templates"
git tag v0.7.0-stage9
```

---

## Stage 10: Component Registry Integration
**Duration**: Week 6  
**Goal**: Integration with templ-store for component sharing

### Tasks:
1. **Design component registry API**
   - Component identifier format: `cc:ui/hero` (registry:category/name)
   - Version specification: `@latest`, `@v1.2.3`, `@^1.0.0`
   - Registry URL configuration
   - Authentication (API keys)

2. **Implement `pkg/components/registry.go`**
   - `Registry` struct
   - `Lookup(identifier string)` - find component
   - `Download(identifier, version string)` - fetch component
   - `Install(destination string)` - install to project
   - Local cache in `~/.templsite/components/`
   - Version resolution logic
   - HTTP client for registry API

3. **Implement `pkg/components/manifest.go`**
   - `Manifest` struct - track installed components
   - `components.yaml` file format
   - Add/remove/update components
   - Version locking
   - Dependency resolution

4. **Implement `cmd/templsite/commands/components.go`**
   - `add` subcommand - install component
   - `list` subcommand - show installed
   - `update` subcommand - update component(s)
   - `remove` subcommand - uninstall component
   - Progress reporting
   - Error handling

5. **Write tests**
   - Mock registry responses
   - Test component installation
   - Test version resolution
   - Test manifest operations

### Deliverables:
- [ ] Can install components from registry
- [ ] Components tracked in manifest
- [ ] Version locking works
- [ ] All subcommands functional
- [ ] Tests passing

### Git:
```bash
git checkout -b stage-10-component-registry
# ... implement ...
git commit -m "Stage 10: Component registry integration"
git tag v0.8.0-stage10
```

---

## Stage 11: Documentation & Examples
**Duration**: Week 6-7  
**Goal**: Complete documentation and working examples

### Tasks:
1. **Write comprehensive README.md**
   - Project overview
   - Key features with examples
   - Installation instructions (releases + source)
   - Quick start guide
   - Command reference with examples
   - Configuration documentation
   - Content authoring guide
   - Component development guide
   - Deployment instructions
   - License information

2. **Create example sites**
   - `examples/blog/` - Full blog example
   - `examples/business/` - Corporate site
   - `examples/docs/` - Documentation site
   - Each with README and build instructions

3. **Write developer documentation**
   - `CONTRIBUTING.md` - How to contribute
   - `ARCHITECTURE.md` - System design overview
   - `API.md` - Public API documentation
   - Extension points for plugins

4. **Add inline code documentation**
   - Package-level comments
   - Function/method documentation
   - Usage examples in comments
   - Run `go doc` to verify

5. **Create tutorials**
   - Getting started tutorial
   - Building a blog tutorial
   - Component development tutorial
   - Deployment tutorial

### Deliverables:
- [ ] Complete README.md
- [ ] Example sites built and tested
- [ ] Developer documentation complete
- [ ] All code documented
- [ ] Tutorials written

### Git:
```bash
git checkout -b stage-11-documentation
# ... implement ...
git commit -m "Stage 11: Documentation and examples"
git tag v0.9.0-stage11
```

---

## Stage 12: Testing & Polish
**Duration**: Week 7-8  
**Goal**: Production-ready quality

### Tasks:
1. **Write comprehensive tests**
   - Unit tests for all packages (pkg/*)
   - Integration tests for build pipeline
   - End-to-end tests for CLI commands
   - Table-driven tests for parsers
   - Mock external dependencies
   - Test edge cases and errors
   - Target: >80% coverage

2. **Error handling improvements**
   - Audit all error messages
   - Add context to errors (wrap with fmt.Errorf)
   - User-friendly error messages
   - Suggest fixes where possible
   - Validate inputs with helpful hints
   - Recovery from common issues

3. **Performance optimization**
   - Profile with pprof
   - Identify bottlenecks
   - Optimize hot paths
   - Parallel processing (goroutines for page rendering)
   - Reduce allocations
   - Benchmark critical paths

4. **Cross-platform testing**
   - Test on Linux
   - Test on macOS
   - Test on Windows
   - Fix path separator issues
   - Test Tailwind CLI download on all platforms
   - Setup GitHub Actions CI/CD

5. **Release preparation**
   - Version tagging strategy (semver)
   - Changelog generation
   - Release notes template
   - Binary distribution (goreleaser)
   - Installation script
   - Update documentation for v1.0.0

### Deliverables:
- [ ] Test coverage >80%
- [ ] All errors handled gracefully
- [ ] Performance benchmarks documented
- [ ] Works on Linux, macOS, Windows
- [ ] CI/CD pipeline setup
- [ ] Ready for v1.0.0 release

### Git:
```bash
git checkout -b stage-12-testing-polish
# ... implement ...
git commit -m "Stage 12: Testing and production polish"
git tag v1.0.0-rc1
# ... final testing ...
git tag v1.0.0
```

---

## Stage 13: Advanced Features (Post-1.0)
**Duration**: Future releases  
**Goal**: Nice-to-have features for future versions

### Potential Features:
- **Syntax highlighting** - Code block highlighting with chroma
- **Image optimization** - Resize, compress, WebP conversion
- **RSS feed generation** - Automatic RSS/Atom feeds
- **Sitemap generation** - XML sitemap for SEO
- **Search index** - Generate JSON search index
- **Multilingual support** - i18n for multi-language sites
- **Git integration** - Source content from Git repos
- **Incremental builds** - Only rebuild changed pages
- **Plugin system** - Extension points for custom behavior
- **Admin UI** - Optional web-based content editor
- **Analytics integration** - Built-in analytics snippets
- **PWA support** - Service worker generation
- **Social cards** - Auto-generate OpenGraph images
- **Content collections** - Taxonomy and filtering

### Implementation:
- Each feature gets its own stage
- Community feedback guides priority
- Maintain backward compatibility
- Comprehensive testing for each

---

## Dependencies

### Required Go Modules:
```go
github.com/a-h/templ v0.2.747           // templ components
github.com/yuin/goldmark v1.7.8         // Markdown parser
gopkg.in/yaml.v3 v3.0.1                 // YAML config parsing
github.com/tdewolff/minify/v2 v2.21.2   // CSS/JS minification
github.com/fsnotify/fsnotify v1.8.0     // File watching
```

### External Binaries:
- **Tailwind CSS v4 standalone CLI** - Auto-downloaded by Makefile
  - URL: https://github.com/tailwindlabs/tailwindcss/releases/

### Development Tools:
```bash
go install github.com/a-h/templ/cmd/templ@latest
```

---

## Development Workflow

### For Each Stage:

1. **Branch**
   ```bash
   git checkout -b stage-N-description
   ```

2. **Implement**
   - Write code following Go best practices
   - Add tests alongside implementation
   - Update documentation as you go
   - Commit frequently with clear messages

3. **Test**
   ```bash
   make test
   go test ./... -v -race -cover
   ```

4. **Manual Testing**
   - Build the binary
   - Test CLI commands
   - Verify output

5. **Commit**
   ```bash
   git add .
   git commit -m "Stage N: Clear description of changes"
   ```

6. **Merge**
   ```bash
   git checkout main
   git merge stage-N-description
   ```

7. **Tag**
   ```bash
   git tag v0.N.0-stageN
   git push origin main --tags
   ```

---

## Quick Start Guide

### Begin Stage 1:

```bash
# Add initial dependencies
go get github.com/a-h/templ@latest
go get github.com/yuin/goldmark@latest
go get gopkg.in/yaml.v3@latest
go get github.com/tdewolff/minify/v2@latest
go get github.com/fsnotify/fsnotify@latest

# Create directory structure
mkdir -p cmd/templsite/{commands,templates}
mkdir -p pkg/{site,content,assets,components}
mkdir -p internal/{server,watch,build}

# Create initial files
touch cmd/templsite/main.go
touch Makefile
touch README.md
touch .gitignore

# Initial commit
git add .
git commit -m "Stage 1: Initial project structure"
```

---

## Success Criteria

### v0.5.0 (MVP)
- Build static sites from Markdown
- templ component rendering
- Asset pipeline (CSS, JS)
- Basic CLI commands

### v0.8.0 (Beta)
- Development server with live reload
- New site scaffolding
- Component registry integration
- Good documentation

### v1.0.0 (Stable)
- Production-ready quality
- Comprehensive tests
- Cross-platform support
- Complete documentation
- Community-ready

---

## Timeline Estimate

- **Weeks 1-2**: Stages 1-4 (Foundation)
- **Weeks 3-4**: Stages 5-7 (Core Features)
- **Weeks 5-6**: Stages 8-10 (Developer Experience)
- **Weeks 7-8**: Stages 11-12 (Polish & Release)
- **Total**: ~2 months to v1.0.0

---

## Notes

- Prioritize working software over perfect code
- Test early and often
- Document as you build
- Get feedback from potential users
- Keep scope manageable for v1.0
- Advanced features can wait for post-1.0

---

## Resources

- [templ documentation](https://templ.guide/)
- [goldmark documentation](https://github.com/yuin/goldmark)
- [Tailwind CSS v4 docs](https://tailwindcss.com/)
- [DaisyUI documentation](https://daisyui.com/)
- [fsnotify documentation](https://github.com/fsnotify/fsnotify)

---

**Last Updated**: 2025-01-15
**Next Review**: After Stage 3 completion