// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package content

import (
	"regexp"
	"strings"
)

// sectionHeadingRegex matches top-level heading tags (h2 by default).
// This is used to find split points for section wrapping.
var sectionHeadingRegex = regexp.MustCompile(`(?i)<h2[\s>]`)

// WrapHeadingSections wraps each h2-delimited block of HTML in <section> tags.
// Content before the first h2 is wrapped in its own section.
// This enables CSS structural selectors like section:nth-of-type(even)
// for alternating backgrounds via the token system.
//
// Input:  <h2 id="a">A</h2><p>...</p><h2 id="b">B</h2><p>...</p>
// Output: <section><h2 id="a">A</h2><p>...</p></section><section><h2 id="b">B</h2><p>...</p></section>
func WrapHeadingSections(html string) string {
	locs := sectionHeadingRegex.FindAllStringIndex(html, -1)
	if len(locs) == 0 {
		return html
	}

	var sb strings.Builder
	sb.Grow(len(html) + len(locs)*19) // <section></section> = 19 bytes

	// Content before first h2 (if any)
	if locs[0][0] > 0 {
		preamble := strings.TrimSpace(html[:locs[0][0]])
		if preamble != "" {
			sb.WriteString("<section>\n")
			sb.WriteString(preamble)
			sb.WriteString("\n</section>\n")
		}
	}

	// Each h2 section
	for i, loc := range locs {
		start := loc[0]
		var end int
		if i+1 < len(locs) {
			end = locs[i+1][0]
		} else {
			end = len(html)
		}

		chunk := strings.TrimSpace(html[start:end])
		if chunk != "" {
			sb.WriteString("<section>\n")
			sb.WriteString(chunk)
			sb.WriteString("\n</section>\n")
		}
	}

	return sb.String()
}
