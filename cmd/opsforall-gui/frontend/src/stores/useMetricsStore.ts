import { create } from 'zustand'
import type { DashboardData, AlertInfo, HealthScorePoint } from '@/types'

interface TimelineEvent {
  id: string
  timestamp: string
  category: string
  level: string
  title: string
  detail: string
  module: string
  related?: string[]
  metadata?: Record<string, string>
}

interface MetricsState {
  latest: DashboardData | null
  alerts: AlertInfo[]
  timeline: TimelineEvent[]
  healthTrend: HealthScorePoint[]
  history: {
    cpu: { time: string; value: number }[]
    memory: { time: string; value: number }[]
    disk: { time: string; value: number }[]
  }
  setMetrics: (data: DashboardData) => void
  setSnapshot: (snapshot: { metrics: DashboardData; alerts: AlertInfo[]; timeline: TimelineEvent[] }) => void
}

export const useMetricsStore = create<MetricsState>((set) => ({
  latest: null,
  alerts: [],
  timeline: [],
  healthTrend: [],
  history: {
    cpu: [],
    memory: [],
    disk: [],
  },
  setMetrics: (data) => {
    const time = new Date().toLocaleTimeString('en-US', { hour12: false })
    set((state) => ({
      latest: data,
      healthTrend: data.health_trend ?? [],
      history: {
        cpu: [...state.history.cpu.slice(-59), { time, value: data.cpu.value }],
        memory: [...state.history.memory.slice(-59), { time, value: data.memory.value }],
        disk: [...state.history.disk.slice(-59), { time, value: data.disk.value }],
      }
    }))
  },
  setSnapshot: (snap) => {
    const time = new Date().toLocaleTimeString('en-US', { hour12: false })
    set((state) => ({
      latest: snap.metrics,
      alerts: snap.alerts ?? [],
      timeline: snap.timeline ?? [],
      healthTrend: snap.metrics.health_trend ?? [],
      history: {
        cpu: [...state.history.cpu.slice(-59), { time, value: snap.metrics.cpu.value }],
        memory: [...state.history.memory.slice(-59), { time, value: snap.metrics.memory.value }],
        disk: [...state.history.disk.slice(-59), { time, value: snap.metrics.disk.value }],
      }
    }))
  }
}))
