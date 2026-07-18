import { create } from 'zustand'
import type { DashboardData } from '@/types'

interface MetricsState {
  latest: DashboardData | null
  history: {
    cpu: { time: string; value: number }[]
    memory: { time: string; value: number }[]
    disk: { time: string; value: number }[]
  }
  setMetrics: (data: DashboardData) => void
}

export const useMetricsStore = create<MetricsState>((set) => ({
  latest: null,
  history: {
    cpu: [],
    memory: [],
    disk: [],
  },
  setMetrics: (data) => {
    const time = new Date().toLocaleTimeString('en-US', { hour12: false })
    set((state) => ({
      latest: data,
      history: {
        cpu: [...state.history.cpu.slice(-59), { time, value: data.cpu.value }],
        memory: [...state.history.memory.slice(-59), { time, value: data.memory.value }],
        disk: [...state.history.disk.slice(-59), { time, value: data.disk.value }],
      }
    }))
  }
}))
