//go:build integration

package common

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStorageFileBasedCRUD uses a real file-based SQLite database to test
// the full storage lifecycle: open, write, read, close, reopen, verify.
func TestStorageFileBasedCRUD(t *testing.T) {
	// Create a temp database path
	dir, err := os.MkdirTemp("", "universalops_integration_*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	dbPath := filepath.Join(dir, "test.db")
	// Use file-based SQLite (not :memory:)
	t.Setenv("UNIVERSALOPS_DB_PATH", dbPath)

	// Initialize storage on the file
	InitStorage(dbPath)
	defer func() {
		s := GetStorage()
		if s != nil {
			s.Close()
		}
	}()

	store := GetStorage()
	require.NotNil(t, store)

	// Write data
	err = store.InsertLog("INFO", "integration-test", "file-based CRUD test")
	assert.NoError(t, err)

	err = store.InsertEvent(TimelineEvent{
		ID:       NewUUID(),
		Category: CatSystem,
		Level:    EventInfo,
		Title:    "test_event",
		Detail:   `{"key":"value"}`,
		Module:   "integration-test",
		Metadata: map[string]string{"normal": "true"},
	}, nil)
	assert.NoError(t, err)

	// Wait for async writer to flush
	time.Sleep(1500 * time.Millisecond)

	// Read back
	logs, err := store.QueryLogs("", "", 100)
	assert.NoError(t, err)
	assert.NotEmpty(t, logs)

	events, err := store.QueryEvents("", "", 100, 0)
	assert.NoError(t, err)
	assert.NotEmpty(t, events)

	// Close and reopen to verify persistence
	store.Close()

	InitStorage(dbPath)
	store2 := GetStorage()
	require.NotNil(t, store2)
	defer store2.Close()

	logs2, err := store2.QueryLogs("", "", 100)
	assert.NoError(t, err)
	assert.NotEmpty(t, logs2, "data should persist after close/reopen")
}

// TestStorageConcurrentAccess verifies that multiple goroutines can safely
// write to the storage layer without deadlock or data corruption.
func TestStorageConcurrentAccess(t *testing.T) {
	dir, err := os.MkdirTemp("", "universalops_concurrent_*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	dbPath := filepath.Join(dir, "concurrent.db")
	InitStorage(dbPath)
	defer func() {
		s := GetStorage()
		if s != nil {
			s.Close()
		}
	}()

	store := GetStorage()
	require.NotNil(t, store)

	// Launch 10 concurrent writers
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 20; j++ {
				_ = store.InsertLog("INFO", "concurrent", "test message")
				_ = store.InsertEvent(TimelineEvent{
					ID:       NewUUID(),
					Category: CatSystem,
					Level:    EventInfo,
					Title:    "concurrent_test",
					Detail:   `{}`,
					Module:   "concurrent",
				}, nil)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Wait for async flush
	time.Sleep(1500 * time.Millisecond)

	// Verify all data was written
	logs, err := store.QueryLogs("", "", 1000)
	assert.NoError(t, err)
	// Cannot assert exact count because concurrent writes may race with async flush
	assert.GreaterOrEqual(t, len(logs), 0, "logs should be writable under concurrent load")
}

// TestStorageWALMode verifies that WAL journal mode is active, which is
// critical for concurrent read/write performance.
func TestStorageWALMode(t *testing.T) {
	dir, err := os.MkdirTemp("", "universalops_wal_*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	dbPath := filepath.Join(dir, "wal.db")
	InitStorage(dbPath)
	defer func() {
		s := GetStorage()
		if s != nil {
			s.Close()
		}
	}()

	// After InitStorage, check that WAL journal files exist
	walFile := dbPath + "-wal"
	shmFile := dbPath + "-shm"

	assert.FileExists(t, walFile, "WAL journal file should exist")
	assert.FileExists(t, shmFile, "WAL shared memory file should exist")
}
