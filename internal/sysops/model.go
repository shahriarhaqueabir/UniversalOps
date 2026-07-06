package sysops

import (
	tea "charm.land/bubbletea/v2"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// Model represents the SysOps (System Operations) layer.
type Model struct {
	// Collected data
	stats *common.SystemStats
	err   error

	// UI state
	ready    bool
	tabIndex int // 0=overview, 1=processes, 2=details

	// Workflow state
	workflowReport string
	showReport     bool
}

// NewModel creates a new SysOps model.
func NewModel() *Model {
	return &Model{
		tabIndex: 0,
	}
}

// Init initializes the SysOps layer.
func (m *Model) Init() tea.Cmd {
	return func() tea.Msg {
		stats, err := CollectAllStats()
		if err != nil {
			return StatsResult{Stats: nil, Err: err}
		}
		return StatsResult{Stats: stats, Err: nil}
	}
}

// StatsResult is the message returned after collecting stats.
type StatsResult struct {
	Stats *common.SystemStats
	Err   error
}

// CollectStats returns the current stats.
func (m *Model) CollectStats() (*common.SystemStats, error) {
	return CollectAllStats()
}

// Ready returns true if data has been collected at least once.
func (m *Model) Ready() bool {
	return m.ready
}

// TabIndex returns the current active tab.
func (m *Model) TabIndex() int {
	return m.tabIndex
}

// Error returns the last error, if any.
func (m *Model) Error() error {
	return m.err
}
