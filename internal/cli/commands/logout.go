package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/studio/platform/internal/cli/config"
)

// LogoutCmd represents the logout command
var LogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from the Studio API",
	Long:  `Remove stored credentials and logout from the Studio platform.`,
	RunE:  runLogout,
}

func runLogout(cmd *cobra.Command, args []string) error {
	store, err := config.NewCredentialsStore()
	if err != nil {
		return fmt.Errorf("failed to create credentials store: %w", err)
	}

	if err := store.Clear(); err != nil {
		return fmt.Errorf("failed to clear credentials: %w", err)
	}

	fmt.Println("✅ Logged out successfully!")

	return nil
}
