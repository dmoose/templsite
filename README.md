# templsite

[![Go](https://github.com/dmoose/templsite/actions/workflows/ci.yml/badge.svg)](https://github.com/dmoose/templsite/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dmoose/templsite)](https://goreportcard.com/report/github.com/dmoose/templsite)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A static site generator built with Go and [templ](https://templ.guide/) components. Zero Node.js dependencies, type-safe templates, modern CSS with Tailwind v4.

## Why templsite?

- **Type-safe templates** — templ components are Go functions, not string templates. Errors are caught at compile time.
- **Zero Node.js** — Pure Go toolchain. Tailwind CSS v4 runs via standalone CLI.
- **Full customization** — Each site compiles its own binary with its own components. No theme layer to fight with.

See [ARCHITECTURE.md](ARCHITECTURE.md) for the design rationale.

## Quick Start

```bash
# Install
git clone https://github.com/dmoose/templsite
cd templsite
make setup && make build

# Create a site
templsite new mysite --template tailwind
cd mysite
make serve
```

Your site is running at `http://localhost:8080` with live reload.

## How It Works

Each site is a standalone Go project. The `templsite` library provides content parsing, asset building, and a dev server. Your site's `main.go` controls rendering using your own templ components.

```
mysite/
├── main.go              # Your binary — renders with YOUR components
├── components/          # templ components (customize these)
│   ├── layout/
│   │   ├── base.templ   # HTML shell
│   │   └── page.templ   # Page layout
│   └── ui/
│       ├── header.templ
│       └── footer.templ
├── content/             # Markdown with YAML frontmatter
├── assets/css/app.css   # Tailwind CSS v4 entry point
├── config.yaml          # Site configuration
└── public/              # Generated output (gitignored)
```

## Features

- **Live reload** — WebSocket-based auto-refresh during development
- **Content organization** — Sections, taxonomies (tags/categories), prev/next navigation
- **Feeds and SEO** — RSS, Atom, JSON Feed, sitemap.xml, robots.txt
- **Pagination** — Built-in paginator with URL generation
- **Page enrichment** — Summary, reading time, word count, table of contents
- **Data files** — Load YAML/JSON data for use in templates
- **Navigation menus** — Configurable menus with active state tracking
- **Environment configs** — `--env` flag with config overlay files
- **Content scaffolding** — Archetypes for `./site new posts/my-post`
- **Syntax highlighting** — Code blocks highlighted with Chroma

## Commands

```bash
# templsite CLI
templsite new <path>                             # Create site (tailwind template)
templsite new <path> --template tailwind         # Explicit template selection
templsite new <path> --templsite-path ../templsite  # Local dev mode

# Site commands (in your site directory)
make serve        # Dev server with live reload
make build        # Compile site binary
make deploy       # Generate static files to public/
make prod         # Production build (uses config.production.yaml)
```

## Configuration

```yaml
title: "My Site"
baseURL: "https://example.com"
description: "Site description for feeds"
language: "en"

taxonomies:
  - tags
  - categories

menus:
  main:
    - name: Home
      url: /
      weight: 1
    - name: About
      url: /about/
      weight: 2

build:
  drafts: false
  future: false
```

For environment overrides, create `config.production.yaml` with only the fields that change.

## Content

Markdown files in `content/` with YAML frontmatter:

```markdown
---
title: "My Blog Post"
date: 2025-01-15
draft: false
tags: ["golang", "web"]
---

Your content here with **Markdown** formatting.

<!--more-->

Content after this marker is excluded from the summary.
```

### URL structure

| File path | URL |
|-----------|-----|
| `content/index.md` | `/` |
| `content/about.md` | `/about/` |
| `content/blog/_index.md` | `/blog/` |
| `content/blog/first-post.md` | `/blog/first-post/` |

## Customizing Components

Edit any `.templ` file in `components/`. Changes are compiled and reloaded automatically during `make serve`.

```go
// components/ui/header.templ
package ui

templ Header(title string) {
    <header class="bg-white shadow">
        <div class="max-w-7xl mx-auto px-4 py-6">
            <a class="text-xl font-bold" href="/">{title}</a>
        </div>
    </header>
}
```

## Deployment

```bash
make deploy    # Generates public/ directory
```

Upload `public/` to any static host — Netlify, Vercel, GitHub Pages, S3 + CloudFront, etc.

## Documentation

- [Guide: Templ Developer](docs/guide-templ-developer.md) — Full guide: content features, pagination, taxonomies, feeds, deployment
- [Architecture](ARCHITECTURE.md) — Why each site compiles its own binary

## FAQ

**Why does each site compile its own binary?**
Templ components are Go functions compiled at build time, not runtime templates. Each site must be a Go project to allow component customization with type safety and IDE support. See [ARCHITECTURE.md](ARCHITECTURE.md).

**Do I need to know Go?**
Basic Go helps, but you can get started by editing Markdown content, customizing templ components (HTML-like syntax), and using Tailwind utility classes. The Makefile handles most complexity.

## Requirements

- Go 1.25+
- [templ](https://templ.guide/) CLI

## License

MIT License — see [LICENSE](LICENSE) for details.

Copyright (c) 2025-2026 Catapulsion LLC and contributors
