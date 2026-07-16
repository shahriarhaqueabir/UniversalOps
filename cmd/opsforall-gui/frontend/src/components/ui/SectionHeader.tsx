import type { LucideIcon } from 'lucide-react'

interface SectionHeaderProps {
  icon: LucideIcon
  title: string
  count?: number
  className?: string
}

export function SectionHeader({ icon: Icon, title, count, className = '' }: SectionHeaderProps) {
  return (
    <div className={`flex items-center gap-3 mb-4 ${className}`}>
      <Icon size={20} className="text-[var(--color-accent)]" />
      <h3 className="text-lg font-bold text-[var(--color-text)] uppercase tracking-widest">{title}</h3>
      {count !== undefined && (
        <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-[var(--color-accent)]/15 text-[var(--color-accent)] border border-[var(--color-accent)]/30">
          {count}
        </span>
      )}
    </div>
  )
}
