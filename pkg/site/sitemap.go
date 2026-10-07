// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package site

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"sort"

	"github.com/dmoose/templsite/pkg/content"
)

// URLSet represents a sitemap.xml document
type URLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []SitemapURL `xml:"url"`
}

// SitemapURL represents a single URL entry in a sitemap
type SitemapURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

// Sitemap generates a sitemap.xml for the site.
//
// A page whose frontmatter sets noindex: true is left out, and so is a
// section with no index page of its own whose every page is noindex
// (its URL would name nothing a search engine should find).
//
// lastmod comes from the page date, and is omitted for undated pages rather
// than filled in from the clock. Sections and taxonomy terms are emitted in
// sorted order so that repeated builds of unchanged content produce identical
// bytes; ranging over the underlying maps directly would not.
func (s *Site) Sitemap() string {
	var urls []SitemapURL

	// Add regular pages
	for _, page := range s.RegularPages() {
		if NoIndex(page) {
			continue
		}
		url := SitemapURL{
			Loc: s.Config.BaseURL + page.URL,
		}
		if !page.Date.IsZero() {
			url.LastMod = page.Date.Format("2006-01-02")
		}
		urls = append(urls, url)
	}

	// Add section index pages
	sectionNames := make([]string, 0, len(s.Sections))
	for name := range s.Sections {
		sectionNames = append(sectionNames, name)
	}
	sort.Strings(sectionNames)

	for _, name := range sectionNames {
		section := s.Sections[name]
		if sectionNoIndex(section) {
			continue
		}
		if section.Name == "_root" {
			// Add homepage
			urls = append(urls, SitemapURL{
				Loc: s.Config.BaseURL + "/",
			})
		} else {
			urls = append(urls, SitemapURL{
				Loc: s.Config.BaseURL + section.URL,
			})
		}
	}

	// Add taxonomy term pages
	taxNames := make([]string, 0, len(s.Taxonomies))
	for name := range s.Taxonomies {
		taxNames = append(taxNames, name)
	}
	sort.Strings(taxNames)

	for _, name := range taxNames {
		for _, term := range s.Taxonomies[name].TermsByName() {
			urls = append(urls, SitemapURL{
				Loc: s.Config.BaseURL + term.URL,
			})
		}
	}

	urlset := URLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")
	_ = encoder.Encode(urlset)

	return buf.String()
}

// RobotsTxt generates a robots.txt file with sitemap reference
func (s *Site) RobotsTxt() string {
	return fmt.Sprintf("User-agent: *\nAllow: /\n\nSitemap: %s/sitemap.xml\n", s.Config.BaseURL)
}

// RobotsTxtWithDisallow generates a robots.txt with custom disallow rules
func (s *Site) RobotsTxtWithDisallow(disallowPaths []string) string {
	var buf bytes.Buffer
	buf.WriteString("User-agent: *\n")

	for _, path := range disallowPaths {
		buf.WriteString("Disallow: ")
		buf.WriteString(path)
		buf.WriteString("\n")
	}

	if len(disallowPaths) == 0 {
		buf.WriteString("Allow: /\n")
	}

	buf.WriteString("\nSitemap: ")
	buf.WriteString(s.Config.BaseURL)
	buf.WriteString("/sitemap.xml\n")

	return buf.String()
}

// NoIndex reports whether a page's frontmatter sets noindex: true, which
// keeps it out of the sitemap. Layouts read it to emit the robots meta.
func NoIndex(p *content.Page) bool {
	v, _ := p.Frontmatter["noindex"].(bool)
	return v
}

// sectionNoIndex reports whether a section has nothing to index: its
// index page is noindex, or it has no index page and every page in it
// is noindex.
func sectionNoIndex(sec *Section) bool {
	if sec.Index != nil {
		return NoIndex(sec.Index)
	}
	if len(sec.Pages) == 0 {
		return false
	}
	for _, p := range sec.Pages {
		if !NoIndex(p) {
			return false
		}
	}
	return true
}
