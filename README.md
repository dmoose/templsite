# templsite

A modern static site generator built with Go and templ components. Zero Node.js dependencies, type-safe templates, and modern CSS with Tailwind + DaisyUI.

## Status

🚧 **Work in Progress** - Currently implementing Stage 1 of the development plan.

See [PLAN.md](PLAN.md) for the complete implementation roadmap and [DESIGN.md](DESIGN.md) for the architectural design.

## Features (Planned)

- **Type-safe templates**: Use Go's templ components instead of error-prone template strings
- **Zero Node.js**: Pure Go toolchain with Tailwind CSS standalone CLI
- **Live reload**: Instant feedback during development
- **Modern CSS**: Tailwind CSS v4 + DaisyUI 5 with CSS variables
- **Component library**: Integration with templ-store for reusable components

## Installation

### From source (requires Go 1.25+)

```bash
git clone https://github.com/yourorg/templsite
cd templsite
make setup
make build
```

## Quick Start

Once implementation is complete:

```bash
# Create a new site
templsite new mysite --template business
cd mysite

# Start development server
templsite serve

# Build for production
templsite build
```

## Development

### Prerequisites

- Go 1.25+
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
```

### Available Make Targets

- `make help` - Show all available targets
- `make setup` - Complete project setup (deps + Tailwind CLI)
- `make build` - Build the binary
- `make test` - Run tests
- `make clean` - Clean build artifacts
- `make install` - Install to GOPATH
- `make fmt` - Format code
- `make vet` - Run go vet

## Project Structure

```
templsite/
├── cmd/templsite/           # CLI application
│   ├── main.go             # Entry point
│   ├── commands/           # Command implementations
│   └── templates/          # Embedded starter templates
├── pkg/                    # Public libraries
│   ├── site/              # Site management
│   ├── content/           # Content parsing
│   ├── assets/            # Asset pipeline
│   └── components/        # Component registry
├── internal/              # Internal packages
│   ├── server/            # Development server
│   ├── watch/             # File watcher
│   └── build/             # Build orchestration
├── go.mod
├── Makefile
├── DESIGN.md              # Architecture & design
└── PLAN.md                # Implementation plan
```

## Documentation

- [DESIGN.md](DESIGN.md) - Complete architectural design and theory
- [PLAN.md](PLAN.md) - Staged implementation plan with timeline
- More documentation will be added as features are implemented

## Current Stage: Stage 1

✅ Project structure created  
✅ Dependencies added  
✅ CLI skeleton implemented  
✅ Makefile created  
⏳ Remaining stages in progress

## License

MIT (to be determined)

## Contributing

This project is currently in early development. Contributions will be welcome once the core functionality is implemented.