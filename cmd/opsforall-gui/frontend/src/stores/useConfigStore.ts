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
}))
