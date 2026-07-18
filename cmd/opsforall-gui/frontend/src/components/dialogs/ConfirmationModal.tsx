import * as Dialog from '@radix-ui/react-dialog'
import { AlertTriangle, XCircle, ShieldCheck } from 'lucide-react'
import type { ActionPreview } from '@/types'

interface ConfirmationModalProps {
  preview: ActionPreview | null
  onConfirm: () => void
  onCancel: () => void
}

export function ConfirmationModal({ preview, onConfirm, onCancel }: ConfirmationModalProps) {
  if (!preview) return null

  return (
    <Dialog.Root open={!!preview}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-[250] bg-black/80 backdrop-blur-md animate-in fade-in duration-300" />
        <Dialog.Content className="fixed left-1/2 top-1/2 z-[300] -translate-x-1/2 -translate-y-1/2 bg-panel border-2 border-danger/20 rounded-[2rem] p-10 w-full max-w-xl shadow-[0_0_100px_rgba(255,50,50,0.1)] animate-in zoom-in-95 duration-300">
          <div className="flex items-center gap-4 mb-8">
            <div className="w-14 h-14 rounded-2xl bg-danger/10 flex items-center justify-center text-danger border border-danger/20">
              <AlertTriangle size={32} />
            </div>
            <div>
              <h2 className="text-2xl font-black text-text uppercase tracking-tight">Impact Warning</h2>
              <p className="text-text-dim font-bold uppercase tracking-widest text-xs mt-1">Execution Safety Required</p>
            </div>
          </div>

          <div className="space-y-6">
            <div>
              <p className="text-[10px] font-black text-text-faint uppercase tracking-widest mb-2">Requested Action</p>
              <p className="text-lg font-bold text-text">{preview.description}</p>
            </div>

            <div className="p-6 rounded-2xl bg-danger/5 border border-danger/10 space-y-3">
              <p className="text-[10px] font-black text-danger uppercase tracking-[0.2em]">Operational Risks</p>
              <ul className="space-y-2">
                {preview.risks.map((risk, i) => (
                  <li key={i} className="flex items-start gap-3 text-sm font-bold text-text-dim leading-snug">
                    <XCircle size={14} className="text-danger mt-0.5 shrink-0" />
                    {risk}
                  </li>
                ))}
              </ul>
            </div>

            <div className="p-4 rounded-xl bg-panel-3 border border-border">
              <p className="text-[10px] font-black text-text-faint uppercase tracking-widest mb-1">Rollback Policy</p>
              <p className="text-xs font-bold text-text-dim italic">{preview.rollback}</p>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4 mt-10">
            <button onClick={onCancel} className="py-4 rounded-xl border border-border text-text-dim font-black uppercase tracking-widest hover:bg-panel-3 transition-all active:scale-95">
              Abort
            </button>
            <button onClick={onConfirm} className="py-4 rounded-xl bg-danger text-white font-black uppercase tracking-widest shadow-xl shadow-danger/20 hover:bg-danger/90 transition-all active:scale-95 flex items-center justify-center gap-2">
              <ShieldCheck size={18} />
              Authorize
            </button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  )
}
