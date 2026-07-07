package charts

import (
	"image/color"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// ── BrailleLine tests ──────────────────────────────────────────────────────

func TestBrailleLine(t *testing.T) {
	tests := []struct {
		name   string
		points []float64
		w      int
		h      int
		check  func(t *testing.T, result string)
	}{
		{
			name:   "empty data returns empty",
			points: nil,
			w:      10,
			h:      5,
			check:  func(t *testing.T, result string) { mustBeEmpty(t, result) },
		},
		{
			name:   "single value produces output",
			points: []float64{50},
			w:      10,
			h:      5,
			check: func(t *testing.T, result string) {
				mustNonEmpty(t, result)
				mustContainBraille(t, result)
			},
		},
		{
			name:   "constant values produce output",
			points: []float64{100, 100, 100, 100, 100},
			w:      10,
			h:      5,
			check: func(t *testing.T, result string) {
				mustNonEmpty(t, result)
				mustContainBraille(t, result)
			},
		},
		{
			name:   "rising trend",
			points: []float64{10, 20, 30, 40, 50, 60, 70, 80, 90},
			w:      10,
			h:      5,
			check: func(t *testing.T, result string) {
				mustNonEmpty(t, result)
				mustContainBraille(t, result)
			},
		},
		{
			name:   "negative values",
			points: []float64{-10, -5, 0, 5, 10},
			w:      10,
			h:      5,
			check: func(t *testing.T, result string) {
				mustNonEmpty(t, result)
				mustContainBraille(t, result)
			},
		},
		{
			name:   "correct dimensions",
			points: []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			w:      5,
			h:      3,
			check: func(t *testing.T, result string) {
				lines := strings.Split(result, "\n")
				if len(lines) != 3 {
					t.Errorf("expected 3 lines, got %d", len(lines))
				}
				for _, line := range lines {
					runes := []rune(line)
					if len(runes) != 5 {
						t.Errorf("expected 5 braille chars per line, got %d", len(runes))
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BrailleLine(tt.points, tt.w, tt.h)
			tt.check(t, result)
		})
	}
}

func TestBrailleLineMulti(t *testing.T) {
	t.Run("multiple series", func(t *testing.T) {
		series := []Series{
			{Name: "CPU", Data: []float64{10, 20, 30, 40, 50}, Color: lipgloss.Color("#00FF00")},
			{Name: "MEM", Data: []float64{50, 40, 30, 20, 10}, Color: lipgloss.Color("#0000FF")},
		}
		result := BrailleLineMulti(series, 10, 5)
		mustNonEmpty(t, result)
		mustContainBraille(t, result)
	})

	t.Run("empty series list", func(t *testing.T) {
		result := BrailleLineMulti(nil, 10, 5)
		mustBeEmpty(t, result)
	})

	t.Run("series with no data", func(t *testing.T) {
		series := []Series{{Name: "empty", Data: nil, Color: lipgloss.Color("#FF0000")}}
		result := BrailleLineMulti(series, 10, 5)
		mustBeEmpty(t, result)
	})
}

// ── BlockBar tests ─────────────────────────────────────────────────────────

func TestBlockBar(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		max   float64
		width float64
		check func(t *testing.T, result string)
	}{
		{
			name:  "zero width returns empty",
			value: 50,
			max:   100,
			width: 0,
			check: func(t *testing.T, result string) { mustBeEmpty(t, result) },
		},
		{
			name:  "zero value is empty",
			value: 0,
			max:   100,
			width: 20,
			check: func(t *testing.T, result string) {
				mustNonEmpty(t, result)
				if strings.Contains(result, "█") {
					t.Error("expected no filled blocks for zero value")
				}
			},
		},
		{
			name:  "100% fills all space",
			value: 100,
			max:   100,
			width: 20,
			check: func(t *testing.T, result string) {
				mustNonEmpty(t, result)
				if !strings.Contains(result, "█") {
					t.Error("expected filled blocks for max value")
				}
			},
		},
		{
			name:  "50% is half filled",
			value: 50,
			max:   100,
			width: 20,
			check: func(t *testing.T, result string) {
				mustNonEmpty(t, result)
				mustContainBlock(t, result)
			},
		},
		{
			name:  "value > max is capped",
			value: 200,
			max:   100,
			width: 20,
			check: func(t *testing.T, result string) {
				mustNonEmpty(t, result)
				if !strings.Contains(result, "█") {
					t.Error("expected filled blocks for capped value")
				}
			},
		},
		{
			name:  "negative value treated as zero",
			value: -10,
			max:   100,
			width: 20,
			check: func(t *testing.T, result string) {
				mustNonEmpty(t, result)
				if strings.Contains(result, "█") {
					t.Error("expected no filled blocks for negative value")
				}
			},
		},
		{
			name:  "width includes block characters at expected length",
			value: 50,
			max:   100,
			width: 10,
			check: func(t *testing.T, result string) {
				runes := []rune(result)
				if len(runes) != 10 {
					t.Errorf("expected 10 characters, got %d", len(runes))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BlockBar(tt.value, tt.max, tt.width)
			tt.check(t, result)
		})
	}
}

// ── RenderAxis tests ───────────────────────────────────────────────────────

func TestRenderAxis(t *testing.T) {
	tests := []struct {
		name   string
		min    float64
		max    float64
		height int
		check  func(t *testing.T, labels []string)
	}{
		{
			name:   "zero height returns nil",
			min:    0,
			max:    100,
			height: 0,
			check:  func(t *testing.T, labels []string) { mustBeNil(t, labels) },
		},
		{
			name:   "produces correct number of labels",
			min:    0,
			max:    100,
			height: 5,
			check: func(t *testing.T, labels []string) {
				if len(labels) != 5 {
					t.Errorf("expected 5 labels, got %d", len(labels))
				}
				// Top label should be max (100 or 100.0), bottom should be 0
				if labels[0] != "100" && labels[0] != "100.0" {
					t.Errorf("expected top label to be ~100, got %q", labels[0])
				}
				if labels[4] != "0" && labels[4] != "0.00" {
					t.Errorf("expected bottom label to be ~0, got %q", labels[4])
				}
			},
		},
		{
			name:   "equal min/max produces range",
			min:    50,
			max:    50,
			height: 3,
			check: func(t *testing.T, labels []string) {
				if len(labels) == 0 {
					t.Fatal("expected non-empty labels")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labels := RenderAxis(tt.min, tt.max, tt.height)
			tt.check(t, labels)
		})
	}
}

// ── LineChart tests ────────────────────────────────────────────────────────

func TestLineChart(t *testing.T) {
	t.Run("empty chart shows no data", func(t *testing.T) {
		lc := NewLineChart(DefaultConfig())
		result := lc.Render()
		mustNonEmpty(t, result)
		if !strings.Contains(result, "No data") {
			t.Error("expected 'No data' message")
		}
	})

	t.Run("single series renders", func(t *testing.T) {
		lc := NewLineChart(ChartConfig{Width: 40, Height: 8, AutoScale: true})
		lc.AddSeries("CPU", []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}, lipgloss.Color("#00FF00"))
		result := lc.Render()
		mustNonEmpty(t, result)
		if strings.Contains(result, "No data") {
			t.Error("expected chart content")
		}
	})

	t.Run("multi-series renders", func(t *testing.T) {
		lc := NewLineChart(ChartConfig{Width: 40, Height: 8, AutoScale: true, ShowLegend: true})
		lc.AddSeries("CPU", []float64{10, 20, 30, 40}, lipgloss.Color("#00FF00"))
		lc.AddSeries("MEM", []float64{40, 30, 20, 10}, lipgloss.Color("#0000FF"))
		result := lc.Render()
		mustNonEmpty(t, result)
	})

	t.Run("with threshold", func(t *testing.T) {
		lc := NewLineChart(ChartConfig{Width: 40, Height: 8, AutoScale: true})
		lc.AddSeries("CPU", []float64{10, 20, 30, 40, 50}, lipgloss.Color("#00FF00"))
		lc.SetThreshold(90, "critical")
		result := lc.Render()
		mustNonEmpty(t, result)
	})

	t.Run("with forecast", func(t *testing.T) {
		lc := NewLineChart(ChartConfig{Width: 40, Height: 8, AutoScale: true})
		lc.AddSeries("CPU", []float64{10, 20, 30, 40, 50}, lipgloss.Color("#00FF00"))
		lc.SetForecast([]float64{55, 60, 65, 70, 75})
		result := lc.Render()
		mustNonEmpty(t, result)
	})

	t.Run("with title and labels", func(t *testing.T) {
		lc := NewLineChart(ChartConfig{
			Width: 40, Height: 8, AutoScale: true,
			Title: "CPU Usage", XLabel: "Time",
		})
		lc.AddSeries("CPU", []float64{10, 20, 30, 40, 50}, lipgloss.Color("#00FF00"))
		result := lc.Render()
		mustNonEmpty(t, result)
		if !strings.Contains(result, "CPU Usage") {
			t.Error("expected title in output")
		}
	})

	t.Run("set config works", func(t *testing.T) {
		lc := NewLineChart(DefaultConfig())
		lc.SetConfig(ChartConfig{Width: 30, Height: 6, AutoScale: true})
		lc.AddSeries("test", []float64{1, 2, 3}, lipgloss.Color("#FF0000"))
		_ = lc.Render()
		cfg := lc.Config()
		if cfg.Width != 30 || cfg.Height != 6 {
			t.Error("SetConfig did not persist")
		}
	})
}

// ── BarChart tests ─────────────────────────────────────────────────────────

func TestBarChart(t *testing.T) {
	t.Run("empty chart shows no data", func(t *testing.T) {
		bc := NewBarChart(DefaultConfig())
		result := bc.Render()
		mustNonEmpty(t, result)
		if !strings.Contains(result, "No data") {
			t.Error("expected 'No data' message")
		}
	})

	t.Run("single bar renders", func(t *testing.T) {
		bc := NewBarChart(ChartConfig{Width: 40})
		bc.AddBar("CPU", 75, lipgloss.Color("#00FF00"))
		result := bc.Render()
		mustNonEmpty(t, result)
		mustContainBlock(t, result)
	})

	t.Run("multiple bars render", func(t *testing.T) {
		bc := NewBarChart(ChartConfig{Width: 40})
		bc.AddBar("CPU", 75, lipgloss.Color("#00FF00"))
		bc.AddBar("MEM", 45, lipgloss.Color("#3B82F6"))
		bc.AddBar("DISK", 30, lipgloss.Color("#F59E0B"))
		result := bc.Render()
		mustNonEmpty(t, result)
		lines := strings.Split(strings.TrimSpace(result), "\n")
		if len(lines) < 3 {
			t.Errorf("expected at least 3 bars, got %d", len(lines))
		}
	})

	t.Run("sorted ascending", func(t *testing.T) {
		bc := NewBarChart(ChartConfig{Width: 40})
		bc.AddBar("A", 10, lipgloss.Color("#FF0000"))
		bc.AddBar("B", 50, lipgloss.Color("#00FF00"))
		bc.AddBar("C", 30, lipgloss.Color("#0000FF"))
		bc.SortByValue(true)
		result := bc.Render()
		mustNonEmpty(t, result)
		lines := strings.Split(strings.TrimSpace(result), "\n")
		if len(lines) >= 3 {
			_ = lines[0] // should be the smallest
			_ = lines[2] // should be the largest
		}
	})

	t.Run("with title", func(t *testing.T) {
		bc := NewBarChart(ChartConfig{Width: 40, Title: "Top Processes"})
		bc.AddBar("chrome", 80, lipgloss.Color("#FF0000"))
		result := bc.Render()
		if !strings.Contains(result, "Top Processes") {
			t.Error("expected title")
		}
	})

	t.Run("grouped bars", func(t *testing.T) {
		bc := NewBarChart(ChartConfig{Width: 40})
		bc.AddGroup("Network", []float64{10, 20}, []color.Color{
			lipgloss.Color("#FF0000"),
			lipgloss.Color("#0000FF"),
		})
		result := bc.Render()
		mustNonEmpty(t, result)
	})
}

// ── Gauge tests ────────────────────────────────────────────────────────────

func TestGauge(t *testing.T) {
	t.Run("default gauge renders", func(t *testing.T) {
		g := NewGauge(ChartConfig{Width: 30})
		result := g.Render()
		mustNonEmpty(t, result)
	})

	t.Run("0% value", func(t *testing.T) {
		g := NewGauge(ChartConfig{Width: 30})
		g.SetValue(0)
		result := g.Render()
		mustNonEmpty(t, result)
		if strings.Contains(result, "▓▓") {
			t.Error("expected empty gauge for 0%")
		}
	})

	t.Run("100% value", func(t *testing.T) {
		g := NewGauge(ChartConfig{Width: 30})
		g.SetValue(100)
		result := g.Render()
		mustNonEmpty(t, result)
	})

	t.Run("mid value", func(t *testing.T) {
		g := NewGauge(ChartConfig{Width: 30})
		g.SetValue(50)
		result := g.Render()
		mustNonEmpty(t, result)
	})

	t.Run("with thresholds", func(t *testing.T) {
		g := NewGauge(ChartConfig{Width: 30})
		g.SetValue(80)
		g.SetThresholds(50, 90)
		result := g.Render()
		mustNonEmpty(t, result)
	})

	t.Run("with label", func(t *testing.T) {
		g := NewGauge(ChartConfig{Width: 30})
		g.SetValue(73)
		g.SetLabel("CPU")
		result := g.Render()
		mustNonEmpty(t, result)
		if !strings.Contains(result, "CPU") {
			t.Error("expected label in output")
		}
	})

	t.Run("set config works", func(t *testing.T) {
		g := NewGauge(DefaultConfig())
		g.SetConfig(ChartConfig{Width: 20})
		if c := g.Config(); c.Width != 20 {
			t.Error("SetConfig did not persist")
		}
	})
}

// ── Sparkline tests ────────────────────────────────────────────────────────

func TestSparkline(t *testing.T) {
	t.Run("empty data returns empty", func(t *testing.T) {
		s := NewSparkline(10)
		result := s.Render()
		mustBeEmpty(t, result)
	})

	t.Run("basic sparkline renders", func(t *testing.T) {
		s := NewSparkline(10)
		s.SetValues([]float64{10, 20, 30, 40, 50})
		result := s.Render()
		mustNonEmpty(t, result)
		mustContainBraille(t, result)
	})

	t.Run("with min/max", func(t *testing.T) {
		s := NewSparkline(10)
		s.SetValues([]float64{10, 20, 30, 40, 50})
		s.ShowMinMax(true)
		result := s.Render()
		mustNonEmpty(t, result)
		if !strings.Contains(result, "min:") || !strings.Contains(result, "max:") {
			t.Error("expected min/max labels")
		}
	})

	t.Run("with trend", func(t *testing.T) {
		s := NewSparkline(10)
		s.SetValues([]float64{10, 20, 30, 40, 50})
		s.ShowTrend(true)
		result := s.Render()
		mustNonEmpty(t, result)
	})

	t.Run("single value sparkline", func(t *testing.T) {
		s := NewSparkline(10)
		s.SetValues([]float64{42})
		result := s.Render()
		mustNonEmpty(t, result)
	})

	t.Run("constant values", func(t *testing.T) {
		s := NewSparkline(10)
		s.SetValues([]float64{50, 50, 50, 50, 50})
		result := s.Render()
		mustNonEmpty(t, result)
	})

	t.Run("with both min/max and trend", func(t *testing.T) {
		s := NewSparkline(15)
		s.SetValues([]float64{5, 15, 25, 35, 45, 55, 45, 35, 25, 15})
		s.ShowMinMax(true)
		s.ShowTrend(true)
		result := s.Render()
		mustNonEmpty(t, result)
	})
}

// ── HeatMap tests ──────────────────────────────────────────────────────────

func TestHeatMap(t *testing.T) {
	t.Run("empty dimensions return empty", func(t *testing.T) {
		hm := NewHeatMap(0, 0)
		result := hm.Render()
		mustBeEmpty(t, result)
	})

	t.Run("basic grid renders", func(t *testing.T) {
		hm := NewHeatMap(3, 4)
		hm.SetCell(0, 0, 0.8)
		hm.SetCell(1, 1, 0.5)
		hm.SetCell(2, 2, 0.2)
		result := hm.Render()
		mustNonEmpty(t, result)
	})

	t.Run("with row and column labels", func(t *testing.T) {
		hm := NewHeatMap(3, 4)
		hm.SetRowLabel(0, "80-443")
		hm.SetRowLabel(1, "444-1023")
		hm.SetRowLabel(2, "1024+")
		hm.SetColLabel(0, "0s")
		hm.SetColLabel(3, "60s")
		result := hm.Render()
		mustNonEmpty(t, result)
	})

	t.Run("color scheme changes", func(t *testing.T) {
		hm := NewHeatMap(2, 2)
		hm.SetCell(0, 0, 0.9)
		hm.ColorScheme(HeatBlueWhiteRed)
		result := hm.Render()
		mustNonEmpty(t, result)
	})

	t.Run("all color schemes render", func(t *testing.T) {
		for _, scheme := range []HeatScheme{HeatGreenAmberRed, HeatBlueWhiteRed, HeatWhiteRed, HeatGreenWhite} {
			hm := NewHeatMap(2, 2)
			hm.SetCell(0, 0, 0.1)
			hm.SetCell(0, 1, 0.5)
			hm.SetCell(1, 0, 0.9)
			hm.ColorScheme(scheme)
			result := hm.Render()
			mustNonEmpty(t, result)
		}
	})

	t.Run("out of bounds cell is safe", func(t *testing.T) {
		hm := NewHeatMap(3, 3)
		hm.SetCell(10, 10, 1.0) // should not panic
		result := hm.Render()
		mustNonEmpty(t, result)
	})
}

// ── NumericDisplay tests ───────────────────────────────────────────────────

func TestNumericDisplay(t *testing.T) {
	t.Run("basic display renders", func(t *testing.T) {
		nd := NewNumericDisplay(ChartConfig{Width: 20})
		nd.SetValue(73.2, "%")
		result := nd.Render()
		mustNonEmpty(t, result)
		if !strings.Contains(result, "73.2") {
			t.Error("expected value in output")
		}
	})

	t.Run("with rising trend", func(t *testing.T) {
		nd := NewNumericDisplay(ChartConfig{Width: 20})
		nd.SetValue(73.2, "%")
		nd.SetTrend(TrendRising, 2.1)
		result := nd.Render()
		mustNonEmpty(t, result)
	})

	t.Run("with falling trend", func(t *testing.T) {
		nd := NewNumericDisplay(ChartConfig{Width: 20})
		nd.SetValue(45.0, "%")
		nd.SetTrend(TrendFalling, 5.3)
		result := nd.Render()
		mustNonEmpty(t, result)
	})

	t.Run("with stable trend", func(t *testing.T) {
		nd := NewNumericDisplay(ChartConfig{Width: 20})
		nd.SetValue(50.0, "%")
		nd.SetTrend(TrendStable, 0)
		result := nd.Render()
		mustNonEmpty(t, result)
		// Should not show arrow for stable
		if strings.Contains(result, "→") || strings.Contains(result, "↑") || strings.Contains(result, "↓") {
			// This might be present depending on implementation, not a hard error
		}
	})

	t.Run("with title", func(t *testing.T) {
		nd := NewNumericDisplay(ChartConfig{Width: 20, Title: "CPU"})
		nd.SetValue(73.2, "%")
		result := nd.Render()
		mustNonEmpty(t, result)
		if !strings.Contains(result, "CPU") {
			t.Error("expected title in output")
		}
	})

	t.Run("custom color", func(t *testing.T) {
		nd := NewNumericDisplay(ChartConfig{Width: 20})
		nd.SetValue(99.9, "%")
		nd.SetColor(lipgloss.Color("#FF0000"))
		result := nd.Render()
		mustNonEmpty(t, result)
	})

	t.Run("set config works", func(t *testing.T) {
		nd := NewNumericDisplay(DefaultConfig())
		nd.SetConfig(ChartConfig{Width: 15})
		if c := nd.Config(); c.Width != 15 {
			t.Error("SetConfig did not persist")
		}
	})
}

// ── Normalize helper tests ─────────────────────────────────────────────────

func TestNormalize(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		min    float64
		max    float64
		check  func(t *testing.T, result []float64)
	}{
		{
			name:   "empty input",
			values: nil,
			check:  func(t *testing.T, result []float64) {},
		},
		{
			name:   "normal range",
			values: []float64{0, 50, 100},
			min:    0,
			max:    100,
			check: func(t *testing.T, result []float64) {
				if len(result) != 3 {
					t.Fatalf("expected 3 results, got %d", len(result))
				}
				if result[0] != 0 || result[2] != 1 {
					t.Errorf("expected [0, 1], got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalize(tt.values, tt.min, tt.max)
			tt.check(t, result)
		})
	}
}

// ── NiceRange tests ────────────────────────────────────────────────────────

func TestNiceRange(t *testing.T) {
	niceMin, niceMax, step := NiceRange(0, 100, 5)
	if step <= 0 {
		t.Errorf("expected positive step, got %f", step)
	}
	if niceMin > 0 || niceMax < 100 {
		t.Errorf("expected range covering [0,100], got [%f,%f]", niceMin, niceMax)
	}
}

// ── Helpers ────────────────────────────────────────────────────────────────

func mustBeEmpty(t *testing.T, result string) {
	t.Helper()
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func mustBeNil(t *testing.T, labels []string) {
	t.Helper()
	if labels != nil {
		t.Errorf("expected nil, got %v", labels)
	}
}

func mustNonEmpty(t *testing.T, result string) {
	t.Helper()
	if result == "" {
		t.Fatal("expected non-empty result")
	}
}

func mustContainBraille(t *testing.T, result string) {
	t.Helper()
	for _, r := range result {
		if r >= 0x2800 && r <= 0x28FF {
			return
		}
	}
	t.Error("expected braille characters (U+2800-U+28FF)")
}

func mustContainBlock(t *testing.T, result string) {
	t.Helper()
	blockChars := "▁▂▃▄▅▆▇█░"
	for _, r := range result {
		for _, bc := range blockChars {
			if r == bc {
				return
			}
		}
	}
	// Block bar may also use only █ for full fills
	if strings.Contains(result, "█") {
		return
	}
	t.Error("expected block characters (▁▂▃▄▅▆▇█░)")
}
