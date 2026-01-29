package watch

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewWatcher(t *testing.T) {
	w, err := New()
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer w.Close()

	if w == nil {
		t.Fatal("watcher is nil")
	}

	if w.watcher == nil {
		t.Fatal("fsnotify watcher is nil")
	}

	if w.events == nil {
		t.Fatal("events channel is nil")
	}

	if w.errors == nil {
		t.Fatal("errors channel is nil")
	}

	if w.debounceDelay != 100*time.Millisecond {
		t.Errorf("debounce delay = %v, want 100ms", w.debounceDelay)
	}
}

func TestWatcherAddRemove(t *testing.T) {
	w, err := New()
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer w.Close()

	tmpDir := t.TempDir()

	// Add directory
	err = w.Add(tmpDir)
	if err != nil {
		t.Fatalf("failed to add directory: %v", err)
	}

	// Remove directory
	err = w.Remove(tmpDir)
	if err != nil {
		t.Fatalf("failed to remove directory: %v", err)
	}
}

func TestWatcherFileChange(t *testing.T) {
	w, err := New()
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer w.Close()

	// Reduce debounce delay for testing
	w.SetDebounceDelay(50 * time.Millisecond)

	tmpDir := t.TempDir()

	// Add directory to watch
	if err := w.Add(tmpDir); err != nil {
		t.Fatalf("failed to add directory: %v", err)
	}

	// Start watcher
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	w.Start(ctx)

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Wait for event
	select {
	case path := <-w.Events():
		t.Logf("received event for: %s", path)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for file change event")
	}
}

func TestWatcherDebouncing(t *testing.T) {
	w, err := New()
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer w.Close()

	// Very short debounce for testing
	w.SetDebounceDelay(100 * time.Millisecond)

	tmpDir := t.TempDir()
	if err := w.Add(tmpDir); err != nil {
		t.Fatalf("failed to add directory: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	w.Start(ctx)

	// Create multiple files quickly
	testFile := filepath.Join(tmpDir, "test.md")
	for i := 0; i < 5; i++ {
		os.WriteFile(testFile, []byte("content"), 0644)
		time.Sleep(10 * time.Millisecond)
	}

	// Should only get one event due to debouncing
	eventCount := 0
	timeout := time.After(500 * time.Millisecond)

loop:
	for {
		select {
		case <-w.Events():
			eventCount++
		case <-timeout:
			break loop
		}
	}

	// Should get 1-2 events, not 5
	if eventCount > 2 {
		t.Errorf("expected 1-2 debounced events, got %d", eventCount)
	}
	if eventCount == 0 {
		t.Error("expected at least 1 event")
	}

	t.Logf("received %d events (debounced from 5 writes)", eventCount)
}

func TestIsRelevantFile(t *testing.T) {
	w, _ := New()
	defer w.Close()

	tests := []struct {
		path     string
		relevant bool
	}{
		// Relevant files
		{"content/index.md", true},
		{"content/blog/post.md", true},
		{"components/layout/base.templ", true},
		{"assets/css/app.css", true},
		{"assets/js/app.js", true},
		{"config.yaml", true},
		{"config.yml", true},
		{"src/main.go", true},

		// Not relevant
		{".git/config", false},
		{".hidden/file.md", false},
		{"public/index.html", false},
		{"dist/bundle.js", false},
		{"build/output.css", false},
		{"_site/page.html", false},
		{"components/layout/base_templ.go", false},
		{"pkg/site/site_test.go", false},
		{".DS_Store", false},
		{".gitignore", false},
		{"README.txt", false},
	}

	for _, tt := range tests {
		result := w.isRelevantFile(tt.path)
		if result != tt.relevant {
			t.Errorf("isRelevantFile(%q) = %v, want %v", tt.path, result, tt.relevant)
		}
	}
}

func TestWatcherContextCancellation(t *testing.T) {
	w, err := New()
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer w.Close()

	tmpDir := t.TempDir()
	if err := w.Add(tmpDir); err != nil {
		t.Fatalf("failed to add directory: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	w.Start(ctx)

	// Cancel context immediately
	cancel()

	// Give it time to stop
	time.Sleep(100 * time.Millisecond)

	// Create a file - should not get an event
	testFile := filepath.Join(tmpDir, "test.md")
	os.WriteFile(testFile, []byte("test"), 0644)

	select {
	case <-w.Events():
		t.Error("received event after context cancellation")
	case <-time.After(200 * time.Millisecond):
		// Good, no event received
	}
}

func TestWatcherMultipleFiles(t *testing.T) {
	w, err := New()
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer w.Close()

	w.SetDebounceDelay(50 * time.Millisecond)

	tmpDir := t.TempDir()
	if err := w.Add(tmpDir); err != nil {
		t.Fatalf("failed to add directory: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	w.Start(ctx)

	// Create different types of files
	files := []string{
		"content.md",
		"styles.css",
		"script.js",
		"config.yaml",
	}

	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", file, err)
		}
	}

	// Should get at least one event
	select {
	case path := <-w.Events():
		t.Logf("received event for: %s", path)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for events")
	}
}

func TestSetDebounceDelay(t *testing.T) {
	w, err := New()
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer w.Close()

	customDelay := 500 * time.Millisecond
	w.SetDebounceDelay(customDelay)

	if w.debounceDelay != customDelay {
		t.Errorf("debounce delay = %v, want %v", w.debounceDelay, customDelay)
	}
}
