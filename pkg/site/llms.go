// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package site

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/dmoose/templsite/pkg/content"
	"gopkg.in/yaml.v3"
)

// LLMsTxt generates the /llms.txt index file following the llms.txt specification.
// It produces an H1 title, blockquote summary, and H2 sections with links to
// companion markdown files for each published page.
func (s *Site) LLMsTxt() string {
	var b strings.Builder

	// H1 title
	b.WriteString("# ")
	b.WriteString(s.Config.Title)
	b.WriteString("\n\n")

	// Blockquote summary
	desc := s.Config.LLMs.Description
	if desc == "" {
		desc = s.Config.Description
	}
	if desc != "" {
		b.WriteString("> ")
		b.WriteString(desc)
		b.WriteString("\n\n")
	}

	baseURL := strings.TrimRight(s.Config.BaseURL, "/")

	if len(s.Config.LLMs.Sections) > 0 {
		s.writeLLMsConfiguredSections(&b, baseURL)
	} else {
		s.writeLLMsAutoSections(&b, baseURL)
	}

	return b.String()
}

// writeLLMsConfiguredSections writes sections defined in config, matching pages
// by section name (the directory under content/).
func (s *Site) writeLLMsConfiguredSections(b *strings.Builder, baseURL string) {
	// Collect pages that go under "## Optional"
	var optionalPages []*content.Page

	for _, sec := range s.Config.LLMs.Sections {
		pages := s.pagesForLLMsSection(sec.Pattern)
		if len(pages) == 0 {
			continue
		}

		if sec.Priority == "optional" {
			optionalPages = append(optionalPages, pages...)
			continue
		}

		b.WriteString("## ")
		b.WriteString(sec.Name)
		b.WriteString("\n\n")
		writePageLinks(b, pages, baseURL)
		b.WriteString("\n")
	}

	if len(optionalPages) > 0 {
		b.WriteString("## Optional\n\n")
		writePageLinks(b, optionalPages, baseURL)
		b.WriteString("\n")
	}
}

// writeLLMsAutoSections writes sections auto-derived from the site's section structure.
func (s *Site) writeLLMsAutoSections(b *strings.Builder, baseURL string) {
	// Collect and sort sections for deterministic output
	sections := s.AllSections()
	sort.Slice(sections, func(i, j int) bool {
		return sections[i].Name < sections[j].Name
	})

	// Write root pages first (pages not in a named section)
	rootPages := s.llmsFilteredPages(s.RegularPagesInSection("_root"))
	if len(rootPages) > 0 {
		b.WriteString("## Pages\n\n")
		writePageLinks(b, rootPages, baseURL)
		b.WriteString("\n")
	}

	for _, section := range sections {
		if section.Name == "_root" {
			continue
		}
		pages := s.llmsFilteredPages(section.RegularPages())
		if len(pages) == 0 {
			continue
		}

		b.WriteString("## ")
		b.WriteString(section.Title)
		b.WriteString("\n\n")
		writePageLinks(b, pages, baseURL)
		b.WriteString("\n")
	}
}

// LLMsFull generates /llms-full.txt with all linked content inlined.
func (s *Site) LLMsFull() string {
	var b strings.Builder

	// H1 title
	b.WriteString("# ")
	b.WriteString(s.Config.Title)
	b.WriteString("\n\n")

	// Blockquote summary
	desc := s.Config.LLMs.Description
	if desc == "" {
		desc = s.Config.Description
	}
	if desc != "" {
		b.WriteString("> ")
		b.WriteString(desc)
		b.WriteString("\n\n")
	}

	baseURL := strings.TrimRight(s.Config.BaseURL, "/")

	if len(s.Config.LLMs.Sections) > 0 {
		s.writeLLMsFullConfigured(&b, baseURL)
	} else {
		s.writeLLMsFullAuto(&b, baseURL)
	}

	return b.String()
}

func (s *Site) writeLLMsFullConfigured(b *strings.Builder, baseURL string) {
	var optionalPages []*content.Page

	for _, sec := range s.Config.LLMs.Sections {
		pages := s.pagesForLLMsSection(sec.Pattern)
		if len(pages) == 0 {
			continue
		}

		if sec.Priority == "optional" {
			optionalPages = append(optionalPages, pages...)
			continue
		}

		b.WriteString("## ")
		b.WriteString(sec.Name)
		b.WriteString("\n\n")
		s.writePageContents(b,pages, baseURL)
	}

	if len(optionalPages) > 0 {
		b.WriteString("## Optional\n\n")
		s.writePageContents(b,optionalPages, baseURL)
	}
}

