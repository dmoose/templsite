// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package site

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dmoose/templsite/pkg/content"
)

// setupLLMsTestSite creates a temp site with content for llms.txt tests.
func setupLLMsTestSite(t *testing.T) *Site {
	t.Helper()
	tmpDir := t.TempDir()

	// Create content directories
	for _, dir := range []string{"content/blog", "content/docs", "content/internal"} {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
	}

	files := map[string]string{
		"content/about.md": `---
title: "About"
description: "About this site"
date: 2024-01-01
---
This is the about page.`,

		"content/blog/post1.md": `---
title: "First Post"
description: "The first blog post"
date: 2024-01-15
tags: ["go"]
---
Hello from the first post.`,

		"content/blog/post2.md": `---
title: "Second Post"
date: 2024-01-10
---
Hello from the second post.`,

		"content/docs/getting-started.md": `---
title: "Getting Started"
description: "How to get started"
date: 2024-01-05
---
# Getting Started

Follow these steps to get started.`,

		"content/internal/secret.md": `---
title: "Internal Notes"
date: 2024-01-01
llms: false
---
This should be excluded.`,
	}

	for name, body := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(body), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	config := DefaultConfig()
	config.BaseURL = "https://example.com"
	config.Title = "Test Site"
	config.Description = "A test site for llms.txt"
	config.LLMs.Enabled = true

	site := NewWithConfig(config)
	site.SetBaseDir(tmpDir)

	if err := site.ProcessContent(context.Background()); err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}

	return site
}

func TestLLMsTxt_AutoSections(t *testing.T) {
	site := setupLLMsTestSite(t)
	result := site.LLMsTxt()

	// H1 title
	if !strings.HasPrefix(result, "# Test Site\n") {
		t.Error("should start with H1 title")
	}

	// Blockquote
	if !strings.Contains(result, "> A test site for llms.txt") {
		t.Error("should contain blockquote description")
	}

	// Section headings (auto-derived)
	if !strings.Contains(result, "## Blog") {
		t.Error("should contain Blog section")
	}
	if !strings.Contains(result, "## Docs") {
		t.Error("should contain Docs section")
	}

	// Page links
	if !strings.Contains(result, "[First Post]") {
		t.Error("should contain First Post link")
	}
	if !strings.Contains(result, "[Getting Started]") {
		t.Error("should contain Getting Started link")
	}

	// Companion URL format
	if !strings.Contains(result, "https://example.com/blog/post1/index.html.md") {
		t.Error("should contain companion URL with index.html.md suffix")
	}

	// Description after colon
	if !strings.Contains(result, "): The first blog post") {
		t.Error("should include page description after colon")
	}

	// Excluded page (frontmatter llms: false)
	if strings.Contains(result, "Internal Notes") {
		t.Error("should not contain page with llms: false")
	}
}

func TestLLMsTxt_ConfiguredSections(t *testing.T) {
	site := setupLLMsTestSite(t)
	site.Config.LLMs.Sections = []LLMsSectionConfig{
		{Name: "Documentation", Pattern: "docs", Priority: "required"},
		{Name: "Blog Posts", Pattern: "blog", Priority: "optional"},
	}

	result := site.LLMsTxt()

	if !strings.Contains(result, "## Documentation") {
		t.Error("should contain configured Documentation section")
	}

	// Optional sections go under ## Optional
	if !strings.Contains(result, "## Optional") {
		t.Error("should contain Optional section")
	}
	if strings.Contains(result, "## Blog Posts") {
		t.Error("should NOT contain Blog Posts as separate section (it's optional)")
	}

	// Blog posts should be under Optional
	optIdx := strings.Index(result, "## Optional")
	if optIdx == -1 {
		t.Fatal("Optional section not found")
	}
	optSection := result[optIdx:]
	if !strings.Contains(optSection, "[First Post]") {
		t.Error("blog posts should appear under Optional section")
	}
}

func TestLLMsTxt_ExcludePattern(t *testing.T) {
	site := setupLLMsTestSite(t)
	site.Config.LLMs.Exclude = []string{"internal"}

	result := site.LLMsTxt()

	if strings.Contains(result, "Internal Notes") {
		t.Error("should not contain pages from excluded section")
	}
	if strings.Contains(result, "## Internal") {
		t.Error("should not contain excluded section heading")
	}
}

func TestLLMsTxt_CustomDescription(t *testing.T) {
	site := setupLLMsTestSite(t)
	site.Config.LLMs.Description = "Custom LLMs description"

	result := site.LLMsTxt()

	if !strings.Contains(result, "> Custom LLMs description") {
		t.Error("should use custom description over site description")
	}
	if strings.Contains(result, "> A test site") {
		t.Error("should not use site description when custom is set")
	}
}

