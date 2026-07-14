import { Cpu, Activity } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { CPUInfo, CPUExtendedInfo, PerformanceData } from '@/types'

interface CpuTabProps {
  cpuInfo: CPUInfo
}

export function CpuTab({ cpuInfo }: CpuTabProps) {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: cpuExtended } = useQuery<CPUExtendedInfo>({
    queryKey: ['sysops-cpu-extended'],
    queryFn: async () => { const r = await call('SysOps.GetCPUExtended'); return r as CPUExtendedInfo },
    refetchInterval: refreshInterval,
  })

  const { data: perfData } = useQuery<PerformanceData>({
    queryKey: ['sysops-performance'],
    queryFn: async () => { const r = await call('SysOps.GetPerformanceStats'); return r as PerformanceData },
    refetchInterval: refreshInterval,
  })

  const saturation = (cpuInfo.load_avg_1 / cpuInfo.logical_cores) * 100

  return (
    <div className="space-y-8">
      {/* CPU Overview */}
      <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-8 shadow-2xl">
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-3">
            <Cpu size={20} className="text-[var(--color-accent)]" />
            <h3 className="text-lg font-bold text-[var(--color-text)] uppercase tracking-widest">Processor Health</h3>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-3xl font-bold text-[var(--color-text)] tabular-nums">{cpuInfo.percent.toFixed(1)}%</span>
            <span className={`text-xs font-bold px-2 py-0.5 rounded border ${saturation > 80 ? 'bg-[var(--color-danger)]/20 text-[var(--color-danger)] border-[var(--color-danger)]/30' : 'bg-[var(--color-success)]/20 text-[var(--color-success)] border-[var(--color-success)]/30'}`}>
              {saturation.toFixed(0)}% Saturation
            </span>
          </div>
        </div>
        <div className="grid grid-cols-3 gap-6 mb-6">
          <div className="text-center">
            <p className="text-2xl font-bold text-[var(--color-text)] tabular-nums">{cpuInfo.physical_cores}</p>
            <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase">Physical Cores</p>
          </div>
          <div className="text-center">
            <p className="text-2xl font-bold text-[var(--color-text)] tabular-nums">{cpuInfo.logical_cores}</p>
            <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase">Logical Cores</p>
          </div>
          <div className="text-center">
            <p className="text-2xl font-bold text-[var(--color-text)] tabular-nums">{cpuExtended?.temperature ? `${cpuExtended.temperature.toFixed(1)}°C` : 'N/A'}</p>
            <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase">Temperature</p>
          </div>
        </div>
        <div className="grid grid-cols-2 gap-x-8 gap-y-3">
          {cpuInfo.per_cpu.map((p, i) => (
            <Bar key={i} label={`Core ${i}`} value={p} />
          ))}
        </div>
      </div>

      {/* Load Average */}
      {perfData && (
        <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-8 shadow-2xl">
          <div className="flex items-center gap-3 mb-6">
            <Activity size={20} className="text-[var(--color-accent)]" />
            <h3 className="text-lg font-bold text-[var(--color-text)] uppercase tracking-widest">Load Average</h3>
          </div>
          <div className="grid grid-cols-3 gap-6">
            <div className="text-center">
              <p className="text-3xl font-bold text-[var(--color-text)] tabular-nums">{perfData.load_average.load_1.toFixed(2)}</p>
              <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase">1 Minute</p>
            </div>
            <div className="text-center">
              <p className="text-3xl font-bold text-[var(--color-text)] tabular-nums">{perfData.load_average.load_5.toFixed(2)}</p>
              <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase">5 Minutes</p>
            </div>
            <div className="text-center">
              <p className="text-3xl font-bold text-[var(--color-text)] tabular-nums">{perfData.load_average.load_15.toFixed(2)}</p>
              <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase">15 Minutes</p>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export function Bar({ label, value, max = 100, unit = '%' }: { label: string; value: number; max?: number; unit?: string }) {
  const pct = Math.min((value / max) * 100, 100)
  const barColor = pct >= 70 ? 'var(--color-danger)' : pct >= 25 ? 'var(--color-warning)' : 'var(--color-success)'
  return (
    <div className="space-y-1">
      <div className="flex items-center justify-between">
        <span className="text-[var(--color-text-dim)] text-sm font-medium">{label}</span>
        <span className="text-[var(--color-text)] font-bold text-sm tabular-nums">{value.toFixed(1)}{unit}</span>
      </div>
      <div className="h-3 bg-[var(--color-panel-3)] rounded-full overflow-hidden border border-[var(--color-border)]">
        <div className="h-full rounded-full transition-all duration-700" style={{ width: `${pct}%`, background: `linear-gradient(90deg, ${barColor}88, ${barColor})` }} />
      </div>
    </div>
  )
}
