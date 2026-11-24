// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package site

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dmoose/templsite/internal/modproxy"
)

// decoyTempl puts a templ on PATH that records any call and fails, so a
// test can tell whether PATH was used. It returns the record's path.
func decoyTempl(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	marker := filepath.Join(dir, "called")
	script := "#!/bin/sh\necho \"$@\" > " + marker + "\nexit 1\n"
	if err := os.WriteFile(filepath.Join(dir, "templ"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return marker
}

// siteTempl is the templ stand-in version the test sites require.
const siteTempl = "v0.1.0"

// templSite writes a module with one component, requiring the templ
// stand-in, with or without templ as a tool. The go command reads only
// the test's own proxy and module cache.
func templSite(t *testing.T, tool bool) string {
	t.Helper()
	modproxy.Serve(t, modproxy.Templ(siteTempl), modproxy.Templ("v0.2.0"))
	dir := t.TempDir()
	mod := "module example.com/site\n\ngo 1.26\n\nrequire " + modproxy.TemplPath + " " + siteTempl + "\n"
	if tool {
		mod += "\ntool " + templTool + "\n"
	}
	modproxy.WriteFiles(t, dir, map[string]string{
		"go.mod":                 mod,
		"components/hello.templ": "package components\n\ntempl Hello() {\n\t<p>hello</p>\n}\n",
		"content/index.md":       "---\ntitle: Home\n---\nHello\n",
	})
	if tool {
		// As a site's own setup does: tidy settles templ's dependencies.
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("go mod tidy: %v\n%s", err, out)
		}
	}
	return dir
}

func siteAt(dir string) *Site {
	s := NewWithConfig(DefaultConfig())
	s.SetBaseDir(dir)
	return s
}

func TestGenerateTemplComponentsUsesTheModulesTempl(t *testing.T) {
	want := siteTempl
	dir := templSite(t, true)
	marker := decoyTempl(t)

	if err := siteAt(dir).GenerateTemplComponents(context.Background()); err != nil {
		t.Fatalf("GenerateTemplComponents: %v", err)
	}
	gen, err := os.ReadFile(filepath.Join(dir, "components", "hello_templ.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(gen), "// templ: version: "+want+"\n") {
		t.Errorf("generated file is not stamped %s:\n%s", want, gen[:min(len(gen), 200)])
	}
	if _, err := os.Stat(marker); err == nil {
		t.Error("templ on PATH was called")
	}
}

func TestGenerateTemplComponentsNeedsTheToolLine(t *testing.T) {
	dir := templSite(t, false)
	marker := decoyTempl(t)

	err := siteAt(dir).GenerateTemplComponents(context.Background())
	if err == nil || !strings.Contains(err.Error(), "go mod edit -tool="+templTool) {
		t.Fatalf("want an error naming the fix, got %v", err)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Error("templ on PATH was called")
	}
}

func TestBuildDoesNotGenerate(t *testing.T) {
	dir := templSite(t, true)
	marker := decoyTempl(t)

	if err := siteAt(dir).Build(context.Background()); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "components", "hello_templ.go")); err == nil {
		t.Error("Build generated templ code")
	}
	if _, err := os.Stat(marker); err == nil {
		t.Error("templ on PATH was called")
	}
}
