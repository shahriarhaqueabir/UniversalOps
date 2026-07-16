import { cn } from '@/lib/utils'

interface ProgressBarProps {
  value: number
  max?: number
  variant?: 'default' | 'success' | 'warning' | 'danger'
  label?: string
  showValue?: boolean
  unit?: string
  className?: string
}

function getAutoVariant(pct: number): 'success' | 'warning' | 'danger' {
  if (pct >= 70) return 'danger'
  if (pct >= 25) return 'warning'
  return 'success'
}

const variantColors = {
  default: 'var(--color-accent)',
  success: 'var(--color-success)',
  warning: 'var(--color-warning)',
  danger: 'var(--color-danger)',
}

export function ProgressBar({ value, max = 100, variant, label, showValue = true, unit = '%', className }: ProgressBarProps) {
  const pct = Math.min((value / max) * 100, 100)
  const effectiveVariant = variant || getAutoVariant(pct)
  const barColor = variantColors[effectiveVariant]

  return (
    <div className={cn('space-y-1', className)}>
      {(label || showValue) && (
        <div className="flex items-center justify-between">
          {label && <span className="text-[var(--color-text-dim)] text-sm font-medium">{label}</span>}
          {showValue && (
            <span className="text-[var(--color-text)] font-bold text-sm tabular-nums">
              {value.toFixed(1)}{unit}
            </span>
          )}
        </div>
      )}
      <div className="h-3 bg-[var(--color-panel-3)] rounded-full overflow-hidden border border-[var(--color-border)]">
        <div
          className="h-full rounded-full transition-all duration-700"
          style={{ width: `${pct}%`, background: `linear-gradient(90deg, ${barColor}88, ${barColor})` }}
        />
      </div>
    </div>
  )
}
