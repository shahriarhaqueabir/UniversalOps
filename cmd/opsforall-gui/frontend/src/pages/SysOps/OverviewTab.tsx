import { useQuery } from '@tanstack/react-query'
import {
  Cpu, MemoryStick, Activity, Clock, Server,
  Users, AlertTriangle, Zap,
} from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { Panel, PanelHeader } from '@/components/ui/Panel'
import { MiniStat } from '@/components/ui/MiniStat'
import { cn } from '@/lib/utils'
import type {
  CPUInfo, MemoryInfo, DiskInfo, SystemInfo,
  ProcessInfo, PerformanceData, LoggedInUserData,
  SystemRecommendation, GPUInfo, BatteryInfo,
} from '@/types'

/* ── Helpers ── */

function parseUptimeDays(uptime: string): number {
  if (!uptime) return 0
  let days = 0, hours = 0, minutes = 0
  const dMatch = uptime.match(/(\d+)d/)
  const hMatch = uptime.match(/(\d+)h/)
  const mMatch = uptime.match(/(\d+)m/)
  if (dMatch) days = parseInt(dMatch[1])
  if (hMatch) hours = parseInt(hMatch[1])
  if (mMatch) minutes = parseInt(mMatch[1])
  return days + hours / 24 + minutes / 1440
}

function healthColor(pct: number): string {
  if (pct >= 90) return 'var(--color-danger)'
  if (pct >= 75) return 'var(--color-warning)'
  return 'var(--color-success)'
}

/* ── Health Ring ── */

function HealthRing({ value, label, subtitle }: { value: number; label: string; subtitle?: string }) {
  const r = 36
  const circ = 2 * Math.PI * r
  const offset = circ - (value / 100) * circ
  const color = healthColor(value)

  return (
    <div className="flex flex-col items-center gap-2 group">
      <div className="relative w-24 h-24">
        <svg className="w-full h-full -rotate-90" viewBox="0 0 80 80">
          <circle cx="40" cy="40" r={r} fill="none" stroke="var(--color-panel-3)" strokeWidth="6" />
          <circle cx="40" cy="40" r={r} fill="none" stroke={color} strokeWidth="6"
            strokeLinecap="round" strokeDasharray={circ} strokeDashoffset={offset}
            className="transition-all duration-1000 ease-out"
          />
        </svg>
        <div className="absolute inset-0 flex items-center justify-center">
          <span className="text-lg font-black tabular-nums" style={{ color }}>{Math.round(value)}%</span>
        </div>
      </div>
      <span className="text-[10px] font-black text-text-dim uppercase tracking-[0.15em]">{label}</span>
      {subtitle && <span className="text-[9px] text-text-faint font-medium">{subtitle}</span>}
    </div>
  )
}

/* ── Main OverviewTab ── */

