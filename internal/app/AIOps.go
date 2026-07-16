package app

import (
	"context"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/ollama/ollama/api"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/aiops"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// titleCaser replaces the deprecated strings.Title (removed guidance since
// Go 1.18); see CROSS-1.
var titleCaser = cases.Title(language.English)

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
		cpuHistory, _ := storage.GetMetricHistory(common.MetricCPU, 10)
		memHistory, _ := storage.GetMetricHistory(common.MetricMem, 10)
		if len(cpuHistory) > 0 {
			historyContext = fmt.Sprintf("\nHistorical Context (10s window):\nCPU usage patterns: %v\nRAM occupancy patterns: %v\n", cpuHistory, memHistory)
		}
	}

	messages := []aiops.ChatMessage{
		{Role: "system", Content: "You are the OpsForAll AI Assistant. Use current stats and historical context provided to answer accurately." + historyContext},
		{Role: "user", Content: message},
	}
	response, err := aiops.Chat(messages)
	if err != nil {
		common.LogWarn("AI Chat failed: %v", err)
		return "Error: " + err.Error()
	}
	return response
}

// sanitizePromptInput strips content that could facilitate prompt injection.
func sanitizePromptInput(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 200 {
		s = s[:200]
	}
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}

// GenerateReport creates a formatted text report from the given sections.
func (a *AIOps) GenerateReport(sections []string) string {
	var reportSections []aiops.ReportSection
	for _, title := range sections {
		title = sanitizePromptInput(title)
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
	binaryExists := aiops.CheckOllamaBinary()

	if err != nil {
		return OllamaStatus{
			Available:    false,
			BinaryExists: binaryExists,
			Error:        err.Error(),
		}
	}
	return OllamaStatus{
		Available:       status.Available,
		BinaryExists:    binaryExists,
		Model:           status.Model,
		Version:         status.Version,
		AvailableModels: status.AvailableModels,
	}
}

// PullModel initiates a model pull and emits progress events to the frontend.
func (a *AIOps) PullModel(modelName string) error {
	common.LogInfo("AIOps: Pulling model %q", modelName)
	err := aiops.PullModel(modelName, func(resp api.ProgressResponse) error {
		if a.app.ctx != nil {
			percent := 0.0
			if resp.Total > 0 {
				percent = float64(resp.Completed) / float64(resp.Total) * 100
			}
			runtime.EventsEmit(a.app.ctx, "ollama:progress", OllamaProgress{
				Status:    resp.Status,
				Percent:   percent,
				Total:     resp.Total,
				Completed: resp.Completed,
			})
		}
		return nil
	})

	if err != nil {
		common.LogWarn("AIOps: PullModel failed: %v", err)
		return err
	}
	return nil
}

// CreateOpsPersona creates the specialized 'opsforall' model from the local Modelfile.
func (a *AIOps) CreateOpsPersona() error {
	common.LogInfo("AIOps: Creating opsforall persona")

	// In a real build, we'd find the Modelfile relative to the executable or in assets.
	// For now, we assume it's in the current working directory as per project root.
	modelfilePath := "Modelfile"

	// If running via 'wails dev', it might be in the project root.
	if _, err := os.Stat(modelfilePath); os.IsNotExist(err) {
		// Try to find it in common locations or from embed if we had it there.
		// For this implementation, we'll try to find it in the project root.
		common.LogWarn("AIOps: Modelfile not found at %s", modelfilePath)
		return fmt.Errorf("Modelfile not found. Please ensure it exists in the application directory.")
	}

	// Read the base model name from the Modelfile 'FROM' line if we wanted to be dynamic,
	// but we'll use structured parameters for the SDK v0.31.2.

	// Default values matching our Modelfile:
	from := "hf.co/empero-ai/Qwythos-9B-Claude-Mythos-5-1M-GGUF:Q6_K"
	system := "You are the OpsForAll Technical Auditor, a professional system, network, and security intelligence agent.\nYour primary objective is to synthesize complex telemetry into high-density technical briefings.\nMuted industrial tone. Mechanics-first explanation. Zero fluff.\nAnchor all claims in metrics and specific protocol mechanisms."
	params := map[string]any{
		"temperature":    0.1,
		"top_p":          0.95,
		"repeat_penalty": 1.1,
		"stop":           []string{"──"},
	}

	err := aiops.CreateModel("opsforall", from, system, params)
	if err != nil {
		common.LogWarn("AIOps: CreateModel failed: %v", err)
		return err
	}

	common.LogInfo("AIOps: Successfully created opsforall persona")
	return nil
}

// DetectAnomalies performs anomaly detection on pipeline metrics.
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
			continue
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

// ── AI Methods for Timeline Integration ──────────────────────────────────────

// AskAI sends a prompt to the AI with the given context and returns the response.
func (a *AIOps) AskAI(ctx context.Context, prompt string) (string, error) {
	messages := []aiops.ChatMessage{
		{Role: "system", Content: "You are the OpsForAll AI assistant, an expert operations analyst. Be concise and specific."},
		{Role: "user", Content: prompt},
	}
	return aiops.Chat(messages)
}

// WithTimeout returns a context with the given timeout.
func (a *AIOps) WithTimeout(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}

// QuerySystemState answers a natural-language system-state question using live metrics.
func (a *AIOps) QuerySystemState(query string) string {
	stats, err := a.app.SysOps.collector.CollectAllStats()
	if err != nil {
		common.LogWarn("QuerySystemState: CollectAllStats failed: %v", err)
	}

	return aiops.AnswerSystemStateQuery(query, stats, nil, nil)
}

// GetAIInsights synthesizes anomaly, trend, and alert data into actionable insights.
func (a *AIOps) GetAIInsights() []AIInsight {
	var insights []AIInsight
	now := time.Now().Format("2006-01-02T15:04:05Z07:00")

	// 1. Pull anomalies from existing detection
	anomalies := a.DetectAnomalies()
	for _, anom := range anomalies {
		insights = append(insights, AIInsight{
			Category:  metricCategory(anom.Metric),
			Severity:  anom.Severity,
			Title:     fmt.Sprintf("%s anomaly detected", titleCaser.String(anom.Metric)),
			Message:   fmt.Sprintf("%s is at %.1f (expected ~%.1f, %.1fσ deviation)", anom.Metric, anom.Value, anom.Expected, anom.Deviation),
			Action:    fmt.Sprintf("Investigate %s usage for runaway processes or resource leaks.", anom.Metric),
			Timestamp: now,
		})
	}

	// 2. Check metric trends
	metricNames := []string{common.MetricCPU, common.MetricMem, common.MetricDisk}
	for _, name := range metricNames {
		mf := a.app.pipeline.GetMetricWithForecast(name)
		if len(mf.Values) < 10 {
			continue
		}
		if mf.Trend.Direction == common.TrendRising && mf.LastValue > 75 {
			insights = append(insights, AIInsight{
				Category:  metricCategory(name),
				Severity:  "warning",
				Title:     fmt.Sprintf("%s trending upward", strings.ToUpper(name)),
				Message:   fmt.Sprintf("%s is at %.1f%% and rising. May reach critical threshold.", name, mf.LastValue),
				Action:    "Monitor closely and consider preemptive action.",
				Timestamp: now,
			})
		} else if mf.Trend.Direction == common.TrendFalling && mf.LastValue > 90 {
			insights = append(insights, AIInsight{
				Category:  metricCategory(name),
				Severity:  "info",
				Title:     fmt.Sprintf("%s recovering", strings.ToUpper(name)),
				Message:   fmt.Sprintf("%s was high but is now trending down from %.1f%%.", name, mf.LastValue),
				Action:    "No action needed — conditions improving.",
				Timestamp: now,
			})
		}
	}

	// 3. Check recent alerts
	storage := common.GetStorage()
	if storage != nil {
		history, _ := storage.QueryAlertHistory(20)
		var activeCritical, activeWarning int
		for _, alert := range history {
			if alert.Resolved {
				continue
			}
			if alert.Level == "CRITICAL" {
				activeCritical++
			} else if alert.Level == "WARNING" {
				activeWarning++
			}
		}
		if activeCritical > 0 {
			insights = append(insights, AIInsight{
				Category:  "alerts",
				Severity:  "critical",
				Title:     fmt.Sprintf("%d active critical alert(s)", activeCritical),
				Message:   "There are unresolved critical alerts that require immediate attention.",
				Action:    "Review and resolve critical alerts in the Alerts dashboard.",
				Timestamp: now,
			})
		}
		if activeWarning > 0 {
			insights = append(insights, AIInsight{
				Category:  "alerts",
				Severity:  "warning",
				Title:     fmt.Sprintf("%d active warning alert(s)", activeWarning),
				Message:   "There are unresolved warnings that should be reviewed.",
				Action:    "Check warning-level alerts for emerging issues.",
				Timestamp: now,
			})
		}
	}

	if len(insights) == 0 {
		insights = append(insights, AIInsight{
			Category:  "general",
			Severity:  "info",
			Title:     "System operating normally",
			Message:   "No anomalies, concerning trends, or active alerts detected.",
			Action:    "Continue monitoring.",
			Timestamp: now,
		})
	}

	// Sort by severity: critical > warning > info
	severityRank := map[string]int{"critical": 0, "warning": 1, "info": 2}
	sort.Slice(insights, func(i, j int) bool {
		return severityRank[insights[i].Severity] < severityRank[insights[j].Severity]
	})

	return insights
}

// GetConfidenceScore computes an overall system confidence score (0-100).
func (a *AIOps) GetConfidenceScore() AIConfidence {
	now := time.Now().Format("2006-01-02T15:04:05Z07:00")
	factors := make(map[string]float64)

	// Factor 1: Data freshness (30%) — how recent is the latest metric?
	storage := common.GetStorage()
	freshnessScore := 50.0 // default if no storage
	if storage != nil {
		recent, _ := storage.GetMetricHistory(common.MetricCPU, 1)
		if len(recent) > 0 {
			freshnessScore = 100.0 // we have recent data
		}
	}
	factors["data_freshness"] = freshnessScore

	// Factor 2: Metric stability (25%) — low variance across metrics
	stabilityScore := 100.0
	metricNames := []string{common.MetricCPU, common.MetricMem, common.MetricDisk}
	for _, name := range metricNames {
		mf := a.app.pipeline.GetMetricWithForecast(name)
		if len(mf.Values) < 5 {
			continue
		}
		stddev := (mf.Stats.Max - mf.Stats.Min) / 2
		if stddev > 20 {
			stabilityScore -= 15
		} else if stddev > 10 {
			stabilityScore -= 5
		}
		// Penalize if value is near critical threshold
		if mf.LastValue > 90 {
			stabilityScore -= 10
		} else if mf.LastValue > 80 {
			stabilityScore -= 3
		}
	}
	if stabilityScore < 0 {
		stabilityScore = 0
	}
	factors["metric_stability"] = stabilityScore

	// Factor 3: Anomaly count (25%)
	anomalies := a.DetectAnomalies()
	anomalyScore := 100.0
	for _, anom := range anomalies {
		if anom.Severity == "critical" {
			anomalyScore -= 25
		} else {
			anomalyScore -= 10
		}
	}
	if anomalyScore < 0 {
		anomalyScore = 0
	}
	factors["anomaly_count"] = anomalyScore

	// Factor 4: Alert history (20%)
	alertScore := 100.0
	if storage != nil {
		history, _ := storage.QueryAlertHistory(20)
		for _, alert := range history {
			if alert.Resolved {
				continue
			}
			if alert.Level == "CRITICAL" {
				alertScore -= 15
			} else if alert.Level == "WARNING" {
				alertScore -= 5
			}
		}
	}
	if alertScore < 0 {
		alertScore = 0
	}
	factors["alert_health"] = alertScore

	overall := (freshnessScore*0.30 + stabilityScore*0.25 + anomalyScore*0.25 + alertScore*0.20)

	return AIConfidence{
		Overall:   math.Round(overall*10) / 10,
		Factors:   factors,
		UpdatedAt: now,
	}
}

// ── Conversation Persistence ───────────────────────────────────────────────

// LearnedBaseline represents the statistical baseline for a metric.
type LearnedBaseline struct {
	Metric string  `json:"metric"`
	Mean   float64 `json:"mean"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	StdDev float64 `json:"stdDev"`
	Count  int     `json:"count"` // number of samples
}

// GetLearnedBaselines returns the statistical baseline for each core metric.
func (a *AIOps) GetLearnedBaselines() []LearnedBaseline {
	metricNames := []string{common.MetricCPU, common.MetricMem, common.MetricDisk, common.MetricNetRX, common.MetricNetTX}
	var baselines []LearnedBaseline

	for _, name := range metricNames {
		mf := a.app.pipeline.GetMetricWithForecast(name)
		if len(mf.Values) < 5 {
			continue
		}
		stddev := (mf.Stats.Max - mf.Stats.Min) / 2
		baselines = append(baselines, LearnedBaseline{
			Metric: name,
			Mean:   math.Round(mf.Stats.Avg*10) / 10,
			Min:    math.Round(mf.Stats.Min*10) / 10,
			Max:    math.Round(mf.Stats.Max*10) / 10,
			StdDev: math.Round(stddev*10) / 10,
			Count:  len(mf.Values),
		})
	}

	if baselines == nil {
		baselines = []LearnedBaseline{}
	}
	return baselines
}

// SaveMessage persists a chat message. If sessionID is empty, generates a new one.
func (a *AIOps) SaveMessage(sessionID, role, content string) string {
	storage := common.GetStorage()
	if storage == nil {
		return sessionID
	}
	if sessionID == "" {
		sessionID = fmt.Sprintf("sess-%d", time.Now().UnixMilli())
	}
	if err := storage.InsertMessage(sessionID, role, content); err != nil {
		common.LogWarn("SaveMessage: %v", err)
	}
	return sessionID
}

// GetMessages returns all messages for a given session.
func (a *AIOps) GetMessages(sessionID string) []ConversationMessage {
	storage := common.GetStorage()
	if storage == nil {
		return []ConversationMessage{}
	}
	msgs, err := storage.GetMessages(sessionID)
	if err != nil {
		common.LogWarn("GetMessages: %v", err)
		return []ConversationMessage{}
	}
	result := make([]ConversationMessage, len(msgs))
	for i, m := range msgs {
		result[i] = ConversationMessage{
			ID:        m.ID,
			SessionID: m.SessionID,
			Role:      m.Role,
			Content:   m.Content,
			Timestamp: m.Timestamp,
		}
	}
	return result
}

// ListSessions returns all chat sessions with metadata.
func (a *AIOps) ListSessions() []map[string]interface{} {
	storage := common.GetStorage()
	if storage == nil {
		return []map[string]interface{}{}
	}
	sessions, err := storage.ListSessions()
	if err != nil {
		common.LogWarn("ListSessions: %v", err)
		return []map[string]interface{}{}
	}
	return sessions
}

// DeleteSession removes a chat session and all its messages.
func (a *AIOps) DeleteSession(sessionID string) {
	storage := common.GetStorage()
	if storage == nil {
		return
	}
	if err := storage.DeleteSession(sessionID); err != nil {
		common.LogWarn("DeleteSession: %v", err)
	}
}

// metricCategory maps a metric name to an insight category.
func metricCategory(name string) string {
	switch {
	case strings.Contains(name, "cpu"):
		return "performance"
	case strings.Contains(name, "mem"):
		return "performance"
	case strings.Contains(name, "disk"):
		return "storage"
	case strings.Contains(name, "net"):
		return "network"
	default:
		return "general"
	}
}
