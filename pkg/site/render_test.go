package site

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"git.catapulsion.com/templsite/pkg/content"
)

func TestRenderPage(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test config
	config := &Config{
		Title:     "Test Site",
		BaseURL:   "http://localhost:8080",
		OutputDir: "public",
		Content: ContentConfig{
			Dir:           "content",
			DefaultLayout: "page",
		},
		Assets: AssetsConfig{
			InputDir:  "assets",
			OutputDir: "assets",
		},
	}

	site := &Site{
		Config:  config,
		baseDir: tmpDir,
	}

	// Create test page
	page := &content.Page{
		Path:        "content/test.md",
		Title:       "Test Page",
		Description: "This is a test page",
		Content:     "<p>Hello World</p>",
		URL:         "/test/",
		Layout:      "page",
		Date:        time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	// Render the page
	ctx := context.Background()
	err := site.renderPage(ctx, page)
	if err != nil {
		t.Fatalf("renderPage failed: %v", err)
	}

	// Verify output file exists
	outputPath := filepath.Join(tmpDir, "public", "test", "index.html")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("output file not created: %s", outputPath)
	}

	// Read and verify content
	html, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	htmlStr := string(html)

	// Debug: print what was actually rendered
	t.Logf("Rendered HTML length: %d bytes", len(html))
	t.Logf("First 500 chars: %s", htmlStr[:min(500, len(htmlStr))])

	// Verify HTML structure (case-insensitive for DOCTYPE)
	if !strings.Contains(strings.ToLower(htmlStr), "<!doctype html>") {
		t.Errorf("output missing DOCTYPE. Full output:\n%s", htmlStr)
	}

	if !strings.Contains(htmlStr, "<title>Test Page</title>") {
		t.Error("output missing page title")
	}

	if !strings.Contains(htmlStr, "Test Page") {
		t.Error("output missing page title in content")
	}

	if !strings.Contains(htmlStr, "Hello World") {
		t.Error("output missing page content")
	}

	if !strings.Contains(htmlStr, "This is a test page") {
		t.Error("output missing page description")
	}

	// Verify CSS link
	if !strings.Contains(htmlStr, "/assets/css/main.css") {
		t.Error("output missing CSS link")
	}

	// Verify JS link
	if !strings.Contains(htmlStr, "/assets/js/main.js") {
		t.Error("output missing JS link")
	}
}

func TestRenderPages(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		Title:     "Test Site",
		BaseURL:   "http://localhost:8080",
		OutputDir: "public",
		Content: ContentConfig{
			Dir:           "content",
			DefaultLayout: "page",
		},
		Assets: AssetsConfig{
			InputDir:  "assets",
			OutputDir: "assets",
		},
	}

	site := &Site{
		Config:  config,
		baseDir: tmpDir,
		Pages: []*content.Page{
			{
				Path:    "content/index.md",
				Title:   "Home",
				Content: "<p>Home page</p>",
				URL:     "/",
				Layout:  "page",
			},
			{
				Path:    "content/about.md",
				Title:   "About",
				Content: "<p>About page</p>",
				URL:     "/about/",
				Layout:  "page",
			},
			{
				Path:    "content/blog/post.md",
				Title:   "Blog Post",
				Content: "<p>Blog content</p>",
				URL:     "/blog/post/",
				Layout:  "page",
			},
		},
	}

	// Render all pages
	ctx := context.Background()
	err := site.renderPages(ctx)
	if err != nil {
		t.Fatalf("renderPages failed: %v", err)
	}

	// Verify all output files exist
	expectedFiles := []struct {
		path  string
		title string
	}{
		{filepath.Join(tmpDir, "public", "index.html"), "Home"},
		{filepath.Join(tmpDir, "public", "about", "index.html"), "About"},
		{filepath.Join(tmpDir, "public", "blog", "post", "index.html"), "Blog Post"},
	}

	for _, expected := range expectedFiles {
		if _, err := os.Stat(expected.path); os.IsNotExist(err) {
			t.Errorf("output file not created: %s", expected.path)
			continue
		}

		// Verify content
		html, err := os.ReadFile(expected.path)
		if err != nil {
			t.Errorf("failed to read %s: %v", expected.path, err)
			continue
		}

		if !strings.Contains(string(html), expected.title) {
			t.Errorf("file %s missing expected title %s", expected.path, expected.title)
		}
	}
}

func TestRenderPageWithContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		Title:     "Test Site",
		BaseURL:   "http://localhost:8080",
		OutputDir: "public",
		Content: ContentConfig{
			Dir:           "content",
			DefaultLayout: "page",
		},
		Assets: AssetsConfig{
			InputDir:  "assets",
			OutputDir: "assets",
		},
	}

	site := &Site{
		Config:  config,
		baseDir: tmpDir,
	}

	page := &content.Page{
		Path:    "content/test.md",
		Title:   "Test",
		Content: "<p>Test</p>",
		URL:     "/test/",
		Layout:  "page",
	}

	// Cancel context immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := site.renderPage(ctx, page)
	if err == nil {
		t.Error("expected error with cancelled context, got nil")
	}
}