export function OverviewTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: cpuInfo } = useQuery<CPUInfo>({
    queryKey: ['sysops-cpu'],
    queryFn: async () => await call('SysOps.GetCPUInfo') as CPUInfo,
    refetchInterval: refreshInterval,
  })

  const { data: memInfo } = useQuery<MemoryInfo>({
    queryKey: ['sysops-mem'],
    queryFn: async () => await call('SysOps.GetMemoryInfo') as MemoryInfo,
    refetchInterval: refreshInterval,
  })

  const { data: diskInfo } = useQuery<DiskInfo>({
    queryKey: ['sysops-disk'],
    queryFn: async () => await call('SysOps.GetDiskInfo') as DiskInfo,
    refetchInterval: refreshInterval,
  })

  const { data: sysInfo } = useQuery<SystemInfo>({
    queryKey: ['sysops-sys'],
    queryFn: async () => await call('SysOps.GetSystemInfo') as SystemInfo,
    refetchInterval: refreshInterval,
  })

  const { data: topProcs = [] } = useQuery<ProcessInfo[]>({
    queryKey: ['sysops-top-procs'],
    queryFn: async () => (await call('SysOps.GetTopProcesses', 5) as ProcessInfo[]) || [],
    refetchInterval: refreshInterval,
  })

  const { data: perfData } = useQuery<PerformanceData>({
    queryKey: ['sysops-perf'],
    queryFn: async () => await call('SysOps.GetPerformanceStats') as PerformanceData,
    refetchInterval: refreshInterval,
  })

  const { data: users = [] } = useQuery<LoggedInUserData[]>({
    queryKey: ['sysops-users'],
    queryFn: async () => (await call('SysOps.GetLoggedInUsers') as LoggedInUserData[]) || [],
    refetchInterval: refreshInterval,
  })

  const { data: gpuInfo } = useQuery<GPUInfo>({
    queryKey: ['sysops-gpu'],
    queryFn: async () => await call('SysOps.GetGPUInfo') as GPUInfo,
    refetchInterval: refreshInterval,
  })

  const { data: battInfo } = useQuery<BatteryInfo>({
    queryKey: ['sysops-battery'],
    queryFn: async () => await call('SysOps.GetBatteryInfo') as BatteryInfo,
    refetchInterval: refreshInterval,
  })

  const { data: recommendations = [] } = useQuery<SystemRecommendation[]>({
    queryKey: ['sysops-recs'],
    queryFn: async () => (await call('SysOps.GetRecommendations') as SystemRecommendation[]) || [],
    refetchInterval: refreshInterval,
  })

  const { data: allProcs = [] } = useQuery<ProcessInfo[]>({
    queryKey: ['sysops-all-procs'],
    queryFn: async () => (await call('SysOps.ListAllProcesses', 500) as ProcessInfo[]) || [],
    refetchInterval: refreshInterval,
  })

  const cpuPct = cpuInfo?.percent ?? 0
  const memPct = memInfo?.used_percent ?? 0
  const uptimeDays = parseUptimeDays(sysInfo?.uptime ?? '')

  return (
    <div className="space-y-6 animate-in fade-in duration-500">

      {/* ── Health Cards ── */}
      <Panel padding="lg">
        <PanelHeader icon={<Activity size={20} />} title="System Health" />
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-6">
          <HealthRing value={cpuPct} label="CPU" subtitle={cpuInfo?.model_name?.split(' ').slice(0, 2).join(' ') || ''} />
          <HealthRing value={memPct} label="Memory" subtitle={memInfo ? `${memInfo.used_gb.toFixed(1)}/${memInfo.total_gb.toFixed(0)} GB` : ''} />
          {diskInfo?.partitions?.map(p => (
            <HealthRing key={p.mountpoint} value={p.used_percent} label={p.mountpoint} subtitle={`${p.fs_type} ${(p.total_bytes / 1e12).toFixed(1)}TB`} />
          ))}
          {gpuInfo?.detected && (
            <HealthRing value={gpuInfo.utilization} label="GPU" subtitle={gpuInfo.name.split(' ').slice(0, 2).join(' ')} />
          )}
          {battInfo?.detected && (
            <HealthRing value={battInfo.percent} label="Battery" subtitle={battInfo.charging ? 'Charging' : `${battInfo.percent}%`} />
          )}
        </div>
      </Panel>

      {/* ── Quick Stats ── */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <MiniStat label="CPU Cores" value={`${cpuInfo?.logical_cores ?? '—'}L / ${cpuInfo?.physical_cores ?? '—'}P`} icon={<Cpu size={24} />}
          variant={cpuPct < 75 ? 'success' : cpuPct < 90 ? 'warning' : 'danger'} />
        <MiniStat label="Memory" value={memInfo ? `${memInfo.used_gb.toFixed(1)} / ${memInfo.total_gb.toFixed(0)} GB` : '—'} icon={<MemoryStick size={24} />}
          variant={memPct < 75 ? 'success' : memPct < 90 ? 'warning' : 'danger'} />
        <MiniStat label="Uptime" value={uptimeDays >= 1 ? `${uptimeDays.toFixed(1)}d` : `${(uptimeDays * 24).toFixed(0)}h`} icon={<Clock size={24} />} variant="success" />
        <MiniStat label="Users" value={users.length} icon={<Users size={24} />} variant="default" />
      </div>

      {/* ── System Inventory + Memory Breakdown ── */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Panel padding="md">
          <PanelHeader icon={<Server size={20} />} title="System Inventory" />
          {sysInfo ? (
            <div className="space-y-3">
              {[
                { label: 'OS', value: sysInfo.os },
                { label: 'Version', value: sysInfo.platform_version },
                { label: 'Kernel', value: sysInfo.kernel_version },
                { label: 'Architecture', value: sysInfo.kernel_arch },
                { label: 'Hostname', value: sysInfo.hostname },
                { label: 'Platform', value: sysInfo.platform },
              ].map(item => (
                <div key={item.label} className="flex items-center justify-between py-1.5 border-b border-border/30 last:border-0">
                  <span className="text-[10px] font-black text-text-dim uppercase tracking-[0.15em]">{item.label}</span>
                  <span className="text-xs font-bold text-text font-mono">{item.value || '—'}</span>
                </div>
              ))}
            </div>
          ) : (
            <div className="h-20 flex items-center justify-center text-text-faint text-xs animate-pulse">Loading...</div>
          )}
        </Panel>

        <Panel padding="md">
          <PanelHeader icon={<MemoryStick size={20} />} title="Memory Breakdown" />
          {memInfo ? (
            <div className="space-y-4">
              {[
                { label: 'Used', value: memInfo.used_gb, total: memInfo.total_gb, pct: memInfo.used_percent, color: 'var(--color-danger)' },
                { label: 'Cached', value: memInfo.cached_bytes / 1e9, total: memInfo.total_gb, pct: (memInfo.cached_bytes / memInfo.total_bytes) * 100, color: 'var(--color-accent)' },
                { label: 'Available', value: memInfo.available_bytes / 1e9, total: memInfo.total_gb, pct: (memInfo.available_bytes / memInfo.total_bytes) * 100, color: 'var(--color-success)' },
                { label: 'Swap', value: memInfo.swap_used / 1e9, total: memInfo.swap_total / 1e9, pct: memInfo.swap_percent, color: 'var(--color-warning)' },
              ].map(item => (
                <div key={item.label}>
                  <div className="flex items-center justify-between mb-1">
                    <span className="text-[10px] font-black text-text-dim uppercase tracking-[0.15em]">{item.label}</span>
                    <span className="text-xs font-bold text-text tabular-nums">{item.value.toFixed(1)} / {item.total.toFixed(0)} GB</span>
                  </div>
                  <div className="h-1.5 bg-panel-3 rounded-full overflow-hidden">
                    <div className="h-full rounded-full transition-all duration-700" style={{ width: `${Math.min(item.pct, 100)}%`, backgroundColor: item.color }} />
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="h-20 flex items-center justify-center text-text-faint text-xs animate-pulse">Loading...</div>
          )}
        </Panel>
      </div>

      {/* ── Top Consumers + Running Processes ── */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Panel padding="md">
          <PanelHeader icon={<Zap size={20} />} title="Top Consumers" />
          {topProcs.length > 0 ? (
            <div className="space-y-2">
              {topProcs.map((proc, i) => (
                <div key={proc.pid || i} className="flex items-center gap-3 py-2 px-3 rounded-xl hover:bg-panel-2 transition-colors">
                  <span className="text-[10px] font-black text-text-faint w-5 text-right tabular-nums">#{i + 1}</span>
                  <div className="flex-1 min-w-0">
                    <p className="text-xs font-bold text-text truncate">{proc.name}</p>
                    <p className="text-[9px] text-text-faint font-mono">PID {proc.pid}</p>
                  </div>
                  <div className="flex items-center gap-4 text-[10px] font-bold tabular-nums">
                    <span className="text-danger">{Math.round(proc.cpu ?? 0)}% CPU</span>
                    <span className="text-accent">{Math.round(proc.mem_pct ?? 0)}% RAM</span>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="h-20 flex items-center justify-center text-text-faint text-xs animate-pulse">Loading...</div>
          )}
        </Panel>

        <Panel padding="md">
          <PanelHeader icon={<Activity size={20} />} title="Process Status" />
          {allProcs.length > 0 ? (
            <div className="space-y-4">
              {/* Status breakdown stat cards */}
              <div className="grid grid-cols-3 gap-3">
                {(() => {
                  const running = allProcs.filter(p => p.status === 'running').length
                  const stopped = allProcs.filter(p => p.status === 'stopped' || p.status === 'idle').length
                  const zombie = allProcs.filter(p => p.status === 'zombie' || p.status === 'zomb').length
                  const other = allProcs.length - running - stopped - zombie
                  return (
                    <>
                      <div className="bg-success/5 border border-success/20 rounded-xl p-4 text-center">
                        <span className="text-2xl font-black text-success tabular-nums">{running}</span>
                        <p className="text-[9px] font-black text-text-faint uppercase tracking-widest mt-1">Running</p>
                      </div>
                      <div className="bg-warning/5 border border-warning/20 rounded-xl p-4 text-center">
                        <span className="text-2xl font-black text-warning tabular-nums">{stopped}</span>
                        <p className="text-[9px] font-black text-text-faint uppercase tracking-widest mt-1">Stopped</p>
                      </div>
                      <div className={cn(
                        "rounded-xl p-4 text-center border",
                        zombie > 0 ? "bg-danger/5 border-danger/20" : "bg-panel-2 border-border/50"
                      )}>
                        <span className={cn("text-2xl font-black tabular-nums", zombie > 0 ? "text-danger" : "text-text-faint")}>{zombie}</span>
                        <p className="text-[9px] font-black text-text-faint uppercase tracking-widest mt-1">Zombie</p>
                      </div>
                      {other > 0 && (
                        <div className="bg-panel-2 border border-border/50 rounded-xl p-4 text-center col-span-3">
                          <span className="text-lg font-black text-text-dim tabular-nums">{other}</span>
                          <p className="text-[9px] font-black text-text-faint uppercase tracking-widest mt-1">Other Statuses</p>
                        </div>
                      )}
                    </>
                  )
                })()}
              </div>
              {/* Status distribution bar */}
              {allProcs.length > 0 && (
                <div className="h-2 bg-panel-3 rounded-full overflow-hidden flex">
                  {(() => {
                    const running = allProcs.filter(p => p.status === 'running').length
                    const stopped = allProcs.filter(p => p.status === 'stopped' || p.status === 'idle').length
                    const zombie = allProcs.filter(p => p.status === 'zombie' || p.status === 'zomb').length
                    const other = allProcs.length - running - stopped - zombie
                    const total = allProcs.length
                    return (
                      <>
                        {running > 0 && <div className="h-full bg-success transition-all" style={{ width: `${(running / total) * 100}%` }} title={`Running: ${running}`} />}
                        {stopped > 0 && <div className="h-full bg-warning transition-all" style={{ width: `${(stopped / total) * 100}%` }} title={`Stopped: ${stopped}`} />}
                        {zombie > 0 && <div className="h-full bg-danger transition-all" style={{ width: `${(zombie / total) * 100}%` }} title={`Zombie: ${zombie}`} />}
                        {other > 0 && <div className="h-full bg-text-faint/30 transition-all" style={{ width: `${(other / total) * 100}%` }} title={`Other: ${other}`} />}
                      </>
                    )
                  })()}
                </div>
              )}
              <p className="text-[10px] text-text-faint font-medium text-center">{allProcs.length} total processes</p>
            </div>
          ) : (
            <div className="h-20 flex items-center justify-center text-text-faint text-xs animate-pulse">Loading...</div>
          )}
        </Panel>
      </div>

      {/* ── Performance Metrics ── */}
      <Panel padding="md">
        <PanelHeader icon={<Activity size={20} />} title="Performance Metrics" />
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {perfData ? (
            <div className="space-y-4">
              <div>
                <span className="text-[10px] font-black text-text-dim uppercase tracking-[0.15em]">CPU Times</span>
                <div className="grid grid-cols-3 gap-3 mt-2">
                  {[
                    { label: 'User', value: perfData.cpu_times.user, color: 'var(--color-accent)' },
                    { label: 'System', value: perfData.cpu_times.system, color: 'var(--color-warning)' },
                    { label: 'Idle', value: perfData.cpu_times.idle, color: 'var(--color-success)' },
                  ].map(item => {
                    const pct = perfData.cpu_times.total > 0 ? (item.value / perfData.cpu_times.total) * 100 : 0
                    return (
                      <div key={item.label} className="bg-panel-2 rounded-xl p-3 text-center">
                        <span className="text-lg font-black tabular-nums" style={{ color: item.color }}>{Math.round(pct)}%</span>
                        <p className="text-[9px] font-black text-text-faint uppercase tracking-wider mt-1">{item.label}</p>
                      </div>
                    )
                  })}
                </div>
              </div>
              <div className="flex items-center justify-between py-2 border-t border-border/30">
                <span className="text-[10px] font-black text-text-dim uppercase tracking-[0.15em]">Load Average</span>
                <span className="text-xs font-bold text-text font-mono tabular-nums">
                  {perfData.load_average.load_1.toFixed(2)} / {perfData.load_average.load_5.toFixed(2)} / {perfData.load_average.load_15.toFixed(2)}
                </span>
              </div>
              <div className="flex items-center justify-between py-2 border-t border-border/30">
                <span className="text-[10px] font-black text-text-dim uppercase tracking-[0.15em]">I/O Wait</span>
                <span className={cn('text-xs font-bold font-mono tabular-nums', perfData.io_wait > 10 ? 'text-danger' : perfData.io_wait > 5 ? 'text-warning' : 'text-success')}>
                  {Math.round(perfData.io_wait)}%
                </span>
              </div>
            </div>
          ) : (
            <div className="h-20 flex items-center justify-center text-text-faint text-xs animate-pulse">Loading...</div>
          )}
        </div>
      </Panel>

      {/* ── Recommendations ── */}
      {recommendations.length > 0 && (
        <Panel padding="md">
          <PanelHeader icon={<AlertTriangle size={20} />} title="Recommendations" />
          <div className="space-y-2">
            {recommendations.map((rec) => (
              <div key={rec.message} className="flex items-start gap-3 py-3 px-4 rounded-xl bg-warning/5 border border-warning/10 hover:border-warning/20 transition-colors">
                <AlertTriangle size={16} className="text-warning mt-0.5 shrink-0" />
                <div>
                  <p className="text-xs font-bold text-text">{rec.message}</p>
                  <p className="text-[10px] text-text-dim mt-0.5 uppercase tracking-wider">{rec.category} · {rec.severity}</p>
                </div>
              </div>
            ))}
          </div>
        </Panel>
      )}

      {/* ── Logged-in Users ── */}
      {users.length > 0 && (
        <Panel padding="md">
          <PanelHeader icon={<Users size={20} />} title="Active Sessions" />
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
            {users.map((u) => (
              <div key={`${u.user}-${u.terminal}`} className="flex items-center gap-3 bg-panel-2 rounded-xl p-4 border border-border/50">
                <div className="w-8 h-8 rounded-lg bg-accent/10 flex items-center justify-center text-accent">
                  <Users size={16} />
                </div>
                <div className="min-w-0">
                  <p className="text-xs font-bold text-text truncate">{u.user}</p>
                  <p className="text-[9px] text-text-faint font-mono truncate">{u.terminal} @ {u.host}</p>
                </div>
              </div>
            ))}
          </div>
        </Panel>
      )}
    </div>
  )
}