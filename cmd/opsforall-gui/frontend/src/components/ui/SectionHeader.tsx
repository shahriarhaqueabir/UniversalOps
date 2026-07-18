interface SectionHeaderProps {
  icon: React.ReactNode
  title: string
  count?: number
  className?: string
}

export function SectionHeader({ icon, title, count, className = '' }: SectionHeaderProps) {
  return (
    <div className={`flex items-center gap-4 mb-8 ${className} group`}>
      <div className="w-10 h-10 rounded-xl bg-accent/10 border border-accent/20 flex items-center justify-center text-accent transition-transform group-hover:scale-110">
        {icon}
      </div>
      <h3 className="text-xl font-black text-text uppercase tracking-[0.2em]">{title}</h3>
      {count !== undefined && (
        <span className="text-xs font-black px-3 py-1 rounded-full bg-panel-3 text-accent border border-border tabular-nums shadow-inner">
          {count}
        </span>
      )}
    </div>
  )
}
