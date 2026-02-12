# Contributing to templsite

Thanks for your interest in contributing! Here's how to get started.

## Development Setup

```bash
git clone https://github.com/dmoose/templsite.git
cd templsite
make setup    # Downloads dependencies + Tailwind CLI
make build    # Build the binary
make test     # Run tests with race detector
```

### Prerequisites

- Go 1.25+
- [templ](https://templ.guide/) CLI (`go install github.com/a-h/templ/cmd/templ@latest`)

## Making Changes

1. Fork the repo and create a feature branch from `main`
2. Make your changes
3. Run `make fmt` and `make vet`
4. Run `make test` to ensure all tests pass
5. Submit a pull request

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Wrap errors with context: `fmt.Errorf("doing something: %w", err)`
- Use `log/slog` for logging with contextual fields
- Check context cancellation in loops
- Write table-driven tests where appropriate
- Use `t.TempDir()` for tests that need filesystem access

## What to Contribute

- Bug fixes with test cases
- Documentation improvements
- New content features (sections, taxonomies, etc.)
- Template improvements (tailwind or fastatic)
- Performance improvements with benchmarks

## Third-Party Content and Redistributability

When contributing templates, assets, snippets, icons, fonts, or images:
- Only include content you created, content in the public domain, or content under a license compatible with MIT distribution.
- Add attribution and license details in the PR description when required by the source license.
- Do not copy proprietary themes/assets from paid or restricted sources.
- Keep generated project templates safe for downstream redistribution by default.

## Copyright and License Metadata

- The repository license is MIT (`LICENSE`), which applies to contributions unless stated otherwise.
- Go source files should include:
  - `// Copyright (c) 2025-2026 Catapulsion LLC and contributors`
  - `// SPDX-License-Identifier: MIT`
- Non-code files are covered by REUSE metadata in `.reuse/dep5` unless a file has a different declared license.
- If you add third-party content, include the correct SPDX identifier and attribution for that content.

## Reporting Issues

Open an issue on GitHub with:
- What you expected to happen
- What actually happened
- Steps to reproduce
- Go version and OS

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
