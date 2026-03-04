# templsite

A modern static site generator built with Go and templ components. Zero Node.js dependencies, type-safe templates, and modern CSS with Tailwind + DaisyUI.

## Status

🚀 **Stage 8 Complete** - Development server with live reload is functional!

Current version: **v0.6.0-stage8**

See [PLAN.md](PLAN.md) for the complete implementation roadmap and [DESIGN.md](DESIGN.md) for the architectural design.

## Features

- ✅ **Type-safe templates**: Use Go's templ components instead of error-prone template strings
- ✅ **Zero Node.js**: Pure Go toolchain with Tailwind CSS standalone CLI
- ✅ **Live reload**: Instant feedback during development with WebSocket-based auto-refresh
- ✅ **Modern CSS**: Tailwind CSS v4 + DaisyUI 5 with CSS variables
- ✅ **Build command**: Full-featured CLI with progress reporting
- 🚧 **Component library**: Integration with templ-store (planned for Stage 10)

## Installation

### From source (requires Go 1.21+)

```bash
git clone https://github.com/yourorg/templsite
cd templsite
make setup
make build
```

The `make setup` command will:
- Download Go dependencies
- Auto-detect or download Tailwind CSS CLI to `bin/`

## Quick Start

Currently at Stage 8, you can manually create a site and use the build/serve commands:

```bash
# Create site structure
mkdir mysite && cd mysite
mkdir -p content assets/css components/layout components/ui

# Create config
cat > config.yaml << EOF
title: "My Site"
baseURL: "http://localhost:8080"
EOF

# Create content
cat > content/index.md << EOF
---
title: "Welcome"
---
# Hello World
EOF

# Create CSS
cat > assets/css/app.css << EOF
@import "tailwindcss";
EOF

# Copy templ components from templsite/components/ directory

# Build the site
templsite build

# Or start development server with live reload
templsite serve
```

## Commands

### Build

Build your site for production:

```bash
templsite build                    # Use default config.yaml
templsite build --config site.yaml # Custom config
templsite build --output dist      # Custom output directory
templsite build --verbose          # Debug logging
templsite build --clean            # Clean before building
```

### Serve

Start development server with live reload:

```bash
templsite serve                  # Start on localhost:8080
templsite serve --port 3000      # Custom port
templsite serve --verbose        # Debug logging
```

The development server will:
- Perform an initial build
- Watch for file changes (content, assets, components, config)
- Automatically rebuild when files change
- Refresh all connected browsers via WebSocket
- Handle multiple browser connections

### New (Coming in Stage 9)

```bash
# Not yet implemented
templsite new mysite --template business
```

## Development

### Prerequisites

- Go 1.21+
- Make
- curl (for downloading Tailwind CLI)

### Setup

```bash
# Download dependencies and Tailwind CLI
make setup

# Build the binary
make build

# Run tests
make test

# Generate templ components (included in make build)
make generate
```

### Available Make Targets

- `make help` - Show all available targets
- `make setup` - Complete project setup (deps + Tailwind CLI)
- `make build` - Build the binary (includes templ generation)
- `make test` - Run tests with race detector
- `make clean` - Clean build artifacts
- `make install` - Install to GOPATH
- `make generate` - Generate templ components
- `make fmt` - Format code
- `make vet` - Run go vet

## Project Structure

```
templsite/
├── cmd/templsite/           # CLI application
│   ├── main.go             # Entry point
│   ├── commands/           # Command implementations
│   │   ├── build.go        # ✅ Build command
│   │   ├── serve.go        # ✅ Serve command
│   │   ├── new.go          # 🚧 New command (Stage 9)
│   │   └── components.go   # 🚧 Component management (Stage 10)
│   └── templates/          # 🚧 Embedded starter templates (Stage 9)
├── components/             # Example templ components
│   ├── layout/            # Layout components
│   └── ui/                # UI components
├── pkg/                    # Public libraries
│   ├── site/              # Site management (86.1% coverage)
│   ├── content/           # Content parsing (85.8% coverage)
│   ├── assets/            # Asset pipeline (78.8% coverage)
│   └── components/        # Component registry (Stage 10)
├── internal/              # Internal packages
│   ├── server/            # Development server (complete)
│   └── watch/             # File watcher (84.1% coverage)
├── go.mod
├── Makefile
├── DESIGN.md              # Architecture & design philosophy
├── PLAN.md                # Implementation plan
└── AGENT.md               # AI agent context document
```

## Documentation

- [DESIGN.md](DESIGN.md) - Complete architectural design and theory
- [PLAN.md](PLAN.md) - Staged implementation plan with timeline
- [AGENT.md](AGENT.md) - Context document for AI agents working on this project

## Implementation Progress

### Completed Stages (v0.6.0)

- ✅ **Stage 1**: CLI Skeleton - Command structure and Makefile
- ✅ **Stage 2**: Configuration - YAML config with defaults
- ✅ **Stage 3**: Content Parser - Markdown with frontmatter
- ✅ **Stage 4**: CSS Pipeline - Tailwind CSS processing
- ✅ **Stage 5**: JS & Static - Asset pipeline completion
- ✅ **Stage 6**: Template System - templ component rendering
- ✅ **Stage 7**: Build Command - Full CLI with stats
- ✅ **Stage 8**: Development Server - Live reload with WebSocket

### Next Stages

- 🚧 **Stage 9**: New Site Command - Project scaffolding with templates
- 🚧 **Stage 10**: Component Registry - templ-store integration
- 🚧 **Stage 11**: Documentation & Examples
- 🚧 **Stage 12**: Testing & Polish (v1.0.0)

## Testing

The project maintains high test coverage:

- Overall: 56 tests passing
- `pkg/site`: 86.1% coverage
- `pkg/content`: 85.8% coverage
- `pkg/assets`: 78.8% coverage
- `internal/watch`: 84.1% coverage

Run tests with:

```bash
make test                              # All tests
go test -v ./pkg/site                 # Specific package
go test -coverprofile=coverage.out ./... # With coverage
```

## Key Technologies

- **Go 1.21+** - Core language
- **templ** - Type-safe HTML components
- **goldmark** - Markdown parser with GFM
- **Tailwind CSS v4** - Modern CSS (standalone CLI)
- **tdewolff/minify** - Pure Go minification
- **fsnotify** - File system watching
- **gorilla/websocket** - WebSocket for live reload

## Contributing

This project is in active development (Stage 8 of 13 complete). The core functionality is working:

- Content parsing ✅
- Asset processing ✅
- Page rendering ✅
- Build command ✅
- Development server with live reload ✅

Contributions will be welcome once we reach v1.0.0 (Stage 12).

## License

MIT (to be determined)

## Contact

For questions or issues, please see the repository's issue tracker.