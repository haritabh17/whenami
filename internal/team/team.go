package team

import (
	"fmt"
	"sort"
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

// MemberTimeGroup is a display group of teammates sharing one UTC offset.
type MemberTimeGroup struct {
	Members []MemberTime
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
	return FormatMenubarTitleAt(cfg, members, time.Now())
}

// FormatMenubarTitleAt builds the compact menu bar string for a specific instant.
func FormatMenubarTitleAt(cfg *config.Config, members []MemberTime, at time.Time) string {
	return truncateMenubarTitle(joinMemberGroupDisplays(cfg, GroupMemberTimesAt(members, at), " | "))
}

// GroupMemberTimes groups teammates by current UTC offset while preserving configured order.
func GroupMemberTimes(members []MemberTime) []MemberTimeGroup {
	return GroupMemberTimesAt(members, time.Now())
}

// SortMemberTimesByUTCOffset orders members by their current UTC offset.
func SortMemberTimesByUTCOffset(members []MemberTime) []MemberTime {
	return SortMemberTimesByUTCOffsetAt(members, time.Now())
}

// SortMemberTimesByUTCOffsetAt orders members by UTC offset at a specific instant.
func SortMemberTimesByUTCOffsetAt(members []MemberTime, at time.Time) []MemberTime {
	return FlattenMemberTimeGroups(GroupMemberTimesAt(members, at))
}

// GroupMemberTimesAt groups teammates by UTC offset at a specific instant.
func GroupMemberTimesAt(members []MemberTime, at time.Time) []MemberTimeGroup {
	type offsetGroup struct {
		group MemberTimeGroup
		valid bool
		value int
		order int
	}

	groups := make([]offsetGroup, 0, len(members))
	byOffset := make(map[int]int)
	for _, m := range members {
		offset, ok := memberUTCOffset(m, at)
		if !ok {
			groups = append(groups, offsetGroup{
				group: MemberTimeGroup{Members: []MemberTime{m}},
				order: len(groups),
			})
			continue
		}
		if idx, ok := byOffset[offset]; ok {
			groups[idx].group.Members = append(groups[idx].group.Members, m)
			continue
		}
		byOffset[offset] = len(groups)
		groups = append(groups, offsetGroup{
			group: MemberTimeGroup{Members: []MemberTime{m}},
			valid: true,
			value: offset,
			order: len(groups),
		})
	}
	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].valid != groups[j].valid {
			return groups[i].valid
		}
		if groups[i].valid && groups[i].value != groups[j].value {
			return groups[i].value < groups[j].value
		}
		return groups[i].order < groups[j].order
	})

	out := make([]MemberTimeGroup, 0, len(groups))
	for _, g := range groups {
		out = append(out, g.group)
	}
	return out
}

// FlattenMemberTimeGroups returns group members in their rendered order.
func FlattenMemberTimeGroups(groups []MemberTimeGroup) []MemberTime {
	count := 0
	for _, g := range groups {
		count += len(g.Members)
	}
	out := make([]MemberTime, 0, count)
	for _, g := range groups {
		out = append(out, g.Members...)
	}
	return out
}

func memberUTCOffset(m MemberTime, at time.Time) (int, bool) {
	if m.TZ == "" {
		return 0, false
	}
	loc, err := time.LoadLocation(m.TZ)
	if err != nil {
		return 0, false
	}
	_, offset := at.In(loc).Zone()
	return offset, true
}

func joinMemberGroupDisplays(cfg *config.Config, groups []MemberTimeGroup, sep string) string {
	parts := make([]string, 0, len(groups))
	for _, g := range groups {
		text := FormatMemberGroupDisplay(cfg, g, "")
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

// FormatMemberGroupDisplay builds visible text for teammates sharing a UTC offset.
func FormatMemberGroupDisplay(cfg *config.Config, group MemberTimeGroup, suffix string) string {
	if len(group.Members) == 0 {
		return ""
	}
	var parts []string
	if config.ShowName(cfg) {
		names := make([]string, 0, len(group.Members))
		for _, m := range group.Members {
			label := m.Label
			if label == "" {
				label = m.DisplayName
			}
			if label != "" {
				names = append(names, label)
			}
		}
		if len(names) > 0 {
			parts = append(parts, strings.Join(names, ", "))
		}
	}
	if config.ShowTime(cfg) {
		t := group.Members[0].Time
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