func TestLLMsFull(t *testing.T) {
	site := setupLLMsTestSite(t)
	result := site.LLMsFull()

	// Should have H1
	if !strings.HasPrefix(result, "# Test Site\n") {
		t.Error("should start with H1 title")
	}

	// Should inline content
	if !strings.Contains(result, "Hello from the first post.") {
		t.Error("should inline page content")
	}
	if !strings.Contains(result, "Follow these steps to get started.") {
		t.Error("should inline docs content")
	}

	// Should have H3 page titles
	if !strings.Contains(result, "### First Post") {
		t.Error("should have H3 page title")
	}

	// Should have source links
	if !strings.Contains(result, "Source: [") {
		t.Error("should have Source: link")
	}

	// Should not include excluded pages
	if strings.Contains(result, "This should be excluded") {
		t.Error("should not inline excluded page content")
	}
}

func TestCompanionMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		page     *content.Page
		wantH1   bool
		contains string
	}{
		{
			name: "adds title when missing",
			page: &content.Page{
				Title:      "My Page",
				RawContent: "Some content here.",
			},
			wantH1:   true,
			contains: "# My Page\n\nSome content here.",
		},
		{
			name: "preserves existing H1",
			page: &content.Page{
				Title:      "My Page",
				RawContent: "# Already Has Title\n\nSome content.",
			},
			wantH1:   false,
			contains: "# Already Has Title\n\nSome content.",
		},
		{
			name: "handles empty title",
			page: &content.Page{
				Title:      "",
				RawContent: "Just content, no title.",
			},
			wantH1:   false,
			contains: "Just content, no title.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompanionMarkdown(nil, tt.page)

			if !strings.Contains(result, tt.contains) {
				t.Errorf("expected result to contain %q, got:\n%s", tt.contains, result)
			}

			if tt.wantH1 && !strings.HasPrefix(result, "# ") {
				t.Error("expected result to start with H1")
			}

			// Should end with newline
			if !strings.HasSuffix(result, "\n") {
				t.Error("companion markdown should end with newline")
			}
		})
	}
}

func TestCompanionMarkdown_WithData(t *testing.T) {
	site := setupLLMsTestSite(t)

	// Add data to the site
	site.Data["projects"] = []any{
		map[string]any{
			"title":       "templsite",
			"description": "A static site generator",
			"status":      "Active",
		},
	}

	page := &content.Page{
		Title:      "Projects",
		RawContent: "",
		URL:        "/projects/",
		Frontmatter: map[string]any{
			"title":     "Projects",
			"llms_data": []any{"projects"},
		},
	}

	result := CompanionMarkdown(site, page)

	if !strings.HasPrefix(result, "# Projects") {
		t.Error("should start with title")
	}
	if !strings.Contains(result, "## Data: projects") {
		t.Error("should contain data section heading")
	}
	if !strings.Contains(result, "templsite") {
		t.Error("should contain data content")
	}
	if !strings.Contains(result, "```yaml") {
		t.Error("should wrap data in yaml code block")
	}
}

func TestCompanionMarkdown_WithDataMissing(t *testing.T) {
	site := setupLLMsTestSite(t)

	page := &content.Page{
		Title:      "Projects",
		RawContent: "Some content.",
		URL:        "/projects/",
		Frontmatter: map[string]any{
			"title":     "Projects",
			"llms_data": []any{"nonexistent"},
		},
	}

	result := CompanionMarkdown(site, page)

	// Should still produce valid output without the missing data
	if !strings.Contains(result, "Some content.") {
		t.Error("should still contain raw content")
	}
	if strings.Contains(result, "## Data:") {
		t.Error("should not contain data section for missing keys")
	}
}

func TestLLMsFull_WithData(t *testing.T) {
	site := setupLLMsTestSite(t)

	// Add data and a data-driven page
	site.Data["faq"] = []any{
		map[string]any{
			"question": "What is templsite?",
			"answer":   "A static site generator.",
		},
	}

	// Find the about page and add llms_data to it
	for _, p := range site.Pages {
		if p.Title == "About" {
			p.Frontmatter["llms_data"] = []any{"faq"}
			break
		}
	}

	result := site.LLMsFull()

	if !strings.Contains(result, "#### Data: faq") {
		t.Error("llms-full.txt should inline data from llms_data references")
	}
	if !strings.Contains(result, "What is templsite?") {
		t.Error("llms-full.txt should contain inlined data content")
	}
}

