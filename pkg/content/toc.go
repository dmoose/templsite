// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package content

import (
	"fmt"
	"regexp"
	"strings"
)

// TOCEntry represents a heading in the table of contents
type TOCEntry struct {
	Level    int
	ID       string
	Text     string
	Children []*TOCEntry
}

// headingRegex matches HTML heading tags with their content
var headingRegex = regexp.MustCompile(`<h([1-6])(?:\s+id="([^"]*)")?[^>]*>(.*?)</h[1-6]>`)

// GenerateTOC extracts headings from HTML and generates a table of contents.
// minLevel and maxLevel control which heading levels to include (default: 2-4).
// The optional tocID parameter scopes the TOC for go-components JS scroll-spy;
// when provided, each link gets data-tui-toc-id for IntersectionObserver binding.
func GenerateTOC(html string, minLevel, maxLevel int, tocID ...string) string {
	if minLevel <= 0 {
		minLevel = 2
	}
	if maxLevel <= 0 {
		maxLevel = 4
	}
	if maxLevel < minLevel {
		maxLevel = minLevel
	}

	entries := extractHeadings(html, minLevel, maxLevel)
	if len(entries) == 0 {
		return ""
	}

	id := ""
	if len(tocID) > 0 {
		id = tocID[0]
	}
	return renderTOC(entries, id)
}

// extractHeadings finds all headings in HTML within the level range
func extractHeadings(html string, minLevel, maxLevel int) []*TOCEntry {
	var entries []*TOCEntry

	matches := headingRegex.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		level := int(match[1][0] - '0')
		if level < minLevel || level > maxLevel {
			continue
		}

		id := match[2]
		text := StripHTML(match[3])

		// Generate ID if not present
		if id == "" {
			id = generateHeadingID(text)
		}

		entries = append(entries, &TOCEntry{
			Level: level,
			ID:    id,
			Text:  text,
		})
	}

	return entries
}

// generateHeadingID creates a URL-safe ID from heading text
func generateHeadingID(text string) string {
	// Convert to lowercase
	id := strings.ToLower(text)

	// Replace spaces with hyphens
	id = strings.ReplaceAll(id, " ", "-")

	// Remove non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	// Remove consecutive hyphens
	id = result.String()
	for strings.Contains(id, "--") {
		id = strings.ReplaceAll(id, "--", "-")
	}

	// Trim hyphens from ends
	id = strings.Trim(id, "-")

	if id == "" {
		id = "heading"
	}

	return id
}

// renderTOC renders entries as nested HTML lists.
// tocID scopes data-tui-toc-id attributes for JS scroll-spy binding.
func renderTOC(entries []*TOCEntry, tocID string) string {
	if len(entries) == 0 {
		return ""
	}

	var buf strings.Builder
	renderTOCLevel(&buf, entries, entries[0].Level, tocID)
	return buf.String()
}

// renderTOCLevel recursively renders TOC entries at a given level.
// Output uses CSS classes and data attributes compatible with go-components:
// toc-list on <ul>, toc-item/toc-item-h{N} on <li>, toc-link on <a>,
// and data-tui-toc-link/data-tui-toc-href/data-tui-toc-id for JS scroll-spy.
func renderTOCLevel(buf *strings.Builder, entries []*TOCEntry, baseLevel int, tocID string) {
	if len(entries) == 0 {
		return
	}

	buf.WriteString(`<ul class="toc-list">` + "\n")

	for i := 0; i < len(entries); {
		entry := entries[i]

		fmt.Fprintf(buf, `<li class="toc-item toc-item-h%d">`, entry.Level)
		if tocID != "" {
			fmt.Fprintf(buf, `<a class="toc-link" href="#%s" data-tui-toc-link data-tui-toc-href="#%s" data-tui-toc-id="%s">%s</a>`,
				entry.ID, entry.ID, tocID, entry.Text)
		} else {
			fmt.Fprintf(buf, `<a class="toc-link" href="#%s">%s</a>`, entry.ID, entry.Text)
		}

		// Find children (entries with higher level numbers until we hit same or lower level)
		children := []*TOCEntry{}
		j := i + 1
		for j < len(entries) && entries[j].Level > entry.Level {
			children = append(children, entries[j])
			j++
		}

		if len(children) > 0 {
			buf.WriteString("\n")
			renderTOCLevel(buf, children, entry.Level+1, tocID)
		}

		buf.WriteString("</li>\n")
		i = j
	}

	buf.WriteString("</ul>\n")
}

// TOCFromEntries builds TOC HTML from a slice of entries (alternative API).
// The optional tocID parameter enables go-components JS scroll-spy binding.
func TOCFromEntries(entries []*TOCEntry, tocID ...string) string {
	id := ""
	if len(tocID) > 0 {
		id = tocID[0]
	}
	return renderTOC(entries, id)
}
