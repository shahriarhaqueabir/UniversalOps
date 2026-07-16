import { cn } from '@/lib/utils'

interface MiniStatProps {
  label: string
  value: string | number
  icon?: React.ReactNode
  unit?: string
  variant?: 'default' | 'success' | 'danger' | 'warning'
  className?: string
}

const variantClasses = {
  default: 'bg-[var(--color-accent)]/10 border-[var(--color-accent)]/30 text-[var(--color-accent)]',
  success: 'bg-[var(--color-success)]/10 border-[var(--color-success)]/30 text-[var(--color-success)]',
  danger: 'bg-[var(--color-danger)]/10 border-[var(--color-danger)]/30 text-[var(--color-danger)]',
  warning: 'bg-[var(--color-warning)]/10 border-[var(--color-warning)]/30 text-[var(--color-warning)]',
}

export function MiniStat({ label, value, icon, unit, variant = 'default', className }: MiniStatProps) {
  return (
    <div className={cn('bg-[var(--color-panel)] border border-[var(--color-border)] rounded-2xl p-6 flex items-center gap-6 shadow-lg transition-all hover:scale-105 active:scale-95 cursor-default group', className)}>
      <div className={cn('w-14 h-14 rounded-2xl flex items-center justify-center border shadow-inner group-hover:scale-110 transition-all', variantClasses[variant])}>
        {icon}
      </div>
      <div>
        <p className="text-sm font-bold text-[var(--color-text-faint)] uppercase tracking-widest mb-1">{label}</p>
        <p className="text-2xl font-bold text-[var(--color-text)] tabular-nums leading-none">
          {value}{unit && <span className="text-base text-[var(--color-text-faint)] ml-1 font-medium">{unit}</span>}
        </p>
      </div>
    </div>
  )
}
