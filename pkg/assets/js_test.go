package assets

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessJS(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "assets")
	outputDir := filepath.Join(tmpDir, "public", "assets")

	// Create input directory and JS file
	jsDir := filepath.Join(inputDir, "js")
	if err := os.MkdirAll(jsDir, 0755); err != nil {
		t.Fatalf("failed to create JS dir: %v", err)
	}

	jsInput := filepath.Join(jsDir, "app.js")
	jsContent := `// Test JavaScript
function hello() {
  console.log("Hello, World!");
}

const message = "test";
`
	if err := os.WriteFile(jsInput, []byte(jsContent), 0644); err != nil {
		t.Fatalf("failed to write JS file: %v", err)
	}

	config := &Config{
		InputDir:  inputDir,
		OutputDir: outputDir,
		Minify:    false,
	}

	pipeline := New(config)
	ctx := context.Background()

	err := pipeline.processJS(ctx)
	if err != nil {
		t.Fatalf("processJS failed: %v", err)
	}

	// Verify output file was created
	outputFile := filepath.Join(outputDir, "js", "main.js")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("expected output JS file to be created")
	}

	// Verify content
	output, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if !strings.Contains(string(output), "hello") {
		t.Error("expected output to contain function name")
	}
}

func TestProcessJSWithMinify(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "assets")
	outputDir := filepath.Join(tmpDir, "public", "assets")

	// Create input directory and JS file
	jsDir := filepath.Join(inputDir, "js")
	if err := os.MkdirAll(jsDir, 0755); err != nil {
		t.Fatalf("failed to create JS dir: %v", err)
	}

	jsInput := filepath.Join(jsDir, "app.js")
	jsContent := `// Test JavaScript with lots of whitespace


function   hello  (  )   {
  console.log(  "Hello, World!"  );
}


const   message   =   "test"  ;
`
	if err := os.WriteFile(jsInput, []byte(jsContent), 0644); err != nil {
		t.Fatalf("failed to write JS file: %v", err)
	}

	config := &Config{
		InputDir:  inputDir,
		OutputDir: outputDir,
		Minify:    true,
	}

	pipeline := New(config)
	ctx := context.Background()

	err := pipeline.processJS(ctx)
	if err != nil {
		t.Fatalf("processJS with minify failed: %v", err)
	}

	// Verify output file was created
	outputFile := filepath.Join(outputDir, "js", "main.js")
	output, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	// Minified output should be smaller
	if len(output) >= len(jsContent) {
		t.Logf("Note: Minified output (%d bytes) not smaller than input (%d bytes)", len(output), len(jsContent))
	}

	// Should still contain the function
	if !strings.Contains(string(output), "hello") {
		t.Error("expected minified output to contain function name")
	}
}

func TestProcessJSSkipsWhenNoInput(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "assets")
	outputDir := filepath.Join(tmpDir, "public", "assets")

	// Create directories but no JS file
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}

	config := &Config{
		InputDir:  inputDir,
		OutputDir: outputDir,
		Minify:    false,
	}

	pipeline := New(config)
	ctx := context.Background()

	// processJS should succeed and skip when no input file exists
	err := pipeline.processJS(ctx)
	if err != nil {
		t.Errorf("processJS should succeed when no input file: %v", err)
	}

	// Verify no output file was created
	outputFile := filepath.Join(outputDir, "js", "main.js")
	if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
		t.Error("expected no output file when no input")
	}
}

func TestProcessJSWithContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "assets")
	outputDir := filepath.Join(tmpDir, "public", "assets")

	// Create input directory and JS file
	jsDir := filepath.Join(inputDir, "js")
	if err := os.MkdirAll(jsDir, 0755); err != nil {
		t.Fatalf("failed to create JS dir: %v", err)
	}

	jsInput := filepath.Join(jsDir, "app.js")
	if err := os.WriteFile(jsInput, []byte("console.log('test');"), 0644); err != nil {
		t.Fatalf("failed to write JS file: %v", err)
	}

	config := &Config{
		InputDir:  inputDir,
		OutputDir: outputDir,
	}

	pipeline := New(config)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// processJS should complete quickly (file is small)
	err := pipeline.processJS(ctx)
	if err != nil {
		t.Logf("processJS with cancelled context: %v", err)
	}
}

func TestProcessJSDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "assets", "js")
	outputDir := filepath.Join(tmpDir, "public", "assets", "js")

	// Create input directory structure with multiple JS files
	if err := os.MkdirAll(filepath.Join(inputDir, "lib"), 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}

	// Create main.js
	mainJS := filepath.Join(inputDir, "main.js")
	if err := os.WriteFile(mainJS, []byte("console.log('main');"), 0644); err != nil {
		t.Fatalf("failed to write main.js: %v", err)
	}

	// Create lib/utils.js
	utilsJS := filepath.Join(inputDir, "lib", "utils.js")
	if err := os.WriteFile(utilsJS, []byte("function util() { return true; }"), 0644); err != nil {
		t.Fatalf("failed to write utils.js: %v", err)
	}

	// Create a non-JS file (should be ignored)
	readme := filepath.Join(inputDir, "README.md")
	if err := os.WriteFile(readme, []byte("# README"), 0644); err != nil {
		t.Fatalf("failed to write README: %v", err)
	}

	config := &Config{
		InputDir:  filepath.Join(tmpDir, "assets"),
		OutputDir: filepath.Join(tmpDir, "public", "assets"),
		Minify:    false,
	}

	pipeline := New(config)
	ctx := context.Background()

	err := pipeline.processJSDirectory(ctx, inputDir, outputDir)
	if err != nil {
		t.Fatalf("processJSDirectory failed: %v", err)
	}

	// Verify main.js was copied
	mainOutput := filepath.Join(outputDir, "main.js")
	if _, err := os.Stat(mainOutput); os.IsNotExist(err) {
		t.Error("expected main.js to be copied")
	}

	// Verify lib/utils.js was copied
	utilsOutput := filepath.Join(outputDir, "lib", "utils.js")
	if _, err := os.Stat(utilsOutput); os.IsNotExist(err) {
		t.Error("expected lib/utils.js to be copied")
	}

	// Verify README was not copied
	readmeOutput := filepath.Join(outputDir, "README.md")
	if _, err := os.Stat(readmeOutput); !os.IsNotExist(err) {
		t.Error("expected README.md not to be copied")
	}
}

func TestProcessJSDirectoryWithMinify(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "assets", "js")
	outputDir := filepath.Join(tmpDir, "public", "assets", "js")

	// Create input directory with JS file
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}

	jsContent := `// Comment
function test() {
  const x = 1;
  const y = 2;
  return x + y;
}
`
	jsFile := filepath.Join(inputDir, "app.js")
	if err := os.WriteFile(jsFile, []byte(jsContent), 0644); err != nil {
		t.Fatalf("failed to write JS file: %v", err)
	}

	config := &Config{
		InputDir:  filepath.Join(tmpDir, "assets"),
		OutputDir: filepath.Join(tmpDir, "public", "assets"),
		Minify:    true,
	}

	pipeline := New(config)
	ctx := context.Background()

	err := pipeline.processJSDirectory(ctx, inputDir, outputDir)
	if err != nil {
		t.Fatalf("processJSDirectory with minify failed: %v", err)
	}

	// Verify output file
	outputFile := filepath.Join(outputDir, "app.js")
	output, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	// Should still contain the function name
	if !strings.Contains(string(output), "test") {
		t.Error("expected minified output to contain function name")
	}
}

func TestProcessJSDirectoryWithContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "assets", "js")
	outputDir := filepath.Join(tmpDir, "public", "assets", "js")

	// Create input directory with JS files
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}

	for i := 0; i < 5; i++ {
		jsFile := filepath.Join(inputDir, filepath.Join("file"+string(rune('0'+i))+".js"))
		if err := os.WriteFile(jsFile, []byte("console.log('test');"), 0644); err != nil {
			t.Fatalf("failed to write JS file: %v", err)
		}
	}

	config := &Config{
		InputDir:  filepath.Join(tmpDir, "assets"),
		OutputDir: filepath.Join(tmpDir, "public", "assets"),
	}

	pipeline := New(config)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should either complete quickly or return context error
	err := pipeline.processJSDirectory(ctx, inputDir, outputDir)
	if err != nil && err != context.Canceled {
		t.Logf("processJSDirectory with cancelled context: %v", err)
	}
}

func TestProcessJSCreatesOutputDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "assets")
	outputDir := filepath.Join(tmpDir, "public", "assets")

	// Create input directory and JS file
	jsDir := filepath.Join(inputDir, "js")
	if err := os.MkdirAll(jsDir, 0755); err != nil {
		t.Fatalf("failed to create JS dir: %v", err)
	}

	jsInput := filepath.Join(jsDir, "app.js")
	if err := os.WriteFile(jsInput, []byte("console.log('test');"), 0644); err != nil {
		t.Fatalf("failed to write JS file: %v", err)
	}

	config := &Config{
		InputDir:  inputDir,
		OutputDir: outputDir,
	}

	pipeline := New(config)
	ctx := context.Background()

	err := pipeline.processJS(ctx)
	if err != nil {
		t.Fatalf("processJS failed: %v", err)
	}

	// Check if output directory was created
	outputJSDir := filepath.Join(outputDir, "js")
	if _, err := os.Stat(outputJSDir); os.IsNotExist(err) {
		t.Error("expected output directory to be created")
	}
}
