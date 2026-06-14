package cmd

import (
	"github.com/haritabh17/theirtime/internal/config"
	"github.com/haritabh17/theirtime/internal/launchagent"
	"github.com/haritabh17/theirtime/internal/ui"
	"github.com/spf13/cobra"
)

var installAgentsQuiet bool

var installAgentsCmd = &cobra.Command{
	Use:   "install-agents",
	Short: "Install or refresh the menu bar LaunchAgent",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Default.Quiet = installAgentsQuiet

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		ui.Heading("Install menu bar agent")

		bin, err := launchagent.CurrentBinary()
		if err != nil {
			return err
		}

		update, done := ui.ProgressSpinner("Installing menu bar agent…")
		err = launchagent.Install(bin, update)
		done()
		if err != nil {
			return err
		}

		if len(cfg.Team) > 0 {
			ui.InstallAgentsDone(len(cfg.Team))
		} else {
			ui.InstallAgentsEmpty()
		}
		return nil
	},
}

func init() {
	installAgentsCmd.Flags().BoolVar(&installAgentsQuiet, "quiet", false, "Minimal output (errors only)")
}
