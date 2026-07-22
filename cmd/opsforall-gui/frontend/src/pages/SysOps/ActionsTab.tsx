import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { ConfirmationModal } from '@/components/dialogs/ConfirmationModal'
import { CommandOutput } from '@/components/sysops/CommandOutput'
import type { ActionPreview } from '@/types'
import { cn } from '@/lib/utils'
import { Power, RotateCcw, Moon, Snowflake, Wifi, Trash2, Download, RefreshCw, Zap } from 'lucide-react'
import { toast } from 'sonner'

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
  { id: 'disk_cleanup', label: 'Disk Cleanup', icon: <Trash2 size={20} />, color: 'text-[var(--color-warning)]', bg: 'bg-[var(--color-warning)]/10', border: 'border-[var(--color-warning)]/30', confirmMessage: 'Run Windows Disk Cleanup?' },
  { id: 'defrag', label: 'Defragment', icon: <Zap size={20} />, color: 'text-[var(--color-accent)]', bg: 'bg-[var(--color-accent)]/10', border: 'border-[var(--color-accent)]/30', confirmMessage: 'Optimize and defragment drive C:?' },
  { id: 'clean_pkg_cache', label: 'Clean Package Cache', icon: <Download size={20} />, color: 'text-[var(--color-warning)]', bg: 'bg-[var(--color-warning)]/10', border: 'border-[var(--color-warning)]/30', confirmMessage: 'Clean package manager cache?' },
  { id: 'system_update', label: 'System Update', icon: <RefreshCw size={20} />, color: 'text-[var(--color-accent)]', bg: 'bg-[var(--color-accent)]/10', border: 'border-[var(--color-accent)]/30', confirmMessage: 'Run system update? This may take a while.' },
]

export function ActionsTab() {
  const { call } = useBackend()
  const queryClient = useQueryClient()
  const [preview, setPreview] = useState<ActionPreview | null>(null)
  const [lastResult, setLastResult] = useState<{ success: boolean; message: string; output?: string } | null>(null)

  const initiateMutation = useMutation({
    mutationFn: async (action: string) => {
      const p = await call('SysOps.RunSystemAction', action)
      return p as ActionPreview
    },
    onSuccess: (p) => setPreview(p),
    onError: (err: any) => toast.error(`Failed to initiate action: ${err.message}`),
  })

  const confirmMutation = useMutation({
    mutationFn: async (handshakeID: string) => {
      const res = await call('App.ConfirmAction', handshakeID) as { success: boolean; message: string; detail?: string; error?: string }
      return res
    },
    onSuccess: (res) => {
      setLastResult({ success: res.success, message: res.message || res.error || '', output: res.detail })
      if (res.success) {
        toast.success('Action executed successfully')
      } else {
        toast.error(`Action failed: ${res.error || res.message}`)
      }
      queryClient.invalidateQueries()
    },
  })

  const handleConfirm = () => {
    if (preview) {
      confirmMutation.mutate(preview.handshake_id)
      setPreview(null)
    }
  }

  return (
    <div className="space-y-6">
      <ConfirmationModal
        preview={preview}
        onConfirm={handleConfirm}
        onCancel={() => setPreview(null)}
      />

      <div className="grid grid-cols-4 gap-4">
        {actions.map(action => (
          <button
            key={action.id}
            onClick={() => { setLastResult(null); initiateMutation.mutate(action.id) }}
            disabled={initiateMutation.isPending || confirmMutation.isPending}
            className={cn('flex flex-col items-center gap-3 p-6 rounded-xl border transition-all hover:scale-[1.02] disabled:opacity-50 disabled:hover:scale-100', action.bg, action.border)}
          >
            <div className={action.color}>{action.icon}</div>
            <span className="text-sm font-bold text-[var(--color-text)]">{action.label}</span>
          </button>
        ))}
      </div>

      {lastResult && (
        <div className={cn('p-5 rounded-xl border animate-in fade-in slide-in-from-top-2', lastResult.success ? 'bg-[var(--color-success)]/10 border-[var(--color-success)]/30' : 'bg-[var(--color-danger)]/10 border-[var(--color-danger)]/30')}>
          <p className={cn('text-sm font-bold', lastResult.success ? 'text-[var(--color-success)]' : 'text-[var(--color-danger)]')}>{lastResult.message}</p>
          {lastResult.output && <CommandOutput output={lastResult.output} title="Command Output" />}
        </div>
      )}

      {(initiateMutation.isPending || confirmMutation.isPending) && (
        <div className="text-center py-8">
          <div className="animate-spin w-8 h-8 border-2 border-[var(--color-accent)] border-t-transparent rounded-full mx-auto" />
          <p className="text-sm text-[var(--color-text-faint)] mt-3">Processing...</p>
        </div>
      )}
    </div>
  )
}
