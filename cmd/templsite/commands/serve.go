package commands

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"git.catapulsion.com/templsite/pkg/server"
	"git.catapulsion.com/templsite/pkg/site"
)

// Serve starts the development server with live reload
func Serve(ctx context.Context, args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	configPath := fs.String("config", "config.yaml", "path to configuration file")
	env := fs.String("env", "", "environment (loads config.<env>.yaml overrides)")
	port := fs.Int("port", 8080, "port to listen on")
	addr := fs.String("addr", "", "address to bind to (default: localhost:<port>)")
	verbose := fs.Bool("verbose", false, "enable verbose logging")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: templsite serve [options]

Start a development server with live reload. Watches for file changes and
automatically rebuilds the site and refreshes the browser.

Options:
`)
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  templsite serve
  templsite serve --env staging
  templsite serve --port 3000
  templsite serve --config site.yaml --verbose
  templsite serve --addr 0.0.0.0:8080

`)
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Setup logging level
	logLevel := slog.LevelInfo
	if *verbose {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	// Determine address
	serverAddr := *addr
	if serverAddr == "" {
		serverAddr = fmt.Sprintf("localhost:%d", *port)
	}

	slog.Debug("serve command started", "config", *configPath, "addr", serverAddr)

	// Check if config file exists
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s\n\nRun 'templsite new <sitename>' to create a new site", *configPath)
	}

	// Load site configuration
	slog.Info("loading configuration", "path", *configPath, "env", *env)
	s, err := site.NewWithEnv(*configPath, *env)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create development server with noop render function
	// (templsite serve is for existing projects that handle their own rendering)
	noopRender := func(ctx context.Context, s *site.Site) error {
		return nil // User's site binary handles rendering
	}
	srv, err := server.New(s, serverAddr, noopRender)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Start server (blocks until context is cancelled or error occurs)
	if err := srv.Serve(ctx); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
