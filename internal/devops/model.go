package devops

import (
	tea "charm.land/bubbletea/v2"
)

// Model represents the DevOps (Development Operations) layer.
type Model struct {
	tabIndex  int // 0=shell, 1=logs, 2=files, 3=processes, 4=services
	input     string
	output    string
	err       error
	files     []FileEntry
	logs      []string
	results   []ShellResult
	processes []ProcessEntry
	services  []ServiceEntry

	// Workflow state
	workflowReport string
	showReport     bool
}

// NewModel creates a new DevOps model.
func NewModel() *Model {
	return &Model{
		tabIndex: 0,
	}
}

// Init initializes the DevOps layer.
func (m *Model) Init() tea.Cmd {
	return nil
}

// Messages for the DevOps layer.

// CommandResultMsg is sent when a shell command completes.
type CommandResultMsg struct {
	Result *ShellResult
	Err    error
}

// LogResultMsg is sent when log content is read.
type LogResultMsg struct {
	Lines []string
	Path  string
	Err   error
}

// FileListMsg is sent when a directory listing is complete.
type FileListMsg struct {
	Files []FileEntry
	Path  string
	Err   error
}

// FileContentMsg is sent when a file is read.
type FileContentMsg struct {
	Content string
	Path    string
	Err     error
}

// ProcessListMsg is sent when process listing completes.
type ProcessListMsg struct {
	Processes []ProcessEntry
	Err       error
}

// ProcessActionMsg is sent when a process action completes.
type ProcessActionMsg struct {
	Message string
	Err     error
}

// ServiceListMsg is sent when service listing completes.
type ServiceListMsg struct {
	Services []ServiceEntry
	Err      error
}

// TabIndex returns the current active tab.
func (m *Model) TabIndex() int {
	return m.tabIndex
}

// Error returns the last error, if any.
func (m *Model) Error() error {
	return m.err
}
