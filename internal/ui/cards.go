package ui

import (
	"fmt"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common/charts"
)

// ── Card Types ──────────────────────────────────────────────────────────────

// CardType identifies the kind of card for type-switching.
type CardType int

const (
	CardGauge CardType = iota
	CardChart
	CardMetric
	CardOperation
	CardStatusCard
	CardDetail
)

// CardStatus represents the live state of a card.
type CardStatus int

const (
	CardLoading   CardStatus = iota // Spinner / "Loading..."
	CardReady                       // Data loaded, rendering
	CardError                       // Error state with message
	CardStreaming                   // Live-updating (animation)
)

// Card is the interface all card types implement.
type Card interface {
	Type() CardType
	Render(width int) string
	Focus() bool
	SetFocused(bool)
	Title() string
	Status() CardStatus
	SetStatus(CardStatus)
}

// StatusIndicator returns the colored status dot for a card.
func StatusIndicator(status CardStatus) string {
	p := common.CurrentPalette()
	switch status {
	case CardLoading:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(p.Warning)).Render("⏳")
	case CardReady:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(p.Secondary)).Render("●")
	case CardError:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(p.Danger)).Render("✕")
	case CardStreaming:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(p.Info)).Render("⟳")
	}
	return ""
}

// ── GaugeCard ───────────────────────────────────────────────────────────────

// GaugeCard shows a health metric with a horizontal gauge bar, percentage, and
// an optional trend arrow.
type GaugeCard struct {
	title    string
	label    string
	value    float64 // 0-100
	warnAt   float64
	critAt   float64
	trendDir charts.TrendDirection
	trendPct float64
	status   CardStatus
	focused  bool
}

// NewGaugeCard creates a ready-to-use health gauge card.
func NewGaugeCard(title string) *GaugeCard {
	return &GaugeCard{
		title:  title,
		warnAt: 70,
		critAt: 90,
		status: CardReady,
	}
}

func (c *GaugeCard) Type() CardType         { return CardGauge }
func (c *GaugeCard) Focus() bool            { return c.focused }
func (c *GaugeCard) SetFocused(v bool)      { c.focused = v }
func (c *GaugeCard) Title() string          { return c.title }
func (c *GaugeCard) Status() CardStatus     { return c.status }
func (c *GaugeCard) SetStatus(s CardStatus) { c.status = s }

// SetValue sets the gauge value (clamped 0-100).
func (c *GaugeCard) SetValue(v float64) {
	c.value = math.Max(0, math.Min(100, v))
}

// Value returns the current gauge value.
func (c *GaugeCard) Value() float64 { return c.value }

// SetLabel sets an optional text label below the gauge bar.
func (c *GaugeCard) SetLabel(l string) { c.label = l }

// SetTrend sets the trend direction and percentage change.
func (c *GaugeCard) SetTrend(dir charts.TrendDirection, pct float64) {
	c.trendDir = dir
	c.trendPct = pct
}

// Render returns the card as a bordered string.
func (c *GaugeCard) Render(width int) string {
	if width < 12 {
		width = 12
	}
	p := common.CurrentPalette()

	// ── Bar dimensions ──
	indicatorW := lipgloss.Width(StatusIndicator(c.status))
	trendStr := c.trendString()
	trendW := lipgloss.Width(trendStr)
	// reserve: indicator + " " + percentage (e.g. " 100%") + trend space
	pctStr := fmt.Sprintf(" %3.0f%%", c.value)
	reserved := indicatorW + 2 + len(pctStr) + trendW
	if reserved > width-4 {
		reserved = width - 4
	}
	barW := width - 4 - reserved
	if barW < 2 {
		barW = 2
	}

	gaugeColor := common.GetHealthColor(c.value)

	// ── Build inner content ──
	var inner strings.Builder

	// Title line
	titleStr := c.title
	if c.focused {
		titleStr = "▸ " + c.title
	}
	titleLine := fmt.Sprintf("%-*s%s", width-4-indicatorW, titleStr, StatusIndicator(c.status))
	inner.WriteString(CardTitleStyle.Render(titleLine))
	inner.WriteByte('\n')

	// Gauge bar line
	bar := strings.Builder{}
	bar.WriteString(strings.Repeat("▓", int(math.Round(c.value/100.0*float64(barW)))))
	bar.WriteString(strings.Repeat("░", barW-int(math.Round(c.value/100.0*float64(barW)))))
	gaugeBar := lipgloss.NewStyle().Foreground(gaugeColor).Render(bar.String())
	gaugeLine := gaugeBar + pctStr + trendStr
	inner.WriteString(CardBodyStyle.Render(gaugeLine))
	inner.WriteByte('\n')

	// Optional label line
	if c.label != "" {
		inner.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)).Render(c.label))
		inner.WriteByte('\n')
	}

	style := CardStyle
	if c.focused {
		style = CardFocusedStyle
	}
	return style.Width(width).Render(inner.String())
}

