# Hawkward Card System Design

> A composable card component system for the Bubble Tea TUI, inspired by squib's card-based dashboard aesthetic.

---

## Design Principles

1. **Composable** — Cards are self-contained widgets that can be arranged in grids, stacks, or standalone
2. **Responsive** — Card grids auto-column based on terminal width (1 col at <80 chars, 2 at <120, 3+ at ≥120)
3. **Interactive** — Cards are focusable, selectable, and drill-downable
4. **Stateful** — Each card tracks its own state (loading, ready, error, streaming)
5. **Theme-aware** — All card colors come from `common.Palette`, respecting dark/light/high-contrast themes

---

## Card Types

### Base Card

```go
// Card is the foundational card component.
type Card struct {
    ID        string       // Unique identifier
    Title     string       // Card header
    Body      string       // Main content (rendered by card type)
    Footer    string       // Optional footer (key hints, timestamps)
    Status    CardStatus   // Current state
    Focusable bool         // Can this card receive focus?
    focused   bool         // Internal focus state
}

type CardStatus int

const (
    CardLoading  CardStatus = iota // Spinner / "Loading..."
    CardReady                       // Data loaded, rendering
    CardError                       // Error state with message
    CardStreaming                   // Live-updating (animation)
)
```

**Rendering**:
```
┌──────────────────────────────┐
│ Title                   ⬤   │  ← Status indicator (color dot)
│                              │
│  Body content here...        │  ← Rendered by specialized card
│                              │
│ footer / key hints           │  ← Optional
└──────────────────────────────┘
```

When focused:
```
┌──────────────────────────────┐  ← Highlighted border (Primary color)
│ Title                   ⬤   │
```

### Specialized Card Types

#### GaugeCard

Health metric display with integrated gauge widget.

```go
// NewGaugeCard creates a card with an embedded Gauge.
func NewGaugeCard(title string, gauge *charts.Gauge) *Card {
    // gauge.Render() provides the body
    // Card shows: title, gauge visualization, percentage, trend arrow
}

// SetValue updates the gauge value and re-renders.
func (c *Card) SetValue(pct float64) {
    c.gauge.SetValue(pct)
}
```

**Layout**:
```
┌──────────────────────────────┐
│ CPU                          │
│ ▓▓▓▓▓▓▓░░░ 73.2%    ↑ 2.1% │
│ Forecast: 78% in 15m         │
└──────────────────────────────┘
```

#### ChartCard

Card with embedded LineChart, BarChart, or AreaChart.

```go
// NewChartCard creates a card with an embedded chart component.
func NewChartCard(title string, chart interface{ Render() string }) *Card {
    // Chart renders to string, embedded as card body
    // Card shows: title, chart, optional legend
}

// AddSeries delegates to the embedded chart.
func (c *Card) AddSeries(name string, data []float64, color lipgloss.Color) {
    c.chart.AddSeries(name, data, color)
}
```

**Layout**:
```
┌──────────────────────────────┐
│ Network Throughput           │
│ ⣀⣤⣶⣿⣶⣤⣀⣤⣶⣿⣶⣤⣀ │
│ ────────────────▶            │
│ RX: 45.2 MB/s  TX: 12.3     │
└──────────────────────────────┘
```

#### TableCard

Card with a sortable data table.

```go
// NewTableCard creates a card with a sortable table.
func NewTableCard(title string, headers []string, rows [][]string) *Card

// SortByColumn sorts rows by the given column.
func (c *Card) SortByColumn(col int, asc bool)

// ExpandRow opens a detail pane below the row.
func (c *Card) ExpandRow(index int)
```

**Layout**:
```
┌──────────────────────────────┐
│ Top Processes       [sort▼] │
│ PID   NAME      CPU%  MEM   │
│ 1234  chrome    23.4  245   │
│ 5678  code      12.1  180   │
│ ...                         │
│ [enter] for detail          │
└──────────────────────────────┘
```

#### StatusCard

Compact card for quick status overviews.

```go
// NewStatusCard creates a compact card with icon, label, value, and sparkline.
func NewStatusCard(icon, title, value string, sparkline string, status CardStatus) *Card
```

**Layout** (normal):
```
┌──────────────────────┐
│ 🖥 45 processes      │
│ ⣀⣤⣶⣿            │
└──────────────────────┘
```

