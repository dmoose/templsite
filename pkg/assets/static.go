package assets

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// copyStatic copies static files (images, fonts, etc.) to the output directory
func (p *Pipeline) copyStatic(ctx context.Context) error {
	// Define directories to skip (already processed)
	skipDirs := map[string]bool{
		"css": true,
		"js":  true,
	}

	// Check if input directory exists
	if _, err := os.Stat(p.config.InputDir); os.IsNotExist(err) {
		slog.Debug("no asset input directory found, skipping static files", "path", p.config.InputDir)
		return nil
	}

	slog.Debug("copying static files", "from", p.config.InputDir, "to", p.config.OutputDir)

	// Walk the input directory
	fileCount := 0
	err := filepath.WalkDir(p.config.InputDir, func(path string, d os.DirEntry, err error) error {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == p.config.InputDir {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(p.config.InputDir, path)
		if err != nil {
			return fmt.Errorf("calculating relative path: %w", err)
		}

		// Get the first path component (e.g., "css", "js", "images")
		firstDir := strings.Split(relPath, string(filepath.Separator))[0]

		// Skip directories we've already processed
		if skipDirs[firstDir] {
			if d.IsDir() && filepath.Base(path) == firstDir {
				return filepath.SkipDir
			}
			return nil
		}

		// If it's a directory, create it in the output
		if d.IsDir() {
			outPath := filepath.Join(p.config.OutputDir, relPath)
			if err := os.MkdirAll(outPath, 0755); err != nil {
				return fmt.Errorf("creating directory %s: %w", outPath, err)
			}
			return nil
		}

		// Copy the file
		outPath := filepath.Join(p.config.OutputDir, relPath)
		if err := copyFile(path, outPath); err != nil {
			return fmt.Errorf("copying %s: %w", relPath, err)
		}

		fileCount++
		slog.Debug("copied static file", "from", path, "to", outPath)
		return nil
	})

	if err != nil {
		return err
	}

	if fileCount > 0 {
		slog.Info("copied static files", "count", fileCount)
	}

	return nil
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening source file: %w", err)
	}
	defer srcFile.Close()

	// Get source file info
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("getting source file info: %w", err)
	}

	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("creating destination directory: %w", err)
	}

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("creating destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copying file content: %w", err)
	}

	// Set permissions
	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("setting file permissions: %w", err)
	}

	return nil
}