func (s *Site) writeLLMsFullAuto(b *strings.Builder, baseURL string) {
	sections := s.AllSections()
	sort.Slice(sections, func(i, j int) bool {
		return sections[i].Name < sections[j].Name
	})

	rootPages := s.llmsFilteredPages(s.RegularPagesInSection("_root"))
	if len(rootPages) > 0 {
		b.WriteString("## Pages\n\n")
		s.writePageContents(b,rootPages, baseURL)
	}

	for _, section := range sections {
		if section.Name == "_root" {
			continue
		}
		pages := s.llmsFilteredPages(section.RegularPages())
		if len(pages) == 0 {
			continue
		}

		b.WriteString("## ")
		b.WriteString(section.Title)
		b.WriteString("\n\n")
		s.writePageContents(b,pages, baseURL)
	}
}

// CompanionMarkdown returns clean markdown for a page's companion .md file.
// It prepends a title heading if the raw content doesn't already start with one.
// If the page's frontmatter includes llms_data, the referenced data files are
// appended as YAML blocks, making data-driven page content available to LLMs.
func CompanionMarkdown(s *Site, p *content.Page) string {
	var b strings.Builder

	raw := strings.TrimSpace(p.RawContent)

	// Prepend title if the raw content doesn't already have one
	if !strings.HasPrefix(raw, "# ") && p.Title != "" {
		b.WriteString("# ")
		b.WriteString(p.Title)
		b.WriteString("\n\n")
	}

	if raw != "" {
		b.WriteString(raw)
		b.WriteString("\n")
	}

	// Append referenced data files
	dataKeys := llmsDataKeys(p)
	if len(dataKeys) > 0 && s != nil {
		for _, key := range dataKeys {
			data := s.GetData(key)
			if data == nil {
				slog.Debug("llms_data key not found", "key", key, "page", p.URL)
				continue
			}

			out, err := yaml.Marshal(data)
			if err != nil {
				slog.Debug("failed to marshal llms_data", "key", key, "error", err)
				continue
			}

			b.WriteString("\n---\n\n")
			b.WriteString("## Data: ")
			b.WriteString(key)
			b.WriteString("\n\n")
			b.WriteString("```yaml\n")
			b.WriteString(string(out))
			b.WriteString("```\n")
		}
	}

	return b.String()
}

// llmsDataKeys extracts the llms_data frontmatter field as a string slice.
func llmsDataKeys(p *content.Page) []string {
	val, ok := p.Frontmatter["llms_data"]
	if !ok {
		return nil
	}

	switch v := val.(type) {
	case []any:
		keys := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				keys = append(keys, s)
			}
		}
		return keys
	case []string:
		return v
	case string:
		return []string{v}
	}

	return nil
}

// companionPath returns the output file path for a page's companion .md file.
func (s *Site) companionPath(p *content.Page) string {
	htmlPath := s.GetOutputPath(p.URL)
	return htmlPath + ".md"
}

// writeLLMsFiles generates and writes llms.txt, llms-full.txt, and companion
// markdown files to the output directory.
func (s *Site) writeLLMsFiles() {
	outputDir := s.OutputDir()

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		slog.Warn("failed to create output directory for llms files", "error", err)
		return
	}

	// LLMs files are always regenerated — they derive from content and must
	// stay current across incremental builds (unlike robots.txt/sitemap.xml
	// which use writeIfMissing to allow static/ overrides).
	writeFile := func(name, content string) {
		path := filepath.Join(outputDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			slog.Warn("failed to write generated file", "file", name, "error", err)
			return
		}
		slog.Debug("generated file written", "file", name)
	}

	writeFile("llms.txt", s.LLMsTxt())
	writeFile("llms-full.txt", s.LLMsFull())

	// Write companion markdown files for each published page
	count := 0
	for _, page := range s.Pages {
		if !page.IsPublished() || s.isLLMsExcluded(page) {
			continue
		}

		mdPath := s.companionPath(page)
		if err := os.MkdirAll(filepath.Dir(mdPath), 0755); err != nil {
			slog.Warn("failed to create directory for companion file", "error", err)
			continue
		}

		md := CompanionMarkdown(s, page)
		if err := os.WriteFile(mdPath, []byte(md), 0644); err != nil {
			slog.Warn("failed to write companion markdown", "path", mdPath, "error", err)
			continue
		}
		count++
	}

	if count > 0 {
		slog.Info("llms.txt files generated", "companions", count)
	}
}

