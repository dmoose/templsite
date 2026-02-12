// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package templates

import (
	"embed"
	"fmt"
	"io/fs"
	"slices"
)

//go:embed tailwind tailwind/.gitignore tailwind/archetypes fastatic fastatic/.gitignore fastatic/archetypes
var Templates embed.FS

// GetTemplate returns the embedded filesystem for a specific template
func GetTemplate(name string) (fs.FS, error) {
	// Check if template exists
	if !templateExists(name) {
		return nil, fmt.Errorf("template %q not found", name)
	}

	// Return a sub-filesystem rooted at the template directory
	subFS, err := fs.Sub(Templates, name)
	if err != nil {
		return nil, fmt.Errorf("accessing template %q: %w", name, err)
	}

	return subFS, nil
}

// ListTemplates returns a list of available template names
func ListTemplates() []string {
	return []string{"tailwind", "fastatic"}
}

// templateExists checks if a template name exists
func templateExists(name string) bool {
	templates := ListTemplates()
	return slices.Contains(templates, name)
}
