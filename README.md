# templsite

A static site generator built with Go and [templ](https://templ.guide/) components. Zero Node.js dependencies, type-safe templates, modern CSS with Tailwind v4.

## Two Paths

templsite serves two audiences with dedicated templates:

| Template | Audience | Styling | Command |
|----------|----------|---------|---------|
| **tailwind** | templ developers using Tailwind CSS | Tailwind v4 utility classes | `templsite new mysite --template tailwind` |
| **fastatic** | Sites using [go-components](https://git.catapulsion.com/go-components) | Token-based design system via fstctl | `templsite new mysite --template fastatic` |

Both templates produce standalone Go projects where each site compiles its own binary. The `templsite` library provides content parsing, asset building, and dev server infrastructure. Your site's `main.go` does the rendering using **your** components.

See [ARCHITECTURE.md](ARCHITECTURE.md) for why this design works.

## Quick Start

### Tailwind path (standalone)

```bash
templsite new mysite --template tailwind
cd mysite
make serve
```

### Fastatic path (go-components)

```bash
templsite new mysite --template fastatic
cd mysite
fstctl init --framework=templ --package=../go-components/dist
fstctl add navbar hero section footer card button
make serve
```

For local development of templsite itself, add `--templsite-path`:

```bash
templsite new mysite --template tailwind --templsite-path /path/to/templsite
```

## Features

- **Type-safe templates** - Go's templ components instead of error-prone template strings
- **Zero Node.js** - Pure Go toolchain with Tailwind CSS standalone CLI
- **Live reload** - WebSocket-based auto-refresh during development
- **Modern CSS** - Tailwind CSS v4 with standalone CLI
- **Project scaffolding** - `templsite new` creates complete Go projects from templates
- **Component customization** - Each site compiles with its own templ components
- **Content organization** - Sections, taxonomies (tags/categories), prev/next navigation
- **Page enrichment** - Summary, reading time, word count, table of contents
- **Feeds and SEO** - RSS, Atom, JSON Feed, sitemap.xml, robots.txt
- **Pagination** - Built-in paginator with URL generation
- **Data files** - Load YAML/JSON data for use in templates
- **Navigation menus** - Configurable menus with active state
- **Environment configs** - `--env` flag with config overlay files

## Installation

Requires Go 1.21+:

```bash
git clone https://github.com/dmoose/templsite
cd templsite
make setup    # Downloads deps + Tailwind CLI
make build    # Builds the templsite binary
```

## Project Structure

A site created with `templsite new mysite`:

```
mysite/
├── main.go              # Your site's binary - renders with YOUR components
├── go.mod               # Go module with templsite as dependency
├── config.yaml          # Site configuration
├── Makefile             # Build automation
├── components/          # templ components (customize these)
│   ├── layout/
│   │   ├── base.templ   # HTML shell (DOCTYPE, head, body)
│   │   └── page.templ   # Page layout
│   └── ui/
│       ├── header.templ # Site header
│       └── footer.templ # Site footer
├── content/             # Markdown files with YAML frontmatter
├── archetypes/          # Content templates (for ./site new)
├── assets/              # CSS, JS, images
│   └── css/app.css      # Tailwind CSS entry point
└── public/              # Generated output (gitignored)
```

## Commands

### templsite CLI

```bash
templsite new <path>                          # Create site (default: tailwind template)
templsite new <path> --template fastatic      # Create site with fastatic template
templsite new <path> --templsite-path ../templsite  # Local dev mode
templsite help                                # Show help
templsite version                             # Show version
```

### Site commands (in your site directory)

```bash
make serve           # Dev server with live reload (port 8080)
make build           # Compile your site's binary
make deploy          # Generate static files to public/
make prod            # Production build (uses config.production.yaml)
make restart         # Kill and restart dev server
```

Or directly:

```bash
./site serve                    # Dev server
./site build                    # Static build
./site build --env production   # Production build
./site new posts/my-post        # Create content file from archetype
```

## Configuration

**config.yaml**:

```yaml
title: "My Site"
baseURL: "https://example.com"
description: "Site description for feeds"
language: "en"

content:
  dir: "content"
  defaultLayout: "page"

assets:
  inputDir: "assets"
  outputDir: "assets"
  minify: true
  fingerprint: false

outputDir: "public"

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

params:
  author: "Your Name"
```

For environment overrides, create `config.production.yaml` with only the fields that change. See `config.example.yaml` for all available fields.

## Content Authoring

Markdown files in `content/` with YAML frontmatter:

```markdown
---
title: "My Blog Post"
date: 2025-01-15
draft: false
description: "A great post"
tags: ["golang", "web"]
categories: ["tutorials"]
author: "Jane Doe"
weight: 10
---

Your content here with **Markdown** formatting.

<!--more-->

Content after this marker is excluded from the summary.
```

### Available page fields

After processing, each page has: `Title`, `Description`, `Date`, `Author`, `Content` (HTML), `Summary`, `WordCount`, `ReadingTime`, `TOC`, `Section`, `Tags`, `Prev`, `Next`, `URL`.

### URL structure

- `content/index.md` → `/`
- `content/about.md` → `/about/`
- `content/blog/post.md` → `/blog/post/`

## Customizing Components

Edit any file in `components/`. Example:

```go
// components/ui/header.templ
package ui

templ Header(title string, links []NavLink) {
    <header class="bg-white shadow">
        <div class="max-w-7xl mx-auto px-4 py-6">
            <a class="text-xl font-bold" href="/">{title}</a>
            <nav class="mt-2">
                for _, link := range links {
                    <a class="mr-4 text-blue-600 hover:underline" href={templ.URL(link.URL)}>{link.Text}</a>
                }
            </nav>
        </div>
    </header>
}
```

Run `make serve` and changes are compiled and reloaded automatically.

## Deployment

```bash
make deploy        # Generates public/ directory
# Upload public/ to any static host
```

Compatible with Netlify, Vercel, GitHub Pages, AWS S3 + CloudFront, or any static file host.

## Guides

- [Guide: templ Developer](docs/guide-templ-developer.md) - Tailwind path, content features, contributing
- [Guide: go-components](docs/guide-go-components.md) - Fastatic path, fstctl, token-based design
- [Guide: LLM Authoring](docs/guide-llm-authoring.md) - For LLMs building sites with the fastatic stack

## FAQ

### Why does each site compile its own binary?

Templ components are Go functions compiled at build time, not runtime templates. Each site must be a Go project to allow component customization with type safety, IDE support, and compile-time validation. See [ARCHITECTURE.md](ARCHITECTURE.md).

### Why not Hugo?

Hugo uses runtime text templates. templsite uses compile-time Go functions via templ, giving you type safety, IDE support, and access to any Go library. Trade-off: Hugo has one universal binary; templsite requires compilation (but the Makefile makes it transparent).

### Do I need to know Go?

Basic Go knowledge helps, but you can get started by editing Markdown content, customizing templ components (HTML-like syntax), and using Tailwind utility classes. The Makefile hides most complexity.

## Key Technologies

- **Go 1.21+** - Core language
- **templ** - Type-safe HTML components
- **goldmark** - Markdown parser with GFM
- **Tailwind CSS v4** - Modern CSS (standalone CLI)
- **tdewolff/minify** - Pure Go CSS/JS minification
- **fsnotify** - File system watching
- **gorilla/websocket** - WebSocket for live reload

## License

MIT License — see [LICENSE](LICENSE) for details.
