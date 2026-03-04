package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git.catapulsion.com/templsite/internal/watch"
	"git.catapulsion.com/templsite/pkg/site"
)

// Server represents the development server
type Server struct {
	site       *site.Site
	watcher    *watch.Watcher
	liveReload *LiveReload
	addr       string
	server     *http.Server
}

// New creates a new development server
func New(s *site.Site, addr string) (*Server, error) {
	watcher, err := watch.New()
	if err != nil {
		return nil, fmt.Errorf("creating watcher: %w", err)
	}

	return &Server{
		site:       s,
		watcher:    watcher,
		liveReload: NewLiveReload(),
		addr:       addr,
	}, nil
}

// Serve starts the development server
func (s *Server) Serve(ctx context.Context) error {
	// Initial build
	slog.Info("performing initial build")
	if err := s.site.Build(ctx); err != nil {
		return fmt.Errorf("initial build failed: %w", err)
	}

	// Start live reload broadcast loop
	s.liveReload.Start(ctx)

	// Setup file watching
	if err := s.setupWatching(); err != nil {
		return fmt.Errorf("setting up file watching: %w", err)
	}

	// Start file change handler
	go s.handleFileChanges(ctx)

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/_live-reload", s.liveReload.HandleWebSocket)
	mux.HandleFunc("/", s.handleRequest)

	s.server = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	// Start HTTP server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("development server started", "addr", s.addr, "url", fmt.Sprintf("http://%s", s.addr))
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
func (s *Server) setupWatching() error {
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
	s.watcher.Start(context.Background())

	slog.Info("file watching started",
		"content", contentDir,
		"assets", assetsDir,
		"components", componentsDir)

	return nil
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

			// Rebuild the site
			buildStart := time.Now()
			if err := s.site.Build(ctx); err != nil {
				slog.Error("rebuild failed", "error", err)
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

	// Get output directory
	outputDir := s.site.Config.OutputPath(".")

	// Construct file path
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	} else if !strings.Contains(filepath.Base(path), ".") {
		// If no extension, try adding /index.html
		path = filepath.Join(path, "index.html")
	}

	filePath := filepath.Join(outputDir, path)

	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			slog.Debug("file not found", "path", filePath)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		slog.Error("stat error", "path", filePath, "error", err)
		return
	}

	// Serve directory index
	if info.IsDir() {
		indexPath := filepath.Join(filePath, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			filePath = indexPath
		} else {
			http.NotFound(w, r)
			return
		}
	}

	// Read file
	content, err := os.ReadFile(filePath)
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
	w.Write(content)
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
