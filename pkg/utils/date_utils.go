package utils

import "time"

// DaysBetween calculates days between two dates
func DaysBetween(start, end time.Time) int {
	duration := end.Sub(start)
	return int(duration.Hours() / 24)
}

