package content

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"gopkg.in/yaml.v3"
)

// Parser handles parsing of Markdown content files
type Parser struct {
	contentDir string
	markdown   goldmark.Markdown
}

// HighlightOptions configures syntax highlighting in the parser.
// Pass to NewParser to enable Chroma-based code highlighting.
type HighlightOptions struct {
	Style       string // Chroma style name (e.g. "monokai", "github")
	LineNumbers bool   // Show line numbers in code blocks
}

// NewParser creates a new Parser instance. Pass a HighlightOptions to enable
// syntax highlighting; omit it to disable highlighting.
func NewParser(contentDir string, opts ...HighlightOptions) *Parser {
	// Configure goldmark with GitHub Flavored Markdown + footnotes
	extensions := []goldmark.Extender{
		extension.GFM,
		extension.Table,
		extension.Strikethrough,
		extension.Linkify,
		extension.TaskList,
		extension.Footnote,
	}

	// Add syntax highlighting if configured
	if len(opts) > 0 && opts[0].Style != "" {
		hl := opts[0]
		extensions = append(extensions, highlighting.NewHighlighting(
			highlighting.WithFormatOptions(
				html.WithClasses(true),
				html.WithLineNumbers(hl.LineNumbers),
			),
		))
	}

	md := goldmark.New(
		goldmark.WithExtensions(extensions...),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithASTTransformers(
				util.Prioritized(&externalLinkTransformer{}, 500),
			),
		),
		goldmark.WithRendererOptions(
			goldmarkhtml.WithHardWraps(),
			goldmarkhtml.WithXHTML(),
		),
	)

	return &Parser{
		contentDir: contentDir,
		markdown:   md,
	}
}

// ParseAll walks the content directory and parses all Markdown files
func (p *Parser) ParseAll(ctx context.Context) ([]*Page, error) {
	var pages []*Page

	err := filepath.WalkDir(p.contentDir, func(path string, d os.DirEntry, err error) error {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process .md files
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		page, err := p.ParseFile(ctx, path)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}

		pages = append(pages, page)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return pages, nil
}

// ParseFile parses a single Markdown file
func (p *Parser) ParseFile(ctx context.Context, path string) (*Page, error) {
	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	// Parse frontmatter and body
	frontmatter, body, err := p.parseFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("parsing frontmatter: %w", err)
	}

	// Store raw markdown content
	rawContent := string(body)

	// Convert Markdown to HTML
	var buf bytes.Buffer
	if err := p.markdown.Convert(body, &buf); err != nil {
		return nil, fmt.Errorf("converting markdown: %w", err)
	}
	htmlContent := buf.String()

	// Wrap h2-delimited blocks in <section> tags when frontmatter requests it.
	// Enables CSS structural selectors (section:nth-of-type) for alternating
	// backgrounds via the token system.
	if getBoolDefault(frontmatter, "sections", false) {
		htmlContent = WrapHeadingSections(htmlContent)
	}

	// Calculate text metrics
	plainText := StripHTML(htmlContent)
	wordCount := WordCount(plainText)
	readingTime := ReadingTime(wordCount)

	// Extract summary (uses <!--more--> marker or first paragraph)
	summary := ExtractSummary(rawContent, htmlContent, 50)

	// Generate table of contents (h2-h4 by default)
	// "page-toc" ID enables go-components scroll-spy JS binding
	toc := GenerateTOC(htmlContent, 2, 4, "page-toc")

	// Create page
	page := &Page{
		Path:        path,
		Content:     htmlContent,
		RawContent:  rawContent,
		Frontmatter: frontmatter,
		Layout:      getStringDefault(frontmatter, "layout", "page"),
		Draft:       getBoolDefault(frontmatter, "draft", false),
		URL:         p.generateURL(path),
		Title:       getStringDefault(frontmatter, "title", ""),
		Description: getStringDefault(frontmatter, "description", ""),
		Tags:        getStringSlice(frontmatter, "tags"),
		Author:      getStringDefault(frontmatter, "author", ""),
		Section:     p.extractSection(path),
		Weight:      getIntDefault(frontmatter, "weight", 0),
		Aliases:     getStringSlice(frontmatter, "aliases"),
		WordCount:   wordCount,
		ReadingTime: readingTime,
		Summary:     summary,
		TOC:         toc,
	}

	// Parse date if present
	if dateStr, ok := frontmatter["date"].(string); ok {
		date, err := parseDate(dateStr)
		if err == nil {
			page.Date = date
		}
	} else if dateTime, ok := frontmatter["date"].(time.Time); ok {
		// YAML parses bare dates (2006-01-02) as UTC midnight, but users mean
		// local time. Normalize date-only values to local timezone.
		page.Date = normalizeToLocal(dateTime)
	}

	return page, nil
}

