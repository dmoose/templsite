package commands

import (
	"context"
	"fmt"
)

// New creates a new site from a template
func New(ctx context.Context, args []string) error {
	// TODO: Implement in Stage 9
	fmt.Println("new command - not yet implemented")
	fmt.Println("Usage: templsite new <path> [--template <name>]")
	fmt.Println("Available templates: minimal, business, blog")
	return nil
}
