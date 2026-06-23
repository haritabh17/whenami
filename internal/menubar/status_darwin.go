//go:build darwin

package menubar

import (
	"time"

	"github.com/getlantern/systray"
	"github.com/haritabh17/theirtime/internal/config"
	"github.com/haritabh17/theirtime/internal/team"
)

func setMenubarContent(cfg *config.Config, members []team.MemberTime, avatars map[string]avatarEntry, at time.Time) {
	segments := buildMenubarSegmentsAt(cfg, members, avatars, at)
	if len(segments) == 0 {
		return
	}
	defer func() {
		if recover() != nil {
			systray.SetTitle(team.FormatMenubarTitleAt(cfg, members, at))
		}
	}()
	systray.SetStatusSegments(segments)
}

func buildMenubarSegments(cfg *config.Config, members []team.MemberTime, avatars map[string]avatarEntry) []systray.StatusSegment {
	return buildMenubarSegmentsAt(cfg, members, avatars, time.Now())
}

func buildMenubarSegmentsAt(cfg *config.Config, members []team.MemberTime, avatars map[string]avatarEntry, at time.Time) []systray.StatusSegment {
	groups := team.GroupMemberTimesAt(members, at)
	segments := make([]systray.StatusSegment, 0, len(members)*2)
	for _, group := range groups {
		groupSegments := buildGroupSegments(cfg, group, avatars)
		if len(groupSegments) == 0 {
			continue
		}
		if len(segments) > 0 {
			segments = append(segments, systray.StatusSegment{Text: " | "})
		}
		segments = append(segments, groupSegments...)
	}
	return segments
}

func buildGroupSegments(cfg *config.Config, group team.MemberTimeGroup, avatars map[string]avatarEntry) []systray.StatusSegment {
	segments := make([]systray.StatusSegment, 0, len(group.Members)+1)
	if config.ShowAvatar(cfg) {
		for _, m := range group.Members {
			entry, ok := avatars[m.ID]
			if !ok || len(entry.data) == 0 {
				continue
			}
			segments = append(segments, systray.StatusSegment{
				Image:       entry.data,
				AvatarSize:  entry.contentSize,
				DisplaySize: config.IconSize(cfg),
			})
		}
	}
	if text := team.FormatMemberGroupDisplay(cfg, group, ""); text != "" {
		segments = append(segments, systray.StatusSegment{Text: text})
	}
	return segments
}
