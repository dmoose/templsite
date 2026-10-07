// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package site

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSitemap(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content", "blog")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	// Create test content
	files := map[string]string{
		"post1.md": `---
title: "Post 1"
date: 2024-01-15
tags: ["go"]
---
Content`,
		"post2.md": `---
title: "Post 2"
date: 2024-01-10
---
Content`,
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(contentDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	config := DefaultConfig()
	config.BaseURL = "https://example.com"

	site := NewWithConfig(config)
	site.SetBaseDir(tmpDir)

	ctx := context.Background()
	if err := site.ProcessContent(ctx); err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}

	sitemap := site.Sitemap()

	// Check XML declaration
	if !strings.HasPrefix(sitemap, "<?xml") {
		t.Error("Sitemap should start with XML declaration")
	}

	// Check namespace
	if !strings.Contains(sitemap, "xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\"") {
		t.Error("Sitemap should have sitemap namespace")
	}

	// Check URLs are present
	if !strings.Contains(sitemap, "<loc>https://example.com/blog/post1/</loc>") {
		t.Error("Sitemap should contain post1 URL")
	}
	if !strings.Contains(sitemap, "<loc>https://example.com/blog/post2/</loc>") {
		t.Error("Sitemap should contain post2 URL")
	}

	// Check section URL
	if !strings.Contains(sitemap, "<loc>https://example.com/blog/</loc>") {
		t.Error("Sitemap should contain blog section URL")
	}

	// Check lastmod for dated content
	if !strings.Contains(sitemap, "<lastmod>2024-01-15</lastmod>") {
		t.Error("Sitemap should contain lastmod date")
	}

	// Check taxonomy term URLs
	if !strings.Contains(sitemap, "<loc>https://example.com/tags/go/</loc>") {
		t.Error("Sitemap should contain tag URL")
	}
}

func TestRobotsTxt(t *testing.T) {
	config := DefaultConfig()
	config.BaseURL = "https://example.com"

	site := NewWithConfig(config)

	robots := site.RobotsTxt()

	if !strings.Contains(robots, "User-agent: *") {
		t.Error("robots.txt should contain User-agent")
	}
	if !strings.Contains(robots, "Allow: /") {
		t.Error("robots.txt should contain Allow")
	}
	if !strings.Contains(robots, "Sitemap: https://example.com/sitemap.xml") {
		t.Error("robots.txt should contain Sitemap URL")
	}
}

func TestRobotsTxtWithDisallow(t *testing.T) {
	config := DefaultConfig()
	config.BaseURL = "https://example.com"

	site := NewWithConfig(config)

	robots := site.RobotsTxtWithDisallow([]string{"/admin/", "/private/"})

	if !strings.Contains(robots, "User-agent: *") {
		t.Error("robots.txt should contain User-agent")
	}
	if !strings.Contains(robots, "Disallow: /admin/") {
		t.Error("robots.txt should contain Disallow for /admin/")
	}
	if !strings.Contains(robots, "Disallow: /private/") {
		t.Error("robots.txt should contain Disallow for /private/")
	}
	if strings.Contains(robots, "Allow: /") {
		t.Error("robots.txt should not contain Allow when disallow rules are present")
	}
	if !strings.Contains(robots, "Sitemap: https://example.com/sitemap.xml") {
		t.Error("robots.txt should contain Sitemap URL")
	}
}

func TestSitemapEmptySite(t *testing.T) {
	config := DefaultConfig()
	config.BaseURL = "https://example.com"

	site := NewWithConfig(config)
	site.Sections = make(map[string]*Section)
	site.Taxonomies = make(map[string]*Taxonomy)

	sitemap := site.Sitemap()

	// Should still be valid XML
	if !strings.HasPrefix(sitemap, "<?xml") {
		t.Error("Empty sitemap should still be valid XML")
	}
	if !strings.Contains(sitemap, "<urlset") {
		t.Error("Empty sitemap should have urlset element")
	}
}

// TestSitemapOmitsNoIndex: a noindex page is not listed, and a section
// left with nothing indexable is not listed either.
func TestSitemapOmitsNoIndex(t *testing.T) {
	tmpDir := t.TempDir()
	files := map[string]string{
		"about.md":              "---\ntitle: About\n---\nBody",
		"404.md":                "---\ntitle: Not Found\nnoindex: true\n---\nBody",
		"purchase/completed.md": "---\ntitle: Thanks\nnoindex: true\n---\nBody",
		"purchase/returned.md":  "---\ntitle: Still thinking\nnoindex: true\n---\nBody",
		"blog/one.md":           "---\ntitle: One\n---\nBody",
		"blog/hidden.md":        "---\ntitle: Hidden\nnoindex: true\n---\nBody",
	}
	for name, body := range files {
		path := filepath.Join(tmpDir, "content", name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}

	config := DefaultConfig()
	config.BaseURL = "https://example.com"
	site := NewWithConfig(config)
	site.SetBaseDir(tmpDir)
	if err := site.ProcessContent(t.Context()); err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}
	sitemap := site.Sitemap()

	for _, want := range []string{"/about/", "/blog/one/", "/blog/"} {
		if !strings.Contains(sitemap, "<loc>https://example.com"+want+"</loc>") {
			t.Errorf("sitemap is missing %s", want)
		}
	}
	for _, gone := range []string{"/404/", "/purchase/", "/purchase/completed/", "/purchase/returned/", "/blog/hidden/"} {
		if strings.Contains(sitemap, "<loc>https://example.com"+gone+"</loc>") {
			t.Errorf("sitemap lists %s, which is noindex or has nothing indexable", gone)
		}
	}
}
