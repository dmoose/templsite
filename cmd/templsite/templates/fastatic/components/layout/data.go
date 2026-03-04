package layout

// PageMeta carries all metadata needed by the base layout template.
type PageMeta struct {
	Title        string // "Page Title - Site Name"
	Description  string // meta description
	SiteName     string // og:site_name
	BaseURL      string // for canonical URLs
	URL          string // page URL path (canonical)
	Language     string // html lang attribute (default "en")
	ThemeColor   string // browser/PWA theme color
	AssetVersion string // cache-busting hash
}
