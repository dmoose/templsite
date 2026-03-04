package site

import (
	"testing"

	"git.catapulsion.com/templsite/pkg/content"
)

func TestNewPaginator(t *testing.T) {
	pages := make([]*content.Page, 25)
	for i := range pages {
		pages[i] = &content.Page{Title: string(rune('A' + i))}
	}

	pager := NewPaginator(pages, 10, "/blog/")

	if pager.TotalItems != 25 {
		t.Errorf("expected TotalItems 25, got %d", pager.TotalItems)
	}
	if pager.TotalPages != 3 {
		t.Errorf("expected TotalPages 3, got %d", pager.TotalPages)
	}
	if pager.PerPage != 10 {
		t.Errorf("expected PerPage 10, got %d", pager.PerPage)
	}
	if pager.PageNum != 1 {
		t.Errorf("expected PageNum 1, got %d", pager.PageNum)
	}
	if pager.HasPrev {
		t.Error("expected HasPrev false for first page")
	}
	if !pager.HasNext {
		t.Error("expected HasNext true for first page")
	}
	if pager.BaseURL != "/blog/" {
		t.Errorf("expected BaseURL '/blog/', got '%s'", pager.BaseURL)
	}
}

func TestPaginatorPageURLs(t *testing.T) {
	pages := make([]*content.Page, 25)
	for i := range pages {
		pages[i] = &content.Page{Title: string(rune('A' + i))}
	}

	pager := NewPaginator(pages, 10, "/blog/")

	if len(pager.PageURLs) != 3 {
		t.Errorf("expected 3 page URLs, got %d", len(pager.PageURLs))
	}
	if pager.PageURLs[0] != "/blog/" {
		t.Errorf("expected first page URL '/blog/', got '%s'", pager.PageURLs[0])
	}
	if pager.PageURLs[1] != "/blog/page/2/" {
		t.Errorf("expected second page URL '/blog/page/2/', got '%s'", pager.PageURLs[1])
	}
	if pager.PageURLs[2] != "/blog/page/3/" {
		t.Errorf("expected third page URL '/blog/page/3/', got '%s'", pager.PageURLs[2])
	}
}

func TestPaginatorPage(t *testing.T) {
	pages := make([]*content.Page, 25)
	for i := range pages {
		pages[i] = &content.Page{Title: string(rune('A' + i))}
	}

	pager := NewPaginator(pages, 10, "/blog/")

	// Test first page
	page1 := pager.Page(1)
	if len(page1.Items) != 10 {
		t.Errorf("expected 10 items on page 1, got %d", len(page1.Items))
	}
	if page1.Items[0].Title != "A" {
		t.Errorf("expected first item 'A', got '%s'", page1.Items[0].Title)
	}
	if page1.HasPrev {
		t.Error("expected HasPrev false for page 1")
	}
	if !page1.HasNext {
		t.Error("expected HasNext true for page 1")
	}
	if page1.PrevURL != "" {
		t.Errorf("expected empty PrevURL for page 1, got '%s'", page1.PrevURL)
	}
	if page1.NextURL != "/blog/page/2/" {
		t.Errorf("expected NextURL '/blog/page/2/', got '%s'", page1.NextURL)
	}

	// Test second page
	page2 := pager.Page(2)
	if len(page2.Items) != 10 {
		t.Errorf("expected 10 items on page 2, got %d", len(page2.Items))
	}
	if page2.Items[0].Title != "K" {
		t.Errorf("expected first item 'K', got '%s'", page2.Items[0].Title)
	}
	if !page2.HasPrev {
		t.Error("expected HasPrev true for page 2")
	}
	if !page2.HasNext {
		t.Error("expected HasNext true for page 2")
	}
	if page2.PrevURL != "/blog/" {
		t.Errorf("expected PrevURL '/blog/', got '%s'", page2.PrevURL)
	}
	if page2.NextURL != "/blog/page/3/" {
		t.Errorf("expected NextURL '/blog/page/3/', got '%s'", page2.NextURL)
	}

	// Test last page (partial)
	page3 := pager.Page(3)
	if len(page3.Items) != 5 {
		t.Errorf("expected 5 items on page 3, got %d", len(page3.Items))
	}
	if page3.Items[0].Title != "U" {
		t.Errorf("expected first item 'U', got '%s'", page3.Items[0].Title)
	}
	if !page3.HasPrev {
		t.Error("expected HasPrev true for page 3")
	}
	if page3.HasNext {
		t.Error("expected HasNext false for last page")
	}
	if page3.PrevURL != "/blog/page/2/" {
		t.Errorf("expected PrevURL '/blog/page/2/', got '%s'", page3.PrevURL)
	}
	if page3.NextURL != "" {
		t.Errorf("expected empty NextURL for last page, got '%s'", page3.NextURL)
	}
}

