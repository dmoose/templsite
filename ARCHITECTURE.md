# Architecture: Component Override System

## The Problem

**How do you build a static site generator using Go's templ components where users can customize the components?**

This is the fundamental architectural challenge of templsite. Templ components are Go functions that compile at build time, not runtime templates like Hugo's text/template.

### The Constraints

1. **Templ requires compilation** - Components are Go functions, not interpreted templates
2. **Go imports are compile-time** - You can't dynamically load Go code at runtime
3. **Static output required** - Final output is HTML/CSS/JS files, not a web server
4. **Users need customization** - The whole point is customizable, type-safe components

### The Failed Approach

Initially, we tried to make `pkg/site` do everything including rendering:

```go
// pkg/site/site.go - THIS DOESN'T WORK
import "git.catapulsion.com/templsite/components/layout"

func (s *Site) renderPage(page *content.Page) error {
    component := layout.Page(...)  // Uses templsite's components, not user's!
    component.Render(...)
}
```

**Problem:** When `pkg/site` imports `templsite/components/layout`, that import is baked into the library at compile time. There's no way for user's custom components to override this.

## The Solution: Separation of Concerns

**Key Insight:** Don't put rendering logic in the library. Let each user's site binary do its own rendering using its own components.

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│ User's Site (Go Project)                                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  main.go                                                    │
│  ├─ Imports: mysite/components/layout (THEIR components)   │
│  ├─ Imports: templsite/pkg/site (helper library)           │
│  └─ Contains: renderPages() function using THEIR components│
│                                                             │
│  components/                                                │
│  ├─ layout/base.templ       (user customizes these)        │
│  ├─ layout/page.templ                                       │
│  └─ ui/header.templ                                         │
│                                                             │
│  go build → ./site (binary with user's components baked in)│
│                                                             │
└─────────────────────────────────────────────────────────────┘
                           │
                           │ uses
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ templsite Library                                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  pkg/site/                                                  │
│  ├─ Config loading                                          │
│  ├─ Content parsing (Markdown → Page structs)              │
│  ├─ Asset building (CSS, JS, static files)                 │
│  └─ Helper: GetOutputPath()                                │
│                                                             │
│  pkg/content/                                               │
│  └─ Markdown parser with goldmark                          │
│                                                             │
│  pkg/assets/                                                │
│  └─ CSS (Tailwind), JS (minify), static files              │
│                                                             │
│  pkg/server/                                                │
│  └─ Dev server with live reload (takes RenderFunc callback)│
│                                                             │
│  NO RENDERING CODE - No component imports!                 │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### What Each Piece Does

#### 1. templsite Library (`pkg/*`)

**Purpose:** Provide infrastructure without enforcing specific components.

**Exports:**
- `site.New()` - Load configuration
- `site.Build()` - Parse content + build assets (NO rendering)
- `site.Pages` - Parsed page data
- `site.GetOutputPath()` - Helper for file paths
- `server.New(site, addr, renderFunc)` - Dev server with callback
- `content.Parser` - Markdown parsing
- `assets.Pipeline` - CSS/JS processing

**Does NOT:**
- Import any templ components
- Contain rendering logic
- Know about user's components

#### 2. User's Site Binary (`main.go`)

**Purpose:** Glue code that uses the library and renders with custom components.

**Example:**
```go
package main

import (
    "mysite/components/layout"  // User's components
    "git.catapulsion.com/templsite/pkg/site"
    "git.catapulsion.com/templsite/pkg/server"
)

func main() {
    if os.Args[1] == "build" {
        buildSite()
    } else {
        serveSite()
    }
}

func buildSite(ctx context.Context) error {
    s, _ := site.New("config.yaml")
    
    // Library does parsing and assets
    s.Build(ctx)
    
    // User does rendering with THEIR components
    renderPages(ctx, s)
}

func renderPages(ctx context.Context, s *site.Site) error {
    for _, page := range s.Pages {
        outputPath := s.GetOutputPath(page.URL)
        f, _ := os.Create(outputPath)
        
        // Uses user's layout.Page, not templsite's!
        component := layout.Page(s.Config.Title, s.Config.BaseURL, page)
        component.Render(ctx, f)
    }
}
```

#### 3. Dev Server Callback Pattern

The development server needs to rebuild when files change, but it can't import user's components. Solution: **callback function**.

```go
// pkg/server/server.go
type RenderFunc func(ctx context.Context, s *site.Site) error

func New(s *site.Site, addr string, renderFunc RenderFunc) (*Server, error) {
    return &Server{
        site: s,
        renderFunc: renderFunc,  // User's rendering function
    }
}

func (s *Server) handleFileChanges(ctx context.Context) {
    // When files change:
    s.site.Build(ctx)         // Parse + assets
    s.renderFunc(ctx, s.site) // User's rendering
    s.notifyReload()          // Browser refresh
}
```

User passes their render function:
```go
srv, _ := server.New(s, "localhost:8080", renderPages)
srv.Serve(ctx)
```

## The User Experience

From the user's perspective, this complexity is hidden:

```bash
# 1. Scaffold new site
templsite new mysite --template minimal --templsite-path /path/to/templsite

# 2. Enter directory
cd mysite

# 3. Just use make
make serve    # Runs: templ generate && go build && ./mysite serve
make build    # Runs: templ generate && go build
make deploy   # Runs: make build && ./mysite build
```

The Makefile hides:
- `templ generate` - Generates `*_templ.go` from `.templ` files
- `go build` - Compiles user's binary with their components
- `./mysite serve` - Runs their binary (not templsite binary!)

### What Users Can Customize

Everything in their site directory:

1. **Components** (`components/`)
   - Full control over all templ components
   - Type-safe with IDE support
   - Compile-time validation

2. **Content** (`content/`)
   - Markdown files with YAML frontmatter
   - Standard static site generator workflow

3. **Assets** (`assets/`)
   - CSS with Tailwind + DaisyUI
   - JavaScript utilities
   - Static files (images, fonts, etc.)

4. **Configuration** (`config.yaml`)
   - Site title, baseURL, paths, etc.

5. **Build Logic** (`main.go`)
   - Can add custom data processing
   - Custom rendering logic
   - Additional build steps

## Why This Works

### For Users

✅ **Looks simple** - Just run `make serve`
✅ **Full customization** - Modify any component
✅ **Type safety** - Components are Go functions
✅ **IDE support** - Autocomplete, refactoring, compile-time errors
✅ **Go ecosystem** - Can use any Go library

### For templsite

✅ **Clean separation** - Library provides tools, user provides components
✅ **No magic** - Everything is explicit Go code
✅ **Maintainable** - Library doesn't know about user's components
✅ **Extensible** - Users can add their own packages/logic

### For the Go/Templ Model

✅ **Respects compilation model** - Doesn't fight Go's design
✅ **Uses templ correctly** - Components are real Go functions
✅ **No runtime hacks** - No dynamic loading, reflection, or code generation tricks

## Comparison to Hugo

| Aspect | Hugo | templsite |
|--------|------|-----------|
| Templates | Runtime text/template | Compile-time Go functions |
| Binary | One universal binary | Each site compiles its own |
| Customization | Template files | Go code with templ |
| Type Safety | No | Yes (compile-time) |
| IDE Support | Limited | Full Go tooling |
| Speed | Fast | Fast (compiled Go) |
| Extensibility | Limited to Hugo's features | Full Go ecosystem |

Hugo can use one binary because text templates are parsed at runtime. templsite requires compilation but gains type safety and IDE support.

## Trade-offs

### Advantages of This Approach

1. **True type safety** - Errors caught at compile time, not when users visit the site
2. **IDE support** - Autocomplete, go-to-definition, refactoring all work
3. **No template DSL** - Just Go and templ, no new language to learn
4. **Full Go power** - Can use any Go library, add custom logic
5. **Clean architecture** - Library and user code properly separated

### Disadvantages

1. **Requires Go** - Users need Go installed (but not Node.js!)
2. **Compilation step** - Changes require rebuild (but Makefile + live reload hides this)
3. **More moving parts** - Each site is a Go module with dependencies
4. **Not "just run templsite"** - Each site has its own binary

### Why We Accept the Trade-offs

**Target Audience:** Developers who value type safety and want to avoid the JavaScript ecosystem quagmire.

These users:
- Already comfortable with compiled languages
- Understand the value of compile-time errors
- Prefer `go build` to `npm install`
- Want IDE support and refactoring
- Are building serious sites, not quick prototypes

For this audience, the compilation requirement is a feature, not a bug. It's what enables the type safety and IDE support they want.

## Implementation Details

### Template Substitution

When `templsite new mysite` creates a site, it needs to customize the template files for that specific site. Specifically, imports need the right module name:

```go
// Template file before substitution
import "{{.ModulePath}}/components/layout"

// After substitution for site "mysite"
import "mysite/components/layout"
```

This is handled in `cmd/templsite/commands/new.go`:

```go
templateData := map[string]string{
    "ModulePath": siteName,
    "SiteName":   siteName,
}
copyFSWithTemplates(templateFS, ".", sitePath, templateData)
```

The `copyFSWithTemplates` function uses Go's `text/template` to process `.go` and `.templ` files during copying.

### Go Module Setup

Each site needs:

1. **go.mod** - Created with `go mod init sitename`
2. **Dependencies** - Added with `go mod tidy`
3. **Replace directive** - For local development:

```go
// go.mod
module mysite

require git.catapulsion.com/templsite v0.x.x

// Local development - remove for production
replace git.catapulsion.com/templsite => /path/to/templsite
```

The replace directive lets users develop against local templsite. For production, remove it and publish templsite to a real module repository.

### Build Process

1. User runs `make serve`
2. Makefile runs `templ generate` - Generates `*_templ.go` files
3. Makefile runs `go build` - Compiles user's binary with their components
4. Binary runs and uses its built-in components

The key: **User's binary imports user's components**, not templsite's components.

## Future Considerations

### Publishing Sites

When templsite is published to a real module repository (not local development):

```bash
# User creates site
templsite new mysite --template minimal

# Edit go.mod to remove replace directive
# Or templsite does this automatically for production

# Build
cd mysite
go build

# Deploy
./mysite build
# Upload public/ to hosting
```

### Component Registry (Stage 10)

Even with this architecture, a component registry (templ-store) can work:

```bash
# Install a component
cd mysite
templsite components add navbar-fancy

# This copies the component into mysite/components/
# User's next build includes it
```

Components are still just `.templ` files that get compiled into the user's binary.

### Multiple Layouts

Users can add as many layouts as they want:

```go
// mysite/components/layout/blog.templ
package layout

templ Blog(title string, page *content.Page) {
    @Base(title, page.Description, baseURL, blogBody(page))
}
```

Then in their render function:

```go
func renderPages(ctx context.Context, s *site.Site) error {
    for _, page := range s.Pages {
        var component templ.Component
        
        if page.Layout == "blog" {
            component = layout.Blog(s.Config.Title, page)
        } else {
            component = layout.Page(s.Config.Title, s.Config.BaseURL, page)
        }
        
        // ... render component
    }
}
```

Full flexibility, full type safety.

## Conclusion

**The solution to customizable templ components: Don't try to make them runtime-customizable. Make each site its own Go program.**

This architecture:
- Respects Go's compilation model
- Respects templ's design (components are functions)
- Gives users full control and type safety
- Keeps the library clean and maintainable
- Provides good UX through the Makefile abstraction

It's not a hack or workaround. It's the correct architecture for a Go-based, type-safe static site generator using templ components.

The "complexity" of each site being a Go project is actually a strength: it gives users the full power of Go and the Go ecosystem. For developers who want type safety and modern tooling without the JavaScript ecosystem, this is the right trade-off.