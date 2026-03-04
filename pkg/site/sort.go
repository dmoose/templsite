package site

import (
	"sort"

	"git.catapulsion.com/templsite/pkg/content"
)

// SortByDate returns a copy of pages sorted by date
// If ascending is true, oldest first; otherwise newest first
func SortByDate(pages []*content.Page, ascending bool) []*content.Page {
	result := make([]*content.Page, len(pages))
	copy(result, pages)

	sort.Slice(result, func(i, j int) bool {
		if ascending {
			return result[i].Date.Before(result[j].Date)
		}
		return result[i].Date.After(result[j].Date)
	})

	return result
}

// SortByTitle returns a copy of pages sorted alphabetically by title
func SortByTitle(pages []*content.Page) []*content.Page {
	result := make([]*content.Page, len(pages))
	copy(result, pages)

	sort.Slice(result, func(i, j int) bool {
		return result[i].Title < result[j].Title
	})

	return result
}

// SortByWeight returns a copy of pages sorted by weight (ascending), then date (descending)
func SortByWeight(pages []*content.Page) []*content.Page {
	result := make([]*content.Page, len(pages))
	copy(result, pages)

	sort.Slice(result, func(i, j int) bool {
		if result[i].Weight != result[j].Weight {
			return result[i].Weight < result[j].Weight
		}
		return result[i].Date.After(result[j].Date)
	})

	return result
}

// Filter returns pages matching a predicate function
func Filter(pages []*content.Page, predicate func(*content.Page) bool) []*content.Page {
	var result []*content.Page
	for _, p := range pages {
		if predicate(p) {
			result = append(result, p)
		}
	}
	return result
}

// Limit returns the first n pages from the slice
// If n >= len(pages), returns the original slice
func Limit(pages []*content.Page, n int) []*content.Page {
	if n <= 0 {
		return nil
	}
	if n >= len(pages) {
		return pages
	}
	return pages[:n]
}

// Offset returns pages starting from index n
// If n >= len(pages), returns nil
func Offset(pages []*content.Page, n int) []*content.Page {
	if n <= 0 {
		return pages
	}
	if n >= len(pages) {
		return nil
	}
	return pages[n:]
}

// Paginate returns a slice of pages for a given page number (1-indexed)
func Paginate(pages []*content.Page, pageNum, perPage int) []*content.Page {
	if pageNum < 1 {
		pageNum = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	start := (pageNum - 1) * perPage
	if start >= len(pages) {
		return nil
	}

	end := start + perPage
	if end > len(pages) {
		end = len(pages)
	}

	return pages[start:end]
}

// First returns the first page, or nil if empty
func First(pages []*content.Page) *content.Page {
	if len(pages) == 0 {
		return nil
	}
	return pages[0]
}

// Last returns the last page, or nil if empty
func Last(pages []*content.Page) *content.Page {
	if len(pages) == 0 {
		return nil
	}
	return pages[len(pages)-1]
}

// Reverse returns a copy of pages in reverse order
func Reverse(pages []*content.Page) []*content.Page {
	result := make([]*content.Page, len(pages))
	for i, p := range pages {
		result[len(pages)-1-i] = p
	}
	return result
}

// GroupBySection groups pages by their section
func GroupBySection(pages []*content.Page) map[string][]*content.Page {
	groups := make(map[string][]*content.Page)
	for _, p := range pages {
		section := p.Section
		if section == "" {
			section = "_root"
		}
		groups[section] = append(groups[section], p)
	}
	return groups
}
