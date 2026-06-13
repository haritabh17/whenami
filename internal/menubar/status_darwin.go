//go:build darwin

package menubar

import (
	"github.com/getlantern/systray"
	"github.com/haritabh17/theirtime/internal/config"
	"github.com/haritabh17/theirtime/internal/team"
)

func setMenubarContent(cfg *config.Config, members []team.MemberTime, avatars map[string][]byte) {
	segments := make([]systray.StatusSegment, 0, len(members)*2)
	for i, m := range members {
		text := team.FormatMemberBarText(cfg, m, "")
		img := team.MemberAvatar(cfg, m.ID, avatars)
		if text == "" && img == nil {
			continue
		}
		segments = append(segments, systray.StatusSegment{
			Text:  text,
			Image: img,
		})
		if i < len(members)-1 {
			segments = append(segments, systray.StatusSegment{Text: " | "})
		}
	}
	if len(segments) == 0 {
		return
	}
	defer func() {
		if recover() != nil {
			systray.SetTitle(team.FormatMenubarTitle(cfg, members))
		}
	}()
	systray.SetStatusSegments(segments)
}
