import { cn } from '@/lib/utils'

interface StatusBadgeProps {
  status: string
  size?: 'sm' | 'md'
  className?: string
}

const colorMap: Record<string, string> = {
  // Success variants
  success: 'bg-[var(--color-success)]/15 text-[var(--color-success)] border-[var(--color-success)]/30',
  running: 'bg-[var(--color-success)]/15 text-[var(--color-success)] border-[var(--color-success)]/30',
  active: 'bg-[var(--color-success)]/15 text-[var(--color-success)] border-[var(--color-success)]/30',
  enabled: 'bg-[var(--color-success)]/15 text-[var(--color-success)] border-[var(--color-success)]/30',
  auto: 'bg-[var(--color-success)]/15 text-[var(--color-success)] border-[var(--color-success)]/30',
  passed: 'bg-[var(--color-success)]/15 text-[var(--color-success)] border-[var(--color-success)]/30',
  listening: 'bg-[var(--color-success)]/15 text-[var(--color-success)] border-[var(--color-success)]/30',
  established: 'bg-[var(--color-accent)]/15 text-[var(--color-accent)] border-[var(--color-accent)]/30',
  allow: 'bg-[var(--color-success)]/15 text-[var(--color-success)] border-[var(--color-success)]/30',
  // Danger variants
  danger: 'bg-[var(--color-danger)]/15 text-[var(--color-danger)] border-[var(--color-danger)]/30',
  stopped: 'bg-[var(--color-danger)]/15 text-[var(--color-danger)] border-[var(--color-danger)]/30',
  failed: 'bg-[var(--color-danger)]/15 text-[var(--color-danger)] border-[var(--color-danger)]/30',
  error: 'bg-[var(--color-danger)]/15 text-[var(--color-danger)] border-[var(--color-danger)]/30',
  critical: 'bg-[var(--color-danger)]/15 text-[var(--color-danger)] border-[var(--color-danger)]/30',
  block: 'bg-[var(--color-danger)]/15 text-[var(--color-danger)] border-[var(--color-danger)]/30',
  disabled: 'bg-[var(--color-text-faint)]/20 text-[var(--color-text-faint)] border-[var(--color-border)]',
  timeout: 'bg-[var(--color-danger)]/15 text-[var(--color-danger)] border-[var(--color-danger)]/30',
  // Warning variants
  warning: 'bg-[var(--color-warning)]/15 text-[var(--color-warning)] border-[var(--color-warning)]/30',
  manual: 'bg-[var(--color-warning)]/15 text-[var(--color-warning)] border-[var(--color-warning)]/30',
  high: 'bg-[var(--color-warning)]/15 text-[var(--color-warning)] border-[var(--color-warning)]/30',
  medium: 'bg-[var(--color-warning)]/15 text-[var(--color-warning)] border-[var(--color-warning)]/30',
  // Info/accent variants
  info: 'bg-[var(--color-accent)]/15 text-[var(--color-accent)] border-[var(--color-accent)]/30',
  low: 'bg-[var(--color-accent)]/15 text-[var(--color-accent)] border-[var(--color-accent)]/30',
  inactive: 'bg-[var(--color-text-faint)]/20 text-[var(--color-text-faint)] border-[var(--color-border)]',
}

const sizeClasses = {
  sm: 'px-2 py-0.5 text-[10px]',
  md: 'px-3 py-1 text-xs',
}

export function StatusBadge({ status, size = 'md', className }: StatusBadgeProps) {
  const s = status.toLowerCase()
  return (
    <span
      className={cn(
        'inline-block font-black uppercase tracking-[0.2em] rounded-full border shadow-sm transition-all duration-300',
        sizeClasses[size],
        colorMap[s] || 'bg-[var(--color-text-faint)]/20 text-[var(--color-text-faint)] border-[var(--color-border)]',
        className
      )}
    >
      {status.replace('_', ' ')}
    </span>
  )
}
