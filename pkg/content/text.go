package content

import (
	"regexp"
	"strings"
)

// moreMarker is the separator for manual summary breaks
const moreMarker = "<!--more-->"

// stripHTMLRegex matches HTML tags
var stripHTMLRegex = regexp.MustCompile(`<[^>]*>`)

// whitespaceRegex matches multiple whitespace characters
var whitespaceRegex = regexp.MustCompile(`\s+`)

// footnoteRefRegex matches goldmark footnote <sup> elements and their contents.
// These reference anchors (#fn:N) that only exist in the full article, so they
// must be stripped from summaries and descriptions.
var footnoteRefRegex = regexp.MustCompile(`<sup[^>]*>.*?</sup>`)

// closeTagRegex matches opening and closing HTML tags
var closeTagRegex = regexp.MustCompile(`<(/?)(\w+)[^>]*>`)

// StripHTML removes HTML tags from a string
func StripHTML(html string) string {
	// Remove HTML tags
	text := stripHTMLRegex.ReplaceAllString(html, " ")
	// Normalize whitespace
	text = whitespaceRegex.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

// WordCount returns the number of words in plain text
func WordCount(text string) int {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}
	return len(strings.Fields(text))
}

// ReadingTime estimates reading time in minutes (assumes 200 words per minute)
func ReadingTime(wordCount int) int {
	if wordCount <= 0 {
		return 0
	}
	// Round up: (wordCount + 199) / 200
	return (wordCount + 199) / 200
}

// ExtractSummary extracts a plain-text summary from markdown and rendered HTML.
// Priority: 1) Content before <!--more-->, 2) First paragraph, 3) Truncated content.
// The result is always plain text — HTML tags and footnote references are stripped
// so the summary is safe for meta descriptions, RSS feeds, and templ text nodes.
func ExtractSummary(markdown, html string, maxWords int) string {
	// Strip footnote refs — they reference anchors that don't exist in summaries
	cleaned := footnoteRefRegex.ReplaceAllString(html, "")

	var raw string

	// Check for <!--more--> marker in markdown
	if idx := strings.Index(markdown, moreMarker); idx != -1 {
		raw = extractBeforeMore(cleaned, markdown[:idx])
	} else if first := extractFirstParagraph(cleaned); first != "" {
		raw = first
	} else {
		raw = truncateHTML(cleaned, maxWords)
	}

	// Convert to plain text for use in meta tags, listings, and feeds
	return StripHTML(raw)
}

// extractBeforeMore extracts HTML content corresponding to markdown before <!--more-->
func extractBeforeMore(html, markdownBefore string) string {
	// Count paragraphs in the markdown before marker
	paragraphs := countParagraphs(markdownBefore)
	if paragraphs == 0 {
		paragraphs = 1
	}

	// Extract that many paragraphs from HTML
	return extractParagraphs(html, paragraphs)
}

// countParagraphs counts paragraph breaks in markdown
func countParagraphs(markdown string) int {
	// Split by double newlines (paragraph breaks)
	parts := strings.Split(markdown, "\n\n")
	count := 0
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			count++
		}
	}
	return count
}

// extractParagraphs extracts the first n paragraphs from HTML
func extractParagraphs(html string, n int) string {
	var result strings.Builder
	count := 0
	tagStart := -1

	for i := 0; i < len(html) && count < n; i++ {
		if html[i] == '<' {
			tagStart = i
		} else if html[i] == '>' && tagStart >= 0 {
			tag := strings.ToLower(html[tagStart : i+1])
			if tag == "</p>" {
				count++
			}
			tagStart = -1
		}

		if count < n {
			result.WriteByte(html[i])
		}
	}

	return strings.TrimSpace(result.String())
}

