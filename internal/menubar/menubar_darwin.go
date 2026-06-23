//go:build darwin

package menubar

import (
	"fmt"
	"sync"
	"time"

	"github.com/getlantern/systray"
	"github.com/haritabh17/theirtime/internal/config"
	"github.com/haritabh17/theirtime/internal/team"
)

const (
	clockInterval = time.Minute
	slackInterval = 15 * time.Minute
)

type app struct {
	mu             sync.Mutex
	cfg            *config.Config
	client         team.InfoClient
	members        []team.MemberTime
	rawAvatarCache map[string][]byte
	avatarCache    map[string]avatarEntry
	rows           []*systray.MenuItem
	refresh        *systray.MenuItem
	iconSizeLabel  *systray.MenuItem
	demo           bool
}

// Run starts the menu bar process (blocks until Quit).
func Run() error {
	return run(false)
}

// RunDemo starts the menu bar with bundled demo avatars and all display fields (avatar, name, 12h time).
func RunDemo() error {
	return run(true)
}

func run(demo bool) error {
	var (
		cfg    *config.Config
		client team.InfoClient
		err    error
	)
	if demo {
		cfg = demoConfig()
	} else {
		cfg, client, err = team.RequireOnboarded()
		if err != nil {
			return err
		}
		if len(cfg.Team) == 0 {
			return fmt.Errorf("no teammates configured — run theirtime team add <label> <U…>")
		}
	}

	a := &app{
		cfg:            cfg,
		client:         client,
		rawAvatarCache: make(map[string][]byte),
		avatarCache:    make(map[string]avatarEntry),
		demo:           demo,
	}
	systray.Run(a.onReady, a.onExit)
	return nil
}

func (a *app) onReady() {
	systray.SetTooltip("theirtime team times")

	a.refresh = systray.AddMenuItem("Refresh now", "")
	go func() {
		for range a.refresh.ClickedCh {
			a.refreshData()
		}
	}()

	systray.AddSeparator()
	a.rebuildMemberItems()
	systray.AddSeparator()
	a.addSettingsItems()
	systray.AddSeparator()
	quit := systray.AddMenuItem("Quit", "")
	go func() {
		<-quit.ClickedCh
		systray.Quit()
	}()

	a.refreshData()
	go a.clockLoop()
	if !a.demo {
		go a.slackLoop()
	}
}

func (a *app) onExit() {}

func (a *app) addSettingsItems() {
	settings := systray.AddMenuItem("Settings", "")
	current := config.IconSize(a.cfg)
	a.iconSizeLabel = settings.AddSubMenuItem(iconSizeTitle(current), "")
	a.iconSizeLabel.Disable()

	slider := settings.AddSubMenuItem("", "")
	slider.SetSlider(config.IconSizeMin, config.IconSizeMax, current)
	go func() {
		for size := range slider.SliderCh {
			a.setIconSize(size)
		}
	}()
}

func (a *app) setIconSize(size int) {
	size = config.NormalizeIconSize(size)
	if a.demo {
		a.mu.Lock()
		if a.cfg == nil {
			a.cfg = demoConfig()
		}
		a.cfg.IconSize = &size
		a.mu.Unlock()
		a.updateIconSizeLabel(size)
		a.updateDisplay(false)
		return
	}

	cfg, err := config.Load()
	if err != nil {
		return
	}
	cfg.IconSize = &size
	if err := config.Save(cfg); err != nil {
		return
	}
	a.mu.Lock()
	a.cfg = cfg
	a.mu.Unlock()
	a.updateIconSizeLabel(size)
	a.updateDisplay(false)
}

func (a *app) updateIconSizeLabel(size int) {
	if a.iconSizeLabel != nil {
		a.iconSizeLabel.SetTitle(iconSizeTitle(size))
	}
}

func iconSizeTitle(size int) string {
	return fmt.Sprintf("Icon size: %d pt", size)
}

func (a *app) rebuildMemberItems() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.rows = nil
	count := 2
	if !a.demo {
		cfg, err := config.Load()
		if err != nil {
			return
		}
		count = len(cfg.Team)
	}
	for i := 0; i < count; i++ {
		item := systray.AddMenuItem("…", "")
		item.Disable()
		a.rows = append(a.rows, item)
	}
}

func (a *app) clockLoop() {
	tick := time.NewTicker(clockInterval)
	defer tick.Stop()
	for range tick.C {
		a.updateDisplay(false)
	}
}

