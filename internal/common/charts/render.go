package charts

import (
	"math"
	"strings"

	"charm.land/lipgloss/v2"
)

// ── Public rendering primitives ────────────────────────────────────────────

// BrailleLine renders a single data series as a braille line chart.
// width  = number of braille characters horizontally
// height = number of terminal rows vertically
// Returns a height×width grid of braille characters separated by newlines.
func BrailleLine(points []float64, width, height int) string {
	if len(points) == 0 || width <= 0 || height <= 0 {
		return ""
	}
	return BrailleLineMulti([]Series{{Data: points}}, width, height)
}

// BrailleLineMulti renders multiple series into a combined braille grid.
// All series are normalized to the same global data range.
func BrailleLineMulti(series []Series, width, height int) string {
	if len(series) == 0 || width <= 0 || height <= 0 {
		return ""
	}

	// Find global min/max across all non-empty series
	var allMin, allMax float64
	first := true
	for _, s := range series {
		if len(s.Data) == 0 {
			continue
		}
		sMin, sMax := minMaxFloat(s.Data)
		if first || sMin < allMin {
			allMin = sMin
		}
		if first || sMax > allMax {
			allMax = sMax
		}
		first = false
	}
	if first {
		return "" // no data in any series
	}
	if allMax == allMin {
		allMax = allMin + 1
	}

	horizRes := width * 2 // effective horizontal subdivisions
	vertRes := height * 4 // effective vertical dot rows

	// brailleGrid[row][col] = accumulated braille dot bitmask
	brailleGrid := make([][]int, height)
	for i := range brailleGrid {
		brailleGrid[i] = make([]int, width)
	}

	for _, s := range series {
		if len(s.Data) == 0 {
			continue
		}

		// Compute Y position (0 = bottom, vertRes-1 = top) for each subdivision
		yPos := make([]int, horizRes)
		for i := 0; i < horizRes; i++ {
			idx := float64(i) / float64(horizRes-1) * float64(len(s.Data)-1)
			low := int(idx)
			high := low + 1
			if high >= len(s.Data) {
				high = len(s.Data) - 1
			}
			frac := idx - float64(low)
			val := s.Data[low]*(1-frac) + s.Data[high]*frac

			normV := clamp((val-allMin)/(allMax-allMin), 0, 1)
			yPos[i] = int(math.Round(normV * float64(vertRes-1)))
		}

		// Light dots in the braille grid
		for col := 0; col < width; col++ {
			for sub := 0; sub < 2; sub++ {
				hPos := col*2 + sub
				if hPos >= horizRes {
					continue
				}
				y := yPos[hPos]
				row := y / 4
				if row < 0 {
					row = 0
				}
				if row >= height {
					row = height - 1
				}
				dot := y - row*4 // 0..3 within the braille cell

				if sub == 0 { // left column
					switch dot {
					case 0:
						brailleGrid[row][col] |= 0x40 // dot7 (bottom)
					case 1:
						brailleGrid[row][col] |= 0x04 // dot3
					case 2:
						brailleGrid[row][col] |= 0x02 // dot2
					case 3:
						brailleGrid[row][col] |= 0x01 // dot1 (top)
					}
				} else { // right column
					switch dot {
					case 0:
						brailleGrid[row][col] |= 0x80 // dot8 (bottom)
					case 1:
						brailleGrid[row][col] |= 0x20 // dot6
					case 2:
						brailleGrid[row][col] |= 0x10 // dot5
					case 3:
						brailleGrid[row][col] |= 0x08 // dot4 (top)
					}
				}
			}
		}
	}

	// Render the grid from top row to bottom row
	var b strings.Builder
	for row := height - 1; row >= 0; row-- {
		for col := 0; col < width; col++ {
			b.WriteRune(rune(0x2800 + brailleGrid[row][col]))
		}
		if row > 0 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// BlockBar renders a horizontal bar using Unicode eighth-block characters.
// Characters: ▁▂▃▄▅▆▇█ (1/8 through 8/8 fill)
func BlockBar(value, max, width float64) string {
	if width <= 0 {
		return ""
	}
	blocks := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	empty := "░"

	if max <= 0 || value <= 0 {
		return strings.Repeat(empty, int(width))
	}

	pct := math.Min(value/max, 1.0)
	fillWidth := pct * width
	full := int(math.Floor(fillWidth))
	frac := fillWidth - float64(full)

	var b strings.Builder
	for i := 0; i < full; i++ {
		b.WriteString("█")
	}
	if full < int(width) {
		if frac > 0 {
			idx := int(math.Floor(frac * 8))
			if idx >= 8 {
				idx = 7
			}
			b.WriteString(blocks[idx])
			full++
		}
		for i := full; i < int(width); i++ {
			b.WriteString(empty)
		}
	}
	return b.String()
}

// RenderAxis generates Y-axis labels for a given range and height.
// Returns labels from top (index 0) to bottom (index height-1).
func RenderAxis(min, max float64, height int) []string {
	if height <= 0 {
		return nil
	}
	if max == min {
		max = min + 1
	}

	labels := make([]string, height)
	for i := 0; i < height; i++ {
		pct := float64(i) / float64(height-1)
		val := min + pct*(max-min)
		labels[height-1-i] = formatFloat(val) // top to bottom
	}
	return labels
}

// RenderGrid generates a styled grid background string.
func RenderGrid(width, height int, style lipgloss.Style) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	dot := style.Render("·")
	lines := make([]string, height)
	for i := range lines {
		var b strings.Builder
		for j := 0; j < width; j++ {
			b.WriteString(dot)
		}
		lines[i] = b.String()
	}
	return strings.Join(lines, "\n")
}

// NiceRange computes a "nice" min, max, and step count for axis labels.
func NiceRange(min, max float64, steps int) (niceMin, niceMax, step float64) {
	if max == min {
		max = min + 1
	}
	rawRange := max - min
	roughStep := rawRange / float64(steps)

	// Normalize to 1, 2, 5, 10, 20, 50, …
	scale := math.Pow(10, math.Floor(math.Log10(roughStep)))
	norm := roughStep / scale

	var niceStep float64
	switch {
	case norm < 1.5:
		niceStep = 1
	case norm < 3.5:
		niceStep = 2
	case norm < 7.5:
		niceStep = 5
	default:
		niceStep = 10
	}
	niceStep *= scale

	niceMin = math.Floor(min/niceStep) * niceStep
	niceMax = math.Ceil(max/niceStep) * niceStep
	step = niceStep
	return
}

// ── Internal helpers ───────────────────────────────────────────────────────

// normalize scales values to the [0, 1] range.
func normalize(values []float64, min, max float64) []float64 {
	if len(values) == 0 {
		return nil
	}
	out := make([]float64, len(values))
	diff := max - min
	if diff == 0 {
		return out // all zeros
	}
	for i, v := range values {
		out[i] = clamp((v-min)/diff, 0, 1)
	}
	return out
}

func minMaxFloat(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 1
	}
	minV := values[0]
	maxV := values[0]
	for _, v := range values {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	return minV, maxV
}
