package devops

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

// WorkflowResultMsg is sent when a workflow completes.
type WorkflowResultMsg struct {
	Report string
	Err    error
}

// Update handles all messages for the DevOps layer.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case WorkflowResultMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.showReport = true
			m.workflowReport = "Error running dev diagnostics: " + msg.Err.Error()
		} else {
			m.workflowReport = msg.Report
			m.showReport = true
		}

	case CommandResultMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.output = fmt.Sprintf("Error: %v", msg.Err)
		} else {
			m.results = append([]ShellResult{*msg.Result}, m.results...)
			m.output = formatShellResult(msg.Result)
			m.err = nil
		}

	case LogResultMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.output = fmt.Sprintf("Error: %v", msg.Err)
		} else {
			m.logs = msg.Lines
			m.output = formatLines(msg.Lines, msg.Path)
			m.err = nil
		}

	case FileListMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.output = fmt.Sprintf("Error: %v", msg.Err)
		} else {
			m.files = msg.Files
			m.output = formatFileEntries(msg.Files, msg.Path)
			m.err = nil
		}

	case FileContentMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.output = fmt.Sprintf("Error: %v", msg.Err)
		} else {
			m.output = fmt.Sprintf("File: %s\n\n%s", msg.Path, msg.Content)
			m.err = nil
		}

	case ProcessListMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.output = fmt.Sprintf("Error: %v", msg.Err)
		} else {
			m.processes = msg.Processes
			m.output = formatProcesses(msg.Processes)
			m.err = nil
		}

	case ProcessActionMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.output = fmt.Sprintf("Error: %v", msg.Err)
		} else {
			m.output = msg.Message
			m.err = nil
		}

	case ServiceListMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.output = fmt.Sprintf("Error: %v", msg.Err)
		} else {
			m.services = msg.Services
			m.output = formatServices(msg.Services)
			m.err = nil
		}

	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}

	return nil
}

// handleKeyPress processes keyboard events for the DevOps view.
func (m *Model) handleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	// Action keys when input is empty
	if m.input == "" {
		switch msg.String() {
		case "R":
			if m.showReport {
				m.showReport = false
				return nil
			}
			return func() tea.Msg {
				report, err := RunDevDiagnostics()
				if err != nil {
					return WorkflowResultMsg{Err: err}
				}
				return WorkflowResultMsg{Report: report.String()}
			}
		case "r":
			m.results = nil
			m.processes = nil
			m.services = nil
			m.output = ""
			m.err = nil
			m.showReport = false
			return nil
		case "1":
			m.tabIndex = 0
			return nil
		case "2":
			m.tabIndex = 1
			return nil
		case "3":
			m.tabIndex = 2
			return nil
		case "4":
			m.tabIndex = 3
			return nil
		case "5":
			m.tabIndex = 4
			return nil
		}
	}

	switch {
	case msg.Text != "":
		m.input += msg.Text
		return nil
	default:
		switch msg.String() {
		case "tab", "l", "right":
			m.tabIndex = (m.tabIndex + 1) % 5
		case "shift+tab", "h", "left":
			m.tabIndex = (m.tabIndex - 1 + 5) % 5
		case "enter":
			return m.submitInput()
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		case "esc":
			return nil
		}
	}
	return nil
}

// submitInput executes the current input based on the active tab.
func (m *Model) submitInput() tea.Cmd {
	input := strings.TrimSpace(m.input)
	if input == "" && m.tabIndex == 3 {
		input = "list"
	} else if input == "" && m.tabIndex == 4 {
		input = "list"
	} else if input == "" {
		return nil
	}
	m.input = ""

	switch m.tabIndex {
	case 0: // Shell
		return func() tea.Msg {
			result, err := RunCommand(input)
			if err != nil {
				return CommandResultMsg{Err: err}
			}
			return CommandResultMsg{Result: result}
		}

	case 1: // Logs
		return func() tea.Msg {
			if strings.HasPrefix(input, "search:") {
				rest := input[7:]
				parts := strings.SplitN(rest, ":", 2)
				if len(parts) == 2 {
					lines, err := SearchLog(parts[0], parts[1])
					if err != nil {
						return LogResultMsg{Err: err}
					}
					return LogResultMsg{Lines: lines, Path: parts[0]}
				}
				return LogResultMsg{
					Err: fmt.Errorf("usage: search:/path:pattern"),
				}
			}
			lines, err := TailLog(input, 50)
			if err != nil {
				return LogResultMsg{Err: err}
			}
			return LogResultMsg{Lines: lines, Path: input}
		}

	case 2: // Files
		return func() tea.Msg {
			if strings.HasPrefix(input, "read:") {
				path := input[5:]
				content, err := ReadFile(path)
				if err != nil {
					return FileContentMsg{Err: err}
				}
				return FileContentMsg{Content: content, Path: path}
			}
			files, err := ListDir(input)
			if err != nil {
				return FileListMsg{Err: err}
			}
			return FileListMsg{Files: files, Path: input}
		}

	case 3: // Processes
		return m.submitProcessInput(input)

	case 4: // Services
		return func() tea.Msg {
			services, err := ListServices(50)
			if err != nil {
				return ServiceListMsg{Err: err}
			}
			return ServiceListMsg{Services: services}
		}
	}
	return nil
}

func (m *Model) submitProcessInput(input string) tea.Cmd {
	switch {
	case input == "list":
		return listProcessesCmd()
	case strings.HasPrefix(input, "kill:"):
		pid, err := parsePID(input, "kill:")
		if err != nil {
			m.err = err
			return nil
		}
		return func() tea.Msg {
			if err := KillProcess(pid); err != nil {
				return ProcessActionMsg{Err: err}
			}
			return ProcessActionMsg{Message: fmt.Sprintf("Killed process %d", pid)}
		}
	case strings.HasPrefix(input, "restart:"):
		pid, err := parsePID(input, "restart:")
		if err != nil {
			m.err = err
			return nil
		}
		return func() tea.Msg {
			if err := RestartProcess(pid); err != nil {
				return ProcessActionMsg{Err: err}
			}
			return ProcessActionMsg{Message: fmt.Sprintf("Restarted process %d", pid)}
		}
	default:
		return listProcessesCmd()
	}
}

func listProcessesCmd() tea.Cmd {
	return func() tea.Msg {
		processes, err := ListProcesses(30)
		if err != nil {
			return ProcessListMsg{Err: err}
		}
		return ProcessListMsg{Processes: processes}
	}
}

// formatShellResult formats a ShellResult for display.
func formatShellResult(r *ShellResult) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("$ %s\n", r.Command))
	b.WriteString(fmt.Sprintf("⏱ %v  |  Exit: %d\n", r.Duration.Round(time.Millisecond), r.ExitCode))
	b.WriteString(strings.Repeat("─", 40) + "\n")
	if r.Output != "" {
		b.WriteString(r.Output)
		if !strings.HasSuffix(r.Output, "\n") {
			b.WriteString("\n")
		}
	}
	return b.String()
}
