# My Site

A statically-generated site built with [templsite](https://github.com/dmoose/templsite) and [Go templ](https://templ.guide/).

## Quick Start

```bash
make build   # Build the site binary
make serve   # Start development server (http://localhost:8080)
make deploy  # Generate static files to public/
make prod    # Build for production (uses config.production.yaml)
```

## Project Structure

```
.
├── main.go              # Site binary (build/serve/new commands)
├── config.yaml          # Site configuration
├── content/             # Markdown content files
├── archetypes/          # Content templates (for ./site new)
├── components/          # Templ components
│   └── layout/          # Page layouts
├── assets/              # CSS and static assets
├── static/              # Files copied to output root (favicons, etc.)
└── public/              # Generated output (gitignored)
```

## Creating Content

Use `./site new` to scaffold content files with the correct frontmatter:

```bash
./site new posts/my-great-idea    # Blog post with date, draft, tags
./site new about                  # Page with title, description
./site new docs/getting-started   # Nested page
```

Or use the Makefile shortcuts:

```bash
make new-post name=my-great-idea  # Creates content/posts/my-great-idea.md
make new-page name=about          # Creates content/about.md
```

### Archetypes

Content templates live in `archetypes/`. When you run `./site new posts/hello`, the tool looks for:

1. `archetypes/posts.md` (section-specific)
2. `archetypes/default.md` (fallback)
3. Built-in default (if no archetype files exist)

Available template variables: `{{.Title}}`, `{{.Date}}`, `{{.Section}}`, `{{.Slug}}`

## Configuration

Edit `config.yaml`:

```yaml
title: "My Site"
baseURL: "https://example.com"
description: "Site description"
language: "en"

content:
  dir: "content"
  defaultLayout: "page"

menus:
  main:
    - name: "Home"
      url: "/"
    - name: "About"
      url: "/about/"

build:
  drafts: false   # Set true to include draft posts in output
```

For environment-specific overrides, create `config.production.yaml` with only the fields you want to change (e.g., `baseURL`).

## Commands

```bash
./site serve                # Development server with live reload
./site build                # Generate static files
./site build --env production  # Build with production config
./site new <path>           # Create new content file
```

## Frontmatter Reference

Pages:
```yaml
---
title: "Page Title"
description: "Page description for SEO"
---
```

Blog posts:
```yaml
---
title: "Post Title"
date: 2024-01-15T10:00:00-05:00
draft: true
tags: [go, web]
---
```

## Customization

- **Components**: Edit templ files in `components/` for full type-safe control over HTML output
- **Styling**: Edit CSS in `assets/`
- **Content**: Add Markdown files to `content/` (or use `./site new`)
