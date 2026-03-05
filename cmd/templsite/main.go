package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"git.catapulsion.com/templsite/cmd/templsite/commands"
)

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Setup context with signal handling
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, os.Args[1:]); err != nil {
		slog.Error("command failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return showHelp()
	}

	switch args[0] {
	case "new":
		return commands.New(ctx, args[1:])
	case "serve":
		return commands.Serve(ctx, args[1:])
	case "build":
		return commands.Build(ctx, args[1:])
	case "components":
		return commands.Components(ctx, args[1:])
	case "help", "--help", "-h":
		return showHelp()
	case "version", "--version", "-v":
		return showVersion()
	default:
		return fmt.Errorf("unknown command: %s\n\nRun 'templsite help' for usage", args[0])
	}
}

func showHelp() error {
	fmt.Println(`templsite - A modern static site generator built with Go and templ

Usage:
  templsite <command> [arguments]

Commands:
  new          Create a new site from a template
  build        Build the site for production
  serve        Start development server with live reload
  components   Manage templ components
  version      Show version information
  help         Show this help message

Examples:
  templsite new mysite --template business
  templsite build
  templsite serve --port 8080

For more information about a command:
  templsite <command> --help

Documentation: https://github.com/yourorg/templsite`)
	return nil
}

func showVersion() error {
	// Version will be set via build flags
	version := "dev"
	fmt.Printf("templsite version %s\n", version)
	return nil
}
