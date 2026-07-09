package app

import (
	"fmt"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/aiops"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// AIOps exposes AI operations bindings to the frontend.
type AIOps struct {
	app *App
}

// NewAIOps creates a new AIOps facade.
func NewAIOps(app *App) *AIOps {
	return &AIOps{app: app}
}

// Chat sends a message to the Ollama chat API and returns the response.
func (a *AIOps) Chat(message string) string {
	storage := common.GetStorage()
	var historyContext string
	if storage != nil {
		// Fetch last 10 CPU and RAM metrics to give AI historical context
		// We use a strictly capped window to keep the prompt lightweight
		cpuHistory, _ := storage.GetMetricHistory(common.MetricCPU, 10)
		memHistory, _ := storage.GetMetricHistory(common.MetricMem, 10)
		if len(cpuHistory) > 0 {
			historyContext = fmt.Sprintf("\nHistorical Context (10s window):\nCPU usage patterns: %v\nRAM occupancy patterns: %v\n", cpuHistory, memHistory)
		}
	}


	messages := []aiops.ChatMessage{
		{Role: "system", Content: "You are the Hawkward AI Assistant. Use current stats and historical context provided to answer accurately." + historyContext},
		{Role: "user", Content: message},
	}
	response, err := aiops.Chat(messages)
	if err != nil {
		common.LogWarn("AI Chat failed: %v", err)
		return "Error: " + err.Error()
	}
	return response
}


// GenerateReport creates a formatted text report from the given sections.
// Sections are provided as a list of section title strings (content is auto-generated).
func (a *AIOps) GenerateReport(sections []string) string {
	var reportSections []aiops.ReportSection
	for _, title := range sections {
		// Build a prompt to generate content for this section using Ollama
		prompt := "Generate a brief operations report section for: " + title +
			". Include key metrics and observations based on recent system data."
		resp, err := aiops.Chat([]aiops.ChatMessage{
			{Role: "system", Content: "You are a system operations analyst. Be concise and factual."},
			{Role: "user", Content: prompt},
		})

		content := ""
		if err == nil {
			content = resp
		} else {
			content = "Content generation unavailable: " + err.Error()
		}
		reportSections = append(reportSections, aiops.ReportSection{
			Title:   title,
			Content: content,
		})
	}
	return aiops.GenerateReport(reportSections)
}

// GetOllamaStatus returns the current Ollama service status.
func (a *AIOps) GetOllamaStatus() OllamaStatus {
	status, err := aiops.CheckOllama()
	if err != nil {
		return OllamaStatus{
			Available: false,
			Error:     err.Error(),
		}
	}
	return OllamaStatus{
		Available:       status.Available,
		Model:           status.Model,
		Version:         status.Version,
		AvailableModels: status.AvailableModels,
	}
}

// DetectAnomalies performs anomaly detection on pipeline metrics.
// Compares current values against rolling window statistics.
func (a *AIOps) DetectAnomalies() []AnomalyInfo {
	var anomalies []AnomalyInfo

	metrics := []string{
		common.MetricCPU,
		common.MetricMem,
		common.MetricDisk,
		common.MetricNetRX,
		common.MetricNetTX,
		common.MetricProcCnt,
	}

	for _, name := range metrics {
		mf := a.app.pipeline.GetMetricWithForecast(name)
		if len(mf.Values) < 10 {
			continue // need enough data
		}

		lastVal := mf.LastValue
		mean := mf.Stats.Avg
		stddev := (mf.Stats.Max - mf.Stats.Min) / 2
		if stddev < 0.1 {
			stddev = 0.1
		}

		deviation := (lastVal - mean) / stddev
		if deviation < 0 {
			deviation = -deviation
		}

		if deviation > 3.0 {
			severity := "warning"
			if deviation > 5.0 {
				severity = "critical"
			}
			anomalies = append(anomalies, AnomalyInfo{
				Metric:    name,
				Value:     lastVal,
				Expected:  mean,
				Deviation: deviation,
				Severity:  severity,
				Timestamp: mf.LastTime.Format("2006-01-02T15:04:05Z07:00"),
			})
		}
	}

	return anomalies
}