func TestLLMsDataKeys(t *testing.T) {
	tests := []struct {
		name string
		fm   map[string]any
		want []string
	}{
		{"nil", map[string]any{}, nil},
		{"string", map[string]any{"llms_data": "projects"}, []string{"projects"}},
		{"slice", map[string]any{"llms_data": []any{"projects", "faq"}}, []string{"projects", "faq"}},
		{"string_slice", map[string]any{"llms_data": []string{"a", "b"}}, []string{"a", "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &content.Page{Frontmatter: tt.fm}
			got := llmsDataKeys(p)
			if len(got) != len(tt.want) {
				t.Errorf("llmsDataKeys() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("llmsDataKeys()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestLLMsPages(t *testing.T) {
	site := setupLLMsTestSite(t)
	pages := site.LLMsPages()

	// Should not include excluded pages
	for _, p := range pages {
		if p.Title == "Internal Notes" {
			t.Error("LLMsPages should not include excluded pages")
		}
	}

	// Should include regular pages
	titles := make(map[string]bool)
	for _, p := range pages {
		titles[p.Title] = true
	}
	if !titles["First Post"] {
		t.Error("should include First Post")
	}
	if !titles["Getting Started"] {
		t.Error("should include Getting Started")
	}
}

func TestWriteLLMsFiles(t *testing.T) {
	site := setupLLMsTestSite(t)

	// Create output directory
	outputDir := filepath.Join(t.TempDir(), "public")
	site.Config.OutputDir = outputDir

	site.writeLLMsFiles()

	// Check llms.txt was written
	llmsTxt, err := os.ReadFile(filepath.Join(outputDir, "llms.txt"))
	if err != nil {
		t.Fatalf("llms.txt not written: %v", err)
	}
	if !strings.Contains(string(llmsTxt), "# Test Site") {
		t.Error("llms.txt should contain site title")
	}

	// Check llms-full.txt was written
	llmsFull, err := os.ReadFile(filepath.Join(outputDir, "llms-full.txt"))
	if err != nil {
		t.Fatalf("llms-full.txt not written: %v", err)
	}
	if !strings.Contains(string(llmsFull), "Hello from the first post") {
		t.Error("llms-full.txt should contain inlined content")
	}

	// Check companion files were written
	companionPath := filepath.Join(outputDir, "blog", "post1", "index.html.md")
	companion, err := os.ReadFile(companionPath)
	if err != nil {
		t.Fatalf("companion file not written at %s: %v", companionPath, err)
	}
	if !strings.Contains(string(companion), "Hello from the first post") {
		t.Error("companion file should contain page content")
	}
}

func TestWriteLLMsFiles_AlwaysRegenerates(t *testing.T) {
	site := setupLLMsTestSite(t)

	outputDir := filepath.Join(t.TempDir(), "public")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatal(err)
	}
	site.Config.OutputDir = outputDir

	// Pre-create llms.txt with stale content
	stale := "stale llms.txt content"
	if err := os.WriteFile(filepath.Join(outputDir, "llms.txt"), []byte(stale), 0644); err != nil {
		t.Fatal(err)
	}

	site.writeLLMsFiles()

	// Should have regenerated with current content
	data, _ := os.ReadFile(filepath.Join(outputDir, "llms.txt"))
	if string(data) == stale {
		t.Error("writeLLMsFiles should regenerate llms.txt on every build")
	}
	if !strings.Contains(string(data), "# Test Site") {
		t.Error("regenerated llms.txt should contain current site title")
	}
}

func TestMatchSection(t *testing.T) {
	tests := []struct {
		section string
		pattern string
		want    bool
	}{
		{"blog", "blog", true},
		{"blog", "docs", false},
		{"docs", "docs/**", true},
		{"docs", "docs/*", true},
		{"blog", "docs/**", false},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.section+"_"+tt.pattern, func(t *testing.T) {
			if got := matchSection(tt.section, tt.pattern); got != tt.want {
				t.Errorf("matchSection(%q, %q) = %v, want %v", tt.section, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestLLMsTxt_EmptySite(t *testing.T) {
	config := DefaultConfig()
	config.BaseURL = "https://example.com"
	config.Title = "Empty Site"
	config.LLMs.Enabled = true

	site := NewWithConfig(config)
	site.Sections = make(map[string]*Section)
	site.Taxonomies = make(map[string]*Taxonomy)

	result := site.LLMsTxt()

	if !strings.HasPrefix(result, "# Empty Site\n") {
		t.Error("empty site should still produce H1 title")
	}
}
