import { useQuery } from '@tanstack/react-query'
import { Activity, Clock } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { PerformanceData } from '@/types'

export function PerformanceTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: perf } = useQuery<PerformanceData>({
    queryKey: ['sysops-performance'],
    queryFn: async () => { const r = await call('SysOps.GetPerformanceStats'); return r as PerformanceData },
    refetchInterval: refreshInterval,
  })

  if (!perf) return <div className="animate-pulse h-32 bg-[var(--color-panel-2)] rounded-xl" />

  const cpuUserPct = perf.cpu_times.total > 0 ? (perf.cpu_times.user / perf.cpu_times.total) * 100 : 0
  const cpuSystemPct = perf.cpu_times.total > 0 ? (perf.cpu_times.system / perf.cpu_times.total) * 100 : 0
  const cpuIdlePct = perf.cpu_times.total > 0 ? (perf.cpu_times.idle / perf.cpu_times.total) * 100 : 0

  return (
    <div className="space-y-8">
      {/* CPU Time Breakdown */}
      <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-8 shadow-2xl">
        <div className="flex items-center gap-3 mb-6">
          <Activity size={20} className="text-[var(--color-accent)]" />
          <h3 className="text-lg font-bold text-[var(--color-text)] uppercase tracking-widest">CPU Time Breakdown</h3>
        </div>
        <div className="grid grid-cols-4 gap-6">
          <div className="text-center">
            <p className="text-2xl font-bold text-[var(--color-accent)] tabular-nums">{cpuUserPct.toFixed(1)}%</p>
            <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase">User</p>
          </div>
          <div className="text-center">
            <p className="text-2xl font-bold text-[var(--color-warning)] tabular-nums">{cpuSystemPct.toFixed(1)}%</p>
            <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase">System</p>
          </div>
          <div className="text-center">
            <p className="text-2xl font-bold text-[var(--color-success)] tabular-nums">{cpuIdlePct.toFixed(1)}%</p>
            <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase">Idle</p>
          </div>
          <div className="text-center">
            <p className="text-2xl font-bold text-[var(--color-danger)] tabular-nums">{perf.io_wait.toFixed(1)}%</p>
            <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase">I/O Wait</p>
          </div>
        </div>
      </div>

      {/* Load Average */}
      <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-8 shadow-2xl">
        <div className="flex items-center gap-3 mb-6">
          <Clock size={20} className="text-[var(--color-accent)]" />
          <h3 className="text-lg font-bold text-[var(--color-text)] uppercase tracking-widest">Load Average</h3>
        </div>
        <div className="grid grid-cols-3 gap-6">
          <div className="text-center">
            <p className="text-3xl font-bold text-[var(--color-text)] tabular-nums">{perf.load_average.load_1.toFixed(2)}</p>
            <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase">1 min</p>
          </div>
          <div className="text-center">
            <p className="text-3xl font-bold text-[var(--color-text)] tabular-nums">{perf.load_average.load_5.toFixed(2)}</p>
            <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase">5 min</p>
          </div>
          <div className="text-center">
            <p className="text-3xl font-bold text-[var(--color-text)] tabular-nums">{perf.load_average.load_15.toFixed(2)}</p>
            <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase">15 min</p>
          </div>
        </div>
      </div>
    </div>
  )
}
