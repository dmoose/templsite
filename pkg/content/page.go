// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package content

import (
	"slices"
	"time"
)

// Page represents a parsed content page with metadata
type Page struct {
	// Path is the original file path
	Path string

	// Content is the rendered HTML content
	Content string

	// RawContent is the original markdown (for custom processing)
	RawContent string

	// Frontmatter contains the parsed YAML frontmatter
	Frontmatter map[string]any

	// Layout specifies which layout template to use
	Layout string

	// URL is the generated URL path for this page
	URL string

	// Date is the publication date
	Date time.Time

	// Draft indicates if this is a draft post
	Draft bool

	// Title is extracted from frontmatter
	Title string

	// Description is extracted from frontmatter
	Description string

	// Tags are extracted from frontmatter
	Tags []string

	// Author is extracted from frontmatter
	Author string

	// Section is the content section (e.g., "blog" from content/blog/post.md)
	Section string

	// Weight is used for manual ordering (from frontmatter)
	Weight int

	// Summary is the page summary (first paragraph or content before <!--more-->)
	Summary string

	// WordCount is the number of words in the content
	WordCount int

	// ReadingTime is the estimated reading time in minutes
	ReadingTime int

	// TOC is the rendered table of contents HTML
	TOC string

	// Prev is the previous page in the section (by date)
	Prev *Page

	// Next is the next page in the section (by date)
	Next *Page

	// Aliases are alternative URLs that redirect to this page
	Aliases []string
}

// getStringDefault retrieves a string value from a map with a default fallback
func getStringDefault(m map[string]any, key, def string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return def
}

// getBoolDefault retrieves a bool value from a map with a default fallback
func getBoolDefault(m map[string]any, key string, def bool) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return def
}

// getIntDefault retrieves an int value from a map with a default fallback
func getIntDefault(m map[string]any, key string, def int) int {
	if v, ok := m[key].(int); ok {
		return v
	}
	// YAML may parse as float64
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return def
}

// getStringSlice retrieves a string slice from a map
func getStringSlice(m map[string]any, key string) []string {
	if v, ok := m[key].([]any); ok {
		result := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	if v, ok := m[key].([]string); ok {
		return v
	}
	return nil
}

// IsPublished returns true if the page is not a draft and has a past or current date
func (p *Page) IsPublished() bool {
	if p.Draft {
		return false
	}
	if p.Date.IsZero() {
		return true
	}
	return !p.Date.After(time.Now())
}

// HasTag returns true if the page has the specified tag
func (p *Page) HasTag(tag string) bool {
	return slices.Contains(p.Tags, tag)
}
