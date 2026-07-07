package ui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common/charts"
)

// ── DashboardModel ──────────────────────────────────────────────────────────

// DashboardModel is the landing-page view which replaces the old main menu. It
// displays system health gauges, quick-access operation cards, and live
// time-series sparklines.
type DashboardModel struct {
	// Health gauge cards for each ops layer
	SysHealth  *GaugeCard
	MemHealth  *GaugeCard
	DiskHealth *GaugeCard
	NetHealth  *GaugeCard
	SecHealth  *GaugeCard

	// Quick-access operation cards (one per ops layer)
	QuickAccess *CardGrid

	// Live sparkline cards for key metrics
	CPUSpark  *StatusCard
	MemSpark  *StatusCard
	DiskSpark *StatusCard

	// Grid combining gauges + quick access + sparklines as one navigable grid
	grid     *CardGrid
	width    int
	height   int
	stats    *common.SystemStats
	store    *common.TimeSeriesStore
	pipeline *common.DataPipeline
	alerts   *common.AlertEngine
}

// NewDashboardModel creates the dashboard with initialized health gauges,
// operation cards, and sparkline cards.
func NewDashboardModel() *DashboardModel {
	d := &DashboardModel{
		grid:  NewCardGrid(1),
		store: common.NewTimeSeriesStore(60),
	}

	// ── Health gauges ──
	d.SysHealth = NewGaugeCard("CPU")
	d.MemHealth = NewGaugeCard("MEM")
	d.DiskHealth = NewGaugeCard("DISK")
	d.NetHealth = NewGaugeCard("NET")
	d.SecHealth = NewGaugeCard("SEC")

	d.SysHealth.SetLabel("System load")
	d.MemHealth.SetLabel("Memory usage")
	d.DiskHealth.SetLabel("Disk usage")
	d.NetHealth.SetLabel("Network activity")
	d.SecHealth.SetLabel("Security posture")

	// Set initial values
	d.SysHealth.SetValue(0)
	d.MemHealth.SetValue(0)
	d.DiskHealth.SetValue(0)
	d.NetHealth.SetValue(0)
	d.SecHealth.SetValue(0)

	// ── Quick-access operation cards ──
	d.QuickAccess = NewCardGrid(1)
	d.QuickAccess.AddCard(NewOperationCard("🖥", "System Operations",
		"CPU, memory, disk, processes, system info", common.ScreenSysOps))
	d.QuickAccess.AddCard(NewOperationCard("🌐", "Network Operations",
		"Ping, traceroute, port scan, DNS, connections", common.ScreenNetOps))
	d.QuickAccess.AddCard(NewOperationCard("🔒", "Security Operations",
		"Firewall, users, listening ports, defender, tasks", common.ScreenSecOps))
	d.QuickAccess.AddCard(NewOperationCard("⚙", "Development Operations",
		"Command runner, log tailer, file browser", common.ScreenDevOps))
	d.QuickAccess.AddCard(NewOperationCard("🤖", "AI Operations",
		"Local AI assistant, report generation", common.ScreenAIOps))

	// ── Sparkline status cards ──
	d.CPUSpark = NewStatusCard("CPU Trend")
	d.CPUSpark.SetLabel("CPU %")
	d.CPUSpark.SetValue("--")
	d.MemSpark = NewStatusCard("MEM Trend")
	d.MemSpark.SetLabel("MEM %")
	d.MemSpark.SetValue("--")
	d.DiskSpark = NewStatusCard("DISK Trend")
	d.DiskSpark.SetLabel("DISK %")
	d.DiskSpark.SetValue("--")

	return d
}

// SetPipeline attaches the data pipeline and alert engine to the dashboard,
// replacing the local TimeSeriesStore as the data source for trends/forecasts.
func (d *DashboardModel) SetPipeline(p *common.DataPipeline, a *common.AlertEngine) {
	d.pipeline = p
	d.alerts = a
}

