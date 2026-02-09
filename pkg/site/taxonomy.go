package site

import (
	"sort"
	"strings"

	"github.com/dmoose/templsite/pkg/content"
)

// Taxonomy represents a classification system (e.g., tags, categories)
type Taxonomy struct {
	// Name is the taxonomy identifier (e.g., "tags")
	Name string

	// Plural is used for URLs (usually same as Name)
	Plural string

	// Terms maps term slugs to Term structs
	Terms map[string]*Term
}

// Term represents a single term within a taxonomy (e.g., "go" in tags)
type Term struct {
	// Name is the display name of the term
	Name string

	// Slug is the URL-safe version of the name
	Slug string

	// Pages contains all pages with this term
	Pages []*content.Page

	// URL is the term's URL path (e.g., /tags/go/)
	URL string

	// Taxonomy is the parent taxonomy name
	Taxonomy string
}

// PageCount returns the number of pages with this term
func (t *Term) PageCount() int {
	return len(t.Pages)
}

// NewTaxonomy creates a new taxonomy with the given name
func NewTaxonomy(name string) *Taxonomy {
	return &Taxonomy{
		Name:   name,
		Plural: name, // Default plural is same as name
		Terms:  make(map[string]*Term),
	}
}

// AddPage adds a page to a term, creating the term if it doesn't exist
func (t *Taxonomy) AddPage(termName string, page *content.Page) {
	slug := slugify(termName)

	term, exists := t.Terms[slug]
	if !exists {
		term = &Term{
			Name:     termName,
			Slug:     slug,
			URL:      "/" + t.Plural + "/" + slug + "/",
			Taxonomy: t.Name,
			Pages:    make([]*content.Page, 0),
		}
		t.Terms[slug] = term
	}

	term.Pages = append(term.Pages, page)
}

// GetTerm returns a term by slug, or nil if not found
func (t *Taxonomy) GetTerm(slug string) *Term {
	return t.Terms[slug]
}

// AllTerms returns all terms sorted by page count (descending)
func (t *Taxonomy) AllTerms() []*Term {
	terms := make([]*Term, 0, len(t.Terms))
	for _, term := range t.Terms {
		terms = append(terms, term)
	}

	sort.Slice(terms, func(i, j int) bool {
		// Sort by page count descending
		if len(terms[i].Pages) != len(terms[j].Pages) {
			return len(terms[i].Pages) > len(terms[j].Pages)
		}
		// Then alphabetically by name
		return terms[i].Name < terms[j].Name
	})

	return terms
}

// TermsByName returns all terms sorted alphabetically by name
func (t *Taxonomy) TermsByName() []*Term {
	terms := make([]*Term, 0, len(t.Terms))
	for _, term := range t.Terms {
		terms = append(terms, term)
	}

	sort.Slice(terms, func(i, j int) bool {
		return terms[i].Name < terms[j].Name
	})

	return terms
}

// slugify converts a term name to a URL-safe slug
func slugify(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	// Remove consecutive hyphens
	slug = result.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Trim hyphens from ends
	slug = strings.Trim(slug, "-")

	if slug == "" {
		slug = "term"
	}

	return slug
}

// SortTermPages sorts pages within each term by date descending
func (t *Taxonomy) SortTermPages() {
	for _, term := range t.Terms {
		sort.Slice(term.Pages, func(i, j int) bool {
			return term.Pages[i].Date.After(term.Pages[j].Date)
		})
	}
}
