// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package watch

import (
	"context"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher watches file system changes and notifies on relevant events
type Watcher struct {
	watcher       *fsnotify.Watcher
	events        chan string
	errors        chan error
	debounceDelay time.Duration
}

// New creates a new file watcher
func New() (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		watcher:       fsWatcher,
		events:        make(chan string, 10),
		errors:        make(chan error, 10),
		debounceDelay: 100 * time.Millisecond,
	}, nil
}

// Add adds a path to watch (directory or file)
func (w *Watcher) Add(path string) error {
	slog.Debug("adding path to watcher", "path", path)
	return w.watcher.Add(path)
}

// Remove removes a path from watching
func (w *Watcher) Remove(path string) error {
	slog.Debug("removing path from watcher", "path", path)
	return w.watcher.Remove(path)
}

// Events returns the channel for file change events
func (w *Watcher) Events() <-chan string {
	return w.events
}

// Errors returns the channel for watcher errors
func (w *Watcher) Errors() <-chan error {
	return w.errors
}

// Start begins watching for file changes
func (w *Watcher) Start(ctx context.Context) {
	go w.watchLoop(ctx)
}

// Close stops the watcher and closes channels
func (w *Watcher) Close() error {
	close(w.events)
	close(w.errors)
	return w.watcher.Close()
}

// watchLoop is the main event processing loop with debouncing
func (w *Watcher) watchLoop(ctx context.Context) {
	// Debouncing: collect events and only emit after delay
	var debounceTimer *time.Timer
	pendingEvents := make(map[string]bool)

	flushEvents := func() {
		if len(pendingEvents) > 0 {
			// Just send a generic change signal
			// The path doesn't matter much since we'll rebuild everything
			for path := range pendingEvents {
				select {
				case w.events <- path:
					slog.Debug("file change event emitted", "path", path)
				case <-ctx.Done():
					return
				default:
					// Channel full, skip
				}
				break // Only send one event
			}
			pendingEvents = make(map[string]bool)
		}
	}

	for {
		select {
		case <-ctx.Done():
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			return

		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Filter relevant file changes
			if !w.isRelevantFile(event.Name) {
				continue
			}

			// Ignore certain operations
			if event.Op&fsnotify.Chmod == fsnotify.Chmod {
				continue
			}

			slog.Debug("file system event", "path", event.Name, "op", event.Op.String())

			// Add to pending events
			pendingEvents[event.Name] = true

			// Reset debounce timer
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(w.debounceDelay, flushEvents)

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}

			slog.Warn("watcher error", "error", err)
			select {
			case w.errors <- err:
			case <-ctx.Done():
				return
			default:
				// Error channel full, skip
			}
		}
	}
}

// isRelevantFile checks if a file change is relevant for rebuilding
func (w *Watcher) isRelevantFile(path string) bool {
	// Normalize path
	path = filepath.Clean(path)

	// Get base filename for specific checks
	base := filepath.Base(path)

	// Ignore hidden files and directories (check all path components)
	parts := strings.SplitSeq(path, string(filepath.Separator))
	for part := range parts {
		if strings.HasPrefix(part, ".") && part != "." && part != ".." {
			return false
		}
	}

	// Ignore build output directories
	if strings.Contains(path, "public/") ||
		strings.Contains(path, "dist/") ||
		strings.Contains(path, "build/") ||
		strings.Contains(path, "_site/") {
		return false
	}

	// Ignore generated templ files
	if strings.HasSuffix(path, "_templ.go") {
		return false
	}

	// Ignore Go test files
	if strings.HasSuffix(path, "_test.go") {
		return false
	}

	// Check file extensions we care about
	ext := strings.ToLower(filepath.Ext(path))
	relevantExts := map[string]bool{
		".md":    true, // Markdown content
		".templ": true, // templ components
		".css":   true, // Stylesheets
		".js":    true, // JavaScript
		".yaml":  true, // Config files
		".yml":   true, // Config files
		".html":  true, // HTML templates (if any)
		".go":    true, // Go files (for templ components)
	}

	// Also watch config.yaml specifically
	if base == "config.yaml" || base == "config.yml" {
		return true
	}

	return relevantExts[ext]
}

// SetDebounceDelay sets the debounce delay (mainly for testing)
func (w *Watcher) SetDebounceDelay(d time.Duration) {
	w.debounceDelay = d
}
