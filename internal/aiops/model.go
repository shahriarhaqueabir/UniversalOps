package aiops

import (
	tea "charm.land/bubbletea/v2"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// Model represents the AI Ops (AI Operations) layer.
type Model struct {
	tabIndex int // 0=chat, 1=reports
	input    string
	output   string
	messages []ChatMessage
	reports  []ReportSection
	err      error

	latestStats   *common.SystemStats
	metricHistory []common.SystemStats

	// Ollama status
	ollamaStatus *OllamaStatus
}

// NewModel creates a new AIOps model.
func NewModel() *Model {
	return &Model{
		tabIndex: 0,
	}
}

// Init initializes the AIOps layer.
func (m *Model) Init() tea.Cmd {
	return m.checkOllama()
}

// Messages for the AIOps layer.

// OllamaStatusMsg is sent when Ollama availability is checked.
type OllamaStatusMsg struct {
	Status *OllamaStatus
}

// ChatResultMsg is sent when a chat response is received.
type ChatResultMsg struct {
	Response string
	Err      error
}

// ReportSavedMsg is sent when a report is saved to disk.
type ReportSavedMsg struct {
	Path string
	Err  error
}

// TabIndex returns the current active tab.
func (m *Model) TabIndex() int {
	return m.tabIndex
}

// Error returns the last error, if any.
func (m *Model) Error() error {
	return m.err
}
