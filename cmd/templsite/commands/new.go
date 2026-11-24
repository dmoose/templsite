// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package commands

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/dmoose/templsite/cmd/templsite/templates"
)

// NewNewCmd creates the "new" command
func NewNewCmd(ctx context.Context) *cobra.Command {
	var (
		templateName  string
		verbose       bool
		templsitePath string
	)

	cmd := &cobra.Command{
		Use:   "new <site-path>",
		Short: "Create a new site from a template",
		Long: `Create a new site from a template. The site path will be used as the
Go module name and directory name.`,
		Example: `  templsite new mysite
  templsite new mysite --template fastatic
  templsite new mysite --template tailwind --templsite-path ../templsite`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runNew(ctx, args[0], templateName, verbose, templsitePath)
		},
	}

	cmd.Flags().StringVar(&templateName, "template", "tailwind", "template to use (tailwind, fastatic)")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "enable verbose logging")
	cmd.Flags().StringVar(&templsitePath, "templsite-path", "", "path to local templsite for development")

	return cmd
}

func runNew(ctx context.Context, sitePath, templateName string, verbose bool, templsitePath string) error {
	// Setup logging
	if verbose {
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
		slog.SetDefault(logger)
	}

	// Validate template exists
	availableTemplates := templates.ListTemplates()
	templateValid := slices.Contains(availableTemplates, templateName)
	if !templateValid {
		return fmt.Errorf("unknown template %q. Available templates: %s", templateName, strings.Join(availableTemplates, ", "))
	}

	// Check if path already exists
	if _, err := os.Stat(sitePath); err == nil {
		return fmt.Errorf("path %q already exists", sitePath)
	}

	slog.Info("creating new site", "path", sitePath, "template", templateName)

	// Get template filesystem
	templateFS, err := templates.GetTemplate(templateName)
	if err != nil {
		return fmt.Errorf("loading template: %w", err)
	}

	// Create site directory
	if err := os.MkdirAll(sitePath, 0o755); err != nil {
		return fmt.Errorf("creating site directory: %w", err)
	}

	// Determine module name from path
	moduleName := filepath.Base(sitePath)
	if strings.Contains(sitePath, "/") {
		moduleName = sitePath
	}

	// Copy template files recursively with substitutions
	templateData := map[string]string{
		"ModulePath": moduleName,
		"SiteName":   filepath.Base(sitePath),
	}
	if err := copyFSWithTemplates(templateFS, ".", sitePath, templateData); err != nil {
		// Clean up on failure
		_ = os.RemoveAll(sitePath)
		return fmt.Errorf("copying template files: %w", err)
	}

	// Get absolute path for display
	absPath, err := filepath.Abs(sitePath)
	if err != nil {
		absPath = sitePath
	}

	if err := setUpModule(ctx, sitePath, moduleName, templsitePath); err != nil {
		return err
	}

	// Success message
	fmt.Println()
	fmt.Println("✓ Created new site at:", absPath)
	fmt.Println("  Template:", templateName)
	fmt.Println("  Module:", moduleName)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", sitePath)
	if templsitePath != "" {
		fmt.Println("  make serve")
	} else {
		fmt.Println("  # Setup dependencies first, then:")
		fmt.Println("  make serve")
	}
	fmt.Println()

	return nil
}

