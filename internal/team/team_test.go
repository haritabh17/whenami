package team

import (
	"testing"
	"time"

	"github.com/haritabh17/theirtime/internal/config"
)

func defaultCfg() *config.Config {
	c := &config.Config{}
	config.ApplyDefaults(c)
	return c
}

func TestFormatMenubarTitleDefault(t *testing.T) {
	cfg := defaultCfg()
	members := []MemberTime{
		{Label: "sugu", Time: "10.46 AM"},
		{Label: "rafa", Time: "3.15 PM"},
	}
	got := FormatMenubarTitle(cfg, members)
	want := "10.46 AM | 3.15 PM"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFormatMenubarTitleWithNames(t *testing.T) {
	cfg := defaultCfg()
	show := true
	cfg.ShowName = &show
	members := []MemberTime{
		{Label: "sugu", Time: "10.46 AM"},
		{Label: "rafa", Time: "3.15 PM"},
	}
	got := FormatMenubarTitle(cfg, members)
	want := "sugu - 10.46 AM | rafa - 3.15 PM"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestTruncateMenubarTitle(t *testing.T) {
	cfg := defaultCfg()
	show := true
	cfg.ShowName = &show
	long := MemberTime{Label: "verylonglabel", Time: "10.46 AM"}
	title := FormatMenubarTitle(cfg, []MemberTime{long, long, long, long, long})
	if len([]rune(title)) > maxMenubarTitleRunes {
		t.Fatalf("title too long: %d runes", len([]rune(title)))
	}
	runes := []rune(title)
	if runes[len(runes)-1] != '…' {
		t.Fatalf("expected ellipsis, got %q", title)
	}
}

func TestFormatMemberTime(t *testing.T) {
	cfg := defaultCfg()
	at := time.Date(2026, 6, 13, 15, 30, 0, 0, time.UTC)
	got := FormatMemberTime(cfg, "America/New_York", at)
	if got == "" || got == "—" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatMemberDisplayAvatarTimeOnly(t *testing.T) {
	cfg := defaultCfg()
	got := FormatMemberDisplay(cfg, MemberTime{Label: "bob", Time: "4.07 PM"}, "")
	if got != "4.07 PM" {
		t.Fatalf("got %q", got)
	}
}
