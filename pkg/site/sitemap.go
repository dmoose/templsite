package site

import (
	"bytes"
	"encoding/xml"
	"fmt"
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

// Sitemap generates a sitemap.xml for the site
func (s *Site) Sitemap() string {
	var urls []SitemapURL

	// Add regular pages
	for _, page := range s.RegularPages() {
		url := SitemapURL{
			Loc: s.Config.BaseURL + page.URL,
		}
		if !page.Date.IsZero() {
			url.LastMod = page.Date.Format("2006-01-02")
		}
		urls = append(urls, url)
	}

	// Add section index pages
	for _, section := range s.Sections {
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
	for _, tax := range s.Taxonomies {
		for _, term := range tax.Terms {
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
