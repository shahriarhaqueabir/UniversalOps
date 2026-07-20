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

// UpdateSecurityState updates the heuristic findings.
func (k *KnowledgeManager) UpdateSecurityState(grade string, anomalies int, conns int, uptime string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.securityGrade = grade
	k.anomalies = anomalies
	k.activeConns = conns
	k.uptime = uptime
}

// GetSnapshot synthesizes the "Current Truth" from all sources.
// This ensures that AI and Dashboard always see the SAME numbers.
func (k *KnowledgeManager) GetSnapshot() SystemKnowledge {
	k.mu.RLock()
	defer k.mu.RUnlock()

	sk := SystemKnowledge{
		SecurityGrade: k.securityGrade,
		Anomalies:     k.anomalies,
		ActiveConns:   k.activeConns,
		Uptime:        k.uptime,
	}

	if k.pipeline != nil {
		sk.CPUUsage = k.pipeline.GetLastValue(MetricCPU)
		sk.MemoryUsage = k.pipeline.GetLastValue(MetricMem)
		sk.DiskUsage = k.pipeline.GetLastValue(MetricDisk)
	}

	return sk
}
