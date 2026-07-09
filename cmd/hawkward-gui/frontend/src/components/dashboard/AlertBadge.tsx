import { useState } from 'react'
import { Bell } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { AlertInfo } from '@/types'

interface AlertBadgeProps {
  alerts: AlertInfo[]
}

const severityConfig = {
  critical: { bg: 'bg-danger/20', text: 'text-danger', border: 'border-danger/30', label: 'Critical' },
  warning: { bg: 'bg-warning/20', text: 'text-warning', border: 'border-warning/30', label: 'Warning' },
  info: { bg: 'bg-accent/20', text: 'text-accent', border: 'border-accent/30', label: 'Info' },
} as const

export function AlertBadge({ alerts }: AlertBadgeProps) {
  const [expanded, setExpanded] = useState(false)
  const activeAlerts = alerts.filter((a) => !a.resolved)
  const criticalCount = activeAlerts.filter((a) => a.level === 'critical').length

  if (alerts.length === 0) {
    return (
      <div className="bg-panel border border-border rounded-lg p-4">
        <div className="flex items-center gap-2 text-text-faint">
          <Bell size={16} />
          <span className="text-sm">No active alerts</span>
        </div>
      </div>
    )
  }

  return (
    <div className="bg-panel border border-border rounded-lg">
      {/* Header */}
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full flex items-center justify-between p-4 hover:bg-sidebar-hover rounded-lg transition-colors"
      >
        <div className="flex items-center gap-2">
          <Bell
            size={16}
            className={cn(criticalCount > 0 ? 'text-danger animate-pulse' : 'text-text-faint')}
          />
          <span className="text-sm font-medium text-text">
            {activeAlerts.length} Active Alert{activeAlerts.length !== 1 ? 's' : ''}
          </span>
        </div>
        <div className="flex items-center gap-2">
          {criticalCount > 0 && (
            <span className="px-1.5 py-0.5 text-xs font-bold bg-danger/20 text-danger rounded-full">
              {criticalCount} critical
            </span>
          )}
          <svg
            className={cn('w-4 h-4 text-text-faint transition-transform', expanded && 'rotate-180')}
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </div>
      </button>

      {/* Expanded list */}
      {expanded && (
        <div className="px-4 pb-4 space-y-2">
          {alerts.map((alert) => {
            const config = severityConfig[alert.level as keyof typeof severityConfig] ?? severityConfig.info
            return (
              <div
                key={alert.id}
                className={cn('flex items-start gap-3 p-3 rounded border', config.bg, config.border)}
              >
                <span className={cn('text-xs font-bold uppercase mt-0.5', config.text)}>
                  {config.label}
                </span>
                <div className="flex-1 min-w-0">
                  <p className="text-sm text-text">{alert.message}</p>
                  <p className="text-xs text-text-faint mt-0.5">
                    {alert.metric} · {alert.value.toFixed(1)} / {alert.threshold}
                  </p>
                </div>
                <span className="text-xs text-text-faint shrink-0">
                  {new Date(alert.timestamp).toLocaleTimeString()}
                </span>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
