package commands

import (
	"context"
	"fmt"
)

// Components manages templ components
func Components(ctx context.Context, args []string) error {
	// TODO: Implement in Stage 10
	if len(args) == 0 {
		return showComponentsHelp()
	}

	subcommand := args[0]
	switch subcommand {
	case "add":
		fmt.Println("components add - not yet implemented")
		fmt.Println("Usage: templsite components add <identifier>[@version]")
		fmt.Println("Example: templsite components add cc:ui/hero@latest")
	case "list":
		fmt.Println("components list - not yet implemented")
		fmt.Println("Usage: templsite components list")
	case "update":
		fmt.Println("components update - not yet implemented")
		fmt.Println("Usage: templsite components update [<identifier>]")
	case "remove":
		fmt.Println("components remove - not yet implemented")
		fmt.Println("Usage: templsite components remove <identifier>")
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}

	return nil
}

func showComponentsHelp() error {
	fmt.Println(`Component management commands

Usage:
  templsite components <subcommand> [arguments]

Subcommands:
  add       Install a component from the registry
  list      List installed components
  update    Update component(s) to latest version
  remove    Uninstall a component

Examples:
  templsite components add cc:ui/hero@latest
  templsite components list
  templsite components update cc:ui/hero
  templsite components remove cc:ui/hero`)
	return nil
}
