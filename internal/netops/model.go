package netops

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// Model represents the NetOps (Network Operations) layer.
type Model struct {
	// Collected data
	PingResult    *PingResult
	DNSResult     *DNSResult
	PortResults   []PortResult
	TraceResult   *TraceRouteResult
	Connections   []ConnectionInfo
	InterfaceData []InterfaceInfo
	err           error

	// UI state
	ready    bool
	tabIndex int // 0=ping, 1=dns, 2=portscan, 3=traceroute, 4=connections, 5=interfaces

	// Input state for ping target
	pingTarget string
	pingCount  int

	// Bandwidth state
	lastCounters  map[string]bandwidthCounter
	lastCapture   time.Time
	selectedIndex int

	// Workflow state
	workflowReport string
	showReport     bool
}

// NewModel creates a new NetOps model.
func NewModel() *Model {
	return &Model{
		tabIndex:  0,
		pingCount: 4,
	}
}

// Init initializes the NetOps layer.
func (m *Model) Init() tea.Cmd {
	return nil
}

// PingResultMsg is sent after a ping completes.
type PingResultMsg struct {
	Result *PingResult
	Err    error
}

// DNSResultMsg is sent after a DNS lookup completes.
type DNSResultMsg struct {
	Result *DNSResult
	Err    error
}

// PortScanResultMsg is sent after a port scan completes.
type PortScanResultMsg struct {
	Results []PortResult
	Err     error
}

// TraceRouteResultMsg is sent after a traceroute completes.
type TraceRouteResultMsg struct {
	Result *TraceRouteResult
	Err    error
}

// ConnectionsResultMsg is sent after collecting connections.
type ConnectionsResultMsg struct {
	Connections []ConnectionInfo
	Err         error
}

// InterfacesResultMsg is sent after collecting interfaces.
type InterfacesResultMsg struct {
	Interfaces []InterfaceInfo
	Counters   map[string]bandwidthCounter
	Err        error
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
