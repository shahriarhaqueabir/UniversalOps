package charts

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// DefaultSeriesColors returns a palette of default chart line colors.
func DefaultSeriesColors() []color.Color {
	return []color.Color{
		lipgloss.Color("#7C3AED"), // Purple
		lipgloss.Color("#10B981"), // Green
		lipgloss.Color("#3B82F6"), // Blue
		lipgloss.Color("#F59E0B"), // Amber
		lipgloss.Color("#EF4444"), // Red
		lipgloss.Color("#EC4899"), // Pink
	}
}

// ConfigOption is a function that modifies a ChartConfig.
type ConfigOption func(*ChartConfig)

// ApplyConfig applies configuration options and validates the result.
func ApplyConfig(cfg *ChartConfig, opts ...ConfigOption) {
	for _, opt := range opts {
		opt(cfg)
	}
	validate(cfg)
}

func validate(cfg *ChartConfig) {
	if cfg.Width <= 0 {
		cfg.Width = 60
	}
	if cfg.Height <= 0 {
		cfg.Height = 10
	}
	if cfg.SeriesColors == nil {
		cfg.SeriesColors = DefaultSeriesColors()
	}
	if cfg.FormatFn == nil {
		cfg.FormatFn = func(v float64) string { return formatFloat(v) }
	}
}

// WithTitle sets the chart title.
func WithTitle(title string) ConfigOption {
	return func(cfg *ChartConfig) {
		cfg.Title = title
	}
}

// WithDimensions sets the chart width and height.
func WithDimensions(w, h int) ConfigOption {
	return func(cfg *ChartConfig) {
		cfg.Width = w
		cfg.Height = h
	}
}

// WithLabels sets the X and Y axis labels.
func WithLabels(x, y string) ConfigOption {
	return func(cfg *ChartConfig) {
		cfg.XLabel = x
		cfg.YLabel = y
	}
}

// WithLegend enables or disables the chart legend.
func WithLegend(show bool) ConfigOption {
	return func(cfg *ChartConfig) {
		cfg.ShowLegend = show
	}
}

// WithGrid enables or disables the chart grid.
func WithGrid(show bool) ConfigOption {
	return func(cfg *ChartConfig) {
		cfg.ShowGrid = show
	}
}

// WithFixedRange disables auto-scaling and sets fixed min/max values.
func WithFixedRange(min, max float64) ConfigOption {
	return func(cfg *ChartConfig) {
		cfg.AutoScale = false
		cfg.MinValue = min
		cfg.MaxValue = max
	}
}

// WithSeriesColors sets the series color palette.
func WithSeriesColors(colors []color.Color) ConfigOption {
	return func(cfg *ChartConfig) {
		cfg.SeriesColors = colors
	}
}