// copyFSWithTemplates recursively copies files from an embedded filesystem with template processing
func copyFSWithTemplates(sourceFS fs.FS, sourcePath, destPath string, data map[string]string) error {
	return fs.WalkDir(sourceFS, sourcePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return fmt.Errorf("calculating relative path: %w", err)
		}
		targetPath := filepath.Join(destPath, relPath)

		// Strip .tmpl suffix from destination if present
		targetPath = strings.TrimSuffix(targetPath, ".tmpl")

		if d.IsDir() {
			// Create directory
			slog.Debug("creating directory", "path", targetPath)
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return fmt.Errorf("creating directory %q: %w", targetPath, err)
			}
			return nil
		}

		// Copy file with template processing for .templ and .go files
		slog.Debug("copying file", "from", path, "to", targetPath)

		// Read source file
		sourceFile, err := sourceFS.Open(path)
		if err != nil {
			return fmt.Errorf("opening source file %q: %w", path, err)
		}
		defer func() { _ = sourceFile.Close() }()

		content, err := io.ReadAll(sourceFile)
		if err != nil {
			return fmt.Errorf("reading file %q: %w", path, err)
		}

		// Process templates for .templ, .go, and .tmpl files
		if strings.HasSuffix(path, ".templ") || strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".tmpl") {
			tmpl, err := template.New("file").Parse(string(content))
			if err != nil {
				return fmt.Errorf("parsing template %q: %w", path, err)
			}
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, data); err != nil {
				return fmt.Errorf("executing template %q: %w", path, err)
			}
			content = buf.Bytes()
		}

		// Create destination file
		destFile, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("creating destination file %q: %w", targetPath, err)
		}
		defer func() { _ = destFile.Close() }()

		// Write contents
		if _, err := destFile.Write(content); err != nil {
			return fmt.Errorf("writing file contents: %w", err)
		}

		// Set writable permissions — embedded FS files are read-only
		// but scaffolded files should be user-editable
		if info, err := d.Info(); err == nil {
			mode := info.Mode() | 0o200 // ensure owner-writable
			if err := os.Chmod(targetPath, mode); err != nil {
				slog.Warn("failed to set file permissions", "path", targetPath, "error", err)
			}
		}

		return nil
	})
}

// setUpModule makes the scaffolded directory a Go module: templsite
// (through a replace when templsitePath is set) is resolved first, then
// templ is added as a tool, so templ is the version templsite requires,
// and the components are generated with it.
func setUpModule(ctx context.Context, sitePath, moduleName, templsitePath string) error {
	slog.Info("initializing Go module", "module", moduleName)

	// Initialize Go module
	if err := runGo(ctx, sitePath, "mod", "init", moduleName); err != nil {
		slog.Warn("go mod init failed, continuing anyway", "error", err)
	}

	// Add replace directive if local templsite path is provided
	if templsitePath != "" {
		absTemplsitePath, err := filepath.Abs(templsitePath)
		if err != nil {
			return fmt.Errorf("resolving templsite path: %w", err)
		}

		slog.Info("adding replace directive for local development", "path", absTemplsitePath)
		goModPath := filepath.Join(sitePath, "go.mod")

		// Read existing go.mod
		goModContent, err := os.ReadFile(goModPath)
		if err != nil {
			return fmt.Errorf("reading go.mod: %w", err)
		}

		// Append replace directive
		replaceDirective := fmt.Sprintf("\n// Local development - remove for production\nreplace github.com/dmoose/templsite => %s\n", absTemplsitePath)
		goModContent = append(goModContent, []byte(replaceDirective)...)
		if err := os.WriteFile(goModPath, goModContent, 0o644); err != nil { //nolint:gosec // the site path is the CLI argument
			return fmt.Errorf("writing go.mod with replace directive: %w", err)
		}
	}

	// Resolve templsite first: templ then comes from templsite's
	// requirement. With the tool line already present, tidy would look the
	// tool up on its own and take the newest templ release.
	slog.Info("downloading dependencies")
	if err := runGo(ctx, sitePath, "mod", "tidy"); err != nil {
		return fmt.Errorf("resolving templsite: %w", err)
	}

	// templ as a tool with no version of its own: it is the templ already
	// in the module graph, and `go tool templ` builds that version.
	if err := runGo(ctx, sitePath, "mod", "edit", "-tool=github.com/a-h/templ/cmd/templ"); err != nil {
		return fmt.Errorf("adding templ as a tool: %w", err)
	}
	if err := runGo(ctx, sitePath, "mod", "tidy"); err != nil {
		return fmt.Errorf("resolving templ: %w", err)
	}

	slog.Info("generating templ components")
	if err := runGo(ctx, sitePath, "tool", "templ", "generate"); err != nil {
		return fmt.Errorf("generating templ components: %w", err)
	}

	// The generated code imports templ directly
	if err := runGo(ctx, sitePath, "mod", "tidy"); err != nil {
		return fmt.Errorf("tidying after templ generate: %w", err)
	}
	return nil
}

// runGo runs a go command in a specific directory
func runGo(ctx context.Context, dir string, args ...string) error {
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
