package assets

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyStatic(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "assets")
	outputDir := filepath.Join(tmpDir, "public", "assets")

	// Create input directory structure with static files
	imagesDir := filepath.Join(inputDir, "images")
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		t.Fatalf("failed to create images dir: %v", err)
	}

	fontsDir := filepath.Join(inputDir, "fonts")
	if err := os.MkdirAll(fontsDir, 0755); err != nil {
		t.Fatalf("failed to create fonts dir: %v", err)
	}

	// Create some static files
	imageFile := filepath.Join(imagesDir, "logo.png")
	if err := os.WriteFile(imageFile, []byte("fake png data"), 0644); err != nil {
		t.Fatalf("failed to write image file: %v", err)
	}

	fontFile := filepath.Join(fontsDir, "font.woff2")
	if err := os.WriteFile(fontFile, []byte("fake font data"), 0644); err != nil {
		t.Fatalf("failed to write font file: %v", err)
	}

	// Create CSS and JS directories (should be skipped)
	cssDir := filepath.Join(inputDir, "css")
	if err := os.MkdirAll(cssDir, 0755); err != nil {
		t.Fatalf("failed to create css dir: %v", err)
	}
	cssFile := filepath.Join(cssDir, "app.css")
	if err := os.WriteFile(cssFile, []byte("body { color: red; }"), 0644); err != nil {
		t.Fatalf("failed to write css file: %v", err)
	}

	config := &Config{
		InputDir:  inputDir,
		OutputDir: outputDir,
	}

	pipeline := New(config)
	ctx := context.Background()

	err := pipeline.copyStatic(ctx)
	if err != nil {
		t.Fatalf("copyStatic failed: %v", err)
	}

	// Verify image was copied
	outputImage := filepath.Join(outputDir, "images", "logo.png")
	if _, err := os.Stat(outputImage); os.IsNotExist(err) {
		t.Error("expected image to be copied")
	}

	// Verify font was copied
	outputFont := filepath.Join(outputDir, "fonts", "font.woff2")
	if _, err := os.Stat(outputFont); os.IsNotExist(err) {
		t.Error("expected font to be copied")
	}

	// Verify CSS was NOT copied
	outputCSS := filepath.Join(outputDir, "css", "app.css")
	if _, err := os.Stat(outputCSS); !os.IsNotExist(err) {
		t.Error("expected CSS file not to be copied by copyStatic")
	}
}

func TestCopyStaticSkipsWhenNoInput(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "assets")
	outputDir := filepath.Join(tmpDir, "public", "assets")

	// Don't create input directory
	config := &Config{
		InputDir:  inputDir,
		OutputDir: outputDir,
	}

	pipeline := New(config)
	ctx := context.Background()

	// copyStatic should succeed and skip when no input directory exists
	err := pipeline.copyStatic(ctx)
	if err != nil {
		t.Errorf("copyStatic should succeed when no input directory: %v", err)
	}
}

func TestCopyStaticWithNestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "assets")
	outputDir := filepath.Join(tmpDir, "public", "assets")

	// Create nested directory structure
	nestedDir := filepath.Join(inputDir, "images", "icons", "social")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}

	// Create file in nested directory
	nestedFile := filepath.Join(nestedDir, "twitter.svg")
	if err := os.WriteFile(nestedFile, []byte("<svg></svg>"), 0644); err != nil {
		t.Fatalf("failed to write nested file: %v", err)
	}

	config := &Config{
		InputDir:  inputDir,
		OutputDir: outputDir,
	}

	pipeline := New(config)
	ctx := context.Background()

	err := pipeline.copyStatic(ctx)
	if err != nil {
		t.Fatalf("copyStatic failed: %v", err)
	}

	// Verify nested file was copied with correct structure
	outputFile := filepath.Join(outputDir, "images", "icons", "social", "twitter.svg")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("expected nested file to be copied")
	}

	// Verify content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if string(content) != "<svg></svg>" {
		t.Errorf("expected content '<svg></svg>', got '%s'", string(content))
	}
}

func TestCopyStaticWithContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "assets")
	outputDir := filepath.Join(tmpDir, "public", "assets")

	// Create input directory with files
	imagesDir := filepath.Join(inputDir, "images")
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		t.Fatalf("failed to create images dir: %v", err)
	}

	for i := 0; i < 5; i++ {
		imageFile := filepath.Join(imagesDir, filepath.Join("image"+string(rune('0'+i))+".png"))
		if err := os.WriteFile(imageFile, []byte("fake image"), 0644); err != nil {
			t.Fatalf("failed to write image file: %v", err)
		}
	}

	config := &Config{
		InputDir:  inputDir,
		OutputDir: outputDir,
	}

	pipeline := New(config)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should either complete quickly or return context error
	err := pipeline.copyStatic(ctx)
	if err != nil && err != context.Canceled {
		t.Logf("copyStatic with cancelled context: %v", err)
	}
}

