package aiops

import (
	"fmt"
	"math"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

const (
	warningThreshold      = 75.0
	criticalThreshold     = 90.0
	sustainedThreshold    = 80.0
	spikeThreshold        = 30.0
	processSpikeRatio     = 1.5
	processSpikeMinDelta  = 50
	maxMetricHistoryItems = 12
)

// Anomaly describes a deterministic metrics finding.
type Anomaly struct {
	Metric   string
	Severity string
	Message  string
}

// AnswerSystemStateQuery answers common system-state questions without a live LLM.
func AnswerSystemStateQuery(query string, stats *common.SystemStats, sections []ReportSection, anomalies []Anomaly) string {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		q = "status"
	}

	if stats == nil {
		if answer := answerFromReportSections(q, sections); answer != "" {
			return answer
		}
		return "I do not have live system metrics yet. Add report sections or refresh SysOps, then ask about CPU, memory, disk, processes, uptime, or anomalies."
	}

	switch {
	case strings.Contains(q, "security") || strings.Contains(q, "firewall") || strings.Contains(q, "user") || strings.Contains(q, "defender"):
		if answer := answerFromReportSections(q, sections); answer != "" {
			return answer
		}
		return "Security state seems stable based on available data. Check SecOps for firewall rules and user audits."
	case strings.Contains(q, "log") || strings.Contains(q, "service") || strings.Contains(q, "deploy"):
		if answer := answerFromReportSections(q, sections); answer != "" {
			return answer
		}
		return "No specific DevOps activity found in recent reports. Use DevOps layer to tail logs or check services."
	case containsAny(q, "anomaly", "anomalies", "issue", "problem", "risk", "health", "status"):
		return formatHealthAnswer(stats, anomalies)
	case strings.Contains(q, "cpu"):
		return fmt.Sprintf("CPU usage is %.1f%%. %s", stats.CPUPercent, usageAdvice("CPU", stats.CPUPercent))
	case strings.Contains(q, "mem") || strings.Contains(q, "ram"):
		return fmt.Sprintf("Memory usage is %.1f%% (%.1f GB of %.1f GB). %s",
			stats.MemoryUsed, stats.MemoryUsedGB, stats.MemoryTotalGB, usageAdvice("memory", stats.MemoryUsed))
	case strings.Contains(q, "disk") || strings.Contains(q, "storage") || strings.Contains(q, "free"):
		return fmt.Sprintf("Disk usage is %.1f%% with %s free. %s",
			stats.DiskUsed, common.FormatBytes(stats.DiskFree), usageAdvice("disk", stats.DiskUsed))
	case strings.Contains(q, "process"):
		var topProc string
		if answer := answerFromReportSections("process", sections); answer != "" {
			topProc = "\nRecent top process: " + answer
		}
		return fmt.Sprintf("There are %d running processes. A sudden jump can indicate runaway work or a service restart loop.%s", stats.ProcessCount, topProc)
	case strings.Contains(q, "network") || strings.Contains(q, "connection") || strings.Contains(q, "port"):
		if answer := answerFromReportSections(q, sections); answer != "" {
			return answer
		}
		return "Network stats look normal based on available data. Check NetOps for real-time connection lists."
	case strings.Contains(q, "uptime") || strings.Contains(q, "up time") || strings.Contains(q, "running"):
		return fmt.Sprintf("System uptime is %s.", valueOrUnknown(stats.Uptime))
	default:
		if answer := answerFromReportSections(q, sections); answer != "" {
			return answer
		}
		return formatSystemSummary(stats, anomalies)
	}
}

