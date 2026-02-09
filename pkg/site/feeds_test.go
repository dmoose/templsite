package site

import (
	"strings"
	"testing"
	"time"

	"github.com/dmoose/templsite/pkg/content"
)

func TestRSS(t *testing.T) {
	config := DefaultConfig()
	config.BaseURL = "https://example.com"
	config.Language = "en-us"

	site := NewWithConfig(config)

	pages := []*content.Page{
		{
			Title:       "First Post",
			URL:         "/blog/first/",
			Description: "The first post description",
			Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Author:      "Jane Doe",
		},
		{
			Title:   "Second Post",
			URL:     "/blog/second/",
			Summary: "Auto-generated summary",
			Date:    time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
		},
	}

	rss := site.RSS(pages, "My Blog", "A great blog")

	// Check XML declaration
	if !strings.HasPrefix(rss, "<?xml") {
		t.Error("RSS should start with XML declaration")
	}

	// Check required elements
	if !strings.Contains(rss, "<rss version=\"2.0\"") {
		t.Error("RSS should have version 2.0")
	}
	if !strings.Contains(rss, "<title>My Blog</title>") {
		t.Error("RSS should contain channel title")
	}
	if !strings.Contains(rss, "<link>https://example.com</link>") {
		t.Error("RSS should contain channel link")
	}
	if !strings.Contains(rss, "<description>A great blog</description>") {
		t.Error("RSS should contain channel description")
	}
	if !strings.Contains(rss, "<language>en-us</language>") {
		t.Error("RSS should contain language")
	}

	// Check atom:link for self-reference
	if !strings.Contains(rss, "atom:link") {
		t.Error("RSS should contain atom:link self-reference")
	}

	// Check items
	if !strings.Contains(rss, "<title>First Post</title>") {
		t.Error("RSS should contain first post title")
	}
	if !strings.Contains(rss, "<link>https://example.com/blog/first/</link>") {
		t.Error("RSS should contain first post link")
	}
	if !strings.Contains(rss, "<description>The first post description</description>") {
		t.Error("RSS should contain first post description")
	}
	if !strings.Contains(rss, "<author>Jane Doe</author>") {
		t.Error("RSS should contain author")
	}

	// Check second post uses summary when no description
	if !strings.Contains(rss, "<description>Auto-generated summary</description>") {
		t.Error("RSS should use summary when description is empty")
	}
}

func TestAtom(t *testing.T) {
	config := DefaultConfig()
	config.BaseURL = "https://example.com"

	site := NewWithConfig(config)

	pages := []*content.Page{
		{
			Title:       "First Post",
			URL:         "/blog/first/",
			Description: "The first post",
			Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Author:      "Jane Doe",
		},
	}

	atom := site.Atom(pages, "My Blog", "A great blog")

	// Check XML declaration
	if !strings.HasPrefix(atom, "<?xml") {
		t.Error("Atom should start with XML declaration")
	}

	// Check Atom namespace
	if !strings.Contains(atom, "xmlns=\"http://www.w3.org/2005/Atom\"") {
		t.Error("Atom should have Atom namespace")
	}

	// Check required elements
	if !strings.Contains(atom, "<title>My Blog</title>") {
		t.Error("Atom should contain feed title")
	}

	// Check entry
	if !strings.Contains(atom, "<title>First Post</title>") {
		t.Error("Atom should contain entry title")
	}
	if !strings.Contains(atom, "https://example.com/blog/first/") {
		t.Error("Atom should contain entry link")
	}
	if !strings.Contains(atom, "<summary>The first post</summary>") {
		t.Error("Atom should contain entry summary")
	}
	if !strings.Contains(atom, "<name>Jane Doe</name>") {
		t.Error("Atom should contain author name")
	}
}

func TestJSON(t *testing.T) {
	config := DefaultConfig()
	config.BaseURL = "https://example.com"
	config.Language = "en"

	site := NewWithConfig(config)

	pages := []*content.Page{
		{
			Title:       "First Post",
			URL:         "/blog/first/",
			Description: "The first post",
			Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Author:      "Jane Doe",
		},
	}

	json := site.JSON(pages, "My Blog", "A great blog")

	// Check JSON Feed version
	if !strings.Contains(json, "https://jsonfeed.org/version/1.1") {
		t.Error("JSON Feed should have version 1.1")
	}

	// Check required fields
	if !strings.Contains(json, `"title": "My Blog"`) {
		t.Error("JSON Feed should contain title")
	}
	if !strings.Contains(json, `"home_page_url": "https://example.com"`) {
		t.Error("JSON Feed should contain home_page_url")
	}
	if !strings.Contains(json, `"feed_url": "https://example.com/feed.json"`) {
		t.Error("JSON Feed should contain feed_url")
	}
	if !strings.Contains(json, `"language": "en"`) {
		t.Error("JSON Feed should contain language")
	}

	// Check items
	if !strings.Contains(json, `"title": "First Post"`) {
		t.Error("JSON Feed should contain item title")
	}
	if !strings.Contains(json, `"url": "https://example.com/blog/first/"`) {
		t.Error("JSON Feed should contain item url")
	}
}

func TestRSSWithEmptyPages(t *testing.T) {
	config := DefaultConfig()
	config.BaseURL = "https://example.com"

	site := NewWithConfig(config)

	rss := site.RSS(nil, "Empty Blog", "No posts yet")

	if !strings.Contains(rss, "<title>Empty Blog</title>") {
		t.Error("RSS should work with empty pages")
	}
	if !strings.Contains(rss, "</channel>") {
		t.Error("RSS should have closing channel tag")
	}
}

func TestEscapeJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{`say "hello"`, `say \"hello\"`},
		{"line1\nline2", `line1\nline2`},
		{"path\\to\\file", `path\\to\\file`},
		{"tab\there", `tab\there`},
	}

	for _, tt := range tests {
		result := escapeJSON(tt.input)
		if result != tt.expected {
			t.Errorf("escapeJSON(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
