package ui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/aiops"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/devops"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/netops"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/secops"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/sysops"
)

// RootModel is the top-level application model.
type RootModel struct {
	// Navigation
	activeScreen    common.Screen
	previousScreens []common.Screen

	// Components
	dashboard  *DashboardModel
	mainMenu   *MainMenu
	help       *HelpOverlay
	onboarding *OnboardingWizard
	statusBar  *StatusBar

	// Ops layers
	sysOps *sysops.Model
	netOps *netops.Model
	secOps *secops.Model
	devOps *devops.Model
	aiOps  *aiops.Model

	// Shared state
	width           int
	height          int
	keys            KeyMap
	stats           *common.SystemStats
	statsHistory    []common.SystemStats
	refreshInterval time.Duration // configurable dashboard refresh interval

	// Command palette
	commandPalette *CommandPalette

	// Data pipeline
	pipeline *common.DataPipeline
	alerts   *common.AlertEngine

	// Network bandwidth tracking (for pipeline ingestion)
	lastNetCounters interface{}
	lastNetCapture  time.Time
}

// NewRootModel creates a new root model.
func NewRootModel() *RootModel {
	startScreen := common.ScreenOnboarding
	if common.IsOnboarded() {
		startScreen = common.ScreenDashboard
	}

	_ = common.InitLogger("hawkward.log")

	// Load and set the saved theme
	common.SetTheme(common.LoadTheme())

	pipeline := common.NewDataPipeline(common.CollectionConfig{
		Capacity: 240,
	})
	alerts := common.NewAlertEngine(pipeline)
	alerts.AddDefaultRules()

	dash := NewDashboardModel()
	dash.SetPipeline(pipeline, alerts)

	m := &RootModel{
		activeScreen:    startScreen,
		dashboard:       dash,
		mainMenu:        NewMainMenu(),
		help:            NewHelpOverlay(),
		onboarding:      NewOnboardingWizard(),
		statusBar:       &StatusBar{},
		sysOps:          sysops.NewModel(),
		netOps:          netops.NewModel(),
		secOps:          secops.NewModel(),
		devOps:          devops.NewModel(),
		aiOps:           aiops.NewModel(),
		keys:            DefaultKeyMap(),
		refreshInterval: DefaultRefreshInterval,
		pipeline:        pipeline,
		alerts:          alerts,
	}

	// Command palette with callbacks (after m is initialized for closure access)
	m.commandPalette = NewCommandPalette(
		func(cmd Command) tea.Cmd {
			if cmd.Action != nil {
				return cmd.Action()
			}
			m.pushScreen(cmd.Screen)
			return nil
		},
		nil,
	)

	return m
}

// Init initializes the application.
func (m *RootModel) Init() tea.Cmd {
	return tea.Batch(
		m.sysOps.Init(),
		m.secOps.Init(),
		m.devOps.Init(),
		m.aiOps.Init(),
		common.StartTickCmd(m.refreshInterval),
	)
}

