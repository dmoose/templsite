package commands

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/dmoose/templsite/pkg/server"
	"github.com/dmoose/templsite/pkg/site"
	"github.com/spf13/cobra"
)

// NewServeCmd creates the "serve" command
func NewServeCmd(ctx context.Context) *cobra.Command {
	var (
		configPath string
		env        string
		port       int
		addr       string
		verbose    bool
	)

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start development server with live reload",
		Long: `Start a development server with live reload. Watches for file changes and
automatically rebuilds the site and refreshes the browser.`,
		Example: `  templsite serve
  templsite serve --env staging
  templsite serve --port 3000
  templsite serve --config site.yaml --verbose
  templsite serve --addr 0.0.0.0:8080`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(ctx, configPath, env, port, addr, verbose)
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "config.yaml", "path to configuration file")
	cmd.Flags().StringVar(&env, "env", "", "environment (loads config.<env>.yaml overrides)")
	cmd.Flags().IntVar(&port, "port", 8080, "port to listen on")
	cmd.Flags().StringVar(&addr, "addr", "", "address to bind to (default: localhost:<port>)")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "enable verbose logging")

	return cmd
}

func runServe(ctx context.Context, configPath, env string, port int, addr string, verbose bool) error {
	// Setup logging level
	if verbose {
		logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
		slog.SetDefault(logger)
	}

	// Determine address
	serverAddr := addr
	if serverAddr == "" {
		serverAddr = fmt.Sprintf("localhost:%d", port)
	}

	slog.Debug("serve command started", "config", configPath, "addr", serverAddr)

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s\n\nRun 'templsite new <sitename>' to create a new site", configPath)
	}

	// Load site configuration
	slog.Info("loading configuration", "path", configPath, "env", env)
	s, err := site.NewWithEnv(configPath, env)
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
