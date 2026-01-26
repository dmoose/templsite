package site

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegrationFullBuild(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure
	contentDir := filepath.Join(tmpDir, "content")
	assetsDir := filepath.Join(tmpDir, "assets")
	assetsCSSDir := filepath.Join(assetsDir, "css")
	assetsJSDir := filepath.Join(assetsDir, "js")

	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}
	if err := os.MkdirAll(assetsCSSDir, 0755); err != nil {
		t.Fatalf("failed to create css dir: %v", err)
	}
	if err := os.MkdirAll(assetsJSDir, 0755); err != nil {
		t.Fatalf("failed to create js dir: %v", err)
	}

	// Create content files
	indexContent := `---
title: "Home Page"
description: "Welcome to our site"
---

# Welcome

This is the home page.
`
	if err := os.WriteFile(filepath.Join(contentDir, "index.md"), []byte(indexContent), 0644); err != nil {
		t.Fatalf("failed to write index.md: %v", err)
	}

	aboutContent := `---
title: "About Us"
description: "Learn more about us"
layout: "page"
---

# About

We build great software.
`
	if err := os.WriteFile(filepath.Join(contentDir, "about.md"), []byte(aboutContent), 0644); err != nil {
		t.Fatalf("failed to write about.md: %v", err)
	}

	// Create blog directory and post
	blogDir := filepath.Join(contentDir, "blog")
	if err := os.MkdirAll(blogDir, 0755); err != nil {
		t.Fatalf("failed to create blog dir: %v", err)
	}

	blogContent := `---
title: "First Post"
description: "Our first blog post"
date: 2025-01-15
---

# First Post

This is our first blog post.
`
	if err := os.WriteFile(filepath.Join(blogDir, "first-post.md"), []byte(blogContent), 0644); err != nil {
		t.Fatalf("failed to write blog post: %v", err)
	}

	// Create CSS file
	cssContent := `@import "tailwindcss";

body {
  font-family: sans-serif;
}
`
	if err := os.WriteFile(filepath.Join(assetsCSSDir, "app.css"), []byte(cssContent), 0644); err != nil {
		t.Fatalf("failed to write css: %v", err)
	}

	// Create JS file
	jsContent := `console.log('Hello from templsite');

function init() {
  console.log('Site initialized');
}

init();
`
	if err := os.WriteFile(filepath.Join(assetsJSDir, "app.js"), []byte(jsContent), 0644); err != nil {
		t.Fatalf("failed to write js: %v", err)
	}

	// Create config
	config := &Config{
		Title:     "Test Site",
		BaseURL:   "https://example.com",
		OutputDir: "public",
		Content: ContentConfig{
			Dir:           "content",
			DefaultLayout: "page",
		},
		Assets: AssetsConfig{
			InputDir:  "assets",
			OutputDir: "assets",
			Minify:    false,
		},
	}

	// Create site
	site := NewWithConfig(config)
	site.SetBaseDir(tmpDir)

	// Build
	ctx := context.Background()
	if err := site.Build(ctx); err != nil {
		t.Fatalf("build failed: %v", err)
	}

	// Verify output structure
	publicDir := filepath.Join(tmpDir, "public")

	// Check HTML files exist
	htmlFiles := []struct {
		path        string
		mustContain []string
	}{
		{
			path: filepath.Join(publicDir, "index.html"),
			mustContain: []string{
				"<!doctype html>",
				"<title>Home Page</title>",
				"Welcome to our site",
				"Welcome",
				"This is the home page",
			},
		},
		{
			path: filepath.Join(publicDir, "about", "index.html"),
			mustContain: []string{
				"<!doctype html>",
				"<title>About Us</title>",
				"Learn more about us",
				"About",
				"We build great software",
			},
		},
		{
			path: filepath.Join(publicDir, "blog", "first-post", "index.html"),
			mustContain: []string{
				"<!doctype html>",
				"<title>First Post</title>",
				"Our first blog post",
				"First Post",
				"This is our first blog post",
				"January 15, 2025", // Date formatting
			},
		},
	}

	for _, file := range htmlFiles {
		t.Run(file.path, func(t *testing.T) {
			if _, err := os.Stat(file.path); os.IsNotExist(err) {
				t.Fatalf("expected file does not exist: %s", file.path)
			}

			content, err := os.ReadFile(file.path)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}

			contentStr := string(content)
			for _, expected := range file.mustContain {
				if !strings.Contains(contentStr, expected) {
					t.Errorf("file missing expected content: %q\nFile preview:\n%s",
						expected, contentStr[:min(500, len(contentStr))])
				}
			}

			// All pages should have CSS and JS links
			if !strings.Contains(contentStr, "/assets/css/main.css") {
				t.Error("missing CSS link")
			}
			if !strings.Contains(contentStr, "/assets/js/main.js") {
				t.Error("missing JS link")
			}

			// All pages should have header and footer
			if !strings.Contains(contentStr, "Test Site") {
				t.Error("missing site title in header")
			}
			if !strings.Contains(contentStr, "Copyright") {
				t.Error("missing copyright in footer")
			}
		})
	}

	// Verify CSS was processed
	cssOutput := filepath.Join(publicDir, "assets", "css", "main.css")
	if _, err := os.Stat(cssOutput); os.IsNotExist(err) {
		t.Error("CSS output file not created")
	} else {
		css, _ := os.ReadFile(cssOutput)
		t.Logf("CSS output size: %d bytes", len(css))
	}

	// Verify JS was processed
	jsOutput := filepath.Join(publicDir, "assets", "js", "main.js")
	if _, err := os.Stat(jsOutput); os.IsNotExist(err) {
		t.Error("JS output file not created")
	} else {
		js, _ := os.ReadFile(jsOutput)
		jsStr := string(js)
		if !strings.Contains(jsStr, "Hello from templsite") {
			t.Error("JS content not preserved")
		}
		if !strings.Contains(jsStr, "init()") {
			t.Error("JS function call not preserved")
		}
	}

	// Verify proper URL structure (clean URLs with index.html)
	expectedPaths := []string{
		filepath.Join(publicDir, "index.html"),
		filepath.Join(publicDir, "about", "index.html"),
		filepath.Join(publicDir, "blog", "first-post", "index.html"),
		filepath.Join(publicDir, "assets", "css", "main.css"),
		filepath.Join(publicDir, "assets", "js", "main.js"),
	}

	for _, path := range expectedPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected path does not exist: %s", path)
		}
	}

	t.Logf("Build created %d pages successfully", len(site.Pages))
}

