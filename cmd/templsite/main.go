// Copyright (c) 2025-2026 Catapulsion LLC and contributors
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/dmoose/templsite/cmd/templsite/commands"
	"github.com/spf13/cobra"
)

// Version information set via ldflags at build time
// go build -ldflags "-X main.version=v1.0.0 -X main.commit=abc123 -X main.buildTime=..."
var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
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

	root := newRootCmd(ctx)
	err := root.Execute()
	cancel()
	if err != nil {
		os.Exit(1)
	}
}

func newRootCmd(ctx context.Context) *cobra.Command {
	c := commit
	if len(c) > 7 {
		c = c[:7]
	}

	root := &cobra.Command{
		Use:          "templsite",
		Short:        "A modern static site generator built with Go and templ",
		Long:         "templsite - A modern static site generator built with Go and templ\n\nDocumentation: https://github.com/dmoose/templsite",
		Version:      fmt.Sprintf("%s (%s) built %s", version, c, buildTime),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	root.SetVersionTemplate("templsite version {{.Version}}\n")

	versionStr := root.Version
	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version of templsite",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("templsite version %s\n", versionStr)
		},
	})

	root.AddCommand(commands.NewNewCmd(ctx))
	root.AddCommand(commands.NewBuildCmd(ctx))
	root.AddCommand(commands.NewServeCmd(ctx))

	return root
}
