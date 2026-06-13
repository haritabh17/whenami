package cmd

import (
	"fmt"

	"github.com/haritabh17/theirtime/internal/auth"
	"github.com/haritabh17/theirtime/internal/config"
	"github.com/haritabh17/theirtime/internal/keychain"
	"github.com/haritabh17/theirtime/internal/onboard"
	"github.com/spf13/cobra"
)

var onboardCmd = &cobra.Command{
	Use:   "onboard",
	Short: "Create your Slack app, authorize, and get ready to watch teammates",
	RunE: func(cmd *cobra.Command, args []string) error {
		clientID, clientSecret, err := onboard.EnsureAppCredentials()
		if err != nil {
			return err
		}

		fmt.Println("Step 2 of 2 — Authorize theirtime")
		fmt.Println("─────────────────────────────────")
		fmt.Println("Opening Slack authorization in your browser…")
		result, err := auth.Authenticate(clientID, clientSecret)
		if err != nil {
			return err
		}

		if err := keychain.SetToken(result.Token); err != nil {
			return fmt.Errorf("store token in Keychain: %w", err)
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if cfg == nil {
			cfg = &config.Config{}
		}
		cfg.SlackUserID = result.UserID
		config.ApplyDefaults(cfg)
		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Println("Connected to Slack.")
		fmt.Println("Add teammates: theirtime team add <label> <U…>")
		fmt.Println("Then run: theirtime install-agents")
		fmt.Println("Run `theirtime status` to check state.")
		return nil
	},
}
