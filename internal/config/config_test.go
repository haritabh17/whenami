package config

import "testing"

func TestDisplayDefaults(t *testing.T) {
	c := &Config{}
	ApplyDefaults(c)
	if !ShowAvatar(c) || ShowName(c) || !ShowTime(c) {
		t.Fatalf("defaults: avatar=%v name=%v time=%v", ShowAvatar(c), ShowName(c), ShowTime(c))
	}
	if TimePrecision(c) != "minutes" {
		t.Fatalf("precision %q", TimePrecision(c))
	}
}

func TestValidateDisplayRejectsAllHidden(t *testing.T) {
	f := false
	c := &Config{ShowAvatar: &f, ShowName: &f, ShowTime: &f}
	if err := ValidateDisplay(c); err == nil {
		t.Fatal("expected error")
	}
}

func TestTimePrecisionHours(t *testing.T) {
	h := "hours"
	c := &Config{TimePrecision: h}
	ApplyDefaults(c)
	if TimePrecision(c) != "hours" {
		t.Fatalf("got %q", TimePrecision(c))
	}
}