func TestPaginatorInvalidPageNumbers(t *testing.T) {
	pages := make([]*content.Page, 25)
	for i := range pages {
		pages[i] = &content.Page{Title: string(rune('A' + i))}
	}

	pager := NewPaginator(pages, 10, "/blog/")

	// Page 0 should be treated as page 1
	page0 := pager.Page(0)
	if page0.PageNum != 1 {
		t.Errorf("expected page 0 to become page 1, got %d", page0.PageNum)
	}

	// Page beyond range should be clamped to last page
	page10 := pager.Page(10)
	if page10.PageNum != 3 {
		t.Errorf("expected page 10 to become page 3, got %d", page10.PageNum)
	}
}

func TestPaginatorEmptyCollection(t *testing.T) {
	pager := NewPaginator(nil, 10, "/blog/")

	if pager.TotalItems != 0 {
		t.Errorf("expected TotalItems 0, got %d", pager.TotalItems)
	}
	if pager.TotalPages != 1 {
		t.Errorf("expected TotalPages 1 for empty collection, got %d", pager.TotalPages)
	}

	page1 := pager.Page(1)
	if len(page1.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(page1.Items))
	}
}

func TestPaginatorBaseURLNormalization(t *testing.T) {
	pages := []*content.Page{{Title: "A"}}

	// Without trailing slash
	pager := NewPaginator(pages, 10, "/blog")
	if pager.BaseURL != "/blog/" {
		t.Errorf("expected BaseURL '/blog/', got '%s'", pager.BaseURL)
	}

	// With trailing slash
	pager2 := NewPaginator(pages, 10, "/blog/")
	if pager2.BaseURL != "/blog/" {
		t.Errorf("expected BaseURL '/blog/', got '%s'", pager2.BaseURL)
	}
}

func TestPaginatorHelperMethods(t *testing.T) {
	pages := make([]*content.Page, 25)
	for i := range pages {
		pages[i] = &content.Page{Title: string(rune('A' + i))}
	}

	pager := NewPaginator(pages, 10, "/blog/")

	// Test First
	if pager.First() != 1 {
		t.Errorf("expected First() 1, got %d", pager.First())
	}

	// Test Last
	if pager.Last() != 3 {
		t.Errorf("expected Last() 3, got %d", pager.Last())
	}

	// Test Pages
	pageNums := pager.Pages()
	if len(pageNums) != 3 {
		t.Errorf("expected 3 page numbers, got %d", len(pageNums))
	}
	if pageNums[0] != 1 || pageNums[1] != 2 || pageNums[2] != 3 {
		t.Errorf("expected [1, 2, 3], got %v", pageNums)
	}

	// Test URL
	if pager.URL(1) != "/blog/" {
		t.Errorf("expected URL(1) '/blog/', got '%s'", pager.URL(1))
	}
	if pager.URL(2) != "/blog/page/2/" {
		t.Errorf("expected URL(2) '/blog/page/2/', got '%s'", pager.URL(2))
	}
	if pager.URL(0) != "" {
		t.Errorf("expected URL(0) '', got '%s'", pager.URL(0))
	}
	if pager.URL(10) != "" {
		t.Errorf("expected URL(10) '', got '%s'", pager.URL(10))
	}
}

func TestPaginatorDefaultPerPage(t *testing.T) {
	pages := make([]*content.Page, 25)
	for i := range pages {
		pages[i] = &content.Page{Title: string(rune('A' + i))}
	}

	// Invalid perPage should default to 10
	pager := NewPaginator(pages, 0, "/blog/")
	if pager.PerPage != 10 {
		t.Errorf("expected default PerPage 10, got %d", pager.PerPage)
	}

	pager2 := NewPaginator(pages, -5, "/blog/")
	if pager2.PerPage != 10 {
		t.Errorf("expected default PerPage 10, got %d", pager2.PerPage)
	}
}

func TestPaginatorSinglePage(t *testing.T) {
	pages := make([]*content.Page, 5)
	for i := range pages {
		pages[i] = &content.Page{Title: string(rune('A' + i))}
	}

	pager := NewPaginator(pages, 10, "/blog/")

	if pager.TotalPages != 1 {
		t.Errorf("expected TotalPages 1, got %d", pager.TotalPages)
	}
	if pager.HasPrev {
		t.Error("expected HasPrev false for single page")
	}
	if pager.HasNext {
		t.Error("expected HasNext false for single page")
	}

	page1 := pager.Page(1)
	if len(page1.Items) != 5 {
		t.Errorf("expected 5 items, got %d", len(page1.Items))
	}
}
