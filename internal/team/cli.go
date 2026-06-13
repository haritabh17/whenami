package team

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/haritabh17/theirtime/internal/config"
	"github.com/haritabh17/theirtime/internal/keychain"
	"github.com/haritabh17/theirtime/internal/slack"
)

// RequireOnboarded returns config and a Slack client, or an error if not set up.
func RequireOnboarded() (*config.Config, *slack.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}
	token, err := keychain.GetToken()
	if err != nil {
		return nil, nil, fmt.Errorf("not onboarded — run theirtime onboard first")
	}
	return cfg, slack.NewClient(token), nil
}

// PrintListTable writes member times to stdout.
func PrintListTable(members []MemberTime) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "LABEL\tID\tLOCAL TIME\tTIMEZONE")
	for _, m := range members {
		tz := m.TZ
		if tz == "" {
			tz = "—"
		}
		t := m.Time
		if t == "" {
			t = "—"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", m.Label, m.ID, t, tz)
	}
	_ = w.Flush()
}

// FetchNow loads team times at the current instant.
func FetchNow(client InfoClient, cfg *config.Config) ([]MemberTime, error) {
	return ListWithTimes(client, cfg, time.Now())
}