func TestCopyStaticSkipsCSSAndJS(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "assets")
	outputDir := filepath.Join(tmpDir, "public", "assets")

	// Create CSS directory
	cssDir := filepath.Join(inputDir, "css")
	if err := os.MkdirAll(cssDir, 0755); err != nil {
		t.Fatalf("failed to create css dir: %v", err)
	}
	cssFile := filepath.Join(cssDir, "app.css")
	if err := os.WriteFile(cssFile, []byte("body {}"), 0644); err != nil {
		t.Fatalf("failed to write css file: %v", err)
	}

	// Create JS directory
	jsDir := filepath.Join(inputDir, "js")
	if err := os.MkdirAll(jsDir, 0755); err != nil {
		t.Fatalf("failed to create js dir: %v", err)
	}
	jsFile := filepath.Join(jsDir, "app.js")
	if err := os.WriteFile(jsFile, []byte("console.log('test');"), 0644); err != nil {
		t.Fatalf("failed to write js file: %v", err)
	}

	// Create images directory (should be copied)
	imagesDir := filepath.Join(inputDir, "images")
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		t.Fatalf("failed to create images dir: %v", err)
	}
	imageFile := filepath.Join(imagesDir, "test.png")
	if err := os.WriteFile(imageFile, []byte("image"), 0644); err != nil {
		t.Fatalf("failed to write image file: %v", err)
	}

	config := &Config{
		InputDir:  inputDir,
		OutputDir: outputDir,
	}

	pipeline := New(config)
	ctx := context.Background()

	err := pipeline.copyStatic(ctx)
	if err != nil {
		t.Fatalf("copyStatic failed: %v", err)
	}

	// Verify CSS was not copied
	outputCSS := filepath.Join(outputDir, "css", "app.css")
	if _, err := os.Stat(outputCSS); !os.IsNotExist(err) {
		t.Error("expected CSS not to be copied")
	}

	// Verify JS was not copied
	outputJS := filepath.Join(outputDir, "js", "app.js")
	if _, err := os.Stat(outputJS); !os.IsNotExist(err) {
		t.Error("expected JS not to be copied")
	}

	// Verify image was copied
	outputImage := filepath.Join(outputDir, "images", "test.png")
	if _, err := os.Stat(outputImage); os.IsNotExist(err) {
		t.Error("expected image to be copied")
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file
	srcFile := filepath.Join(tmpDir, "source.txt")
	srcContent := "test content"
	if err := os.WriteFile(srcFile, []byte(srcContent), 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	// Copy to destination
	dstFile := filepath.Join(tmpDir, "subdir", "dest.txt")
	err := copyFile(srcFile, dstFile)
	if err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	// Verify destination exists
	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Error("expected destination file to exist")
	}

	// Verify content matches
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}

	if string(dstContent) != srcContent {
		t.Errorf("expected content '%s', got '%s'", srcContent, string(dstContent))
	}
}

func TestCopyFilePreservesPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file with specific permissions
	srcFile := filepath.Join(tmpDir, "source.txt")
	if err := os.WriteFile(srcFile, []byte("test"), 0755); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	// Copy file
	dstFile := filepath.Join(tmpDir, "dest.txt")
	err := copyFile(srcFile, dstFile)
	if err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	// Check permissions
	srcInfo, err := os.Stat(srcFile)
	if err != nil {
		t.Fatalf("failed to stat source file: %v", err)
	}

	dstInfo, err := os.Stat(dstFile)
	if err != nil {
		t.Fatalf("failed to stat destination file: %v", err)
	}

	if srcInfo.Mode() != dstInfo.Mode() {
		t.Errorf("expected permissions %v, got %v", srcInfo.Mode(), dstInfo.Mode())
	}
}

func TestCopyStaticWithMultipleFileTypes(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "assets")
	outputDir := filepath.Join(tmpDir, "public", "assets")

	// Create various file types
	files := map[string]string{
		"images/logo.png":     "png data",
		"images/hero.jpg":     "jpg data",
		"fonts/regular.woff2": "font data",
		"fonts/bold.ttf":      "ttf data",
		"files/document.pdf":  "pdf data",
		"videos/demo.mp4":     "video data",
	}

	for path, content := range files {
		fullPath := filepath.Join(inputDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("failed to create dir for %s: %v", path, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", path, err)
		}
	}

	config := &Config{
		InputDir:  inputDir,
		OutputDir: outputDir,
	}

	pipeline := New(config)
	ctx := context.Background()

	err := pipeline.copyStatic(ctx)
	if err != nil {
		t.Fatalf("copyStatic failed: %v", err)
	}

	// Verify all files were copied
	for path := range files {
		outputPath := filepath.Join(outputDir, path)
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Errorf("expected %s to be copied", path)
		}
	}
}
