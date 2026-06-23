//go:build darwin

package menubar

import (
	"bytes"
	"testing"
	"time"

	"github.com/getlantern/systray"
	"github.com/haritabh17/theirtime/internal/config"
	"github.com/haritabh17/theirtime/internal/team"
)

func defaultStatusCfg() *config.Config {
	cfg := &config.Config{}
	config.ApplyDefaults(cfg)
	return cfg
}

func TestBuildMenubarSegmentsGroupsAvatarsBeforeSharedTime(t *testing.T) {
	cfg := defaultStatusCfg()
	at := time.Date(2026, 6, 23, 12, 0, 0, 0, time.UTC)
	members := []team.MemberTime{
		{ID: "u1", Label: "sugu", TZ: "America/Los_Angeles", Time: "10.46 AM"},
		{ID: "u2", Label: "rafa", TZ: "America/Los_Angeles", Time: "10.46 AM"},
		{ID: "u3", Label: "ann", TZ: "America/New_York", Time: "1.46 PM"},
	}
	avatars := map[string]avatarEntry{
		"u1": {data: []byte("avatar-1"), contentSize: 48},
		"u2": {data: []byte("avatar-2"), contentSize: 48},
		"u3": {data: []byte("avatar-3"), contentSize: 48},
	}

	segments := buildMenubarSegmentsAt(cfg, members, avatars, at)
	if len(segments) != 6 {
		t.Fatalf("got %d segments want 6: %#v", len(segments), segments)
	}
	assertImageSegment(t, segments[0], []byte("avatar-1"), 48)
	assertImageSegment(t, segments[1], []byte("avatar-2"), 48)
	assertTextSegment(t, segments[2], "10.46 AM")
	assertTextSegment(t, segments[3], " | ")
	assertImageSegment(t, segments[4], []byte("avatar-3"), 48)
	assertTextSegment(t, segments[5], "1.46 PM")
}

func TestBuildMenubarSegmentsUsesGroupedNamesWhenEnabled(t *testing.T) {
	cfg := defaultStatusCfg()
	at := time.Date(2026, 6, 23, 12, 0, 0, 0, time.UTC)
	showName := true
	cfg.ShowName = &showName
	members := []team.MemberTime{
		{ID: "u1", Label: "sugu", TZ: "America/Los_Angeles", Time: "10.46 AM"},
		{ID: "u2", Label: "rafa", TZ: "America/Los_Angeles", Time: "10.46 AM"},
	}

	segments := buildMenubarSegmentsAt(cfg, members, nil, at)
	if len(segments) != 1 {
		t.Fatalf("got %d segments want 1: %#v", len(segments), segments)
	}
	assertTextSegment(t, segments[0], "sugu, rafa - 10.46 AM")
}

func TestBuildMenubarSegmentsGroupsSameCurrentOffset(t *testing.T) {
	cfg := defaultStatusCfg()
	at := time.Date(2026, 6, 23, 12, 0, 0, 0, time.UTC)
	members := []team.MemberTime{
		{ID: "u1", Label: "manan", TZ: "Europe/Belgrade", Time: "3.02 PM"},
		{ID: "u2", Label: "nico", TZ: "Europe/Amsterdam", Time: "3.02 PM"},
		{ID: "u3", Label: "mats", TZ: "Europe/Amsterdam", Time: "3.02 PM"},
	}
	avatars := map[string]avatarEntry{
		"u1": {data: []byte("avatar-1")},
		"u2": {data: []byte("avatar-2")},
		"u3": {data: []byte("avatar-3")},
	}

	segments := buildMenubarSegmentsAt(cfg, members, avatars, at)
	if len(segments) != 4 {
		t.Fatalf("got %d segments want 4: %#v", len(segments), segments)
	}
	assertImageSegment(t, segments[0], []byte("avatar-1"), 0)
	assertImageSegment(t, segments[1], []byte("avatar-2"), 0)
	assertImageSegment(t, segments[2], []byte("avatar-3"), 0)
	assertTextSegment(t, segments[3], "3.02 PM")
}