func (c *GaugeCard) trendString() string {
	if c.trendDir == charts.TrendStable {
		return ""
	}
	p := common.CurrentPalette()
	var arrow string
	var clr string
	switch c.trendDir {
	case charts.TrendRising:
		arrow = "↑"
		clr = p.Danger
	case charts.TrendFalling:
		arrow = "↓"
		clr = p.Secondary
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(clr)).Render(
		fmt.Sprintf(" %s%.1f%%", arrow, c.trendPct))
}

// ── OperationCard ───────────────────────────────────────────────────────────

// OperationCard is a navigable card that links to an ops layer. It shows an
// optional icon, the layer name, a short description, and a status indicator.
type OperationCard struct {
	title       string
	description string
	icon        string
	screen      common.Screen
	status      CardStatus
	focused     bool
}

// NewOperationCard creates a card representing an ops layer.
func NewOperationCard(icon, title, description string, screen common.Screen) *OperationCard {
	return &OperationCard{
		icon:        icon,
		title:       title,
		description: description,
		screen:      screen,
		status:      CardReady,
	}
}

func (c *OperationCard) Type() CardType         { return CardOperation }
func (c *OperationCard) Focus() bool            { return c.focused }
func (c *OperationCard) SetFocused(v bool)      { c.focused = v }
func (c *OperationCard) Title() string          { return c.title }
func (c *OperationCard) Status() CardStatus     { return c.status }
func (c *OperationCard) SetStatus(s CardStatus) { c.status = s }

// Screen returns the target ops layer screen.
func (c *OperationCard) Screen() common.Screen { return c.screen }

// Render returns the operation card as a bordered string.
func (c *OperationCard) Render(width int) string {
	if width < 16 {
		width = 16
	}
	p := common.CurrentPalette()

	var inner strings.Builder

	// Icon + title + status
	indicator := StatusIndicator(c.status)
	indicatorW := lipgloss.Width(indicator)
	titlePrefix := ""
	if c.focused {
		titlePrefix = "▸ "
	}
	titleStr := fmt.Sprintf("%s%s%s %s", titlePrefix, c.icon, c.title, indicator)
	// Pad to full width so focus border renders correctly
	titleLine := fmt.Sprintf("%-*s", width-4-indicatorW, titleStr)
	inner.WriteString(CardTitleStyle.Render(titleLine))
	inner.WriteByte('\n')
	inner.WriteByte('\n')

	// Description
	desc := common.TruncateString(c.description, width-6)
	inner.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)).Render(desc))
	inner.WriteByte('\n')
	inner.WriteByte('\n')

	// Footer hint
	hint := fmt.Sprintf("[Enter] open %s", c.title)
	inner.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(p.Border)).Render(hint))
	inner.WriteByte('\n')

	style := CardStyle
	if c.focused {
		style = CardFocusedStyle
	}
	return style.Width(width).Render(inner.String())
}

// ── ChartCard ───────────────────────────────────────────────────────────────

// ChartCard embeds a LineChart with a title and summary labels.
type ChartCard struct {
	title    string
	chart    *charts.LineChart
	leftLbl  string
	rightLbl string
	status   CardStatus
	focused  bool
}

// NewChartCard creates a card wrapping a line chart.
func NewChartCard(title string, cfg charts.ChartConfig) *ChartCard {
	return &ChartCard{
		title:  title,
		chart:  charts.NewLineChart(cfg),
		status: CardReady,
	}
}

func (c *ChartCard) Type() CardType         { return CardChart }
func (c *ChartCard) Focus() bool            { return c.focused }
func (c *ChartCard) SetFocused(v bool)      { c.focused = v }
func (c *ChartCard) Title() string          { return c.title }
func (c *ChartCard) Status() CardStatus     { return c.status }
func (c *ChartCard) SetStatus(s CardStatus) { c.status = s }

