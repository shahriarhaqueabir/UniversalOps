import { useQuery } from '@tanstack/react-query'
import { Users } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { LoggedInUserData } from '@/types'

export function UsersTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: users = [] } = useQuery<LoggedInUserData[]>({
    queryKey: ['sysops-users'],
    queryFn: async () => { const r = await call('SysOps.GetLoggedInUsers'); return (r as LoggedInUserData[]) || [] },
    refetchInterval: refreshInterval,
  })

  return (
    <div className="space-y-6">
      <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-8 shadow-2xl">
        <div className="flex items-center gap-3 mb-6">
          <Users size={20} className="text-[var(--color-success)]" />
          <h3 className="text-lg font-bold text-[var(--color-text)] uppercase tracking-widest">Active Users</h3>
          <span className="text-sm text-[var(--color-text-faint)] ml-auto">{users.length} logged in</span>
        </div>
        {users.length === 0 ? (
          <p className="text-[var(--color-text-dim)] text-center py-8">No active users detected</p>
        ) : (
          <div className="space-y-3">
            {users.map((u, i) => (
              <div key={i} className="flex items-center justify-between p-4 bg-[var(--color-panel-2)] rounded-xl border border-[var(--color-border)]">
                <div>
                  <p className="text-sm font-bold text-[var(--color-text)]">{u.user}</p>
                  <p className="text-xs text-[var(--color-text-faint)]">{u.terminal} from {u.host || 'local'}</p>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
