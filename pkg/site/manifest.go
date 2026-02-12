// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package site

import (
	"encoding/json"
)

// webManifest represents a PWA web manifest
type webManifest struct {
	Name        string         `json:"name"`
	ShortName   string         `json:"short_name"`
	Description string         `json:"description,omitempty"`
	StartURL    string         `json:"start_url"`
	Display     string         `json:"display"`
	ThemeColor  string         `json:"theme_color,omitempty"`
	BgColor     string         `json:"background_color"`
	Icons       []manifestIcon `json:"icons"`
}

// manifestIcon represents an icon entry in the manifest
type manifestIcon struct {
	Src   string `json:"src"`
	Sizes string `json:"sizes,omitempty"`
	Type  string `json:"type"`
}

// Manifest generates a site.webmanifest JSON string from site configuration
func (s *Site) Manifest() string {
	m := webManifest{
		Name:      s.Config.Title,
		ShortName: s.Config.Title,
		StartURL:  "/",
		Display:   "standalone",
		BgColor:   "#ffffff",
		Icons: []manifestIcon{
			{Src: "/favicon.svg", Type: "image/svg+xml"},
			{Src: "/android-chrome-192x192.png", Sizes: "192x192", Type: "image/png"},
			{Src: "/android-chrome-512x512.png", Sizes: "512x512", Type: "image/png"},
		},
	}

	if s.Config.Description != "" {
		m.Description = s.Config.Description
	}

	if s.Config.ThemeColor != "" {
		m.ThemeColor = s.Config.ThemeColor
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(data) + "\n"
}