// DetectAnomalies evaluates current and recent system stats with deterministic rules.
func DetectAnomalies(history []common.SystemStats) []Anomaly {
	if len(history) == 0 {
		return nil
	}

	latest := history[len(history)-1]
	var findings []Anomaly

	findings = append(findings, thresholdAnomaly("CPU", latest.CPUPercent)...)
	findings = append(findings, thresholdAnomaly("memory", latest.MemoryUsed)...)
	findings = append(findings, thresholdAnomaly("disk", latest.DiskUsed)...)

	// Uptime anomaly (recent reboot)
	if strings.Contains(latest.Uptime, "second") || (strings.Contains(latest.Uptime, "minute") && !strings.Contains(latest.Uptime, "minutes")) {
		findings = append(findings, Anomaly{
			Metric:   "uptime",
			Severity: "info",
			Message:  "System was recently rebooted (uptime: " + latest.Uptime + ").",
		})
	}

	if len(history) >= 6 {
		recent6 := history[len(history)-6:]
		if allAtLeast(recent6, func(s common.SystemStats) float64 { return s.CPUPercent }, 95.0) {
			findings = append(findings, Anomaly{
				Metric:   "CPU",
				Severity: "critical",
				Message:  "CPU is pinned at 95%+ for the last 6 samples.",
			})
		}
	}

	if len(history) >= 3 {
		recent := history[len(history)-3:]
		if allAtLeast(recent, func(s common.SystemStats) float64 { return s.CPUPercent }, sustainedThreshold) {
			findings = append(findings, Anomaly{
				Metric:   "CPU",
				Severity: "warning",
				Message:  fmt.Sprintf("CPU has stayed at or above %.0f%% for the last 3 samples.", sustainedThreshold),
			})
		}
		if allAtLeast(recent, func(s common.SystemStats) float64 { return s.MemoryUsed }, sustainedThreshold) {
			findings = append(findings, Anomaly{
				Metric:   "memory",
				Severity: "warning",
				Message:  fmt.Sprintf("Memory has stayed at or above %.0f%% for the last 3 samples.", sustainedThreshold),
			})
		}
		if steadilyIncreasing(recent, func(s common.SystemStats) float64 { return s.MemoryUsed }, 15) && latest.MemoryUsed >= 70 {
			findings = append(findings, Anomaly{
				Metric:   "memory",
				Severity: "warning",
				Message:  "Memory usage is climbing across the last 3 samples.",
			})
		}
	}

	if len(history) >= 2 {
		prev := history[:len(history)-1]
		prevCPUAvg := average(prev, func(s common.SystemStats) float64 { return s.CPUPercent })
		if latest.CPUPercent >= 70 && latest.CPUPercent-prevCPUAvg >= spikeThreshold {
			findings = append(findings, Anomaly{
				Metric:   "CPU",
				Severity: "warning",
				Message:  fmt.Sprintf("CPU jumped %.1f points above the previous average.", latest.CPUPercent-prevCPUAvg),
			})
		}

		prevProcessAvg := average(prev, func(s common.SystemStats) float64 { return float64(s.ProcessCount) })
		processDelta := latest.ProcessCount - int(math.Round(prevProcessAvg))
		if prevProcessAvg > 0 && float64(latest.ProcessCount) >= prevProcessAvg*processSpikeRatio && processDelta >= processSpikeMinDelta {
			findings = append(findings, Anomaly{
				Metric:   "processes",
				Severity: "warning",
				Message:  fmt.Sprintf("Process count increased by about %d versus the previous average.", processDelta),
			})
		}

		prevDiskAvg := average(prev, func(s common.SystemStats) float64 { return s.DiskUsed })
		if latest.DiskUsed-prevDiskAvg >= 5.0 {
			findings = append(findings, Anomaly{
				Metric:   "disk",
				Severity: "warning",
				Message:  fmt.Sprintf("Disk usage jumped %.1f%% across the last few samples.", latest.DiskUsed-prevDiskAvg),
			})
		}
	}

	return findings
}

func thresholdAnomaly(metric string, value float64) []Anomaly {
	switch {
	case value >= criticalThreshold:
		return []Anomaly{{Metric: metric, Severity: "critical", Message: fmt.Sprintf("%s usage is critical at %.1f%%.", titleMetric(metric), value)}}
	case value >= warningThreshold:
		return []Anomaly{{Metric: metric, Severity: "warning", Message: fmt.Sprintf("%s usage is elevated at %.1f%%.", titleMetric(metric), value)}}
	default:
		return nil
	}
}

