interface SectionHeaderProps {
  icon: React.ReactNode
  title: string
  count?: number
  className?: string
}

export function SectionHeader({ icon, title, count, className = '' }: SectionHeaderProps) {
  return (
    <div className={`flex items-center gap-3 mb-4 ${className}`}>
      <span className="text-[var(--color-accent)]">{icon}</span>
      <h3 className="text-lg font-bold text-[var(--color-text)] uppercase tracking-widest">{title}</h3>
      {count !== undefined && (
        <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-[var(--color-accent)]/15 text-[var(--color-accent)] border border-[var(--color-accent)]/30">
          {count}
        </span>
      )}
    </div>
  )
}
