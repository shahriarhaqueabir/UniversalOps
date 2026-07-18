package app

import (
	"fmt"
	"time"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// Logs exposes log management bindings to the frontend.
type Logs struct {
	app *App
}

// NewLogs creates a new Logs facade.
func NewLogs(app *App) *Logs {
	return &Logs{app: app}
}

// GetLogs returns filtered log entries from the opsforall database.
func (l *Logs) GetLogs(level string, since string, n int) []LogEntry {
	if n <= 0 {
		n = 200
	}

	storage := common.GetStorage()
	if storage == nil {
		return []LogEntry{}
	}

	data, err := storage.QueryLogs(level, "", n)
	if err != nil {
		common.LogWarn("QueryLogs failed: %v", err)
		return []LogEntry{}
	}

	out := make([]LogEntry, 0, len(data))
	for _, d := range data {
		out = append(out, LogEntry{
			Timestamp: d.Timestamp,
			Level:     d.Level,
			Module:    d.Module,
			Message:   d.Message,
			Line:      "",
		})
	}
	return out
}

// GetLogStats returns aggregated log statistics for the Overview tab.
func (l *Logs) GetLogStats() LogStats {
	storage := common.GetStorage()
	if storage == nil {
		return LogStats{TopSources: []LogSourceCount{}, TrendingErrors: []TrendingError{}}
	}

	stats := LogStats{}
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	hourStart := now.Add(-time.Hour)
	minStart := now.Add(-time.Minute)

	// Time-window counts
	stats.TotalToday = storage.CountLogsAfter(todayStart)
	stats.TotalThisHour = storage.CountLogsAfter(hourStart)
	stats.TotalLastMin = storage.CountLogsAfter(minStart)

	// Level counts
	stats.ErrorCount = storage.CountLogsByLevel("ERROR")
	stats.WarningCount = storage.CountLogsByLevel("WARN")
	stats.InfoCount = storage.CountLogsByLevel("INFO")
	stats.DebugCount = storage.CountLogsByLevel("DEBUG")

	// Top 5 sources
	sources, err := storage.TopLogSources(5)
	if err == nil {
		stats.TopSources = make([]LogSourceCount, len(sources))
		for i, s := range sources {
			stats.TopSources[i] = LogSourceCount{Source: s.Source, Count: s.Count}
		}
	} else {
		stats.TopSources = []LogSourceCount{}
	}

	// Top 5 trending errors
	errs, err := storage.TrendingLogErrors(5)
	if err == nil {
		stats.TrendingErrors = make([]TrendingError, len(errs))
		for i, e := range errs {
			stats.TrendingErrors[i] = TrendingError{
				Message:  e.Message,
				Count:    e.Count,
				LastSeen: e.LastSeen.Format("2006/01/02 15:04:05"),
			}
		}
	} else {
		stats.TrendingErrors = []TrendingError{}
	}

	return stats
}

// GetLogTimeline returns time-bucketed log counts for charting.
// hours >= 24 buckets by hour; hours < 24 buckets by 5 minutes.
func (l *Logs) GetLogTimeline(hours int) []LogTimelinePoint {
	if hours <= 0 {
		hours = 24
	}

	storage := common.GetStorage()
	if storage == nil {
		return []LogTimelinePoint{}
	}

	buckets, err := storage.QueryLogTimeline(hours)
	if err != nil {
		common.LogWarn("QueryLogTimeline failed: %v", err)
		return []LogTimelinePoint{}
	}

	out := make([]LogTimelinePoint, len(buckets))
	for i, b := range buckets {
		out[i] = LogTimelinePoint{
			Timestamp: b.Bucket,
			Total:     b.Total,
			Errors:    b.Errors,
			Warnings:  b.Warnings,
			Info:      b.Info,
		}
	}
	return out
}

// GenerateLogSummary returns a deterministic summary of recent log activity.
func (l *Logs) GenerateLogSummary() LogSummary {
	storage := common.GetStorage()
	if storage == nil {
		return LogSummary{}
	}

	summary := LogSummary{}

	// Top source
	sources, err := storage.TopLogSources(1)
	if err == nil && len(sources) > 0 {
		summary.TopSource = sources[0].Source
	}

	// Most common error message
	errList, err := storage.TrendingLogErrors(1)
	if err == nil && len(errList) > 0 {
		summary.TopMessage = errList[0].Message
	}

	// Error trend: compare last hour vs previous hour
	now := time.Now()
	lastHour := now.Add(-time.Hour)
	prevHour := now.Add(-2 * time.Hour)

	errorsLastHour := storage.CountLogsByLevelInRange("ERROR", lastHour, now)
	errorsPrevHour := storage.CountLogsByLevelInRange("ERROR", prevHour, lastHour)

	switch {
	case errorsLastHour > errorsPrevHour*12/10:
		summary.ErrorTrend = "increasing"
	case errorsLastHour < errorsPrevHour*8/10:
		summary.ErrorTrend = "decreasing"
	default:
		summary.ErrorTrend = "stable"
	}

	if summary.TopSource != "" {
		summary.SummaryText = fmt.Sprintf("Most errors originate from %s. Error rate is %s.", summary.TopSource, summary.ErrorTrend)
	} else {
		summary.SummaryText = "No significant log activity detected."
	}

	return summary
}
