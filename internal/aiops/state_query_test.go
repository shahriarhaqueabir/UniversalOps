package aiops

import (
	"strings"
	"testing"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

func TestDetectAnomalies(t *testing.T) {
	// Test case: Elevated CPU
	history := []common.SystemStats{
		{SystemCPUUtilization: 10.0},
		{SystemCPUUtilization: 15.0},
		{SystemCPUUtilization: 85.0}, // Elevated
	}
	anomalies := DetectAnomalies(history)
	found := false
	for _, a := range anomalies {
		if a.Metric == "CPU" && a.Severity == "warning" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected CPU anomaly, got none")
	}

	// Test case: CPU Spike
	history = []common.SystemStats{
		{SystemCPUUtilization: 10.0},
		{SystemCPUUtilization: 10.0},
		{SystemCPUUtilization: 75.0}, // Spike
	}
	anomalies = DetectAnomalies(history)
	found = false
	for _, a := range anomalies {
		if a.Metric == "CPU" && a.Severity == "warning" && strings.Contains(a.Message, "jumped") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected CPU spike anomaly, got none")
	}

	// Test case: Sustained Memory Pressure
	history = []common.SystemStats{
		{SystemMemoryUsage: 82.0},
		{SystemMemoryUsage: 85.0},
		{SystemMemoryUsage: 81.0},
	}
	anomalies = DetectAnomalies(history)
	found = false
	for _, a := range anomalies {
		if a.Metric == "memory" && strings.Contains(a.Message, "stayed at or above") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected sustained memory anomaly, got none")
	}

	// Test case: Recent Reboot
	history = []common.SystemStats{
		{SystemUptime: "45 seconds"},
	}
	anomalies = DetectAnomalies(history)
	found = false
	for _, a := range anomalies {
		if a.Metric == "uptime" && strings.Contains(a.Message, "recently rebooted") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected recent reboot anomaly, got none")
	}
}

func TestAnswerSystemStateQuery(t *testing.T) {
	stats := &common.SystemStats{
		SystemCPUUtilization: 45.0,
		SystemMemoryUsage:    60.0,
		SystemDiskUsage:      30.0,
	}

	answer := AnswerSystemStateQuery("cpu usage", stats, nil, nil)
	if !strings.Contains(strings.ToLower(answer), "45.0%") {
		t.Errorf("Expected answer to contain 45.0%%, got: %s", answer)
	}

	answer = AnswerSystemStateQuery("status", stats, nil, nil)
	if !strings.Contains(strings.ToLower(answer), "no anomalies") {
		t.Errorf("Expected answer to contain 'no anomalies', got: %s", answer)
	}

	// Test new security/devops patterns
	sections := []ReportSection{
		{Title: "Firewall", Content: "Allow all on port 80"},
		{Title: "Logs", Content: "Errors found in auth.log"},
	}

	answer = AnswerSystemStateQuery("firewall status", stats, sections, nil)
	if !strings.Contains(answer, "Allow all on port 80") {
		t.Errorf("Expected answer to contain firewall info, got: %s", answer)
	}

	answer = AnswerSystemStateQuery("what is in the logs?", stats, sections, nil)
	if !strings.Contains(answer, "Errors found in auth.log") {
		t.Errorf("Expected answer to contain log info, got: %s", answer)
	}
}

func TestDetectAnomalies_EmptyHistory(t *testing.T) {
	anomalies := DetectAnomalies(nil)
	if len(anomalies) != 0 {
		t.Errorf("Expected 0 anomalies for nil history, got %d", len(anomalies))
	}

	anomalies = DetectAnomalies([]common.SystemStats{})
	if len(anomalies) != 0 {
		t.Errorf("Expected 0 anomalies for empty history, got %d", len(anomalies))
	}
}

func TestDetectAnomalies_DiskPressure(t *testing.T) {
	history := []common.SystemStats{
		{SystemDiskUsage: 96.0},
	}
	anomalies := DetectAnomalies(history)
	found := false
	for _, a := range anomalies {
		if a.Metric == "disk" && a.Severity == "critical" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected critical disk anomaly for 96%% usage")
	}
}
