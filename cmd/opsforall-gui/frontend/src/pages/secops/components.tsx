import { cn } from '@/lib/utils'
import { Info, Shield } from 'lucide-react'

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

export function MiniStat({ label, value, icon, variant }: { label: string; value: string | number; icon?: React.ReactNode; variant?: 'default' | 'success' | 'danger' | 'warning' }) {
  const variantClasses = {
    default: 'bg-accent/10 border-accent/30 text-accent',
    success: 'bg-success/10 border-success/30 text-success',
    danger: 'bg-danger/10 border-danger/30 text-danger',
    warning: 'bg-warning/10 border-warning/30 text-warning',
  }
  return (
    <div className="bg-panel border border-border rounded-2xl p-6 flex items-center gap-6 shadow-lg transition-all hover:scale-105 active:scale-95 cursor-default group">
      <div className={cn('w-14 h-14 rounded-2xl flex items-center justify-center border shadow-inner group-hover:scale-110 transition-all', variantClasses[variant || 'default'])}>
        {icon}
      </div>
      <div>
        <p className="text-sm font-bold text-text-faint uppercase tracking-widest mb-1">{label}</p>
        <p className="text-2xl font-bold text-text tabular-nums leading-none">{value}</p>
      </div>
    </div>
  )
}

export function StatusBadge({ status }: { status: string }) {
  const colorMap: Record<string, string> = {
    allow: 'bg-success/20 text-success border-success/30',
    block: 'bg-danger/20 text-danger border-danger/30',
    enabled: 'bg-success/20 text-success border-success/30',
    disabled: 'bg-text-faint/20 text-text-faint border-border',
    listening: 'bg-accent/15 text-accent border-accent/30',
    error: 'bg-danger/20 text-danger border-danger/30',
    warning: 'bg-warning/20 text-warning border-warning/30',
    running: 'bg-success/20 text-success border-success/30',
    active: 'bg-success/20 text-success border-success/30',
    inactive: 'bg-text-faint/20 text-text-faint border-border',
    passed: 'bg-success/20 text-success border-success/30',
    failed: 'bg-danger/20 text-danger border-danger/30',
    critical: 'bg-danger/20 text-danger border-danger/30',
    high: 'bg-orange-500/20 text-orange-400 border-orange-500/30',
    medium: 'bg-warning/20 text-warning border-warning/30',
    low: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
    info: 'bg-accent/15 text-accent border-accent/30',
  }
  const s = status.toLowerCase()
  return (
    <span className={cn('inline-block px-3 py-1 text-xs font-bold uppercase tracking-widest rounded-full border shadow-sm', colorMap[s] || 'bg-text-faint/20 text-text-faint border-border')}>
      {status}
    </span>
  )
}

export function ExpertInsight({ title, content }: { title: string; content: string }) {
  return (
    <div className="bg-accent-soft border border-accent/20 rounded-2xl p-8 flex gap-8 animate-in slide-in-from-left-4 duration-500">
      <div className="w-14 h-14 rounded-2xl bg-accent flex items-center justify-center shrink-0 shadow-xl">
        <Shield size={32} className="text-white" />
      </div>
      <div>
        <h4 className="text-2xl font-bold text-text mb-2">{title}</h4>
        <p className="text-text-dim text-xl leading-relaxed">{content}</p>
      </div>
    </div>
  )
}
