// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package site

import (
	"testing"
	"time"

	"github.com/dmoose/templsite/pkg/content"
)

func TestSectionPageCount(t *testing.T) {
	section := &Section{
		Name: "blog",
		URL:  "/blog/",
		Pages: []*content.Page{
			{Title: "Post 1", URL: "/blog/post1/"},
			{Title: "Post 2", URL: "/blog/post2/"},
			{Title: "Blog Index", URL: "/blog/"}, // index page
		},
	}

	count := section.PageCount()
	if count != 2 {
		t.Errorf("expected 2 regular pages, got %d", count)
	}
}

func TestSectionRegularPages(t *testing.T) {
	section := &Section{
		Name: "blog",
		URL:  "/blog/",
		Pages: []*content.Page{
			{Title: "Post 1", URL: "/blog/post1/"},
			{Title: "Blog Index", URL: "/blog/"},
			{Title: "Post 2", URL: "/blog/post2/"},
		},
	}

	regularPages := section.RegularPages()
	if len(regularPages) != 2 {
		t.Errorf("expected 2 regular pages, got %d", len(regularPages))
	}

	// Verify index page is excluded
	for _, p := range regularPages {
		if p.URL == "/blog/" {
			t.Error("index page should be excluded from RegularPages")
		}
	}
}

func TestSectionSortPages(t *testing.T) {
	now := time.Now()
	section := &Section{
		Name: "blog",
		URL:  "/blog/",
		Pages: []*content.Page{
			{Title: "Old Post", URL: "/blog/old/", Date: now.Add(-48 * time.Hour)},
			{Title: "New Post", URL: "/blog/new/", Date: now},
			{Title: "Mid Post", URL: "/blog/mid/", Date: now.Add(-24 * time.Hour)},
		},
	}

	section.sortPages()

	// Should be sorted newest first
	if section.Pages[0].Title != "New Post" {
		t.Errorf("expected newest post first, got %s", section.Pages[0].Title)
	}
	if section.Pages[1].Title != "Mid Post" {
		t.Errorf("expected mid post second, got %s", section.Pages[1].Title)
	}
	if section.Pages[2].Title != "Old Post" {
		t.Errorf("expected old post last, got %s", section.Pages[2].Title)
	}
}

func TestSectionSortByWeight(t *testing.T) {
	// Same date, different weights
	date := time.Now()
	section := &Section{
		Name: "docs",
		URL:  "/docs/",
		Pages: []*content.Page{
			{Title: "Third", URL: "/docs/third/", Date: date, Weight: 30},
			{Title: "First", URL: "/docs/first/", Date: date, Weight: 10},
			{Title: "Second", URL: "/docs/second/", Date: date, Weight: 20},
		},
	}

	section.sortPages()

	// Same date, so should be sorted by weight ascending
	if section.Pages[0].Title != "First" {
		t.Errorf("expected weight 10 first, got %s (weight %d)", section.Pages[0].Title, section.Pages[0].Weight)
	}
	if section.Pages[1].Title != "Second" {
		t.Errorf("expected weight 20 second, got %s", section.Pages[1].Title)
	}
	if section.Pages[2].Title != "Third" {
		t.Errorf("expected weight 30 last, got %s", section.Pages[2].Title)
	}
}

func TestSectionLinkPrevNext(t *testing.T) {
	now := time.Now()
	section := &Section{
		Name: "blog",
		URL:  "/blog/",
		Pages: []*content.Page{
			{Title: "New Post", URL: "/blog/new/", Date: now},
			{Title: "Mid Post", URL: "/blog/mid/", Date: now.Add(-24 * time.Hour)},
			{Title: "Old Post", URL: "/blog/old/", Date: now.Add(-48 * time.Hour)},
		},
	}

	section.sortPages()
	section.linkPrevNext()

	// New Post (first) - no Prev, Next = Mid
	if section.Pages[0].Prev != nil {
		t.Error("newest post should have no Prev")
	}
	if section.Pages[0].Next == nil || section.Pages[0].Next.Title != "Mid Post" {
		t.Error("newest post's Next should be Mid Post")
	}

	// Mid Post - Prev = New, Next = Old
	if section.Pages[1].Prev == nil || section.Pages[1].Prev.Title != "New Post" {
		t.Error("mid post's Prev should be New Post")
	}
	if section.Pages[1].Next == nil || section.Pages[1].Next.Title != "Old Post" {
		t.Error("mid post's Next should be Old Post")
	}

	// Old Post (last) - Prev = Mid, no Next
	if section.Pages[2].Prev == nil || section.Pages[2].Prev.Title != "Mid Post" {
		t.Error("old post's Prev should be Mid Post")
	}
	if section.Pages[2].Next != nil {
		t.Error("oldest post should have no Next")
	}
}

func TestSectionLinkPrevNextExcludesIndex(t *testing.T) {
	now := time.Now()
	section := &Section{
		Name: "blog",
		URL:  "/blog/",
		Pages: []*content.Page{
			{Title: "Blog Index", URL: "/blog/", Date: now},
			{Title: "Post 1", URL: "/blog/post1/", Date: now.Add(-24 * time.Hour)},
			{Title: "Post 2", URL: "/blog/post2/", Date: now.Add(-48 * time.Hour)},
		},
	}

	section.sortPages()
	section.linkPrevNext()

	// Find the regular pages
	var post1, post2 *content.Page
	for _, p := range section.Pages {
		if p.Title == "Post 1" {
			post1 = p
		}
		if p.Title == "Post 2" {
			post2 = p
		}
	}

	// Post 1 should be first regular page - no Prev, Next = Post 2
	if post1.Prev != nil {
		t.Error("Post 1 should have no Prev (index page excluded)")
	}
	if post1.Next == nil || post1.Next.Title != "Post 2" {
		t.Error("Post 1's Next should be Post 2")
	}

	// Post 2 should be last - Prev = Post 1, no Next
	if post2.Prev == nil || post2.Prev.Title != "Post 1" {
		t.Error("Post 2's Prev should be Post 1")
	}
	if post2.Next != nil {
		t.Error("Post 2 should have no Next")
	}
}
