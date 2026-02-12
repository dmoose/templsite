// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package site

import (
	"bytes"
	"encoding/xml"
	"time"

	"github.com/dmoose/templsite/pkg/content"
)

// RSSFeed represents an RSS 2.0 feed
type RSSFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Atom    string     `xml:"xmlns:atom,attr"`
	Channel RSSChannel `xml:"channel"`
}

// RSSChannel represents the channel element of an RSS feed
type RSSChannel struct {
	Title         string    `xml:"title"`
	Link          string    `xml:"link"`
	Description   string    `xml:"description"`
	Language      string    `xml:"language,omitempty"`
	LastBuildDate string    `xml:"lastBuildDate,omitempty"`
	AtomLink      AtomLink  `xml:"atom:link"`
	Items         []RSSItem `xml:"item"`
}

// AtomLink represents the atom:link element for RSS feed self-reference
type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

// RSSItem represents an item in an RSS feed
type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description,omitempty"`
	PubDate     string `xml:"pubDate,omitempty"`
	GUID        string `xml:"guid"`
	Author      string `xml:"author,omitempty"`
}

// RSS generates an RSS 2.0 feed for the given pages
func (s *Site) RSS(pages []*content.Page, title, description string) string {
	feedURL := s.Config.BaseURL + "/feed.xml"

	items := make([]RSSItem, 0, len(pages))
	for _, page := range pages {
		item := RSSItem{
			Title: page.Title,
			Link:  s.Config.BaseURL + page.URL,
			GUID:  s.Config.BaseURL + page.URL,
		}

		if page.Description != "" {
			item.Description = page.Description
		} else if page.Summary != "" {
			item.Description = page.Summary
		}

		if !page.Date.IsZero() {
			item.PubDate = page.Date.Format(time.RFC1123Z)
		}

		if page.Author != "" {
			item.Author = page.Author
		}

		items = append(items, item)
	}

	feed := RSSFeed{
		Version: "2.0",
		Atom:    "http://www.w3.org/2005/Atom",
		Channel: RSSChannel{
			Title:       title,
			Link:        s.Config.BaseURL,
			Description: description,
			AtomLink: AtomLink{
				Href: feedURL,
				Rel:  "self",
				Type: "application/rss+xml",
			},
			Items: items,
		},
	}

	// Add language if configured
	if s.Config.Language != "" {
		feed.Channel.Language = s.Config.Language
	}

	// Add last build date
	feed.Channel.LastBuildDate = time.Now().Format(time.RFC1123Z)

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")
	if err := encoder.Encode(feed); err != nil {
		return ""
	}

	return buf.String()
}

// AtomFeed represents an Atom 1.0 feed
type AtomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Xmlns   string      `xml:"xmlns,attr"`
	Title   string      `xml:"title"`
	Link    []AtomFLink `xml:"link"`
	Updated string      `xml:"updated"`
	ID      string      `xml:"id"`
	Author  *AtomAuthor `xml:"author,omitempty"`
	Entries []AtomEntry `xml:"entry"`
}

// AtomFLink represents a link in an Atom feed
type AtomFLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
}

// AtomAuthor represents an author in an Atom feed
type AtomAuthor struct {
	Name string `xml:"name"`
}

// AtomEntry represents an entry in an Atom feed
type AtomEntry struct {
	Title   string      `xml:"title"`
	Link    AtomFLink   `xml:"link"`
	ID      string      `xml:"id"`
	Updated string      `xml:"updated"`
	Summary string      `xml:"summary,omitempty"`
	Author  *AtomAuthor `xml:"author,omitempty"`
}

