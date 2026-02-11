package content

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewParser(t *testing.T) {
	parser := NewParser("content")

	if parser.contentDir != "content" {
		t.Errorf("expected contentDir 'content', got '%s'", parser.contentDir)
	}

	if parser.markdown == nil {
		t.Error("expected markdown parser to be initialized")
	}
}

func TestParseFrontmatter(t *testing.T) {
	parser := NewParser("content")

	tests := []struct {
		name        string
		content     string
		wantTitle   string
		wantLayout  string
		wantBody    string
		shouldError bool
	}{
		{
			name: "valid frontmatter",
			content: `---
title: "Test Post"
layout: "post"
---
# Hello World`,
			wantTitle:  "Test Post",
			wantLayout: "post",
			wantBody:   "# Hello World",
		},
		{
			name: "no frontmatter",
			content: `# Hello World
This is content without frontmatter.`,
			wantBody: `# Hello World
This is content without frontmatter.`,
		},
		{
			name: "empty frontmatter",
			content: `---
---
# Hello World`,
			wantBody: "# Hello World",
		},
		{
			name: "frontmatter with multiple fields",
			content: `---
title: "Multi Field"
author: "John Doe"
date: 2025-01-15
draft: true
tags:
  - go
  - static-site
---
Content here`,
			wantTitle: "Multi Field",
			wantBody:  "Content here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frontmatter, body, err := parser.parseFrontmatter([]byte(tt.content))

			if tt.shouldError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.wantTitle != "" {
				if title := getStringDefault(frontmatter, "title", ""); title != tt.wantTitle {
					t.Errorf("expected title '%s', got '%s'", tt.wantTitle, title)
				}
			}

			if tt.wantLayout != "" {
				if layout := getStringDefault(frontmatter, "layout", ""); layout != tt.wantLayout {
					t.Errorf("expected layout '%s', got '%s'", tt.wantLayout, layout)
				}
			}

			if tt.wantBody != "" {
				bodyStr := strings.TrimSpace(string(body))
				wantBody := strings.TrimSpace(tt.wantBody)
				if bodyStr != wantBody {
					t.Errorf("expected body '%s', got '%s'", wantBody, bodyStr)
				}
			}
		})
	}
}

