package common

import (
	"fmt"
	"strings"
	"time"
)

// FormatBytes converts bytes to a human-readable string (KB, MB, GB, TB).
func FormatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// FormatPercent formats a percentage with 1 decimal place.
func FormatPercent(pct float64) string {
	return fmt.Sprintf("%.1f%%", pct)
}

// FormatUptime converts uptime seconds to a human string.
func FormatUptime(seconds uint64) string {
	d := time.Duration(seconds) * time.Second
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}

// RepeatString returns a string repeated n times.
func RepeatString(s string, n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(s, n)
}

// TruncateString truncates a string to maxLen, appending "..." if truncated.
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
