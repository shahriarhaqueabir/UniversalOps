package app

import (
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// KnowledgeAPI exposes the unified System Knowledge Layer to the frontend.
type KnowledgeAPI struct {
	app *App
}

// NewKnowledgeAPI creates a new KnowledgeAPI facade.
func NewKnowledgeAPI(app *App) *KnowledgeAPI {
	return &KnowledgeAPI{app: app}
}

// GetSnapshot returns the current unified system state.
func (k *KnowledgeAPI) GetSnapshot() common.SystemKnowledge {
	return common.GetKnowledge().GetSnapshot()
}
