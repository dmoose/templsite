// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

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

	if rss := site.RSS(nil, "Empty Blog", "No posts yet"); rss != "" {
		t.Errorf("RSS with no pages should be suppressed, got %q", rss)
	}
}

func TestRSSSuppressedWithoutDatedPages(t *testing.T) {
	config := DefaultConfig()
	config.BaseURL = "https://example.com"

	site := NewWithConfig(config)

	pages := []*content.Page{
		{Title: "About", URL: "/about/"},
		{Title: "Contact", URL: "/contact/"},
	}

	if rss := site.RSS(pages, "Product Site", "No blog here"); rss != "" {
		t.Errorf("RSS with only undated pages should be suppressed, got %q", rss)
	}
}

func TestRSSSkipsUndatedPages(t *testing.T) {
	config := DefaultConfig()
	config.BaseURL = "https://example.com"

	site := NewWithConfig(config)

	pages := []*content.Page{
		{Title: "Dated Post", URL: "/blog/dated/", Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		{Title: "Undated Page", URL: "/about/"},
	}

	rss := site.RSS(pages, "My Blog", "A great blog")

	if !strings.Contains(rss, "<title>Dated Post</title>") {
		t.Error("RSS should contain the dated post")
	}
	if strings.Contains(rss, "Undated Page") {
		t.Error("RSS should not contain an undated page")
	}
}

func TestRSSLastBuildDateIsNewestItem(t *testing.T) {
	config := DefaultConfig()
	config.BaseURL = "https://example.com"

	site := NewWithConfig(config)

	newest := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	pages := []*content.Page{
		{Title: "Older", URL: "/blog/older/", Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		{Title: "Newest", URL: "/blog/newest/", Date: newest},
		{Title: "Middle", URL: "/blog/middle/", Date: time.Date(2024, 3, 2, 0, 0, 0, 0, time.UTC)},
	}

	rss := site.RSS(pages, "My Blog", "A great blog")

	want := "<lastBuildDate>" + newest.Format(time.RFC1123Z) + "</lastBuildDate>"
	if !strings.Contains(rss, want) {
		t.Errorf("RSS lastBuildDate should be the newest item date %s", newest.Format(time.RFC1123Z))
	}
}

func TestAtomSuppressedWithoutDatedPages(t *testing.T) {
	config := DefaultConfig()
	config.BaseURL = "https://example.com"

	site := NewWithConfig(config)

	pages := []*content.Page{{Title: "About", URL: "/about/"}}

	if atom := site.Atom(pages, "Product Site", "No blog here"); atom != "" {
		t.Errorf("Atom with only undated pages should be suppressed, got %q", atom)
	}
}

func TestAtomUpdatedIsNewestEntry(t *testing.T) {
	config := DefaultConfig()
	config.BaseURL = "https://example.com"

	site := NewWithConfig(config)

	newest := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	pages := []*content.Page{
		{Title: "Older", URL: "/blog/older/", Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		{Title: "Newest", URL: "/blog/newest/", Date: newest},
		{Title: "Undated", URL: "/about/"},
	}

	atom := site.Atom(pages, "My Blog", "A great blog")

	want := "<updated>" + newest.Format(time.RFC3339) + "</updated>"
	if !strings.Contains(atom, want) {
		t.Errorf("Atom feed <updated> should be the newest entry date %s", newest.Format(time.RFC3339))
	}
	if strings.Contains(atom, "Undated") {
		t.Error("Atom should not contain an undated page")
	}
}

func TestJSONSuppressedWithoutDatedPages(t *testing.T) {
	config := DefaultConfig()
	config.BaseURL = "https://example.com"

	site := NewWithConfig(config)

	pages := []*content.Page{{Title: "About", URL: "/about/"}}

	if feed := site.JSON(pages, "Product Site", "No blog here"); feed != "" {
		t.Errorf("JSON feed with only undated pages should be suppressed, got %q", feed)
	}
}

func TestJSONSkipsUndatedPages(t *testing.T) {
	config := DefaultConfig()
	config.BaseURL = "https://example.com"

	site := NewWithConfig(config)

	pages := []*content.Page{
		{Title: "Dated Post", URL: "/blog/dated/", Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		{Title: "Undated Page", URL: "/about/"},
	}

	feed := site.JSON(pages, "My Blog", "A great blog")

	if !strings.Contains(feed, `"title": "Dated Post"`) {
		t.Error("JSON feed should contain the dated post")
	}
	if strings.Contains(feed, "Undated Page") {
		t.Error("JSON feed should not contain an undated page")
	}
}

// TestFeedWritersAreClockIndependent asserts the property the whole change
// exists for: the same content rendered at two different moments produces
// identical feed bytes.
func TestFeedWritersAreClockIndependent(t *testing.T) {
	config := DefaultConfig()
	config.BaseURL = "https://example.com"

	site := NewWithConfig(config)

	pages := []*content.Page{
		{Title: "First Post", URL: "/blog/first/", Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		{Title: "Second Post", URL: "/blog/second/", Date: time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC)},
	}

	writers := map[string]func() string{
		"RSS":  func() string { return site.RSS(pages, "My Blog", "A great blog") },
		"Atom": func() string { return site.Atom(pages, "My Blog", "A great blog") },
		"JSON": func() string { return site.JSON(pages, "My Blog", "A great blog") },
	}

	first := make(map[string]string, len(writers))
	for name, write := range writers {
		out := write()
		if out == "" {
			t.Fatalf("%s produced no output for dated pages", name)
		}
		first[name] = out
	}

	// The coarsest timestamp format in play resolves to one second, so a build
	// that still consulted the clock would differ across this pause.
	time.Sleep(1100 * time.Millisecond)

	for name, write := range writers {
		if second := write(); first[name] != second {
			t.Errorf("%s output changed between calls; wall-clock time is leaking into the feed", name)
		}
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
