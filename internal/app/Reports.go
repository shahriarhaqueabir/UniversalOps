package app

import (
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// ReportsAPI exposes consolidated report retrieval to the frontend.
type ReportsAPI struct{}

// NewReportsAPI creates a new ReportsAPI facade.
func NewReportsAPI() *ReportsAPI {
	return &ReportsAPI{}
}

// ListAllReports returns all persisted reports (health, security, auto_diag) aggregated by recency.
func (r *ReportsAPI) ListAllReports() []common.ReportRecord {
	storage := common.GetStorage()
	if storage == nil {
		return []common.ReportRecord{}
	}
	reports, _ := storage.ListAllReports()
	return reports
}

// DeleteReport removes a single report by ID.
func (r *ReportsAPI) DeleteReport(id string) bool {
	storage := common.GetStorage()
	if storage == nil {
		return false
	}
	err := storage.DeleteReport(id)
	return err == nil
}
