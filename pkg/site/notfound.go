// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package site

import "fmt"

// Default404 generates a minimal 404 page as a fallback.
// Users can override this by creating content/404.md or placing 404.html in static/.
func (s *Site) Default404() string {
	title := "Page Not Found"
	siteName := s.Config.Title

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s - %s</title>
  <style>
    body { font-family: system-ui, -apple-system, sans-serif; display: flex; align-items: center; justify-content: center; min-height: 100vh; margin: 0; background: #fafafa; color: #333; }
    .container { text-align: center; padding: 2rem; }
    h1 { font-size: 4rem; margin: 0; color: #999; }
    p { font-size: 1.25rem; margin: 1rem 0; }
    a { color: inherit; text-decoration: underline; }
  </style>
</head>
<body>
  <div class="container">
    <h1>404</h1>
    <p>The page you're looking for doesn't exist.</p>
    <p><a href="/">Back to %s</a></p>
  </div>
</body>
</html>
`, title, siteName, siteName)
}
