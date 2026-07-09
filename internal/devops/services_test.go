package devops

import (
	"runtime"
	"testing"
)

func TestStringifyServiceField(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{"plain string", "Running", "Running"},
		{"string with spaces", "  Manual  ", "Manual"},
		{"float number", float64(1), "Stopped"},
		{"float 4", float64(4), "Running"},
		{"unknown float", float64(7), "Status(7)"},
		{"Value object", map[string]interface{}{"Value": "Automatic"}, "Automatic"},
		{"Value object with number", map[string]interface{}{"Value": float64(1)}, "Stopped"},
		{"nil value", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stringifyServiceField(tt.input); got != tt.want {
				t.Errorf("stringifyServiceField(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseSCQuery(t *testing.T) {
	input := `
SERVICE_NAME: Spooler
DISPLAY_NAME: Print Spooler
TYPE               : 10
STATE              : 4 RUNNING
								  STOP_STATE       : 0

SERVICE_NAME: WSearch
DISPLAY_NAME: Windows Search
TYPE               : 20
STATE              : 1 STOPPED
								  STOP_STATE       : 0
`
	services := parseSCQuery(input)
	if len(services) != 2 {
		t.Fatalf("parseSCQuery returned %d services, want 2", len(services))
	}

	if services[0].Name != "Spooler" {
		t.Errorf("services[0].Name = %q, want %q", services[0].Name, "Spooler")
	}
	if services[0].Status != "Running" {
		t.Errorf("services[0].Status = %q, want %q", services[0].Status, "Running")
	}
	if services[0].DisplayName != "Print Spooler" {
		t.Errorf("services[0].DisplayName = %q, want %q", services[0].DisplayName, "Print Spooler")
	}

	if services[1].Name != "WSearch" {
		t.Errorf("services[1].Name = %q, want %q", services[1].Name, "WSearch")
	}
	if services[1].Status != "Stopped" {
		t.Errorf("services[1].Status = %q, want %q", services[1].Status, "Stopped")
	}
}

func TestParseSCQueryEmpty(t *testing.T) {
	services := parseSCQuery("")
	if len(services) != 0 {
		t.Errorf("parseSCQuery empty returned %d services, want 0", len(services))
	}
}

func TestParseWindowsServicesJSON(t *testing.T) {
	jsonInput := `[
	{
		"Name": "Spooler",
		"DisplayName": "Print Spooler",
		"Status": "Running",
		"StartType": "Automatic"
	},
	{
		"Name": "WSearch",
		"DisplayName": "Windows Search",
		"Status": "Stopped",
		"StartType": "Manual"
	}
]`
	services, err := parseWindowsServicesJSON(jsonInput)
	if err != nil {
		t.Fatalf("parseWindowsServicesJSON failed: %v", err)
	}
	if len(services) != 2 {
		t.Fatalf("parseWindowsServicesJSON returned %d services, want 2", len(services))
	}
	if services[0].Name != "Spooler" || services[0].Status != "Running" {
		t.Errorf("unexpected service[0]: %+v", services[0])
	}
	if services[1].Name != "WSearch" || services[1].StartType != "Manual" {
		t.Errorf("unexpected service[1]: %+v", services[1])
	}
}

func TestParseWindowsServicesJSON_WithDashes(t *testing.T) {
	// PowerShell sometimes emits "-" for unset fields which is invalid JSON
	jsonInput := `[
	{
		"Name": "TestSvc",
		"DisplayName": -,
		"Status": "Running",
		"StartType": -
	}
]`
	services, err := parseWindowsServicesJSON(jsonInput)
	if err != nil {
		t.Fatalf("parseWindowsServicesJSON with dashes failed: %v", err)
	}
	if len(services) != 1 {
		t.Fatalf("parseWindowsServicesJSON with dashes returned %d services, want 1", len(services))
	}
	if services[0].Name != "TestSvc" {
		t.Errorf("service.Name = %q, want %q", services[0].Name, "TestSvc")
	}
}

func TestParseWindowsServicesJSON_WithValueObjects(t *testing.T) {
	jsonInput := `[
	{
		"Name": "Spooler",
		"DisplayName": "Print Spooler",
		"Status": {"Value": "Running"},
		"StartType": {"Value": "Automatic"}
	}
]`
	services, err := parseWindowsServicesJSON(jsonInput)
	if err != nil {
		t.Fatalf("parseWindowsServicesJSON with Value objects failed: %v", err)
	}
	if len(services) != 1 {
		t.Fatalf("parseWindowsServicesJSON returned %d services, want 1", len(services))
	}
	if services[0].Status != "Running" {
		t.Errorf("service.Status = %q, want %q", services[0].Status, "Running")
	}
	if services[0].StartType != "Automatic" {
		t.Errorf("service.StartType = %q, want %q", services[0].StartType, "Automatic")
	}
}

func TestParseWindowsServicesJSON_Empty(t *testing.T) {
	_, err := parseWindowsServicesJSON("")
	if err == nil {
		t.Error("parseWindowsServicesJSON empty should return error")
	}
}

func TestParseWindowsServicesJSON_SingleObject(t *testing.T) {
	jsonInput := `{"Name": "SingleSvc", "DisplayName": "Single", "Status": "Running", "StartType": "Auto"}`
	services, err := parseWindowsServicesJSON(jsonInput)
	if err != nil {
		t.Fatalf("parseWindowsServicesJSON single object failed: %v", err)
	}
	if len(services) != 1 {
		t.Fatalf("parseWindowsServicesJSON returned %d services, want 1", len(services))
	}
	if services[0].Name != "SingleSvc" {
		t.Errorf("service.Name = %q, want %q", services[0].Name, "SingleSvc")
	}
}

func TestListServices(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Services test only for Windows")
	}

	services, err := ListServices(0)
	if err != nil {
		t.Fatalf("ListServices failed: %v", err)
	}

	// Windows always has services
	if len(services) == 0 {
		t.Error("No services found")
	}

	foundSpooler := false
	for _, s := range services {
		if s.Name == "Spooler" {
			foundSpooler = true
			break
		}
	}

	if !foundSpooler {
		t.Log("Spooler service not found (unusual for Windows)")
	}
}