func formatHealthAnswer(stats *common.SystemStats, anomalies []Anomaly) string {
	if len(anomalies) == 0 {
		return fmt.Sprintf("No anomalies detected. CPU %.1f%%, memory %.1f%%, disk %.1f%%, processes %d.",
			stats.CPUPercent, stats.MemoryUsed, stats.DiskUsed, stats.ProcessCount)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Detected %d anomaly(s):", len(anomalies)))
	for _, finding := range anomalies {
		b.WriteString(fmt.Sprintf("\n- %s %s: %s", finding.Severity, finding.Metric, finding.Message))
	}
	return b.String()
}

func formatSystemSummary(stats *common.SystemStats, anomalies []Anomaly) string {
	summary := fmt.Sprintf("System state: CPU %.1f%%, memory %.1f%% (%.1f/%.1f GB), disk %.1f%% (%s free), processes %d, uptime %s.",
		stats.CPUPercent,
		stats.MemoryUsed,
		stats.MemoryUsedGB,
		stats.MemoryTotalGB,
		stats.DiskUsed,
		common.FormatBytes(stats.DiskFree),
		stats.ProcessCount,
		valueOrUnknown(stats.Uptime),
	)
	if len(anomalies) == 0 {
		return summary + " No anomalies detected."
	}
	return summary + fmt.Sprintf(" %d anomaly(s) detected; ask `anomalies` for details.", len(anomalies))
}

func answerFromReportSections(query string, sections []ReportSection) string {
	if len(sections) == 0 {
		return ""
	}

	for _, section := range sections {
		title := strings.ToLower(section.Title)
		if title != "" && strings.Contains(query, title) {
			return fmt.Sprintf("%s: %s", section.Title, section.Content)
		}
	}

	for _, keyword := range []string{"cpu", "memory", "disk", "process", "uptime", "network", "security", "firewall", "user", "log", "service", "defender"} {
		if strings.Contains(query, keyword) {
			for _, section := range sections {
				if strings.Contains(strings.ToLower(section.Title+" "+section.Content), keyword) {
					return fmt.Sprintf("%s: %s", section.Title, section.Content)
				}
			}
		}
	}
	return ""
}

func usageAdvice(metric string, value float64) string {
	switch {
	case value >= criticalThreshold:
		return fmt.Sprintf("%s is in a critical range; investigate top consumers soon.", titleMetric(metric))
	case value >= warningThreshold:
		return fmt.Sprintf("%s is elevated; keep watching for sustained pressure.", titleMetric(metric))
	default:
		return fmt.Sprintf("%s looks normal.", titleMetric(metric))
	}
}

func containsAny(s string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(s, needle) {
			return true
		}
	}
	return false
}

func allAtLeast(stats []common.SystemStats, value func(common.SystemStats) float64, threshold float64) bool {
	for _, stat := range stats {
		if value(stat) < threshold {
			return false
		}
	}
	return true
}

func steadilyIncreasing(stats []common.SystemStats, value func(common.SystemStats) float64, minDelta float64) bool {
	if len(stats) < 2 {
		return false
	}
	for i := 1; i < len(stats); i++ {
		if value(stats[i]) <= value(stats[i-1]) {
			return false
		}
	}
	return value(stats[len(stats)-1])-value(stats[0]) >= minDelta
}

func average(stats []common.SystemStats, value func(common.SystemStats) float64) float64 {
	if len(stats) == 0 {
		return 0
	}
	total := 0.0
	for _, stat := range stats {
		total += value(stat)
	}
	return total / float64(len(stats))
}

func titleMetric(metric string) string {
	if metric == "CPU" {
		return metric
	}
	if metric == "" {
		return "Metric"
	}
	return strings.ToUpper(metric[:1]) + metric[1:]
}

func valueOrUnknown(value string) string {
	if strings.TrimSpace(value) == "" {
		return "unknown"
	}
	return value
}
