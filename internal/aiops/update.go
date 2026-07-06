package aiops

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

// Update handles all messages for the AIOps layer.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case OllamaStatusMsg:
		m.ollamaStatus = msg.Status

	case ChatResultMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.messages = append(m.messages, ChatMessage{Role: "assistant", Content: msg.Response})
			m.err = nil
		}

	case ReportSavedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.output = fmt.Sprintf("Error saving report: %v", msg.Err)
		} else {
			m.output = fmt.Sprintf("Report saved to: %s", msg.Path)
			m.err = nil
		}

	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}

	return nil
}

// handleKeyPress processes keyboard events for the AIOps view.
func (m *Model) handleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	// Report action keys (only when input is empty)
	if m.input == "" && m.tabIndex == 1 {
		switch msg.String() {
		case "R":
			m.output = GenerateReport(m.reports)
			return nil
		case "g":
			m.output = GenerateReport(m.reports)
			return nil
		case "m":
			m.output = ReportToMarkdown(m.reports)
			return nil
		case "j":
			return m.renderJSONReport()
		case "v":
			return m.renderCSVReport()
		case "s":
			path := fmt.Sprintf("report_%d.txt", time.Now().Unix())
			content := m.generateTextReport()
			if content == "" {
				m.err = fmt.Errorf("no report content to save")
				return nil
			}
			return m.saveReport(content, path)
		case "c":
			m.reports = nil
			m.output = ""
			m.err = nil
			return nil
		}
	}

	// Chat action keys (only when input is empty)
	if m.input == "" && m.tabIndex == 0 {
		switch msg.String() {
		case "c":
			m.messages = nil
			m.output = ""
			m.err = nil
			return nil
		}
	}

	// Regular input handling
	switch {
	case msg.Text != "":
		m.input += msg.Text
		return nil
	default:
		switch msg.String() {
		case "tab", "l", "right":
			m.tabIndex = (m.tabIndex + 1) % 2
		case "shift+tab", "h", "left":
			m.tabIndex = (m.tabIndex - 1 + 2) % 2
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

// submitInput processes the current input based on the active tab.
func (m *Model) submitInput() tea.Cmd {
	input := strings.TrimSpace(m.input)
	if input == "" {
		return nil
	}
	m.input = ""

	switch m.tabIndex {
	case 0: // Chat
		return m.handleChatSubmit(input)

	case 1: // Reports
		return m.handleReportSubmit(input)
	}

	return nil
}

// handleChatSubmit handles chat message submission.
func (m *Model) handleChatSubmit(input string) tea.Cmd {
	// Check for local system query first
	if isLocalSystemQuery(input) {
		cleanInput := stripLocalCommand(input)
		m.messages = append(m.messages, ChatMessage{Role: "user", Content: input})
		answer := m.localSystemAnswer(cleanInput)
		m.messages = append(m.messages, ChatMessage{Role: "assistant", Content: answer})
		return nil
	}

	if !m.ollamaAvailable() {
		m.err = fmt.Errorf("Ollama is not available. Start it with 'ollama serve'")
		return nil
	}

	// Add user message
	m.messages = append(m.messages, ChatMessage{Role: "user", Content: input})

	// Build API messages with system prompt (not stored in history)
	apiMsgs := []ChatMessage{
		{Role: "system", Content: "You are a helpful AI assistant in a terminal operations dashboard."},
	}
	apiMsgs = append(apiMsgs, m.messages...)

	return func() tea.Msg {
		response, err := Chat(apiMsgs)
		if err != nil {
			return ChatResultMsg{Err: err}
		}
		return ChatResultMsg{Response: response}
	}
}

// handleReportSubmit handles report section input.
func (m *Model) handleReportSubmit(input string) tea.Cmd {
	switch {
	case input == "generate":
		m.output = GenerateReport(m.reports)
	case input == "markdown":
		m.output = ReportToMarkdown(m.reports)
	case input == "json":
		return m.renderJSONReport()
	case input == "csv":
		return m.renderCSVReport()
	case strings.HasPrefix(input, "save-json"):
		path := reportPathFromCommand(input, "save-json", "report.json")
		content, err := ReportToJSON(m.reports)
		if err != nil {
			m.err = err
			return nil
		}
		if content == "" || content == "null\n" {
			m.err = fmt.Errorf("no report content to save")
			return nil
		}
		return m.saveReport(content, path)
	case strings.HasPrefix(input, "save-csv"):
		path := reportPathFromCommand(input, "save-csv", "report.csv")
		content, err := ReportToCSV(m.reports)
		if err != nil {
			m.err = err
			return nil
		}
		if len(m.reports) == 0 {
			m.err = fmt.Errorf("no report content to save")
			return nil
		}
		return m.saveReport(content, path)
	case strings.HasPrefix(input, "save"):
		path := "report.txt"
		if len(input) > 5 && input[4] == ':' {
			path = strings.TrimSpace(input[5:])
		}
		if path == "" || path == "save" {
			path = fmt.Sprintf("report_%d.txt", time.Now().Unix())
		}
		content := m.generateTextReport()
		if content == "" {
			m.err = fmt.Errorf("no report content to save")
			return nil
		}
		return m.saveReport(content, path)
	case input == "clear":
		m.reports = nil
		m.output = ""
		m.err = nil
	default:
		// Add a section from input (format: "Title|Content")
		m.reports = addSection(m.reports, input)
		m.output = ""
		m.err = nil
	}
	return nil
}

func (m *Model) renderJSONReport() tea.Cmd {
	content, err := ReportToJSON(m.reports)
	if err != nil {
		m.err = err
		return nil
	}
	m.output = content
	m.err = nil
	return nil
}

func (m *Model) renderCSVReport() tea.Cmd {
	content, err := ReportToCSV(m.reports)
	if err != nil {
		m.err = err
		return nil
	}
	m.output = content
	m.err = nil
	return nil
}

// generateTextReport generates a plain text report from all sections.
func (m *Model) generateTextReport() string {
	return GenerateReport(m.reports)
}

// saveReport saves the current report to a file.
func (m *Model) saveReport(content, path string) tea.Cmd {
	return func() tea.Msg {
		err := SaveReport(content, path)
		if err != nil {
			return ReportSavedMsg{Err: err}
		}
		return ReportSavedMsg{Path: path}
	}
}

func reportPathFromCommand(input, command, fallback string) string {
	path := fallback
	if len(input) > len(command)+1 && input[len(command)] == ':' {
		path = strings.TrimSpace(input[len(command)+1:])
	}
	if path == "" || path == command {
		return fallback
	}
	return path
}

// ollamaAvailable returns true if Ollama is connected.
func (m *Model) ollamaAvailable() bool {
	return m.ollamaStatus != nil && m.ollamaStatus.Available
}