// SetSize updates the dashboard dimensions for responsive layout.
func (d *DashboardModel) SetSize(w, h int) {
	d.width = w
	d.height = h
}

// UpdateStats refreshes dashboard data with the latest system stats and
// pushes values into the time-series store.
func (d *DashboardModel) UpdateStats(stats *common.SystemStats) {
	d.stats = stats

	// Push into time-series store
	now := time.Now()
	d.store.Push("cpu", "%", now, stats.CPUPercent)
	d.store.Push("mem", "%", now, stats.MemoryUsed)
	d.store.Push("disk", "%", now, stats.DiskUsed)
	d.store.Push("net", "conn", now, float64(stats.ProcessCount))         // proxy metric
	d.store.Push("sec", "score", now, float64(100-stats.AnomalyCount*10)) // inverse anomaly

	// Update gauge values
	d.SysHealth.SetValue(stats.CPUPercent)
	d.MemHealth.SetValue(stats.MemoryUsed)
	d.DiskHealth.SetValue(stats.DiskUsed)
	// NET and SEC are estimated from available data
	netVal := float64(stats.ProcessCount)
	if netVal > 100 {
		netVal = 100
	}
	d.NetHealth.SetValue(netVal)
	secVal := float64(100 - stats.AnomalyCount*10)
	if secVal < 0 {
		secVal = 0
	}
	d.SecHealth.SetValue(secVal)

	// ── Pipeline-based trend info ──
	if d.pipeline != nil {
		// CPU trend from pipeline forecast engine
		if trend := d.pipeline.GetTrend(common.MetricCPU); trend.Direction != common.TrendStable {
			dir := toChartTrend(trend.Direction)
			d.SysHealth.SetTrend(dir, trend.ChangePct)
		}
		if trend := d.pipeline.GetTrend(common.MetricMem); trend.Direction != common.TrendStable {
			dir := toChartTrend(trend.Direction)
			d.MemHealth.SetTrend(dir, trend.ChangePct)
		}
		if trend := d.pipeline.GetTrend(common.MetricDisk); trend.Direction != common.TrendStable {
			dir := toChartTrend(trend.Direction)
			d.DiskHealth.SetTrend(dir, trend.ChangePct)
		}
	} else {
		// Fallback: simple trend from local store
		if cpuTS := d.store.Get("cpu", "%"); cpuTS.Count() >= 2 {
			cpuVals := cpuTS.Values()
			change := cpuVals[len(cpuVals)-1] - cpuVals[0]
			dir := charts.TrendStable
			pct := 0.0
			if change > 0.5 {
				dir = charts.TrendRising
				pct = change
			} else if change < -0.5 {
				dir = charts.TrendFalling
				pct = -change
			}
			d.SysHealth.SetTrend(dir, pct)
		}
	}

	// Update sparkline cards (prefer pipeline data when available)
	if d.pipeline != nil {
		if ts := d.pipeline.GetTimeSeries(common.MetricCPU); ts != nil && ts.Count() > 0 {
			vals := ts.Values()
			d.CPUSpark.SetValue(fmt.Sprintf("%.1f%%", vals[len(vals)-1]))
			d.CPUSpark.SetSparklineData(vals)
		}
		if ts := d.pipeline.GetTimeSeries(common.MetricMem); ts != nil && ts.Count() > 0 {
			vals := ts.Values()
			d.MemSpark.SetValue(fmt.Sprintf("%.1f%%", vals[len(vals)-1]))
			d.MemSpark.SetSparklineData(vals)
		}
		if ts := d.pipeline.GetTimeSeries(common.MetricDisk); ts != nil && ts.Count() > 0 {
			vals := ts.Values()
			d.DiskSpark.SetValue(fmt.Sprintf("%.1f%%", vals[len(vals)-1]))
			d.DiskSpark.SetSparklineData(vals)
		}
		// Forecast indicator text on sparkline cards
		if cpuMF := d.pipeline.GetMetricWithForecast(common.MetricCPU); len(cpuMF.Forecast) > 0 {
			nextVal := cpuMF.Forecast[len(cpuMF.Forecast)-1]
			d.CPUSpark.SetLabel(fmt.Sprintf("Forecast: %.1f%%", nextVal))
		}
		if memMF := d.pipeline.GetMetricWithForecast(common.MetricMem); len(memMF.Forecast) > 0 {
			nextVal := memMF.Forecast[len(memMF.Forecast)-1]
			d.MemSpark.SetLabel(fmt.Sprintf("Forecast: %.1f%%", nextVal))
		}
	} else {
		// Fallback to local store
		if cpuTS := d.store.Get("cpu", "%"); cpuTS.Count() > 0 {
			vals := cpuTS.Values()
			last := vals[len(vals)-1]
			d.CPUSpark.SetValue(fmt.Sprintf("%.1f%%", last))
			d.CPUSpark.SetSparklineData(vals)
		}
		if memTS := d.store.Get("mem", "%"); memTS.Count() > 0 {
			vals := memTS.Values()
			last := vals[len(vals)-1]
			d.MemSpark.SetValue(fmt.Sprintf("%.1f%%", last))
			d.MemSpark.SetSparklineData(vals)
		}
		if diskTS := d.store.Get("disk", "%"); diskTS.Count() > 0 {
			vals := diskTS.Values()
			last := vals[len(vals)-1]
			d.DiskSpark.SetValue(fmt.Sprintf("%.1f%%", last))
			d.DiskSpark.SetSparklineData(vals)
		}
	}
}

