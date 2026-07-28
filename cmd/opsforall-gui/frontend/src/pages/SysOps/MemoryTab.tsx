import { MemoryStick, Activity } from 'lucide-react'
import { StatCard } from '@/components/ui/StatCard'
import { Panel } from '@/components/ui/Panel'
import type { MemoryInfo, ProcessInfo } from '@/types'
import { useQuery } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'

interface MemoryTabProps {
  memInfo: MemoryInfo
}

export function MemoryTab({ memInfo }: MemoryTabProps) {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const availableGB = memInfo.total_gb - memInfo.used_gb
  const cachedGB = memInfo.cached_bytes / (1024 * 1024 * 1024)
  const swapUsedGB = memInfo.swap_used / (1024 * 1024 * 1024)
  const swapTotalGB = memInfo.swap_total / (1024 * 1024 * 1024)

  const { data: processes = [] } = useQuery<ProcessInfo[]>({
    queryKey: ['sysops-memory-processes'],
    queryFn: async () => {
      const r = await call('SysOps.ListAllProcesses', 100)
      const list = (r as ProcessInfo[]) || []
      return list.sort((a, b) => b.memory - a.memory).slice(0, 10)
    },
    refetchInterval: refreshInterval,
  })

  return (
    <div className="space-y-8">
      <Panel variant="elevated" padding="lg" category="system">
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-3">
            <MemoryStick size={20} className="text-[var(--color-success)]" />
            <h3 className="text-lg font-bold text-[var(--color-text)] uppercase tracking-widest">RAM Usage</h3>
          </div>
          <span className="text-3xl font-bold text-[var(--color-success)] tabular-nums">{Math.round(memInfo.used_percent)}%</span>
        </div>
        <div className="h-6 bg-[var(--color-panel-3)] rounded-full overflow-hidden border border-[var(--color-border)] mb-6">
          <div className="h-full rounded-full bg-gradient-to-r from-[var(--color-success)]/60 to-[var(--color-success)] transition-all duration-700" style={{ width: `${memInfo.used_percent}%` }} />
        </div>
        <div className="grid grid-cols-4 gap-6">
          <StatCard label="Total" value={`${memInfo.total_gb.toFixed(1)} GB`} />
          <StatCard label="Used" value={`${memInfo.used_gb.toFixed(1)} GB`} valueClassName="text-[var(--color-success)]" />
          <StatCard label="Available" value={`${availableGB.toFixed(1)} GB`} />
          <StatCard label="Cached" value={`${cachedGB.toFixed(1)} GB`} />
        </div>
      </Panel>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
        <Panel variant="default" padding="lg" category="system">
          <div className="flex items-center gap-3 mb-6">
            <Activity size={18} className="text-[var(--color-accent)]" />
            <h3 className="text-sm font-black text-[var(--color-text)] uppercase tracking-widest">Top RAM Consumers</h3>
          </div>
          <div className="space-y-4">
            {processes.map((p) => (
              <div key={p.pid} className="flex items-center justify-between group">
                <div className="min-w-0">
                  <p className="text-sm font-bold text-[var(--color-text)] truncate">{p.name}</p>
                  <p className="text-[10px] font-black text-[var(--color-text-faint)] uppercase tracking-tighter">PID {p.pid}</p>
                </div>
                <div className="text-right">
                  <p className="text-sm font-black text-[var(--color-accent)] tabular-nums">{p.memory.toFixed(0)} MB</p>
                  <p className="text-[10px] font-bold text-[var(--color-text-dim)] tabular-nums">{Math.round(p.mem_pct)}%</p>
                </div>
              </div>
            ))}
          </div>
        </Panel>

        {memInfo.swap_total > 0 && (
          <Panel variant="elevated" padding="lg" category="system">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-sm font-black text-[var(--color-text)] uppercase tracking-widest">Swap Usage</h3>
              <span className="text-2xl font-bold text-[var(--color-warning)] tabular-nums">{Math.round(memInfo.swap_percent)}%</span>
            </div>
            <div className="h-4 bg-[var(--color-panel-3)] rounded-full overflow-hidden border border-[var(--color-border)]">
              <div className="h-full rounded-full bg-gradient-to-r from-[var(--color-warning)]/60 to-[var(--color-warning)] transition-all duration-700" style={{ width: `${memInfo.swap_percent}%` }} />
            </div>
            <div className="flex justify-between mt-3 text-sm">
              <span className="text-[var(--color-text-dim)]">{swapUsedGB.toFixed(1)} GB used</span>
              <span className="text-[var(--color-text-faint)]">{swapTotalGB.toFixed(1)} GB total</span>
            </div>
          </Panel>
        )}
      </div>
    </div>
  )
}
