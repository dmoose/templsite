// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package site

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// datedFixture is a small site with dated posts across two sections plus an
// undated page, exercising feeds, the sitemap, and taxonomy output.
var datedFixture = map[string]string{
	"content/blog/first.md": `---
title: "First Post"
date: 2024-01-15
description: "The first post"
tags: ["go", "static"]
---
First body.`,
	"content/blog/second.md": `---
title: "Second Post"
date: 2024-06-20
description: "The second post"
tags: ["static"]
---
Second body.`,
	"content/notes/note.md": `---
title: "A Note"
date: 2024-03-02
tags: ["go"]
---
Note body.`,
	"content/about.md": `---
title: "About"
---
About body.`,
}

// undatedFixture is a product site with no dated content at all.
var undatedFixture = map[string]string{
	"content/about.md": `---
title: "About"
---
About body.`,
	"content/pricing.md": `---
title: "Pricing"
---
Pricing body.`,
}

// writeFixture materializes a fixture map under dir.
func writeFixture(t *testing.T, dir string, files map[string]string) {
	t.Helper()

	for name, body := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("creating dir for %s: %v", name, err)
		}
		if err := os.WriteFile(path, []byte(body), 0644); err != nil {
			t.Fatalf("writing %s: %v", name, err)
		}
	}
}

// buildFixture writes the fixture into a fresh directory, builds it, and
// returns the output directory.
func buildFixture(t *testing.T, files map[string]string) string {
	t.Helper()

	baseDir := t.TempDir()
	writeFixture(t, baseDir, files)

	config := DefaultConfig()
	config.BaseURL = "https://example.com"

	site := NewWithConfig(config)
	site.SetBaseDir(baseDir)

	if err := site.Build(context.Background()); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	return site.OutputDir()
}

// collectOutput walks dir and returns slash-separated relative paths mapped to
// file contents.
func collectOutput(t *testing.T, dir string) map[string][]byte {
	t.Helper()

	out := make(map[string][]byte)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = data
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", dir, err)
	}

	return out
}

// assertIdenticalBuilds builds the fixture twice into separate directories and
// byte-compares every emitted file.
func assertIdenticalBuilds(t *testing.T, files map[string]string) map[string][]byte {
	t.Helper()

	firstOut := buildFixture(t, files)

	// Sleep past the one-second resolution of the feed timestamp formats so a
	// build that still read the clock would produce different bytes.
	time.Sleep(1100 * time.Millisecond)

	secondOut := buildFixture(t, files)

	first := collectOutput(t, firstOut)
	second := collectOutput(t, secondOut)

	for name := range first {
		if _, ok := second[name]; !ok {
			t.Errorf("%s present in first build but missing from second", name)
		}
	}
	for name := range second {
		if _, ok := first[name]; !ok {
			t.Errorf("%s present in second build but missing from first", name)
		}
	}

	for name, want := range first {
		got, ok := second[name]
		if !ok {
			continue
		}
		if !bytes.Equal(want, got) {
			t.Errorf("%s differs between builds:\n--- first ---\n%s\n--- second ---\n%s", name, want, got)
		}
	}

	if len(first) == 0 {
		t.Fatal("build produced no output files")
	}

	return first
}

// firstElement returns the text content of the first <name>…</name> element.
func firstElement(t *testing.T, doc, name string) string {
	t.Helper()

	_, rest, found := strings.Cut(doc, "<"+name+">")
	if !found {
		t.Fatalf("no <%s> element found in:\n%s", name, doc)
	}
	text, _, found := strings.Cut(rest, "</"+name+">")
	if !found {
		t.Fatalf("unterminated <%s> element in:\n%s", name, doc)
	}
	return text
}

