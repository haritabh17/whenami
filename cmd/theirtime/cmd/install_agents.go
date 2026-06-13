package cmd

import (
	"fmt"

	"github.com/haritabh17/theirtime/internal/config"
	"github.com/haritabh17/theirtime/internal/launchagent"
	"github.com/spf13/cobra"
)

var installAgentsCmd = &cobra.Command{
	Use:   "install-agents",
	Short: "Install or refresh the menu bar LaunchAgent",
	RunE: func(cmd *cobra.Command, args []string) error {
		bin, err := launchagent.CurrentBinary()
		if err != nil {
			return err
		}
		if err := launchagent.Install(bin); err != nil {
			return err
		}
		cfg, _ := config.Load()
		if len(cfg.Team) > 0 {
			fmt.Println("Menu bar LaunchAgent installed.")
		} else {
			fmt.Println("No teammates configured — menu bar agent not started.")
			fmt.Println("Add teammates with theirtime team add …, then run install-agents again.")
		}
		return nil
	},
}
