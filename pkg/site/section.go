package site

import (
	"sort"

	"git.catapulsion.com/templsite/pkg/content"
)

// Section represents a content section (e.g., blog, docs)
type Section struct {
	// Name is the section identifier (directory name)
	Name string

	// Title is the display title (from _index.md or derived from Name)
	Title string

	// Description is from _index.md frontmatter
	Description string

	// Pages contains all pages in this section, sorted by date descending
	Pages []*content.Page

	// URL is the section's URL path (e.g., /blog/)
	URL string

	// Index is the _index.md page for this section (if it exists)
	Index *content.Page
}

// PageCount returns the number of regular pages (excluding _index.md)
func (s *Section) PageCount() int {
	count := 0
	for _, p := range s.Pages {
		if !s.isIndexPage(p) {
			count++
		}
	}
	return count
}

// RegularPages returns pages excluding _index.md
func (s *Section) RegularPages() []*content.Page {
	var pages []*content.Page
	for _, p := range s.Pages {
		if !s.isIndexPage(p) {
			pages = append(pages, p)
		}
	}
	return pages
}

// isIndexPage checks if a page is an _index.md file
func (s *Section) isIndexPage(p *content.Page) bool {
	// Check if the page's URL matches the section URL (it's the index)
	return p.URL == s.URL
}

// sortPages sorts section pages by date descending, then weight, then title
func (s *Section) sortPages() {
	sort.Slice(s.Pages, func(i, j int) bool {
		// First compare by date (descending)
		if !s.Pages[i].Date.Equal(s.Pages[j].Date) {
			return s.Pages[i].Date.After(s.Pages[j].Date)
		}
		// Then by weight (ascending - lower weight first)
		if s.Pages[i].Weight != s.Pages[j].Weight {
			return s.Pages[i].Weight < s.Pages[j].Weight
		}
		// Finally by title
		return s.Pages[i].Title < s.Pages[j].Title
	})
}

// linkPrevNext establishes Prev/Next links for regular pages in this section
func (s *Section) linkPrevNext() {
	regularPages := s.RegularPages()

	for i, page := range regularPages {
		if i > 0 {
			page.Prev = regularPages[i-1]
		} else {
			page.Prev = nil
		}
		if i < len(regularPages)-1 {
			page.Next = regularPages[i+1]
		} else {
			page.Next = nil
		}
	}
}
