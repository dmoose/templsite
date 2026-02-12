// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package site

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/dmoose/templsite/pkg/content"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple lowercase",
			input:    "golang",
			expected: "golang",
		},
		{
			name:     "with spaces",
			input:    "Web Development",
			expected: "web-development",
		},
		{
			name:     "with special chars",
			input:    "C++ Programming!",
			expected: "c-programming",
		},
		{
			name:     "multiple spaces",
			input:    "machine   learning",
			expected: "machine-learning",
		},
		{
			name:     "already slug",
			input:    "my-tag",
			expected: "my-tag",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "term",
		},
		{
			name:     "only special chars",
			input:    "!@#$%",
			expected: "term",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slugify(tt.input)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestNewTaxonomy(t *testing.T) {
	tax := NewTaxonomy("tags")

	if tax.Name != "tags" {
		t.Errorf("expected name 'tags', got '%s'", tax.Name)
	}
	if tax.Plural != "tags" {
		t.Errorf("expected plural 'tags', got '%s'", tax.Plural)
	}
	if tax.Terms == nil {
		t.Error("expected Terms map to be initialized")
	}
}

func TestTaxonomyAddPage(t *testing.T) {
	tax := NewTaxonomy("tags")

	page1 := &content.Page{Title: "Post 1"}
	page2 := &content.Page{Title: "Post 2"}
	page3 := &content.Page{Title: "Post 3"}

	tax.AddPage("Go", page1)
	tax.AddPage("Go", page2)
	tax.AddPage("Rust", page3)

	// Check Go term
	goTerm := tax.GetTerm("go")
	if goTerm == nil {
		t.Fatal("expected 'go' term to exist")
	}
	if goTerm.Name != "Go" {
		t.Errorf("expected term name 'Go', got '%s'", goTerm.Name)
	}
	if goTerm.Slug != "go" {
		t.Errorf("expected slug 'go', got '%s'", goTerm.Slug)
	}
	if len(goTerm.Pages) != 2 {
		t.Errorf("expected 2 pages, got %d", len(goTerm.Pages))
	}
	if goTerm.URL != "/tags/go/" {
		t.Errorf("expected URL '/tags/go/', got '%s'", goTerm.URL)
	}

	// Check Rust term
	rustTerm := tax.GetTerm("rust")
	if rustTerm == nil {
		t.Fatal("expected 'rust' term to exist")
	}
	if len(rustTerm.Pages) != 1 {
		t.Errorf("expected 1 page, got %d", len(rustTerm.Pages))
	}
}

func TestTaxonomyAllTerms(t *testing.T) {
	tax := NewTaxonomy("tags")

	// Add pages with different counts
	for range 5 {
		tax.AddPage("popular", &content.Page{Title: "P"})
	}
	for range 2 {
		tax.AddPage("medium", &content.Page{Title: "M"})
	}
	tax.AddPage("rare", &content.Page{Title: "R"})

	terms := tax.AllTerms()

	if len(terms) != 3 {
		t.Errorf("expected 3 terms, got %d", len(terms))
	}

	// Should be sorted by page count descending
	if terms[0].Slug != "popular" {
		t.Errorf("expected first term 'popular', got '%s'", terms[0].Slug)
	}
	if terms[1].Slug != "medium" {
		t.Errorf("expected second term 'medium', got '%s'", terms[1].Slug)
	}
	if terms[2].Slug != "rare" {
		t.Errorf("expected third term 'rare', got '%s'", terms[2].Slug)
	}
}

func TestTaxonomyTermsByName(t *testing.T) {
	tax := NewTaxonomy("tags")

	tax.AddPage("Zebra", &content.Page{})
	tax.AddPage("Apple", &content.Page{})
	tax.AddPage("Mango", &content.Page{})

	terms := tax.TermsByName()

	if len(terms) != 3 {
		t.Errorf("expected 3 terms, got %d", len(terms))
	}

	// Should be sorted alphabetically
	if terms[0].Name != "Apple" {
		t.Errorf("expected first term 'Apple', got '%s'", terms[0].Name)
	}
	if terms[1].Name != "Mango" {
		t.Errorf("expected second term 'Mango', got '%s'", terms[1].Name)
	}
	if terms[2].Name != "Zebra" {
		t.Errorf("expected third term 'Zebra', got '%s'", terms[2].Name)
	}
}

func TestTermPageCount(t *testing.T) {
	term := &Term{
		Name:  "test",
		Pages: []*content.Page{{}, {}, {}},
	}

	if term.PageCount() != 3 {
		t.Errorf("expected page count 3, got %d", term.PageCount())
	}
}

func TestBuildTaxonomies(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content", "blog")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	files := map[string]string{
		"post1.md": `---
title: "Go Post"
tags: ["go", "programming"]
categories: ["tutorials"]
---
Content`,
		"post2.md": `---
title: "Rust Post"
tags: ["rust", "programming"]
categories: ["tutorials"]
---
Content`,
		"post3.md": `---
title: "Go Advanced"
tags: ["go", "advanced"]
categories: ["deep-dive"]
---
Content`,
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(contentDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	config := DefaultConfig()
	config.Taxonomies = []string{"tags", "categories"}

	site := NewWithConfig(config)
	site.SetBaseDir(tmpDir)

	ctx := context.Background()
	if err := site.ProcessContent(ctx); err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}

	// Check taxonomies were created
	if len(site.Taxonomies) != 2 {
		t.Errorf("expected 2 taxonomies, got %d", len(site.Taxonomies))
	}

	// Check tags taxonomy
	tags := site.GetTaxonomy("tags")
	if tags == nil {
		t.Fatal("expected 'tags' taxonomy")
	}

	goTerm := tags.GetTerm("go")
	if goTerm == nil {
		t.Fatal("expected 'go' term in tags")
	}
	if len(goTerm.Pages) != 2 {
		t.Errorf("expected 2 pages with 'go' tag, got %d", len(goTerm.Pages))
	}

	progTerm := tags.GetTerm("programming")
	if progTerm == nil {
		t.Fatal("expected 'programming' term in tags")
	}
	if len(progTerm.Pages) != 2 {
		t.Errorf("expected 2 pages with 'programming' tag, got %d", len(progTerm.Pages))
	}

	// Check categories taxonomy
	cats := site.GetTaxonomy("categories")
	if cats == nil {
		t.Fatal("expected 'categories' taxonomy")
	}

	tutorials := cats.GetTerm("tutorials")
	if tutorials == nil {
		t.Fatal("expected 'tutorials' term in categories")
	}
	if len(tutorials.Pages) != 2 {
		t.Errorf("expected 2 pages in 'tutorials' category, got %d", len(tutorials.Pages))
	}
}

func TestTaxonomyTermsQuery(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content", "blog")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	files := map[string]string{
		"post1.md": `---
title: "Post 1"
tags: ["go"]
---
Content`,
		"post2.md": `---
title: "Post 2"
tags: ["go", "web"]
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

	// Test TaxonomyTerms
	terms := site.TaxonomyTerms("tags")
	if len(terms) != 2 {
		t.Errorf("expected 2 terms, got %d", len(terms))
	}

	// Should be sorted by page count (go has 2, web has 1)
	if terms[0].Slug != "go" {
		t.Errorf("expected 'go' first (most pages), got '%s'", terms[0].Slug)
	}

	// Test PagesByTaxonomy
	goPages := site.PagesByTaxonomy("tags", "go")
	if len(goPages) != 2 {
		t.Errorf("expected 2 pages with 'go' tag, got %d", len(goPages))
	}

	webPages := site.PagesByTaxonomy("tags", "web")
	if len(webPages) != 1 {
		t.Errorf("expected 1 page with 'web' tag, got %d", len(webPages))
	}

	// Test nonexistent
	nilPages := site.PagesByTaxonomy("tags", "nonexistent")
	if nilPages != nil {
		t.Error("expected nil for nonexistent term")
	}

	nilPages = site.PagesByTaxonomy("nonexistent", "go")
	if nilPages != nil {
		t.Error("expected nil for nonexistent taxonomy")
	}
}

func TestGetTerm(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content", "blog")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(contentDir, "post.md"),
		[]byte("---\ntitle: Post\ntags: [\"go\"]\n---\nContent"),
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

	term := site.GetTerm("tags", "go")
	if term == nil {
		t.Fatal("expected to find 'go' term")
	}
	if term.Name != "go" {
		t.Errorf("expected term name 'go', got '%s'", term.Name)
	}

	// Test nonexistent
	if site.GetTerm("tags", "nonexistent") != nil {
		t.Error("expected nil for nonexistent term")
	}
	if site.GetTerm("nonexistent", "go") != nil {
		t.Error("expected nil for nonexistent taxonomy")
	}
}

func TestTagsConvenienceMethod(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content", "blog")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(contentDir, "post.md"),
		[]byte("---\ntitle: Post\ntags: [\"go\", \"web\"]\n---\nContent"),
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

	tags := site.Tags()
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}
}
