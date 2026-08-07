package common

import (
	"sync"
)

// KnowledgeManager provides a unified view of the system's "Current Truth".
// PERFORMANCE OPTIMIZATION: It no longer stores duplicated state.
// Instead, it acts as a high-performance facade over the DataPipeline
// and other subsystems to ensure data consistency across the app (AI vs Dashboard).
type KnowledgeManager struct {
	pipeline *DataPipeline

	// mu protects non-metric state (SecurityGrade, etc.)
	mu            sync.RWMutex
	securityGrade string
	anomalies     int
	activeConns   int
	uptime        string
	interfaces    []string
}

var globalKnowledge *KnowledgeManager

func InitKnowledge(pipeline *DataPipeline) {
	globalKnowledge = &KnowledgeManager{
		pipeline: pipeline,
	}
}

func GetKnowledge() *KnowledgeManager {
	return globalKnowledge
}

// UpdateSystemState updates the heuristic findings and environmental context.
func (k *KnowledgeManager) UpdateSystemState(grade string, anomalies int, conns int, uptime string, ifaces []string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.securityGrade = grade
	k.anomalies = anomalies
	k.activeConns = conns
	k.uptime = uptime
	k.interfaces = ifaces
}

// GetSnapshot synthesizes the "Current Truth" from all sources.
// This ensures that AI and Dashboard always see the SAME numbers.
func (k *KnowledgeManager) GetSnapshot() SystemKnowledge {
	k.mu.RLock()
	defer k.mu.RUnlock()

	sk := SystemKnowledge{
		SecurityGrade:     k.securityGrade,
		Anomalies:         k.anomalies,
		ActiveConns:       k.activeConns,
		SystemUptime:      k.uptime,
		NetworkInterfaces: k.interfaces,
	}

	if k.pipeline != nil {
		// Core system metrics
		sk.SystemCPUUtilization = k.pipeline.GetLastValue(MetricCPU)
		sk.CPUTrend = k.pipeline.GetTrend(MetricCPU).Direction.String()
		sk.SystemMemoryUsage = k.pipeline.GetLastValue(MetricMem)
		sk.MemoryTrend = k.pipeline.GetTrend(MetricMem).Direction.String()
		sk.SystemDiskUsage = k.pipeline.GetLastValue(MetricDisk)
		sk.DiskTrend = k.pipeline.GetTrend(MetricDisk).Direction.String()
		sk.ProcessCount = int(k.pipeline.GetLastValue(MetricProcCnt))
		sk.ConnectionCount = k.activeConns

		// Network metrics
		sk.SystemNetRX = k.pipeline.GetLastValue(MetricNetRX)
		sk.NetRXTrend = k.pipeline.GetTrend(MetricNetRX).Direction.String()
		sk.SystemNetTX = k.pipeline.GetLastValue(MetricNetTX)
		sk.NetTXTrend = k.pipeline.GetTrend(MetricNetTX).Direction.String()

		// Load averages
		sk.SystemLoad1 = k.pipeline.GetLastValue(MetricLoad1)
		sk.SystemLoad5 = k.pipeline.GetLastValue(MetricLoad5)
		sk.SystemLoad15 = k.pipeline.GetLastValue(MetricLoad15)

		// Swap
		sk.SystemSwapUsage = k.pipeline.GetLastValue(MetricSwap)
		sk.SwapTrend = k.pipeline.GetTrend(MetricSwap).Direction.String()

		// Disk I/O
		sk.SystemDiskIORead = k.pipeline.GetLastValue(MetricDiskIORead)
		sk.DiskIOReadTrend = k.pipeline.GetTrend(MetricDiskIORead).Direction.String()
		sk.SystemDiskIOWrite = k.pipeline.GetLastValue(MetricDiskIOWrite)
		sk.DiskIOWriteTrend = k.pipeline.GetTrend(MetricDiskIOWrite).Direction.String()
	}

	return sk
}

func (d TrendDirection) String() string {
	switch d {
	case TrendRising:
		return "rising"
	case TrendFalling:
		return "falling"
	default:
		return "stable"
	}
}