// Update handles all messages.
func (m *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.statusBar.Width = msg.Width
		m.dashboard.SetSize(msg.Width, msg.Height)

	case tea.KeyPressMsg:
		// Command palette takes priority when visible
		if m.commandPalette.IsVisible() {
			switch msg.String() {
			case "esc":
				m.commandPalette.Hide()
				return m, nil
			case "enter":
				cmd := m.commandPalette.SelectedCommand()
				if cmd != nil {
					m.commandPalette.Hide()
					if cmd.Action != nil {
						cmds = append(cmds, cmd.Action())
					} else {
						m.pushScreen(cmd.Screen)
					}
				}
				return m, tea.Batch(cmds...)
			default:
				m.commandPalette.HandleKey(msg)
				return m, nil
			}
		}

		// Help overlay gets first priority
		if m.help.HandleKey(msg) {
			return m, nil
		}

		// Onboarding gets next priority
		if !m.onboarding.IsComplete() {
			m.onboarding.HandleKey(msg)
			if m.onboarding.IsComplete() {
				m.activeScreen = common.ScreenDashboard
				m.statusBar.Screen = common.ScreenNames[common.ScreenDashboard]
			}
			return m, nil
		}

		// Global key handlers
		switch msg.String() {
		case "/":
			// Open command palette from any screen
			m.commandPalette.Show()
			return m, nil
		case "?":
			m.help.Toggle()
			return m, nil
		case "t":
			common.NextTheme()
			return m, nil
		case "d":
			// Direct nav to dashboard from any screen
			if m.activeScreen != common.ScreenDashboard {
				m.pushScreen(common.ScreenDashboard)
			} else {
				// Already on dashboard, consume the key
				return m, nil
			}
			return m, nil
		case "q", "ctrl+c":
			return m, tea.Quit
		}

		// Route to active screen
		cmd := m.routeKey(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case common.TickMsg:
		stats, err := m.sysOps.CollectStats()
		if err == nil {
			m.statsHistory = append(m.statsHistory, *stats)
			if len(m.statsHistory) > 24 {
				m.statsHistory = m.statsHistory[len(m.statsHistory)-24:]
			}

			// Build history for sparklines
			var cpuHist, memHist []float64
			for _, s := range m.statsHistory {
				cpuHist = append(cpuHist, s.CPUPercent)
				memHist = append(memHist, s.MemoryUsed)
			}
			stats.CPUHistory = cpuHist
			stats.MemHistory = memHist

			// Detect anomalies for status bar
			anomalies := aiops.DetectAnomalies(m.statsHistory)
			stats.AnomalyCount = len(anomalies)

			m.stats = stats
			m.statusBar.Stats = stats

			// ── Pipeline ingestion ──
			// Push core system metrics
			m.pipeline.PushMetric(common.MetricCPU, "%", stats.CPUPercent)
			m.pipeline.PushMetric(common.MetricMem, "%", stats.MemoryUsed)
			m.pipeline.PushMetric(common.MetricDisk, "%", stats.DiskUsed)
			m.pipeline.PushMetric(common.MetricProcCnt, "count", float64(stats.ProcessCount))

			// Push network bandwidth rates
			ifaces, netErr := m.netOps.CollectInterfaces()
			if netErr == nil && len(ifaces) > 0 {
				// Sum across all non-loopback interfaces for aggregate rates
				var rxTotal, txTotal float64
				for _, iface := range ifaces {
					rxTotal += iface.RXRateBps
					txTotal += iface.TXRateBps
				}
				m.pipeline.PushMetric(common.MetricNetRX, "bps", rxTotal)
				m.pipeline.PushMetric(common.MetricNetTX, "bps", txTotal)
			}

			// Evaluate alert rules
			newAlerts := m.alerts.Evaluate()
			stats.AnomalyCount += len(newAlerts)
			_ = newAlerts // available for notification display

			// Update dashboard with pipeline data
			m.dashboard.UpdateStats(stats)
		}
		cmds = append(cmds, common.StartTickCmd(m.refreshInterval))

	default:
		// Route message to active ops layer or dashboard
		cmd := m.routeMessage(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// routeKey routes a key press to the active screen.
// It handles navigation (esc to go back) and delegates to the active layer.
func (m *RootModel) routeKey(msg tea.KeyPressMsg) tea.Cmd {
	// Screen switching (number keys) - from dashboard or main menu
	if m.activeScreen == common.ScreenDashboard || m.activeScreen == common.ScreenMainMenu {
		switch msg.String() {
		case "1":
			m.pushScreen(common.ScreenSysOps)
			return nil
		case "2":
			m.pushScreen(common.ScreenNetOps)
			return nil
		case "3":
			m.pushScreen(common.ScreenSecOps)
			return nil
		case "4":
			m.pushScreen(common.ScreenDevOps)
			return nil
		case "5":
			m.pushScreen(common.ScreenAIOps)
			return nil
		}
	}

	// Esc: go back to previous screen
	if msg.String() == "esc" {
		m.popScreen()
		return nil
	}

	// Delegate to the active screen's handler
	switch m.activeScreen {
	case common.ScreenDashboard:
		return m.dashboard.Update(msg)

	case common.ScreenMainMenu:
		selected, screen := m.mainMenu.Update(msg)
		if selected {
			m.pushScreen(screen)
		}
		return nil

	case common.ScreenSysOps:
		return m.sysOps.Update(msg)

	case common.ScreenNetOps:
		return m.netOps.Update(msg)
	case common.ScreenSecOps:
		return m.secOps.Update(msg)
	case common.ScreenDevOps:
		return m.devOps.Update(msg)
	case common.ScreenAIOps:
		return m.aiOps.Update(msg)
	}

	return nil
}

// pushScreen navigates to a new screen, saving the current one.
func (m *RootModel) pushScreen(screen common.Screen) {
	common.LogInfo("Navigating to %s", common.ScreenNames[screen])
	m.previousScreens = append(m.previousScreens, m.activeScreen)
	m.activeScreen = screen
	m.statusBar.Screen = common.ScreenNames[screen]
}

// popScreen returns to the previous screen.
func (m *RootModel) popScreen() {
	if len(m.previousScreens) > 0 {
		prev := m.previousScreens[len(m.previousScreens)-1]
		m.previousScreens = m.previousScreens[:len(m.previousScreens)-1]
		m.activeScreen = prev
		common.LogInfo("Returning to %s", common.ScreenNames[prev])
	} else {
		m.activeScreen = common.ScreenDashboard
		common.LogInfo("Returning to Dashboard")
	}
	m.statusBar.Screen = common.ScreenNames[m.activeScreen]
}

// routeMessage routes non-keyboard messages to the active ops layer.
func (m *RootModel) routeMessage(msg tea.Msg) tea.Cmd {
	switch m.activeScreen {
	case common.ScreenDashboard:
		return m.dashboard.Update(msg)
	case common.ScreenSysOps:
		return m.sysOps.Update(msg)
	case common.ScreenNetOps:
		return m.netOps.Update(msg)
	case common.ScreenSecOps:
		return m.secOps.Update(msg)
	case common.ScreenDevOps:
		return m.devOps.Update(msg)
	case common.ScreenAIOps:
		return m.aiOps.Update(msg)
	}
	return nil
}

// View renders the current screen.
func (m *RootModel) View() tea.View {
	var content strings.Builder

	// Onboarding always takes priority
	if !m.onboarding.IsComplete() {
		content.WriteString(m.onboarding.Render())
	} else {
		// Render the active screen
		switch m.activeScreen {
		case common.ScreenDashboard:
			content.WriteString(m.dashboard.View())

		case common.ScreenMainMenu:
			m.mainMenu.SetSize(m.width, m.height)
			content.WriteString(m.mainMenu.Render())

		case common.ScreenSysOps:
			content.WriteString(m.sysOps.View(m.width, m.height, m.stats))

		case common.ScreenNetOps:
			content.WriteString(m.netOps.View(m.width, m.height))

		case common.ScreenSecOps:
			content.WriteString(m.secOps.View(m.width, m.height, m.stats))

		case common.ScreenDevOps:
			content.WriteString(m.devOps.View(m.width, m.height, m.stats))

		case common.ScreenAIOps:
			content.WriteString(m.aiOps.View(m.width, m.height, m.stats))
		}
	}

	// Status bar (always visible)
	m.statusBar.Screen = common.ScreenNames[m.activeScreen]
	content.WriteString("\n")
	content.WriteString(m.statusBar.Render())

	// Help overlay (takes over the full content if visible)
	if m.help.Visible() {
		helpContent := m.help.Render()
		content.Reset()
		content.WriteString(helpContent)
		content.WriteString("\n")
		content.WriteString(m.statusBar.Render())
	}

	// Command palette overlay (rendered on top when visible)
	if m.commandPalette.IsVisible() {
		paletteView := m.commandPalette.View(m.width, m.height)
		if paletteView != "" {
			// Insert palette at the top of the content
			combined := paletteView + "\n" + content.String()
			v := tea.NewView(combined)
			v.AltScreen = true
			v.MouseMode = tea.MouseModeCellMotion
			return v
		}
	}

	v := tea.NewView(content.String())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// DefaultRefreshInterval for dashboard updates.
const DefaultRefreshInterval = 3 * time.Second

// RefreshInterval returns the current dashboard refresh interval.
func (m *RootModel) RefreshInterval() time.Duration {
	return m.refreshInterval
}

// SetRefreshInterval sets the dashboard refresh interval.
// Returns the previous interval value.
func (m *RootModel) SetRefreshInterval(d time.Duration) time.Duration {
	prev := m.refreshInterval
	m.refreshInterval = d
	return prev
}
