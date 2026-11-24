// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dmoose/templsite/internal/watch"
	"github.com/dmoose/templsite/pkg/site"
)

// RenderFunc is a function that renders all pages for a site
// Users provide this to use their own components
type RenderFunc func(ctx context.Context, s *site.Site) error

// Server represents the development server
type Server struct {
	site       *site.Site
	watcher    *watch.Watcher
	liveReload *LiveReload
	addr       string
	server     *http.Server
	renderFunc RenderFunc
}

// New creates a new development server
func New(s *site.Site, addr string, renderFunc RenderFunc) (*Server, error) {
	watcher, err := watch.New()
	if err != nil {
		return nil, fmt.Errorf("creating watcher: %w", err)
	}

	return &Server{
		site:       s,
		watcher:    watcher,
		liveReload: NewLiveReload(),
		addr:       addr,
		renderFunc: renderFunc,
	}, nil
}

// Serve starts the development server
func (s *Server) Serve(ctx context.Context) error {
	// Initial build
	slog.Info("performing initial build")
	if err := s.site.Build(ctx); err != nil {
		return fmt.Errorf("initial build failed: %w", err)
	}

	// Render pages using user's components
	if err := s.renderFunc(ctx, s.site); err != nil {
		return fmt.Errorf("initial render failed: %w", err)
	}

	// Start live reload broadcast loop
	s.liveReload.Start(ctx)

	// Setup file watching
	s.setupWatching(ctx)

	// Start file change handler
	go s.handleFileChanges(ctx)

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/_live-reload", s.liveReload.HandleWebSocket)
	mux.HandleFunc("/", s.handleRequest)

	s.server = &http.Server{
		Addr:              s.addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Start HTTP server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("development server started", "addr", s.addr, "url", "http://"+s.addr)
		fmt.Printf("\n✓ Development server running at http://%s\n", s.addr)
		fmt.Println("  Press Ctrl+C to stop")
		fmt.Println()
		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		slog.Info("shutting down server")
		return s.shutdown()
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	}
}

// setupWatching configures file system watching
func (s *Server) setupWatching(ctx context.Context) {
	// Watch content directory
	contentDir := s.site.Config.ContentPath(".")
	if _, err := os.Stat(contentDir); err == nil {
		if err := s.addDirRecursive(contentDir); err != nil {
			slog.Warn("failed to watch content directory", "dir", contentDir, "error", err)
		}
	}

	// Watch assets directory
	assetsDir := s.site.Config.AssetsInputPath(".")
	if _, err := os.Stat(assetsDir); err == nil {
		if err := s.addDirRecursive(assetsDir); err != nil {
			slog.Warn("failed to watch assets directory", "dir", assetsDir, "error", err)
		}
	}

	// Watch components directory
	componentsDir := "components"
	if _, err := os.Stat(componentsDir); err == nil {
		if err := s.addDirRecursive(componentsDir); err != nil {
			slog.Warn("failed to watch components directory", "dir", componentsDir, "error", err)
		}
	}

	// Watch config file
	if _, err := os.Stat("config.yaml"); err == nil {
		if err := s.watcher.Add("config.yaml"); err != nil {
			slog.Warn("failed to watch config.yaml", "error", err)
		}
	}

	// Start watcher
	s.watcher.Start(ctx)

	slog.Info("file watching started",
		"content", contentDir,
		"assets", assetsDir,
		"components", componentsDir)
}

// addDirRecursive adds a directory and all subdirectories to the watcher
func (s *Server) addDirRecursive(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip hidden directories and output directories
			base := filepath.Base(path)
			if strings.HasPrefix(base, ".") ||
				base == "public" ||
				base == "dist" ||
				base == "build" ||
				base == "_site" {
				return filepath.SkipDir
			}
			return s.watcher.Add(path)
		}
		return nil
	})
}

// handleFileChanges watches for file changes and triggers rebuilds
func (s *Server) handleFileChanges(ctx context.Context) {
	// Prevent rapid rebuilds
	var lastBuild time.Time
	minRebuildInterval := 500 * time.Millisecond

	for {
		select {
		case <-ctx.Done():
			return

		case path := <-s.watcher.Events():
			// Check if enough time has passed since last build
			if time.Since(lastBuild) < minRebuildInterval {
				slog.Debug("skipping rebuild, too soon since last build")
				continue
			}

			slog.Info("file changed, rebuilding", "file", path)
			lastBuild = time.Now()
			if strings.HasSuffix(path, ".templ") {
				slog.Warn("components changed: the running site binary still has the old ones; run make serve again", "file", path)
			}

			// Rebuild the site
			buildStart := time.Now()
			if err := s.site.Build(ctx); err != nil {
				slog.Error("rebuild failed", "error", err)
				continue
			}

			// Render pages using user's components
			if err := s.renderFunc(ctx, s.site); err != nil {
				slog.Error("render failed", "error", err)
				continue
			}

			buildDuration := time.Since(buildStart)

			slog.Info("rebuild complete", "duration", buildDuration)

			// Notify browsers to reload
			s.liveReload.NotifyReload()

		case err := <-s.watcher.Errors():
			slog.Warn("watcher error", "error", err)
		}
	}
}

