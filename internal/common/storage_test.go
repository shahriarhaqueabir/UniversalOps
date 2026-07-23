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
