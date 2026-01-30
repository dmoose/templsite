# My Site

Built with [templsite](https://github.com/yourorg/templsite) - a type-safe static site generator using Go and templ.

## Quick Start

```bash
make serve    # Start development server
make build    # Build the site binary
make deploy   # Generate static files
```

## Project Structure

```
.
├── main.go              # Site binary entry point
├── config.yaml          # Site configuration
├── content/             # Markdown content
├── components/          # Templ components (customize these!)
│   ├── layout/         # Page layouts
│   └── ui/             # UI components
├── assets/             # CSS, JS, static files
├── bin/                # daisyUI bundle files
└── public/             # Generated output (gitignored)
```

## Customization

### Content
Edit Markdown files in `content/` - they support frontmatter.

### Components
Customize templ components in `components/` - these compile to Go functions with full type safety!

### Styling
Edit `assets/css/app.css` or use DaisyUI component classes directly in your templ files.

## Commands

Your site is a Go program. After modifying components:

```bash
make build   # Regenerates templ and rebuilds binary
./site serve # Start server
./site build # Generate static files
```

## Learn More

- [templsite](https://github.com/yourorg/templsite)
- [templ](https://templ.guide/)
- [DaisyUI](https://daisyui.com/components/)
- [Tailwind CSS](https://tailwindcss.com/)
