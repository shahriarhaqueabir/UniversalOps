package common

import (
	"context"
	"time"
)

type CollectorID string

const (
	CollectorCPU    CollectorID = "cpu"
	CollectorMem    CollectorID = "memory"
	CollectorDisk   CollectorID = "disk"
	CollectorNet    CollectorID = "network"
	CollectorTemp   CollectorID = "temperature"
	CollectorProc   CollectorID = "processes"
)

type MetricSample struct {
	Name  string
	Unit  string
	Value float64
}

type CollectorInfo struct {
	ID              CollectorID
	Name            string
	Description     string
	DefaultInterval time.Duration
	DefaultEnabled  bool
}

type Collector interface {
	Info() CollectorInfo
	Collect(ctx context.Context) ([]MetricSample, error)
}