func TestGenerateURL(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content")
	parser := NewParser(contentDir)

	tests := []struct {
		name    string
		path    string
		wantURL string
	}{
		{
			name:    "index file",
			path:    filepath.Join(contentDir, "index.md"),
			wantURL: "/",
		},
		{
			name:    "simple page",
			path:    filepath.Join(contentDir, "about.md"),
			wantURL: "/about/",
		},
		{
			name:    "nested page",
			path:    filepath.Join(contentDir, "blog", "post.md"),
			wantURL: "/blog/post/",
		},
		{
			name:    "nested index",
			path:    filepath.Join(contentDir, "blog", "index.md"),
			wantURL: "/blog/",
		},
		{
			name:    "deeply nested",
			path:    filepath.Join(contentDir, "docs", "guide", "getting-started.md"),
			wantURL: "/docs/guide/getting-started/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := parser.generateURL(tt.path)
			if url != tt.wantURL {
				t.Errorf("expected URL '%s', got '%s'", tt.wantURL, url)
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(contentDir, "test.md")
	testContent := `---
title: "Test Page"
date: 2025-01-15
layout: "post"
draft: false
tags:
  - test
  - markdown
author: "John Doe"
description: "A test page"
---
# Test Heading

This is a **test** page with some content.

## Subheading

- List item 1
- List item 2
`

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	parser := NewParser(contentDir)
	ctx := context.Background()

	page, err := parser.ParseFile(ctx, testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Verify page fields
	if page.Title != "Test Page" {
		t.Errorf("expected title 'Test Page', got '%s'", page.Title)
	}

	if page.Layout != "post" {
		t.Errorf("expected layout 'post', got '%s'", page.Layout)
	}

	if page.Draft {
		t.Error("expected draft to be false")
	}

	if page.Author != "John Doe" {
		t.Errorf("expected author 'John Doe', got '%s'", page.Author)
	}

	if page.Description != "A test page" {
		t.Errorf("expected description 'A test page', got '%s'", page.Description)
	}

	if len(page.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(page.Tags))
	}

	if page.URL != "/test/" {
		t.Errorf("expected URL '/test/', got '%s'", page.URL)
	}

	// Date-only values are parsed in local timezone (what the user intends)
	expectedDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.Local)
	if !page.Date.Equal(expectedDate) {
		t.Errorf("expected date %v, got %v", expectedDate, page.Date)
	}

	// Verify HTML content is generated
	if page.Content == "" {
		t.Error("expected content to be generated")
	}

	if !strings.Contains(page.Content, "<h1") {
		t.Error("expected content to contain h1 tag")
	}

	if !strings.Contains(page.Content, "<strong>test</strong>") {
		t.Error("expected content to contain bold text")
	}
}

func TestParseAll(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	// Create multiple test files
	files := map[string]string{
		"index.md": `---
title: "Home"
---
# Welcome`,
		"about.md": `---
title: "About"
---
# About Us`,
		"blog/post1.md": `---
title: "Post 1"
---
# First Post`,
		"blog/post2.md": `---
title: "Post 2"
draft: true
---
# Second Post`,
	}

	for path, content := range files {
		fullPath := filepath.Join(contentDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	// Create a non-markdown file (should be ignored)
	nonMdFile := filepath.Join(contentDir, "readme.txt")
	if err := os.WriteFile(nonMdFile, []byte("not markdown"), 0644); err != nil {
		t.Fatalf("failed to write non-md file: %v", err)
	}

	parser := NewParser(contentDir)
	ctx := context.Background()

	pages, err := parser.ParseAll(ctx)
	if err != nil {
		t.Fatalf("ParseAll failed: %v", err)
	}

	// Should only parse .md files
	if len(pages) != 4 {
		t.Errorf("expected 4 pages, got %d", len(pages))
	}

	// Verify all pages have content
	for _, page := range pages {
		if page.Content == "" {
			t.Errorf("page %s has no content", page.Path)
		}
		if page.Title == "" {
			t.Errorf("page %s has no title", page.Path)
		}
	}
}

func TestParseAllWithContext(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(contentDir, "test.md")
	if err := os.WriteFile(testFile, []byte("# Test"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	parser := NewParser(contentDir)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := parser.ParseAll(ctx)
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		name        string
		dateStr     string
		shouldError bool
	}{
		{
			name:        "YYYY-MM-DD",
			dateStr:     "2025-01-15",
			shouldError: false,
		},
		{
			name:        "RFC3339",
			dateStr:     "2025-01-15T10:30:00Z",
			shouldError: false,
		},
		{
			name:        "with time",
			dateStr:     "2025-01-15 10:30:00",
			shouldError: false,
		},
		{
			name:        "invalid format",
			dateStr:     "Jan 15, 2025",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, err := parseDate(tt.dateStr)

			if tt.shouldError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.shouldError {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if date.IsZero() {
					t.Error("expected non-zero date")
				}
			}
		})
	}
}

func TestGetStringDefault(t *testing.T) {
	m := map[string]any{
		"exists": "value",
		"number": 42,
	}

	if got := getStringDefault(m, "exists", "default"); got != "value" {
		t.Errorf("expected 'value', got '%s'", got)
	}

	if got := getStringDefault(m, "missing", "default"); got != "default" {
		t.Errorf("expected 'default', got '%s'", got)
	}

	if got := getStringDefault(m, "number", "default"); got != "default" {
		t.Errorf("expected 'default' for non-string, got '%s'", got)
	}
}

func TestGetBoolDefault(t *testing.T) {
	m := map[string]any{
		"true":   true,
		"false":  false,
		"string": "not a bool",
	}

	if got := getBoolDefault(m, "true", false); !got {
		t.Error("expected true")
	}

	if got := getBoolDefault(m, "false", true); got {
		t.Error("expected false")
	}

	if got := getBoolDefault(m, "missing", true); !got {
		t.Error("expected default true")
	}

	if got := getBoolDefault(m, "string", true); !got {
		t.Error("expected default true for non-bool")
	}
}

func TestGetStringSlice(t *testing.T) {
	m := map[string]any{
		"tags":   []any{"tag1", "tag2", "tag3"},
		"direct": []string{"a", "b"},
		"mixed":  []any{"string", 42, "another"},
		"empty":  []any{},
	}

	tags := getStringSlice(m, "tags")
	if len(tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(tags))
	}
	if tags[0] != "tag1" {
		t.Errorf("expected 'tag1', got '%s'", tags[0])
	}

	direct := getStringSlice(m, "direct")
	if len(direct) != 2 {
		t.Errorf("expected 2 items, got %d", len(direct))
	}

	mixed := getStringSlice(m, "mixed")
	if len(mixed) != 2 { // Should only include string items
		t.Errorf("expected 2 string items, got %d", len(mixed))
	}

	empty := getStringSlice(m, "empty")
	if len(empty) != 0 {
		t.Errorf("expected empty slice, got %d items", len(empty))
	}

	missing := getStringSlice(m, "missing")
	if missing != nil {
		t.Error("expected nil for missing key")
	}
}

func TestPageIsPublished(t *testing.T) {
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)

	tests := []struct {
		name          string
		page          *Page
		wantPublished bool
	}{
		{
			name: "not draft, no date",
			page: &Page{
				Draft: false,
			},
			wantPublished: true,
		},
		{
			name: "draft",
			page: &Page{
				Draft: true,
				Date:  past,
			},
			wantPublished: false,
		},
		{
			name: "published with past date",
			page: &Page{
				Draft: false,
				Date:  past,
			},
			wantPublished: true,
		},
		{
			name: "future date",
			page: &Page{
				Draft: false,
				Date:  future,
			},
			wantPublished: false,
		},
		{
			name: "current date",
			page: &Page{
				Draft: false,
				Date:  now,
			},
			wantPublished: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.page.IsPublished(); got != tt.wantPublished {
				t.Errorf("expected IsPublished() = %v, got %v", tt.wantPublished, got)
			}
		})
	}
}

func TestPageHasTag(t *testing.T) {
	page := &Page{
		Tags: []string{"go", "programming", "web"},
	}

	if !page.HasTag("go") {
		t.Error("expected page to have 'go' tag")
	}

	if !page.HasTag("programming") {
		t.Error("expected page to have 'programming' tag")
	}

	if page.HasTag("python") {
		t.Error("expected page not to have 'python' tag")
	}

	emptyPage := &Page{}
	if emptyPage.HasTag("any") {
		t.Error("expected page with no tags to return false")
	}
}

func TestMarkdownRendering(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	testFile := filepath.Join(contentDir, "markdown.md")
	testContent := `# Heading 1

## Heading 2

This is a paragraph with **bold** and *italic* text.

- List item 1
- List item 2

[Link](https://example.com)

` + "```go\nfunc main() {\n}\n```"

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	parser := NewParser(contentDir)
	ctx := context.Background()

	page, err := parser.ParseFile(ctx, testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Check various HTML elements are rendered
	expectations := []string{
		"<h1",
		"<h2",
		"<strong>bold</strong>",
		"<em>italic</em>",
		"<ul>",
		"<li>",
		"<a href=\"https://example.com\"",
		"<pre>",
		"<code",
	}

	for _, exp := range expectations {
		if !strings.Contains(page.Content, exp) {
			t.Errorf("expected content to contain '%s'", exp)
		}
	}
}

func TestExtractSection(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content")
	parser := NewParser(contentDir)

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "root file",
			path:     filepath.Join(contentDir, "about.md"),
			expected: "",
		},
		{
			name:     "blog section",
			path:     filepath.Join(contentDir, "blog", "post.md"),
			expected: "blog",
		},
		{
			name:     "nested in section",
			path:     filepath.Join(contentDir, "docs", "guide", "intro.md"),
			expected: "docs",
		},
		{
			name:     "section index",
			path:     filepath.Join(contentDir, "blog", "index.md"),
			expected: "blog",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.extractSection(tt.path)
			if result != tt.expected {
				t.Errorf("expected section '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestParseFileNewFields(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content", "blog")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	testFile := filepath.Join(contentDir, "test-post.md")
	testContent := `---
title: "Test Post"
weight: 10
aliases:
  - /old-url/
  - /another-old-url/
---
This is the first paragraph of the blog post and it contains enough words to test our word counting and reading time functionality.

<!--more-->

## Details

This is the second paragraph with more details that should not appear in the summary.

### Subheading

And even more content here with many words to test word counting thoroughly.
`

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	parser := NewParser(filepath.Join(tmpDir, "content"))
	ctx := context.Background()

	page, err := parser.ParseFile(ctx, testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Test Section
	if page.Section != "blog" {
		t.Errorf("expected section 'blog', got '%s'", page.Section)
	}

	// Test Weight
	if page.Weight != 10 {
		t.Errorf("expected weight 10, got %d", page.Weight)
	}

	// Test Aliases
	if len(page.Aliases) != 2 {
		t.Errorf("expected 2 aliases, got %d", len(page.Aliases))
	}

	// Test RawContent (should not include frontmatter)
	if page.RawContent == "" {
		t.Error("expected non-empty RawContent")
	}
	if strings.Contains(page.RawContent, "title:") {
		t.Error("RawContent should not contain frontmatter")
	}

	// Test WordCount
	if page.WordCount == 0 {
		t.Error("expected non-zero WordCount")
	}

	// Test ReadingTime (with the current content, should be at least 1 minute)
	if page.ReadingTime == 0 {
		t.Error("expected non-zero ReadingTime")
	}

	// Test Summary (should contain first paragraph content)
	if page.Summary == "" {
		t.Error("expected non-empty Summary")
	}
	if !strings.Contains(page.Summary, "first paragraph") {
		t.Error("Summary should contain first paragraph content")
	}
	// Content after <!--more--> should not be in summary
	if strings.Contains(page.Summary, "should not appear") {
		t.Error("Summary should not contain content after <!--more-->")
	}

	// Test TOC (should have entries for h2, h3)
	if page.TOC == "" {
		t.Error("expected non-empty TOC")
	}
	if !strings.Contains(page.TOC, "Details") {
		t.Error("TOC should contain h2 heading 'Details'")
	}
	if !strings.Contains(page.TOC, "Subheading") {
		t.Error("TOC should contain h3 heading 'Subheading'")
	}
}

func TestGetIntDefault(t *testing.T) {
	m := map[string]any{
		"int":    42,
		"float":  3.14,
		"string": "not an int",
	}

	if got := getIntDefault(m, "int", 0); got != 42 {
		t.Errorf("expected 42, got %d", got)
	}

	if got := getIntDefault(m, "float", 0); got != 3 {
		t.Errorf("expected 3 (from float64), got %d", got)
	}

	if got := getIntDefault(m, "string", 99); got != 99 {
		t.Errorf("expected default 99 for non-int, got %d", got)
	}

	if got := getIntDefault(m, "missing", 100); got != 100 {
		t.Errorf("expected default 100, got %d", got)
	}
}
