package site

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dmoose/templsite/pkg/content"
)

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configYAML := `
title: "Test Site"
baseURL: "https://example.com"
content:
  dir: "content"
  defaultLayout: "page"
assets:
  inputDir: "assets"
  outputDir: "assets"
outputDir: "public"
`

	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	site, err := New(configPath)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if site.Config == nil {
		t.Error("expected site to have config")
	}

	if site.Config.Title != "Test Site" {
		t.Errorf("expected title 'Test Site', got '%s'", site.Config.Title)
	}
}

func TestNewWithNonexistentConfig(t *testing.T) {
	_, err := New("nonexistent.yaml")
	if err == nil {
		t.Error("expected error when loading nonexistent config")
	}
}

func TestNewWithConfig(t *testing.T) {
	config := DefaultConfig()
	config.Title = "Direct Config"

	site := NewWithConfig(config)
	if site.Config == nil {
		t.Error("expected site to have config")
	}

	if site.Config.Title != "Direct Config" {
		t.Errorf("expected title 'Direct Config', got '%s'", site.Config.Title)
	}
}

func TestSetBaseDir(t *testing.T) {
	site := NewWithConfig(DefaultConfig())

	site.SetBaseDir("/test/base")

	if site.baseDir != "/test/base" {
		t.Errorf("expected baseDir '/test/base', got '%s'", site.baseDir)
	}
}

func TestBuild(t *testing.T) {
	tmpDir := t.TempDir()

	// Create content directory
	contentDir := filepath.Join(tmpDir, "content")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	site := NewWithConfig(DefaultConfig())
	site.SetBaseDir(tmpDir)

	ctx := context.Background()
	err := site.Build(ctx)

	// Build should succeed
	if err != nil {
		t.Errorf("Build failed: %v", err)
	}

	// BuildTime should be set
	if site.BuildTime.IsZero() {
		t.Error("expected BuildTime to be set after build")
	}

	// BuildTime should be recent
	elapsed := time.Since(site.BuildTime)
	if elapsed > time.Second {
		t.Errorf("BuildTime seems too old: %v ago", elapsed)
	}
}

func TestBuildWithContext(t *testing.T) {
	tmpDir := t.TempDir()

	// Create content directory
	contentDir := filepath.Join(tmpDir, "content")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	site := NewWithConfig(DefaultConfig())
	site.SetBaseDir(tmpDir)

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Build should fail with context error
	err := site.Build(ctx)
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestBuildSetsTimestamp(t *testing.T) {
	tmpDir := t.TempDir()

	// Create content directory
	contentDir := filepath.Join(tmpDir, "content")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}

	site := NewWithConfig(DefaultConfig())
	site.SetBaseDir(tmpDir)

	before := time.Now()
	ctx := context.Background()

	if err := site.Build(ctx); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	after := time.Now()

	// Verify BuildTime is between before and after
	if site.BuildTime.Before(before) {
		t.Error("BuildTime is before build started")
	}

	if site.BuildTime.After(after) {
		t.Error("BuildTime is after build finished")
	}
}

func TestClean(t *testing.T) {
	site := NewWithConfig(DefaultConfig())

	// Clean should not error (even though it's a stub)
	err := site.Clean()
	if err != nil {
		t.Errorf("Clean failed: %v", err)
	}
}

