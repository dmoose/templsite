// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package site

import (
	"testing"
	"time"

	"github.com/dmoose/templsite/pkg/content"
)

func TestSortByDate(t *testing.T) {
	now := time.Now()
	pages := []*content.Page{
		{Title: "Middle", Date: now.Add(-24 * time.Hour)},
		{Title: "Oldest", Date: now.Add(-48 * time.Hour)},
		{Title: "Newest", Date: now},
	}

	// Test descending (newest first)
	desc := SortByDate(pages, false)
	if desc[0].Title != "Newest" {
		t.Errorf("expected 'Newest' first, got '%s'", desc[0].Title)
	}
	if desc[2].Title != "Oldest" {
		t.Errorf("expected 'Oldest' last, got '%s'", desc[2].Title)
	}

	// Test ascending (oldest first)
	asc := SortByDate(pages, true)
	if asc[0].Title != "Oldest" {
		t.Errorf("expected 'Oldest' first, got '%s'", asc[0].Title)
	}
	if asc[2].Title != "Newest" {
		t.Errorf("expected 'Newest' last, got '%s'", asc[2].Title)
	}

	// Verify original slice is unchanged
	if pages[0].Title != "Middle" {
		t.Error("original slice was modified")
	}
}

func TestSortByTitle(t *testing.T) {
	pages := []*content.Page{
		{Title: "Zebra"},
		{Title: "Apple"},
		{Title: "Mango"},
	}

	sorted := SortByTitle(pages)

	if sorted[0].Title != "Apple" {
		t.Errorf("expected 'Apple' first, got '%s'", sorted[0].Title)
	}
	if sorted[1].Title != "Mango" {
		t.Errorf("expected 'Mango' second, got '%s'", sorted[1].Title)
	}
	if sorted[2].Title != "Zebra" {
		t.Errorf("expected 'Zebra' last, got '%s'", sorted[2].Title)
	}

	// Verify original unchanged
	if pages[0].Title != "Zebra" {
		t.Error("original slice was modified")
	}
}

func TestSortByWeight(t *testing.T) {
	now := time.Now()
	pages := []*content.Page{
		{Title: "Weight 20", Weight: 20, Date: now},
		{Title: "Weight 10 New", Weight: 10, Date: now},
		{Title: "Weight 10 Old", Weight: 10, Date: now.Add(-24 * time.Hour)},
	}

	sorted := SortByWeight(pages)

	// Weight 10 items first (sorted by date within same weight)
	if sorted[0].Title != "Weight 10 New" {
		t.Errorf("expected 'Weight 10 New' first, got '%s'", sorted[0].Title)
	}
	if sorted[1].Title != "Weight 10 Old" {
		t.Errorf("expected 'Weight 10 Old' second, got '%s'", sorted[1].Title)
	}
	if sorted[2].Title != "Weight 20" {
		t.Errorf("expected 'Weight 20' last, got '%s'", sorted[2].Title)
	}
}

func TestFilter(t *testing.T) {
	pages := []*content.Page{
		{Title: "Go Post", Tags: []string{"go"}},
		{Title: "Rust Post", Tags: []string{"rust"}},
		{Title: "Another Go", Tags: []string{"go", "web"}},
	}

	// Filter for pages with "go" tag
	goPages := Filter(pages, func(p *content.Page) bool {
		return p.HasTag("go")
	})

	if len(goPages) != 2 {
		t.Errorf("expected 2 go pages, got %d", len(goPages))
	}

	// Filter for empty result
	pythonPages := Filter(pages, func(p *content.Page) bool {
		return p.HasTag("python")
	})

	if len(pythonPages) != 0 {
		t.Errorf("expected 0 python pages, got %d", len(pythonPages))
	}
}

func TestLimit(t *testing.T) {
	pages := []*content.Page{
		{Title: "A"},
		{Title: "B"},
		{Title: "C"},
		{Title: "D"},
	}

	// Normal limit
	limited := Limit(pages, 2)
	if len(limited) != 2 {
		t.Errorf("expected 2 pages, got %d", len(limited))
	}
	if limited[0].Title != "A" || limited[1].Title != "B" {
		t.Error("unexpected page order")
	}

	// Limit larger than slice
	all := Limit(pages, 10)
	if len(all) != 4 {
		t.Errorf("expected 4 pages, got %d", len(all))
	}

	// Zero limit
	zero := Limit(pages, 0)
	if zero != nil {
		t.Error("expected nil for zero limit")
	}

	// Negative limit
	negative := Limit(pages, -1)
	if negative != nil {
		t.Error("expected nil for negative limit")
	}
}