func TestBuildIsByteReproducibleWithDatedContent(t *testing.T) {
	files := assertIdenticalBuilds(t, datedFixture)

	feed, ok := files["feed.xml"]
	if !ok {
		t.Fatal("expected feed.xml for a site with dated posts")
	}

	// The feed's own <updated> is the first one in the document and must carry
	// the newest post's date rather than a value read off the clock. Dates come
	// out of frontmatter in the local zone, so compare the timestamp without
	// its offset.
	firstUpdated := firstElement(t, string(feed), "updated")
	if !strings.HasPrefix(firstUpdated, "2024-06-20T00:00:00") {
		t.Errorf("feed.xml <updated> should be the newest post date 2024-06-20, got %q; full feed:\n%s", firstUpdated, feed)
	}
	if strings.Contains(string(feed), "About") {
		t.Error("feed.xml should not contain the undated About page")
	}
}

func TestBuildWithoutDatedContentEmitsNoFeed(t *testing.T) {
	files := assertIdenticalBuilds(t, undatedFixture)

	if _, ok := files["feed.xml"]; ok {
		t.Error("expected no feed.xml for a site with no dated pages")
	}

	// The rest of the generated files should still be there.
	for _, name := range []string{"sitemap.xml", "robots.txt", "site.webmanifest", "404.html"} {
		if _, ok := files[name]; !ok {
			t.Errorf("expected %s to still be generated", name)
		}
	}
}

func TestSitemapOmitsLastModForUndatedPages(t *testing.T) {
	config := DefaultConfig()
	config.BaseURL = "https://example.com"

	baseDir := t.TempDir()
	writeFixture(t, baseDir, undatedFixture)

	site := NewWithConfig(config)
	site.SetBaseDir(baseDir)

	if err := site.ProcessContent(context.Background()); err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}

	if sitemap := site.Sitemap(); strings.Contains(sitemap, "<lastmod>") {
		t.Errorf("sitemap should omit lastmod when no page is dated; got:\n%s", sitemap)
	}
}

// TestBuildIsIdenticalAcrossTimezones is the property the UTC date
// change exists to create.
//
// A bare `date: 2024-01-15` in frontmatter names a day, not an instant.
// While it was read in the builder's local zone, the RFC3339 and
// RFC1123Z timestamps in feed output carried that machine's offset —
// so the same repository produced different bytes on a laptop in
// California and a CI runner in UTC, and neither was wrong. Building
// the same fixture under two zones and comparing every emitted file is
// the only assertion that actually pins this.
func TestBuildIsIdenticalAcrossTimezones(t *testing.T) {
	zones := []string{"UTC", "America/Los_Angeles", "Asia/Tokyo"}

	var reference map[string][]byte
	var referenceZone string

	for _, tz := range zones {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			t.Skipf("no tzdata for %s: %v", tz, err)
		}

		out := func() map[string][]byte {
			orig := time.Local
			time.Local = loc
			defer func() { time.Local = orig }()
			return collectOutput(t, buildFixture(t, datedFixture))
		}()

		if reference == nil {
			reference, referenceZone = out, tz
			continue
		}

		if len(out) != len(reference) {
			t.Fatalf("TZ=%s produced %d files, TZ=%s produced %d", tz, len(out), referenceZone, len(reference))
		}
		for name, want := range reference {
			got, ok := out[name]
			if !ok {
				t.Errorf("TZ=%s did not emit %s", tz, name)
				continue
			}
			if !bytes.Equal(got, want) {
				t.Errorf("%s differs between TZ=%s and TZ=%s — the builder's timezone reached the output\n%s",
					name, referenceZone, tz, firstDifferingLine(string(want), string(got)))
			}
		}
	}
}

// firstDifferingLine reports the first line that differs, which for a
// timezone leak is the timestamp that carried the offset.
func firstDifferingLine(a, b string) string {
	al, bl := strings.Split(a, "\n"), strings.Split(b, "\n")
	for i := 0; i < len(al) && i < len(bl); i++ {
		if al[i] != bl[i] {
			return "  first: " + strings.TrimSpace(al[i]) + "\n  other: " + strings.TrimSpace(bl[i])
		}
	}
	return ""
}
