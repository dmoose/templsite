package site

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRegularPages(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content")

	// Create blog section with posts
	blogDir := filepath.Join(contentDir, "blog")
	if err := os.MkdirAll(blogDir, 0755); err != nil {
		t.Fatalf("failed to create blog dir: %v", err)
	}

	files := map[string]string{
		"content/index.md": `---
title: "Home"
---
Welcome`,
		"content/blog/_index.md": `---
title: "Blog"
---
Blog listing`,
		"content/blog/post1.md": `---
title: "Post 1"
date: 2025-01-15
---
Content`,
		"content/blog/post2.md": `---
title: "Post 2"
date: 2025-01-10
draft: true
---
Draft content`,
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	site := NewWithConfig(DefaultConfig())
	site.SetBaseDir(tmpDir)

	ctx := context.Background()
	if err := site.ProcessContent(ctx); err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}

	regularPages := site.RegularPages()

	// Should only include Post 1 (not Home, Blog index, or draft Post 2)
	if len(regularPages) != 1 {
		t.Errorf("expected 1 regular page, got %d", len(regularPages))
		for _, p := range regularPages {
			t.Logf("  - %s (%s)", p.Title, p.URL)
		}
	}

	if len(regularPages) > 0 && regularPages[0].Title != "Post 1" {
		t.Errorf("expected Post 1, got %s", regularPages[0].Title)
	}
}

func TestPagesInSection(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content")

	// Create blog and docs sections
	for _, dir := range []string{"blog", "docs"} {
		if err := os.MkdirAll(filepath.Join(contentDir, dir), 0755); err != nil {
			t.Fatalf("failed to create %s dir: %v", dir, err)
		}
	}

	files := map[string]string{
		"content/blog/post1.md": `---
title: "Blog Post 1"
---
Content`,
		"content/blog/post2.md": `---
title: "Blog Post 2"
---
Content`,
		"content/docs/guide.md": `---
title: "Guide"
---
Content`,
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	site := NewWithConfig(DefaultConfig())
	site.SetBaseDir(tmpDir)

	ctx := context.Background()
	if err := site.ProcessContent(ctx); err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}

	// Test blog section
	blogPages := site.PagesInSection("blog")
	if len(blogPages) != 2 {
		t.Errorf("expected 2 blog pages, got %d", len(blogPages))
	}

	// Test docs section
	docsPages := site.PagesInSection("docs")
	if len(docsPages) != 1 {
		t.Errorf("expected 1 docs page, got %d", len(docsPages))
	}

	// Test nonexistent section
	nopages := site.PagesInSection("nonexistent")
	if nopages != nil {
		t.Error("expected nil for nonexistent section")
	}
}

func TestGetSection(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content", "blog")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	files := map[string]string{
		"content/blog/_index.md": `---
title: "My Blog"
description: "A great blog"
---
Blog content`,
		"content/blog/post.md": `---
title: "Post"
---
Content`,
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	site := NewWithConfig(DefaultConfig())
	site.SetBaseDir(tmpDir)

	ctx := context.Background()
	if err := site.ProcessContent(ctx); err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}

	section := site.GetSection("blog")
	if section == nil {
		t.Fatal("expected blog section")
	}

	if section.Title != "My Blog" {
		t.Errorf("expected title 'My Blog', got '%s'", section.Title)
	}

	if section.Description != "A great blog" {
		t.Errorf("expected description 'A great blog', got '%s'", section.Description)
	}

	if section.URL != "/blog/" {
		t.Errorf("expected URL '/blog/', got '%s'", section.URL)
	}
}

func TestAllSections(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple sections
	for _, dir := range []string{"blog", "docs", "projects"} {
		if err := os.MkdirAll(filepath.Join(tmpDir, "content", dir), 0755); err != nil {
			t.Fatalf("failed to create %s dir: %v", dir, err)
		}
		// Add a page to each section
		path := filepath.Join(tmpDir, "content", dir, "page.md")
		content := "---\ntitle: \"Page\"\n---\nContent"
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	site := NewWithConfig(DefaultConfig())
	site.SetBaseDir(tmpDir)

	ctx := context.Background()
	if err := site.ProcessContent(ctx); err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}

	sections := site.AllSections()
	if len(sections) != 3 {
		t.Errorf("expected 3 sections, got %d", len(sections))
	}
}

func TestPageByURL(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content", "blog")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(contentDir, "my-post.md"),
		[]byte("---\ntitle: \"My Post\"\n---\nContent"),
		0644,
	); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	site := NewWithConfig(DefaultConfig())
	site.SetBaseDir(tmpDir)

	ctx := context.Background()
	if err := site.ProcessContent(ctx); err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}

	// Test with various URL formats
	tests := []struct {
		url      string
		expected string
	}{
		{"/blog/my-post/", "My Post"},
		{"blog/my-post/", "My Post"},
		{"/blog/my-post", "My Post"},
		{"/nonexistent/", ""},
	}

	for _, tt := range tests {
		page := site.PageByURL(tt.url)
		if tt.expected == "" {
			if page != nil {
				t.Errorf("expected nil for URL %s, got %s", tt.url, page.Title)
			}
		} else {
			if page == nil {
				t.Errorf("expected page for URL %s, got nil", tt.url)
			} else if page.Title != tt.expected {
				t.Errorf("expected title %s for URL %s, got %s", tt.expected, tt.url, page.Title)
			}
		}
	}
}

func TestPagesByTag(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content", "blog")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	files := map[string]string{
		"post1.md": `---
title: "Go Post"
tags: ["go", "programming"]
---
Content`,
		"post2.md": `---
title: "Rust Post"
tags: ["rust", "programming"]
---
Content`,
		"post3.md": `---
title: "Go Advanced"
tags: ["go", "advanced"]
---
Content`,
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(contentDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	site := NewWithConfig(DefaultConfig())
	site.SetBaseDir(tmpDir)

	ctx := context.Background()
	if err := site.ProcessContent(ctx); err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}

	// Test go tag
	goPages := site.PagesByTag("go")
	if len(goPages) != 2 {
		t.Errorf("expected 2 pages with 'go' tag, got %d", len(goPages))
	}

	// Test programming tag
	progPages := site.PagesByTag("programming")
	if len(progPages) != 2 {
		t.Errorf("expected 2 pages with 'programming' tag, got %d", len(progPages))
	}

	// Test nonexistent tag
	nopages := site.PagesByTag("nonexistent")
	if len(nopages) != 0 {
		t.Errorf("expected 0 pages with 'nonexistent' tag, got %d", len(nopages))
	}
}

func TestAllTags(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content", "blog")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	files := map[string]string{
		"post1.md": `---
title: "Post 1"
tags: ["go", "web"]
---
Content`,
		"post2.md": `---
title: "Post 2"
tags: ["go", "cli"]
---
Content`,
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(contentDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	site := NewWithConfig(DefaultConfig())
	site.SetBaseDir(tmpDir)

	ctx := context.Background()
	if err := site.ProcessContent(ctx); err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}

	tags := site.AllTags()

	if tags["go"] != 2 {
		t.Errorf("expected 'go' tag count 2, got %d", tags["go"])
	}
	if tags["web"] != 1 {
		t.Errorf("expected 'web' tag count 1, got %d", tags["web"])
	}
	if tags["cli"] != 1 {
		t.Errorf("expected 'cli' tag count 1, got %d", tags["cli"])
	}
}