// SetLabels sets left and right summary labels below the chart.
func (c *ChartCard) SetLabels(left, right string) {
	c.leftLbl = left
	c.rightLbl = right
}

// Chart returns the underlying LineChart for adding series.
func (c *ChartCard) Chart() *charts.LineChart { return c.chart }

// Render returns the chart card as a bordered string.
func (c *ChartCard) Render(width int) string {
	if width < 20 {
		width = 20
	}

	// Configure chart for available space
	cfg := c.chart.Config()
	cfg.Width = width - 4
	cfg.Height = 3
	c.chart.SetConfig(cfg)

	var inner strings.Builder

	// Title + status
	indicator := StatusIndicator(c.status)
	titleStr := c.title
	if c.focused {
		titleStr = "▸ " + c.title
	}
	indicatorW := lipgloss.Width(indicator)
	titleLine := fmt.Sprintf("%-*s%s", width-4-indicatorW, titleStr, indicator)
	inner.WriteString(CardTitleStyle.Render(titleLine))
	inner.WriteByte('\n')

	// Chart
	chartStr := c.chart.Render()
	inner.WriteString(chartStr)
	if !strings.HasSuffix(chartStr, "\n") {
		inner.WriteByte('\n')
	}

	// Labels
	if c.leftLbl != "" || c.rightLbl != "" {
		p := common.CurrentPalette()
		lblStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted))
		avail := width - 4
		left := lblStyle.Render(c.leftLbl)
		right := lblStyle.Render(c.rightLbl)
		pad := avail - lipgloss.Width(left) - lipgloss.Width(right)
		if pad < 1 {
			pad = 1
		}
		inner.WriteString(left + strings.Repeat(" ", pad) + right)
		inner.WriteByte('\n')
	}

	style := CardStyle
	if c.focused {
		style = CardFocusedStyle
	}
	return style.Width(width).Render(inner.String())
}

// ── StatusCard ──────────────────────────────────────────────────────────────

// StatusCard is a compact card with an icon, label, value and optional sparkline.
type StatusCard struct {
	title     string
	label     string
	value     string
	sparkData []float64
	status    CardStatus
	focused   bool
}

// NewStatusCard creates a compact status card.
func NewStatusCard(title string) *StatusCard {
	return &StatusCard{
		title:  title,
		status: CardReady,
	}
}

func (c *StatusCard) Type() CardType         { return CardStatusCard }
func (c *StatusCard) Focus() bool            { return c.focused }
func (c *StatusCard) SetFocused(v bool)      { c.focused = v }
func (c *StatusCard) Title() string          { return c.title }
func (c *StatusCard) Status() CardStatus     { return c.status }
func (c *StatusCard) SetStatus(s CardStatus) { c.status = s }

// SetValue sets the displayed value text.
func (c *StatusCard) SetValue(v string) { c.value = v }

// SetLabel sets the label text.
func (c *StatusCard) SetLabel(l string) { c.label = l }

// SetSparklineData sets the data for the sparkline.
func (c *StatusCard) SetSparklineData(data []float64) { c.sparkData = data }

// Render returns the status card as a bordered string.
func (c *StatusCard) Render(width int) string {
	if width < 12 {
		width = 12
	}
	p := common.CurrentPalette()

	var inner strings.Builder

	indicator := StatusIndicator(c.status)
	indicatorW := lipgloss.Width(indicator)
	titleStr := c.title
	if c.focused {
		titleStr = "▸ " + c.title
	}
	titleLine := fmt.Sprintf("%-*s%s", width-4-indicatorW, titleStr, indicator)
	inner.WriteString(CardTitleStyle.Render(titleLine))
	inner.WriteByte('\n')

	// Value + label line
	if c.value != "" || c.label != "" {
		valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Secondary)).Bold(true)
		inner.WriteString(valStyle.Render(c.value))
		if c.label != "" {
			inner.WriteString(" ")
			inner.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)).Render(c.label))
		}
		inner.WriteByte('\n')
	}

	// Sparkline
	if len(c.sparkData) > 0 {
		sparkW := width - 6
		if sparkW < 4 {
			sparkW = 4
		}
		sp := charts.NewSparkline(sparkW)
		sp.SetValues(c.sparkData)
		inner.WriteString(sp.Render())
	}

	style := CardStyle
	if c.focused {
		style = CardFocusedStyle
	}
	return style.Width(width).Render(inner.String())
}

