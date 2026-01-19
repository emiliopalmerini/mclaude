package util

import (
	"fmt"
	"time"
)

// FormatNumber formats an int64 with K/M suffix for readability.
// Examples: 500 -> "500", 1500 -> "1.5K", 1500000 -> "1.5M"
func FormatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%.1fM", float64(n)/1000000)
}

// FormatTokens formats a float64 token count with K/M suffix for readability.
// Examples: 500 -> "500", 1500 -> "1.5K", 1500000 -> "1.5M"
func FormatTokens(n float64) string {
	if n < 1000 {
		return fmt.Sprintf("%.0f", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%.1fK", n/1000)
	}
	return fmt.Sprintf("%.1fM", n/1000000)
}

// FormatTokensInt formats an int64 token count with K/M suffix for readability.
// Examples: 500 -> "500", 1500 -> "1.5K", 1500000 -> "1.5M"
func FormatTokensInt(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%.1fM", float64(n)/1000000)
}

// FormatDateISO formats an RFC3339 timestamp string to ISO date format (2006-01-02).
// Returns the original string if parsing fails.
func FormatDateISO(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.Format("2006-01-02")
}

// FormatDateHuman formats an RFC3339 timestamp string to human-readable format (Jan 2, 2006).
// Returns the original string if parsing fails.
func FormatDateHuman(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.Format("Jan 2, 2006")
}

// FormatDateTime formats an RFC3339 timestamp string to date-time format (2006-01-02 15:04).
// Returns the original string if parsing fails.
func FormatDateTime(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.Format("2006-01-02 15:04")
}

// ParseTimeRFC3339 parses an RFC3339 timestamp string to time.Time.
// Returns zero time if parsing fails.
func ParseTimeRFC3339(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

// ParseTimeSQLite parses a SQLite datetime or RFC3339 string to time.Time.
// Handles "YYYY-MM-DD HH:MM:SS" (SQLite) and RFC3339 formats.
// Returns zero time if parsing fails.
func ParseTimeSQLite(s string) time.Time {
	// Try SQLite datetime format first (most common from DB)
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return t
	}
	// Fall back to RFC3339
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

// GetStartDateForPeriod returns the start date for a given period as an RFC3339 string.
// Supported periods: "today", "week", "month", "all" (or any other value for all time).
func GetStartDateForPeriod(period string) string {
	now := time.Now().UTC()
	var start time.Time

	switch period {
	case "today":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start = time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, time.UTC)
	case "month":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	default:
		start = time.Unix(0, 0)
	}

	return start.Format(time.RFC3339)
}
