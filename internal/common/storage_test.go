package common

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Helper to create a Storage with temp DB for tests.
func newTestStorage(t *testing.T) *Storage {
	t.Helper()

	f, err := os.CreateTemp("", "hawkward-test-*.db")
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
		globalStorage = nil
	})
	return s
}

func TestInitStoragePragmas(t *testing.T) {
	f, err := os.CreateTemp("", "hawkward-test-*.db")
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

	// Wait a brief moment for the InitStorage LogInfo goroutine to settle
	// to make the total count deterministic (either 3 or 4).
	time.Sleep(50 * time.Millisecond)

	if err := s.InsertLog("INFO", "test", "hello world"); err != nil {
		t.Fatal(err)
	}
	if err := s.InsertLog("ERROR", "network", "connection refused"); err != nil {
		t.Fatal(err)
	}
	if err := s.InsertLog("INFO", "network", "retry succeeded"); err != nil {
		t.Fatal(err)
	}

	entries, err := s.QueryLogs("", "", 10)
	if err != nil {
		t.Fatal(err)
	}
	// We expect at least the 3 entries we just inserted.
	// The 4th entry comes from InitStorage's LogInfo (module=SYSTEM) but is async.
	if len(entries) < 3 {
		t.Fatalf("got %d log entries, want at least 3", len(entries))
	}

	// Filter by level
	entries, err = s.QueryLogs("ERROR", "", 10)
	if err != nil {
		t.Fatal(err)
	}
	// We check for ERROR entries specifically; should be at least 1.
	foundError := false
	for _, e := range entries {
		if e.Level == "ERROR" && e.Message == "connection refused" {
			foundError = true
			break
		}
	}
	if !foundError {
		t.Errorf("filtered by ERROR: did not find 'connection refused'. entries: %+v", entries)
	}

	// Filter by search
	entries, err = s.QueryLogs("", "network", 10)
	if err != nil {
		t.Fatal(err)
	}
	// Should find our 2 network entries.
	if len(entries) < 2 {
		t.Errorf("search 'network': got %d entries, want at least 2", len(entries))
	}

	// Limit
	entries, err = s.QueryLogs("INFO", "", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("limited INFO query: got %d, want 1", len(entries))
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
	f, err := os.CreateTemp("", "hawkward-test-*.db")
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