func TestBuildMenubarSegmentsSortsGroupsByUTCOffset(t *testing.T) {
	cfg := defaultStatusCfg()
	at := time.Date(2026, 6, 23, 12, 0, 0, 0, time.UTC)
	members := []team.MemberTime{
		{ID: "ny", Label: "ny", TZ: "America/New_York", Time: "9.02 AM"},
		{ID: "la", Label: "la", TZ: "America/Los_Angeles", Time: "6.02 AM"},
		{ID: "in", Label: "in", TZ: "Asia/Kolkata", Time: "6.32 PM"},
	}
	avatars := map[string]avatarEntry{
		"ny": {data: []byte("avatar-ny")},
		"la": {data: []byte("avatar-la")},
		"in": {data: []byte("avatar-in")},
	}

	segments := buildMenubarSegmentsAt(cfg, members, avatars, at)
	if len(segments) != 8 {
		t.Fatalf("got %d segments want 8: %#v", len(segments), segments)
	}
	assertImageSegment(t, segments[0], []byte("avatar-la"), 0)
	assertTextSegment(t, segments[1], "6.02 AM")
	assertTextSegment(t, segments[2], " | ")
	assertImageSegment(t, segments[3], []byte("avatar-ny"), 0)
	assertTextSegment(t, segments[4], "9.02 AM")
	assertTextSegment(t, segments[5], " | ")
	assertImageSegment(t, segments[6], []byte("avatar-in"), 0)
	assertTextSegment(t, segments[7], "6.32 PM")
}

func TestBuildMemberMenuRowsSortsByUTCOffset(t *testing.T) {
	cfg := defaultStatusCfg()
	at := time.Date(2026, 6, 23, 12, 0, 0, 0, time.UTC)
	members := []team.MemberTime{
		{ID: "ny", Label: "ny", TZ: "America/New_York", Time: "9.02 AM"},
		{ID: "la", Label: "la", TZ: "America/Los_Angeles", Time: "6.02 AM"},
		{ID: "in", Label: "in", TZ: "Asia/Kolkata", Time: "6.32 PM"},
	}
	avatars := map[string]avatarEntry{
		"ny": {data: []byte("avatar-ny")},
		"la": {data: []byte("avatar-la")},
		"in": {data: []byte("avatar-in")},
	}

	rows := buildMemberMenuRows(cfg, members, avatars, at)
	wantTitles := []string{
		"6.02 AM (America/Los_Angeles)",
		"9.02 AM (America/New_York)",
		"6.32 PM (Asia/Kolkata)",
	}
	wantImages := [][]byte{[]byte("avatar-la"), []byte("avatar-ny"), []byte("avatar-in")}
	if len(rows) != len(wantTitles) {
		t.Fatalf("got %d rows want %d", len(rows), len(wantTitles))
	}
	for i := range wantTitles {
		if rows[i].title != wantTitles[i] {
			t.Fatalf("row %d title got %q want %q", i, rows[i].title, wantTitles[i])
		}
		if !rows[i].hasIcon {
			t.Fatalf("row %d expected icon", i)
		}
		if !bytes.Equal(rows[i].icon.data, wantImages[i]) {
			t.Fatalf("row %d icon got %q want %q", i, rows[i].icon.data, wantImages[i])
		}
	}
}

func assertImageSegment(t *testing.T, got systray.StatusSegment, want []byte, wantSize int) {
	t.Helper()
	if got.Text != "" {
		t.Fatalf("image segment text got %q want empty", got.Text)
	}
	if !bytes.Equal(got.Image, want) {
		t.Fatalf("image segment bytes got %q want %q", got.Image, want)
	}
	if got.AvatarSize != wantSize {
		t.Fatalf("avatar size got %d want %d", got.AvatarSize, wantSize)
	}
}

func assertTextSegment(t *testing.T, got systray.StatusSegment, want string) {
	t.Helper()
	if got.Text != want {
		t.Fatalf("text segment got %q want %q", got.Text, want)
	}
	if len(got.Image) != 0 {
		t.Fatalf("text segment image got %q want empty", got.Image)
	}
	if got.AvatarSize != 0 {
		t.Fatalf("text segment avatar size got %d want 0", got.AvatarSize)
	}
}
