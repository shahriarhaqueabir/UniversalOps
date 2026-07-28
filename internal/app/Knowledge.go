package app

import (
	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

// KnowledgeAPI exposes the unified System Knowledge Layer to the frontend.
type KnowledgeAPI struct{}

// NewKnowledgeAPI creates a new KnowledgeAPI facade.
func NewKnowledgeAPI() *KnowledgeAPI {
	return &KnowledgeAPI{}
}

// GetSnapshot returns the current unified system state.
// Returns a zero-value SystemKnowledge if the KnowledgeManager has not been
// initialized yet (e.g. in tests that call NewApp without Startup).
func (k *KnowledgeAPI) GetSnapshot() common.SystemKnowledge {
	defer common.RecoverPanic()
	km := common.GetKnowledge()
	if km == nil {
		return common.SystemKnowledge{}
	}
	return km.GetSnapshot()
}
