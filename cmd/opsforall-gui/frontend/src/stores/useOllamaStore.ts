import { create } from 'zustand'
import type { OllamaStatus } from '@/types'

interface OllamaState {
  status: OllamaStatus
  setStatus: (status: OllamaStatus) => void
}

export const useOllamaStore = create<OllamaState>((set) => ({
  status: { available: false, binary_exists: false, model: '', version: '' },
  setStatus: (status) => set({ status }),
}))
