import { cn } from '@/lib/utils'

interface StatCardProps {
  label: string
  value: string | number
  icon?: React.ReactNode
  valueClassName?: string
  className?: string
}

export function StatCard({ label, value, icon, valueClassName, className }: StatCardProps) {
  return (
    <div className={cn('bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-xl p-4', className)}>
      <div className="flex items-center gap-2 mb-1">
        {icon && <span className="text-[var(--color-text-faint)]">{icon}</span>}
        <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase tracking-wider">{label}</p>
      </div>
      <p className={cn('text-xl font-bold text-[var(--color-text)] tabular-nums', valueClassName)}>{value}</p>
    </div>
  )
}