func TestIntegrationBuildWithMinification(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup directories
	contentDir := filepath.Join(tmpDir, "content")
	assetsDir := filepath.Join(tmpDir, "assets")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(assetsDir, "js"), 0755); err != nil {
		t.Fatalf("failed to create js dir: %v", err)
	}

	// Create simple content
	content := `---
title: "Test"
---
# Test
`
	if err := os.WriteFile(filepath.Join(contentDir, "index.md"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write content: %v", err)
	}

	// Create JS with lots of whitespace and comments
	js := `
// This is a comment
function hello() {
    console.log( "Hello World" );
    return true;
}

// Another comment
hello();
`
	if err := os.WriteFile(filepath.Join(assetsDir, "js", "app.js"), []byte(js), 0644); err != nil {
		t.Fatalf("failed to write js: %v", err)
	}

	// Create config with minification enabled
	config := &Config{
		Title:     "Test Site",
		BaseURL:   "https://example.com",
		OutputDir: "public",
		Content: ContentConfig{
			Dir:           "content",
			DefaultLayout: "page",
		},
		Assets: AssetsConfig{
			InputDir:  "assets",
			OutputDir: "assets",
			Minify:    true,
		},
	}

	site := NewWithConfig(config)
	site.SetBaseDir(tmpDir)

	ctx := context.Background()
	if err := site.Build(ctx); err != nil {
		t.Fatalf("build with minification failed: %v", err)
	}

	// Verify JS was minified
	jsOutput := filepath.Join(tmpDir, "public", "assets", "js", "main.js")
	minified, err := os.ReadFile(jsOutput)
	if err != nil {
		t.Fatalf("failed to read minified js: %v", err)
	}

	minifiedStr := string(minified)

	// Minified JS should be smaller
	if len(minifiedStr) >= len(js) {
		t.Error("minified JS is not smaller than source")
	}

	// Should still contain the actual code
	if !strings.Contains(minifiedStr, "hello") {
		t.Error("minified JS missing function name")
	}

	// Should not contain comments
	if strings.Contains(minifiedStr, "// This is a comment") {
		t.Error("minified JS still contains comments")
	}

	t.Logf("Original JS: %d bytes, Minified: %d bytes (%.1f%% reduction)",
		len(js), len(minifiedStr), 100*(1-float64(len(minifiedStr))/float64(len(js))))
}