func TestGetOutputPath(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		OutputDir: "public",
	}

	site := &Site{
		Config:  config,
		baseDir: tmpDir,
	}

	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "root URL",
			url:      "/",
			expected: filepath.Join(tmpDir, "public", "index.html"),
		},
		{
			name:     "simple page",
			url:      "/about/",
			expected: filepath.Join(tmpDir, "public", "about", "index.html"),
		},
		{
			name:     "nested page",
			url:      "/blog/post/",
			expected: filepath.Join(tmpDir, "public", "blog", "post", "index.html"),
		},
		{
			name:     "deeply nested",
			url:      "/docs/guide/install/",
			expected: filepath.Join(tmpDir, "public", "docs", "guide", "install", "index.html"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := site.getOutputPath(tt.url)
			if result != tt.expected {
				t.Errorf("getOutputPath(%q) = %q, want %q", tt.url, result, tt.expected)
			}
		})
	}
}

func TestRenderPageCreatesOutputDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		Title:     "Test Site",
		BaseURL:   "http://localhost:8080",
		OutputDir: "public",
		Content: ContentConfig{
			Dir:           "content",
			DefaultLayout: "page",
		},
		Assets: AssetsConfig{
			InputDir:  "assets",
			OutputDir: "assets",
		},
	}

	site := &Site{
		Config:  config,
		baseDir: tmpDir,
	}

	// Test nested path that doesn't exist
	page := &content.Page{
		Path:    "content/deeply/nested/page.md",
		Title:   "Nested Page",
		Content: "<p>Nested content</p>",
		URL:     "/deeply/nested/page/",
		Layout:  "page",
	}

	ctx := context.Background()
	err := site.renderPage(ctx, page)
	if err != nil {
		t.Fatalf("renderPage failed: %v", err)
	}

	// Verify directory was created
	outputPath := filepath.Join(tmpDir, "public", "deeply", "nested", "page", "index.html")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("output file not created: %s", outputPath)
	}
}

func TestRenderPageWithDifferentLayouts(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		Title:     "Test Site",
		BaseURL:   "http://localhost:8080",
		OutputDir: "public",
		Content: ContentConfig{
			Dir:           "content",
			DefaultLayout: "page",
		},
		Assets: AssetsConfig{
			InputDir:  "assets",
			OutputDir: "assets",
		},
	}

	site := &Site{
		Config:  config,
		baseDir: tmpDir,
	}

	layouts := []string{"page", "", "unknown-layout"}

	for _, layout := range layouts {
		t.Run("layout_"+layout, func(t *testing.T) {
			page := &content.Page{
				Path:    "content/test.md",
				Title:   "Test " + layout,
				Content: "<p>Content</p>",
				URL:     "/test-" + layout + "/",
				Layout:  layout,
			}

			ctx := context.Background()
			err := site.renderPage(ctx, page)
			if err != nil {
				t.Fatalf("renderPage with layout %q failed: %v", layout, err)
			}

			// Verify output exists
			outputPath := site.getOutputPath(page.URL)
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Errorf("output file not created for layout %q: %s", layout, outputPath)
			}
		})
	}
}

func TestRenderPageWithEmptyContent(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		Title:     "Test Site",
		BaseURL:   "http://localhost:8080",
		OutputDir: "public",
		Content: ContentConfig{
			Dir:           "content",
			DefaultLayout: "page",
		},
		Assets: AssetsConfig{
			InputDir:  "assets",
			OutputDir: "assets",
		},
	}

	site := &Site{
		Config:  config,
		baseDir: tmpDir,
	}

	page := &content.Page{
		Path:    "content/empty.md",
		Title:   "Empty Page",
		Content: "",
		URL:     "/empty/",
		Layout:  "page",
	}

	ctx := context.Background()
	err := site.renderPage(ctx, page)
	if err != nil {
		t.Fatalf("renderPage with empty content failed: %v", err)
	}

	// Verify output exists and has basic structure
	outputPath := site.getOutputPath(page.URL)
	html, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	htmlStr := string(html)

	// Debug output
	t.Logf("Empty page HTML length: %d bytes", len(html))
	t.Logf("First 200 chars: %s", htmlStr[:min(200, len(htmlStr))])

	if !strings.Contains(strings.ToLower(htmlStr), "<!doctype html>") {
		t.Errorf("output missing DOCTYPE even with empty content. Full output:\n%s", htmlStr)
	}

	if !strings.Contains(htmlStr, "Empty Page") {
		t.Error("output missing page title")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
