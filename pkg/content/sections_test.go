package content

import (
	"strings"
	"testing"
)

func TestWrapHeadingSections(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		input     string
		wantParts []string
		wantCount int // expected number of <section> tags
	}{
		{
			name:      "no headings returns unchanged",
			input:     "<p>Hello world</p>",
			wantParts: []string{"<p>Hello world</p>"},
			wantCount: 0,
		},
		{
			name:  "single h2 wraps in section",
			input: `<h2 id="intro">Intro</h2><p>Content here</p>`,
			wantParts: []string{
				"<section>",
				`<h2 id="intro">Intro</h2>`,
				"<p>Content here</p>",
				"</section>",
			},
			wantCount: 1,
		},
		{
			name:  "two h2s create two sections",
			input: `<h2 id="a">A</h2><p>First</p><h2 id="b">B</h2><p>Second</p>`,
			wantParts: []string{
				"<section>",
				`<h2 id="a">A</h2>`,
				"<p>First</p>",
				"</section>",
				`<h2 id="b">B</h2>`,
				"<p>Second</p>",
			},
			wantCount: 2,
		},
		{
			name:  "content before first h2 gets its own section",
			input: `<p>Preamble</p><h2 id="a">A</h2><p>Body</p>`,
			wantParts: []string{
				"<section>",
				"<p>Preamble</p>",
				"</section>",
				`<h2 id="a">A</h2>`,
			},
			wantCount: 2,
		},
		{
			name:  "h3 inside h2 section is not split",
			input: `<h2 id="a">A</h2><p>Intro</p><h3 id="sub">Sub</h3><p>Detail</p>`,
			wantParts: []string{
				`<h2 id="a">A</h2>`,
				`<h3 id="sub">Sub</h3>`,
				"<p>Detail</p>",
			},
			wantCount: 1,
		},
		{
			name:      "empty string returns empty",
			input:     "",
			wantParts: []string{},
			wantCount: 0,
		},
		{
			name:  "h2 without id attribute",
			input: `<h2>No ID</h2><p>Content</p>`,
			wantParts: []string{
				"<section>",
				"<h2>No ID</h2>",
			},
			wantCount: 1,
		},
		{
			name:  "multiple elements between h2s preserved",
			input: `<h2 id="a">A</h2><p>P1</p><ul><li>Item</li></ul><p>P2</p><h2 id="b">B</h2><p>P3</p>`,
			wantParts: []string{
				"<p>P1</p>",
				"<ul><li>Item</li></ul>",
				"<p>P2</p>",
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := WrapHeadingSections(tt.input)

			for _, part := range tt.wantParts {
				if !strings.Contains(got, part) {
					t.Errorf("WrapHeadingSections() missing %q in output:\n%s", part, got)
				}
			}

			sectionCount := strings.Count(got, "<section>")
			if sectionCount != tt.wantCount {
				t.Errorf("WrapHeadingSections() has %d <section> tags, want %d.\nOutput:\n%s",
					sectionCount, tt.wantCount, got)
			}

			// Verify matching open/close tags
			openCount := strings.Count(got, "<section>")
			closeCount := strings.Count(got, "</section>")
			if openCount != closeCount {
				t.Errorf("Mismatched section tags: %d open, %d close", openCount, closeCount)
			}
		})
	}
}
