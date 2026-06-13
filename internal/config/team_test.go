package config

import (
	"os"
	"path/filepath"
	"testing"
)

func testConfigDir(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, "Library", "Application Support", "theirtime")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
}

func TestValidateTeamLabel(t *testing.T) {
	if err := ValidateTeamLabel("sugu"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateTeamLabel(""); err == nil {
		t.Fatal("expected error for empty label")
	}
	if err := ValidateTeamLabel("bad name"); err == nil {
		t.Fatal("expected error for spaces")
	}
}

func TestValidateSlackMemberID(t *testing.T) {
	if err := ValidateSlackMemberID("U012ABCDEFGH"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateSlackMemberID("not-an-id"); err == nil {
		t.Fatal("expected error")
	}
}

func TestAddRemoveTeamMember(t *testing.T) {
	testConfigDir(t)

	if err := AddTeamMember("sugu", "U012ABCDEFGH"); err != nil {
		t.Fatal(err)
	}
	if err := AddTeamMember("sugu", "U099ZZZZZZZZ"); err == nil {
		t.Fatal("duplicate label")
	}
	if err := AddTeamMember("other", "U012ABCDEFGH"); err == nil {
		t.Fatal("duplicate id")
	}

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Team) != 1 || cfg.Team[0].Label != "sugu" {
		t.Fatalf("team: %+v", cfg.Team)
	}

	if err := RemoveTeamMember("sugu"); err != nil {
		t.Fatal(err)
	}
	cfg, _ = Load()
	if len(cfg.Team) != 0 {
		t.Fatalf("expected empty team, got %+v", cfg.Team)
	}
}
