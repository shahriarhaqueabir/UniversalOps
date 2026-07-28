// ── Health types ──

export type HealthStatus = 'healthy' | 'degraded' | 'critical' | 'unknown'

export interface SystemHealthMetric {
  name: string
  value: number
  unit: string
  status: HealthStatus
  message: string
}

export interface CollectorHealth {
  id: string
  name: string
  enabled: boolean
  status: 'healthy' | 'degraded' | 'critical' | 'stale'
  last_run: string
  interval_ms: number
  error?: string
}

export interface HealthSummary {
  overall: HealthStatus
  uptime: string
  cpu: SystemHealthMetric
  memory: SystemHealthMetric
  disk: SystemHealthMetric
  load: SystemHealthMetric
  alerts: number
  collectors: CollectorHealth[]
}
