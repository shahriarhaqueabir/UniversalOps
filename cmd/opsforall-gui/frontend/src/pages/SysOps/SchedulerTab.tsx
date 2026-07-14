import { useQuery } from '@tanstack/react-query'
import { Calendar, CheckCircle, XCircle } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { ScheduledTaskData } from '@/types'

export function SchedulerTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: tasks = [] } = useQuery<ScheduledTaskData[]>({
    queryKey: ['sysops-scheduler'],
    queryFn: async () => { const r = await call('SysOps.GetScheduledTasks'); return (r as ScheduledTaskData[]) || [] },
    refetchInterval: refreshInterval,
  })

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3 mb-2">
        <Calendar size={20} className="text-[var(--color-accent)]" />
        <h3 className="text-lg font-bold text-[var(--color-text)] uppercase tracking-widest">Scheduled Tasks</h3>
        <span className="text-sm text-[var(--color-text-faint)] ml-auto">{tasks.length} tasks</span>
      </div>

      {tasks.length === 0 ? (
        <div className="text-center py-12 bg-[var(--color-panel)] border border-[var(--color-border)] rounded-xl">
          <Calendar size={48} className="text-[var(--color-text-faint)] mx-auto mb-4" />
          <p className="text-[var(--color-text-dim)]">No scheduled tasks found</p>
        </div>
      ) : (
        <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-xl overflow-hidden">
          <div className="max-h-[500px] overflow-y-auto">
            <table className="w-full text-left">
              <thead className="sticky top-0 bg-[var(--color-panel-2)] border-b border-[var(--color-border)]">
                <tr>
                  <th className="px-4 py-3 text-xs font-bold text-[var(--color-text-faint)] uppercase">Task</th>
                  <th className="px-4 py-3 text-xs font-bold text-[var(--color-text-faint)] uppercase">Schedule</th>
                  <th className="px-4 py-3 text-xs font-bold text-[var(--color-text-faint)] uppercase">Status</th>
                  <th className="px-4 py-3 text-xs font-bold text-[var(--color-text-faint)] uppercase">Next Run</th>
                </tr>
              </thead>
              <tbody>
                {tasks.map((t, i) => (
                  <tr key={i} className="border-b border-[var(--color-border)]/20 hover:bg-[var(--color-sidebar-hover)]">
                    <td className="px-4 py-3 text-sm font-medium text-[var(--color-text)]">{t.name}</td>
                    <td className="px-4 py-3 text-sm text-[var(--color-text-dim)] font-mono">{t.schedule}</td>
                    <td className="px-4 py-3">
                      {t.enabled ? <CheckCircle size={14} className="text-[var(--color-success)]" /> : <XCircle size={14} className="text-[var(--color-text-faint)]" />}
                    </td>
                    <td className="px-4 py-3 text-sm text-[var(--color-text-faint)]">{t.next_run || 'N/A'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}
