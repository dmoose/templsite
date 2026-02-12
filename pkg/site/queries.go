// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package site

import (
	"strings"

	"github.com/dmoose/templsite/pkg/content"
)

// RegularPages returns all non-index, published pages across all sections
func (s *Site) RegularPages() []*content.Page {
	var pages []*content.Page
	for _, page := range s.Pages {
		if page.IsPublished() && !s.isIndexPage(page) {
			pages = append(pages, page)
		}
	}
	return pages
}

// PagesInSection returns all pages in a specific section
// Returns nil if section doesn't exist
func (s *Site) PagesInSection(sectionName string) []*content.Page {
	if sectionName == "" {
		sectionName = "_root"
	}
	if section, ok := s.Sections[sectionName]; ok {
		return section.Pages
	}
	return nil
}

// RegularPagesInSection returns non-index, published pages in a section
func (s *Site) RegularPagesInSection(sectionName string) []*content.Page {
	if sectionName == "" {
		sectionName = "_root"
	}
	section, ok := s.Sections[sectionName]
	if !ok {
		return nil
	}

	var pages []*content.Page
	for _, page := range section.RegularPages() {
		if page.IsPublished() {
			pages = append(pages, page)
		}
	}
	return pages
}

// GetSection returns a section by name
func (s *Site) GetSection(name string) *Section {
	if name == "" {
		name = "_root"
	}
	return s.Sections[name]
}

// AllSections returns all sections (excluding _root if empty)
func (s *Site) AllSections() []*Section {
	var sections []*Section
	for _, section := range s.Sections {
		// Include _root only if it has regular pages
		if section.Name == "_root" && section.PageCount() == 0 {
			continue
		}
		sections = append(sections, section)
	}
	return sections
}

// SectionNames returns names of all sections with content
func (s *Site) SectionNames() []string {
	var names []string
	for name, section := range s.Sections {
		if name == "_root" && section.PageCount() == 0 {
			continue
		}
		names = append(names, name)
	}
	return names
}

// PageByURL finds a page by its URL path
func (s *Site) PageByURL(url string) *content.Page {
	// Normalize URL
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	if !strings.HasSuffix(url, "/") && url != "/" {
		url += "/"
	}

	for _, page := range s.Pages {
		if page.URL == url {
			return page
		}
	}
	return nil
}

// PagesByTag returns all published pages with a specific tag
// Uses the taxonomy system if available, falls back to direct tag check
func (s *Site) PagesByTag(tag string) []*content.Page {
	// Try to use taxonomy system first
	if pages := s.PagesByTaxonomy("tags", tag); pages != nil {
		return pages
	}

	// Fallback to direct check (for when taxonomies aren't built)
	var pages []*content.Page
	for _, page := range s.Pages {
		if page.IsPublished() && page.HasTag(tag) {
			pages = append(pages, page)
		}
	}
	return pages
}

// AllTags returns all unique tags across all pages with their counts
func (s *Site) AllTags() map[string]int {
	tags := make(map[string]int)
	for _, page := range s.Pages {
		if !page.IsPublished() {
			continue
		}
		for _, tag := range page.Tags {
			tags[tag]++
		}
	}
	return tags
}

// GetTaxonomy returns a taxonomy by name, or nil if not found
func (s *Site) GetTaxonomy(name string) *Taxonomy {
	if s.Taxonomies == nil {
		return nil
	}
	return s.Taxonomies[name]
}

// TaxonomyTerms returns all terms for a taxonomy, sorted by page count descending
// Returns nil if taxonomy doesn't exist
func (s *Site) TaxonomyTerms(name string) []*Term {
	tax := s.GetTaxonomy(name)
	if tax == nil {
		return nil
	}
	return tax.AllTerms()
}

// TaxonomyTermsByName returns all terms for a taxonomy, sorted alphabetically
func (s *Site) TaxonomyTermsByName(name string) []*Term {
	tax := s.GetTaxonomy(name)
	if tax == nil {
		return nil
	}
	return tax.TermsByName()
}

// PagesByTaxonomy returns pages for a specific taxonomy term
// Returns nil if taxonomy or term doesn't exist
func (s *Site) PagesByTaxonomy(taxonomy, termSlug string) []*content.Page {
	tax := s.GetTaxonomy(taxonomy)
	if tax == nil {
		return nil
	}

	term := tax.GetTerm(termSlug)
	if term == nil {
		return nil
	}

	return term.Pages
}

// GetTerm returns a specific term from a taxonomy
func (s *Site) GetTerm(taxonomy, termSlug string) *Term {
	tax := s.GetTaxonomy(taxonomy)
	if tax == nil {
		return nil
	}
	return tax.GetTerm(termSlug)
}

// Tags returns all tags as Term structs, sorted by page count descending
// This is a convenience method equivalent to TaxonomyTerms("tags")
func (s *Site) Tags() []*Term {
	return s.TaxonomyTerms("tags")
}
