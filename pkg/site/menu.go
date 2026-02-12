// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package site

import "sort"

// MenuItem represents a menu item with runtime state
type MenuItem struct {
	Name   string
	URL    string
	Weight int
	Active bool // Set during rendering based on current page
}

// buildMenus converts config menus to runtime MenuItem slices
func (s *Site) buildMenus() {
	s.Menus = make(map[string][]*MenuItem)

	if s.Config.Menus == nil {
		return
	}

	for name, items := range s.Config.Menus {
		menuItems := make([]*MenuItem, len(items))
		for i, item := range items {
			menuItems[i] = &MenuItem{
				Name:   item.Name,
				URL:    item.URL,
				Weight: item.Weight,
				Active: false,
			}
		}

		// Sort by weight, then by name
		sort.Slice(menuItems, func(i, j int) bool {
			if menuItems[i].Weight != menuItems[j].Weight {
				return menuItems[i].Weight < menuItems[j].Weight
			}
			return menuItems[i].Name < menuItems[j].Name
		})

		s.Menus[name] = menuItems
	}
}

// Menu returns a menu by name, or nil if not found
func (s *Site) Menu(name string) []*MenuItem {
	if s.Menus == nil {
		return nil
	}
	return s.Menus[name]
}

// MenuWithActive returns a menu with the Active flag set for the current URL
// This returns a new slice to avoid mutating the original
func (s *Site) MenuWithActive(name, currentURL string) []*MenuItem {
	menu := s.Menu(name)
	if menu == nil {
		return nil
	}

	result := make([]*MenuItem, len(menu))
	for i, item := range menu {
		result[i] = &MenuItem{
			Name:   item.Name,
			URL:    item.URL,
			Weight: item.Weight,
			Active: item.URL == currentURL,
		}
	}

	return result
}

// AllMenus returns all menu names
func (s *Site) AllMenus() []string {
	if s.Menus == nil {
		return nil
	}

	names := make([]string, 0, len(s.Menus))
	for name := range s.Menus {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
