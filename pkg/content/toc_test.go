// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package content

import (
	"strings"
	"testing"
)

func TestGenerateTOC(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		minLevel    int
		maxLevel    int
		wantEmpty   bool
		wantContain []string
	}{
		{
			name:      "no headings",
			html:      "<p>Just a paragraph</p>",
			minLevel:  2,
			maxLevel:  4,
			wantEmpty: true,
		},
		{
			name:      "single h2",
			html:      `<h2 id="intro">Introduction</h2>`,
			minLevel:  2,
			maxLevel:  4,
			wantEmpty: false,
			wantContain: []string{
				`<a class="toc-link" href="#intro">Introduction</a>`,
				`<ul class="toc-list">`,
				`<li class="toc-item toc-item-h2">`,
			},
		},
		{
			name:      "h1 excluded by default",
			html:      `<h1 id="title">Title</h1><h2 id="intro">Introduction</h2>`,
			minLevel:  2,
			maxLevel:  4,
			wantEmpty: false,
			wantContain: []string{
				`<a class="toc-link" href="#intro">Introduction</a>`,
			},
		},
		{
			name:      "nested headings",
			html:      `<h2 id="section">Section</h2><h3 id="subsection">Subsection</h3>`,
			minLevel:  2,
			maxLevel:  4,
			wantEmpty: false,
			wantContain: []string{
				`<a class="toc-link" href="#section">Section</a>`,
				`<a class="toc-link" href="#subsection">Subsection</a>`,
				`<li class="toc-item toc-item-h2">`,
				`<li class="toc-item toc-item-h3">`,
			},
		},
		{
			name:      "heading without id gets generated id",
			html:      `<h2>My Heading</h2>`,
			minLevel:  2,
			maxLevel:  4,
			wantEmpty: false,
			wantContain: []string{
				`<a class="toc-link" href="#my-heading">My Heading</a>`,
			},
		},
		{
			name:      "custom min/max levels",
			html:      `<h1 id="h1">H1</h1><h2 id="h2">H2</h2><h3 id="h3">H3</h3><h4 id="h4">H4</h4>`,
			minLevel:  1,
			maxLevel:  2,
			wantEmpty: false,
			wantContain: []string{
				`<a class="toc-link" href="#h1">H1</a>`,
				`<a class="toc-link" href="#h2">H2</a>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateTOC(tt.html, tt.minLevel, tt.maxLevel)

			if tt.wantEmpty {
				if result != "" {
					t.Errorf("expected empty TOC, got '%s'", result)
				}
				return
			}

			if result == "" {
				t.Error("expected non-empty TOC")
				return
			}

			for _, want := range tt.wantContain {
				if !strings.Contains(result, want) {
					t.Errorf("expected TOC to contain '%s', got:\n%s", want, result)
				}
			}
		})
	}
}

func TestExtractHeadings(t *testing.T) {
	html := `<h1 id="title">Title</h1>
<h2 id="intro">Introduction</h2>
<h3 id="sub1">Sub One</h3>
<h3 id="sub2">Sub Two</h3>
<h2 id="conclusion">Conclusion</h2>`

	entries := extractHeadings(html, 2, 4)

	if len(entries) != 4 {
		t.Errorf("expected 4 headings, got %d", len(entries))
	}

	// Check first entry
	if entries[0].Level != 2 || entries[0].ID != "intro" || entries[0].Text != "Introduction" {
		t.Errorf("unexpected first entry: %+v", entries[0])
	}

	// Check sub-headings
	if entries[1].Level != 3 || entries[1].ID != "sub1" {
		t.Errorf("unexpected second entry: %+v", entries[1])
	}
}

func TestGenerateHeadingID(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "simple text",
			text:     "Hello World",
			expected: "hello-world",
		},
		{
			name:     "with special characters",
			text:     "Hello, World!",
			expected: "hello-world",
		},
		{
			name:     "with numbers",
			text:     "Chapter 1: Introduction",
			expected: "chapter-1-introduction",
		},
		{
			name:     "multiple spaces",
			text:     "Hello   World",
			expected: "hello-world",
		},
		{
			name:     "leading/trailing spaces",
			text:     "  Hello  ",
			expected: "hello",
		},
		{
			name:     "only special chars",
			text:     "!@#$%",
			expected: "heading",
		},
		{
			name:     "empty string",
			text:     "",
			expected: "heading",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateHeadingID(tt.text)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestTOCFromEntries(t *testing.T) {
	entries := []*TOCEntry{
		{Level: 2, ID: "one", Text: "One"},
		{Level: 3, ID: "one-a", Text: "One A"},
		{Level: 2, ID: "two", Text: "Two"},
	}

	result := TOCFromEntries(entries)

	if result == "" {
		t.Error("expected non-empty result")
	}

	if !strings.Contains(result, `<a class="toc-link" href="#one">One</a>`) {
		t.Error("expected TOC to contain link to #one")
	}

	if !strings.Contains(result, `<a class="toc-link" href="#one-a">One A</a>`) {
		t.Error("expected TOC to contain link to #one-a")
	}
}

func TestTOCNesting(t *testing.T) {
	html := `<h2 id="section1">Section 1</h2>
<h3 id="sub1-1">Sub 1.1</h3>
<h3 id="sub1-2">Sub 1.2</h3>
<h4 id="sub1-2-1">Sub 1.2.1</h4>
<h2 id="section2">Section 2</h2>`

	result := GenerateTOC(html, 2, 4)

	// Count nested ul tags to verify nesting
	ulCount := strings.Count(result, `<ul class="toc-list">`)
	if ulCount < 2 {
		t.Errorf("expected at least 2 nested <ul> tags for proper nesting, got %d", ulCount)
	}

	// Verify structure contains all links
	expectedLinks := []string{
		"#section1",
		"#sub1-1",
		"#sub1-2",
		"#sub1-2-1",
		"#section2",
	}

	for _, link := range expectedLinks {
		if !strings.Contains(result, link) {
			t.Errorf("expected TOC to contain link '%s'", link)
		}
	}
}

func TestTOCWithScrollSpy(t *testing.T) {
	html := `<h2 id="intro">Introduction</h2><h3 id="details">Details</h3>`

	// Without tocID — no data attributes
	plain := GenerateTOC(html, 2, 4)
	if strings.Contains(plain, "data-tui-toc-link") {
		t.Error("expected no data-tui-toc-link without tocID")
	}

	// With tocID — data attributes present
	result := GenerateTOC(html, 2, 4, "page-toc")

	wantContain := []string{
		`data-tui-toc-link`,
		`data-tui-toc-href="#intro"`,
		`data-tui-toc-id="page-toc"`,
		`data-tui-toc-href="#details"`,
	}
	for _, want := range wantContain {
		if !strings.Contains(result, want) {
			t.Errorf("expected TOC to contain '%s', got:\n%s", want, result)
		}
	}
}

func TestTOCFromEntriesWithScrollSpy(t *testing.T) {
	entries := []*TOCEntry{
		{Level: 2, ID: "one", Text: "One"},
		{Level: 3, ID: "two", Text: "Two"},
	}

	result := TOCFromEntries(entries, "my-toc")
	if !strings.Contains(result, `data-tui-toc-id="my-toc"`) {
		t.Errorf("expected data-tui-toc-id, got:\n%s", result)
	}
	if !strings.Contains(result, `data-tui-toc-href="#one"`) {
		t.Errorf("expected data-tui-toc-href, got:\n%s", result)
	}
}