**Layout** (compact, collapsed mode):
```
🖥 45p ▓▓▓░░
```

---

## Card Layout Engine

### CardGrid

```go
// CardGrid arranges cards in a responsive grid.
type CardGrid struct {
    Cards   []*Card
    Columns int          // Auto-computed: 1, 2, or 3
    Gap     int          // Horizontal gap between cards (default: 1)
    focused int          // Index of focused card (-1 if none)
}

// NewCardGrid creates an empty grid.
func NewCardGrid(gap int) *CardGrid

// AddCard appends a card to the grid.
func (g *CardGrid) AddCard(card *Card)

// AutoColumns computes optimal column count based on terminal width.
func (g *CardGrid) AutoColumns(width int) int {
    switch {
    case width < 80:
        return 1    // Narrow: single column
    case width < 120:
        return 2    // Medium: two columns
    default:
        return 3    // Wide: three columns
    }
}

// Render renders all cards in the grid layout.
func (g *CardGrid) Render(termWidth int) string {
    cols := g.AutoColumns(termWidth)
    // Distribute cards into columns (left-to-right, top-to-bottom)
    // Render each column, join horizontally with Gap
    // Render each card, join vertically
    return rendered
}

// NextCard / PrevCard moves focus between cards.
func (g *CardGrid) NextCard() { g.focused = min(g.focused+1, len(g.Cards)-1) }
func (g *CardGrid) PrevCard() { g.focused = max(g.focused-1, 0) }

// FocusedCard returns the currently focused card, or nil.
func (g *CardGrid) FocusedCard() *Card
```

**Column distribution algorithm**:
```
Cards [A, B, C, D, E, F] in 3 columns:
  Col 1: A, D
  Col 2: B, E
  Col 3: C, F
  Each column has equal height (+/- 1 card)
  Columns rendered side by side with Gap
```

### Stack Layout

For detail views or sequential operations:

```go
// CardStack vertically stacks cards with full width.
func RenderCardStack(cards []*Card, width int) string {
    var b strings.Builder
    for i, card := range cards {
        b.WriteString(card.Render(width))
        if i < len(cards)-1 {
            b.WriteString("\n") // Separator
        }
    }
    return b.String()
}
```

### DetailPanel

Overlay panel for drill-down content. Captures all input until dismissed.

```go
// DetailPanel shows detailed content as a layered overlay.
type DetailPanel struct {
    Title      string
    Content    string
    Size       PanelSize      // Small, Medium, Large
    onClose    func()         // Called when panel is dismissed
}

type PanelSize int

const (
    PanelSmall  PanelSize = iota  // 40% of terminal width
    PanelMedium                    // 60% of terminal width
    PanelLarge                     // 80% of terminal width
)

// Render renders the detail panel as an overlay on the current view.
// The panel is centered and layered above the current content.
func (p *DetailPanel) Render(width, height int) string {
    panelWidth := int(float64(width) * p.panelWidthFraction())
    panelHeight := int(float64(height) * 0.8)
    // Use lipgloss.Place to center the panel
    // Content scrollable if exceeds panel height
    return lipgloss.Place(width, height,
        lipgloss.Center, lipgloss.Center,
        panelStyle.Render(content),
    )
}
```

---

## Card States

### State Transition Diagram

```
          Init
            │
            ▼
      ┌─────────┐
      │ Loading  │ ←── Spinner / "Collecting data..."
      └────┬────┘
           │
     ┌─────┴─────┐
     │           │
     ▼           ▼
 ┌───────┐  ┌───────┐
 │ Ready │  │ Error │ ←── Retry button on Enter
 └───┬───┘  └───────┘
     │           │
     ▼           │
 ┌─────────┐    │
 │Streaming│    │
 └─────────┘    │
     │          │
     └────┬─────┘
          │
          ▼
      ┌───────┐
      │Re-init│ (new data arrives)
      └───────┘
```

### Visual Indicators

| State | Visual | Body Content |
|-------|--------|-------------|
| `CardLoading` | ⏳ yellow pulsing dot | "Loading..." + spinner |
| `CardReady` | ● green dot | Rendered card data |
| `CardError` | ✕ red dot | Error message + "Press Enter to retry" |
| `CardStreaming` | ⟳ blue pulsing dot | Live-updating data with animation |