// extractFirstParagraph extracts the first <p>...</p> from HTML
func extractFirstParagraph(html string) string {
	lower := strings.ToLower(html)

	// Find opening <p> tag (may have attributes)
	pStart := strings.Index(lower, "<p")
	if pStart == -1 {
		return ""
	}

	// Find the end of the opening tag
	tagEnd := strings.Index(html[pStart:], ">")
	if tagEnd == -1 {
		return ""
	}
	contentStart := pStart + tagEnd + 1

	// Find closing </p>
	pEnd := strings.Index(lower[contentStart:], "</p>")
	if pEnd == -1 {
		return ""
	}

	// Return just the inner content (without tags)
	return strings.TrimSpace(html[contentStart : contentStart+pEnd])
}

// truncateHTML truncates HTML content to approximately maxWords words
// It tries to preserve complete sentences and HTML structure
func truncateHTML(html string, maxWords int) string {
	if maxWords <= 0 {
		maxWords = 50
	}

	// Strip HTML to count words
	plain := StripHTML(html)
	words := strings.Fields(plain)

	if len(words) <= maxWords {
		return html
	}

	// Find the position in the original HTML that corresponds to maxWords
	// This is approximate - we find the Nth word and cut there
	targetText := strings.Join(words[:maxWords], " ")

	// Find position of the end of our target in the plain text
	targetEnd := len(targetText)

	// Map back to HTML position (approximate)
	htmlPos := mapPlainPosToHTML(html, targetEnd)

	// Find a good breaking point (end of sentence or tag)
	breakPos := findBreakPoint(html, htmlPos)

	result := html[:breakPos]

	// Close any open tags
	result = closeOpenTags(result)

	// Add ellipsis if we truncated
	if breakPos < len(html) {
		result = strings.TrimRight(result, " ")
		if !strings.HasSuffix(result, ".") && !strings.HasSuffix(result, "?") && !strings.HasSuffix(result, "!") {
			result += "…"
		}
	}

	return result
}

// mapPlainPosToHTML maps a position in plain text to approximate position in HTML
func mapPlainPosToHTML(html string, plainPos int) int {
	plainCount := 0
	inTag := false

	for i := 0; i < len(html); i++ {
		if html[i] == '<' {
			inTag = true
		} else if html[i] == '>' {
			inTag = false
		} else if !inTag {
			plainCount++
			if plainCount >= plainPos {
				return i + 1
			}
		}
	}
	return len(html)
}

// findBreakPoint finds a good position to break the HTML
func findBreakPoint(html string, pos int) int {
	if pos >= len(html) {
		return len(html)
	}

	// Look for sentence end within next 50 chars
	searchEnd := pos + 50
	if searchEnd > len(html) {
		searchEnd = len(html)
	}

	for i := pos; i < searchEnd; i++ {
		c := html[i]
		if c == '.' || c == '?' || c == '!' {
			// Skip if inside a tag
			if !isInsideTag(html, i) {
				return i + 1
			}
		}
	}

	// Look for word boundary
	for i := pos; i < searchEnd; i++ {
		if html[i] == ' ' && !isInsideTag(html, i) {
			return i
		}
	}

	return pos
}

// isInsideTag checks if position is inside an HTML tag
func isInsideTag(html string, pos int) bool {
	lastOpen := strings.LastIndex(html[:pos], "<")
	lastClose := strings.LastIndex(html[:pos], ">")
	return lastOpen > lastClose
}

// closeOpenTags closes any unclosed HTML tags
func closeOpenTags(html string) string {
	var openTags []string
	matches := closeTagRegex.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		isClosing := match[1] == "/"
		tagName := strings.ToLower(match[2])

		// Skip self-closing tags
		if tagName == "br" || tagName == "hr" || tagName == "img" || tagName == "input" {
			continue
		}

		if isClosing {
			// Remove from open tags
			for i := len(openTags) - 1; i >= 0; i-- {
				if openTags[i] == tagName {
					openTags = append(openTags[:i], openTags[i+1:]...)
					break
				}
			}
		} else {
			openTags = append(openTags, tagName)
		}
	}

	// Close open tags in reverse order
	var result strings.Builder
	result.WriteString(html)
	for i := len(openTags) - 1; i >= 0; i-- {
		result.WriteString("</")
		result.WriteString(openTags[i])
		result.WriteString(">")
	}

	return result.String()
}
