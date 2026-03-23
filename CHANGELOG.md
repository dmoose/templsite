# Changelog

All notable changes to templsite are documented in this file.

## Unreleased

## v1.0.0

### Added
- Sections, taxonomies, and pagination support
- RSS/Atom/JSON feed generation
- Menu system with nested items
- Data files (YAML/JSON) accessible in templates
- Build options (drafts, future-dated posts)
- Content scaffolding from archetypes
- Sitemap and robots.txt generation
- Syntax highlighting for code blocks
- Asset versioning with content hashing
- Multiple template support (tailwind, fastatic)
- llms.txt generation with data-driven page support
- Cobra CLI with subcommands and auto-generated help

### Fixed
- Resolved all golangci-lint errcheck and unused warnings
- Lint fix for defer/exit ordering

## v0.7.0 — Project Scaffolding

- `templsite new` command for creating sites from templates
- Template substitution for module paths and project names
- Automatic `go mod init` and `go mod tidy` on new projects

## v0.6.0 — Dev Server

- Development server with WebSocket live reload
- File watching with fsnotify and debouncing
- Automatic rebuild on content/asset/template changes
- Graceful shutdown with context cancellation

## v0.5.0 — Build Command

- `templsite build` command for production output
- Full site build orchestration
- Content processing pipeline
- Asset pipeline integration

## v0.4.0 — Template Rendering

- templ component integration
- Base layout with composable page templates
- Render callback pattern (user binary does rendering)

## v0.3.0 — Asset Pipeline

- Tailwind CSS v4 standalone CLI integration (no Node.js)
- JavaScript minification with tdewolff/minify
- Static file copying with directory structure preservation
- CSS asset pipeline with Tailwind

## v0.2.0 — Content Parser

- Markdown parsing with goldmark
- YAML frontmatter extraction
- Page metadata (title, date, draft, tags, etc.)
- Content directory walking and parsing

## v0.1.0 — Foundation

- CLI skeleton with command routing
- Configuration system with YAML and environment overrides
- Core type definitions (Site, Page, Config)
- Project structure and build system
