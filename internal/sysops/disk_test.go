package sysops

import (
	"testing"
)

func TestGetDiskStats(t *testing.T) {
	stats, err := GetDiskStats()
	if err != nil {
		t.Fatalf("GetDiskStats failed: %v", err)
	}

	// Should have at least one partition on a live system
	if len(stats.Usage) == 0 {
		t.Error("No disk partitions found")
	}

	for _, u := range stats.Usage {
		if u.Mountpoint == "" {
			t.Error("Disk partition missing Mountpoint")
		}
		if u.TotalBytes == 0 {
			t.Errorf("Disk %s reported 0 bytes", u.Mountpoint)
		}
	}
}
