package content

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v3"
)

// Parser handles parsing of Markdown content files
type Parser struct {
	contentDir string
	markdown   goldmark.Markdown
}

// NewParser creates a new Parser instance
func NewParser(contentDir string) *Parser {
	// Configure goldmark with GitHub Flavored Markdown
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.Linkify,
			extension.TaskList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
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

	// Convert Markdown to HTML
	var buf bytes.Buffer
	if err := p.markdown.Convert(body, &buf); err != nil {
		return nil, fmt.Errorf("converting markdown: %w", err)
	}

	// Create page
	page := &Page{
		Path:        path,
		Content:     buf.String(),
		Frontmatter: frontmatter,
		Layout:      getStringDefault(frontmatter, "layout", "page"),
		Draft:       getBoolDefault(frontmatter, "draft", false),
		URL:         p.generateURL(path),
		Title:       getStringDefault(frontmatter, "title", ""),
		Description: getStringDefault(frontmatter, "description", ""),
		Tags:        getStringSlice(frontmatter, "tags"),
		Author:      getStringDefault(frontmatter, "author", ""),
	}

	// Parse date if present
	if dateStr, ok := frontmatter["date"].(string); ok {
		date, err := parseDate(dateStr)
		if err == nil {
			page.Date = date
		}
	} else if dateTime, ok := frontmatter["date"].(time.Time); ok {
		// YAML might parse dates directly as time.Time
		page.Date = dateTime
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

	// Handle index files
	if rel == "index" || strings.HasSuffix(rel, "/index") {
		rel = strings.TrimSuffix(rel, "index")
		if rel == "" {
			return "/"
		}
		return "/" + rel
	}

	// Return with leading and trailing slash
	return "/" + rel + "/"
}

// parseDate attempts to parse a date string in multiple formats
func parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
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
