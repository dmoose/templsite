package commands

import (
	"context"
	"fmt"
)

// Build builds the site for production
func Build(ctx context.Context, args []string) error {
	// TODO: Implement in Stage 7
	fmt.Println("build command - not yet implemented")
	fmt.Println("Usage: templsite build [--config <path>] [--output <dir>] [--verbose]")
	return nil
}