// Atom generates an Atom 1.0 feed for the given pages
func (s *Site) Atom(pages []*content.Page, title, subtitle string) string {
	feedURL := s.Config.BaseURL + "/atom.xml"

	entries := make([]AtomEntry, 0, len(pages))
	for _, page := range pages {
		entry := AtomEntry{
			Title: page.Title,
			Link: AtomFLink{
				Href: s.Config.BaseURL + page.URL,
				Rel:  "alternate",
			},
			ID: s.Config.BaseURL + page.URL,
		}

		if !page.Date.IsZero() {
			entry.Updated = page.Date.Format(time.RFC3339)
		} else {
			entry.Updated = time.Now().Format(time.RFC3339)
		}

		if page.Description != "" {
			entry.Summary = page.Description
		} else if page.Summary != "" {
			entry.Summary = page.Summary
		}

		if page.Author != "" {
			entry.Author = &AtomAuthor{Name: page.Author}
		}

		entries = append(entries, entry)
	}

	feed := AtomFeed{
		Xmlns: "http://www.w3.org/2005/Atom",
		Title: title,
		Link: []AtomFLink{
			{Href: s.Config.BaseURL, Rel: "alternate"},
			{Href: feedURL, Rel: "self", Type: "application/atom+xml"},
		},
		Updated: time.Now().Format(time.RFC3339),
		ID:      s.Config.BaseURL + "/",
		Entries: entries,
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")
	if err := encoder.Encode(feed); err != nil {
		return ""
	}

	return buf.String()
}

// JSONFeed represents a JSON Feed 1.1
type JSONFeed struct {
	Version     string         `json:"version"`
	Title       string         `json:"title"`
	HomePageURL string         `json:"home_page_url"`
	FeedURL     string         `json:"feed_url"`
	Description string         `json:"description,omitempty"`
	Language    string         `json:"language,omitempty"`
	Items       []JSONFeedItem `json:"items"`
}

// JSONFeedItem represents an item in a JSON Feed
type JSONFeedItem struct {
	ID            string `json:"id"`
	URL           string `json:"url"`
	Title         string `json:"title"`
	ContentText   string `json:"content_text,omitempty"`
	Summary       string `json:"summary,omitempty"`
	DatePublished string `json:"date_published,omitempty"`
	Author        *struct {
		Name string `json:"name"`
	} `json:"author,omitempty"`
}

// JSON generates a JSON Feed 1.1 for the given pages
func (s *Site) JSON(pages []*content.Page, title, description string) string {
	feedURL := s.Config.BaseURL + "/feed.json"

	items := make([]JSONFeedItem, 0, len(pages))
	for _, page := range pages {
		item := JSONFeedItem{
			ID:    s.Config.BaseURL + page.URL,
			URL:   s.Config.BaseURL + page.URL,
			Title: page.Title,
		}

		if page.Description != "" {
			item.Summary = page.Description
		} else if page.Summary != "" {
			item.Summary = page.Summary
		}

		if !page.Date.IsZero() {
			item.DatePublished = page.Date.Format(time.RFC3339)
		}

		if page.Author != "" {
			item.Author = &struct {
				Name string `json:"name"`
			}{Name: page.Author}
		}

		items = append(items, item)
	}

	feed := JSONFeed{
		Version:     "https://jsonfeed.org/version/1.1",
		Title:       title,
		HomePageURL: s.Config.BaseURL,
		FeedURL:     feedURL,
		Description: description,
		Items:       items,
	}

	if s.Config.Language != "" {
		feed.Language = s.Config.Language
	}

	// Use encoding/json for proper JSON output
	var buf bytes.Buffer
	encoder := jsonEncoder{&buf}
	if err := encoder.Encode(feed); err != nil {
		return ""
	}

	return buf.String()
}

// jsonEncoder is a simple JSON encoder that produces readable output
type jsonEncoder struct {
	buf *bytes.Buffer
}

func (e jsonEncoder) Encode(v any) error {
	data, err := marshalJSONIndent(v)
	if err != nil {
		return err
	}
	e.buf.Write(data)
	return nil
}

// marshalJSONIndent marshals v to indented JSON
func marshalJSONIndent(v any) ([]byte, error) {
	// Import encoding/json inline to avoid adding it at package level
	// when it might not be needed
	import_json := func() ([]byte, error) {
		var buf bytes.Buffer
		encoder := struct {
			*bytes.Buffer
		}{&buf}

		// Manual JSON encoding for JSONFeed
		feed, ok := v.(JSONFeed)
		if !ok {
			return nil, nil
		}

		buf.WriteString("{\n")
		buf.WriteString(`  "version": "` + feed.Version + "\",\n")
		buf.WriteString(`  "title": "` + escapeJSON(feed.Title) + "\",\n")
		buf.WriteString(`  "home_page_url": "` + feed.HomePageURL + "\",\n")
		buf.WriteString(`  "feed_url": "` + feed.FeedURL + "\"")

		if feed.Description != "" {
			buf.WriteString(",\n")
			buf.WriteString(`  "description": "` + escapeJSON(feed.Description) + "\"")
		}

		if feed.Language != "" {
			buf.WriteString(",\n")
			buf.WriteString(`  "language": "` + feed.Language + "\"")
		}

		buf.WriteString(",\n  \"items\": [")
		for i, item := range feed.Items {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString("\n    {\n")
			buf.WriteString(`      "id": "` + item.ID + "\",\n")
			buf.WriteString(`      "url": "` + item.URL + "\",\n")
			buf.WriteString(`      "title": "` + escapeJSON(item.Title) + "\"")

			if item.Summary != "" {
				buf.WriteString(",\n")
				buf.WriteString(`      "summary": "` + escapeJSON(item.Summary) + "\"")
			}

			if item.DatePublished != "" {
				buf.WriteString(",\n")
				buf.WriteString(`      "date_published": "` + item.DatePublished + "\"")
			}

			if item.Author != nil {
				buf.WriteString(",\n")
				buf.WriteString(`      "author": {"name": "` + escapeJSON(item.Author.Name) + `"}`)
			}

			buf.WriteString("\n    }")
		}
		buf.WriteString("\n  ]\n}")

		return encoder.Bytes(), nil
	}

	return import_json()
}

// escapeJSON escapes special characters in JSON strings
func escapeJSON(s string) string {
	var buf bytes.Buffer
	for _, r := range s {
		switch r {
		case '"':
			buf.WriteString(`\"`)
		case '\\':
			buf.WriteString(`\\`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		default:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}
