package charts

import (
	"fmt"
	"image/color"
)

// ChartType identifies the kind of chart.
type ChartType int

const (
	ChartLine ChartType = iota
	ChartBar
	ChartArea
	ChartGauge
	ChartSparkline
	ChartHeatMap
	ChartNumeric
)

// ChartConfig holds layout and display options for any chart type.
type ChartConfig struct {
	// Dimensions
	Width  int
	Height int

	// Labels
	Title  string
	XLabel string
	YLabel string

	// Display options
	ShowLegend bool
	ShowGrid   bool
	AutoScale  bool
	MinValue   float64 // used when AutoScale is false
	MaxValue   float64

	// Series metadata
	SeriesNames  []string
	SeriesColors []color.Color

	// Value formatting
	FormatFn func(v float64) string
}

// DefaultConfig returns a sensible ChartConfig with auto-scaling enabled.
func DefaultConfig() ChartConfig {
	return ChartConfig{
		Width:      60,
		Height:     10,
		AutoScale:  true,
		ShowGrid:   true,
		ShowLegend: false,
		FormatFn:   func(v float64) string { return formatFloat(v) },
	}
}

// DataPoint is a single value with optional label.
type DataPoint struct {
	Value float64
	Label string // optional x-axis or series label
}

// Series is a named collection of data points.
type Series struct {
	Name  string
	Data  []float64
	Color color.Color
}

// SeriesSet is a collection of named series for multi-line charts.
type SeriesSet struct {
	Series  []Series
	XLabels []string // shared x-axis labels
}

// Threshold marks a critical value on a chart.
type Threshold struct {
	Value float64
	Label string
	Color color.Color
}

// TrendDirection indicates movement direction.
type TrendDirection int

const (
	TrendStable  TrendDirection = 0
	TrendRising  TrendDirection = 1
	TrendFalling TrendDirection = -1
)

// TrendInfo describes a data series' recent trend.
type TrendInfo struct {
	Direction    TrendDirection
	ChangePct    float64 // percent change over the window
	Slope        float64 // linear regression slope
	WindowValues []float64
}

// Chart interface that all chart types implement.
type Chart interface {
	// Render produces a styled terminal string of the chart.
	Render() string
	// SetConfig updates the chart configuration.
	SetConfig(cfg ChartConfig)
	// Config returns the current configuration.
	Config() ChartConfig
}

// ── helpers ──────────────────────────────────────────────────────────────

func formatFloat(v float64) string {
	if v >= 1000 {
		return fmt.Sprintf("%.0f", v)
	}
	if v >= 100 {
		return fmt.Sprintf("%.1f", v)
	}
	return fmt.Sprintf("%.2f", v)
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
