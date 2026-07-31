import { Info, Target, AlertTriangle, Activity, Zap } from 'lucide-react'

interface SectionBriefingProps {
  title: string
  objective: string
  checklist: string[]
  why?: string
  risks?: string[]
  typicalValues?: string
  recommendedActions?: string[]
  className?: string
}

export function SectionBriefing({
  title,
  objective,
  checklist,
  why,
  risks,
  typicalValues,
  recommendedActions,
  className = ''
}: SectionBriefingProps) {
  return (
    <div className={`bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-8 shadow-xl mb-8 ${className}`}>
      <div className="flex items-center justify-between gap-4 mb-6">
        <div className="flex items-center gap-4">
          <Info size={24} className="text-[var(--color-accent)]" />
          <h3 className="text-2xl font-black text-[var(--color-text)] uppercase tracking-[0.2em] italic">{title}</h3>
        </div>
      </div>

      <p className="text-lg text-[var(--color-text-dim)] leading-relaxed mb-8 italic border-l-4 border-[var(--color-accent)]/30 pl-6">{objective}</p>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-8">
        <div className="space-y-4">
          <h4 className="text-[10px] font-black uppercase tracking-widest text-[var(--color-text-faint)]">Checklist</h4>
          <div className="grid grid-cols-1 gap-3">
            {checklist.map((item) => (
              <div key={item} className="flex items-center gap-3">
                <div className="w-1.5 h-1.5 rounded-full bg-[var(--color-accent)] shadow-[0_0_6px_var(--color-accent)]" />
                <span className="text-sm font-bold text-[var(--color-text-dim)]">{item}</span>
              </div>
            ))}
          </div>
        </div>

        {why && (
          <div className="space-y-4">
            <h4 className="text-[10px] font-black uppercase tracking-widest text-accent flex items-center gap-2">
              <Target size={12} /> Why it matters
            </h4>
            <p className="text-xs text-[var(--color-text-dim)] leading-relaxed font-medium">{why}</p>
          </div>
        )}
      </div>

      {(risks || typicalValues || recommendedActions) && (
        <div className="pt-8 border-t border-[var(--color-border)]/50 grid grid-cols-1 md:grid-cols-3 gap-8">
          {risks && (
            <div className="space-y-3">
              <h4 className="text-[10px] font-black uppercase tracking-widest text-danger flex items-center gap-2">
                <AlertTriangle size={12} /> Risks
              </h4>
              <ul className="space-y-1.5">
                {risks.map((r) => (
                  <li key={r} className="text-[10px] text-[var(--color-text-dim)] font-medium">• {r}</li>
                ))}
              </ul>
            </div>
          )}

          {typicalValues && (
            <div className="space-y-3">
              <h4 className="text-[10px] font-black uppercase tracking-widest text-success flex items-center gap-2">
                <Activity size={12} /> Typical Values
              </h4>
              <p className="text-[10px] text-[var(--color-text-dim)] font-mono leading-relaxed">{typicalValues}</p>
            </div>
          )}

          {recommendedActions && (
            <div className="space-y-3">
              <h4 className="text-[10px] font-black uppercase tracking-widest text-warning flex items-center gap-2">
                <Zap size={12} /> Recommended
              </h4>
              <ul className="space-y-1.5">
                {recommendedActions.map((a) => (
                  <li key={a} className="text-[10px] text-[var(--color-text-dim)] font-medium">• {a}</li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

