package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// StatusBar renders the bottom status bar with system health and navigation hints.
type StatusBar struct {
	Stats  *common.SystemStats
	Screen string
	Width  int
}

// Render builds the status bar string.
func (s *StatusBar) Render() string {
	if s == nil {
		return ""
	}

	// Left: current screen
	left := StatusBarStyle.Render(fmt.Sprintf(" %s ", s.Screen))

	// Center: quick health
	cpuPct := 0.0
	memPct := 0.0
	diskPct := 0.0
	if s.Stats != nil {
		cpuPct = s.Stats.CPUPercent
		memPct = s.Stats.MemoryUsed
		diskPct = s.Stats.DiskUsed
	}

	center := fmt.Sprintf(" CPU:%s | MEM:%s | DISK:%s ",
		common.FormatPercent(cpuPct),
		common.FormatPercent(memPct),
		common.FormatPercent(diskPct),
	)

	// Color the center section based on health
	centerStyle := StatusGoodStyle
	if cpuPct > 90 || memPct > 90 || diskPct > 90 {
		centerStyle = StatusBadStyle
	} else if cpuPct > 70 || memPct > 70 || diskPct > 70 {
		centerStyle = StatusWarnStyle
	}
	center = centerStyle.Render(center)

	// Anomaly indicator
	anomalyText := ""
	if s.Stats != nil && s.Stats.AnomalyCount > 0 {
		anomalyText = lipgloss.NewStyle().
			Foreground(lipgloss.Color(common.CurrentPalette().Danger)).
			Bold(true).
			Render(fmt.Sprintf(" ⚠ %d ANOMALIES ", s.Stats.AnomalyCount))
	}

	// Right: keyboard hints
	themeName := strings.ToUpper(string(common.CurrentTheme()))
	right := StatusBarStyle.Render(fmt.Sprintf(" [%s] [?] Help  [q] Quit  [1-5] Switch ", themeName))

	// Calculate padding to fill the width
	barWidth := s.Width
	leftW := lipgloss.Width(left)
	centerW := lipgloss.Width(center)
	anomalyW := lipgloss.Width(anomalyText)
	rightW := lipgloss.Width(right)
	totalW := leftW + centerW + anomalyW + rightW

	if totalW < barWidth {
		padding := common.RepeatString(" ", barWidth-totalW)
		return left + center + anomalyText + padding + right
	}

	return left + center + anomalyText + right
}
