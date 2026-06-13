package cmd

import (
	"fmt"

	"github.com/haritabh17/theirtime/internal/auth"
	"github.com/haritabh17/theirtime/internal/config"
	"github.com/haritabh17/theirtime/internal/keychain"
	"github.com/haritabh17/theirtime/internal/onboard"
	"github.com/haritabh17/theirtime/internal/ui"
	"github.com/spf13/cobra"
)

var (
	onboardQuiet   bool
	onboardVerbose bool
)

var onboardCmd = &cobra.Command{
	Use:   "onboard",
	Short: "Create your Slack app, authorize, and get ready to watch teammates",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Default.Quiet = onboardQuiet
		ui.Default.Verbose = onboardVerbose

		ui.SetupIntro()

		hadCreds := keychain.HasAppCredentials()
		clientID, clientSecret, err := onboard.EnsureAppCredentials()
		if err != nil {
			return err
		}
		if hadCreds {
			ui.Success("Slack app credentials found — skipping app creation")
			ui.Blank()
		}
		_ = clientID
		_ = clientSecret

		ui.Step(2, 2, "Authorize")
		ui.Action("Opening Slack in your browser…")
		stop := ui.Spinner("Waiting for sign-in…")
		result, err := auth.Authenticate(clientID, clientSecret)
		stop()
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

		ui.Success("Connected to Slack")
		ui.DoneCard()
		return nil
	},
}

func init() {
	onboardCmd.Flags().BoolVar(&onboardQuiet, "quiet", false, "Minimal output (errors only)")
	onboardCmd.Flags().BoolVar(&onboardVerbose, "verbose", false, "Show detailed browser instructions")
}
