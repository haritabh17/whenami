package timeformat

import (
	"testing"
	"time"
)

func TestFormatClock12hMinutes(t *testing.T) {
	at := time.Date(2026, 6, 13, 16, 7, 0, 0, time.UTC)
	got := FormatClock("UTC", false, PrecisionMinutes, at)
	if got != "4.07 PM" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatClock12hHours(t *testing.T) {
	at := time.Date(2026, 6, 13, 16, 7, 0, 0, time.UTC)
	got := FormatClock("UTC", false, PrecisionHours, at)
	if got != "4 PM" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatClock24hMinutes(t *testing.T) {
	at := time.Date(2026, 6, 13, 16, 7, 0, 0, time.UTC)
	got := FormatClock("UTC", true, PrecisionMinutes, at)
	if got != "16.07" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatClock24hHours(t *testing.T) {
	at := time.Date(2026, 6, 13, 16, 7, 0, 0, time.UTC)
	got := FormatClock("UTC", true, PrecisionHours, at)
	if got != "16" {
		t.Fatalf("got %q", got)
	}
}
