package charts

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
)

// ── HeatMap ────────────────────────────────────────────────────────────────

// HeatScheme defines a color interpolation scheme for the heatmap.
type HeatScheme int

const (
	HeatGreenAmberRed HeatScheme = iota
	HeatBlueWhiteRed
	HeatWhiteRed
	HeatGreenWhite
)

// HeatMap renders a grid of colored cells representing data intensity.
type HeatMap struct {
	rows      int
	cols      int
	cells     [][]float64 // [row][col] value 0–1
	rowLabels map[int]string
	colLabels map[int]string
	scheme    HeatScheme
}

// NewHeatMap creates a heatmap with the given grid dimensions.
func NewHeatMap(rows, cols int) *HeatMap {
	cells := make([][]float64, rows)
	for i := range cells {
		cells[i] = make([]float64, cols)
	}
	return &HeatMap{
		rows:      rows,
		cols:      cols,
		cells:     cells,
		rowLabels: make(map[int]string),
		colLabels: make(map[int]string),
		scheme:    HeatGreenAmberRed,
	}
}

// SetCell sets the value (0–1) for a cell at the given position.
func (hm *HeatMap) SetCell(row, col int, value float64) {
	if row < 0 || row >= hm.rows || col < 0 || col >= hm.cols {
		return
	}
	hm.cells[row][col] = clamp(value, 0, 1)
}

// SetRowLabel sets the label for a row.
func (hm *HeatMap) SetRowLabel(row int, label string) {
	hm.rowLabels[row] = label
}

// SetColLabel sets the label for a column.
func (hm *HeatMap) SetColLabel(col int, label string) {
	hm.colLabels[col] = label
}

// ColorScheme sets the color interpolation scheme.
func (hm *HeatMap) ColorScheme(scheme HeatScheme) {
	hm.scheme = scheme
}

// SetConfig is a no-op (config is via NewHeatMap).
func (hm *HeatMap) SetConfig(cfg ChartConfig) {}

// Config returns a basic configuration.
func (hm *HeatMap) Config() ChartConfig {
	return ChartConfig{
		Width:  hm.cols * 4,
		Height: hm.rows*2 + 1,
	}
}

// Render returns the heatmap as a styled string.
func (hm *HeatMap) Render() string {
	if hm.rows == 0 || hm.cols == 0 {
		return ""
	}

	var b strings.Builder

	// ── Column labels ──
	if len(hm.colLabels) > 0 {
		// Leave space for row labels
		maxRowLabelLen := 0
		for _, l := range hm.rowLabels {
			if len(l) > maxRowLabelLen {
				maxRowLabelLen = len(l)
			}
		}
		colLabelPadding := maxRowLabelLen + 1
		b.WriteString(strings.Repeat(" ", colLabelPadding))

		for c := 0; c < hm.cols; c++ {
			label, ok := hm.colLabels[c]
			if !ok {
				label = ""
			}
			// Center label over cell (4 chars wide)
			centered := centerText(label, 4)
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(centered))
		}
		b.WriteString("\n")
	}

	// ── Density characters ──
	densities := []string{" ", "░", "▒", "▓", "█"}

	// ── Grid ──
	for r := 0; r < hm.rows; r++ {
		// Row label
		if label, ok := hm.rowLabels[r]; ok {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(fmt.Sprintf("%-*s", maxRowLabelLen(hm.rowLabels), label)))
			b.WriteString(" ")
		}

		// Cells
		b.WriteString("│")
		for c := 0; c < hm.cols; c++ {
			val := hm.cells[r][c]
			densityIdx := int(math.Round(val * 4))
			if densityIdx < 0 {
				densityIdx = 0
			}
			if densityIdx > 4 {
				densityIdx = 4
			}

			cellColor := hm.schemeColor(val)
			cellChar := densities[densityIdx]
			// Render cell twice for 4-char wide visual (2 spaces + char + space)
			b.WriteString(lipgloss.NewStyle().Foreground(cellColor).Background(cellColor).Render(" "))
			b.WriteString(lipgloss.NewStyle().Foreground(cellColor).Background(cellColor).Render(cellChar))
			b.WriteString(lipgloss.NewStyle().Foreground(cellColor).Background(cellColor).Render(" "))
		}
		b.WriteString("│")
		b.WriteByte('\n')
	}

	return b.String()
}

func maxRowLabelLen(labels map[int]string) int {
	maxLen := 0
	for _, l := range labels {
		if len(l) > maxLen {
			maxLen = len(l)
		}
	}
	return maxLen
}

func centerText(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	padding := width - len(s)
	left := padding / 2
	right := padding - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

// schemeColor maps a value 0–1 to a color based on the active scheme.
func (hm *HeatMap) schemeColor(value float64) color.Color {
	switch hm.scheme {
	case HeatGreenAmberRed:
		switch {
		case value < 0.33:
			return lipgloss.Color("#10B981") // green
		case value < 0.66:
			return lipgloss.Color("#F59E0B") // amber
		default:
			return lipgloss.Color("#EF4444") // red
		}
	case HeatBlueWhiteRed:
		switch {
		case value < 0.33:
			return lipgloss.Color("#3B82F6") // blue
		case value < 0.66:
			return lipgloss.Color("#F3F4F6") // white
		default:
			return lipgloss.Color("#EF4444") // red
		}
	case HeatWhiteRed:
		switch {
		case value < 0.25:
			return lipgloss.Color("#F3F4F6") // white
		case value < 0.5:
			return lipgloss.Color("#FCA5A5") // light red
		case value < 0.75:
			return lipgloss.Color("#F87171") // medium red
		default:
			return lipgloss.Color("#EF4444") // full red
		}
	case HeatGreenWhite:
		switch {
		case value < 0.25:
			return lipgloss.Color("#F3F4F6") // white
		case value < 0.5:
			return lipgloss.Color("#A7F3D0") // light green
		case value < 0.75:
			return lipgloss.Color("#34D399") // medium green
		default:
			return lipgloss.Color("#10B981") // full green
		}
	default:
		return lipgloss.Color("#888888")
	}
}
