package config

import (
	"fmt"
	"regexp"
)

var (
	labelRE  = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	slackIDRE = regexp.MustCompile(`^[UW][A-Z0-9]{8,}$`)
)

// TeamMember is a watched teammate (label is your short name for the menu bar).
type TeamMember struct {
	Label string `yaml:"label"`
	ID    string `yaml:"id"`
}

func ValidateTeamLabel(label string) error {
	if label == "" {
		return fmt.Errorf("label must not be empty")
	}
	if !labelRE.MatchString(label) {
		return fmt.Errorf("label %q: use letters, numbers, _ or - only", label)
	}
	return nil
}

func ValidateSlackMemberID(id string) error {
	if !slackIDRE.MatchString(id) {
		return fmt.Errorf("invalid Slack member ID %q (expected U… or W…)", id)
	}
	return nil
}

// AddTeamMember appends a member after validation and duplicate checks.
func AddTeamMember(label, id string) error {
	if err := ValidateTeamLabel(label); err != nil {
		return err
	}
	if err := ValidateSlackMemberID(id); err != nil {
		return err
	}

	cfg, err := Load()
	if err != nil {
		return err
	}
	for _, m := range cfg.Team {
		if m.Label == label {
			return fmt.Errorf("label %q already exists", label)
		}
		if m.ID == id {
			return fmt.Errorf("member %s already watched as %q", id, m.Label)
		}
	}
	cfg.Team = append(cfg.Team, TeamMember{Label: label, ID: id})
	return Save(cfg)
}

// RemoveTeamMember removes a member by label (case-sensitive).
func RemoveTeamMember(label string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	var kept []TeamMember
	var found bool
	for _, m := range cfg.Team {
		if m.Label == label {
			found = true
			continue
		}
		kept = append(kept, m)
	}
	if !found {
		return fmt.Errorf("no team member with label %q", label)
	}
	cfg.Team = kept
	return Save(cfg)
}
