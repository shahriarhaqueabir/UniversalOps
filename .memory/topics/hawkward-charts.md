# Hawkward Chart Library Design

This document describes the chart rendering engine — a pure Lip Gloss + Unicode visualization library with zero external dependencies.

## Design Principles

1. **No external charting deps** — single binary constraint. All rendering uses Lip Gloss styling + Unicode block/braille characters.
2. **Character-level precision** — Every cell in the terminal is a rendering primitive (8×16 pixels effectively).
3. **Automatic scaling** — Charts auto-compute range, tick marks, and labels from data.
4. **Theme-aware** — All chart colors come from the active `Palette`, respecting dark/light/high-contrast themes.
5. **Composable** — Charts return `lipgloss.Style`-compatible strings that can be embedded in cards, panels, or standalone.

---

## Unicode Rendering Primitives

### Braille Line Charts (8-dot braille)
```
Characters: ⠀ ⠁ ⠂ ⠃ ⠄ ⠅ ⠆ ⠇ ⠈ ⠉ ⠊ ⠋ ⠌ ⠍ ⠎ ⠏
            ⠐ ⠑ ⠒ ⠓ ⠔ ⠕ ⠖ ⠗ ⠘ ⠙ ⠚ ⠛ ⠜ ⠝ ⠞ ⠟
            ⠠ ⠡ ⠢ ⠣ ⠤ ⠥ ⠦ ⠧ ⠨ ⠩ ⠪ ⠫ ⠬ ⠭ ⠮ ⠯
            ⠰ ⠱ ⠲ ⠳ ⠴ ⠵ ⠶ ⠷ ⠸ ⠹ ⠺ ⠻ ⠼ ⠽ ⠾ ⠿
            ⡀ ⡁ ⡂ ... ⣿
```
Each braille cell is 2×4 dots (8 total). This gives 4 vertical positions per terminal row.

**Rendering algorithm for line charts:**
1. Map data values to Y positions in the braille grid
2. For each X position, compute which of the 8 dots should be lit
3. Combine adjacent 4-dot columns into braille characters
4. Result: effectively 2× horizontal resolution and 4× vertical resolution vs block chars

### Block Element Bar Charts
```
8-level bar: ▁▂▃▄▅▆▇█
```
Each character is 1/8 of full height. Used for:
- Horizontal bar charts (single row per data point)
- Vertical bar charts (stack characters)
- Gauges/progress bars
- Sparklines (compact trend)

### Box Drawing for Charts
```
Grid lines:   ─ │ ┌ ┐ └ ┘ ├ ┤ ┬ ┴ ┼
Thick:        ━ ┃ ┏ ┓ ┗ ┛ ┣ ┫ ┳ ┻ ╋
Double:       ═ ║ ╔ ╗ ╚ ╝ ╠ ╣ ╦ ╩ ╬
```
Used for chart frames, axis lines, grid backgrounds.

---

## Component Specifications

### LineChart

```
  ⣀⣤⣶⣿⣶⣤⣀⣤⣶⣿   CPU Usage (last 60 samples)
 90% ┤    ⣀⣤⣶⣿⣶⣤⣀
     │  ⣀⣤⣶⣿  ⣀⣤⣶⣿
     │ ⣀⣶⣿⣶⣶⣤⣀
     │⣀⣿⣶⣤⣀⣀
 10% ┼─────────────────────▶
     13:00        13:01
```

| Method | Description |
|--------|-------------|
| `NewLineChart(config)` | Create with config: width, height, series colors, labels |
| `AddSeries(name string, data []float64, color lipgloss.Color)` | Add data series |
| `SetForecast(data []float64)` | Overlay forecast as dashed region |
| `SetThreshold(value float64, label string)` | Add threshold line (e.g., 90% CPU) |
| `Render() string` | Return chart as styled string |
| `CursorAt(index int) string` | Tooltip at cursor position |

**Auto-features:**
- Y-axis labels (auto-scaled 0, 25, 50, 75, 100 or computed)
- X-axis time labels (auto-formatted: HH:MM, MM:SS)
- Grid lines (configurable)
- Legend (color + name)
- Multiple series (up to 6 with distinct colors)
- Threshold line with label
- Forecast overlay (dashed/predicted region)

### BarChart

```
  Process CPU Usage (top 5)
  chrome.exe  ████████████ 23.4%
  code.exe    ██████░░░░░░ 12.1%
  terminal    ████░░░░░░░░  8.3%
  slack.exe   ██░░░░░░░░░░  4.2%
  explorer    █░░░░░░░░░░░  2.1%
```

| Method | Description |
|--------|-------------|
| `NewBarChart(config)` | Horizontal or vertical |
| `AddBar(label string, value float64, color lipgloss.Color)` | Add single bar |
| `AddGroup(label string, values []float64, colors []lipgloss.Color)` | Grouped/stacked bars |
| `SortByValue(asc bool)` | Sort bars |
| `Render() string` | |