func TestLinkPrevNext(t *testing.T) {
	tmpDir := t.TempDir()

	// Create content directory with blog posts
	blogDir := filepath.Join(tmpDir, "content", "blog")
	if err := os.MkdirAll(blogDir, 0755); err != nil {
		t.Fatalf("failed to create blog dir: %v", err)
	}

	// Create content directory for root pages
	contentDir := filepath.Join(tmpDir, "content")

	// Create blog posts with different dates
	posts := []struct {
		filename string
		content  string
	}{
		{
			filename: "post1.md",
			content: `---
title: "Post 1 - Oldest"
date: 2025-01-01
---
Content`,
		},
		{
			filename: "post2.md",
			content: `---
title: "Post 2 - Middle"
date: 2025-01-15
---
Content`,
		},
		{
			filename: "post3.md",
			content: `---
title: "Post 3 - Newest"
date: 2025-01-20
---
Content`,
		},
	}

	for _, post := range posts {
		path := filepath.Join(blogDir, post.filename)
		if err := os.WriteFile(path, []byte(post.content), 0644); err != nil {
			t.Fatalf("failed to write post: %v", err)
		}
	}

	// Create a root page (different section)
	rootPage := filepath.Join(contentDir, "about.md")
	if err := os.WriteFile(rootPage, []byte("---\ntitle: About\n---\nAbout"), 0644); err != nil {
		t.Fatalf("failed to write root page: %v", err)
	}

	site := NewWithConfig(DefaultConfig())
	site.SetBaseDir(tmpDir)

	ctx := context.Background()
	if err := site.ProcessContent(ctx); err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}

	// Find blog posts by title
	var post1, post2, post3 *content.Page
	for _, p := range site.Pages {
		switch p.Title {
		case "Post 1 - Oldest":
			post1 = p
		case "Post 2 - Middle":
			post2 = p
		case "Post 3 - Newest":
			post3 = p
		}
	}

	if post1 == nil || post2 == nil || post3 == nil {
		t.Fatal("failed to find all blog posts")
	}

	// Verify all are in blog section
	if post1.Section != "blog" || post2.Section != "blog" || post3.Section != "blog" {
		t.Error("expected all posts to be in 'blog' section")
	}

	// Verify Prev/Next links
	// Posts are sorted newest first: post3 -> post2 -> post1
	// Prev points to newer (earlier in list), Next points to older (later in list)

	// Post 3 (newest) - no Prev, Next = Post 2
	if post3.Prev != nil {
		t.Errorf("newest post should have no Prev, got '%s'", post3.Prev.Title)
	}
	if post3.Next == nil || post3.Next.Title != "Post 2 - Middle" {
		t.Error("newest post's Next should be middle post")
	}

	// Post 2 (middle) - Prev = Post 3, Next = Post 1
	if post2.Prev == nil || post2.Prev.Title != "Post 3 - Newest" {
		t.Error("middle post's Prev should be newest post")
	}
	if post2.Next == nil || post2.Next.Title != "Post 1 - Oldest" {
		t.Error("middle post's Next should be oldest post")
	}

	// Post 1 (oldest) - Prev = Post 2, no Next
	if post1.Prev == nil || post1.Prev.Title != "Post 2 - Middle" {
		t.Error("oldest post's Prev should be middle post")
	}
	if post1.Next != nil {
		t.Errorf("oldest post should have no Next, got '%s'", post1.Next.Title)
	}
}

func TestLinkPrevNextAcrossSections(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two sections
	blogDir := filepath.Join(tmpDir, "content", "blog")
	docsDir := filepath.Join(tmpDir, "content", "docs")
	if err := os.MkdirAll(blogDir, 0755); err != nil {
		t.Fatalf("failed to create blog dir: %v", err)
	}
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("failed to create docs dir: %v", err)
	}

	// Create posts in blog section
	blogPost := filepath.Join(blogDir, "post.md")
	if err := os.WriteFile(blogPost, []byte("---\ntitle: Blog Post\ndate: 2025-01-15\n---\nContent"), 0644); err != nil {
		t.Fatalf("failed to write blog post: %v", err)
	}

	// Create pages in docs section
	docsPage := filepath.Join(docsDir, "intro.md")
	if err := os.WriteFile(docsPage, []byte("---\ntitle: Docs Intro\ndate: 2025-01-10\n---\nContent"), 0644); err != nil {
		t.Fatalf("failed to write docs page: %v", err)
	}

	site := NewWithConfig(DefaultConfig())
	site.SetBaseDir(tmpDir)

	ctx := context.Background()
	if err := site.ProcessContent(ctx); err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}

	// Find pages
	var blogPage, docPage *content.Page
	for _, p := range site.Pages {
		switch p.Title {
		case "Blog Post":
			blogPage = p
		case "Docs Intro":
			docPage = p
		}
	}

	// Verify pages are in different sections
	if blogPage.Section != "blog" || docPage.Section != "docs" {
		t.Error("pages should be in different sections")
	}

	// Verify no cross-section linking
	if blogPage.Prev != nil || blogPage.Next != nil {
		t.Error("blog post should not be linked to docs page")
	}
	if docPage.Prev != nil || docPage.Next != nil {
		t.Error("docs page should not be linked to blog post")
	}
}
