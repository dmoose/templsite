package site

import (
	"testing"
)

func TestBuildMenus(t *testing.T) {
	config := DefaultConfig()
	config.Menus = map[string][]MenuItemConfig{
		"main": {
			{Name: "Home", URL: "/", Weight: 1},
			{Name: "Blog", URL: "/blog/", Weight: 2},
			{Name: "About", URL: "/about/", Weight: 3},
		},
		"footer": {
			{Name: "Privacy", URL: "/privacy/", Weight: 1},
			{Name: "Terms", URL: "/terms/", Weight: 2},
		},
	}

	site := NewWithConfig(config)
	site.buildMenus()

	// Check main menu
	mainMenu := site.Menu("main")
	if mainMenu == nil {
		t.Fatal("expected 'main' menu")
	}
	if len(mainMenu) != 3 {
		t.Errorf("expected 3 items in main menu, got %d", len(mainMenu))
	}

	// Check items are sorted by weight
	if mainMenu[0].Name != "Home" {
		t.Errorf("expected 'Home' first, got '%s'", mainMenu[0].Name)
	}
	if mainMenu[1].Name != "Blog" {
		t.Errorf("expected 'Blog' second, got '%s'", mainMenu[1].Name)
	}
	if mainMenu[2].Name != "About" {
		t.Errorf("expected 'About' third, got '%s'", mainMenu[2].Name)
	}

	// Check footer menu
	footerMenu := site.Menu("footer")
	if footerMenu == nil {
		t.Fatal("expected 'footer' menu")
	}
	if len(footerMenu) != 2 {
		t.Errorf("expected 2 items in footer menu, got %d", len(footerMenu))
	}
}

func TestMenuSortByWeightThenName(t *testing.T) {
	config := DefaultConfig()
	config.Menus = map[string][]MenuItemConfig{
		"main": {
			{Name: "Zebra", URL: "/z/", Weight: 1},
			{Name: "Apple", URL: "/a/", Weight: 1},
			{Name: "Mango", URL: "/m/", Weight: 1},
		},
	}

	site := NewWithConfig(config)
	site.buildMenus()

	menu := site.Menu("main")

	// Same weight, should be sorted alphabetically
	if menu[0].Name != "Apple" {
		t.Errorf("expected 'Apple' first, got '%s'", menu[0].Name)
	}
	if menu[1].Name != "Mango" {
		t.Errorf("expected 'Mango' second, got '%s'", menu[1].Name)
	}
	if menu[2].Name != "Zebra" {
		t.Errorf("expected 'Zebra' third, got '%s'", menu[2].Name)
	}
}

func TestMenuWithActive(t *testing.T) {
	config := DefaultConfig()
	config.Menus = map[string][]MenuItemConfig{
		"main": {
			{Name: "Home", URL: "/", Weight: 1},
			{Name: "Blog", URL: "/blog/", Weight: 2},
		},
	}

	site := NewWithConfig(config)
	site.buildMenus()

	// Get menu with /blog/ as active
	menu := site.MenuWithActive("main", "/blog/")

	if menu[0].Active {
		t.Error("expected Home to not be active")
	}
	if !menu[1].Active {
		t.Error("expected Blog to be active")
	}

	// Verify original menu is unchanged
	original := site.Menu("main")
	if original[0].Active || original[1].Active {
		t.Error("original menu should not be modified")
	}
}

func TestMenuNotFound(t *testing.T) {
	config := DefaultConfig()
	site := NewWithConfig(config)
	site.buildMenus()

	menu := site.Menu("nonexistent")
	if menu != nil {
		t.Error("expected nil for nonexistent menu")
	}

	menuActive := site.MenuWithActive("nonexistent", "/")
	if menuActive != nil {
		t.Error("expected nil for nonexistent menu with active")
	}
}

func TestAllMenus(t *testing.T) {
	config := DefaultConfig()
	config.Menus = map[string][]MenuItemConfig{
		"main":   {{Name: "Home", URL: "/"}},
		"footer": {{Name: "Privacy", URL: "/privacy/"}},
		"social": {{Name: "Twitter", URL: "https://twitter.com"}},
	}

	site := NewWithConfig(config)
	site.buildMenus()

	names := site.AllMenus()

	if len(names) != 3 {
		t.Errorf("expected 3 menu names, got %d", len(names))
	}

	// Should be sorted alphabetically
	if names[0] != "footer" {
		t.Errorf("expected 'footer' first, got '%s'", names[0])
	}
	if names[1] != "main" {
		t.Errorf("expected 'main' second, got '%s'", names[1])
	}
	if names[2] != "social" {
		t.Errorf("expected 'social' third, got '%s'", names[2])
	}
}

func TestEmptyMenus(t *testing.T) {
	config := DefaultConfig()
	site := NewWithConfig(config)
	site.buildMenus()

	if len(site.Menus) != 0 {
		t.Errorf("expected 0 menus, got %d", len(site.Menus))
	}

	names := site.AllMenus()
	if len(names) != 0 {
		t.Errorf("expected 0 menu names, got %d", len(names))
	}
}
