package cmd

import (
	"fmt"
	"os"

	"github.com/haritabh17/theirtime/internal/config"
	"github.com/haritabh17/theirtime/internal/slack"
	"github.com/haritabh17/theirtime/internal/team"
	"github.com/spf13/cobra"
)

var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Manage teammates to watch in the menu bar",
}

var teamAddCmd = &cobra.Command{
	Use:   "add <label> <slack_member_id>",
	Short: "Add a teammate by your label and their Slack member ID",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		label, id := args[0], args[1]

		cfg, client, err := team.RequireOnboarded()
		if err != nil {
			return err
		}

		info, err := client.GetUserInfo(id)
		if err != nil {
			if slack.IsMissingScope(err) {
				return fmt.Errorf("%w\n\nRe-run theirtime auth to grant users:read", err)
			}
			return fmt.Errorf("verify member: %w", err)
		}
		if info.TZ == "" {
			fmt.Fprintf(os.Stderr, "warning: %s has no timezone in Slack — menu bar will show — until they set one\n", label)
		}

		if err := config.AddTeamMember(label, id); err != nil {
			return err
		}
		_ = cfg
		fmt.Printf("Added %q (%s)", label, id)
		if info.DisplayName != "" {
			fmt.Printf(" — Slack: %s", info.DisplayName)
		}
		fmt.Println()
		fmt.Println("Run theirtime install-agents to start or refresh the menu bar.")
		return nil
	},
}

var teamListCmd = &cobra.Command{
	Use:   "list",
	Short: "List watched teammates and their current local times",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, client, err := team.RequireOnboarded()
		if err != nil {
			return err
		}
		if len(cfg.Team) == 0 {
			fmt.Println("No teammates watched. Add one: theirtime team add <label> <U…>")
			return nil
		}
		members, err := team.FetchNow(client, cfg)
		if err != nil {
			return err
		}
		team.PrintListTable(members)
		return nil
	},
}

var teamRemoveCmd = &cobra.Command{
	Use:   "remove <label>",
	Short: "Stop watching a teammate by label",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.RemoveTeamMember(args[0]); err != nil {
			return err
		}
		fmt.Printf("Removed %q\n", args[0])
		fmt.Println("Run theirtime install-agents to refresh LaunchAgents if the menu bar should stop.")
		return nil
	},
}

func init() {
	teamCmd.AddCommand(teamAddCmd)
	teamCmd.AddCommand(teamListCmd)
	teamCmd.AddCommand(teamRemoveCmd)
}