// handleRequest handles HTTP requests for static files
func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Log request
	slog.Debug("request", "method", r.Method, "path", r.URL.Path)

	// Serve from the output directory only: an os.Root refuses any path,
	// symlink included, that leads outside it.
	root, err := os.OpenRoot(s.site.OutputDir())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		slog.Error("opening output directory", "error", err)
		return
	}
	defer func() { _ = root.Close() }()

	// Construct file path, relative to the output directory
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	} else if !strings.Contains(filepath.Base(path), ".") {
		// If no extension, try adding /index.html
		path = filepath.Join(path, "index.html")
	}
	filePath := strings.TrimPrefix(filepath.Clean(path), "/")

	// Check if file exists
	info, err := root.Stat(filePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			s.serve404(w, r, root)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		slog.Error("stat error", "path", filePath, "error", err)
		return
	}

	// Serve directory index
	if info.IsDir() {
		indexPath := filepath.Join(filePath, "index.html")
		if _, err := root.Stat(indexPath); err != nil {
			s.serve404(w, r, root)
			return
		}
		filePath = indexPath
	}

	// Read file
	content, err := root.ReadFile(filePath)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		slog.Error("read error", "path", filePath, "error", err)
		return
	}

	// Inject live reload script for HTML files
	if strings.HasSuffix(filePath, ".html") {
		contentStr := string(content)
		// Inject before closing </body> tag
		if strings.Contains(contentStr, "</body>") {
			contentStr = strings.Replace(contentStr, "</body>",
				LiveReloadScript()+"\n</body>", 1)
			content = []byte(contentStr)
		}
	}

	// Determine content type
	contentType := getContentType(filePath)
	w.Header().Set("Content-Type", contentType)

	// Disable caching for development
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Write response
	if _, err := w.Write(content); err != nil { //nolint:gosec // dev server returns the site's own built files
		slog.Debug("error writing response", "path", filePath, "error", err)
	}
}

// serve404 serves the custom 404.html page if available, otherwise falls back to default
func (s *Server) serve404(w http.ResponseWriter, r *http.Request, root *os.Root) {
	slog.Debug("file not found", "path", r.URL.Path)

	// Try to serve custom 404.html
	content, err := root.ReadFile("404.html")
	if err != nil {
		// No custom 404 page, use default
		http.NotFound(w, r)
		return
	}

	// Inject live reload script
	contentStr := string(content)
	if strings.Contains(contentStr, "</body>") {
		contentStr = strings.Replace(contentStr, "</body>",
			LiveReloadScript()+"\n</body>", 1)
		content = []byte(contentStr)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write(content)
}

// getContentType returns the content type based on file extension
func getContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	contentTypes := map[string]string{
		".html":  "text/html; charset=utf-8",
		".css":   "text/css; charset=utf-8",
		".js":    "application/javascript; charset=utf-8",
		".json":  "application/json; charset=utf-8",
		".png":   "image/png",
		".jpg":   "image/jpeg",
		".jpeg":  "image/jpeg",
		".gif":   "image/gif",
		".svg":   "image/svg+xml",
		".ico":   "image/x-icon",
		".woff":  "font/woff",
		".woff2": "font/woff2",
		".ttf":   "font/ttf",
		".eot":   "application/vnd.ms-fontobject",
		".pdf":   "application/pdf",
		".xml":   "application/xml; charset=utf-8",
		".txt":   "text/plain; charset=utf-8",
	}

	if contentType, ok := contentTypes[ext]; ok {
		return contentType
	}
	return "application/octet-stream"
}

// shutdown gracefully shuts down the server
func (s *Server) shutdown() error {
	// Stop watcher
	if s.watcher != nil {
		if err := s.watcher.Close(); err != nil {
			slog.Warn("error closing watcher", "error", err)
		}
	}

	// Close live reload connections
	if s.liveReload != nil {
		s.liveReload.Close()
	}

	// Shutdown HTTP server
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("server shutdown error: %w", err)
		}
	}

	slog.Info("server shutdown complete")
	return nil
}