### Gauge

```
  CPU: ▓▓▓▓▓▓▓░░░ 73%
  ▓▓▓▓▓▓▓░░░░░░░░░░░░░░░░░░░░
  └── green ──┘└─yellow──┘└red┘
      0-70%     70-90%    90-100%
```

| Method | Description |
|--------|-------------|
| `NewGauge(config)` | Horizontal or vertical |
| `SetValue(pct float64)` | Set 0-100 |
| `SetThresholds(warn, crit float64)` | Color zone thresholds |
| `SetLabel(text string)` | Label below gauge |
| `ShowPercent(show bool)` | Toggle percentage display |
| `Render() string` | |

### Sparkline v2 (enhanced)

```
  ⣀⣤⣶⣿⣶⣤⣀⣤⣶⣿⣶⣤⣀  ↑ 12.3%
  min: 23.1    max: 87.4    avg: 45.2
```

| Method | Description |
|--------|-------------|
| `NewSparkline(width int)` | Create compact trend |
| `SetValues(data []float64)` | Set data |
| `ShowMinMax(show bool)` | Toggle min/max markers |
| `ShowTrend(show bool)` | Show trend arrow + percent |
| `ShowLabels(show bool)` | Show start/end labels |
| `Render() string` | |

### HeatMap

```
  Port Activity (last 60s)
  ┌────┬────┬────┬────┬────┬────┐
  │ ██ │ ░░ │ ██ │ ░░ │ ░░ │ ██ │ 80-443
  │ ██ │ ██ │ ░░ │ ░░ │ ██ │ ░░ │ 444-1023
  │ ░░ │ ░░ │ ░░ │ ██ │ ░░ │ ░░ │ 1024-49151
  │ ░░ │ ░░ │ ░░ │ ░░ │ ░░ │ ░░ │ 49152+
  └────┴────┴────┴────┴────┴────┘
     0s         30s        60s
```

| Method | Description |
|--------|-------------|
| `NewHeatMap(rows, cols int)` | Create grid |
| `SetCell(row, col int, value float64)` | Set cell value |
| `SetRowLabel(row int, label string)` | Y-axis label |
| `SetColLabel(col int, label string)` | X-axis label |
| `ColorScheme(scheme HeatScheme)` | Color interpolation |
| `Render() string` | |

### NumericDisplay

```
  ┌──────────┐
  │  73.2%   │
  │  CPU     │
  │  ↑ 2.1%  │
  └──────────┘
```

| Method | Description |
|--------|-------------|
| `NewNumericDisplay(config)` | Large centered number |
| `SetValue(value float64, unit string)` | Main value |
| `SetTrend(dir TrendDirection, pct float64)` | Trend indicator |
| `SetColor(color lipgloss.Color)` | Value color |
| `Render() string` | |

---

## Data Pipeline Integration

```
gopsutil (every 3s)
  │
  ▼
Collector (collects CPU, MEM, DISK, NET, PROC)
  │
  ▼
TimeSeriesStore (ring buffer, 240 samples @ 3s = 12min window)
  ├─ CPU: [45.2, 46.1, 47.3, ...]
  ├─ MEM: [52.1, 52.0, 51.8, ...]
  ├─ DISK: [38.1, 38.1, 38.2, ...]
  ├─ NET RX: [1.2e6, 1.3e6, ...]
  └─ NET TX: [0.8e6, 0.9e6, ...]
  │
  ▼
Forecast Engine
  ├─ LinearRegression → trend slope + projection
  ├─ ExponentialSmoothing → next N values
  └─ ThresholdPrediction → "time to 90%"
  │
  ▼
Chart Components (Render)
  ├─ LineChart(data, forecast)
  ├─ Gauge(current, thresholds)
  └─ Sparkline(history)
  │
  ▼
Lip Gloss styled string → BubbleTea View()
```

---

## Theme Integration

Each palette gains 4 chart-specific colors:

```go
type Palette struct {
    // Existing
    Primary, Secondary, Warning, Danger, Info string
    Muted, Text, Background, Border          string
    
    // New for charts
    ChartLine1 string  // First series line color
    ChartLine2 string  // Second series line color  
    ChartLine3 string  // Third series line color
    ChartGrid  string  // Grid line color
}
```

All 10 themes define these chart colors to ensure cohesive visualization across themes.

---

## Rendering Strategy

For performance, charts use a **two-pass rendering** approach:

1. **Layout pass**: Calculate dimensions, grid positions, label widths, content box
2. **Render pass**: Build strings using Lip Gloss styles, compose final output

This enables:
- Efficient re-rendering (only data changes trigger a re-render)
- Width/height adaptation on terminal resize
- Style reuse across chart types
