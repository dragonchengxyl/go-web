package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/studio/platform/internal/cli/commands"
)

var (
	version = "1.0.0"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "studio-cli",
		Short: "Studio CLI - Game release management tool",
		Long: `Studio CLI is a command-line tool for managing game releases on the Studio platform.

It provides commands for:
- Authenticating with the API
- Publishing new game releases
- Uploading game packages with progress tracking
- Managing release metadata`,
		Version: version,
	}

	// Add commands
	rootCmd.AddCommand(commands.LoginCmd)
	rootCmd.AddCommand(commands.LogoutCmd)
	rootCmd.AddCommand(commands.PublishCmd)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
