// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package site

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadData(t *testing.T) {
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}

	// Create YAML data file
	yamlData := `
- name: Alice
  role: CEO
- name: Bob
  role: CTO
`
	if err := os.WriteFile(filepath.Join(dataDir, "team.yaml"), []byte(yamlData), 0644); err != nil {
		t.Fatalf("failed to write YAML file: %v", err)
	}

	// Create JSON data file
	jsonData := `{"version": "1.0", "features": ["a", "b", "c"]}`
	if err := os.WriteFile(filepath.Join(dataDir, "config.json"), []byte(jsonData), 0644); err != nil {
		t.Fatalf("failed to write JSON file: %v", err)
	}

	config := DefaultConfig()
	site := NewWithConfig(config)
	site.SetBaseDir(tmpDir)

	if err := site.loadData(); err != nil {
		t.Fatalf("loadData failed: %v", err)
	}

	// Check YAML data
	team := site.GetData("team")
	if team == nil {
		t.Fatal("expected 'team' data")
	}

	teamSlice, ok := team.([]any)
	if !ok {
		t.Fatalf("expected team to be []any, got %T", team)
	}
	if len(teamSlice) != 2 {
		t.Errorf("expected 2 team members, got %d", len(teamSlice))
	}

	// Check JSON data
	configData := site.GetData("config")
	if configData == nil {
		t.Fatal("expected 'config' data")
	}

	configMap, ok := configData.(map[string]any)
	if !ok {
		t.Fatalf("expected config to be map[string]any, got %T", configData)
	}
	if configMap["version"] != "1.0" {
		t.Errorf("expected version '1.0', got '%v'", configMap["version"])
	}
}

func TestLoadDataNestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "data", "nested", "deep")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}

	yamlData := `key: value`
	if err := os.WriteFile(filepath.Join(nestedDir, "settings.yaml"), []byte(yamlData), 0644); err != nil {
		t.Fatalf("failed to write YAML file: %v", err)
	}

	config := DefaultConfig()
	site := NewWithConfig(config)
	site.SetBaseDir(tmpDir)

	if err := site.loadData(); err != nil {
		t.Fatalf("loadData failed: %v", err)
	}

	// Nested path should use forward slashes
	data := site.GetData("nested/deep/settings")
	if data == nil {
		t.Fatal("expected nested data at 'nested/deep/settings'")
	}

	dataMap, ok := data.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", data)
	}
	if dataMap["key"] != "value" {
		t.Errorf("expected key 'value', got '%v'", dataMap["key"])
	}
}

func TestLoadDataNoDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// Don't create data directory

	config := DefaultConfig()
	site := NewWithConfig(config)
	site.SetBaseDir(tmpDir)

	// Should not error when data directory doesn't exist
	if err := site.loadData(); err != nil {
		t.Errorf("loadData should not error when data dir missing: %v", err)
	}

	if len(site.Data) != 0 {
		t.Errorf("expected empty data, got %d items", len(site.Data))
	}
}

func TestLoadDataCustomDataDir(t *testing.T) {
	tmpDir := t.TempDir()
	customDir := filepath.Join(tmpDir, "custom_data")
	if err := os.MkdirAll(customDir, 0755); err != nil {
		t.Fatalf("failed to create custom dir: %v", err)
	}

	yamlData := `name: custom`
	if err := os.WriteFile(filepath.Join(customDir, "info.yaml"), []byte(yamlData), 0644); err != nil {
		t.Fatalf("failed to write YAML file: %v", err)
	}

	config := DefaultConfig()
	config.DataDir = "custom_data"
	site := NewWithConfig(config)
	site.SetBaseDir(tmpDir)

	if err := site.loadData(); err != nil {
		t.Fatalf("loadData failed: %v", err)
	}

	data := site.GetData("info")
	if data == nil {
		t.Fatal("expected 'info' data from custom directory")
	}
}

func TestGetDataNotFound(t *testing.T) {
	config := DefaultConfig()
	site := NewWithConfig(config)
	site.Data = make(map[string]any)

	data := site.GetData("nonexistent")
	if data != nil {
		t.Error("expected nil for nonexistent data")
	}
}

func TestGetDataNilData(t *testing.T) {
	config := DefaultConfig()
	site := NewWithConfig(config)
	// Don't initialize Data

	data := site.GetData("anything")
	if data != nil {
		t.Error("expected nil when Data is nil")
	}
}

func TestLoadDataIgnoresOtherFiles(t *testing.T) {
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}

	// Create various file types
	if err := os.WriteFile(filepath.Join(dataDir, "valid.yaml"), []byte("key: value"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "readme.md"), []byte("# Readme"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "script.sh"), []byte("#!/bin/bash"), 0644); err != nil {
		t.Fatal(err)
	}

	config := DefaultConfig()
	site := NewWithConfig(config)
	site.SetBaseDir(tmpDir)

	if err := site.loadData(); err != nil {
		t.Fatalf("loadData failed: %v", err)
	}

	// Only YAML file should be loaded
	if len(site.Data) != 1 {
		t.Errorf("expected 1 data file, got %d", len(site.Data))
	}
	if site.GetData("valid") == nil {
		t.Error("expected 'valid' data")
	}
}
