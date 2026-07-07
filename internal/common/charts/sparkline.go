package charts

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
)

// ── Sparkline ──────────────────────────────────────────────────────────────

// Sparkline renders a compact single-line braille trend with optional stats.
type Sparkline struct {
	width      int
	data       []float64
	showMinMax bool
	showTrend  bool
}

// NewSparkline creates a compact sparkline with the given braille width.
func NewSparkline(width int) *Sparkline {
	if width < 2 {
		width = 2
	}
	return &Sparkline{width: width}
}

// SetValues sets the data to render.
func (s *Sparkline) SetValues(data []float64) {
	s.data = data
}

// ShowMinMax toggles min/max marker labels.
func (s *Sparkline) ShowMinMax(show bool) {
	s.showMinMax = show
}

// ShowTrend toggles trend arrow and percentage.
func (s *Sparkline) ShowTrend(show bool) {
	s.showTrend = show
}

// SetConfig is a no-op for sparklines (embedded in other views).
func (s *Sparkline) SetConfig(cfg ChartConfig) {}

// Config returns a minimal ChartConfig.
func (s *Sparkline) Config() ChartConfig {
	return ChartConfig{Width: s.width, Height: 1}
}

// Render returns the sparkline as a string (1–2 lines).
func (s *Sparkline) Render() string {
	if len(s.data) == 0 {
		return ""
	}

	// Use braille rendering for the trend line
	brailleLine := BrailleLine(s.data, s.width, 1)
	if brailleLine == "" {
		return ""
	}

	// Trend indicator
	trendStr := ""
	if s.showTrend && len(s.data) >= 2 {
		first := s.data[0]
		last := s.data[len(s.data)-1]
		change := last - first

		var arrow string
		var arrowColor color.Color
		if change > 0 {
			arrow = "↑"
			arrowColor = lipgloss.Color("#10B981") // green
		} else if change < 0 {
			arrow = "↓"
			arrowColor = lipgloss.Color("#EF4444") // red
		} else {
			arrow = "→"
			arrowColor = lipgloss.Color("#888888") // gray
		}

		pct := 0.0
		if first != 0 {
			pct = math.Abs(change) / math.Abs(first) * 100
		}

		trendStr = " " + lipgloss.NewStyle().Foreground(arrowColor).Render(
			fmt.Sprintf("%s %.1f%%", arrow, pct))
	}

	var b strings.Builder
	b.WriteString(brailleLine)
	b.WriteString(trendStr)
	b.WriteByte('\n')

	// Min/max labels
	if s.showMinMax {
		minV, maxV := minMaxFloat(s.data)

		// Compute avg
		sum := 0.0
		for _, v := range s.data {
			sum += v
		}
		avg := sum / float64(len(s.data))

		label := fmt.Sprintf("min: %s  max: %s  avg: %s",
			formatFloat(minV), formatFloat(maxV), formatFloat(avg))
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(label))
		b.WriteByte('\n')
	}

	return b.String()
}
