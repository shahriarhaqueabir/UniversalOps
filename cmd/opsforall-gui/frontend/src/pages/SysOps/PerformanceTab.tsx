import { useQuery } from '@tanstack/react-query'
import { Activity, Clock } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { Panel } from '@/components/ui/Panel'
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
      <Panel variant="elevated" padding="lg" category="system">
        <div className="flex items-center gap-3 mb-6">
          <Activity size={20} className="text-[var(--color-accent)]" />
          <h3 className="text-lg font-bold text-[var(--color-text)] uppercase tracking-widest">CPU Time Breakdown</h3>
        </div>
        <div className="grid grid-cols-4 gap-6">
          <div className="text-center">
            <p className="text-2xl font-bold text-[var(--color-accent)] tabular-nums">{Math.round(cpuUserPct)}%</p>
            <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase">User</p>
          </div>
          <div className="text-center">
            <p className="text-2xl font-bold text-[var(--color-warning)] tabular-nums">{Math.round(cpuSystemPct)}%</p>
            <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase">System</p>
          </div>
          <div className="text-center">
            <p className="text-2xl font-bold text-[var(--color-success)] tabular-nums">{Math.round(cpuIdlePct)}%</p>
            <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase">Idle</p>
          </div>
          <div className="text-center">
            <p className="text-2xl font-bold text-[var(--color-danger)] tabular-nums">{Math.round(perf.io_wait)}%</p>
            <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase">I/O Wait</p>
          </div>
        </div>
      </Panel>

      {/* Load Average */}
      <Panel variant="elevated" padding="lg" category="system">
        <div className="flex items-center gap-3 mb-6">
          <Clock size={20} className="text-[var(--color-accent)]" />
          <h3 className="text-lg font-bold text-[var(--color-text)] uppercase tracking-widest">Load Average</h3>
        </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
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
      </Panel>
    </div>
  )
}
