package charts

import (
	"fmt"
	"image/color"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
)

// ── BarChart ───────────────────────────────────────────────────────────────

// BarChart renders horizontal bars using Unicode block characters.
type BarChart struct {
	config ChartConfig
	bars   []barEntry
	groups []barGroup
	sorted bool
}

type barEntry struct {
	label string
	value float64
	color color.Color
}

type barGroup struct {
	label  string
	values []float64
	colors []color.Color
}

// NewBarChart creates a horizontal bar chart.
func NewBarChart(cfg ChartConfig) *BarChart {
	if cfg.FormatFn == nil {
		cfg.FormatFn = func(v float64) string { return formatFloat(v) }
	}
	return &BarChart{config: cfg}
}

// AddBar adds a single bar with a label, value, and color.
func (bc *BarChart) AddBar(label string, value float64, c color.Color) {
	bc.bars = append(bc.bars, barEntry{label: label, value: value, color: c})
}

// AddGroup adds a grouped/stacked bar with multiple segments.
func (bc *BarChart) AddGroup(label string, values []float64, colors []color.Color) {
	bc.groups = append(bc.groups, barGroup{label: label, values: values, colors: colors})
}

// SortByValue sorts single bars by value (ascending if true).
func (bc *BarChart) SortByValue(asc bool) {
	bc.sorted = true
	sort.Slice(bc.bars, func(i, j int) bool {
		if asc {
			return bc.bars[i].value < bc.bars[j].value
		}
		return bc.bars[i].value > bc.bars[j].value
	})
}

// SetConfig updates the chart configuration.
func (bc *BarChart) SetConfig(cfg ChartConfig) {
	if cfg.FormatFn == nil {
		cfg.FormatFn = func(v float64) string { return formatFloat(v) }
	}
	bc.config = cfg
}

// Config returns the current configuration.
func (bc *BarChart) Config() ChartConfig { return bc.config }

// Render returns the styled chart as a string.
func (bc *BarChart) Render() string {
	if len(bc.bars) == 0 && len(bc.groups) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("No data")
	}

	var b strings.Builder

	if bc.config.Title != "" {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render(bc.config.Title))
		b.WriteByte('\n')
	}

	// Find the longest label for alignment
	maxLabelLen := 0
	for _, bar := range bc.bars {
		if len(bar.label) > maxLabelLen {
			maxLabelLen = len(bar.label)
		}
	}
	for _, g := range bc.groups {
		if len(g.label) > maxLabelLen {
			maxLabelLen = len(g.label)
		}
	}
	if maxLabelLen > 20 {
		maxLabelLen = 20
	}

	// Compute the bar width (config width minus label and value)
	barWidth := float64(bc.config.Width - maxLabelLen - 12)
	if barWidth < 5 {
		barWidth = 5
	}

	// Find global max
	maxVal := 0.0
	for _, bar := range bc.bars {
		if bar.value > maxVal {
			maxVal = bar.value
		}
	}
	for _, g := range bc.groups {
		sum := 0.0
		for _, v := range g.values {
			sum += v
		}
		if sum > maxVal {
			maxVal = sum
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}

	// Render single bars
	for _, bar := range bc.bars {
		label := fmt.Sprintf("%-*s", maxLabelLen, truncate(bar.label, maxLabelLen))
		barStr := BlockBar(bar.value, maxVal, barWidth)
		valueStr := bc.config.FormatFn(bar.value)

		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(label))
		b.WriteString(" ")
		b.WriteString(lipgloss.NewStyle().Foreground(bar.color).Render(barStr))
		b.WriteString(" ")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Render(valueStr))
		b.WriteByte('\n')
	}

	// Render groups
	for _, g := range bc.groups {
		label := fmt.Sprintf("%-*s", maxLabelLen, truncate(g.label, maxLabelLen))
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(label))
		b.WriteString(" ")

		// Each segment of the group
		for i, v := range g.values {
			segWidth := (v / maxVal) * barWidth
			if segWidth < 1 && v > 0 {
				segWidth = 1
			}
			segStr := BlockBar(v, maxVal, segWidth)
			c := lipgloss.Color("#888888")
			if i < len(g.colors) {
				c = g.colors[i]
			}
			b.WriteString(lipgloss.NewStyle().Foreground(c).Render(segStr))
		}

		// Total value
		sum := 0.0
		for _, v := range g.values {
			sum += v
		}
		valueStr := bc.config.FormatFn(sum)
		b.WriteString(" ")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Render(valueStr))
		b.WriteByte('\n')
	}

	return b.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}
