// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package site

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFilterPagesDrafts(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content", "blog")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	files := map[string]string{
		"published.md": `---
title: "Published Post"
date: 2024-01-15
draft: false
---
Content`,
		"draft.md": `---
title: "Draft Post"
date: 2024-01-10
draft: true
---
Content`,
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(contentDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	// Test without drafts enabled
	config := DefaultConfig()
	config.Build.Drafts = false

	site := NewWithConfig(config)
	site.SetBaseDir(tmpDir)

	ctx := context.Background()
	if err := site.ProcessContent(ctx); err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}

	if len(site.Pages) != 1 {
		t.Errorf("expected 1 page without drafts, got %d", len(site.Pages))
	}
	if site.Pages[0].Title != "Published Post" {
		t.Errorf("expected 'Published Post', got '%s'", site.Pages[0].Title)
	}

	// Test with drafts enabled
	config2 := DefaultConfig()
	config2.Build.Drafts = true

	site2 := NewWithConfig(config2)
	site2.SetBaseDir(tmpDir)

	if err := site2.ProcessContent(ctx); err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}

	if len(site2.Pages) != 2 {
		t.Errorf("expected 2 pages with drafts enabled, got %d", len(site2.Pages))
	}
}

func TestFilterPagesFuture(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content", "blog")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	// Create a post with a future date
	futureDate := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	pastDate := time.Now().Add(-24 * time.Hour).Format("2006-01-02")

	files := map[string]string{
		"past.md": `---
title: "Past Post"
date: ` + pastDate + `
---
Content`,
		"future.md": `---
title: "Future Post"
date: ` + futureDate + `
---
Content`,
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(contentDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	// Test without future enabled
	config := DefaultConfig()
	config.Build.Future = false

	site := NewWithConfig(config)
	site.SetBaseDir(tmpDir)

	ctx := context.Background()
	if err := site.ProcessContent(ctx); err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}

	if len(site.Pages) != 1 {
		t.Errorf("expected 1 page without future, got %d", len(site.Pages))
	}
	if site.Pages[0].Title != "Past Post" {
		t.Errorf("expected 'Past Post', got '%s'", site.Pages[0].Title)
	}

	// Test with future enabled
	config2 := DefaultConfig()
	config2.Build.Future = true

	site2 := NewWithConfig(config2)
	site2.SetBaseDir(tmpDir)

	if err := site2.ProcessContent(ctx); err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}

	if len(site2.Pages) != 2 {
		t.Errorf("expected 2 pages with future enabled, got %d", len(site2.Pages))
	}
}

func TestFilterPagesBothOptions(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content", "blog")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	futureDate := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	files := map[string]string{
		"published.md": `---
title: "Published"
date: 2024-01-15
---
Content`,
		"draft.md": `---
title: "Draft"
draft: true
---
Content`,
		"future.md": `---
title: "Future"
date: ` + futureDate + `
---
Content`,
		"draft-future.md": `---
title: "Draft Future"
date: ` + futureDate + `
draft: true
---
Content`,
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(contentDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	ctx := context.Background()

	// Test with both disabled (default)
	config := DefaultConfig()
	site := NewWithConfig(config)
	site.SetBaseDir(tmpDir)

	if err := site.ProcessContent(ctx); err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}

	if len(site.Pages) != 1 {
		t.Errorf("expected 1 page with both disabled, got %d", len(site.Pages))
	}

	// Test with both enabled
	config2 := DefaultConfig()
	config2.Build.Drafts = true
	config2.Build.Future = true

	site2 := NewWithConfig(config2)
	site2.SetBaseDir(tmpDir)

	if err := site2.ProcessContent(ctx); err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}

	if len(site2.Pages) != 4 {
		t.Errorf("expected 4 pages with both enabled, got %d", len(site2.Pages))
	}
}

func TestBuildConfigDefaults(t *testing.T) {
	config := DefaultConfig()

	if config.Build.Drafts {
		t.Error("expected Build.Drafts to default to false")
	}
	if config.Build.Future {
		t.Error("expected Build.Future to default to false")
	}
}
