import { useEffect, useState } from 'react'
import { cn } from '@/lib/utils'
import { MiniSparkline } from '@/components/charts/MiniSparkline'
import type { HealthStatus, TrendDirection, TimeSeriesPoint } from '@/types'

interface HealthCardProps {
  icon: React.ReactNode
  label: string
  value: string | number
  unit?: string
  status: HealthStatus
  trend?: TrendDirection
  onClick?: () => void
  sparklineData?: TimeSeriesPoint[]
}

const statusColors: Record<HealthStatus, string> = {
  healthy: 'bg-success',
  warning: 'bg-warning',
  critical: 'bg-danger',
}

const statusGlow: Record<HealthStatus, string> = {
  healthy: 'shadow-[0_0_8px_rgba(74,222,128,0.3)]',
  warning: 'shadow-[0_0_8px_rgba(251,191,36,0.3)]',
  critical: 'shadow-[0_0_8px_rgba(248,113,113,0.3)]',
}

export function HealthCard({
  icon,
  label,
  value,
  unit,
  status,
  onClick,
  sparklineData,
}: HealthCardProps) {
  const [animate, setAnimate] = useState(false)

  useEffect(() => {
    setAnimate(true)
    const t = setTimeout(() => setAnimate(false), 600)
    return () => clearTimeout(t)
  }, [value])

  return (
    <button
      onClick={onClick}
      className={cn(
        'relative bg-panel border border-border rounded-lg p-4 text-left transition-all duration-200',
        'hover:border-accent/40 hover:shadow-[0_0_12px_rgba(124,108,255,0.08)]',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent/50',
        onClick ? 'cursor-pointer' : 'cursor-default',
      )}
    >
      {/* Top row: icon, label, status dot */}
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <span className="text-accent">{icon}</span>
          <span className="text-xs font-medium text-text-faint uppercase tracking-wider">{label}</span>
        </div>
        <span
          className={cn('inline-block w-2 h-2 rounded-full', statusColors[status], statusGlow[status])}
        />
      </div>

      {/* Value */}
      <div className="flex items-baseline gap-1">
        <span
          className={cn(
            'text-2xl font-bold text-text transition-all duration-500',
            animate && 'scale-110 text-accent',
          )}
        >
          {value}
        </span>
        {unit && <span className="text-xs text-text-faint">{unit}</span>}
      </div>

      {/* Sparkline */}
      {sparklineData && sparklineData.length > 0 && (
        <div className="mt-2">
          <MiniSparkline data={sparklineData} height={32} />
        </div>
      )}
    </button>
  )
}
