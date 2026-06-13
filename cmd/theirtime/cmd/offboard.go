package cmd

import (
	"fmt"
	"os"

	"github.com/haritabh17/theirtime/internal/config"
	"github.com/haritabh17/theirtime/internal/keychain"
	"github.com/haritabh17/theirtime/internal/launchagent"
	"github.com/spf13/cobra"
)

var offboardCmd = &cobra.Command{
	Use:   "offboard",
	Short: "Remove LaunchAgent and delete local data",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := launchagent.Uninstall(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: uninstall LaunchAgent: %v\n", err)
		}
		_ = keychain.DeleteToken()
		_ = keychain.DeleteAppCredentials()

		path, err := config.Path()
		if err == nil {
			_ = os.Remove(path)
		}
		fmt.Println("theirtime removed from this Mac.")
		return nil
	},
}
