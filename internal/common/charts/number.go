package charts

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// ── NumericDisplay ─────────────────────────────────────────────────────────

// NumericDisplay renders a large centered value with trend and unit.
type NumericDisplay struct {
	config ChartConfig
	value  float64
	unit   string
	trend  TrendInfo
	color  color.Color
}

// NewNumericDisplay creates a numeric display card.
func NewNumericDisplay(cfg ChartConfig) *NumericDisplay {
	if cfg.FormatFn == nil {
		cfg.FormatFn = func(v float64) string { return formatFloat(v) }
	}
	return &NumericDisplay{
		config: cfg,
		color:  lipgloss.Color("#10B981"), // default green
	}
}

// SetValue sets the displayed value and optional unit.
func (nd *NumericDisplay) SetValue(value float64, unit string) {
	nd.value = value
	nd.unit = unit
}

// SetTrend sets the trend arrow and percentage change.
func (nd *NumericDisplay) SetTrend(dir TrendDirection, pct float64) {
	nd.trend = TrendInfo{Direction: dir, ChangePct: pct}
}

// SetColor sets the value color.
func (nd *NumericDisplay) SetColor(c color.Color) {
	nd.color = c
}

// SetConfig updates the chart configuration.
func (nd *NumericDisplay) SetConfig(cfg ChartConfig) {
	if cfg.FormatFn == nil {
		cfg.FormatFn = func(v float64) string { return formatFloat(v) }
	}
	nd.config = cfg
}

// Config returns the current configuration.
func (nd *NumericDisplay) Config() ChartConfig { return nd.config }

// Render returns the numeric display as a styled string.
func (nd *NumericDisplay) Render() string {
	width := nd.config.Width
	if width < 10 {
		width = 10
	}

	var b strings.Builder

	// ── Top border ──
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("┌" + strings.Repeat("─", width-2) + "┐"))
	b.WriteByte('\n')

	// ── Value line ──
	valueStr := nd.config.FormatFn(nd.value)
	if nd.unit != "" {
		valueStr += nd.unit
	}
	centeredVal := centerText(valueStr, width-2)
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("│"))
	b.WriteString(lipgloss.NewStyle().Foreground(nd.color).Bold(true).Render(centeredVal))
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("│"))
	b.WriteByte('\n')

	// ── Unit/name line ──
	if nd.config.Title != "" {
		centeredTitle := centerText(nd.config.Title, width-2)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("│"))
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(centeredTitle))
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("│"))
		b.WriteByte('\n')
	}

	// ── Trend line ──
	if nd.trend.Direction != TrendStable {
		var arrow string
		var arrowColor color.Color
		switch nd.trend.Direction {
		case TrendRising:
			arrow = "↑"
			arrowColor = lipgloss.Color("#10B981")
		case TrendFalling:
			arrow = "↓"
			arrowColor = lipgloss.Color("#EF4444")
		}
		trendStr := fmt.Sprintf("%s %.1f%%", arrow, nd.trend.ChangePct)
		centeredTrend := centerText(trendStr, width-2)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("│"))
		b.WriteString(lipgloss.NewStyle().Foreground(arrowColor).Render(centeredTrend))
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("│"))
		b.WriteByte('\n')
	}

	// ── Bottom border ──
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("└" + strings.Repeat("─", width-2) + "┘"))

	return b.String()
}
