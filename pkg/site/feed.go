// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package site

import (
	"encoding/xml"
	"fmt"
	"time"
)

// atomFeed represents an Atom 1.0 feed
type atomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Xmlns   string      `xml:"xmlns,attr"`
	Title   string      `xml:"title"`
	Link    []atomLink  `xml:"link"`
	Updated string      `xml:"updated"`
	ID      string      `xml:"id"`
	Author  *atomAuthor `xml:"author,omitempty"`
	Entries []atomEntry `xml:"entry"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
}

type atomAuthor struct {
	Name string `xml:"name"`
}

type atomEntry struct {
	Title   string     `xml:"title"`
	Link    atomLink   `xml:"link"`
	ID      string     `xml:"id"`
	Updated string     `xml:"updated"`
	Summary string     `xml:"summary,omitempty"`
	Content *atomCDATA `xml:"content,omitempty"`
}

type atomCDATA struct {
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

// Feed generates an Atom 1.0 feed for the site.
//
// Only dated pages become entries: an undated page has no honest <updated>
// value, and substituting the clock would make two builds of unchanged sources
// differ. The feed-level <updated> is the newest entry date.
//
// If the site has no dated pages there is nothing to syndicate and Feed returns
// the empty string; callers should skip writing the file in that case.
func (s *Site) Feed() string {
	baseURL := s.Config.BaseURL

	var entries []atomEntry
	var newest time.Time

	for _, page := range s.RegularPages() {
		if page.Date.IsZero() {
			continue
		}

		entry := atomEntry{
			Title:   page.Title,
			Link:    atomLink{Href: baseURL + page.URL, Rel: "alternate", Type: "text/html"},
			ID:      baseURL + page.URL,
			Updated: page.Date.Format(time.RFC3339),
		}

		if page.Description != "" {
			entry.Summary = page.Description
		}

		if page.Content != "" {
			entry.Content = &atomCDATA{
				Type:    "html",
				Content: page.Content,
			}
		}

		if page.Date.After(newest) {
			newest = page.Date
		}

		entries = append(entries, entry)
	}

	if len(entries) == 0 {
		return ""
	}

	feed := atomFeed{
		Xmlns: "http://www.w3.org/2005/Atom",
		Title: s.Config.Title,
		Link: []atomLink{
			{Href: baseURL + "/", Rel: "alternate", Type: "text/html"},
			{Href: baseURL + "/feed.xml", Rel: "self", Type: "application/atom+xml"},
		},
		Updated: newest.Format(time.RFC3339),
		ID:      baseURL + "/",
		Entries: entries,
	}

	var buf []byte
	buf = append(buf, []byte(xml.Header)...)
	data, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return fmt.Sprintf("%s<feed/>", xml.Header)
	}
	buf = append(buf, data...)
	buf = append(buf, '\n')
	return string(buf)
}
