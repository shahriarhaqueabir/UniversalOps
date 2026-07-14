package sysops

import (
	"fmt"
	"time"
)

// DiagnosticCheck holds a single diagnostic check result.
type DiagnosticCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "pass", "warn", "fail"
	Message string `json:"message"`
	Value   string `json:"value"`
}

// DiagnosticResult holds the full diagnostic result.
type DiagnosticResult struct {
	Checks    []DiagnosticCheck `json:"checks"`
	Score     int               `json:"score"`
	Timestamp string            `json:"timestamp"`
}

// RunExtendedDiagnostics runs a set of system health checks and returns a score from 0-100.
func RunExtendedDiagnostics() (*DiagnosticResult, error) {
	result := &DiagnosticResult{
		Timestamp: time.Now().Format(time.RFC3339),
	}
	score := 100

	// CPU check
	cpuStats, err := GetCPUStats()
	if err == nil {
		check := DiagnosticCheck{Name: "CPU Usage", Value: fmt.Sprintf("%.1f%%", cpuStats.Percent)}
		if cpuStats.Percent > 90 {
			check.Status = "fail"
			check.Message = "CPU usage critical"
			score -= 30
		} else if cpuStats.Percent > 80 {
			check.Status = "warn"
			check.Message = "CPU usage elevated"
			score -= 15
		} else {
			check.Status = "pass"
			check.Message = "CPU usage normal"
		}
		result.Checks = append(result.Checks, check)
	}

	// Memory check
	mem, err := GetMemoryStats()
	if err == nil {
		check := DiagnosticCheck{Name: "Memory Usage", Value: fmt.Sprintf("%.1f%%", mem.UsedPercent)}
		if mem.UsedPercent > 90 {
			check.Status = "fail"
			check.Message = "Memory usage critical"
			score -= 25
		} else if mem.UsedPercent > 80 {
			check.Status = "warn"
			check.Message = "Memory usage elevated"
			score -= 10
		} else {
			check.Status = "pass"
			check.Message = "Memory usage normal"
		}
		result.Checks = append(result.Checks, check)
	}

	// Disk check
	disk, err := GetDiskStats()
	if err == nil {
		for _, p := range disk.Usage {
			check := DiagnosticCheck{
				Name:  fmt.Sprintf("Disk %s", p.Mountpoint),
				Value: fmt.Sprintf("%.1f%%", p.UsedPercent),
			}
			if p.UsedPercent > 95 {
				check.Status = "fail"
				check.Message = "Disk space critical"
				score -= 20
			} else if p.UsedPercent > 85 {
				check.Status = "warn"
				check.Message = "Disk space low"
				score -= 10
			} else {
				check.Status = "pass"
				check.Message = "Disk space OK"
			}
			result.Checks = append(result.Checks, check)
		}
	}

	// Swap check
	if mem != nil && mem.SwapTotal > 0 {
		check := DiagnosticCheck{
			Name:  "Swap Usage",
			Value: fmt.Sprintf("%.1f%%", mem.SwapPercent),
		}
		if mem.SwapPercent > 80 {
			check.Status = "warn"
			check.Message = "Swap usage high"
			score -= 10
		} else {
			check.Status = "pass"
			check.Message = "Swap usage normal"
		}
		result.Checks = append(result.Checks, check)
	}

	// Temperature check
	ext, err := GetCPUExtended()
	if err == nil && ext.Temperature > 0 {
		check := DiagnosticCheck{
			Name:  "CPU Temperature",
			Value: fmt.Sprintf("%.1f°C", ext.Temperature),
		}
		if ext.Temperature > 85 {
			check.Status = "fail"
			check.Message = "Temperature critical"
			score -= 20
		} else if ext.Temperature > 70 {
			check.Status = "warn"
			check.Message = "Temperature elevated"
			score -= 5
		} else {
			check.Status = "pass"
			check.Message = "Temperature normal"
		}
		result.Checks = append(result.Checks, check)
	}

	if score < 0 {
		score = 0
	}
	result.Score = score

	return result, nil
}