// pagesForLLMsSection returns published, non-excluded pages matching a section
// pattern. The pattern is matched against the page's section name (directory).
func (s *Site) pagesForLLMsSection(pattern string) []*content.Page {
	var pages []*content.Page
	for _, page := range s.Pages {
		if !page.IsPublished() || s.isIndexPage(page) || s.isLLMsExcluded(page) {
			continue
		}
		if matchSection(page.Section, pattern) {
			pages = append(pages, page)
		}
	}
	return pages
}

// llmsFilteredPages filters out excluded and unpublished pages.
func (s *Site) llmsFilteredPages(pages []*content.Page) []*content.Page {
	var filtered []*content.Page
	for _, p := range pages {
		if p.IsPublished() && !s.isLLMsExcluded(p) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// isLLMsExcluded checks if a page should be excluded from llms.txt output.
// A page is excluded if its frontmatter sets llms: false, or if it matches
// any pattern in the config's exclude list.
func (s *Site) isLLMsExcluded(p *content.Page) bool {
	// Check frontmatter opt-out
	if v, ok := p.Frontmatter["llms"].(bool); ok && !v {
		return true
	}

	// Check exclude patterns
	for _, pattern := range s.Config.LLMs.Exclude {
		if matchSection(p.Section, pattern) {
			return true
		}
	}

	return false
}

// matchSection checks if a section name matches a pattern.
// Patterns can be exact names ("blog") or use a trailing wildcard ("docs/**").
func matchSection(section, pattern string) bool {
	// Strip trailing /** or /* for section matching
	clean := strings.TrimSuffix(pattern, "/**")
	clean = strings.TrimSuffix(clean, "/*")

	return section == clean
}

// writePageLinks writes a markdown list of page links.
func writePageLinks(b *strings.Builder, pages []*content.Page, baseURL string) {
	for _, p := range pages {
		companionURL := pageCompanionURL(p, baseURL)
		b.WriteString("- [")
		b.WriteString(p.Title)
		b.WriteString("](")
		b.WriteString(companionURL)
		b.WriteString(")")
		if p.Description != "" {
			b.WriteString(": ")
			b.WriteString(p.Description)
		}
		b.WriteString("\n")
	}
}

// writePageContents writes each page's title, link, and full content including
// any referenced data files.
func (s *Site) writePageContents(b *strings.Builder, pages []*content.Page, baseURL string) {
	for _, p := range pages {
		companionURL := pageCompanionURL(p, baseURL)
		b.WriteString("### ")
		b.WriteString(p.Title)
		b.WriteString("\n\n")
		b.WriteString(fmt.Sprintf("Source: [%s](%s)\n\n", companionURL, companionURL))
		raw := strings.TrimSpace(p.RawContent)
		if raw != "" {
			b.WriteString(raw)
			b.WriteString("\n\n")
		}

		// Inline referenced data files
		for _, key := range llmsDataKeys(p) {
			data := s.GetData(key)
			if data == nil {
				continue
			}
			out, err := yaml.Marshal(data)
			if err != nil {
				continue
			}
			b.WriteString("#### Data: ")
			b.WriteString(key)
			b.WriteString("\n\n")
			b.WriteString("```yaml\n")
			b.WriteString(string(out))
			b.WriteString("```\n\n")
		}
	}
}

// pageCompanionURL returns the full URL to a page's companion .md file.
func pageCompanionURL(p *content.Page, baseURL string) string {
	url := p.URL

	// Handle URLs that end with / (directory-style clean URLs)
	if strings.HasSuffix(url, "/") {
		return baseURL + url + "index.html.md"
	}

	return baseURL + url + ".md"
}

// LLMsPages returns all pages that would be included in llms.txt output,
// respecting exclusion rules. Useful for templates or custom rendering.
func (s *Site) LLMsPages() []*content.Page {
	var pages []*content.Page
	for _, page := range s.Pages {
		if page.IsPublished() && !s.isIndexPage(page) && !s.isLLMsExcluded(page) {
			pages = append(pages, page)
		}
	}
	return pages
}

// LLMsSectionNames returns the section names that will appear in llms.txt,
// either from config or auto-derived. Useful for debugging configuration.
func (s *Site) LLMsSectionNames() []string {
	if len(s.Config.LLMs.Sections) > 0 {
		names := make([]string, len(s.Config.LLMs.Sections))
		for i, sec := range s.Config.LLMs.Sections {
			names[i] = sec.Name
		}
		return names
	}

	names := s.SectionNames()
	slices.Sort(names)
	return names
}
