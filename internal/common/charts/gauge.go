package charts

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// ── Gauge ──────────────────────────────────────────────────────────────────

// Gauge renders a horizontal or vertical progress gauge with color zones.
type Gauge struct {
	config        ChartConfig
	value         float64 // 0–100
	warnThreshold float64
	critThreshold float64
	label         string
	showPct       bool
}

// NewGauge creates a progress gauge.
func NewGauge(cfg ChartConfig) *Gauge {
	if cfg.FormatFn == nil {
		cfg.FormatFn = func(v float64) string { return formatFloat(v) }
	}
	return &Gauge{
		config:        cfg,
		warnThreshold: 70,
		critThreshold: 90,
		showPct:       true,
	}
}

// SetValue sets the gauge value (clamped to 0–100).
func (g *Gauge) SetValue(pct float64) {
	g.value = clamp(pct, 0, 100)
}

// SetThresholds sets the warning and critical thresholds.
func (g *Gauge) SetThresholds(warn, crit float64) {
	g.warnThreshold = clamp(warn, 0, 100)
	g.critThreshold = clamp(crit, 0, 100)
}

// SetLabel sets a text label below the gauge.
func (g *Gauge) SetLabel(text string) {
	g.label = text
}

// SetConfig updates the chart configuration.
func (g *Gauge) SetConfig(cfg ChartConfig) {
	if cfg.FormatFn == nil {
		cfg.FormatFn = func(v float64) string { return formatFloat(v) }
	}
	g.config = cfg
}

// Config returns the current configuration.
func (g *Gauge) Config() ChartConfig { return g.config }

// Render returns the styled gauge as a string.
func (g *Gauge) Render() string {
	var b strings.Builder

	// Gauge color based on value relative to thresholds
	gaugeColor := zoneColor(g.value, g.warnThreshold, g.critThreshold)

	// First line: gauge bar
	barWidth := float64(g.config.Width - 6) // room for " 100%"
	if barWidth < 10 {
		barWidth = 10
	}

	// Block fill characters for the gauge
	fillChars := "▓"
	emptyChars := "░"

	filled := int((g.value / 100.0) * barWidth)
	if filled > int(barWidth) {
		filled = int(barWidth)
	}

	var bar strings.Builder
	for i := 0; i < filled; i++ {
		bar.WriteString(fillChars)
	}
	remaining := int(barWidth) - filled
	if remaining < 0 {
		remaining = 0
	}
	for i := 0; i < remaining; i++ {
		bar.WriteString(emptyChars)
	}

	// Gauge line with optional label
	if g.label != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Render(g.label + " "))
	}
	b.WriteString(lipgloss.NewStyle().Foreground(gaugeColor).Render(bar.String()))

	if g.showPct {
		pctStr := fmt.Sprintf(" %.0f%%", g.value)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Render(pctStr))
	}
	b.WriteByte('\n')

	return b.String()
}

// zoneColor returns the appropriate color based on thresholds.
func zoneColor(value, warn, crit float64) color.Color {
	switch {
	case value >= crit:
		return lipgloss.Color("#EF4444") // red
	case value >= warn:
		return lipgloss.Color("#F59E0B") // amber
	default:
		return lipgloss.Color("#10B981") // green
	}
}
