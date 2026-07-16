import { Info } from 'lucide-react'

interface SectionBriefingProps {
  title: string
  objective: string
  checklist: string[]
  className?: string
}

export function SectionBriefing({ title, objective, checklist, className = '' }: SectionBriefingProps) {
  return (
    <div className={`bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-8 shadow-xl mb-8 ${className}`}>
      <div className="flex items-center gap-4 mb-4">
        <Info size={24} className="text-[var(--color-accent)]" />
        <h3 className="text-2xl font-bold text-[var(--color-text)] uppercase tracking-widest">{title}</h3>
      </div>
      <p className="text-lg text-[var(--color-text-dim)] leading-relaxed mb-6 italic">{objective}</p>
      <div className="grid grid-cols-2 gap-4">
        {checklist.map((item, i) => (
          <div key={i} className="flex items-center gap-3">
            <div className="w-1.5 h-1.5 rounded-full bg-[var(--color-accent)] shadow-[0_0_6px_var(--color-accent)]" />
            <span className="text-sm font-bold text-[var(--color-text-faint)]">{item}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
