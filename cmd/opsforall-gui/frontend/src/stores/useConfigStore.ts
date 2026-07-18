import { create } from 'zustand'
import { useSettingsStore } from './useSettingsStore'

interface ConfigState {
  stagedChanges: Map<string, any>
  stageChange: (key: string, value: any) => void
  discardAll: () => void
  getOriginalValue: (key: string) => any
}

/**
 * useConfigStore — manages the "Shadow State" for configuration changes.
 * Allows users to stage modifications and review them before deployment.
 */
export interface StagedRisk {
  level: 'low' | 'med' | 'high'
  message: string
}

export const useConfigStore = create<ConfigState>((set, get) => ({
  stagedChanges: new Map(),

  stageChange: (key, value) => {
    const next = new Map(get().stagedChanges)
    next.set(key, value)
    set({ stagedChanges: next })
  },

  discardAll: () => {
    set({ stagedChanges: new Map() })
  },

  getOriginalValue: (key) => {
    const settings = useSettingsStore.getState()
    return (settings as any)[key]
  },

  getRiskLevel: (key: string, value: any): StagedRisk => {
    switch (key) {
      case 'refreshInterval':
        if (value < 1000) return { level: 'high', message: 'Sub-second refresh will significantly increase CPU & Disk I/O load.' }
        if (value < 3000) return { level: 'med', message: 'Frequent updates may cause minor UI jitter on lower-end hardware.' }
        return { level: 'low', message: 'Standard monitoring frequency.' }
      case 'pingCount':
        if (value > 15) return { level: 'med', message: 'High ping counts may be flagged as suspicious by some firewalls.' }
        return { level: 'low', message: 'Normal network diagnostic load.' }
      case 'dnsTimeout':
        if (value > 8000) return { level: 'med', message: 'High timeouts may slow down NetOps page responsiveness.' }
        return { level: 'low', message: 'Standard lookup timeout.' }
      case 'modelfile':
        return { level: 'med', message: 'Rebuilding the neural core may take 10-30 seconds depending on hardware.' }
      default:
        return { level: 'low', message: 'Safe configuration change.' }
    }
  },
}))
