package cmd

import (
	"fmt"
	"os"

	"github.com/haritabh17/theirtime/internal/buildinfo"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "theirtime",
	Short: "Teammates' local times in your macOS menu bar",
	Long:  "theirtime — see coworkers' local times in your macOS menu bar, with Slack profile avatars. Token stays in macOS Keychain.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(onboardCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(offboardCmd)
	rootCmd.AddCommand(installAgentsCmd)
	rootCmd.AddCommand(teamCmd)
	rootCmd.AddCommand(menubarCmd)
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("theirtime %s (commit %s, built %s)\n", buildinfo.Version, buildinfo.Commit, buildinfo.Date)
	},
}
