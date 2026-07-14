import { cn } from '@/lib/utils'
import { Info } from 'lucide-react'

export function SectionBriefing({ title, objective, checklist }: { title: string; objective: string; checklist: string[] }) {
  return (
    <div className="bg-panel-2 border border-border rounded-[var(--radius-lg)] p-8 shadow-xl mb-8">
      <div className="flex items-center gap-4 mb-4">
        <Info size={24} className="text-accent" />
        <h3 className="text-2xl font-bold text-text uppercase tracking-widest">{title}</h3>
      </div>
      <p className="text-lg text-text-dim leading-relaxed mb-6 italic">{objective}</p>
      <div className="grid grid-cols-2 gap-4">
        {checklist.map((item, i) => (
          <div key={i} className="flex items-center gap-3">
            <div className="w-1.5 h-1.5 rounded-full bg-accent shadow-[0_0_6px_var(--color-accent)]" />
            <span className="text-sm font-bold text-text-faint">{item}</span>
          </div>
        ))}
      </div>
    </div>
  )
}

export function MiniStat({ label, value, icon, unit }: { label: string; value: string | number; icon?: React.ReactNode; unit?: string }) {
  return (
    <div className="bg-panel border border-border rounded-2xl p-6 flex items-center gap-6 shadow-lg transition-all hover:scale-105 active:scale-95 cursor-default group">
      <div className="w-14 h-14 rounded-2xl bg-panel-3 flex items-center justify-center text-accent border border-border shadow-inner group-hover:bg-accent-soft group-hover:text-white transition-all">
        {icon}
      </div>
      <div>
        <p className="text-sm font-bold text-text-faint uppercase tracking-widest mb-1">{label}</p>
        <p className="text-2xl font-bold text-text tabular-nums leading-none">
          {value}{unit && <span className="text-base text-text-faint ml-1 font-medium">{unit}</span>}
        </p>
      </div>
    </div>
  )
}

export function StatusBadge({ status }: { status: string }) {
  const colorMap: Record<string, string> = {
    success: 'bg-success/15 text-success border-success/30',
    timeout: 'bg-danger/15 text-danger border-danger/30',
    listening: 'bg-success/15 text-success border-success/30',
    established: 'bg-accent/15 text-accent border-accent/30',
  }
  return (
    <span className={cn('inline-block px-3 py-1 text-xs font-bold uppercase tracking-widest rounded-full border shadow-sm', colorMap[status.toLowerCase()] || 'bg-text-faint/20 text-text-faint border-border')}>
      {status.replace('_', ' ')}
    </span>
  )
}