// Update handles messages directed at the dashboard.
func (d *DashboardModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "tab", "right", "l":
			d.grid.NextCard()
		case "shift+tab", "left", "h":
			d.grid.PrevCard()
		case "up", "k":
			// Move up one row in the grid
			d.moveFocusVertical(-1)
		case "down", "j":
			// Move down one row
			d.moveFocusVertical(1)
		}
	}
	return nil
}

// moveFocusVertical shifts focus by one row (up = -1, down = +1). It
// calculates the current row/col from the grid's card count and columns.
func (d *DashboardModel) moveFocusVertical(dir int) {
	if d.grid.Len() == 0 {
		return
	}
	focused := d.grid.Focused()
	if focused < 0 {
		focused = 0
	}

	cols := d.grid.AutoColumns(d.width)
	if cols < 1 {
		cols = 1
	}
	rows := (d.grid.Len() + cols - 1) / cols
	if rows < 1 {
		rows = 1
	}

	row := focused / cols
	col := focused % cols

	newRow := row + dir
	if newRow < 0 {
		newRow = rows - 1
	} else if newRow >= rows {
		newRow = 0
	}

	newIdx := newRow*cols + col
	if newIdx >= d.grid.Len() {
		// Clamp to last card
		newIdx = d.grid.Len() - 1
	}
	d.grid.SetFocused(newIdx)
}

