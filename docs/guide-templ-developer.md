# Guide for Templ Developers Using Tailwind CSS

A practical guide to building static sites with **templsite** -- a Go-based static site generator that uses [templ](https://templ.guide/) components and Tailwind CSS v4. No Node.js required.

---

## 1. Overview

templsite is a static site generator where each site is its own standalone Go project that compiles its own binary. The `templsite` CLI scaffolds new sites, but your site's binary handles rendering using your own templ components.

This design gives you:

- **Full component customization** -- templ components are Go functions compiled into your binary, not runtime templates. You edit them directly.
- **Type safety at compile time** -- errors are caught before deployment, not at runtime.
- **IDE support** -- autocomplete, go-to-definition, and refactoring work out of the box.
- **No template DSL** -- just Go and templ's HTML-like syntax.

The `templsite` library (`github.com/dmoose/templsite`) provides content parsing, asset building, and a dev server. Your binary imports the library, adds its own components, and controls how pages are rendered.

**Repository:** [github.com/dmoose/templsite](https://github.com/dmoose/templsite)
**Module path:** `github.com/dmoose/templsite`

---

## 2. Quick Start

### Install templsite

```bash
git clone https://github.com/dmoose/templsite
cd templsite
make setup   # Downloads Go deps + Tailwind CSS standalone CLI
make build   # Builds the templsite binary
```

### Create a new site

```bash
templsite new mysite --template tailwind
cd mysite
make serve
```

Your site is now running at `http://localhost:8080` with live reload.

### Build for production

```bash
make build         # Compile your site binary (runs templ generate + go build)
./site build       # Generate static files in public/
```

Upload the `public/` directory to any static host.

---

## 3. Project Structure

After running `templsite new mysite --template tailwind`, you get:

```
mysite/
├── main.go              # Your site binary -- build, serve, and new commands
├── go.mod               # Go module; imports github.com/dmoose/templsite
├── config.yaml          # Site configuration
├── Makefile             # Build automation
│
├── components/          # Templ components -- customize these
│   ├── layout/
│   │   ├── base.templ   # HTML shell (DOCTYPE, head, body, stylesheets)
│   │   ├── page.templ   # Default page layout (header + content + footer)
│   │   └── data.go      # Shared data types for components
│   └── ui/
│       ├── header.templ  # Site header with navigation
│       └── footer.templ  # Site footer
│
├── content/             # Markdown files with YAML frontmatter
│   ├── index.md         # Home page (renders at /)
│   └── about.md         # About page (renders at /about/)
│
├── archetypes/          # Templates for ./site new
│   ├── default.md       # Fallback archetype
│   └── posts.md         # Blog post archetype (date, draft, tags)
│
├── assets/
│   ├── css/
│   │   └── app.css      # Tailwind CSS v4 entry point
│   └── js/
│       └── app.js       # Site-wide JavaScript (keep minimal)
│
└── public/              # Generated output (gitignored)
```

### Key files explained

- **main.go** -- The entry point for your site binary. Contains `buildSite()`, `serveSite()`, `newContent()`, and the `renderPages()` function that iterates over pages and renders them using your templ components.
- **components/layout/page.templ** -- The default page layout. Wraps content with your header, footer, and base HTML shell. This is where you control how every page looks.
- **config.yaml** -- Site title, base URL, content directory, asset settings, taxonomies, menus, and build options.
- **Makefile** -- Provides `make serve`, `make build`, `make deploy`, `make prod`, `make new-post`, `make new-page`, and `make clean`.

---

## 4. Creating and Customizing Components

### How templ components work

Templ components are Go functions that return HTML. They live in `.templ` files and are compiled to `*_templ.go` files by `templ generate` (which the Makefile runs automatically).

The scaffolded `components/layout/page.templ` looks like this:

```go
package layout

import (
    "github.com/dmoose/templsite/pkg/content"
    "mysite/components/ui"
)

templ Page(siteTitle, baseURL string, page *content.Page) {
    @Base(page.Title, page.Description, baseURL, pageBody(siteTitle, baseURL, page))
}

templ pageBody(siteTitle, baseURL string, page *content.Page) {
    @ui.Header(siteTitle, baseURL)
    <main class="container mx-auto px-4 py-8 max-w-4xl">
        <article class="prose lg:prose-xl">
            if page.Title != "" {
                <h1>{ page.Title }</h1>
            }
            if page.Description != "" {
                <p class="lead text-base-content/70">{ page.Description }</p>
            }
            <div class="mt-8">
                @templ.Raw(page.Content)
            </div>
        </article>
    </main>
    @ui.Footer(siteTitle)
}
```

### Editing existing components

Open any `.templ` file, make changes, and `make serve` will detect the change, regenerate Go code, recompile, rebuild the site, and reload your browser automatically.

For example, to add a sidebar to the page layout:

```go
templ pageBody(siteTitle, baseURL string, page *content.Page) {
    @ui.Header(siteTitle, baseURL)
    <div class="container mx-auto px-4 py-8 flex gap-8">
        <main class="flex-1 max-w-3xl">
            <article class="prose lg:prose-xl">
                @templ.Raw(page.Content)
            </article>
        </main>
        <aside class="w-64 hidden lg:block">
            <div class="sticky top-8">
                if page.TOC != "" {
                    <nav class="text-sm">
                        <h3 class="font-bold mb-2">On this page</h3>
                        @templ.Raw(page.TOC)
                    </nav>
                }
            </div>
        </aside>
    </div>
    @ui.Footer(siteTitle)
}
```

### Creating new components

Create a new `.templ` file in the appropriate package:

```go
// components/ui/card.templ
package ui

templ Card(title, description string) {
    <div class="rounded-lg border border-gray-200 p-6 shadow-sm hover:shadow-md transition-shadow">
        <h3 class="text-lg font-semibold">{ title }</h3>
        if description != "" {
            <p class="mt-2 text-gray-600">{ description }</p>
        }
    </div>
}
```

Then use it in other components:

```go
@ui.Card("My Title", "Some description")
```

### Layout switching

The `renderPages` function in `main.go` decides which layout each page gets. By default, every page uses `layout.Page()`. To use different layouts:

1. Create a new layout component (e.g., `components/layout/wide.templ`).
2. Update `renderPages` in `main.go` to select layouts based on page metadata:

```go
func renderPages(ctx context.Context, s *site.Site) error {
    for _, page := range s.Pages {
        outputPath := s.GetOutputPath(page.URL)
        os.MkdirAll(filepath.Dir(outputPath), 0755)

        f, err := os.Create(outputPath)
        if err != nil {
            return fmt.Errorf("creating output file: %w", err)
        }

        // Choose layout based on frontmatter
        var component templ.Component
        switch page.Layout {
        case "wide":
            component = layout.Wide(s.Config.Title, s.Config.BaseURL, page)
        default:
            component = layout.Page(s.Config.Title, s.Config.BaseURL, page)
        }

        if err := component.Render(ctx, f); err != nil {
            f.Close()
            return fmt.Errorf("rendering page %s: %w", page.Path, err)
        }
        f.Close()
    }
    return nil
}
```

Then set `layout: "wide"` in a page's frontmatter to use the wide layout.

---

## 5. Tailwind CSS Styling

### Using utility classes in templ files

Apply Tailwind classes directly in your templ components:

```go
templ Hero(title, subtitle string) {
    <section class="min-h-[60vh] flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
        <div class="text-center px-4">
            <h1 class="text-5xl font-bold tracking-tight text-gray-900">{ title }</h1>
            <p class="mt-4 text-xl text-gray-600 max-w-2xl mx-auto">{ subtitle }</p>
        </div>
    </section>
}
```

The Tailwind standalone CLI scans all your `.templ` files (and generated `*_templ.go` files) for class names and produces a CSS file containing only the classes you actually use.

### Customizing app.css

The entry point is `assets/css/app.css`:

```css
/* Tailwind CSS v4 */
@import "tailwindcss";

/* Add your custom styles below */
```

With Tailwind v4, you can define custom theme values using the `@theme` directive:

```css
@import "tailwindcss";

@theme {
  --color-brand-50: oklch(0.97 0.02 210);
  --color-brand-500: oklch(0.6 0.15 210);
  --color-brand-900: oklch(0.3 0.1 210);
  --font-family-display: "Inter", sans-serif;
}
```

These become usable as `bg-brand-500`, `text-brand-900`, `font-display`, etc.

### Tailwind v4 features

Tailwind v4 introduces several changes from v3:

- **No `tailwind.config.js`** -- configuration is done in CSS with `@theme`.
- **`@import "tailwindcss"`** replaces the old `@tailwind` directives.
- **Automatic content detection** -- no need to configure `content` paths.
- **CSS-first configuration** -- theme customization lives in your CSS file, not a JavaScript config.

### Component-specific styles

For styles that apply to rendered Markdown content (where you cannot add classes directly), use Tailwind's `prose` utility:

```go
<article class="prose lg:prose-xl prose-headings:text-gray-900 prose-a:text-blue-600">
    @templ.Raw(page.Content)
</article>
```

### Using templUI or other component libraries

If you are using a Tailwind-based component library like [templUI](https://www.templui.com/), install it as a Go dependency and import its components in your templ files. Since your site is a standalone Go project, you have full control over dependencies. Configure the library's Tailwind plugin in your `app.css` as needed.

---

## 6. Content Authoring

### Markdown files

Create Markdown files in `content/` with YAML frontmatter:

```markdown
---
title: "Building a Blog with Go"
date: 2025-06-15
draft: false
description: "A step-by-step tutorial on building a blog with templsite"
tags: ["golang", "web", "tutorial"]
categories: ["tutorials"]
author: "Jane Doe"
weight: 10
aliases:
  - /old-blog-post-url/
---

# Building a Blog with Go

Your content here with **Markdown** formatting, [links](https://example.com),
and code blocks.

<!--more-->

Content after the `<!--more-->` marker is excluded from the auto-generated summary.
```

### Frontmatter fields

| Field         | Type       | Description                                          |
|---------------|------------|------------------------------------------------------|
| `title`       | string     | Page title                                           |
| `date`        | date       | Publication date (ISO 8601)                          |
| `draft`       | bool       | If true, excluded from build unless `build.drafts` is set |
| `description` | string     | Meta description for SEO and feeds                   |
| `tags`        | []string   | Taxonomy terms for the "tags" taxonomy               |
| `categories`  | []string   | Taxonomy terms for the "categories" taxonomy         |
| `author`      | string     | Author name                                          |
| `weight`      | int        | Manual sort order (lower = first)                    |
| `aliases`     | []string   | Alternative URLs that redirect to this page          |
| `layout`      | string     | Layout name (used by `renderPages` in `main.go`)     |

### Sections

Directory structure under `content/` defines sections:

```
content/
├── index.md              # Home page (section: _root)
├── about.md              # Standalone page (section: _root)
├── blog/
│   ├── _index.md         # Section index page for /blog/
│   ├── first-post.md     # Blog post
│   └── second-post.md    # Blog post
└── docs/
    ├── _index.md         # Section index page for /docs/
    └── getting-started.md
```

The `_index.md` file in a section directory becomes the section index page. Its frontmatter provides the section title and description.

### URL structure

URLs are generated automatically from file paths:

| File path                     | URL              |
|-------------------------------|------------------|
| `content/index.md`           | `/`              |
| `content/about.md`           | `/about/`        |
| `content/blog/_index.md`     | `/blog/`         |
| `content/blog/first-post.md` | `/blog/first-post/` |

### Creating content with archetypes

Use the site binary's `new` command to scaffold content files:

```bash
./site new posts/my-great-idea    # Uses archetypes/posts.md template
./site new about                  # Uses archetypes/default.md template
```

Or with the Makefile:

```bash
make new-post name=my-great-idea
make new-page name=about
```

Archetype files support template variables: `{{.Title}}`, `{{.Date}}`, `{{.Section}}`, `{{.Slug}}`.

---

## 7. Using Page Data

After `Site.Build()` processes content, each `*content.Page` has these computed fields available in your components:

### Summary

Auto-extracted from content -- either everything before `<!--more-->` or the first paragraph:

```go
if page.Summary != "" {
    <p class="text-gray-600">{ page.Summary }</p>
}
```

### Reading time and word count

```go
<span class="text-sm text-gray-500">
    { fmt.Sprintf("%d min read", page.ReadingTime) }
    { fmt.Sprintf(" (%d words)", page.WordCount) }
</span>
```

### Table of contents

The `TOC` field contains rendered HTML of headings found in the page:

```go
if page.TOC != "" {
    <nav class="text-sm border-l-2 border-gray-200 pl-4">
        @templ.Raw(page.TOC)
    </nav>
}
```

### Prev/Next navigation

Pages within the same section are linked via `Prev` and `Next`:

```go
<nav class="flex justify-between mt-12 pt-6 border-t">
    if page.Prev != nil {
        <a href={ templ.URL(page.Prev.URL) } class="text-blue-600 hover:underline">
            { page.Prev.Title }
        </a>
    }
    if page.Next != nil {
        <a href={ templ.URL(page.Next.URL) } class="text-blue-600 hover:underline ml-auto">
            { page.Next.Title }
        </a>
    }
</nav>
```

### Data files

Place YAML or JSON files in the `data/` directory (configurable via `dataDir` in `config.yaml`). They are loaded automatically during build:

```
data/
├── team.yaml        # Accessible as s.GetData("team")
├── nav.json         # Accessible as s.GetData("nav")
└── config/
    └── settings.yaml  # Accessible as s.GetData("config/settings")
```

Access in `renderPages`:

```go
team := s.GetData("team")
```

Or with type safety using the generic helper:

```go
team := site.GetDataAs[[]map[string]any](s, "team")
```

### Menus

Define menus in `config.yaml`:

```yaml
menus:
  main:
    - name: Home
      url: /
      weight: 1
    - name: Blog
      url: /blog/
      weight: 2
    - name: About
      url: /about/
      weight: 3
  footer:
    - name: Privacy
      url: /privacy/
      weight: 1
```

Access in your render function or pass to components:

```go
// Get menu items with active state based on current page URL
mainMenu := s.MenuWithActive("main", page.URL)
```

Each `MenuItem` has `Name`, `URL`, `Weight`, and `Active` (bool) fields. Use `Active` to highlight the current page in navigation.

---

## 8. Taxonomies

### Configuration

In `config.yaml`, list the taxonomies you want to build:

```yaml
taxonomies:
  - tags
  - categories
```

Each taxonomy name corresponds to a frontmatter field. Pages with `tags: ["go", "web"]` will be indexed under the "tags" taxonomy with terms "go" and "web".

### Accessing taxonomy data

In your `renderPages` function or components:

```go
// Get all terms for a taxonomy, sorted by page count (descending)
tagTerms := s.TaxonomyTerms("tags")

// Get terms sorted alphabetically
tagTerms := s.TaxonomyTermsByName("tags")

// Get pages for a specific term
goPages := s.PagesByTaxonomy("tags", "go")

// Convenience: get all tags
allTags := s.Tags()

// Get a specific term
term := s.GetTerm("tags", "go")
// term.Name, term.Slug, term.URL, term.Pages, term.PageCount()
```

### Term URLs

Term URLs follow the pattern `/<taxonomy>/<slug>/`. For example, a tag "Go Programming" becomes `/tags/go-programming/`.

### Custom taxonomies

Any frontmatter field can be a taxonomy. To add a "series" taxonomy:

1. Add to config:
   ```yaml
   taxonomies:
     - tags
     - categories
     - series
   ```

2. Add to page frontmatter:
   ```yaml
   series: ["go-basics"]
   ```

3. Query in your render code:
   ```go
   seriesTerms := s.TaxonomyTerms("series")
   ```

### Rendering taxonomy pages

To generate taxonomy listing pages (e.g., `/tags/` and `/tags/go/`), add rendering logic in `renderPages`:

```go
// Render tag listing pages
for _, term := range s.Tags() {
    outputPath := s.GetOutputPath(term.URL)
    os.MkdirAll(filepath.Dir(outputPath), 0755)

    f, _ := os.Create(outputPath)
    component := layout.TagPage(s.Config.Title, s.Config.BaseURL, term)
    component.Render(ctx, f)
    f.Close()
}
```

---

## 9. Pagination

### Creating a paginator

Use `site.NewPaginator` to paginate any collection of pages:

```go
import "github.com/dmoose/templsite/pkg/site"

// Paginate blog posts, 10 per page, base URL /blog/
blogPages := s.RegularPagesInSection("blog")
paginator := site.NewPaginator(blogPages, 10, "/blog/")
```

### Paginator fields

| Field        | Type              | Description                                   |
|--------------|-------------------|-----------------------------------------------|
| `Items`      | `[]*content.Page` | Pages for the current page number             |
| `TotalItems` | `int`             | Total pages across all paginated pages        |
| `PerPage`    | `int`             | Items per page                                |
| `TotalPages` | `int`             | Total number of paginated pages               |
| `PageNum`    | `int`             | Current page number (1-indexed)               |
| `HasPrev`    | `bool`            | Whether a previous page exists                |
| `HasNext`    | `bool`            | Whether a next page exists                    |
| `PrevURL`    | `string`          | URL of the previous page                      |
| `NextURL`    | `string`          | URL of the next page                          |
| `PageURLs`   | `[]string`        | URLs for all pages                            |
| `BaseURL`    | `string`          | Base URL for this paginated section           |

### Rendering paginated lists

In `renderPages`, iterate over all paginated pages:

```go
blogPages := s.RegularPagesInSection("blog")
pager := site.NewPaginator(blogPages, 10, "/blog/")

for i := 1; i <= pager.TotalPages; i++ {
    p := pager.Page(i)       // Returns paginator configured for page i
    outputPath := s.GetOutputPath(p.PageURLs[i-1])
    os.MkdirAll(filepath.Dir(outputPath), 0755)

    f, _ := os.Create(outputPath)
    component := layout.BlogList(s.Config.Title, s.Config.BaseURL, p)
    component.Render(ctx, f)
    f.Close()
}
```

### URL scheme

- Page 1: `/blog/` (the base URL)
- Page 2: `/blog/page/2/`
- Page 3: `/blog/page/3/`

### Pagination navigation component

```go
templ PaginationNav(pager *site.Paginator) {
    if pager.TotalPages > 1 {
        <nav class="flex items-center justify-center gap-2 mt-8">
            if pager.HasPrev {
                <a href={ templ.URL(pager.PrevURL) }
                   class="px-4 py-2 border rounded hover:bg-gray-50">
                    Previous
                </a>
            }
            for _, num := range pager.Pages() {
                if num == pager.PageNum {
                    <span class="px-4 py-2 bg-blue-600 text-white rounded">
                        { fmt.Sprintf("%d", num) }
                    </span>
                } else {
                    <a href={ templ.URL(pager.URL(num)) }
                       class="px-4 py-2 border rounded hover:bg-gray-50">
                        { fmt.Sprintf("%d", num) }
                    </a>
                }
            }
            if pager.HasNext {
                <a href={ templ.URL(pager.NextURL) }
                   class="px-4 py-2 border rounded hover:bg-gray-50">
                    Next
                </a>
            }
        </nav>
    }
}
```

---

## 10. Feeds and SEO

### Automatic generation

`Site.Build()` automatically generates these files in `public/` (unless overridden by a file in `static/`):

- **feed.xml** -- Atom 1.0 feed of all regular pages
- **sitemap.xml** -- XML sitemap with all pages, sections, and taxonomy terms
- **robots.txt** -- Allows all crawlers, references sitemap.xml

### Custom feeds

For more control, use the feed generation methods directly in `renderPages`:

```go
// RSS 2.0
rss := s.RSS(blogPages, "My Blog", "Latest posts from my blog")
os.WriteFile(filepath.Join(s.OutputDir(), "rss.xml"), []byte(rss), 0644)

// Atom 1.0
atom := s.Atom(blogPages, "My Blog", "Latest posts")
os.WriteFile(filepath.Join(s.OutputDir(), "atom.xml"), []byte(atom), 0644)

// JSON Feed 1.1
jsonFeed := s.JSON(blogPages, "My Blog", "Latest posts from my blog")
os.WriteFile(filepath.Join(s.OutputDir(), "feed.json"), []byte(jsonFeed), 0644)
```

### Linking feeds in HTML

In your `components/layout/base.templ`, add feed discovery links in the `<head>`:

```go
<link rel="alternate" type="application/atom+xml" title="Feed" href={ baseURL + "/feed.xml" }/>
```

### Sitemap

The sitemap includes all regular pages, section index pages, and taxonomy term pages. Each page entry includes the `loc` (URL) and `lastmod` (date from frontmatter) when available.

### Overriding generated files

Place a custom `robots.txt`, `sitemap.xml`, or `feed.xml` in the `static/` directory. Files in `static/` are copied to the output root and take precedence over auto-generated versions.

---

## 11. Environment Configs

### The --env flag

Build for different environments by passing `--env`:

```bash
./site build --env production
```

Or use the Makefile shortcut:

```bash
make prod    # Runs: ./site build --env production
```

### config.production.yaml

Create an environment-specific override file. Only include fields you want to change:

```yaml
# config.production.yaml
baseURL: "https://example.com"

assets:
  minify: true
  fingerprint: true
```

### Params map

Use the `params` map for arbitrary values accessible in your render code:

```yaml
# config.yaml
params:
  analytics_id: ""
  show_drafts_banner: true
```

```yaml
# config.production.yaml
params:
  analytics_id: "G-XXXXXXXXXX"
  show_drafts_banner: false
```

Access in `renderPages`:

```go
analyticsID, _ := s.Config.Params["analytics_id"].(string)
```

### Merge behavior

When `--env production` is used, `config.production.yaml` is loaded and merged on top of `config.yaml`:

- **Scalar fields** (title, baseURL, etc.): overwritten if non-zero in the override.
- **Map fields** (params, menus): key-merged. Keys in the base config that are absent from the override are preserved.
- **Taxonomies**: replaced entirely if the override specifies any.
- **Environment variable `SITE_BASE_URL`**: takes highest precedence over both config files.

---

## 12. Dev Server and Live Reload

### Starting the dev server

```bash
make serve
```

This runs `templ generate`, `go build`, then `./site serve`, starting the dev server at `http://localhost:8080`.

### How it works

1. The server performs an initial `Site.Build()` and `renderPages()`.
2. It watches `content/`, `assets/`, `components/`, and `config.yaml` using fsnotify.
3. When a file changes, it rebuilds content, regenerates assets, re-renders pages, and sends a reload signal over WebSocket.
4. A small JavaScript snippet is automatically injected into every HTML page that connects to `/_live-reload` via WebSocket.
5. On receiving a reload message, the browser refreshes the page.

### Debouncing

Rapid file changes (e.g., from saving multiple files) are debounced with a 500ms minimum interval between rebuilds.

### What triggers a rebuild

Any change to files with these extensions: `.md`, `.templ`, `.css`, `.js`, `.yaml`, `.go`. The watcher monitors directories recursively, skipping hidden directories and output directories (`public/`, `dist/`, `build/`, `_site/`).

### Custom port

Set the `PORT` environment variable or edit `main.go`:

```bash
PORT=3000 ./site serve
```

---

## 13. Deployment

### Static hosting

```bash
make deploy   # Builds binary and generates public/
```

Upload the `public/` directory. Compatible with Netlify, Vercel, GitHub Pages, AWS S3 + CloudFront, or any static file host.

### Production build with environment config

```bash
make prod     # Uses config.production.yaml (minified, correct baseURL)
```

### Docker

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN apk add --no-cache make curl
RUN make build

FROM alpine:latest
COPY --from=builder /app/site /usr/local/bin/site
COPY --from=builder /app/public /var/www
EXPOSE 8080
CMD ["site", "serve"]
```

### CI/CD example (GitHub Actions)

```yaml
- name: Build site
  run: |
    make build
    ./site build --env production

- name: Deploy
  uses: actions/upload-artifact@v4
  with:
    name: site
    path: public/
```

---

## 14. Contributing to templsite

### Code style

**Error wrapping** -- always wrap errors with context:

```go
return fmt.Errorf("processing content: %w", err)
return fmt.Errorf("config file not found: %s\nCreate with: templsite new mysite", path)
```

**Context cancellation** -- check in loops:

```go
for _, page := range pages {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    // process page
}
```

**Structured logging** -- use `log/slog` with contextual fields:

```go
slog.Info("building site", "title", s.Config.Title, "pages", len(s.Pages))
```

### Testing patterns

**Temporary directories** -- use `t.TempDir()` for file operations:

```go
func TestFeature(t *testing.T) {
    tmpDir := t.TempDir()
    // create test files, run test, assert results
}
```

**Table-driven tests** -- for parsers and transformations:

```go
tests := []struct {
    name     string
    input    string
    expected Page
    wantErr  bool
}{
    {"valid", "# Title", Page{...}, false},
    {"empty", "", Page{}, false},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // test implementation
    })
}
```

**Context cancellation tests**:

```go
ctx, cancel := context.WithCancel(context.Background())
cancel()
err := operation(ctx)
if err != context.Canceled {
    t.Errorf("expected context.Canceled, got %v", err)
}
```

### Running tests

```bash
make test                                  # All tests with race detector
go test ./...                              # All tests
go test -v ./pkg/site/                     # Specific package
go test -v -run TestBuild ./pkg/site/      # Specific test
go test -coverprofile=coverage.out ./...   # With coverage
go tool cover -html=coverage.out           # View coverage report
```

### Repository structure

```
templsite/
├── cmd/templsite/           # CLI application
│   ├── main.go              # Entry point
│   ├── commands/            # Command implementations (new, build, serve)
│   └── templates/           # Embedded site templates (go:embed)
│       └── tailwind/        # The "tailwind" template
├── pkg/                     # Public library API
│   ├── site/                # Site management, config, sections, taxonomies, feeds
│   ├── content/             # Markdown parser, Page struct
│   ├── assets/              # CSS, JS, static file pipeline
│   ├── server/              # Dev server, WebSocket live reload
│   └── scaffold/            # Content scaffolding (archetypes)
└── internal/
    └── watch/               # File system watching (fsnotify wrapper)
```

---

## 15. Key Dependencies

| Dependency                       | Purpose                                      |
|----------------------------------|----------------------------------------------|
| `github.com/a-h/templ`          | Type-safe HTML components for Go             |
| `github.com/yuin/goldmark`      | Markdown parser with GFM support             |
| `github.com/tdewolff/minify/v2` | Pure Go CSS/JS minification                  |
| `github.com/fsnotify/fsnotify`  | Cross-platform file system notifications     |
| `github.com/gorilla/websocket`  | WebSocket connections for live reload        |
| `github.com/alecthomas/chroma`  | Syntax highlighting for code blocks          |
| Tailwind CSS v4 standalone CLI   | Utility-first CSS (downloaded automatically) |
| `gopkg.in/yaml.v3`              | YAML parsing for config and frontmatter      |

No Node.js, npm, or webpack is required anywhere in the toolchain.

---

## License

templsite is licensed under the MIT License. See the [LICENSE](../LICENSE) file in the repository for details.
