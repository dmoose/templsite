// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package site

import (
	"strings"
	"testing"
	"time"

	"github.com/dmoose/templsite/pkg/content"
)

func TestFeedDatedPagesOnly(t *testing.T) {
	config := DefaultConfig()
	config.BaseURL = "https://example.com"
	config.Title = "My Site"

	site := NewWithConfig(config)
	site.Pages = []*content.Page{
		{Title: "Dated Post", URL: "/blog/dated/", Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		{Title: "Undated Page", URL: "/about/"},
	}

	feed := site.Feed()

	if !strings.Contains(feed, "<title>Dated Post</title>") {
		t.Error("feed should contain the dated post")
	}
	if strings.Contains(feed, "Undated Page") {
		t.Error("feed should not contain an undated page")
	}
}

func TestFeedUpdatedIsNewestEntry(t *testing.T) {
	config := DefaultConfig()
	config.BaseURL = "https://example.com"

	newest := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

	site := NewWithConfig(config)
	site.Pages = []*content.Page{
		{Title: "Older", URL: "/blog/older/", Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		{Title: "Newest", URL: "/blog/newest/", Date: newest},
		{Title: "Middle", URL: "/blog/middle/", Date: time.Date(2024, 3, 2, 0, 0, 0, 0, time.UTC)},
	}

	feed := site.Feed()

	want := "<updated>" + newest.Format(time.RFC3339) + "</updated>"
	if !strings.Contains(feed, want) {
		t.Errorf("feed-level <updated> should be the newest entry date %s", newest.Format(time.RFC3339))
	}

	// The feed-level <updated> is the first one in the document; it must not be
	// a value the clock supplied.
	if strings.Index(feed, want) > strings.Index(feed, "<entry>") {
		t.Error("feed-level <updated> should precede the first entry")
	}
}

func TestFeedSuppressedWithoutDatedPages(t *testing.T) {
	config := DefaultConfig()
	config.BaseURL = "https://example.com"

	site := NewWithConfig(config)
	site.Pages = []*content.Page{
		{Title: "About", URL: "/about/"},
		{Title: "Contact", URL: "/contact/"},
	}

	if feed := site.Feed(); feed != "" {
		t.Errorf("feed should be suppressed when no page is dated, got %q", feed)
	}
}

func TestFeedSuppressedWithNoPages(t *testing.T) {
	config := DefaultConfig()
	config.BaseURL = "https://example.com"

	site := NewWithConfig(config)

	if feed := site.Feed(); feed != "" {
		t.Errorf("feed should be suppressed for an empty site, got %q", feed)
	}
}

func TestFeedIgnoresBuildTime(t *testing.T) {
	config := DefaultConfig()
	config.BaseURL = "https://example.com"

	pageDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	site := NewWithConfig(config)
	site.Pages = []*content.Page{
		{Title: "Post", URL: "/blog/post/", Date: pageDate},
	}

	site.BuildTime = time.Date(2030, 12, 25, 0, 0, 0, 0, time.UTC)
	withBuildTime := site.Feed()

	site.BuildTime = time.Time{}
	withoutBuildTime := site.Feed()

	if withBuildTime != withoutBuildTime {
		t.Error("feed output should not depend on Site.BuildTime")
	}
	if strings.Contains(withBuildTime, "2030") {
		t.Error("feed should not contain the build timestamp")
	}
}
