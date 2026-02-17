// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package layout

import (
	"fmt"
	"strings"

	"github.com/dmoose/templsite/pkg/content"
	"github.com/dmoose/templsite/pkg/site"
)

// PageMeta carries all metadata needed by the base layout template.
type PageMeta struct {
	Title        string // "Page Title — Site Name"
	Description  string // meta description
	SiteName     string // og:site_name
	BaseURL      string // for canonical URLs
	URL          string // page URL path (canonical)
	Language     string // html lang attribute (default "en")
	ThemeColor   string // browser/PWA theme color
	AssetVersion string // cache-busting hash
}

func langOrDefault(lang string) string {
	if lang == "" {
		return "en"
	}
	return lang
}

// pageTitle: home page gets "SiteTitle — description"; others get "PageTitle — SiteTitle"
func pageTitle(page *content.Page, s *site.Site) string {
	if page.URL == "/" {
		return s.Config.Title + " — " + s.Config.Description
	}
	return page.Title + " — " + s.Config.Title
}

// paramString reads a value from site config params by key.
func paramString(s *site.Site, key string) string {
	if v, ok := s.Config.Params[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// fmString reads a dot-separated path from page frontmatter.
func fmString(page *content.Page, path string) string {
	parts := strings.Split(path, ".")
	var current any = page.Frontmatter
	for _, key := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current = m[key]
	}
	s, _ := current.(string)
	return s
}
