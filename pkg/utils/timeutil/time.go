package timeutil

import "time"

// IsToday is ...
func IsToday(t time.Time) bool {
	now := time.Now().Truncate(24 * time.Hour)
	t = t.Truncate(24 * time.Hour)
	return now.Equal(t)
}

// DaysBetween is ...
func DaysBetween(start, end time.Time) int {
	start = start.Truncate(24 * time.Hour)
	end = end.Truncate(24 * time.Hour)
	duration := end.Sub(start)
	return int(duration.Hours() / 24)
}
