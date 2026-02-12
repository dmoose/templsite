// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package site

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// loadData reads all YAML/JSON files from the data directory
func (s *Site) loadData() error {
	s.Data = make(map[string]any)

	dataDir := s.Config.DataPath(s.baseDir)

	// Check if data directory exists
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		return nil // No data directory is fine
	}

	err := filepath.WalkDir(dataDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" && ext != ".json" {
			return nil
		}

		// Get relative path from data dir for nested keys
		relPath, err := filepath.Rel(dataDir, path)
		if err != nil {
			return err
		}

		// Convert path to key (e.g., "team.yaml" -> "team", "nested/config.yaml" -> "nested/config")
		key := strings.TrimSuffix(relPath, filepath.Ext(relPath))
		key = filepath.ToSlash(key) // Normalize to forward slashes

		data, err := loadDataFile(path)
		if err != nil {
			return fmt.Errorf("loading data file %s: %w", path, err)
		}

		s.Data[key] = data
		return nil
	})

	return err
}

// loadDataFile reads and parses a single data file
func loadDataFile(path string) (any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(path))

	var result any

	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("parsing JSON: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("parsing YAML: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported data file format: %s", ext)
	}

	return result, nil
}

// GetData returns data by key, supporting nested paths with "/"
// For example: GetData("team") or GetData("config/settings")
func (s *Site) GetData(key string) any {
	if s.Data == nil {
		return nil
	}
	return s.Data[key]
}

// GetDataAs attempts to return data cast to a specific type
// Returns nil if not found or type assertion fails
func GetDataAs[T any](s *Site, key string) T {
	var zero T
	data := s.GetData(key)
	if data == nil {
		return zero
	}
	if typed, ok := data.(T); ok {
		return typed
	}
	return zero
}
