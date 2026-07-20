import { X, ShieldAlert, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'

interface ConfirmDialogProps {
  open: boolean
  title: string
  description: string
  confirmText?: string
  cancelText?: string
  type?: 'danger' | 'info' | 'warning'
  onConfirm: () => void
  onClose: () => void
}

export function ConfirmDialog({
  open,
  title,
  description,
  confirmText = 'Confirm Action',
  cancelText = 'Cancel',
  type = 'warning',
  onConfirm,
  onClose,
}: ConfirmDialogProps) {
  if (!open) return null

  const colors = {
    danger: 'text-danger bg-danger/10 border-danger/20',
    info: 'text-accent bg-accent/10 border-accent/20',
    warning: 'text-warning bg-warning/10 border-warning/20',
  }

  const buttonColors = {
    danger: 'bg-danger hover:bg-danger/90',
    info: 'bg-accent hover:bg-accent/90',
    warning: 'bg-warning hover:bg-warning/90 text-black',
  }

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/70 backdrop-blur-md">
      <div className="bg-panel border border-border rounded-2xl shadow-[0_0_50px_rgba(0,0,0,0.5)] w-full max-w-lg overflow-hidden animate-in zoom-in-95 duration-200">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-border bg-panel-2">
          <div className="flex items-center gap-4">
            <div className={cn("w-12 h-12 rounded-xl flex items-center justify-center border", colors[type])}>
              <ShieldAlert size={28} />
            </div>
            <h2 className="text-2xl font-bold text-text">{title}</h2>
          </div>
          <button
            onClick={onClose}
            className="p-2 rounded-lg hover:bg-sidebar-hover text-text-faint hover:text-text transition-colors"
          >
            <X size={24} />
          </button>
        </div>

        {/* Body */}
        <div className="p-8">
          <p className="text-text-dim text-lg leading-relaxed">
            {description}
          </p>

          <div className="mt-8 p-5 rounded-xl bg-panel-3 border border-border flex items-start gap-4">
            <div className="mt-1">
              <ShieldAlert size={20} className="text-warning" />
            </div>
            <p className="text-sm text-text-faint italic">
              Permissions Required: This action will modify workstation state or execute commands outside the default sandbox environment.
            </p>
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-4 p-6 border-t border-border bg-panel-2">
          <button
            onClick={onClose}
            className="px-6 py-3 text-lg font-bold text-text-dim hover:text-text transition-colors"
          >
            {cancelText}
          </button>
          <button
            onClick={() => {
              onConfirm()
              onClose()
            }}
            className={cn(
              "flex items-center gap-2 px-8 py-3 text-lg font-bold rounded-xl transition-all shadow-lg active:scale-95",
              buttonColors[type]
            )}
          >
            {confirmText}
            <ChevronRight size={20} />
          </button>
        </div>
      </div>
    </div>
  )
}
