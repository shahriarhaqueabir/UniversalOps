package common

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// Color palette constants.
const (
	ColorPrimary   = "#7C3AED" // Purple
	ColorSecondary = "#10B981" // Green
	ColorWarning   = "#F59E0B" // Amber
	ColorDanger    = "#EF4444" // Red
	ColorInfo      = "#3B82F6" // Blue
	ColorMuted     = "#6B7280" // Gray
	ColorText      = "#F3F4F6" // Light gray
	ColorDarkBg    = "#1F2937" // Dark background
	ColorBorder    = "#374151" // Border
)

// Common styles shared across all ops layers.
var (
	// PanelTitle is the standard section title style.
	PanelTitle lipgloss.Style

	// Label is used for field labels / descriptions.
	Label lipgloss.Style

	// Value is used for highlighted data values.
	Value lipgloss.Style

	// Muted is for secondary / less important text.
	Muted lipgloss.Style

	// Error is for error messages.
	Error lipgloss.Style

	// Panel frames content in a rounded border.
	Panel lipgloss.Style

	// Output is for monospace / command output text.
	Output lipgloss.Style

	// Section is for subsection headers (like DNS record types).
	Section lipgloss.Style
)

// Health colors using image/color.Color for lipgloss v2 compatibility.
var (
	HealthyColor color.Color
	WarningColor color.Color
	DangerColor  color.Color
	InfoColor    color.Color
)

func init() {
	rebuildCommonStyles(CurrentPalette())
}

func rebuildCommonStyles(p Palette) {
	PanelTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Primary)).
		Bold(true).
		Padding(0, 0, 1, 0)

	Label = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Muted))

	Value = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Secondary)).
		Bold(true)

	Muted = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Muted))

	Error = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Danger))

	Panel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(p.Border)).
		Padding(1, 2)

	Output = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Text))

	Section = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Info)).
		Bold(true)

	HealthyColor = lipgloss.Color(p.Secondary)
	WarningColor = lipgloss.Color(p.Warning)
	DangerColor = lipgloss.Color(p.Danger)
	InfoColor = lipgloss.Color(p.Info)
}

// RenderSparkline renders a compact ASCII-only trend line.
func RenderSparkline(values []float64, width int) string {
	if width <= 0 {
		return ""
	}
	if len(values) == 0 {
		return strings.Repeat(".", width)
	}

	window := values
	if len(window) > width {
		window = window[len(window)-width:]
	}

	maxValue := 0.0
	for _, value := range window {
		if value > maxValue {
			maxValue = value
		}
	}

	levels := "._-=+#"
	var b strings.Builder
	for i := 0; i < width-len(window); i++ {
		b.WriteByte('.')
	}
	for _, value := range window {
		if maxValue == 0 || value <= 0 {
			b.WriteByte(levels[0])
			continue
		}
		idx := int((value / maxValue) * float64(len(levels)-1))
		if idx >= len(levels) {
			idx = len(levels) - 1
		}
		b.WriteByte(levels[idx])
	}
	return b.String()
}

// GetHealthColor returns a color based on the percentage (green/yellow/red).
func GetHealthColor(pct float64) color.Color {
	switch {
	case pct > 90:
		return DangerColor
	case pct > 70:
		return WarningColor
	default:
		return HealthyColor
	}
}
