import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Play, Square, RotateCcw } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { ServiceEntry } from '@/types'
import { cn } from '@/lib/utils'

export function ServicesTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const queryClient = useQueryClient()

  const { data: services = [] } = useQuery<ServiceEntry[]>({
    queryKey: ['devops-services'],
    queryFn: async () => { const r = await call('DevOps.GetServices'); return (r as ServiceEntry[]) || [] },
    refetchInterval: refreshInterval,
  })

  const controlService = async (name: string, action: string) => {
    await call('DevOps.ControlService', name, action)
    queryClient.invalidateQueries({ queryKey: ['devops-services'] })
  }

  const running = services.filter(s => s.status === 'Running').length
  const stopped = services.filter(s => s.status !== 'Running').length

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-6">
        <div className="flex items-center gap-2 px-4 py-2 bg-[var(--color-success)]/10 border border-[var(--color-success)]/30 rounded-xl">
          <Play size={14} className="text-[var(--color-success)]" />
          <span className="text-sm font-bold text-[var(--color-success)] tabular-nums">{running}</span>
          <span className="text-xs text-[var(--color-text-faint)]">Running</span>
        </div>
        <div className="flex items-center gap-2 px-4 py-2 bg-[var(--color-warning)]/10 border border-[var(--color-warning)]/30 rounded-xl">
          <Square size={14} className="text-[var(--color-warning)]" />
          <span className="text-sm font-bold text-[var(--color-warning)] tabular-nums">{stopped}</span>
          <span className="text-xs text-[var(--color-text-faint)]">Stopped</span>
        </div>
      </div>

      <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-xl overflow-hidden">
        <div className="max-h-[600px] overflow-y-auto">
          <table className="w-full text-left">
            <thead className="sticky top-0 bg-[var(--color-panel-2)] border-b border-[var(--color-border)]">
              <tr>
                <th className="px-4 py-3 text-xs font-bold text-[var(--color-text-faint)] uppercase">Service</th>
                <th className="px-4 py-3 text-xs font-bold text-[var(--color-text-faint)] uppercase">Status</th>
                <th className="px-4 py-3 text-xs font-bold text-[var(--color-text-faint)] uppercase">Start Type</th>
                <th className="px-4 py-3 text-xs font-bold text-[var(--color-text-faint)] uppercase text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              {services.map((s, i) => (
                <tr key={i} className="border-b border-[var(--color-border)]/20 hover:bg-[var(--color-sidebar-hover)]">
                  <td className="px-4 py-3">
                    <span className="text-sm font-medium text-[var(--color-text)]">{s.display_name || s.name}</span>
                  </td>
                  <td className="px-4 py-3">
                    <span className={cn('text-xs font-bold px-2 py-0.5 rounded-full', s.status === 'Running' ? 'bg-[var(--color-success)]/20 text-[var(--color-success)]' : 'bg-[var(--color-warning)]/20 text-[var(--color-warning)]')}>
                      {s.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-[var(--color-text-dim)]">{s.start_type}</td>
                  <td className="px-4 py-3 text-right">
                    <div className="flex gap-1 justify-end">
                      {s.status !== 'Running' && (
                        <button onClick={() => controlService(s.name, 'start')} className="p-1.5 text-[var(--color-success)] hover:bg-[var(--color-success)]/10 rounded-lg"><Play size={14} /></button>
                      )}
                      {s.status === 'Running' && (
                        <>
                          <button onClick={() => controlService(s.name, 'stop')} className="p-1.5 text-[var(--color-danger)] hover:bg-[var(--color-danger)]/10 rounded-lg"><Square size={14} /></button>
                          <button onClick={() => controlService(s.name, 'restart')} className="p-1.5 text-[var(--color-warning)] hover:bg-[var(--color-warning)]/10 rounded-lg"><RotateCcw size={14} /></button>
                        </>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
