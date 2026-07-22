package app

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/sysops"
)

func TestCollectTelemetry(t *testing.T) {
	// Initialize subsystems needed for collectors
	cfg := common.DefaultCollectionConfig()
	pipeline := common.NewDataPipeline(cfg)
	eventBus := common.NewEventBus(100)
	sysOps := NewSysOps()
	netOps := NewNetOps(eventBus)

	// Create app instance (needed for some collectors)
	app := &App{
		pipeline: pipeline,
		SysOps:   sysOps,
		NetOps:   netOps,
	}

	registry := common.NewCollectorRegistry()
	RegisterCollectors(registry, app)

	// Phase 2: Warm up background data caches
	sysops.UpdateProcessSnapshot()

	fmt.Println("\n| Collector ID | Duration | Mallocs | Total Alloc | Samples |")
	fmt.Println("| :--- | :--- | :--- | :--- | :--- |")

	for _, c := range registry.AllCollectors() {
		info := c.Info()

		// Warm up
		c.Collect(context.Background())

		var m1, m2 runtime.MemStats
		runtime.GC() // Clean start
		runtime.ReadMemStats(&m1)
		start := time.Now()

		samples, _ := c.Collect(context.Background())

		duration := time.Since(start)
		runtime.ReadMemStats(&m2)

		allocs := m2.Mallocs - m1.Mallocs
		bytes := m2.TotalAlloc - m1.TotalAlloc

		fmt.Printf("| %s | %v | %d | %d B | %d |\n",
			info.ID, duration, allocs, bytes, len(samples))
	}
	fmt.Println()
}

func TestProcessSnapshotRace(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(10)

	// Concurrent readers
	for i := 0; i < 5; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, _ = sysops.GetTopProcesses(10)
				_ = sysops.GetProcessCount()
				_ = sysops.GetTotalOpenFDs()
			}
		}()
	}

	// Concurrent writers
	for i := 0; i < 5; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_ = sysops.UpdateProcessSnapshot()
			}
		}()
	}

	wg.Wait()
}

// TestColdStartProcessCount verifies that GetProcessCount and GetTotalOpenFDs
// return 0 (not panic or stale data) when called before the first snapshot.
// This simulates the race window between scheduler.Start() and the first
// process worker tick in Startup().
func TestColdStartProcessCount(t *testing.T) {
	// Simulate cold-start: no snapshot taken yet
	count := sysops.GetProcessCount()
	if count != 0 {
		t.Logf("GetProcessCount before snapshot: %d (expected 0)", count)
	}

	fds := sysops.GetTotalOpenFDs()
	if fds != 0 {
		t.Logf("GetTotalOpenFDs before snapshot: %d (expected 0)", fds)
	}

	// GetTopProcesses should trigger a synchronous snapshot
	procs, err := sysops.GetTopProcesses(5)
	if err != nil {
		t.Fatalf("GetTopProcesses before snapshot failed: %v", err)
	}
	if len(procs) == 0 {
		t.Log("GetTopProcesses returned 0 processes after cold-start fallback")
	} else {
		t.Logf("GetTopProcesses returned %d processes after cold-start fallback", len(procs))
	}

	// After GetTopProcesses triggered a snapshot, GetProcessCount should be > 0
	countAfter := sysops.GetProcessCount()
	if countAfter == 0 {
		t.Error("GetProcessCount still 0 after GetTopProcesses triggered snapshot")
	} else {
		t.Logf("GetProcessCount after fallback snapshot: %d", countAfter)
	}
}

func BenchmarkCollectors(b *testing.B) {
	cfg := common.DefaultCollectionConfig()
	pipeline := common.NewDataPipeline(cfg)
	eventBus := common.NewEventBus(100)
	sysOps := NewSysOps()
	netOps := NewNetOps(eventBus)
	app := &App{pipeline: pipeline, SysOps: sysOps, NetOps: netOps}
	registry := common.NewCollectorRegistry()
	RegisterCollectors(registry, app)
	sysops.UpdateProcessSnapshot()

	for _, c := range registry.AllCollectors() {
		info := c.Info()
		b.Run(string(info.ID), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = c.Collect(context.Background())
			}
		})
	}
}
