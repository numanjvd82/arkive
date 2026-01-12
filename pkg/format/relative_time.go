package format

import (
	"strconv"
	"time"
)

func RelativeTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	now := time.Now()
	if value.After(now) {
		return value.Format("Jan 2, 2006 15:04")
	}
	diff := now.Sub(value)
	switch {
	case diff < time.Minute:
		return "Just now"
	case diff < time.Hour:
		return formatMinutes(diff)
	case diff < 24*time.Hour:
		return formatHours(diff)
	case diff < 7*24*time.Hour:
		return formatDays(diff)
	case diff < 30*24*time.Hour:
		return formatWeeks(diff)
	default:
		return value.Format("Jan 2, 2006")
	}
}

func formatMinutes(diff time.Duration) string {
	return formatUnit(int(diff.Minutes()), "m")
}

func formatHours(diff time.Duration) string {
	return formatUnit(int(diff.Hours()), "h")
}

func formatDays(diff time.Duration) string {
	return formatUnit(int(diff.Hours()/24), "d")
}

func formatWeeks(diff time.Duration) string {
	return formatUnit(int(diff.Hours()/(24*7)), "w")
}

func formatUnit(value int, suffix string) string {
	return strconv.Itoa(value) + suffix + " ago"
}
