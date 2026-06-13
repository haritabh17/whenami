package team

import (
	"fmt"
	"strings"
	"time"

	"github.com/haritabh17/theirtime/internal/config"
	"github.com/haritabh17/theirtime/internal/slack"
	"github.com/haritabh17/theirtime/internal/timeformat"
)

const (
	maxMenubarTitleRunes = 72
)

// MemberTime is a watched teammate with a formatted local time.
type MemberTime struct {
	Label       string
	ID          string
	TZ          string
	DisplayName string
	Time        string
	AvatarURL   string
}

// InfoClient fetches Slack user profile fields.
type InfoClient interface {
	GetUserInfo(userID string) (slack.UserInfo, error)
}

// ListWithTimes resolves each configured member's current local time.
func ListWithTimes(client InfoClient, cfg *config.Config, at time.Time) ([]MemberTime, error) {
	if cfg == nil {
		return nil, fmt.Errorf("not configured")
	}
	out := make([]MemberTime, 0, len(cfg.Team))
	for _, m := range cfg.Team {
		mt := MemberTime{Label: m.Label, ID: m.ID}
		info, err := client.GetUserInfo(m.ID)
		if err != nil {
			mt.Time = "—"
			out = append(out, mt)
			continue
		}
		mt.DisplayName = info.DisplayName
		mt.TZ = info.TZ
		mt.AvatarURL = info.AvatarURL
		if info.TZ == "" {
			mt.Time = "—"
		} else {
			mt.Time = FormatMemberTime(cfg, info.TZ, at)
		}
		out = append(out, mt)
	}
	return out, nil
}

// FormatMenubarTitle builds the compact menu bar string (text-only fallback).
func FormatMenubarTitle(cfg *config.Config, members []MemberTime) string {
	return truncateMenubarTitle(joinMemberDisplays(cfg, members, " | "))
}

func joinMemberDisplays(cfg *config.Config, members []MemberTime, sep string) string {
	parts := make([]string, 0, len(members))
	for _, m := range members {
		text := FormatMemberDisplay(cfg, m, "")
		if text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, sep)
}

func truncateMenubarTitle(title string) string {
	if len([]rune(title)) <= maxMenubarTitleRunes {
		return title
	}
	runes := []rune(title)
	return string(runes[:maxMenubarTitleRunes-1]) + "…"
}

// FormatMemberTime formats local time for a timezone using config preferences.
func FormatMemberTime(cfg *config.Config, tz string, at time.Time) string {
	if tz == "" {
		return "—"
	}
	return timeformat.FormatClock(tz, cfg.Format24h, config.TimePrecision(cfg), at)
}

// FormatMemberDisplay builds visible text for a member from config toggles.
func FormatMemberDisplay(cfg *config.Config, m MemberTime, suffix string) string {
	var parts []string
	if config.ShowName(cfg) {
		label := m.Label
		if label == "" {
			label = m.DisplayName
		}
		if label != "" {
			parts = append(parts, label)
		}
	}
	if config.ShowTime(cfg) {
		t := m.Time
		if t == "" {
			t = "—"
		}
		parts = append(parts, t)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " - ") + suffix
}

// MemberAvatar returns avatar bytes when show_avatar is enabled.
func MemberAvatar(cfg *config.Config, id string, avatars map[string][]byte) []byte {
	if !config.ShowAvatar(cfg) {
		return nil
	}
	return avatars[id]
}

func FormatMenuLine(cfg *config.Config, m MemberTime) string {
	text := FormatMemberDisplay(cfg, m, "")
	if text == "" && config.ShowAvatar(cfg) {
		text = m.Label
	}
	if config.ShowTime(cfg) {
		tz := m.TZ
		if tz == "" {
			tz = "no timezone"
		}
		if text == "" {
			return fmt.Sprintf("%s (%s)", m.Time, tz)
		}
		return fmt.Sprintf("%s (%s)", text, tz)
	}
	if text == "" {
		return m.Label
	}
	return text
}

// FormatMemberBarText returns the text chunk shown beside an avatar in the menu bar.
func FormatMemberBarText(cfg *config.Config, m MemberTime, separator string) string {
	return FormatMemberDisplay(cfg, m, separator)
}