func TestOffset(t *testing.T) {
	pages := []*content.Page{
		{Title: "A"},
		{Title: "B"},
		{Title: "C"},
		{Title: "D"},
	}

	// Normal offset
	offset := Offset(pages, 2)
	if len(offset) != 2 {
		t.Errorf("expected 2 pages, got %d", len(offset))
	}
	if offset[0].Title != "C" {
		t.Errorf("expected 'C' first, got '%s'", offset[0].Title)
	}

	// Zero offset
	noOffset := Offset(pages, 0)
	if len(noOffset) != 4 {
		t.Errorf("expected 4 pages, got %d", len(noOffset))
	}

	// Offset beyond length
	beyond := Offset(pages, 10)
	if beyond != nil {
		t.Error("expected nil for offset beyond length")
	}
}

func TestPaginate(t *testing.T) {
	pages := make([]*content.Page, 25)
	for i := range 25 {
		pages[i] = &content.Page{Title: string(rune('A' + i))}
	}

	// First page
	page1 := Paginate(pages, 1, 10)
	if len(page1) != 10 {
		t.Errorf("expected 10 pages on first page, got %d", len(page1))
	}
	if page1[0].Title != "A" {
		t.Errorf("expected 'A' first, got '%s'", page1[0].Title)
	}

	// Second page
	page2 := Paginate(pages, 2, 10)
	if len(page2) != 10 {
		t.Errorf("expected 10 pages on second page, got %d", len(page2))
	}
	if page2[0].Title != "K" {
		t.Errorf("expected 'K' first on page 2, got '%s'", page2[0].Title)
	}

	// Last page (partial)
	page3 := Paginate(pages, 3, 10)
	if len(page3) != 5 {
		t.Errorf("expected 5 pages on last page, got %d", len(page3))
	}

	// Page beyond range
	page4 := Paginate(pages, 4, 10)
	if page4 != nil {
		t.Error("expected nil for page beyond range")
	}

	// Invalid page number
	page0 := Paginate(pages, 0, 10)
	if len(page0) != 10 {
		t.Errorf("expected page 0 to be treated as page 1, got %d items", len(page0))
	}
}

func TestFirst(t *testing.T) {
	pages := []*content.Page{
		{Title: "A"},
		{Title: "B"},
	}

	first := First(pages)
	if first == nil || first.Title != "A" {
		t.Error("expected first page to be 'A'")
	}

	// Empty slice
	if First(nil) != nil {
		t.Error("expected nil for empty slice")
	}
	if First([]*content.Page{}) != nil {
		t.Error("expected nil for empty slice")
	}
}

func TestLast(t *testing.T) {
	pages := []*content.Page{
		{Title: "A"},
		{Title: "B"},
	}

	last := Last(pages)
	if last == nil || last.Title != "B" {
		t.Error("expected last page to be 'B'")
	}

	// Empty slice
	if Last(nil) != nil {
		t.Error("expected nil for empty slice")
	}
}

func TestReverse(t *testing.T) {
	pages := []*content.Page{
		{Title: "A"},
		{Title: "B"},
		{Title: "C"},
	}

	reversed := Reverse(pages)

	if reversed[0].Title != "C" {
		t.Errorf("expected 'C' first, got '%s'", reversed[0].Title)
	}
	if reversed[2].Title != "A" {
		t.Errorf("expected 'A' last, got '%s'", reversed[2].Title)
	}

	// Original unchanged
	if pages[0].Title != "A" {
		t.Error("original slice was modified")
	}
}

func TestGroupBySection(t *testing.T) {
	pages := []*content.Page{
		{Title: "Blog 1", Section: "blog"},
		{Title: "Blog 2", Section: "blog"},
		{Title: "Doc 1", Section: "docs"},
		{Title: "Root", Section: ""},
	}

	groups := GroupBySection(pages)

	if len(groups["blog"]) != 2 {
		t.Errorf("expected 2 blog pages, got %d", len(groups["blog"]))
	}
	if len(groups["docs"]) != 1 {
		t.Errorf("expected 1 docs page, got %d", len(groups["docs"]))
	}
	if len(groups["_root"]) != 1 {
		t.Errorf("expected 1 root page, got %d", len(groups["_root"]))
	}
}
