import { cn } from '@/lib/utils'
import { CheckCircle, AlertTriangle, XCircle } from 'lucide-react'

interface HealthBadgeProps {
  status: 'pass' | 'warn' | 'fail'
  value: string
  label: string
}

const statusConfig = {
  pass: { icon: CheckCircle, color: 'text-[var(--color-success)]', bg: 'bg-[var(--color-success)]/20', border: 'border-[var(--color-success)]/30' },
  warn: { icon: AlertTriangle, color: 'text-[var(--color-warning)]', bg: 'bg-[var(--color-warning)]/20', border: 'border-[var(--color-warning)]/30' },
  fail: { icon: XCircle, color: 'text-[var(--color-danger)]', bg: 'bg-[var(--color-danger)]/20', border: 'border-[var(--color-danger)]/30' },
}

export function HealthBadge({ status, value, label }: HealthBadgeProps) {
  const config = statusConfig[status]
  const Icon = config.icon

  return (
    <div className={cn('flex items-center gap-3 px-4 py-2 rounded-xl border', config.bg, config.border)}>
      <Icon size={16} className={config.color} />
      <div className="flex flex-col">
        <span className={cn('text-sm font-bold tabular-nums', config.color)}>{value}</span>
        <span className="text-[10px] font-bold text-[var(--color-text-faint)] uppercase tracking-widest">{label}</span>
      </div>
    </div>
  )
}
