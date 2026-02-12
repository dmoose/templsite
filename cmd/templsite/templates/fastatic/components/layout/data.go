// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package layout

import (
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
