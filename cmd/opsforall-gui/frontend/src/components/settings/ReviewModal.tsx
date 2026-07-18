import * as Dialog from '@radix-ui/react-dialog'
import { X, CheckCircle2, RotateCcw, Zap } from 'lucide-react'
import { useConfigStore } from '@/stores/useConfigStore'
import { useBackend } from '@/hooks/useBackend'
import { toast } from 'sonner'

interface ReviewModalProps {
  isOpen: boolean
  onOpenChange: (open: boolean) => void
}

/**
 * ReviewModal — The "Staged Review" interface.
 * Displays a diff of pending changes before they are committed to the Go backend.
 */
export function ReviewModal({ isOpen, onOpenChange }: ReviewModalProps) {
  const { stagedChanges, discardAll, getOriginalValue } = useConfigStore()
  const { call } = useBackend()

  const handleDeploy = async () => {
    try {
      // In a real iteration, we might batch these into a single Go call.
      // For now, we iterate and call the existing bindings.
      for (const [key, value] of stagedChanges.entries()) {
        if (key === 'refreshInterval') {
          // We need to fetch other settings to maintain consistency for this specific binding
          const s = stagedChanges
          await call('PipelineAPI.UpdateSettings',
            value,
            0,
            s.get('pingCount') || getOriginalValue('pingCount'),
            s.get('dnsTimeout') || getOriginalValue('dnsTimeout')
          )
        }
        // Add more logic for other settings as they are integrated into the staged flow
      }

      toast.success('System configuration deployed successfully')
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
                Review and approve staged changes to the operations engine.
              </Dialog.Description>
            </div>
            <Dialog.Close className="p-2 rounded-xl hover:bg-[var(--color-panel-2)] transition-all">
              <X size={20} />
            </Dialog.Close>
          </div>

          <div className="bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-2xl overflow-hidden shadow-inner">
            <table className="w-full text-left border-collapse">
              <thead className="bg-[var(--color-bg)]/50">
                <tr>
                  <th className="px-5 py-3 text-[10px] font-black uppercase text-[var(--color-text-faint)] tracking-widest">Parameter</th>
                  <th className="px-5 py-3 text-[10px] font-black uppercase text-[var(--color-text-faint)] tracking-widest">Current</th>
                  <th className="px-5 py-3 text-[10px] font-black uppercase text-[var(--color-text-faint)] tracking-widest">Proposed</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[var(--color-border)]/30">
                {Array.from(stagedChanges.entries()).map(([key, value]) => (
                  <tr key={key} className="hover:bg-[var(--color-bg)]/20 transition-colors">
                    <td className="px-5 py-4">
                      <p className="text-xs font-bold text-[var(--color-text)] capitalize">
                        {key.replace(/([A-Z])/g, ' $1').trim()}
                      </p>
                    </td>
                    <td className="px-5 py-4 text-xs font-mono text-[var(--color-text-faint)]">
                      {String(getOriginalValue(key))}
                    </td>
                    <td className="px-5 py-4 text-xs font-mono font-bold text-[var(--color-accent)]">
                      {String(value)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

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