// parseFrontmatter extracts YAML frontmatter from content
func (p *Parser) parseFrontmatter(content []byte) (map[string]any, []byte, error) {
	// Check for frontmatter delimiter
	if !bytes.HasPrefix(content, []byte("---\n")) && !bytes.HasPrefix(content, []byte("---\r\n")) {
		// No frontmatter, return empty map and full content
		return make(map[string]any), content, nil
	}

	// Determine line ending and skip opening delimiter
	var closingDelimiter []byte
	var skipBytes int
	var lineEnding []byte
	if bytes.HasPrefix(content, []byte("---\n")) {
		closingDelimiter = []byte("---\n")
		skipBytes = 4 // Skip "---\n"
		lineEnding = []byte("\n")
	} else {
		closingDelimiter = []byte("---\r\n")
		skipBytes = 5 // Skip "---\r\n"
		lineEnding = []byte("\r\n")
	}

	// Skip the opening delimiter
	remaining := content[skipBytes:]

	// Look for closing delimiter (it should be at the start of a line)
	closingIndex := -1
	searchPos := 0
	for {
		idx := bytes.Index(remaining[searchPos:], closingDelimiter)
		if idx == -1 {
			break
		}
		// Check if this closing delimiter is at the start of a line
		actualIdx := searchPos + idx
		if actualIdx == 0 || bytes.HasSuffix(remaining[:actualIdx], lineEnding) {
			closingIndex = actualIdx
			break
		}
		searchPos = actualIdx + 1
	}

	if closingIndex == -1 {
		// No closing delimiter found, treat as no frontmatter
		return make(map[string]any), content, nil
	}

	// Extract frontmatter content and body
	frontmatterContent := remaining[:closingIndex]
	bodyStart := closingIndex + len(closingDelimiter)
	body := remaining[bodyStart:]

	// Parse YAML frontmatter
	var frontmatter map[string]any

	// Only parse if there's actual content between delimiters
	if len(bytes.TrimSpace(frontmatterContent)) > 0 {
		if err := yaml.Unmarshal(frontmatterContent, &frontmatter); err != nil {
			return nil, nil, fmt.Errorf("parsing YAML: %w", err)
		}
	}

	if frontmatter == nil {
		frontmatter = make(map[string]any)
	}

	return frontmatter, body, nil
}

// generateURL creates a clean URL from the file path
func (p *Parser) generateURL(path string) string {
	// Get relative path from content directory
	rel, err := filepath.Rel(p.contentDir, path)
	if err != nil {
		rel = filepath.Base(path)
	}

	// Remove .md extension
	rel = strings.TrimSuffix(rel, ".md")

	// Convert to forward slashes
	rel = filepath.ToSlash(rel)

	// Handle index files (both index.md and _index.md)
	// _index.md is Hugo's convention for section list pages
	isIndex := rel == "index" || rel == "_index" ||
		strings.HasSuffix(rel, "/index") || strings.HasSuffix(rel, "/_index")

	if isIndex {
		// Remove the index filename, keeping just the directory path
		rel = strings.TrimSuffix(rel, "_index")
		rel = strings.TrimSuffix(rel, "index")
		rel = strings.TrimSuffix(rel, "/") // Remove trailing slash from directory

		if rel == "" {
			return "/"
		}
		return "/" + rel + "/"
	}

	// Return with leading and trailing slash
	return "/" + rel + "/"
}

// extractSection extracts the content section from a file path
// For example, "content/blog/post.md" returns "blog"
// Files at the root of content return empty string
func (p *Parser) extractSection(path string) string {
	// Get relative path from content directory
	rel, err := filepath.Rel(p.contentDir, path)
	if err != nil {
		return ""
	}

	// Convert to forward slashes for consistency
	rel = filepath.ToSlash(rel)

	// Split by directory separator
	parts := strings.Split(rel, "/")

	// If there's only one part (filename), there's no section
	if len(parts) <= 1 {
		return ""
	}

	// First directory is the section
	return parts[0]
}

// parseDate attempts to parse a date string in multiple formats.
// Date-only formats are parsed in the local timezone (what the user intends
// when writing "date: 2026-02-12"), while datetime formats preserve their zone.
func parseDate(dateStr string) (time.Time, error) {
	// Date-only format — parse in local timezone
	if t, err := time.ParseInLocation("2006-01-02", dateStr, time.Local); err == nil {
		return t, nil
	}

	// Datetime formats — preserve timezone info
	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC822,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// normalizeToLocal converts a UTC midnight time (from YAML auto-parsing bare dates)
// to local midnight. If the time has a non-zero hour/minute/second, it's left as-is.
func normalizeToLocal(t time.Time) time.Time {
	if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 && t.Location() == time.UTC {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	}
	return t
}

// externalLinkTransformer adds target="_blank" and rel="noopener noreferrer"
// to links pointing to external URLs (http:// or https://).
type externalLinkTransformer struct{}

func (t *externalLinkTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		link, ok := n.(*ast.Link)
		if !ok {
			return ast.WalkContinue, nil
		}
		dest := string(link.Destination)
		if strings.HasPrefix(dest, "http://") || strings.HasPrefix(dest, "https://") {
			link.SetAttributeString("target", "_blank")
			link.SetAttributeString("rel", "noopener noreferrer")
		}
		return ast.WalkContinue, nil
	})
}
