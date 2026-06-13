package cmd

import (
	"github.com/spf13/cobra"
)

var menubarDemo bool

var menubarCmd = &cobra.Command{
	Use:   "menubar",
	Short: "Show watched teammates' local times in the macOS menu bar",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMenubar(menubarDemo)
	},
}

func init() {
	menubarCmd.Flags().BoolVar(&menubarDemo, "demo", false, "Preview menu bar with bundled cartoon avatars (no Slack; for screenshots)")
}
