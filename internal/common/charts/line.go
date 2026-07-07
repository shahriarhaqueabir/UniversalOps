package charts

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
)

// ── LineChart ──────────────────────────────────────────────────────────────

// LineChart renders one or more series as a braille line chart with axes.
type LineChart struct {
	config     ChartConfig
	series     []lineSeries
	forecast   []float64
	thresholds []Threshold
}

type lineSeries struct {
	name  string
	data  []float64
	color color.Color
}

// NewLineChart creates a braille line chart.
func NewLineChart(cfg ChartConfig) *LineChart {
	if cfg.FormatFn == nil {
		cfg.FormatFn = func(v float64) string { return formatFloat(v) }
	}
	return &LineChart{config: cfg}
}

// AddSeries adds a named data series with a line color.
func (lc *LineChart) AddSeries(name string, data []float64, c color.Color) {
	lc.series = append(lc.series, lineSeries{name: name, data: data, color: c})
}

// SetForecast overlays a forecast series as a dashed/muted region.
func (lc *LineChart) SetForecast(data []float64) {
	lc.forecast = data
}

// SetThreshold adds a horizontal threshold line with an optional label.
func (lc *LineChart) SetThreshold(value float64, label string) {
	lc.thresholds = append(lc.thresholds, Threshold{
		Value: value,
		Label: label,
		Color: lipgloss.Color("#EF4444"), // default red
	})
}

// SetConfig updates the chart configuration.
func (lc *LineChart) SetConfig(cfg ChartConfig) {
	if cfg.FormatFn == nil {
		cfg.FormatFn = func(v float64) string { return formatFloat(v) }
	}
	lc.config = cfg
}

// Config returns the current configuration.
func (lc *LineChart) Config() ChartConfig { return lc.config }

// Render returns the styled chart as a string.
func (lc *LineChart) Render() string {
	if len(lc.series) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("No data")
	}

	// ── Compute range ──
	minVal, maxVal := lc.computeRange()
	if maxVal == minVal {
		maxVal = minVal + 1
	}

	// ── Layout measurements ──
	yLabels := RenderAxis(minVal, maxVal, lc.config.Height)
	labelWidth := 0
	for _, l := range yLabels {
		if len(l) > labelWidth {
			labelWidth = len(l)
		}
	}
	if labelWidth < 3 {
		labelWidth = 3
	}

	chartW := lc.config.Width - labelWidth - 2
	if chartW < 4 {
		chartW = 4
	}

	// ── Render braille grid (all series combined) ──
	seriesList := make([]Series, len(lc.series))
	for i, s := range lc.series {
		seriesList[i] = Series{Name: s.name, Data: s.data, Color: s.color}
	}
	brailleStr := BrailleLineMulti(seriesList, chartW, lc.config.Height)
	brailleLines := strings.Split(brailleStr, "\n")

	for len(brailleLines) < lc.config.Height {
		brailleLines = append(brailleLines, strings.Repeat(" ", chartW))
	}

	// ── Compute threshold row positions ──
	thresholdRowLabels := make(map[int]string)
	for _, t := range lc.thresholds {
		norm := clamp((t.Value-minVal)/(maxVal-minVal), 0, 1)
		row := int(math.Round(norm * float64(lc.config.Height-1)))
		thresholdRowLabels[row] = lipgloss.NewStyle().Foreground(t.Color).Render("─── " + t.Label)
	}

	// ── Compose output ──
	var b strings.Builder

	// Title
	if lc.config.Title != "" {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render(lc.config.Title))
		b.WriteByte('\n')
	}

	// Legend
	if lc.config.ShowLegend && len(lc.series) > 0 {
		legendParts := make([]string, 0, len(lc.series))
		for _, s := range lc.series {
			dot := lipgloss.NewStyle().Foreground(s.color).Render("■")
			legendParts = append(legendParts, fmt.Sprintf("%s %s", dot, s.name))
		}
		b.WriteString(strings.Join(legendParts, "  "))
		b.WriteByte('\n')
	}

	// Chart body: Y-axis labels + braille + threshold labels
	for i := 0; i < lc.config.Height; i++ {
		label := ""
		if i < len(yLabels) {
			label = fmt.Sprintf("%*s", labelWidth, yLabels[i])
		} else {
			label = strings.Repeat(" ", labelWidth)
		}

		line := ""
		if i < len(brailleLines) {
			line = brailleLines[i]
		}

		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(label))
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("┤"))
		b.WriteString(line)

		if tLabel, ok := thresholdRowLabels[i]; ok {
			b.WriteString(" ")
			b.WriteString(tLabel)
		}

		b.WriteByte('\n')
	}

	// X-axis label
	if lc.config.XLabel != "" {
		padding := labelWidth + 1
		b.WriteString(strings.Repeat(" ", padding))
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(lc.config.XLabel))
		b.WriteByte('\n')
	}

	return b.String()
}

func (lc *LineChart) computeRange() (float64, float64) {
	var minVal, maxVal float64
	if lc.config.AutoScale {
		first := true
		for _, s := range lc.series {
			if len(s.data) == 0 {
				continue
			}
			sMin, sMax := minMaxFloat(s.data)
			if first || sMin < minVal {
				minVal = sMin
			}
			if first || sMax > maxVal {
				maxVal = sMax
			}
			first = false
		}
		if len(lc.forecast) > 0 {
			fMin, fMax := minMaxFloat(lc.forecast)
			if fMin < minVal {
				minVal = fMin
			}
			if fMax > maxVal {
				maxVal = fMax
			}
		}
		for _, t := range lc.thresholds {
			if t.Value < minVal {
				minVal = t.Value
			}
			if t.Value > maxVal {
				maxVal = t.Value
			}
		}
	} else {
		minVal = lc.config.MinValue
		maxVal = lc.config.MaxValue
	}
	return minVal, maxVal
}