```go
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
```

---

## Card Interaction

### Focus Navigation

```go
// Card focus methods on CardGrid:
// Tab / Shift+Tab → NextCard() / PrevCard()
// Focused card has:
//   - Primary-colored border (replaces default Border)
//   - ">" prefix in title
//   - Key hints visible in footer

func (g *CardGrid) FocusedRender(termWidth int) string {
    // Same as Render() but focused card gets special styling
}

// Focused card border style:
var FocusedBorderStyle = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color(p.Primary)).
    Padding(1, 2)
```

### Selection & Drill-Down

```
Enter on focused card → activates card-specific action:
  - GaugeCard:     Show forecast detail / time-to-threshold
  - ChartCard:     Open expanded chart in DetailPanel
  - TableCard:     Open row detail in DetailPanel (if row selected)
  - StatusCard:    Open full view of that ops layer
```

```go
// HandleCardEnter processes Enter on the focused card.
func (g *CardGrid) HandleCardEnter() *DetailPanel {
    card := g.FocusedCard()
    if card == nil {
        return nil
    }
    
    switch card.Type {
    case "gauge":
        return &DetailPanel{
            Title: card.Title + " — Forecast Detail",
            Content: renderForecastDetail(card),
        }
    case "chart":
        return &DetailPanel{
            Title: card.Title + " — Expanded",
            Content: renderExpandedChart(card),
        }
    case "table":
        return &DetailPanel{
            Title: card.Title + " — Row Detail",
            Content: renderRowDetail(card, card.selectedRow),
        }
    case "status":
        // Navigate to ops layer screen
        return nil // handled by caller
    }
    return nil
}
```

### Key Bindings

```go
// Card navigation keys (added to KeyMap):
CardFocusNext  = key.NewBinding(key.WithKeys("tab"),      key.WithHelp("tab", "next card"))
CardFocusPrev  = key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("S+tab", "prev card"))
CardSelect     = key.NewBinding(key.WithKeys("enter"),     key.WithHelp("enter", "select card"))
CardDrillDown  = key.NewBinding(key.WithKeys("enter"),     key.WithHelp("enter", "drill down"))
CardContext    = key.NewBinding(key.WithKeys("ctrl+space"),key.WithHelp("C-space", "context menu"))
SplitToggle    = key.NewBinding(key.WithKeys("|"),         key.WithHelp("|", "toggle split pane"))
SplitResizeL   = key.NewBinding(key.WithKeys("["),         key.WithHelp("[", "shrink split"))
SplitResizeR   = key.NewBinding(key.WithKeys("]"),         key.WithHelp("]", "grow split"))
```

---

## Implementation Plan

### Files
- `internal/ui/cards.go` — Card, CardGrid, specialized card constructors, focus navigation, DetailPanel, SplitPane

### Integration Points

| Integration | File | What Changes |
|-------------|------|-------------|
| Root model | `internal/ui/root.go` | Add card grid state per screen; route Tab/Enter to card system |
| SysOps view | `internal/sysops/view.go` | Build card grid in View(); return grid.Render() |
| NetOps view | `internal/netops/view.go` | Build card grid per tab; return grid.Render() |
| SecOps view | `internal/secops/view.go` | Build card grids for each view section |
| DevOps view | `internal/devops/view.go` | Build session cards, log tail card |
| AIOps view | `internal/aiops/view.go` | Build chat cards, insight cards |
| Dashboard | `internal/ui/dashboard.go` | CardGrid for health gauges + op category cards |
| Types | `internal/common/types.go` | Add CardStatus enum |
| Styles | `internal/common/styles.go` | Add CardBg, FocusedBorder style primitives |
| Keys | `internal/ui/keys.go` | Add card navigation key bindings |

### Tests

| Test | Description |
|------|-------------|
| `TestCardGrid_Layout` | Verify 1/2/3 column distribution at various widths |
| `TestCard_Focus` | Focus navigation through cards, wrap-around |
| `TestCard_States` | Loading → Ready, Loading → Error transitions |
| `TestDetailPanel_Overlay` | Panel renders correctly over existing content |
| `TestSplitPane_Resize` | Split pane ratio changes with [ / ] keys |
| `TestTableCard_Sort` | Column sorting preserves data integrity |

---

*Last updated: 2026-07-07*
