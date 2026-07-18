package common

import (
	"sync"
)

type KnowledgeManager struct {
	mu    sync.RWMutex
	state SystemKnowledge
}

var globalKnowledge *KnowledgeManager

func InitKnowledge() {
	globalKnowledge = &KnowledgeManager{
		state: SystemKnowledge{},
	}
}

func GetKnowledge() *KnowledgeManager {
	if globalKnowledge == nil {
		InitKnowledge()
	}
	return globalKnowledge
}

func (k *KnowledgeManager) Update(fn func(*SystemKnowledge)) {
	k.mu.Lock()
	defer k.mu.Unlock()
	fn(&k.state)
}

func (k *KnowledgeManager) GetSnapshot() SystemKnowledge {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.state
}
