import { useQuery } from '@tanstack/react-query'
import { Activity, ShieldCheck, AlertTriangle, Cpu, MemoryStick, HardDrive, Zap } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { useConfigStore } from '@/stores/useConfigStore'
import { cn } from '@/lib/utils'
import type { DashboardData } from '@/types'

export function ContextPreview() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const { stagedChanges } = useConfigStore()

  const { data: dashboard } = useQuery<DashboardData>({
    queryKey: ['dashboard-mini-preview'],
    queryFn: async () => await call('Dashboard.GetDashboardData') as DashboardData,
    refetchInterval: refreshInterval,
  })

  if (!dashboard) return null

  return (
    <div className="w-80 border-l border-[var(--color-border)] bg-[var(--color-panel)] flex flex-col shrink-0 animate-in slide-in-from-right-4 duration-500">
      <div className="p-8 border-b border-[var(--color-border)] bg-[var(--color-panel-2)]/50">
        <h3 className="text-xs font-black text-text-faint uppercase tracking-[0.2em] flex items-center gap-2">
          <Activity size={14} className="text-accent" />
          Live Context
        </h3>
        <p className="text-[10px] text-text-dim mt-1 font-medium leading-relaxed italic">
          "Verifying system stability during configuration drift."
        </p>
      </div>

      <div className="flex-1 overflow-y-auto p-8 space-y-8">
        {/* Core Metrics */}
        <div className="space-y-5">
          <ContextMetricItem
            icon={<Cpu size={18} />}
            label="Processor"
            value={dashboard.cpu.value}
            unit="%"
            color={dashboard.cpu.value > 80 ? 'text-danger' : 'text-success'}
          />
          <ContextMetricItem
            icon={<MemoryStick size={18} />}
            label="Memory"
            value={dashboard.memory.value}
            unit="%"
            color={dashboard.memory.value > 85 ? 'text-warning' : 'text-success'}
          />
          <ContextMetricItem
            icon={<HardDrive size={18} />}
            label="Storage"
            value={dashboard.disk.value}
            unit="%"
            color={dashboard.disk.value > 90 ? 'text-danger' : 'text-success'}
          />
        </div>

        {/* Shadow State Summary */}
        {stagedChanges.size > 0 && (
          <div className="pt-8 border-t border-[var(--color-border)]">
            <h4 className="text-[10px] font-black text-warning uppercase tracking-widest mb-4 flex items-center gap-2">
              <Zap size={12} /> Shadow State Active
            </h4>
            <div className="space-y-3">
              {Array.from(stagedChanges.entries()).map(([key, val]) => (
                <div key={key} className="p-3 rounded-xl bg-warning/5 border border-warning/20">
                  <div className="flex justify-between items-center mb-1">
                    <span className="text-[10px] font-bold text-text-dim uppercase">{key}</span>
                    <span className="text-[10px] font-black text-warning uppercase">Staged</span>
                  </div>
                  <p className="text-xs font-mono font-bold text-text truncate">{String(val)}</p>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* System Health */}
        <div className="pt-8 border-t border-[var(--color-border)]">
          <h4 className="text-[10px] font-black text-text-faint uppercase tracking-widest mb-4">Neural Awareness</h4>
          <div className="p-5 rounded-2xl bg-accent-soft border border-accent/20 space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                {dashboard.alerts > 0 ? <AlertTriangle size={14} className="text-warning" /> : <ShieldCheck size={14} className="text-success" />}
                <span className="text-[10px] font-bold text-text-dim uppercase">Alerts</span>
              </div>
              <span className={cn("text-[10px] font-black tabular-nums", dashboard.alerts > 0 ? "text-warning" : "text-success")}>
                {dashboard.alerts > 0 ? dashboard.alerts : 'NONE'}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-[10px] font-bold text-text-dim uppercase tracking-tighter">Uptime</span>
              <span className="text-[10px] font-black text-text tabular-nums">{dashboard.uptime}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function ContextMetricItem({ icon, label, value, unit, color }: { icon: React.ReactNode, label: string, value: number, unit: string, color: string }) {
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 text-text-dim">
          {icon}
          <span className="text-[10px] font-black uppercase tracking-wider">{label}</span>
        </div>
        <span className={cn("text-sm font-black tabular-nums", color)}>
          {Math.round(value)}{unit}
        </span>
      </div>
      <div className="h-1 bg-[var(--color-panel-3)] rounded-full overflow-hidden border border-[var(--color-border)]/50">
        <div
          className={cn("h-full transition-all duration-1000", color.replace('text-', 'bg-'))}
          style={{ width: `${value}%` }}
        />
      </div>
    </div>
  )
}
