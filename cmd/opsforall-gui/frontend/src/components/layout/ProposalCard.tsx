import { useConfigStore } from '@/stores/useConfigStore'
import { Check, Rocket, Info } from 'lucide-react'
import { toast } from 'sonner'
import { useState } from 'react'

interface ProposalCardProps {
  reasoning: string
  payload: Record<string, any>
}

/**
 * ProposalCard — An interactive AI optimization card.
 * Allows users to review and batch-stage Hawk's suggested engine improvements.
 */
export function ProposalCard({ reasoning, payload }: ProposalCardProps) {
  const { stageBatch, getOriginalValue } = useConfigStore()
  const [staged, setStaged] = useState(false)

  const handleStage = () => {
    stageBatch(payload)
    setStaged(true)
    toast.success('AI recommendations staged for review')
  }

  const entries = Object.entries(payload)

  return (
    <div className="bg-[var(--color-panel-2)] border border-[var(--color-accent)]/30 rounded-2xl overflow-hidden shadow-lg animate-in slide-in-from-left-4 duration-300">
      <div className="p-4 bg-[var(--color-accent)]/5 border-b border-[var(--color-accent)]/10">
        <div className="flex items-center gap-2 mb-2">
          <Rocket size={14} className="text-[var(--color-accent)]" />
          <p className="text-[10px] font-black uppercase tracking-widest text-[var(--color-accent)]">Optimization Proposal</p>
        </div>
        <p className="text-xs leading-relaxed text-[var(--color-text)]">
          {reasoning}
        </p>
      </div>

      <div className="p-4 space-y-3 bg-[var(--color-panel-3)]/50">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="border-b border-[var(--color-border)]/50">
              <th className="pb-2 text-[9px] font-black uppercase text-[var(--color-text-faint)]">Parameter</th>
              <th className="pb-2 text-[9px] font-black uppercase text-[var(--color-text-faint)]">Proposal</th>
            </tr>
          </thead>
          <tbody>
            {entries.map(([key, val]) => (
              <tr key={key}>
                <td className="py-2 text-[10px] font-bold text-[var(--color-text-dim)] capitalize">
                  {key.replace(/([A-Z])/g, ' $1').trim()}
                </td>
                <td className="py-2 text-[10px] font-mono">
                  <div className="flex items-center gap-2">
                    <span className="opacity-40">{String(getOriginalValue(key))}</span>
                    <span className="text-[var(--color-accent)] font-bold">➔ {String(val)}</span>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {staged ? (
          <div className="flex items-center justify-center gap-2 py-2 px-4 rounded-xl bg-[var(--color-success)]/10 text-[var(--color-success)] border border-[var(--color-success)]/20">
            <Check size={14} />
            <span className="text-[10px] font-black uppercase tracking-widest">Changes Staged</span>
          </div>
        ) : (
          <button
            onClick={handleStage}
            className="w-full py-2.5 px-4 rounded-xl bg-[var(--color-accent)] text-white text-[10px] font-black uppercase tracking-widest hover:opacity-90 transition-all active:scale-95 shadow-md shadow-[var(--color-accent)]/20"
          >
            Stage AI Proposals
          </button>
        )}
      </div>

      <div className="px-4 py-2 bg-[var(--color-panel)] border-t border-[var(--color-border)] flex items-center gap-2">
        <Info size={10} className="text-[var(--color-text-faint)]" />
        <p className="text-[9px] text-[var(--color-text-faint)] italic">
          Proposals must be manually deployed from the Control Plane.
        </p>
      </div>
    </div>
  )
}
