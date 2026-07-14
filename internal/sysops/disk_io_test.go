package sysops

import "testing"

func TestGetDiskIO(t *testing.T) {
	stats, err := GetDiskIO()
	if err != nil {
		t.Fatalf("GetDiskIO returned error: %v", err)
	}
	if len(stats.Disks) == 0 {
		t.Error("No disk I/O stats returned")
	}
}