func (a *app) slackLoop() {
	tick := time.NewTicker(slackInterval)
	defer tick.Stop()
	for range tick.C {
		a.refreshData()
	}
}

func (a *app) refreshData() {
	if a.demo {
		a.mu.Lock()
		a.cfg = demoConfig()
		a.members = demoMembers()
		a.rawAvatarCache = demoAvatars()
		a.rebuildDisplayAvatars(a.members)
		a.mu.Unlock()
		a.updateDisplay(false)
		return
	}

	cfg, err := config.Load()
	if err != nil {
		return
	}
	a.mu.Lock()
	a.cfg = cfg
	a.mu.Unlock()

	members, err := team.FetchNow(a.client, cfg)
	if err != nil {
		systray.SetTitle("theirtime — error")
		return
	}

	a.mu.Lock()
	a.members = members
	a.refreshAvatars(members)
	a.rebuildDisplayAvatars(members)
	a.mu.Unlock()

	a.updateDisplay(true)
}

func (a *app) refreshAvatars(members []team.MemberTime) {
	if !config.ShowAvatar(a.cfg) {
		a.rawAvatarCache = nil
		a.avatarCache = nil
		return
	}
	next := make(map[string][]byte, len(members))
	for _, m := range members {
		if m.AvatarURL == "" {
			continue
		}
		if b, err := fetchAvatar(m.AvatarURL); err == nil {
			next[m.ID] = b
			continue
		}
		if cached, ok := a.rawAvatarCache[m.ID]; ok {
			next[m.ID] = cached
		}
	}
	a.rawAvatarCache = next
}

func (a *app) rebuildDisplayAvatars(_ []team.MemberTime) {
	if !config.ShowAvatar(a.cfg) {
		a.avatarCache = nil
		return
	}
	next := make(map[string]avatarEntry, len(a.rawAvatarCache))
	for id, raw := range a.rawAvatarCache {
		next[id] = avatarEntry{data: append([]byte(nil), raw...)}
	}
	a.avatarCache = next
}

func (a *app) updateDisplay(rebuildRows bool) {
	a.mu.Lock()
	members := append([]team.MemberTime(nil), a.members...)
	cfg := a.cfg
	a.mu.Unlock()

	if cfg == nil || len(members) == 0 {
		return
	}

	now := time.Now()
	for i := range members {
		if members[i].TZ != "" {
			members[i].Time = team.FormatMemberTime(cfg, members[i].TZ, now)
		}
	}

	title := team.FormatMenubarTitleAt(cfg, members, now)
	systray.SetTooltip(title)
	a.updateIconSizeLabel(config.IconSize(cfg))

	a.mu.Lock()
	avatars := make(map[string]avatarEntry, len(a.avatarCache))
	for k, v := range a.avatarCache {
		avatars[k] = avatarEntry{
			data:        append([]byte(nil), v.data...),
			contentSize: v.contentSize,
		}
	}
	a.mu.Unlock()

	setMenubarContent(cfg, members, avatars, now)

	if rebuildRows && len(a.rows) != len(members) {
		// Team list changed; systray cannot remove items — full restart would be needed.
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	rows := buildMemberMenuRows(cfg, members, avatars, now)
	for i := range a.rows {
		if i >= len(rows) {
			break
		}
		a.rows[i].SetTitle(rows[i].title)
		if rows[i].hasIcon {
			a.rows[i].SetIconWithSize(rows[i].icon.data, rows[i].icon.contentSize, config.IconSize(cfg))
		}
	}
}

type memberMenuRow struct {
	title   string
	icon    avatarEntry
	hasIcon bool
}

func buildMemberMenuRows(cfg *config.Config, members []team.MemberTime, avatars map[string]avatarEntry, at time.Time) []memberMenuRow {
	rowMembers := team.SortMemberTimesByUTCOffsetAt(members, at)
	rows := make([]memberMenuRow, 0, len(rowMembers))
	for _, m := range rowMembers {
		row := memberMenuRow{title: team.FormatMenuLine(cfg, m)}
		if entry, ok := avatars[m.ID]; ok && len(entry.data) > 0 && config.ShowAvatar(cfg) {
			row.icon = entry
			row.hasIcon = true
		}
		rows = append(rows, row)
	}
	return rows
}