// ── CardGrid ────────────────────────────────────────────────────────────────

// CardGrid arranges cards in a responsive grid with keyboard focus navigation.
type CardGrid struct {
	cards   []Card
	gap     int
	focused int
}

// NewCardGrid creates an empty grid with the given inter-card gap.
func NewCardGrid(gap int) *CardGrid {
	return &CardGrid{
		gap:     gap,
		focused: 0,
	}
}

// AddCard appends a card to the grid.
func (g *CardGrid) AddCard(card Card) {
	g.cards = append(g.cards, card)
}

// Cards returns the underlying slice for direct access.
func (g *CardGrid) Cards() []Card { return g.cards }

// Len returns the number of cards.
func (g *CardGrid) Len() int { return len(g.cards) }

// Focused returns the index of the focused card, or -1 if empty.
func (g *CardGrid) Focused() int {
	if len(g.cards) == 0 {
		return -1
	}
	return g.focused
}

// FocusedCard returns the currently focused card, or nil.
func (g *CardGrid) FocusedCard() Card {
	if g.focused < 0 || g.focused >= len(g.cards) {
		return nil
	}
	return g.cards[g.focused]
}

// NextCard moves focus to the next card in reading order.
func (g *CardGrid) NextCard() {
	if len(g.cards) == 0 {
		return
	}
	g.focused = (g.focused + 1) % len(g.cards)
}

// PrevCard moves focus to the previous card.
func (g *CardGrid) PrevCard() {
	if len(g.cards) == 0 {
		return
	}
	g.focused = (g.focused - 1 + len(g.cards)) % len(g.cards)
}

// SetFocused sets focus to a specific card index.
func (g *CardGrid) SetFocused(idx int) {
	if idx >= 0 && idx < len(g.cards) {
		g.focused = idx
	}
}

// AutoColumns computes the optimal column count based on terminal width.
func (g *CardGrid) AutoColumns(width int) int {
	switch {
	case width < 80:
		return 1
	case width < 120:
		return 2
	default:
		return 3
	}
}

// Render renders all cards in the grid layout.
func (g *CardGrid) Render(width int) string {
	if len(g.cards) == 0 {
		return ""
	}

	cols := g.AutoColumns(width)
	if cols < 1 {
		cols = 1
	}

	// Total gap consumed between columns
	totalGap := (cols - 1) * g.gap
	cardW := (width - totalGap) / cols
	if cardW < 10 {
		cardW = 10
	}

	rows := (len(g.cards) + cols - 1) / cols
	var rowStrs []string

	for r := 0; r < rows; r++ {
		rowCards := make([]string, 0, cols)
		for c := 0; c < cols; c++ {
			idx := r*cols + c
			if idx >= len(g.cards) {
				// Fill remaining column slots with blank spacer
				rowCards = append(rowCards, strings.Repeat(" ", cardW))
				continue
			}
			card := g.cards[idx]
			card.SetFocused(idx == g.focused)
			rowCards = append(rowCards, card.Render(cardW))
		}

		// Normalise all cards in this row to the same height (max lines found)
		maxH := 0
		var allLines [][]string
		for _, rc := range rowCards {
			lines := strings.Split(rc, "\n")
			allLines = append(allLines, lines)
			if len(lines) > maxH {
				maxH = len(lines)
			}
		}

		// Pad shorter cards with blank lines
		var padded []string
		for _, lines := range allLines {
			for len(lines) < maxH {
				lines = append(lines, strings.Repeat(" ", cardW))
			}
			padded = append(padded, strings.Join(lines, "\n"))
		}

		// Build gap slice
		joinParts := make([]string, 0, len(padded)*2-1)
		for i, p := range padded {
			if i > 0 {
				joinParts = append(joinParts, strings.Repeat(" ", g.gap)) // gap column
			}
			joinParts = append(joinParts, p)
		}

		rowStr := lipgloss.JoinHorizontal(lipgloss.Top, joinParts...)
		rowStrs = append(rowStrs, rowStr)
	}

	return strings.Join(rowStrs, "\n")
}
