package scaffold

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// NewContentOptions holds parameters for content scaffolding.
type NewContentOptions struct {
	ContentDir    string // e.g., "content"
	ArchetypesDir string // e.g., "archetypes"
	Path          string // e.g., "posts/my-great-idea"
}

// ArchetypeVars are the variables available in archetype templates.
type ArchetypeVars struct {
	Title   string // "My Great Idea"
	Date    string // RFC3339 timestamp
	Section string // "posts"
	Slug    string // "my-great-idea"
}

// NewContent creates a new content file from an archetype template.
func NewContent(opts NewContentOptions) (string, error) {
	section, slug := ParseContentPath(opts.Path)

	// Strip .md suffix from slug if provided
	slug = strings.TrimSuffix(slug, ".md")

	// Determine output path
	outputPath := filepath.Join(opts.ContentDir, opts.Path)
	if !strings.HasSuffix(outputPath, ".md") {
		outputPath += ".md"
	}

	// Don't overwrite existing files
	if _, err := os.Stat(outputPath); err == nil {
		return "", fmt.Errorf("%s already exists", outputPath)
	}

	// Find archetype: section-specific → default → built-in
	templateContent := findArchetype(opts.ArchetypesDir, section)

	// Prepare variables
	vars := ArchetypeVars{
		Title:   SlugToTitle(slug),
		Date:    time.Now().Format(time.RFC3339),
		Section: section,
		Slug:    slug,
	}

	// Execute template
	tmpl, err := template.New("archetype").Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("parsing archetype: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("executing archetype: %w", err)
	}

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return "", fmt.Errorf("creating directories: %w", err)
	}

	// Write file
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}

	return outputPath, nil
}

// findArchetype looks for a section-specific archetype, then default, then built-in.
func findArchetype(archetypesDir, section string) string {
	// Try section-specific archetype
	if section != "" {
		path := filepath.Join(archetypesDir, section+".md")
		if data, err := os.ReadFile(path); err == nil {
			return string(data)
		}
	}

	// Try default archetype
	path := filepath.Join(archetypesDir, "default.md")
	if data, err := os.ReadFile(path); err == nil {
		return string(data)
	}

	// Built-in fallback
	return builtInArchetype(section)
}

// ParseContentPath splits "posts/my-slug" into (section, slug).
// For "about" with no slash, returns ("", "about").
// For "docs/guide/intro", returns ("docs/guide", "intro").
func ParseContentPath(path string) (section, slug string) {
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash == -1 {
		return "", path
	}
	return path[:lastSlash], path[lastSlash+1:]
}

// SlugToTitle converts "my-great-idea" to "My Great Idea".
func SlugToTitle(slug string) string {
	words := strings.Split(strings.ReplaceAll(slug, "_", "-"), "-")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

// builtInArchetype returns the default archetype for a section.
func builtInArchetype(section string) string {
	if section == "posts" {
		return `---
title: "{{.Title}}"
date: {{.Date}}
draft: true
tags: []
---
`
	}
	return `---
title: "{{.Title}}"
description: ""
---
`
}
