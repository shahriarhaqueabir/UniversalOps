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

// CleanJSON removes control characters and BOM from JSON strings
// that can cause json.Unmarshal to fail on PowerShell output.
func CleanJSON(s string) string {
	// Remove BOM
	s = strings.TrimLeft(s, "\ufeff\u00ef\u00bb\u00bf")
	// Remove ASCII control characters except \t, \n, \r
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= 32 || r == '\t' || r == '\n' || r == '\r' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// FixPowerShellDashes replaces standalone "-" values with "\"\"" (empty string in JSON)
// to handle PowerShell's use of dash for unset/empty fields.
// This is needed because PowerShell may output "FieldName": -  as a null placeholder.
func FixPowerShellDashes(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		if s[i] == ':' {
			b.WriteByte(':')
			i++
			// Skip spaces after colon
			j := i
			for j < len(s) && (s[j] == ' ' || s[j] == '\t') {
				j++
			}
			// Check if followed by a dash then comma, space, brace, or newline
			if j < len(s) && s[j] == '-' {
				k := j + 1
				// Skip spaces after dash
				for k < len(s) && (s[k] == ' ' || s[k] == '\t') {
					k++
				}
				// Dash is a null placeholder if followed by comma, brace, or newline
				if k >= len(s) || s[k] == ',' || s[k] == '}' || s[k] == '\n' || s[k] == '\r' {
					// Write the spaces between colon and dash
					b.WriteString(s[i:j])
					b.WriteString("\"\"")
					i = j + 1
					continue
				}
			}
			// Not a dash value, write the spaces we skipped
			b.WriteString(s[i:j])
			i = j
		} else {
			b.WriteByte(s[i])
			i++
		}
	}
	return b.String()
}
