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
			check.Message = fmt.Sprintf("CPU at %.1f%% exceeds critical threshold of 90%% — close high-CPU processes (Task Manager → sort by CPU)", cpuStats.Percent)
			score -= 30
		} else if cpuStats.Percent > 80 {
			check.Status = "warn"
			check.Message = fmt.Sprintf("CPU at %.1f%% approaching 90%% critical threshold — monitor for spikes", cpuStats.Percent)
			score -= 15
		} else {
			check.Status = "pass"
			check.Message = fmt.Sprintf("CPU at %.1f%% within healthy range (<80%% threshold)", cpuStats.Percent)
		}
		result.Checks = append(result.Checks, check)
	}

	// Memory check
	mem, err := GetMemoryStats()
	if err == nil {
		check := DiagnosticCheck{Name: "Memory Usage", Value: fmt.Sprintf("%.1f%%", mem.UsedPercent)}
		if mem.UsedPercent > 90 {
			check.Status = "fail"
			check.Message = fmt.Sprintf("Memory at %.1f%% exceeds critical threshold of 90%% — close unused applications or add more RAM", mem.UsedPercent)
			score -= 25
		} else if mem.UsedPercent > 80 {
			check.Status = "warn"
			check.Message = fmt.Sprintf("Memory at %.1f%% approaching 90%% critical threshold — review background processes", mem.UsedPercent)
			score -= 10
		} else {
			check.Status = "pass"
			check.Message = fmt.Sprintf("Memory at %.1f%% within healthy range (<80%% threshold)", mem.UsedPercent)
		}
		result.Checks = append(result.Checks, check)
	}

	// Disk check
	disk, err := GetDiskStats()
	if err == nil {
		for _, p := range disk.Usage {
			freeGB := float64(p.FreeBytes) / (1024 * 1024 * 1024)
			totalGB := float64(p.TotalBytes) / (1024 * 1024 * 1024)
			check := DiagnosticCheck{
				Name:  fmt.Sprintf("Disk %s", p.Mountpoint),
				Value: fmt.Sprintf("%.1f%%", p.UsedPercent),
			}
			if p.UsedPercent > 95 {
				check.Status = "fail"
				check.Message = fmt.Sprintf("Disk %s at %.1f%% — only %.0f GB free of %.0f GB total. Run disk cleanup or extend volume", p.Mountpoint, p.UsedPercent, freeGB, totalGB)
				score -= 20
			} else if p.UsedPercent > 85 {
				check.Status = "warn"
				check.Message = fmt.Sprintf("Disk %s at %.1f%% — %.0f GB free of %.0f GB. Free up space soon to avoid slowdowns", p.Mountpoint, p.UsedPercent, freeGB, totalGB)
				score -= 10
			} else {
				check.Status = "pass"
				check.Message = fmt.Sprintf("Disk %s at %.1f%% — %.0f GB free of %.0f GB (threshold <85%%)", p.Mountpoint, p.UsedPercent, freeGB, totalGB)
			}
			result.Checks = append(result.Checks, check)
		}
	}

	// Swap check
	if mem != nil && mem.SwapTotal > 0 {
		swapUsedGB := float64(mem.SwapUsed) / (1024 * 1024 * 1024)
		swapTotalGB := float64(mem.SwapTotal) / (1024 * 1024 * 1024)
		check := DiagnosticCheck{
			Name:  "Swap Usage",
			Value: fmt.Sprintf("%.1f%%", mem.SwapPercent),
		}
		if mem.SwapPercent > 80 {
			check.Status = "warn"
			check.Message = fmt.Sprintf("Swap at %.1f%% (%.1f GB / %.1f GB). High swap suggests physical RAM is insufficient", mem.SwapPercent, swapUsedGB, swapTotalGB)
			score -= 10
		} else {
			check.Status = "pass"
			check.Message = fmt.Sprintf("Swap at %.1f%% (%.1f GB / %.1f GB). Normal utilization", mem.SwapPercent, swapUsedGB, swapTotalGB)
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
			check.Message = fmt.Sprintf("CPU at %.1f°C exceeds critical threshold of 85°C — check cooling system and ventilation", ext.Temperature)
			score -= 20
		} else if ext.Temperature > 70 {
			check.Status = "warn"
			check.Message = fmt.Sprintf("CPU at %.1f°C approaching critical threshold of 85°C — ensure adequate airflow", ext.Temperature)
			score -= 5
		} else {
			check.Status = "pass"
			check.Message = fmt.Sprintf("CPU at %.1f°C within normal range (<70°C threshold)", ext.Temperature)
		}
		result.Checks = append(result.Checks, check)
	}

	if score < 0 {
		score = 0
	}
	result.Score = score

	return result, nil
}
