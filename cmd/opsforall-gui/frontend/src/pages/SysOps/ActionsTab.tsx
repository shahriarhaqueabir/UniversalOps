import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { ConfirmDialog } from '@/components/dialogs/ConfirmDialog'
import { CommandOutput } from '@/components/sysops/CommandOutput'
import type { ActionResult } from '@/types'
import { cn } from '@/lib/utils'
import { Power, RotateCcw, Moon, Snowflake, Wifi, Trash2, Download, RefreshCw } from 'lucide-react'

interface ActionDef {
  id: string
  label: string
  icon: React.ReactNode
  color: string
  bg: string
  border: string
  confirmMessage: string
  danger?: boolean
}

const actions: ActionDef[] = [
  { id: 'reboot', label: 'Reboot', icon: <RotateCcw size={20} />, color: 'text-[var(--color-warning)]', bg: 'bg-[var(--color-warning)]/10', border: 'border-[var(--color-warning)]/30', confirmMessage: 'Are you sure you want to reboot the system?' },
  { id: 'shutdown', label: 'Shutdown', icon: <Power size={20} />, color: 'text-[var(--color-danger)]', bg: 'bg-[var(--color-danger)]/10', border: 'border-[var(--color-danger)]/30', confirmMessage: 'Are you sure you want to shutdown the system?', danger: true },
  { id: 'sleep', label: 'Sleep', icon: <Moon size={20} />, color: 'text-[var(--color-accent)]', bg: 'bg-[var(--color-accent)]/10', border: 'border-[var(--color-accent)]/30', confirmMessage: 'Put system to sleep?' },
  { id: 'hibernate', label: 'Hibernate', icon: <Snowflake size={20} />, color: 'text-[var(--color-accent)]', bg: 'bg-[var(--color-accent)]/10', border: 'border-[var(--color-accent)]/30', confirmMessage: 'Put system to hibernate?' },
  { id: 'flush_dns', label: 'Flush DNS', icon: <Wifi size={20} />, color: 'text-[var(--color-success)]', bg: 'bg-[var(--color-success)]/10', border: 'border-[var(--color-success)]/30', confirmMessage: 'Flush DNS cache?' },
  { id: 'clear_temp', label: 'Clear Temp Files', icon: <Trash2 size={20} />, color: 'text-[var(--color-warning)]', bg: 'bg-[var(--color-warning)]/10', border: 'border-[var(--color-warning)]/30', confirmMessage: 'Clear temporary files?' },
  { id: 'clean_pkg_cache', label: 'Clean Package Cache', icon: <Download size={20} />, color: 'text-[var(--color-warning)]', bg: 'bg-[var(--color-warning)]/10', border: 'border-[var(--color-warning)]/30', confirmMessage: 'Clean package manager cache?' },
  { id: 'system_update', label: 'System Update', icon: <RefreshCw size={20} />, color: 'text-[var(--color-accent)]', bg: 'bg-[var(--color-accent)]/10', border: 'border-[var(--color-accent)]/30', confirmMessage: 'Run system update? This may take a while.' },
]

export function ActionsTab() {
  const { call } = useBackend()
  const [confirmTarget, setConfirmTarget] = useState<ActionDef | null>(null)
  const [result, setResult] = useState<ActionResult | null>(null)

  const executeMutation = useMutation({
    mutationFn: async (action: string) => {
      const r = await call('SysOps.RunSystemAction', action)
      return r as ActionResult
    },
    onSuccess: (data) => setResult(data),
  })

  const handleConfirm = () => {
    if (confirmTarget) {
      executeMutation.mutate(confirmTarget.id)
      setConfirmTarget(null)
    }
  }

  return (
    <div className="space-y-6">
      <ConfirmDialog
        open={confirmTarget !== null}
        title={`Execute ${confirmTarget?.label}`}
        description={confirmTarget?.confirmMessage || ''}
        type={confirmTarget?.danger ? 'danger' : 'warning'}
        confirmText="Execute"
        onConfirm={handleConfirm}
        onClose={() => setConfirmTarget(null)}
      />

      <div className="grid grid-cols-4 gap-4">
        {actions.map(action => (
          <button
            key={action.id}
            onClick={() => { setResult(null); setConfirmTarget(action) }}
            className={cn('flex flex-col items-center gap-3 p-6 rounded-xl border transition-all hover:scale-[1.02]', action.bg, action.border)}
          >
            <div className={action.color}>{action.icon}</div>
            <span className="text-sm font-bold text-[var(--color-text)]">{action.label}</span>
          </button>
        ))}
      </div>

      {result && (
        <div className={cn('p-5 rounded-xl border', result.success ? 'bg-[var(--color-success)]/10 border-[var(--color-success)]/30' : 'bg-[var(--color-danger)]/10 border-[var(--color-danger)]/30')}>
          <p className={cn('text-sm font-bold', result.success ? 'text-[var(--color-success)]' : 'text-[var(--color-danger)]')}>{result.message}</p>
          {result.output && <CommandOutput output={result.output} title="Command Output" />}
        </div>
      )}

      {executeMutation.isPending && (
        <div className="text-center py-8">
          <div className="animate-spin w-8 h-8 border-2 border-[var(--color-accent)] border-t-transparent rounded-full mx-auto" />
          <p className="text-sm text-[var(--color-text-faint)] mt-3">Executing action...</p>
        </div>
      )}
    </div>
  )
}
