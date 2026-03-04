package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"git.catapulsion.com/templsite/pkg/site"
)

func TestGetContentType(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"file.html", "text/html; charset=utf-8"},
		{"file.css", "text/css; charset=utf-8"},
		{"file.js", "application/javascript; charset=utf-8"},
		{"file.json", "application/json; charset=utf-8"},
		{"file.png", "image/png"},
		{"file.jpg", "image/jpeg"},
		{"file.jpeg", "image/jpeg"},
		{"file.gif", "image/gif"},
		{"file.svg", "image/svg+xml"},
		{"file.ico", "image/x-icon"},
		{"file.woff", "font/woff"},
		{"file.woff2", "font/woff2"},
		{"file.ttf", "font/ttf"},
		{"file.eot", "application/vnd.ms-fontobject"},
		{"file.pdf", "application/pdf"},
		{"file.xml", "application/xml; charset=utf-8"},
		{"file.txt", "text/plain; charset=utf-8"},
		{"file.unknown", "application/octet-stream"},
		{"file", "application/octet-stream"},
		{"FILE.HTML", "text/html; charset=utf-8"}, // Case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := getContentType(tt.path)
			if result != tt.expected {
				t.Errorf("getContentType(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestLiveReloadScript(t *testing.T) {
	script := LiveReloadScript()

	// Check that it contains key elements
	if script == "" {
		t.Error("LiveReloadScript returned empty string")
	}

	// Check for required script elements
	requiredStrings := []string{
		"<script>",
		"</script>",
		"WebSocket",
		"/_live-reload",
		"reload",
		"window.location.reload()",
	}

	for _, s := range requiredStrings {
		if !contains(script, s) {
			t.Errorf("LiveReloadScript missing required string: %q", s)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestNewLiveReload(t *testing.T) {
	lr := NewLiveReload()

	if lr == nil {
		t.Fatal("NewLiveReload returned nil")
	}

	if lr.connections == nil {
		t.Error("connections map not initialized")
	}

	if lr.broadcast == nil {
		t.Error("broadcast channel not initialized")
	}

	if lr.ConnectionCount() != 0 {
		t.Errorf("initial connection count should be 0, got %d", lr.ConnectionCount())
	}
}

func TestLiveReloadNotifyReload(t *testing.T) {
	lr := NewLiveReload()

	// Start the broadcast loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	lr.Start(ctx)

	// NotifyReload should not block
	done := make(chan bool)
	go func() {
		lr.NotifyReload()
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("NotifyReload blocked")
	}
}

func TestLiveReloadClose(t *testing.T) {
	lr := NewLiveReload()

	// Close should not panic
	lr.Close()

	// After close, connection count should be 0
	if lr.ConnectionCount() != 0 {
		t.Errorf("connection count after close should be 0, got %d", lr.ConnectionCount())
	}
}

func TestNewServer(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal config
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `title: "Test Site"
baseURL: "http://localhost:8080"
content:
  dir: "content"
assets:
  inputDir: "assets"
  outputDir: "assets"
outputDir: "public"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Create required directories
	if err := os.MkdirAll(filepath.Join(tmpDir, "content"), 0755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "assets"), 0755); err != nil {
		t.Fatalf("failed to create assets dir: %v", err)
	}

	// Load site
	s, err := site.New(configPath)
	if err != nil {
		t.Fatalf("failed to create site: %v", err)
	}
	s.SetBaseDir(tmpDir)

	// Create server
	renderFunc := func(ctx context.Context, s *site.Site) error {
		return nil
	}

	srv, err := New(s, "localhost:0", renderFunc)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	if srv == nil {
		t.Fatal("New returned nil server")
	}

	if srv.site != s {
		t.Error("server site not set correctly")
	}

	if srv.addr != "localhost:0" {
		t.Errorf("server addr = %q, want %q", srv.addr, "localhost:0")
	}

	if srv.watcher == nil {
		t.Error("server watcher not initialized")
	}

	if srv.liveReload == nil {
		t.Error("server liveReload not initialized")
	}
}

func TestServerHandleRequest(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `title: "Test Site"
baseURL: "http://localhost:8080"
outputDir: "public"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Create public directory with test files
	publicDir := filepath.Join(tmpDir, "public")
	if err := os.MkdirAll(publicDir, 0755); err != nil {
		t.Fatalf("failed to create public dir: %v", err)
	}

	// Create index.html
	indexContent := "<html><body>Hello World</body></html>"
	if err := os.WriteFile(filepath.Join(publicDir, "index.html"), []byte(indexContent), 0644); err != nil {
		t.Fatalf("failed to write index.html: %v", err)
	}

	// Create about/index.html for clean URL test
	aboutDir := filepath.Join(publicDir, "about")
	if err := os.MkdirAll(aboutDir, 0755); err != nil {
		t.Fatalf("failed to create about dir: %v", err)
	}
	aboutContent := "<html><body>About Page</body></html>"
	if err := os.WriteFile(filepath.Join(aboutDir, "index.html"), []byte(aboutContent), 0644); err != nil {
		t.Fatalf("failed to write about/index.html: %v", err)
	}

	// Create CSS file
	cssDir := filepath.Join(publicDir, "assets", "css")
	if err := os.MkdirAll(cssDir, 0755); err != nil {
		t.Fatalf("failed to create css dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cssDir, "main.css"), []byte("body { color: black; }"), 0644); err != nil {
		t.Fatalf("failed to write main.css: %v", err)
	}

	// Load site
	s, err := site.New(configPath)
	if err != nil {
		t.Fatalf("failed to create site: %v", err)
	}
	s.SetBaseDir(tmpDir)

	// Create server
	renderFunc := func(ctx context.Context, s *site.Site) error {
		return nil
	}

	srv, err := New(s, "localhost:0", renderFunc)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedType   string
		containsBody   string
	}{
		{
			name:           "root path",
			path:           "/",
			expectedStatus: http.StatusOK,
			expectedType:   "text/html; charset=utf-8",
			containsBody:   "Hello World",
		},
		{
			name:           "clean URL",
			path:           "/about/",
			expectedStatus: http.StatusOK,
			expectedType:   "text/html; charset=utf-8",
			containsBody:   "About Page",
		},
		{
			name:           "CSS file",
			path:           "/assets/css/main.css",
			expectedStatus: http.StatusOK,
			expectedType:   "text/css; charset=utf-8",
			containsBody:   "body",
		},
		{
			name:           "not found",
			path:           "/nonexistent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			srv.handleRequest(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.expectedStatus)
			}

			if tt.expectedType != "" {
				contentType := resp.Header.Get("Content-Type")
				if contentType != tt.expectedType {
					t.Errorf("Content-Type = %q, want %q", contentType, tt.expectedType)
				}
			}

			if tt.containsBody != "" {
				body, _ := io.ReadAll(resp.Body)
				if !contains(string(body), tt.containsBody) {
					t.Errorf("body does not contain %q", tt.containsBody)
				}
			}
		})
	}
}

func TestServerLiveReloadInjection(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `title: "Test Site"
baseURL: "http://localhost:8080"
outputDir: "public"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Create public directory with HTML file containing </body>
	publicDir := filepath.Join(tmpDir, "public")
	if err := os.MkdirAll(publicDir, 0755); err != nil {
		t.Fatalf("failed to create public dir: %v", err)
	}
	htmlContent := "<html><body>Test</body></html>"
	if err := os.WriteFile(filepath.Join(publicDir, "index.html"), []byte(htmlContent), 0644); err != nil {
		t.Fatalf("failed to write index.html: %v", err)
	}

	// Load site
	s, err := site.New(configPath)
	if err != nil {
		t.Fatalf("failed to create site: %v", err)
	}
	s.SetBaseDir(tmpDir)

	// Create server
	renderFunc := func(ctx context.Context, s *site.Site) error {
		return nil
	}

	srv, err := New(s, "localhost:0", renderFunc)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Request the page
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	srv.handleRequest(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Check that live reload script was injected
	if !contains(bodyStr, "/_live-reload") {
		t.Error("live reload script not injected into HTML")
	}

	if !contains(bodyStr, "<script>") {
		t.Error("script tag not found in response")
	}
}

func TestServerCacheHeaders(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `title: "Test Site"
baseURL: "http://localhost:8080"
outputDir: "public"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Create public directory with file
	publicDir := filepath.Join(tmpDir, "public")
	if err := os.MkdirAll(publicDir, 0755); err != nil {
		t.Fatalf("failed to create public dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(publicDir, "index.html"), []byte("<html></html>"), 0644); err != nil {
		t.Fatalf("failed to write index.html: %v", err)
	}

	// Load site
	s, err := site.New(configPath)
	if err != nil {
		t.Fatalf("failed to create site: %v", err)
	}
	s.SetBaseDir(tmpDir)

	// Create server
	renderFunc := func(ctx context.Context, s *site.Site) error {
		return nil
	}

	srv, err := New(s, "localhost:0", renderFunc)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	srv.handleRequest(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	// Check cache-control headers
	cacheControl := resp.Header.Get("Cache-Control")
	if cacheControl != "no-cache, no-store, must-revalidate" {
		t.Errorf("Cache-Control = %q, want %q", cacheControl, "no-cache, no-store, must-revalidate")
	}

	pragma := resp.Header.Get("Pragma")
	if pragma != "no-cache" {
		t.Errorf("Pragma = %q, want %q", pragma, "no-cache")
	}

	expires := resp.Header.Get("Expires")
	if expires != "0" {
		t.Errorf("Expires = %q, want %q", expires, "0")
	}
}
