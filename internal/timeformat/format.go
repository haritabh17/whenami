package timeformat

import (
	"strings"
	"time"
)

const (
	PrecisionMinutes = "minutes"
	PrecisionHours   = "hours"
)

// FormatClock returns the local time in tz using 12h/24h and hour/minute precision.
func FormatClock(tz string, format24h bool, precision string, at time.Time) string {
	if at.IsZero() {
		at = time.Now()
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	local := at.In(loc)
	if precision == PrecisionHours {
		if format24h {
			return local.Format("15")
		}
		return local.Format("3 PM")
	}
	if format24h {
		return local.Format("15.04")
	}
	return strings.Replace(local.Format("3:04 PM"), ":", ".", 1)
}

// TZLabelFromIANA turns "America/Los_Angeles" into "Los Angeles".
func TZLabelFromIANA(tz string) string {
	part := tz
	if i := strings.LastIndex(tz, "/"); i >= 0 {
		part = tz[i+1:]
	}
	return strings.ReplaceAll(part, "_", " ")
}
