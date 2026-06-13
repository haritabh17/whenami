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
	mu          sync.Mutex
	cfg         *config.Config
	client      team.InfoClient
	members     []team.MemberTime
	avatarCache map[string][]byte
	rows        []*systray.MenuItem
	refresh     *systray.MenuItem
	demo        bool
}

// Run starts the menu bar process (blocks until Quit).
func Run() error {
	return run(false)
}

// RunDemo starts the menu bar with bundled cartoon avatars for screenshots.
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
		cfg:         cfg,
		client:      client,
		avatarCache: make(map[string][]byte),
		demo:        demo,
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
		a.avatarCache = demoAvatars()
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
	a.mu.Unlock()

	a.updateDisplay(true)
}

func (a *app) refreshAvatars(members []team.MemberTime) {
	if !config.ShowAvatar(a.cfg) {
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
		if cached, ok := a.avatarCache[m.ID]; ok {
			next[m.ID] = cached
		}
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

	title := team.FormatMenubarTitle(cfg, members)
	systray.SetTooltip(title)

	a.mu.Lock()
	avatars := make(map[string][]byte, len(a.avatarCache))
	for k, v := range a.avatarCache {
		copied := append([]byte(nil), v...)
		avatars[k] = copied
	}
	a.mu.Unlock()

	setMenubarContent(cfg, members, avatars)

	if rebuildRows && len(a.rows) != len(members) {
		// Team list changed; systray cannot remove items — full restart would be needed.
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range a.rows {
		if i >= len(members) {
			break
		}
		a.rows[i].SetTitle(team.FormatMenuLine(cfg, members[i]))
		if img := team.MemberAvatar(cfg, members[i].ID, avatars); len(img) > 0 {
			a.rows[i].SetIcon(img)
		}
	}
}
