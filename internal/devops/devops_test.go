package devops

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name string
		size int64
		want string
	}{
		{"zero bytes", 0, "0 B"},
		{"one byte", 1, "1 B"},
		{"512 bytes", 512, "512 B"},
		{"1 KB", 1024, "1.0 KB"},
		{"1.5 KB", 1536, "1.5 KB"},
		{"1 MB", 1048576, "1.0 MB"},
		{"2.5 MB", 2621440, "2.5 MB"},
		{"1 GB", 1073741824, "1.0 GB"},
		{"1.5 GB", 1610612736, "1.5 GB"},
		{"1 TB", 1099511627776, "1.0 TB"},
		{"negative", -100, "-100 B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSize(tt.size)
			if got != tt.want {
				t.Errorf("formatSize(%d) = %q, want %q", tt.size, got, tt.want)
			}
		})
	}
}

func TestParsePID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		prefix  string
		want    int32
		wantErr bool
	}{
		{name: "kill pid", input: "kill:1234", prefix: "kill:", want: 1234},
		{name: "restart pid with spaces", input: "restart: 42", prefix: "restart:", want: 42},
		{name: "missing pid", input: "kill:", prefix: "kill:", wantErr: true},
		{name: "bad pid", input: "kill:abc", prefix: "kill:", wantErr: true},
		{name: "negative pid", input: "kill:-1", prefix: "kill:", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePID(tt.input, tt.prefix)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parsePID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("parsePID() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParseWindowsServicesJSON(t *testing.T) {
	t.Run("array response", func(t *testing.T) {
		input := `[
			{"Name":"WinDefend","DisplayName":"Microsoft Defender Antivirus Service","Status":"Running","StartType":"Automatic"},
			{"Name":"Spooler","DisplayName":"Print Spooler","Status":{"Value":"Stopped"},"StartType":"Manual"}
		]`

		got, err := parseWindowsServicesJSON(input)
		if err != nil {
			t.Fatalf("parseWindowsServicesJSON() error = %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len(services) = %d, want 2", len(got))
		}
		if got[0].Name != "WinDefend" || got[0].Status != "Running" {
			t.Errorf("first service = %+v, want WinDefend Running", got[0])
		}
		if got[1].Status != "Stopped" {
			t.Errorf("second service status = %q, want Stopped", got[1].Status)
		}
	})

	t.Run("single object response", func(t *testing.T) {
		input := `{"Name":"EventLog","DisplayName":"Windows Event Log","Status":4,"StartType":"Automatic"}`

		got, err := parseWindowsServicesJSON(input)
		if err != nil {
			t.Fatalf("parseWindowsServicesJSON() error = %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("len(services) = %d, want 1", len(got))
		}
		if got[0].Status != "Running" {
			t.Errorf("status = %q, want Running", got[0].Status)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		if _, err := parseWindowsServicesJSON("{bad}"); err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestParseSystemctlServices(t *testing.T) {
	output := `ssh.service loaded active running OpenBSD Secure Shell server
cron.service loaded active running Regular background program processing daemon
badline
`
	got := parseSystemctlServices(output)
	if len(got) != 2 {
		t.Fatalf("len(services) = %d, want 2", len(got))
	}
	if got[0].Name != "ssh.service" || got[0].Status != "running" {
		t.Errorf("first service = %+v, want ssh.service running", got[0])
	}
	if !strings.Contains(got[1].DisplayName, "Regular background") {
		t.Errorf("second display name = %q, want description", got[1].DisplayName)
	}
}

func TestFormatProcesses(t *testing.T) {
	got := formatProcesses([]ProcessEntry{
		{PID: 12, Name: "example", CPU: 3.5, Memory: 42.25, Status: "running"},
	})
	if !strings.Contains(got, "Processes: 1 shown") {
		t.Errorf("formatProcesses() = %q, want count", got)
	}
	if !strings.Contains(got, "example") {
		t.Errorf("formatProcesses() = %q, want process name", got)
	}
}

func TestFormatServices(t *testing.T) {
	got := formatServices([]ServiceEntry{
		{Name: "svc-a", Status: "Running", StartType: "Automatic", DisplayName: "Service A"},
		{Name: "svc-b", Status: "Stopped", StartType: "Manual", DisplayName: "Service B"},
	})
	if !strings.Contains(got, "Running: 1") {
		t.Errorf("formatServices() = %q, want running count", got)
	}
	if !strings.Contains(got, "Stopped/Inactive: 1") {
		t.Errorf("formatServices() = %q, want stopped count", got)
	}
}

func TestModelProcessAndServiceTabs(t *testing.T) {
	m := NewModel()

	m.handleKeyPress(tea.KeyPressMsg{Text: "4"})
	if m.tabIndex != 3 {
		t.Fatalf("tabIndex = %d, want processes tab", m.tabIndex)
	}
	cmd := m.submitInput()
	if cmd == nil {
		t.Fatal("empty process submit should list processes")
	}

	m.handleKeyPress(tea.KeyPressMsg{Text: "5"})
	if m.tabIndex != 4 {
		t.Fatalf("tabIndex = %d, want services tab", m.tabIndex)
	}
	cmd = m.submitInput()
	if cmd == nil {
		t.Fatal("empty service submit should list services")
	}
}

func TestModelProcessMessages(t *testing.T) {
	m := NewModel()
	m.Update(ProcessListMsg{Processes: []ProcessEntry{{PID: 1, Name: "init"}}})
	if len(m.processes) != 1 {
		t.Fatalf("processes = %d, want 1", len(m.processes))
	}
	if m.output == "" {
		t.Error("output should be populated")
	}

	m.Update(ServiceListMsg{Services: []ServiceEntry{{Name: "svc", Status: "Running"}}})
	if len(m.services) != 1 {
		t.Fatalf("services = %d, want 1", len(m.services))
	}
	if m.err != nil {
		t.Errorf("err = %v, want nil", m.err)
	}
}

func TestModelUpdate_Messages(t *testing.T) {
	m := NewModel()

	// CommandResultMsg
	res := &ShellResult{Command: "ls", Output: "cmd output", Duration: time.Second}
	m.Update(CommandResultMsg{Result: res})
	if len(m.results) != 1 || m.output == "" {
		t.Error("CommandResultMsg not handled correctly")
	}

	// FileListMsg
	files := []FileEntry{{Name: "test.txt", Size: "100 B"}}
	m.Update(FileListMsg{Files: files, Path: "."})
	if len(m.files) != 1 || m.files[0].Name != "test.txt" {
		t.Error("FileListMsg not handled correctly")
	}

	// LogResultMsg
	m.Update(LogResultMsg{Lines: []string{"line 1", "line 2"}, Path: "app.log"})
	if len(m.logs) != 2 || m.logs[0] != "line 1" {
		t.Error("LogResultMsg not handled correctly")
	}

	// WorkflowResultMsg
	m.Update(WorkflowResultMsg{Report: "DevOps report"})
	if m.workflowReport != "DevOps report" || !m.showReport {
		t.Error("WorkflowResultMsg not handled correctly")
	}
}

func TestModelHandleKeyPress(t *testing.T) {
	m := NewModel()

	// Tab navigation
	m.handleKeyPress(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.tabIndex != 1 {
		t.Errorf("tabIndex = %d, want 1", m.tabIndex)
	}

	m.handleKeyPress(tea.KeyPressMsg{Code: tea.KeyRight})
	if m.tabIndex != 2 {
		t.Errorf("tabIndex = %d, want 2", m.tabIndex)
	}

	// Direct jump
	m.handleKeyPress(tea.KeyPressMsg{Text: "1"})
	if m.tabIndex != 0 {
		t.Errorf("tabIndex = %d, want 0", m.tabIndex)
	}

	// Input handling
	m.handleKeyPress(tea.KeyPressMsg{Text: "d"})
	m.handleKeyPress(tea.KeyPressMsg{Text: "i"})
	m.handleKeyPress(tea.KeyPressMsg{Text: "r"})
	if m.input != "dir" {
		t.Errorf("input = %q, want %q", m.input, "dir")
	}

	m.handleKeyPress(tea.KeyPressMsg{Code: tea.KeyBackspace})
	if m.input != "di" {
		t.Errorf("input = %q, want %q", m.input, "di")
	}
}

func TestModelView(t *testing.T) {
	m := NewModel()

	t.Run("shell-tab", func(t *testing.T) {
		m.tabIndex = 0
		m.output = "Welcome to shell"
		view := m.View(80, 24, nil)
		if !strings.Contains(view, "Development Operations") || !strings.Contains(view, "Welcome to shell") {
			t.Error("Shell tab view incorrect")
		}
	})

	t.Run("logtail-tab", func(t *testing.T) {
		m.tabIndex = 1
		m.logs = []string{"Error: disk full"}
		m.output = "Error: disk full"
		view := m.View(80, 24, nil)
		if !strings.Contains(view, "Log Viewer") || !strings.Contains(view, "disk full") {
			t.Error("Log viewer tab view incorrect")
		}
	})

	t.Run("filebrowser-tab", func(t *testing.T) {
		m.tabIndex = 2
		m.files = []FileEntry{{Name: "README.md", Size: "1024 B", IsDir: false}}
		m.output = "README.md"
		view := m.View(80, 24, nil)
		if !strings.Contains(view, "File Browser") || !strings.Contains(view, "README.md") {
			t.Error("File browser tab view incorrect")
		}
	})

	t.Run("report-view", func(t *testing.T) {
		m.showReport = true
		m.workflowReport = "DevOps health OK"
		view := m.View(80, 24, nil)
		if !strings.Contains(view, "Development Operations Report") || !strings.Contains(view, "DevOps health OK") {
			t.Error("Report view incorrect")
		}
	})
}
