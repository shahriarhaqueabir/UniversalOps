import * as Dialog from '@radix-ui/react-dialog'
import { X, CheckCircle2, RotateCcw, Zap, ShieldAlert, Sparkles, AlertTriangle, FileCode } from 'lucide-react'
import { useConfigStore } from '@/stores/useConfigStore'
import { useBackend } from '@/hooks/useBackend'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'

interface ReviewModalProps {
  isOpen: boolean
  onOpenChange: (open: boolean) => void
}

/**
 * ReviewModal — The "Staged Review" interface.
 * Implements Capability, Risk, and Intent Gateways before deployment.
 */
export function ReviewModal({ isOpen, onOpenChange }: ReviewModalProps) {
  const { stagedChanges, discardAll, getOriginalValue, getRiskLevel } = useConfigStore()
  const { call } = useBackend()

  const changes = Array.from(stagedChanges.entries())
  const paramChanges = changes.filter(([key]) => key !== 'modelfile')
  const modelfileChange = stagedChanges.get('modelfile')

  const highestRisk = changes.reduce((prev, [key, val]) => {
    const r = getRiskLevel(key, val).level
    if (r === 'high' || prev === 'high') return 'high'
    if (r === 'med' || prev === 'med') return 'med'
    return 'low'
  }, 'low' as 'low' | 'med' | 'high')

  const handleDeploy = async () => {
    try {
      const id = toast.loading('Initiating system deployment...')

      // 1. Handle standard settings
      if (stagedChanges.has('refreshInterval') || stagedChanges.has('pingCount') || stagedChanges.has('dnsTimeout')) {
        await call('PipelineAPI.UpdateSettings',
          stagedChanges.get('refreshInterval') || getOriginalValue('refreshInterval'),
          0,
          stagedChanges.get('pingCount') || getOriginalValue('pingCount'),
          stagedChanges.get('dnsTimeout') || getOriginalValue('dnsTimeout')
        )
      }

      // 2. Handle Modelfile rebuild
      if (modelfileChange) {
        toast.loading('Rebuilding Neural Core... (may take a moment)', { id })
        await call('AIOps.SaveModelfile', modelfileChange)
        await call('AIOps.CreateOpsPersona')
      }

      toast.success('System configuration deployed successfully', { id })
      discardAll()
      onOpenChange(false)
    } catch (err) {
      toast.error('Deployment failed: ' + err)
    }
  }

  return (
    <Dialog.Root open={isOpen} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-[60] bg-black/60 backdrop-blur-sm animate-in fade-in" />
        <Dialog.Content className="fixed left-1/2 top-1/2 z-[70] -translate-x-1/2 -translate-y-1/2 bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[24px] p-8 w-full max-w-2xl shadow-2xl animate-in zoom-in-95 duration-200">
          <div className="flex items-center justify-between mb-6">
            <div>
              <Dialog.Title className="text-xl font-black text-[var(--color-text)] uppercase tracking-tight flex items-center gap-2">
                <Zap size={20} className="text-[var(--color-accent)]" /> Deployment Review
              </Dialog.Title>
              <Dialog.Description className="text-sm text-[var(--color-text-dim)] mt-1">
                Aura Gateway: Pre-flight check and risk assessment.
              </Dialog.Description>
            </div>
            <Dialog.Close className="p-2 rounded-xl hover:bg-[var(--color-panel-2)] transition-all">
              <X size={20} />
            </Dialog.Close>
          </div>

          {/* Hawk Pre-Flight Gate */}
          <div className="mb-6 p-5 rounded-2xl bg-[var(--color-panel-2)] border border-[var(--color-border)] flex items-start gap-4">
            <div className={cn(
              "w-10 h-10 rounded-xl flex items-center justify-center shrink-0 border shadow-inner",
              highestRisk === 'high' ? "bg-danger/10 text-danger border-danger/20" :
              highestRisk === 'med' ? "bg-warning/10 text-warning border-warning/20" :
              "bg-[var(--color-accent)]/10 text-[var(--color-accent)] border-[var(--color-accent)]/20"
            )}>
              {highestRisk === 'high' ? <ShieldAlert size={20} /> : <Sparkles size={20} />}
            </div>
            <div>
              <p className="text-[10px] font-black uppercase tracking-[0.2em] text-[var(--color-text-faint)]">
                Hawk Pre-Flight Status
              </p>
              <p className="text-sm font-bold text-[var(--color-text)] mt-1">
                {highestRisk === 'high' ? "Attention: Elevated Risk Deployment" :
                 highestRisk === 'med' ? "Warning: Standard Adjustments Pending" :
                 "System Nominal: Safe to Deploy"}
              </p>
              <p className="text-xs text-[var(--color-text-dim)] mt-1 leading-relaxed">
                Hawk has analyzed {changes.length} staged changes. No capability conflicts detected.
              </p>
            </div>
          </div>

          <div className="bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-2xl overflow-hidden shadow-inner">
            <table className="w-full text-left border-collapse">
              <thead className="bg-[var(--color-bg)]/50">
                <tr>
                  <th className="px-5 py-3 text-[10px] font-black uppercase text-[var(--color-text-faint)] tracking-widest">Parameter</th>
                  <th className="px-5 py-3 text-[10px] font-black uppercase text-[var(--color-text-faint)] tracking-widest">Diff</th>
                  <th className="px-5 py-3 text-[10px] font-black uppercase text-[var(--color-text-faint)] tracking-widest">Risk Note</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[var(--color-border)]/30">
                {paramChanges.map(([key, value]) => {
                  const risk = getRiskLevel(key, value)
                  return (
                    <tr key={key} className="hover:bg-[var(--color-bg)]/20 transition-colors">
                      <td className="px-5 py-4">
                        <p className="text-xs font-bold text-[var(--color-text)] capitalize">
                          {key.replace(/([A-Z])/g, ' $1').trim()}
                        </p>
                      </td>
                      <td className="px-5 py-4 text-xs font-mono">
                        <div className="flex items-center gap-2">
                          <span className="text-[var(--color-text-faint)]">{String(getOriginalValue(key))}</span>
                          <ChevronRight size={10} className="text-[var(--color-text-faint)]" />
                          <span className="font-bold text-[var(--color-accent)]">{String(value)}</span>
                        </div>
                      </td>
                      <td className="px-5 py-4">
                        <div className="flex items-start gap-2">
                          {risk.level !== 'low' && <AlertTriangle size={12} className={risk.level === 'high' ? 'text-danger' : 'text-warning'} />}
                          <p className="text-[10px] text-[var(--color-text-dim)] leading-tight">{risk.message}</p>
                        </div>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>

          {/* Neural Core Instruction Diff */}
          {modelfileChange && (
            <div className="mt-4 p-5 rounded-2xl bg-violet-400/5 border border-violet-400/20 shadow-inner">
              <div className="flex items-center gap-2 mb-3">
                <FileCode size={14} className="text-violet-400" />
                <p className="text-[10px] font-black uppercase tracking-widest text-violet-400">Neural Core Instructions Updated</p>
              </div>
              <div className="max-h-32 overflow-y-auto bg-black/20 rounded-xl p-4 border border-violet-400/10">
                <pre className="text-[9px] font-mono text-violet-200/70 whitespace-pre-wrap leading-relaxed">
                  {modelfileChange.split('\n').slice(0, 10).join('\n')}
                  {modelfileChange.split('\n').length > 10 && '\n... [truncated]'}
                </pre>
              </div>
              <p className="mt-2 text-[9px] text-[var(--color-text-dim)] italic">
                * Deployment will trigger a rebuild of the 'opsforall' model in Ollama.
              </p>
            </div>
          )}

          <div className="mt-8 flex items-center justify-between gap-4">
            <button
              onClick={() => { discardAll(); onOpenChange(false); toast.info('All changes discarded') }}
              className="px-6 py-3 rounded-xl bg-danger/10 text-danger text-xs font-black uppercase tracking-widest hover:bg-danger/20 transition-all border border-danger/20"
            >
              <RotateCcw size={14} className="inline mr-2" />
              Discard All
            </button>
            <button
              onClick={handleDeploy}
              className="flex-1 py-3 rounded-xl bg-[var(--color-accent)] text-white text-xs font-black uppercase tracking-widest hover:opacity-90 shadow-xl shadow-[var(--color-accent)]/20 transition-all active:scale-[0.98]"
            >
              <CheckCircle2 size={14} className="inline mr-2" />
              Deploy to Engine
            </button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  )
}

function ChevronRight({ size, className }: { size: number, className?: string }) {
  return (
    <svg
      width={size} height={size} viewBox="0 0 24 24" fill="none"
      stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"
      className={className}
    >
      <path d="m9 18 6-6-6-6"/>
    </svg>
  )
}
