// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseContentPath(t *testing.T) {
	tests := []struct {
		input   string
		section string
		slug    string
	}{
		{"posts/my-post", "posts", "my-post"},
		{"about", "", "about"},
		{"docs/guide/intro", "docs/guide", "intro"},
		{"a/b", "a", "b"},
	}

	for _, tt := range tests {
		section, slug := ParseContentPath(tt.input)
		if section != tt.section || slug != tt.slug {
			t.Errorf("ParseContentPath(%q) = (%q, %q), want (%q, %q)",
				tt.input, section, slug, tt.section, tt.slug)
		}
	}
}

func TestSlugToTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"my-great-idea", "My Great Idea"},
		{"hello", "Hello"},
		{"multi-word-title-here", "Multi Word Title Here"},
		{"ios_privacy_policy", "Ios Privacy Policy"},
	}

	for _, tt := range tests {
		result := SlugToTitle(tt.input)
		if result != tt.expected {
			t.Errorf("SlugToTitle(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNewContentWithBuiltInArchetype(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	opts := NewContentOptions{
		ContentDir:    contentDir,
		ArchetypesDir: filepath.Join(tmpDir, "archetypes"), // doesn't exist
		Path:          "about",
	}

	outputPath, err := NewContent(opts)
	if err != nil {
		t.Fatalf("NewContent failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, `title: "About"`) {
		t.Errorf("missing title, got:\n%s", s)
	}
	if !strings.Contains(s, `description: ""`) {
		t.Errorf("missing description, got:\n%s", s)
	}
}

func TestNewContentPostsBuiltIn(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	opts := NewContentOptions{
		ContentDir:    contentDir,
		ArchetypesDir: filepath.Join(tmpDir, "archetypes"),
		Path:          "posts/hello-world",
	}

	outputPath, err := NewContent(opts)
	if err != nil {
		t.Fatalf("NewContent failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, `title: "Hello World"`) {
		t.Errorf("missing title, got:\n%s", s)
	}
	if !strings.Contains(s, "draft: true") {
		t.Errorf("missing draft, got:\n%s", s)
	}
	if !strings.Contains(s, "tags: []") {
		t.Errorf("missing tags, got:\n%s", s)
	}
	if !strings.Contains(s, "date: ") {
		t.Errorf("missing date, got:\n%s", s)
	}
}

func TestNewContentWithCustomArchetype(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content")
	archetypesDir := filepath.Join(tmpDir, "archetypes")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}
	if err := os.MkdirAll(archetypesDir, 0755); err != nil {
		t.Fatalf("failed to create archetypes dir: %v", err)
	}

	archetype := `---
title: "{{.Title}}"
section: {{.Section}}
slug: {{.Slug}}
custom: true
---
`
	if err := os.WriteFile(filepath.Join(archetypesDir, "docs.md"), []byte(archetype), 0644); err != nil {
		t.Fatalf("failed to write archetype: %v", err)
	}

	opts := NewContentOptions{
		ContentDir:    contentDir,
		ArchetypesDir: archetypesDir,
		Path:          "docs/getting-started",
	}

	outputPath, err := NewContent(opts)
	if err != nil {
		t.Fatalf("NewContent failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, `title: "Getting Started"`) {
		t.Errorf("missing title, got:\n%s", s)
	}
	if !strings.Contains(s, "section: docs") {
		t.Errorf("missing section, got:\n%s", s)
	}
	if !strings.Contains(s, "custom: true") {
		t.Errorf("missing custom field, got:\n%s", s)
	}
}

func TestNewContentCreatesDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content")

	opts := NewContentOptions{
		ContentDir:    contentDir,
		ArchetypesDir: filepath.Join(tmpDir, "archetypes"),
		Path:          "docs/guide/intro",
	}

	outputPath, err := NewContent(opts)
	if err != nil {
		t.Fatalf("NewContent failed: %v", err)
	}

	expected := filepath.Join(contentDir, "docs", "guide", "intro.md")
	if outputPath != expected {
		t.Errorf("outputPath = %q, want %q", outputPath, expected)
	}
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("file not created at %s", outputPath)
	}
}

func TestNewContentRefusesOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(contentDir, "about.md"), []byte("existing"), 0644); err != nil {
		t.Fatalf("failed to write existing file: %v", err)
	}

	opts := NewContentOptions{
		ContentDir:    contentDir,
		ArchetypesDir: filepath.Join(tmpDir, "archetypes"),
		Path:          "about",
	}

	_, err := NewContent(opts)
	if err == nil {
		t.Fatal("expected error for existing file, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewContentDefaultArchetypeFallback(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content")
	archetypesDir := filepath.Join(tmpDir, "archetypes")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}
	if err := os.MkdirAll(archetypesDir, 0755); err != nil {
		t.Fatalf("failed to create archetypes dir: %v", err)
	}

	// Create only default.md — no section-specific archetype
	defaultArchetype := `---
title: "{{.Title}}"
layout: "default"
---
`
	if err := os.WriteFile(filepath.Join(archetypesDir, "default.md"), []byte(defaultArchetype), 0644); err != nil {
		t.Fatalf("failed to write default archetype: %v", err)
	}

	opts := NewContentOptions{
		ContentDir:    contentDir,
		ArchetypesDir: archetypesDir,
		Path:          "services",
	}

	outputPath, err := NewContent(opts)
	if err != nil {
		t.Fatalf("NewContent failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, `layout: "default"`) {
		t.Errorf("should use default archetype, got:\n%s", s)
	}
}