// View renders the complete dashboard.
func (d *DashboardModel) View() string {
	if d.width <= 0 {
		d.width = 80
	}
	if d.height <= 0 {
		d.height = 24
	}
	p := common.CurrentPalette()

	var b strings.Builder

	// ── Dashboard title ──
	b.WriteString(DashboardTitleStyle.Render("🚀 Hawkward Dashboard"))
	b.WriteByte('\n')

	// Alert count indicator
	alertCount := 0
	if d.alerts != nil {
		alertCount = d.alerts.AlertCount()
	}
	subtitle := "Real-time system overview  •  [Tab/↑↓] navigate cards  •  [1-5] jump to layer  •  [?] help"
	if alertCount > 0 {
		alertStr := lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Danger)).
			Bold(true).
			Render(fmt.Sprintf("⚠ %d alert(s)", alertCount))
		subtitle = subtitle + "  •  " + alertStr
	}
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Muted)).
		Render(subtitle))
	b.WriteByte('\n')
	b.WriteByte('\n')

	// ── Health gauge row ──
	b.WriteString(DashboardSectionStyle.Render("System Health"))
	b.WriteByte('\n')

	gaugeW := (d.width - 4) / 5
	if gaugeW < 14 {
		gaugeW = 14
	}
	// Adjust for narrow terminals
	if d.width < 80 {
		gaugeW = (d.width - 4) / 3
		if gaugeW < 14 {
			gaugeW = 14
		}
	}

	gauges := []Card{d.SysHealth, d.MemHealth, d.DiskHealth, d.NetHealth, d.SecHealth}
	var gaugeStrs []string
	for i, g := range gauges {
		g.SetFocused(false) // gauges aren't focus-navigable individually
		rendered := g.Render(gaugeW)
		gaugeStrs = append(gaugeStrs, rendered)

		_ = i
	}

	// Normalise gauge row heights
	var gaugeMaxH int
	var gaugeLines [][]string
	for _, gs := range gaugeStrs {
		lines := strings.Split(gs, "\n")
		gaugeLines = append(gaugeLines, lines)
		if len(lines) > gaugeMaxH {
			gaugeMaxH = len(lines)
		}
	}
	for i, lines := range gaugeLines {
		for len(lines) < gaugeMaxH {
			lines = append(lines, strings.Repeat(" ", gaugeW))
		}
		gaugeStrs[i] = strings.Join(lines, "\n")
	}

	rowParts := make([]string, 0, len(gaugeStrs)*2-1)
	for i, gs := range gaugeStrs {
		if i > 0 {
			rowParts = append(rowParts, strings.Repeat(" ", 1))
		}
		rowParts = append(rowParts, gs)
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, rowParts...))
	b.WriteByte('\n')
	b.WriteByte('\n')

	// ── Quick-access operation cards ──
	b.WriteString(DashboardSectionStyle.Render("Operations"))
	b.WriteByte('\n')

	// Rebuild the grid's card list each render cycle
	d.grid = NewCardGrid(1)
	for _, c := range d.QuickAccess.Cards() {
		d.grid.AddCard(c)
	}
	b.WriteString(d.grid.Render(d.width - 2))
	b.WriteByte('\n')

	// ── Sparkline row ──
	b.WriteString(DashboardSectionStyle.Render("Live Metrics"))
	b.WriteByte('\n')

	sparkW := (d.width - 4) / 3
	if sparkW < 16 {
		sparkW = 16
	}
	sparks := []Card{d.CPUSpark, d.MemSpark, d.DiskSpark}
	var sparkStrs []string
	for _, s := range sparks {
		s.SetFocused(false)
		sparkStrs = append(sparkStrs, s.Render(sparkW))
	}

	// Normalise sparkline heights
	var sparkMaxH int
	var sparkLines [][]string
	for _, ss := range sparkStrs {
		lines := strings.Split(ss, "\n")
		sparkLines = append(sparkLines, lines)
		if len(lines) > sparkMaxH {
			sparkMaxH = len(lines)
		}
	}
	for i, lines := range sparkLines {
		for len(lines) < sparkMaxH {
			lines = append(lines, strings.Repeat(" ", sparkW))
		}
		sparkStrs[i] = strings.Join(lines, "\n")
	}

	sparkParts := make([]string, 0, len(sparkStrs)*2-1)
	for i, ss := range sparkStrs {
		if i > 0 {
			sparkParts = append(sparkParts, strings.Repeat(" ", 1))
		}
		sparkParts = append(sparkParts, ss)
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, sparkParts...))
	b.WriteByte('\n')

	return b.String()
}

// toChartTrend converts a common.TrendDirection to the charts package type.
// Both have identical numeric values, so this is a direct cast.
func toChartTrend(d common.TrendDirection) charts.TrendDirection {
	return charts.TrendDirection(d)
}
