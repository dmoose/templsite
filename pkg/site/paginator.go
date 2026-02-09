package site

import (
	"fmt"
	"strings"

	"github.com/dmoose/templsite/pkg/content"
)

// Paginator handles pagination of page collections with navigation URLs
type Paginator struct {
	Items      []*content.Page // pages for current page
	TotalItems int             // total number of items across all pages
	PerPage    int             // items per page
	TotalPages int             // total number of pages
	PageNum    int             // current page number (1-indexed)
	HasPrev    bool            // true if there's a previous page
	HasNext    bool            // true if there's a next page
	PrevURL    string          // URL of previous page (empty if no prev)
	NextURL    string          // URL of next page (empty if no next)
	PageURLs   []string        // URLs for all pages (for building navigation)
	BaseURL    string          // base URL for this paginated section
}

// NewPaginator creates a paginator for a page collection
// baseURL should end with a slash (e.g., "/blog/")
func NewPaginator(pages []*content.Page, perPage int, baseURL string) *Paginator {
	if perPage < 1 {
		perPage = 10
	}

	// Ensure baseURL ends with slash
	if !strings.HasSuffix(baseURL, "/") {
		baseURL = baseURL + "/"
	}

	total := len(pages)
	totalPages := max((total+perPage-1)/perPage, 1)

	// Generate all page URLs
	urls := make([]string, totalPages)
	for i := range totalPages {
		if i == 0 {
			urls[i] = baseURL
		} else {
			urls[i] = fmt.Sprintf("%spage/%d/", baseURL, i+1)
		}
	}

	return &Paginator{
		Items:      pages,
		TotalItems: total,
		PerPage:    perPage,
		TotalPages: totalPages,
		PageNum:    1,
		HasPrev:    false,
		HasNext:    totalPages > 1,
		PrevURL:    "",
		NextURL:    safeIndex(urls, 1),
		PageURLs:   urls,
		BaseURL:    baseURL,
	}
}

// Page returns a new paginator configured for a specific page number (1-indexed)
// The returned paginator has Items set to just the items for that page
func (p *Paginator) Page(num int) *Paginator {
	if num < 1 {
		num = 1
	}
	if num > p.TotalPages {
		num = p.TotalPages
	}

	start := (num - 1) * p.PerPage
	end := min(start+p.PerPage, len(p.Items))

	// Handle empty items case
	var items []*content.Page
	if start < len(p.Items) {
		items = p.Items[start:end]
	}

	return &Paginator{
		Items:      items,
		TotalItems: p.TotalItems,
		PerPage:    p.PerPage,
		TotalPages: p.TotalPages,
		PageNum:    num,
		HasPrev:    num > 1,
		HasNext:    num < p.TotalPages,
		PrevURL:    safeIndex(p.PageURLs, num-2),
		NextURL:    safeIndex(p.PageURLs, num),
		PageURLs:   p.PageURLs,
		BaseURL:    p.BaseURL,
	}
}

// First returns the first page number (always 1)
func (p *Paginator) First() int {
	return 1
}

// Last returns the last page number
func (p *Paginator) Last() int {
	return p.TotalPages
}

// Pages returns a slice of page numbers for building navigation
// For example, if on page 5 of 10, might return [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
func (p *Paginator) Pages() []int {
	pages := make([]int, p.TotalPages)
	for i := range p.TotalPages {
		pages[i] = i + 1
	}
	return pages
}

// URL returns the URL for a specific page number
func (p *Paginator) URL(pageNum int) string {
	if pageNum < 1 || pageNum > p.TotalPages {
		return ""
	}
	return p.PageURLs[pageNum-1]
}

// safeIndex returns the element at index i, or empty string if out of bounds
func safeIndex(slice []string, i int) string {
	if i < 0 || i >= len(slice) {
		return ""
	}
	return slice[i]
}
