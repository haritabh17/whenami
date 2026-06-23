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
	if IconSize(c) != IconSizeDefault {
		t.Fatalf("icon size %d", IconSize(c))
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

func TestIconSizeClampedToSupportedRange(t *testing.T) {
	small := 1
	c := &Config{IconSize: &small}
	ApplyDefaults(c)
	if IconSize(c) != IconSizeMin {
		t.Fatalf("small icon size got %d want %d", IconSize(c), IconSizeMin)
	}

	large := 99
	c = &Config{IconSize: &large}
	ApplyDefaults(c)
	if IconSize(c) != IconSizeMax {
		t.Fatalf("large icon size got %d want %d", IconSize(c), IconSizeMax)
	}
}
