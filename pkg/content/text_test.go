package content

import (
	"strings"
	"testing"
)

func TestStripHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple paragraph",
			input:    "<p>Hello World</p>",
			expected: "Hello World",
		},
		{
			name:     "nested tags",
			input:    "<div><p>Hello <strong>World</strong></p></div>",
			expected: "Hello World",
		},
		{
			name:     "multiple paragraphs",
			input:    "<p>First</p><p>Second</p>",
			expected: "First Second",
		},
		{
			name:     "with whitespace",
			input:    "<p>Hello</p>\n\n<p>World</p>",
			expected: "Hello World",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no tags",
			input:    "Plain text",
			expected: "Plain text",
		},
		{
			name:     "self-closing tags",
			input:    "<p>Line<br/>break</p>",
			expected: "Line break",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripHTML(tt.input)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestWordCount(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "simple sentence",
			input:    "Hello World",
			expected: 2,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "only whitespace",
			input:    "   \t\n   ",
			expected: 0,
		},
		{
			name:     "single word",
			input:    "Hello",
			expected: 1,
		},
		{
			name:     "multiple spaces between words",
			input:    "Hello    World    Test",
			expected: 3,
		},
		{
			name:     "paragraph",
			input:    "This is a test paragraph with several words in it.",
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WordCount(tt.input)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestReadingTime(t *testing.T) {
	tests := []struct {
		name      string
		wordCount int
		expected  int
	}{
		{
			name:      "zero words",
			wordCount: 0,
			expected:  0,
		},
		{
			name:      "negative words",
			wordCount: -10,
			expected:  0,
		},
		{
			name:      "under 200 words",
			wordCount: 100,
			expected:  1,
		},
		{
			name:      "exactly 200 words",
			wordCount: 200,
			expected:  1,
		},
		{
			name:      "201 words",
			wordCount: 201,
			expected:  2,
		},
		{
			name:      "400 words",
			wordCount: 400,
			expected:  2,
		},
		{
			name:      "1000 words",
			wordCount: 1000,
			expected:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReadingTime(tt.wordCount)
			if result != tt.expected {
				t.Errorf("expected %d minutes, got %d", tt.expected, result)
			}
		})
	}
}

func TestExtractSummary(t *testing.T) {
	tests := []struct {
		name        string
		markdown    string
		html        string
		maxWords    int
		shouldMatch func(result string) bool
		desc        string
	}{
		{
			name:     "with more marker",
			markdown: "First paragraph.\n\n<!--more-->\n\nSecond paragraph.",
			html:     "<p>First paragraph.</p>\n<p>Second paragraph.</p>",
			maxWords: 50,
			shouldMatch: func(result string) bool {
				// Should contain first paragraph but not second
				return strings.Contains(result, "First paragraph") &&
					!strings.Contains(result, "Second paragraph")
			},
			desc: "should contain first paragraph but not second",
		},
		{
			name:     "first paragraph extraction",
			markdown: "First paragraph here.\n\nSecond paragraph here.",
			html:     "<p>First paragraph here.</p>\n<p>Second paragraph here.</p>",
			maxWords: 50,
			shouldMatch: func(result string) bool {
				return strings.Contains(result, "First paragraph here")
			},
			desc: "should contain first paragraph",
		},
		{
			name:     "multiple paragraphs before more",
			markdown: "First para.\n\nSecond para.\n\n<!--more-->\n\nThird para.",
			html:     "<p>First para.</p>\n<p>Second para.</p>\n<p>Third para.</p>",
			maxWords: 50,
			shouldMatch: func(result string) bool {
				// Should contain both paragraphs before <!--more-->
				return strings.Contains(result, "First para") &&
					strings.Contains(result, "Second para") &&
					!strings.Contains(result, "Third para")
			},
			desc: "should contain both paragraphs before <!--more-->",
		},
		{
			name:     "strips footnote references",
			markdown: "AI tools boost productivity by 55%[^1]. The reality is sobering.",
			html:     `<p>AI tools boost productivity by 55%<sup id="fnref:1"><a href="#fn:1" class="footnote-ref" role="doc-noteref">1</a></sup>. The reality is sobering.</p>`,
			maxWords: 50,
			shouldMatch: func(result string) bool {
				return result == "AI tools boost productivity by 55%. The reality is sobering." &&
					!strings.Contains(result, "sup") &&
					!strings.Contains(result, "fnref")
			},
			desc: "should strip footnote <sup> elements and return plain text",
		},
		{
			name:     "strips HTML tags from summary",
			markdown: "Hello **bold** world.\n\nSecond paragraph.",
			html:     "<p>Hello <strong>bold</strong> world.</p>\n<p>Second paragraph.</p>",
			maxWords: 50,
			shouldMatch: func(result string) bool {
				return result == "Hello bold world." &&
					!strings.Contains(result, "<strong>")
			},
			desc: "should return plain text without HTML tags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractSummary(tt.markdown, tt.html, tt.maxWords)
			if !tt.shouldMatch(result) {
				t.Errorf("%s, got '%s'", tt.desc, result)
			}
		})
	}
}

func TestExtractFirstParagraph(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "simple paragraph",
			html:     "<p>Hello World</p>",
			expected: "Hello World",
		},
		{
			name:     "multiple paragraphs",
			html:     "<p>First</p><p>Second</p>",
			expected: "First",
		},
		{
			name:     "no paragraph",
			html:     "<div>No paragraph</div>",
			expected: "",
		},
		{
			name:     "paragraph with attributes",
			html:     `<p class="intro">Introduction text</p>`,
			expected: "Introduction text",
		},
		{
			name:     "paragraph with nested tags",
			html:     "<p>Hello <strong>World</strong></p>",
			expected: "Hello <strong>World</strong>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFirstParagraph(tt.html)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestCloseOpenTags(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "already closed",
			html:     "<p>Complete</p>",
			expected: "<p>Complete</p>",
		},
		{
			name:     "unclosed paragraph",
			html:     "<p>Unclosed",
			expected: "<p>Unclosed</p>",
		},
		{
			name:     "nested unclosed",
			html:     "<div><p>Nested",
			expected: "<div><p>Nested</p></div>",
		},
		{
			name:     "self-closing ignored",
			html:     "<p>With<br>break",
			expected: "<p>With<br>break</p>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := closeOpenTags(tt.html)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
