# Changelog

All notable changes to templsite are documented in this file.

## v1.3.0

### Changed
- templ v0.3.1118: regenerate with v0.3.1118 when you upgrade.
- Building a site no longer runs `templ generate`. Generate before
  `go build`, as the scaffold Makefile does, or call
  `GenerateTemplComponents`, which runs `go tool templ` and needs templ
  as a tool in the site's go.mod.
- `templsite new` adds templ as a go.mod tool with no version of its
  own, so it resolves to the templ templsite requires. The site Makefile
  runs `go tool templ`.
- `templsite new` resolves the site's dependencies with or without
  `--templsite-path`, and stops at a failed step instead of warning.
- The dev server logs that a changed `.templ` file needs `make serve`
  again; the running binary keeps its compiled components.
- chroma v2.27.0: finer YAML highlighting, and `chroma.css` gains the
  `.or` class.
- goldmark v1.8.7, minify v2.25.0.

### Fixed
- A static directory that cannot be read fails the build instead of
  shipping without its files.
- An asset that cannot be read fails the build instead of being left out
  of the asset version.
- The build summary says when its file count is incomplete.
- `Atom` writes the subtitle it is given.
- The dev server serves only files inside the output directory, symlinks
  included, and sets a read-header timeout.

### Development
- templ is a `tool` in go.mod; `make generate` and `make ci` run
  `go tool templ` at that version.
- golangci-lint runs the gochecks profile; `make check` runs gochecks
  and the tests.

## v1.2.0

### Changed
- templ v0.3.1100. Its runtime moved the attribute helpers to
  `templ/runtime`: regenerate with v0.3.1100 when you upgrade.
- **Breaking:** soft-wrapped lines in Markdown join into one paragraph.
  Two trailing spaces or a backslash still break the line.
- The sitemap leaves out `noindex: true` pages, and sections whose every
  page is noindex.
- Date-only frontmatter is parsed as UTC, so feed timestamps no longer
  carry the builder's local offset.
- fastatic scaffold: `theme-init.js` loads before the stylesheets and the
  head declares `color-scheme` (meta + `:root`/`[data-theme]` CSS), so a
  dark-mode first paint can never flash light.
- fastatic template: page and site name are joined with a hyphen in
  `<title>` and `og:title`.

### Fixed
- `watch`: a data race between the debounce timer and the event loop.

### Development
- `make ci` runs what the workflow runs: templ at the go.mod pin, race
  tests, lint and the `go fix` check.

## v1.1.0 — Reproducible Builds

### Changed
- Builds are byte-reproducible: two builds of unchanged sources produce
  identical output. Feed timestamps now derive from content instead of the
  wall clock.
- **Breaking:** feeds contain only dated pages. An undated page has no honest
  `<updated>`/`pubDate`, so it is no longer syndicated.
- **Breaking:** the feed-level `<updated>` (Atom) and `lastBuildDate` (RSS) are
  the newest entry date rather than the build time.
- **Breaking:** a site with no dated pages emits no feed at all. `feed.xml` is
  no longer written, and `Site.RSS`, `Site.Atom`, and `Site.JSON` return an
  empty string so callers can skip writing the file.
- Sitemap sections and taxonomy terms are emitted in sorted order; previously
  they followed Go map iteration order and varied between builds.
- Upgraded templ to v0.3.1020. CI installs the generator at the version
  `go.mod` pins rather than `@latest`, so the generator and the runtime can no
  longer drift apart.

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
