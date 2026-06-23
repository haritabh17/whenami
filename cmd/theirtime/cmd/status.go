package cmd

import (
	"fmt"

	"github.com/haritabh17/theirtime/internal/config"
	"github.com/haritabh17/theirtime/internal/keychain"
	"github.com/haritabh17/theirtime/internal/launchagent"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show configuration and agent state",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		path, _ := config.Path()

		hasToken, hasApp := keychain.Presence()

		fmt.Println("theirtime")
		fmt.Println("──────────")
		fmt.Printf("Config:          %s\n", path)
		fmt.Printf("Keychain token:  %v\n", hasToken)
		fmt.Printf("Slack app creds: %v\n", hasApp)
		fmt.Printf("Menu bar agent:  %v (%d teammate(s) watched)\n", launchagent.IsMenubarInstalled(), len(cfg.Team))
		fmt.Printf("Show avatar:     %v\n", config.ShowAvatar(cfg))
		fmt.Printf("Show name:       %v\n", config.ShowName(cfg))
		fmt.Printf("Show time:       %v\n", config.ShowTime(cfg))
		fmt.Printf("24-hour:         %v\n", cfg.Format24h)
		fmt.Printf("Time precision:  %s\n", config.TimePrecision(cfg))
		fmt.Printf("Icon size:       %dpt\n", config.IconSize(cfg))
		if !hasToken {
			fmt.Println("\nNot onboarded. Run: theirtime onboard")
		}
		return nil
	},
}
