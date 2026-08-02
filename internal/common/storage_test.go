package common

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Helper to create a Storage with temp DB for tests.
func newTestStorage(t *testing.T) *Storage {
	t.Helper()

	f, err := os.CreateTemp("", "universalops-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	path := f.Name()
	t.Cleanup(func() { os.Remove(path) })

	// Use InitStorage for a realistic setup
	if err := InitStorage(path); err != nil {
		t.Fatal(err)
	}

	s := GetStorage()
	t.Cleanup(func() {
		s.Close()
		globalStorageMu.Lock()
		globalStorage = nil
		globalStorageMu.Unlock()
	})
	return s
}

func TestInitStoragePragmas(t *testing.T) {
	f, err := os.CreateTemp("", "universalops-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	path := f.Name()
	t.Cleanup(func() { os.Remove(path) })

	if err := InitStorage(path); err != nil {
		t.Fatal(err)
	}
	s := GetStorage()
	defer s.Close()

	// Verify WAL mode was set
	var mode string
	if err := s.db.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatal(err)
	}
	if mode != "wal" {
		t.Errorf("journal_mode = %q, want %q", mode, "wal")
	}

	// Verify synchronous mode
	var syncMode int
	if err := s.db.QueryRow("PRAGMA synchronous").Scan(&syncMode); err != nil {
		t.Fatal(err)
	}
	if syncMode != 1 { // NORMAL = 1
		t.Errorf("synchronous = %d, want %d", syncMode, 1)
	}

	// Verify cache_size
	var cacheSize int
	if err := s.db.QueryRow("PRAGMA cache_size").Scan(&cacheSize); err != nil {
		t.Fatal(err)
	}
	// The PRAGMA cache_size=-8000 means 8000 * 1KB pages = ~8MB cache
	// The reported value may differ in sign across versions; check reasonable range
	if cacheSize > -1000 && cacheSize < 1000 {
		t.Errorf("cache_size = %d, expected value <= -8000 (or equivalent)", cacheSize)
	}
}

func TestInsertAndGetMetric(t *testing.T) {
	s := newTestStorage(t)

	if err := s.InsertMetric("cpu_usage", "%", 42.5); err != nil {
		t.Fatal(err)
	}
	if err := s.InsertMetric("cpu_usage", "%", 67.3); err != nil {
		t.Fatal(err)
	}
	if err := s.InsertMetric("memory_used", "MB", 2048); err != nil {
		t.Fatal(err)
	}

	s.flushMetrics()

	// Retrieve cpu_usage, expect 2 values in chronological order
	vals, err := s.GetMetricHistory("cpu_usage", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(vals) != 2 {
		t.Fatalf("got %d values, want 2: %v", len(vals), vals)
	}
	if vals[0] != 42.5 || vals[1] != 67.3 {
		t.Errorf("cpu_usage = %v, want [42.5, 67.3]", vals)
	}

	// Retrieve memory_used
	vals, err = s.GetMetricHistory("memory_used", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(vals) != 1 || vals[0] != 2048 {
		t.Errorf("memory_used = %v, want [2048]", vals)
	}

	// Limit test
	vals, err = s.GetMetricHistory("cpu_usage", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(vals) != 1 {
		t.Fatalf("limited query got %d values, want 1", len(vals))
	}
}

func TestConcurrentMetricWrites(t *testing.T) {
	s := newTestStorage(t)

	const writers = 10
	const metricsPerWriter = 20
	done := make(chan error, writers)

	for i := 0; i < writers; i++ {
		go func(id int) {
			for j := 0; j < metricsPerWriter; j++ {
				if err := s.InsertMetric("concurrent_test", "count", float64(id*1000+j)); err != nil {
					done <- err
					return
				}
			}
			done <- nil
		}(i)
	}

	for i := 0; i < writers; i++ {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}

	s.flushMetrics()

	vals, err := s.GetMetricHistory("concurrent_test", 1000)
	if err != nil {
		t.Fatal(err)
	}
	expectedTotal := writers * metricsPerWriter
	if len(vals) != expectedTotal {
		t.Errorf("got %d metrics, want %d", len(vals), expectedTotal)
	}

	// Verify all values are present (dedup check: no duplicates expected since IDs are unique)
	seen := make(map[float64]bool)
	for _, v := range vals {
		if seen[v] {
			t.Errorf("duplicate value %v found", v)
		}
		seen[v] = true
	}
	if len(seen) != expectedTotal {
		t.Errorf("got %d unique values, want %d", len(seen), expectedTotal)
	}
}

func TestInsertLogAndQuery(t *testing.T) {
	s := newTestStorage(t)

	// Use a unique module name for our test logs to avoid interference
	// from the async system log triggered by InitStorage.
	testModule := "storage_test_module"

	if err := s.InsertLog("INFO", testModule, "hello world"); err != nil {
		t.Fatal(err)
	}
	if err := s.InsertLog("ERROR", testModule, "connection refused"); err != nil {
		t.Fatal(err)
	}
	if err := s.InsertLog("INFO", testModule, "retry succeeded"); err != nil {
		t.Fatal(err)
	}

	// Flush to ensure async writes are committed
	s.flushMetrics()

	// Query with specific module to be deterministic
	entries, err := s.QueryLogs("", testModule, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("got %d log entries for module %q, want 3", len(entries), testModule)
	}

	// Filter by level and module
	entries, err = s.QueryLogs("ERROR", testModule, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("filtered by ERROR: got %d entries, want 1", len(entries))
	} else if entries[0].Message != "connection refused" {
		t.Errorf("got message %q, want 'connection refused'", entries[0].Message)
	}

	// Filter by search and module
	entries, err = s.QueryLogs("", "retry", 10)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range entries {
		if e.Module == testModule && e.Message == "retry succeeded" {
			found = true
			break
		}
	}
	if !found {
		t.Error("search 'retry': did not find expected entry")
	}
}

func TestQueryLogTimeline(t *testing.T) {
	s := newTestStorage(t)

	// Insert log entries with explicit timestamps over the last 24 hours
	now := time.Now().UTC()
	for i := 0; i < 10; i++ {
		ts := now.Add(-time.Duration(i) * 3 * time.Hour)
		s.insertLogsBatch([]LogWrite{{
			Level:   "INFO",
			Module:  "test",
			Message: "test info",
			Time:    ts,
		}})
		s.insertLogsBatch([]LogWrite{{
			Level:   "ERROR",
			Module:  "test",
			Message: "test error",
			Time:    ts,
		}})
		s.insertLogsBatch([]LogWrite{{
			Level:   "WARN",
			Module:  "test",
			Message: "test warn",
			Time:    ts,
		}})
	}

	// Flush to persist
	s.flushMetrics()

	// Query 24h timeline (hourly buckets)
	buckets, err := s.QueryLogTimeline(24)
	if err != nil {
		t.Fatalf("QueryLogTimeline(24) failed: %v", err)
	}

	if len(buckets) == 0 {
		t.Fatal("QueryLogTimeline(24) returned no buckets — expected at least 1")
	}

	// Each bucket should have 3 logs (INFO, ERROR, WARN) unless bucket spans
	// across the current hour boundary
	for _, b := range buckets {
		if b.Total == 0 {
			t.Errorf("bucket %q has total=0", b.Bucket)
		}
		if b.Bucket == "" {
			t.Error("found bucket with empty timestamp")
		}
	}

	// Query 1h timeline (5-minute buckets)
	buckets, err = s.QueryLogTimeline(1)
	if err != nil {
		t.Fatalf("QueryLogTimeline(1) failed: %v", err)
	}
	if len(buckets) == 0 {
		t.Fatal("QueryLogTimeline(1) returned no buckets — expected at least 1")
	}

	// Verify ordering (ascending)
	for i := 1; i < len(buckets); i++ {
		if buckets[i].Bucket < buckets[i-1].Bucket {
			t.Errorf("buckets not in ascending order: %q before %q", buckets[i-1].Bucket, buckets[i].Bucket)
		}
	}

	// Verify total counts
	var totalLogs int
	for _, b := range buckets {
		totalLogs += b.Total
	}
	if totalLogs == 0 {
		t.Error("total log count across buckets is 0")
	}
}

func TestPruneData(t *testing.T) {
	s := newTestStorage(t)

	// Insert a metric with a manually old timestamp by going around the channel
	if _, err := s.db.Exec(
		`INSERT INTO metrics (name, unit, value, timestamp) VALUES (?, ?, ?, ?)`,
		"old_metric", "count", 1.0, time.Now().Add(-10*24*time.Hour),
	); err != nil {
		t.Fatal(err)
	}

	// Insert a recent metric via the normal path
	if err := s.InsertMetric("recent_metric", "count", 2.0); err != nil {
		t.Fatal(err)
	}
	s.flushMetrics()

	// Verify both exist
	oldVals, err := s.GetMetricHistory("old_metric", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(oldVals) != 1 {
		t.Fatalf("expected 1 old metric before prune, got %d", len(oldVals))
	}

	// Prune data older than 7 days
	s.Prune(7 * 24 * time.Hour)

	// Old metric should be gone
	oldVals, err = s.GetMetricHistory("old_metric", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(oldVals) != 0 {
		t.Errorf("expected 0 old metrics after prune, got %d", len(oldVals))
	}

	// Recent metric should still exist
	recentVals, err := s.GetMetricHistory("recent_metric", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(recentVals) != 1 {
		t.Errorf("expected 1 recent metric after prune, got %d", len(recentVals))
	}
}

func TestStorageClose(t *testing.T) {
	f, err := os.CreateTemp("", "universalops-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	path := f.Name()
	t.Cleanup(func() { os.Remove(path) })

	if err := InitStorage(path); err != nil {
		t.Fatal(err)
	}

	s := GetStorage()

	// Insert a metric
	if err := s.InsertMetric("before_close", "unit", 99.9); err != nil {
		t.Fatal(err)
	}

	// Close to flush and shutdown
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	// Re-init to verify data persisted
	if err := InitStorage(path); err != nil {
		t.Fatal(err)
	}
	s2 := GetStorage()
	defer s2.Close()

	vals, err := s2.GetMetricHistory("before_close", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(vals) != 1 || vals[0] != 99.9 {
		t.Errorf("after close/reopen: got %v, want [99.9]", vals)
	}
}

func TestStorageDefaultPath(t *testing.T) {
	origDB := DefaultDBName
	t.Cleanup(func() {
		DefaultDBName = origDB
		os.Remove(origDB)
	})

	DefaultDBName = filepath.Join(t.TempDir(), "custom.db")
	if err := InitStorage(""); err != nil {
		t.Fatal(err)
	}
	s := GetStorage()
	defer s.Close()

	if err := s.InsertMetric("test", "unit", 1.0); err != nil {
		t.Fatal(err)
	}
	s.flushMetrics()

	vals, err := s.GetMetricHistory("test", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(vals) != 1 {
		t.Errorf("expected 1 metric with default path, got %d", len(vals))
	}
}

func TestEventsCRUD(t *testing.T) {
	s := newTestStorage(t)

	// Insert events
	evt1 := TimelineEvent{
		ID: NewUUID(), Timestamp: time.Now(), Category: CatSystem,
		Level: EventInfo, Title: "system started", Detail: "boot ok", Module: "kernel",
		Related: []string{}, Metadata: map[string]string{"pid": "1"},
	}
	evt2 := TimelineEvent{
		ID: NewUUID(), Timestamp: time.Now(), Category: CatNetwork,
		Level: EventWarning, Title: "latency spike", Detail: "500ms", Module: "net",
	}
	evt3 := TimelineEvent{
		ID: NewUUID(), Timestamp: time.Now(), Category: CatSecurity,
		Level: EventCritical, Title: "auth failure", Detail: "brute force", Module: "sec",
		Related: []string{"evt-ref-1"},
	}

	if err := s.InsertEvent(evt1, nil); err != nil {
		t.Fatal(err)
	}
	if err := s.InsertEvent(evt2, nil); err != nil {
		t.Fatal(err)
	}
	if err := s.InsertEvent(evt3, nil); err != nil {
		t.Fatal(err)
	}

	// Query all (no filters)
	events, err := s.QueryEvents("", "", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) < 3 {
		t.Fatalf("QueryEvents returned %d events, want at least 3", len(events))
	}

	// Filter by category
	sysEvents, err := s.QueryEvents("system", "", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(sysEvents) != 1 {
		t.Errorf("system events: got %d, want 1", len(sysEvents))
	}

	// Filter by level
	critEvents, err := s.QueryEvents("", "critical", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(critEvents) != 1 {
		t.Errorf("critical events: got %d, want 1", len(critEvents))
	}

	// Get by ID
	got, err := s.GetEventByID(evt3.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("GetEventByID returned nil")
	}
	if got.Title != "auth failure" {
		t.Errorf("event title = %q, want %q", got.Title, "auth failure")
	}
	if len(got.Related) != 1 || got.Related[0] != "evt-ref-1" {
		t.Errorf("related = %v, want [evt-ref-1]", got.Related)
	}

	// Get evt1 (has metadata)
	got1, err := s.GetEventByID(evt1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got1 == nil {
		t.Fatal("GetEventByID returned nil for evt1")
	}
	if got1.Metadata == nil || got1.Metadata["pid"] != "1" {
		t.Errorf("metadata = %v, want map[pid:1]", got1.Metadata)
	}
	if len(got1.Related) != 0 {
		t.Errorf("expected empty Related for evt1, got %v", got1.Related)
	}

	// GetByID not found
	missing, err := s.GetEventByID("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if missing != nil {
		t.Error("expected nil for nonexistent event")
	}

	// Offset test (skip newest 3, should return empty or fewer)
	offsetEvents, err := s.QueryEvents("", "", 10, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(offsetEvents) > 0 && len(offsetEvents) >= len(events) {
		t.Error("offset did not reduce results")
	}

	// Related and metadata should be empty when not set
	got2, err := s.GetEventByID(evt2.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got2 == nil {
		t.Fatal("GetEventByID returned nil for evt2")
	}
	if len(got2.Related) != 0 {
		t.Errorf("expected empty Related, got %v", got2.Related)
	}
}

func TestAlertsCRUD(t *testing.T) {
	s := newTestStorage(t)

	now := time.Now()
	alerts := []AlertRecord{
		{ID: NewUUID(), Timestamp: now, Level: "WARN", Metric: "cpu", Message: "cpu > 90%", Value: 95, Threshold: 90, Resolved: false},
		{ID: NewUUID(), Timestamp: now, Level: "CRITICAL", Metric: "memory", Message: "OOM", Value: 99, Threshold: 80, Resolved: false},
		{ID: NewUUID(), Timestamp: now, Level: "INFO", Metric: "disk", Message: "disk > 85%", Value: 86, Threshold: 85, Resolved: false},
	}

	for _, a := range alerts {
		if err := s.InsertAlert(a, nil); err != nil {
			t.Fatal(err)
		}
	}

	// Query history
	history, err := s.QueryAlertHistory(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) < 3 {
		t.Fatalf("got %d alerts, want at least 3", len(history))
	}

	// Resolve an alert
	if err := s.UpdateAlertResolved(alerts[0].ID, nil); err != nil {
		t.Fatal(err)
	}

	history2, err := s.QueryAlertHistory(10)
	if err != nil {
		t.Fatal(err)
	}
	foundResolved := false
	for _, a := range history2 {
		if a.ID == alerts[0].ID {
			foundResolved = true
			if !a.Resolved {
				t.Error("alert should be resolved after UpdateAlertResolved")
			}
			break
		}
	}
	if !foundResolved {
		t.Error("resolved alert not found in history")
	}

	// Limit test
	limited, err := s.QueryAlertHistory(2)
	if err != nil {
		t.Fatal(err)
	}
	if len(limited) > 2 {
		t.Errorf("limited query returned %d, want at most 2", len(limited))
	}

	// Default limit (0 uses 100)
	allAlerts, err := s.QueryAlertHistory(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(allAlerts) < 3 {
		t.Errorf("default limit query returned %d, want at least 3", len(allAlerts))
	}
}

func TestForensicsCRUD(t *testing.T) {
	s := newTestStorage(t)

	f1 := ForensicRecord{ID: NewUUID(), Timestamp: time.Now().UTC().Format(time.RFC3339), Type: "process", Title: "process scan", DataJSON: `{"pid":123}`, Metadata: "snapshot"}
	f2 := ForensicRecord{ID: NewUUID(), Timestamp: time.Now().UTC().Format(time.RFC3339), Type: "network", Title: "netstat", DataJSON: `{"port":80}`, Metadata: "snapshot"}

	if err := s.InsertForensic(f1); err != nil {
		t.Fatal(err)
	}
	if err := s.InsertForensic(f2); err != nil {
		t.Fatal(err)
	}

	// ListForensics returns summary fields only
	list, err := s.ListForensics()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) < 2 {
		t.Fatalf("ListForensics returned %d, want at least 2", len(list))
	}

	// GetForensic full record
	got, err := s.GetForensic(f1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("GetForensic returned nil")
	}
	if got.Type != "process" {
		t.Errorf("type = %q, want %q", got.Type, "process")
	}
	if got.DataJSON != `{"pid":123}` {
		t.Errorf("data_json = %q, want %q", got.DataJSON, `{"pid":123}`)
	}

	// Not found
	missing, err := s.GetForensic("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if missing != nil {
		t.Error("expected nil for nonexistent forensic")
	}
}

func TestReportsCRUD(t *testing.T) {
	s := newTestStorage(t)

	r1 := ReportRecord{ID: NewUUID(), Timestamp: time.Now().UTC().Format(time.RFC3339), Type: "health", Score: 85, DataJSON: `{"status":"good"}`}
	r2 := ReportRecord{ID: NewUUID(), Timestamp: time.Now().UTC().Format(time.RFC3339), Type: "security", Score: 60, DataJSON: `{"risk":"medium"}`}
	r3 := ReportRecord{ID: NewUUID(), Timestamp: time.Now().UTC().Format(time.RFC3339), Type: "health", Score: 90, DataJSON: `{"status":"excellent"}`}

	for _, r := range []ReportRecord{r1, r2, r3} {
		if err := s.InsertReport(r); err != nil {
			t.Fatal(err)
		}
	}

	// List by type
	healthReports, err := s.ListReportsByType("health")
	if err != nil {
		t.Fatal(err)
	}
	if len(healthReports) != 2 {
		t.Errorf("health reports: got %d, want 2", len(healthReports))
	}

	secReports, err := s.ListReportsByType("security")
	if err != nil {
		t.Fatal(err)
	}
	if len(secReports) != 1 {
		t.Errorf("security reports: got %d, want 1", len(secReports))
	}

	// List all
	all, err := s.ListAllReports()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) < 3 {
		t.Errorf("all reports: got %d, want at least 3", len(all))
	}

	// Get by ID
	got, err := s.GetReport(r2.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("GetReport returned nil")
	}
	if got.Type != "security" {
		t.Errorf("type = %q, want %q", got.Type, "security")
	}
	if got.DataJSON != `{"risk":"medium"}` {
		t.Errorf("data_json = %q, want %q", got.DataJSON, `{"risk":"medium"}`)
	}

	// Not found
	missing, err := s.GetReport("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if missing != nil {
		t.Error("expected nil for nonexistent report")
	}

	// Delete
	if err := s.DeleteReport(r1.ID); err != nil {
		t.Fatal(err)
	}
	got2, err := s.GetReport(r1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got2 != nil {
		t.Error("expected nil after delete")
	}
}

func TestIncidentsCRUD(t *testing.T) {
	s := newTestStorage(t)

	inc := IncidentRecord{
		ID: NewUUID(), Timestamp: time.Now().UTC().Format(time.RFC3339),
		Title: "breach attempt", Details: "multiple auth failures",
		ReportIDs: []string{"r1", "r2"}, Severity: "high",
	}
	if err := s.InsertIncident(inc); err != nil {
		t.Fatal(err)
	}

	list, err := s.ListIncidents()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("list incidents: got %d, want 1", len(list))
	}
	if list[0].Title != "breach attempt" {
		t.Errorf("title = %q, want %q", list[0].Title, "breach attempt")
	}
	if list[0].Severity != "high" {
		t.Errorf("severity = %q, want %q", list[0].Severity, "high")
	}
}

func TestBaselinesCRUD(t *testing.T) {
	s := newTestStorage(t)

	// Upsert baseline
	b1 := BaselineEntry{Metric: "cpu_usage", Avg: 45.2, StdDev: 10.1, LastUpdated: time.Now()}
	if err := s.UpsertBaseline(b1); err != nil {
		t.Fatal(err)
	}

	// Get
	got, err := s.GetBaseline("cpu_usage")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("GetBaseline returned nil")
	}
	if got.Avg != 45.2 {
		t.Errorf("avg = %f, want 45.2", got.Avg)
	}
	if got.StdDev != 10.1 {
		t.Errorf("stddev = %f, want 10.1", got.StdDev)
	}
	if got.Metric != "cpu_usage" {
		t.Errorf("metric = %q, want %q", got.Metric, "cpu_usage")
	}

	// Upsert again (should update)
	b2 := BaselineEntry{Metric: "cpu_usage", Avg: 50.0, StdDev: 15.0}
	if err := s.UpsertBaseline(b2); err != nil {
		t.Fatal(err)
	}
	got2, err := s.GetBaseline("cpu_usage")
	if err != nil {
		t.Fatal(err)
	}
	if got2.Avg != 50.0 {
		t.Errorf("after upsert, avg = %f, want 50.0", got2.Avg)
	}

	// Not found
	missing, err := s.GetBaseline("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if missing != nil {
		t.Error("expected nil for nonexistent baseline")
	}
}

func TestHealthScores(t *testing.T) {
	s := newTestStorage(t)

	if err := s.UpsertHealthScore(85); err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertHealthScore(92); err != nil {
		t.Fatal(err)
	}

	// Get today's trend
	trend, err := s.GetHealthScoreTrend(7)
	if err != nil {
		t.Fatal(err)
	}
	if len(trend) == 0 {
		t.Fatal("health score trend returned empty")
	}
	// Should have at most 1 entry (today was upserted twice)
	if len(trend) > 1 {
		t.Errorf("trend has %d entries, expected 1", len(trend))
	}
	// The score should be 92 (last upsert wins)
	for day, score := range trend {
		if score != 92 {
			t.Errorf("day %q score = %d, want 92", day, score)
		}
	}
}

func TestCustomWorkflows(t *testing.T) {
	s := newTestStorage(t)

	if err := s.UpsertCustomWorkflow("wf1", "health check", "runs health diag", `{"steps":[]}`); err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertCustomWorkflow("wf2", "cleanup", "removes temp", `{"steps":["rm"]}`); err != nil {
		t.Fatal(err)
	}

	list, err := s.ListCustomWorkflows()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) < 2 {
		t.Fatalf("list workflows: got %d, want at least 2", len(list))
	}

	// Verify contents
	found := false
	for _, wf := range list {
		if wf["id"] == "wf1" && wf["name"] == "health check" {
			found = true
			break
		}
	}
	if !found {
		t.Error("workflow wf1 not found in list")
	}

	// Delete
	if err := s.DeleteCustomWorkflow("wf1"); err != nil {
		t.Fatal(err)
	}
	list2, err := s.ListCustomWorkflows()
	if err != nil {
		t.Fatal(err)
	}
	for _, wf := range list2 {
		if wf["id"] == "wf1" {
			t.Error("wf1 should be deleted")
		}
	}
}

func TestReportRulesCRUD(t *testing.T) {
	s := newTestStorage(t)

	r1 := AutoReportRule{
		ID: NewUUID(), Name: "cpu alert rule", Description: "alert when cpu high",
		Metric: "cpu_usage", Condition: "GT", Threshold: 90,
		ReportType: "health", Schedule: "on_alert", Enabled: true,
	}
	r2 := AutoReportRule{
		ID: NewUUID(), Name: "memory rule", Description: "alert on memory",
		Metric: "memory_used", Condition: "GT", Threshold: 80,
		ReportType: "health", Schedule: "daily", Enabled: false,
	}

	if err := s.InsertReportRule(r1); err != nil {
		t.Fatal(err)
	}
	if err := s.InsertReportRule(r2); err != nil {
		t.Fatal(err)
	}

	// List all rules
	list, err := s.ListReportRules()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) < 2 {
		t.Fatalf("list rules: got %d, want at least 2", len(list))
	}

	// Get by ID
	got, err := s.GetReportRule(r1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("GetReportRule returned nil")
	}
	if got.Name != "cpu alert rule" {
		t.Errorf("name = %q, want %q", got.Name, "cpu alert rule")
	}
	if !got.Enabled {
		t.Error("rule should be enabled")
	}

	// Update
	r1.Enabled = false
	r1.Threshold = 95
	if err := s.UpdateReportRule(r1); err != nil {
		t.Fatal(err)
	}
	got2, err := s.GetReportRule(r1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got2.Enabled {
		t.Error("rule should be disabled after update")
	}
	if got2.Threshold != 95 {
		t.Errorf("threshold = %f, want 95", got2.Threshold)
	}

	// Not found
	missing, err := s.GetReportRule("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if missing != nil {
		t.Error("expected nil for nonexistent rule")
	}

	// Delete
	if err := s.DeleteReportRule(r1.ID); err != nil {
		t.Fatal(err)
	}
	got3, err := s.GetReportRule(r1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got3 != nil {
		t.Error("expected nil after delete")
	}
}

func TestPruneTextPK(t *testing.T) {
	s := newTestStorage(t)

	// Insert records with manually old timestamps via raw SQL to bypass helper functions
	oldTime := time.Now().Add(-30 * 24 * time.Hour).UTC().Format(time.RFC3339)

	// Events (text PK)
	evtID := "prune-test-event"
	if _, err := s.db.Exec(
		`INSERT INTO events (id, timestamp, category, level, title) VALUES (?, ?, ?, ?, ?)`,
		evtID, oldTime, "system", "info", "old event",
	); err != nil {
		t.Fatal(err)
	}

	// Alerts (text PK, resolved)
	alertID := "prune-test-alert"
	if _, err := s.db.Exec(
		`INSERT INTO alerts (id, timestamp, level, metric, message, value, threshold, resolved) VALUES (?, ?, ?, ?, ?, ?, ?, 1)`,
		alertID, oldTime, "WARN", "test", "old", 0, 0,
	); err != nil {
		t.Fatal(err)
	}

	// Forensics (text PK)
	forensicID := "prune-test-forensic"
	if _, err := s.db.Exec(
		`INSERT INTO forensics (id, timestamp, type, title, data_json) VALUES (?, ?, ?, ?, ?)`,
		forensicID, oldTime, "snapshot", "old snapshot", "{}",
	); err != nil {
		t.Fatal(err)
	}

	// Reports (text PK)
	reportID := "prune-test-report"
	if _, err := s.db.Exec(
		`INSERT INTO reports (id, timestamp, type, score, data_json) VALUES (?, ?, ?, ?, ?)`,
		reportID, oldTime, "health", 50, "{}",
	); err != nil {
		t.Fatal(err)
	}

	// Incidents (text PK)
	incidentID := "prune-test-incident"
	if _, err := s.db.Exec(
		`INSERT INTO incidents (id, timestamp, title, details, severity) VALUES (?, ?, ?, ?, ?)`,
		incidentID, oldTime, "old incident", "details", "low",
	); err != nil {
		t.Fatal(err)
	}

	// Verify they exist before prune
	for _, tc := range []struct{ table, id string }{
		{"events", evtID}, {"alerts", alertID}, {"forensics", forensicID},
		{"reports", reportID}, {"incidents", incidentID},
	} {
		var count int
		if err := s.db.QueryRow(
			fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE id = ?", tc.table), tc.id,
		).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Errorf("before prune: %s id=%s count=%d, want 1", tc.table, tc.id, count)
		}
	}

	// Prune with 14-day retention (should remove all old records)
	s.Prune(14 * 24 * time.Hour)

	// Verify they were pruned
	for _, tc := range []struct{ table, id string }{
		{"events", evtID}, {"alerts", alertID}, {"forensics", forensicID},
		{"reports", reportID}, {"incidents", incidentID},
	} {
		var count int
		if err := s.db.QueryRow(
			fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE id = ?", tc.table), tc.id,
		).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Errorf("after prune: %s id=%s still exists (count=%d), want 0", tc.table, tc.id, count)
		}
	}
}

func TestMigrationVersioning(t *testing.T) {
	s := newTestStorage(t)

	version, err := s.getSchemaVersion()
	if err != nil {
		t.Fatal(err)
	}
	if version < 2 {
		t.Errorf("schema version = %d, want at least 2", version)
	}

	// Verify v2 index exists
	var idxName string
	err = s.db.QueryRow(
		`SELECT name FROM sqlite_master WHERE type = 'index' AND name = 'idx_logs_level_time'`,
	).Scan(&idxName)
	if err != nil {
		t.Fatal("idx_logs_level_time index not found after migration")
	}
}

func TestTimestampRoundTrip(t *testing.T) {
	// Test formatStorageTimestamp and parseStorageTimestamp with various times
	now := time.Now().UTC().Truncate(time.Second) // truncate to match SQLite precision
	formatted := formatStorageTimestamp(now)

	parsed, err := parseStorageTimestamp(formatted)
	if err != nil {
		t.Fatalf("parseStorageTimestamp(%q) failed: %v", formatted, err)
	}

	if !parsed.Equal(now) {
		t.Errorf("round-trip: got %v, want %v", parsed, now)
	}

	// Test legacy format
	legacy := now.Format("2006-01-02 15:04:05")
	parsed2, err := parseStorageTimestamp(legacy)
	if err != nil {
		t.Fatalf("parseStorageTimestamp(legacy %q) failed: %v", legacy, err)
	}
	if !parsed2.Equal(now) {
		t.Errorf("legacy parse: got %v, want %v", parsed2, now)
	}
}

func TestSettingsCRUD(t *testing.T) {
	s := newTestStorage(t)

	// Insert settings using raw SQL (no InsertSetting method exists)
	if _, err := s.db.Exec(`INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)`, "test_key", "test_value"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(`INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)`, "retention_days", "14"); err != nil {
		t.Fatal(err)
	}

	// GetSetting
	val, err := s.GetSetting("test_key")
	if err != nil {
		t.Fatal(err)
	}
	if val != "test_value" {
		t.Errorf("GetSetting = %q, want %q", val, "test_value")
	}

	// GetSetting not found
	val2, err := s.GetSetting("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if val2 != "" {
		t.Errorf("expected empty for nonexistent setting, got %q", val2)
	}

	// ListSettings
	all, err := s.ListSettings()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) < 2 {
		t.Errorf("ListSettings returned %d, want at least 2", len(all))
	}
	if all["test_key"] != "test_value" {
		t.Errorf("settings map: got %v", all)
	}
}

func TestGetMetricHistoryWithTimestamps(t *testing.T) {
	s := newTestStorage(t)

	if err := s.InsertMetric("timed_metric", "ms", 100.0); err != nil {
		t.Fatal(err)
	}
	if err := s.InsertMetric("timed_metric", "ms", 200.0); err != nil {
		t.Fatal(err)
	}
	s.flushMetrics()

	points, err := s.GetMetricHistoryWithTimestamps("timed_metric", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 2 {
		t.Fatalf("got %d points, want 2", len(points))
	}
	if points[0].Value != 100.0 || points[1].Value != 200.0 {
		t.Errorf("values = [%f, %f], want [100, 200]", points[0].Value, points[1].Value)
	}
	if points[0].Timestamp.IsZero() {
		t.Error("first point has zero timestamp")
	}
}

func TestQueryLogTimeline_EmptyDB(t *testing.T) {
	s := newTestStorage(t)
	s.flushMetrics()

	// After init + flush, the DB has exactly 1 log (the "Storage initialized" message).
	// Verify the timeline query returns it as one bucket.
	buckets, err := s.QueryLogTimeline(24)
	if err != nil {
		t.Fatalf("QueryLogTimeline(24) failed: %v", err)
	}
	if buckets == nil {
		t.Error("QueryLogTimeline(24) returned nil, expected non-nil slice")
	}
	if len(buckets) < 1 {
		t.Error("QueryLogTimeline(24) returned 0 buckets, expected at least 1 for init log")
	}
	if len(buckets) > 0 && buckets[0].Total < 1 {
		t.Errorf("QueryLogTimeline(24) first bucket has Total=%d, want >= 1", buckets[0].Total)
	}

	// 1h timeline (5-min buckets)
	buckets, err = s.QueryLogTimeline(1)
	if err != nil {
		t.Fatalf("QueryLogTimeline(1) failed: %v", err)
	}
	if len(buckets) < 1 {
		t.Error("QueryLogTimeline(1) returned 0 buckets, expected at least 1 for init log")
	}
}

func TestQueryLogTimeline_ZeroHours(t *testing.T) {
	s := newTestStorage(t)
	s.flushMetrics()

	// hours=0 should be handled gracefully (returns empty since it matches future only)
	buckets, err := s.QueryLogTimeline(0)
	if err != nil {
		t.Fatalf("QueryLogTimeline(0) failed: %v", err)
	}
	if buckets == nil {
		t.Error("QueryLogTimeline(0) returned nil, expected empty slice")
	}

	// Negative hours should also work without panic
	buckets, err = s.QueryLogTimeline(-1)
	if err != nil {
		t.Fatalf("QueryLogTimeline(-1) failed: %v", err)
	}
	if buckets == nil {
		t.Error("QueryLogTimeline(-1) returned nil, expected empty slice")
	}
}

func TestCountLogsAfter(t *testing.T) {
	s := newTestStorage(t)

	// Use direct batch inserts to avoid the InitStorage log
	now := time.Now().UTC().Truncate(time.Millisecond)

	s.insertLogsBatch([]LogWrite{
		{Level: "INFO", Module: "test", Message: "msg1", Time: now.Add(-30 * time.Minute)},
		{Level: "ERROR", Module: "test", Message: "msg2", Time: now.Add(-5 * time.Minute)},
	})

	// Should find both when looking 1 hour back
	if n := s.CountLogsAfter(now.Add(-1 * time.Hour)); n != 2 {
		t.Errorf("CountLogsAfter(-1h) = %d, want 2", n)
	}

	// Should find none in the future
	if n := s.CountLogsAfter(now.Add(1 * time.Hour)); n != 0 {
		t.Errorf("CountLogsAfter(+1h) = %d, want 0", n)
	}

	// Both logs are within 45 minutes
	if n := s.CountLogsAfter(now.Add(-45 * time.Minute)); n != 2 {
		t.Errorf("CountLogsAfter(-45m) = %d, want 2", n)
	}

	// Only the -5m log is within 10 minutes
	if n := s.CountLogsAfter(now.Add(-10 * time.Minute)); n != 1 {
		t.Errorf("CountLogsAfter(-10m) = %d, want 1", n)
	}
}

func TestCountLogsByLevel(t *testing.T) {
	s := newTestStorage(t)

	// Use direct batch inserts to avoid the InitStorage log
	s.insertLogsBatch([]LogWrite{
		{Level: "INFO", Module: "test", Message: "info1", Time: time.Now()},
		{Level: "INFO", Module: "test", Message: "info2", Time: time.Now()},
		{Level: "ERROR", Module: "test", Message: "error1", Time: time.Now()},
		{Level: "WARN", Module: "test", Message: "warn1", Time: time.Now()},
	})

	if n := s.CountLogsByLevel("INFO"); n != 2 {
		t.Errorf("CountLogsByLevel(INFO) = %d, want 2", n)
	}
	if n := s.CountLogsByLevel("ERROR"); n != 1 {
		t.Errorf("CountLogsByLevel(ERROR) = %d, want 1", n)
	}
	if n := s.CountLogsByLevel("WARN"); n != 1 {
		t.Errorf("CountLogsByLevel(WARN) = %d, want 1", n)
	}
	if n := s.CountLogsByLevel("DEBUG"); n != 0 {
		t.Errorf("CountLogsByLevel(DEBUG) = %d, want 0", n)
	}
}

func TestCountLogsByLevelInRange(t *testing.T) {
	s := newTestStorage(t)

	now := time.Now().UTC().Truncate(time.Millisecond)

	// Use direct batch inserts with explicit timestamps
	s.insertLogsBatch([]LogWrite{
		{Level: "INFO", Module: "test", Message: "old", Time: now.Add(-2 * time.Hour)},
		{Level: "ERROR", Module: "test", Message: "recent1", Time: now},
		{Level: "ERROR", Module: "test", Message: "recent2", Time: now.Add(30 * time.Second)},
		{Level: "WARN", Module: "test", Message: "recent3", Time: now.Add(time.Minute)},
	})

	// Count recent errors (last/next hour)
	if n := s.CountLogsByLevelInRange("ERROR", now.Add(-1*time.Hour), now.Add(1*time.Hour)); n != 2 {
		t.Errorf("CountLogsByLevelInRange(ERROR, -1h, +1h) = %d, want 2", n)
	}

	// Count INFO in same range — the old INFO is at now-2h, outside range
	if n := s.CountLogsByLevelInRange("INFO", now.Add(-1*time.Hour), now.Add(1*time.Hour)); n != 0 {
		t.Errorf("CountLogsByLevelInRange(INFO, -1h, +1h) = %d, want 0", n)
	}

	// Count outside the range — nothing
	if n := s.CountLogsByLevelInRange("ERROR", now.Add(2*time.Hour), now.Add(3*time.Hour)); n != 0 {
		t.Errorf("CountLogsByLevelInRange(ERROR, +2h, +3h) = %d, want 0", n)
	}
}

func TestTopLogSources(t *testing.T) {
	s := newTestStorage(t)

	// Use direct batch inserts to avoid InitStorage log
	// Insert logs with known modules
	s.insertLogsBatch([]LogWrite{
		{Level: "INFO", Module: "alpha", Message: "msg1", Time: time.Now()},
		{Level: "ERROR", Module: "alpha", Message: "msg2", Time: time.Now()},
		{Level: "WARN", Module: "beta", Message: "msg3", Time: time.Now()},
		{Level: "DEBUG", Module: "gamma", Message: "msg4", Time: time.Now()},
	})

	sources, err := s.TopLogSources(10)
	if err != nil {
		t.Fatalf("TopLogSources(10) failed: %v", err)
	}
	if len(sources) == 0 {
		t.Fatal("TopLogSources(10) returned 0 sources, expected at least 1")
	}

	// alpha should be first with 2 logs
	found := false
	for _, src := range sources {
		if src.Source == "alpha" {
			found = true
			if src.Count != 2 {
				t.Errorf("alpha source count = %d, want 2", src.Count)
			}
			break
		}
	}
	if !found {
		t.Errorf("alpha not found in top sources: %+v", sources)
	}
}

func TestAlertRulePersistence(t *testing.T) {
	s := newTestStorage(t)

	// Insert a rule
	rule := AlertRuleRecord{
		Metric:    "cpu.percent",
		Condition: ">",
		Threshold: 85.0,
		FlapCount: 3,
		Severity:  "WARNING",
		Message:   "CPU high: {value}",
	}
	if err := s.InsertAlertRule(rule, nil); err != nil {
		t.Fatal(err)
	}

	// Query it back
	rules, err := s.QueryAlertRules()
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 1 {
		t.Fatalf("QueryAlertRules returned %d rules, want 1", len(rules))
	}
	got := rules[0]
	if got.Metric != "cpu.percent" || got.Threshold != 85.0 || got.FlapCount != 3 ||
		got.Severity != "WARNING" || got.Condition != ">" || got.Message != "CPU high: {value}" {
		t.Errorf("rule round-trip mismatch: %+v", got)
	}

	// Upsert same metric+threshold updates in place (no duplicate)
	rule.Threshold = 90.0
	if err := s.InsertAlertRule(rule, nil); err != nil {
		t.Fatal(err)
	}
	rules, err = s.QueryAlertRules()
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 2 {
		t.Fatalf("after upsert with new threshold, got %d rules, want 2", len(rules))
	}

	// Delete one
	if err := s.DeleteAlertRule("cpu.percent", 85.0, nil); err != nil {
		t.Fatal(err)
	}
	rules, err = s.QueryAlertRules()
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 1 {
		t.Fatalf("after delete, got %d rules, want 1", len(rules))
	}
	if rules[0].Threshold != 90.0 {
		t.Errorf("remaining rule threshold = %v, want 90.0", rules[0].Threshold)
	}
}

func TestAlertEngineRuleRestore(t *testing.T) {
	s := newTestStorage(t)

	// Persist a couple of rules directly to storage
	_ = s.InsertAlertRule(AlertRuleRecord{Metric: "cpu.percent", Condition: ">", Threshold: 90, FlapCount: 2, Severity: "CRITICAL"}, nil)
	_ = s.InsertAlertRule(AlertRuleRecord{Metric: "mem.used", Condition: "<", Threshold: 10, FlapCount: 1, Severity: "INFO"}, nil)

	// Build an engine and restore rules from DB
	dp := NewDataPipeline(DefaultCollectionConfig())
	ae := NewAlertEngine(dp)
	records, err := s.QueryAlertRules()
	if err != nil {
		t.Fatal(err)
	}
	ae.RestoreRulesFromDB(records)

	rules := ae.GetRules()
	if len(rules) != 2 {
		t.Fatalf("RestoreRulesFromDB restored %d rules, want 2", len(rules))
	}

	// Verify condition parsing (">" -> GT, "<" -> LT)
	var cpuRule, memRule *AlertRule
	for i := range rules {
		switch rules[i].Metric {
		case "cpu.percent":
			cpuRule = &rules[i]
		case "mem.used":
			memRule = &rules[i]
		}
	}
	if cpuRule == nil || cpuRule.Condition != AlertGT || cpuRule.Severity != AlertCritical {
		t.Errorf("cpu rule restore mismatch: %+v", cpuRule)
	}
	if memRule == nil || memRule.Condition != AlertLT || memRule.Severity != AlertInfo {
		t.Errorf("mem rule restore mismatch: %+v", memRule)
	}
}
