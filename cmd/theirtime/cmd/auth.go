package cmd

import (
	"fmt"

	"github.com/haritabh17/theirtime/internal/auth"
	"github.com/haritabh17/theirtime/internal/config"
	"github.com/haritabh17/theirtime/internal/keychain"
	"github.com/haritabh17/theirtime/internal/onboard"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Re-authenticate with Slack",
	RunE: func(cmd *cobra.Command, args []string) error {
		clientID, clientSecret, err := onboard.EnsureAppCredentials()
		if err != nil {
			return err
		}
		fmt.Println("Opening Slack authorization…")
		result, err := auth.Authenticate(clientID, clientSecret)
		if err != nil {
			return err
		}
		if err := keychain.SetToken(result.Token); err != nil {
			return err
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cfg.SlackUserID = result.UserID
		return config.Save(cfg)
	},
}
