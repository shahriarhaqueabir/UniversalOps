import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Monitor,
  LayoutList,
  Info,
  Cpu,
  MemoryStick,
  Search,
  Trash2,
  Disc,
  Box,
  Clock,
  Timer,
  HardDrive,
  Play,
  PauseCircle,
  AlertTriangle,
  Gpu,
  Battery,
  Lightbulb,
  ChevronRight,
} from 'lucide-react'
import {
  AreaChart as RechartsAreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
} from 'recharts'
import { format } from 'date-fns'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { ConfirmDialog } from '@/components/dialogs/ConfirmDialog'
import type { CPUInfo, MemoryInfo, ProcessInfo, SystemInfo, DiskInfo, MetricDataPoint } from '@/types'
import { DataFreshnessIndicator } from '@/components/ui/DataFreshnessIndicator'

// ── Types ──

type SysOpsTab = 'overview' | 'processes' | 'system-info'

interface TabDef {
  id: SysOpsTab
  label: string
  icon: React.ReactNode
}

const tabs: TabDef[] = [
  { id: 'overview', label: 'Analysis', icon: <Monitor size={20} /> },
  { id: 'processes', label: 'Runtime', icon: <LayoutList size={20} /> },
  { id: 'system-info', label: 'Inventory', icon: <Box size={20} /> },
]

interface GPUInfo {
  detected: boolean
  name: string
  vendor: string
  memory_gb: number
  driver: string
}

interface BatteryInfo {
  detected: boolean
  percent: number
  charging: boolean
  time_left_sec: number
  status: string
}

interface SystemRecommendation {
  category: string
  severity: string
  message: string
}

const BAR_GREEN = '#4ade80'
const BAR_AMBER = '#fbbf24'
const BAR_RED = '#f87171'

function pctColor(pct: number): string {
  return pct >= 70 ? BAR_RED : pct >= 25 ? BAR_AMBER : BAR_GREEN
}

const severityStyles: Record<string, { icon: string; color: string; bg: string; border: string }> = {
  critical: { icon: 'text-danger', color: 'text-danger', bg: 'bg-danger/20', border: 'border-danger/30' },
  warning: { icon: 'text-warning', color: 'text-warning', bg: 'bg-warning/20', border: 'border-warning/30' },
  info: { icon: 'text-accent', color: 'text-accent', bg: 'bg-accent/20', border: 'border-accent/30' },
}

// ── Enhanced Components ──

function SectionBriefing({ title, steps }: { title: string, steps: string[] }) {
  return (
    <div className="bg-panel-2 border border-border rounded-[var(--radius-lg)] p-8 shadow-xl">
      <div className="flex items-center gap-4 mb-6">
        <div className="w-10 h-10 rounded-xl bg-panel-3 border border-border flex items-center justify-center text-accent shadow-inner">
          <Info size={20} />
        </div>
        <h3 className="text-xl font-bold text-text uppercase tracking-widest">{title}</h3>
      </div>
      <div className="space-y-4">
        {steps.map((step, i) => (
          <div key={i} className="flex gap-4 group">
            <div className="flex flex-col items-center">
              <div className="w-6 h-6 rounded-full bg-accent/20 border border-accent/40 flex items-center justify-center text-[10px] font-bold text-accent">{i + 1}</div>
              {i < steps.length - 1 && <div className="w-px flex-1 bg-border my-1" />}
            </div>
            <p className="text-base text-text-dim leading-snug pb-2 group-hover:text-text transition-colors">{step}</p>
          </div>
        ))}
      </div>
    </div>
  )
}

function Bar({ label, value, max = 100, color, unit = '%', showLabel = true }: { label: string, value: number, max?: number, color?: string, unit?: string, showLabel?: boolean }) {
  const pct = Math.min((value / max) * 100, 100)
  const barColor = color ?? pctColor(pct)

  return (
    <div className="space-y-2">
      {showLabel && (
        <div className="flex items-center justify-between">
          <span className="text-text-dim text-lg font-medium">{label}</span>
          <span className="text-text font-bold text-lg tabular-nums">{value.toFixed(1)}{unit}</span>
        </div>
      )}
      <div className="h-4 bg-panel-3 rounded-full overflow-hidden border border-border shadow-inner">
        <div className="h-full rounded-full transition-all duration-700" style={{ width: `${pct}%`, background: `linear-gradient(90deg, ${barColor}88, ${barColor})` }} />
      </div>
    </div>
  )
}

function ProcessTreeItem({ proc, allProcs, depth = 0, onKill }: { proc: ProcessInfo, allProcs: ProcessInfo[], depth?: number, onKill: (p: ProcessInfo) => void }) {
  const children = allProcs.filter(p => p.ppid === proc.pid && p.pid !== proc.pid)
  const [expanded, setExpanded] = useState(depth < 2)

  return (
    <div className="flex flex-col">
      <div className={cn(
        "flex items-center gap-4 py-3 px-4 border-b border-border/10 hover:bg-sidebar-hover transition-colors group",
        depth > 0 && "ml-6 border-l border-border/30"
      )}>
        <div className="flex items-center gap-2 min-w-0 flex-1">
          {children.length > 0 && (
            <button onClick={() => setExpanded(!expanded)} className="text-text-faint hover:text-accent">
              <ChevronRight size={14} className={cn("transition-transform", expanded && "rotate-90")} />
            </button>
          )}
          {children.length === 0 && <div className="w-[14px]" />}
          <div className="flex flex-col min-w-0">
            <span className="text-sm font-bold text-text truncate">{proc.name}</span>
            <span className="text-[10px] font-black uppercase text-text-faint tracking-tighter">PID: {proc.pid}</span>
          </div>
        </div>
        <div className="flex items-center gap-6 shrink-0">
          <span className="text-xs font-mono font-bold text-accent tabular-nums w-12 text-right">{proc.cpu.toFixed(1)}%</span>
          <span className="text-xs font-mono text-text-dim tabular-nums w-12 text-right">{proc.memory.toFixed(0)}MB</span>
          <button onClick={() => onKill(proc)} className="opacity-0 group-hover:opacity-100 p-1.5 text-text-faint hover:text-danger hover:bg-danger/10 rounded-lg transition-all">
            <Trash2 size={16} />
          </button>
        </div>
      </div>
      {expanded && children.length > 0 && (
        <div className="flex flex-col">
          {children.map(c => <ProcessTreeItem key={c.pid} proc={c} allProcs={allProcs} depth={depth + 1} onKill={onKill} />)}
        </div>
      )}
    </div>
  )
}

function InfoRow({ label, value, copyable = false }: { label: string, value: string | number, copyable?: boolean }) {
  return (
    <div className="flex items-center justify-between py-4 border-b border-border last:border-0 group">
      <span className="text-xs font-semibold text-[var(--color-text-faint)] uppercase tracking-wider">{label}</span>
      {copyable ? (
        <button onClick={() => navigator.clipboard.writeText(String(value))} className="px-4 py-1.5 bg-panel-3 border border-border rounded-full text-sm font-medium text-[var(--color-text)] hover:border-accent transition-all shadow-md active:scale-95">
          {String(value)}
        </button>
      ) : (
        <span className="text-sm font-medium text-[var(--color-text)] tabular-nums">{String(value)}</span>
      )}
    </div>
  )
}

export function SysOps() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const queryClient = useQueryClient()
  const [activeTab, setActiveTab] = useState<SysOpsTab>('overview')
  const [search, setSearch] = useState('')
  const [killTarget, setKillTarget] = useState<{ pid: number, name: string } | null>(null)

  const { data: cpuInfo, dataUpdatedAt: cpuUpdatedAt } = useQuery<CPUInfo>({
    queryKey: ['sysops-cpu'],
    queryFn: async () => { const r = await call('SysOps.GetCPUInfo'); return r as CPUInfo },
    refetchInterval: refreshInterval,
  })

  const { data: memInfo } = useQuery<MemoryInfo>({
    queryKey: ['sysops-mem'],
    queryFn: async () => { const r = await call('SysOps.GetMemoryInfo'); return r as MemoryInfo },
    refetchInterval: refreshInterval,
  })

  const { data: sysInfo } = useQuery<SystemInfo>({
    queryKey: ['sysops-sys'],
    queryFn: async () => { const r = await call('SysOps.GetSystemInfo'); return r as SystemInfo },
    refetchInterval: refreshInterval,
  })

  const { data: diskInfo } = useQuery<DiskInfo>({
    queryKey: ['sysops-disk'],
    queryFn: async () => { const r = await call('SysOps.GetDiskInfo'); return r as DiskInfo },
    refetchInterval: refreshInterval,
  })

  const { data: processes = [] } = useQuery<ProcessInfo[]>({
    queryKey: ['sysops-processes'],
    queryFn: async () => { const r = await call('SysOps.ListAllProcesses', 100); return (r as ProcessInfo[]) || [] },
    refetchInterval: refreshInterval,
  })

  const { data: processTree = [] } = useQuery<ProcessInfo[]>({
    queryKey: ['sysops-process-tree'],
    queryFn: async () => { const r = await call('SysOps.GetProcessTree'); return (r as ProcessInfo[]) || [] },
    refetchInterval: refreshInterval,
  })

  const [processView, setProcessView] = useState<'list' | 'tree'>('list')

  const { data: topProcesses = [] } = useQuery<ProcessInfo[]>({
    queryKey: ['sysops-top-processes'],
    queryFn: async () => { const r = await call('SysOps.GetTopProcesses', 20); return (r as ProcessInfo[]) || [] },
    refetchInterval: refreshInterval,
  })

  const killProcess = async (pid: number) => {
    await call('DevOps.KillProcess', pid)
    queryClient.invalidateQueries({ queryKey: ['sysops-processes'] })
    setKillTarget(null)
  }

  const { data: cpuHistory = [] } = useQuery<MetricDataPoint[]>({
    queryKey: ['sysops-cpu-history'],
    queryFn: async () => {
      const r = await call('PipelineAPI.GetMetricHistoryWithTimestamps', 'cpu.percent', 60)
      return (r as MetricDataPoint[]) || []
    },
    refetchInterval: refreshInterval,
  })

  const { data: memHistory = [] } = useQuery<MetricDataPoint[]>({
    queryKey: ['sysops-mem-history'],
    queryFn: async () => {
      const r = await call('PipelineAPI.GetMetricHistoryWithTimestamps', 'memory.percent', 60)
      return (r as MetricDataPoint[]) || []
    },
    refetchInterval: refreshInterval,
  })

  const { data: diskHistory = [] } = useQuery<MetricDataPoint[]>({
    queryKey: ['sysops-disk-history'],
    queryFn: async () => {
      const r = await call('PipelineAPI.GetMetricHistoryWithTimestamps', 'disk.percent', 60)
      return (r as MetricDataPoint[]) || []
    },
    refetchInterval: refreshInterval,
  })

  const { data: gpuInfo } = useQuery<GPUInfo>({
    queryKey: ['sysops-gpu'],
    queryFn: async () => { const r = await call('SysOps.GetGPUInfo'); return r as GPUInfo },
    refetchInterval: refreshInterval,
  })

  const { data: batteryInfo } = useQuery<BatteryInfo>({
    queryKey: ['sysops-battery'],
    queryFn: async () => { const r = await call('SysOps.GetBatteryInfo'); return r as BatteryInfo },
    refetchInterval: refreshInterval,
  })

  const { data: recommendations = [] } = useQuery<SystemRecommendation[]>({
    queryKey: ['sysops-recommendations'],
    queryFn: async () => { const r = await call('SysOps.GetRecommendations'); return (r as SystemRecommendation[]) || [] },
    refetchInterval: refreshInterval,
  })

  if (!cpuInfo || !memInfo || !sysInfo || !diskInfo) {
    return (
      <div className="p-6 space-y-4 animate-pulse">
        <div className="h-8 w-48 bg-panel-2 rounded" />
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="h-32 bg-panel-2 rounded" />
          <div className="h-32 bg-panel-2 rounded" />
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full bg-[var(--color-bg)]">
      <ConfirmDialog
        open={killTarget !== null}
        title="Impactful Intervention"
        description={`Forcing the termination of "${killTarget?.name}" (PID: ${killTarget?.pid}) will immediately release its held resources. Ensure no unsaved work exists within this process.`}
        type="danger"
        confirmText="Execute Kill"
        onConfirm={() => killProcess(killTarget!.pid)}
        onClose={() => setKillTarget(null)}
      />

      <div className="p-8 border-b border-border bg-panel-2 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-text flex items-center gap-4">
            <Cpu size={32} className="text-accent" /> SYSTEM OPERATIONS
          </h1>
          <p className="text-text-dim text-lg mt-2">Architecture monitoring, runtime thread audit, and resource inventory.</p>
          <DataFreshnessIndicator lastUpdated={cpuUpdatedAt ? new Date(cpuUpdatedAt) : null} className="mt-1" />
        </div>
        <div className="flex gap-1 bg-panel border border-border rounded-2xl p-1.5 shadow-inner">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              role="tab"
              aria-selected={activeTab === tab.id}
              aria-label={`${tab.label} tab`}
              className={cn(
                'flex items-center gap-3 px-8 py-3 rounded-xl text-lg font-bold transition-all',
                activeTab === tab.id ? 'bg-accent text-white shadow-lg' : 'text-text-dim hover:text-text hover:bg-[var(--color-sidebar-hover)]',
              )}
            >
              {tab.icon}
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-6 space-y-6">
        {activeTab === 'overview' && (
          <div className="space-y-6">
            <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
              <div className="lg:col-span-1 space-y-8">
                <SectionBriefing
                  title="Compute Audit"
                  steps={[
                    "Identify CPU spikes (>80%) that correlate with specific tasks.",
                    "Verify Load Average stability across 1/5/15 minute windows.",
                    "Check RAM Occupancy for evidence of memory exhaustion.",
                    "Audit swap partition if physical RAM is > 90%."
                  ]}
                />
              </div>

              <div className="lg:col-span-3 grid grid-cols-1 md:grid-cols-2 gap-8">
                {/* CPU Card */}
                <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
                  <div className="flex items-center justify-between mb-8">
                    <div className="flex flex-col">
                      <h3 className="text-xl font-bold text-text uppercase tracking-widest flex items-center gap-3"><Cpu size={24} className="text-accent" /> Processor Health</h3>
                      <p className="text-sm font-bold text-text-faint mt-1 uppercase tracking-tighter">
                        {cpuInfo.physical_cores} Physical • {cpuInfo.logical_cores} Logical Cores
                      </p>
                    </div>
                    <div className="flex flex-col items-end">
                      <span className="text-2xl font-bold text-text">{cpuInfo.percent.toFixed(1)}%</span>
                      <span className={cn(
                        "text-[10px] font-bold px-2 py-0.5 rounded border mt-1 uppercase tracking-widest",
                        (cpuInfo.load_avg_1 / cpuInfo.logical_cores) > 0.8 ? "bg-danger/20 text-danger border-danger/30" : "bg-success/20 text-success border-success/30"
                      )}>
                        {((cpuInfo.load_avg_1 / cpuInfo.logical_cores) * 100).toFixed(0)}% Saturation
                      </span>
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-x-10 gap-y-4 max-h-[300px] overflow-y-auto pr-4 custom-scrollbar">
                    {cpuInfo.per_cpu.map((p, i) => <Bar key={i} label={`Core ${i}`} value={p} />)}
                  </div>
                </div>

                {/* Memory Card */}
                <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
                  <div className="flex items-center justify-between mb-8">
                    <h3 className="text-xl font-bold text-text uppercase tracking-widest flex items-center gap-3"><MemoryStick size={24} className="text-success" /> Volatile RAM</h3>
                    <span className="text-2xl font-bold text-[var(--color-success)] tabular-nums">{memInfo.used_percent.toFixed(1)}%</span>
                  </div>
                  <Bar label="Physical Allocation" value={memInfo.used_percent} color="#2dd4a7" showLabel={false} />
                  <div className="mt-8 pt-8 border-t border-border grid grid-cols-3 gap-6">
                    <div>
                      <p className="text-xs font-bold text-text-faint uppercase mb-1">Available</p>
                      <p className="text-sm font-bold text-[var(--color-text)]">{(memInfo.total_gb - memInfo.used_gb).toFixed(2)} GB</p>
                    </div>
                    <div>
                      <p className="text-xs font-bold text-text-faint uppercase mb-1">Cached</p>
                      <p className="text-sm font-bold text-[var(--color-text)]">{(memInfo.cached_bytes / (1024 * 1024 * 1024)).toFixed(2)} GB</p>
                    </div>
                    <div>
                      <p className="text-xs font-bold text-text-faint uppercase mb-1">Swap Usage</p>
                      <p className="text-sm font-bold text-[var(--color-warning)] tabular-nums">{memInfo.swap_percent.toFixed(1)}%</p>
                    </div>
                  </div>
                </div>

                {/* Disk Card */}
                <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
                  <div className="flex items-center justify-between mb-8">
                    <h3 className="text-xl font-bold text-text uppercase tracking-widest flex items-center gap-3"><Disc size={24} className="text-accent" /> Storage Analysis</h3>
                    <span className="text-2xl font-bold text-[var(--color-accent)] tabular-nums">
                      {diskInfo.partitions.length > 0
                        ? (diskInfo.partitions.reduce((a, p) => a + p.used_bytes, 0) / Math.max(diskInfo.partitions.reduce((a, p) => a + p.total_bytes, 0), 1) * 100).toFixed(1)
                        : 0}%
                    </span>
                  </div>
                  <div className="space-y-6">
                    {diskInfo.partitions.map((p, i) => (
                      <div key={i}>
                        <Bar label={p.mountpoint.length > 20 ? p.mountpoint.slice(0, 20) + '…' : p.mountpoint} value={p.used_percent} />
                        <div className="flex items-center justify-between text-sm mt-1">
                          <span className="text-text-dim">
                            {(p.total_bytes / 1e9).toFixed(1)} GB total · {(p.free_bytes / 1e9).toFixed(1)} GB free · {(p.used_bytes / 1e9).toFixed(1)} GB used
                          </span>
                          <span className="text-text-faint text-xs">{p.fs_type} · {p.device}</span>
                        </div>
                      </div>
                    ))}
                    {diskInfo.partitions.length === 0 && (
                      <p className="text-text-dim text-center py-4">No partitions detected.</p>
                    )}
                  </div>
                </div>

                {/* Uptime Card */}
                <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
                  <div className="flex items-center justify-between mb-8">
                    <h3 className="text-xl font-bold text-text uppercase tracking-widest flex items-center gap-3"><Clock size={24} className="text-success" /> System Uptime</h3>
                  </div>
                  <div className="flex flex-col items-center justify-center py-6">
                    <p className="text-3xl font-bold text-text tabular-nums">{sysInfo.uptime}</p>
                    <p className="text-sm font-bold text-text-faint mt-3 uppercase tracking-widest">Continuous Operation</p>
                  </div>
                  <div className="mt-6 pt-6 border-t border-border grid grid-cols-2 gap-6">
                    <div>
                      <p className="text-xs font-bold text-text-faint uppercase mb-1">Processes</p>
                      <p className="text-sm font-bold text-[var(--color-text)] tabular-nums">{sysInfo.process_count}</p>
                    </div>
                    <div>
                      <p className="text-xs font-bold text-text-faint uppercase mb-1">Hostname</p>
                      <p className="text-sm font-bold text-[var(--color-text)]">{sysInfo.hostname}</p>
                    </div>
                  </div>
                </div>

                {/* GPU Card */}
                {gpuInfo?.detected === true && (
                  <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
                    <div className="flex items-center justify-between mb-8">
                      <h3 className="text-xl font-bold text-text uppercase tracking-widest flex items-center gap-3"><Gpu size={24} className="text-accent" /> GPU</h3>
                      <span className="text-xs font-bold px-2 py-0.5 rounded bg-accent/20 text-accent border border-accent/30 uppercase tracking-widest">Detected</span>
                    </div>
                    <div className="space-y-4">
                      <div className="flex items-center justify-between">
                        <span className="text-sm font-bold text-text-faint uppercase">Name</span>
                        <span className="text-sm font-bold text-[var(--color-text)]">{gpuInfo.name}</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm font-bold text-text-faint uppercase">Vendor</span>
                        <span className="text-sm font-bold text-[var(--color-text)]">{gpuInfo.vendor}</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm font-bold text-text-faint uppercase">Memory</span>
                        <span className="text-sm font-bold text-[var(--color-accent)] tabular-nums">{gpuInfo.memory_gb.toFixed(1)} GB</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm font-bold text-text-faint uppercase">Driver</span>
                        <span className="text-sm font-bold text-[var(--color-text)]">{gpuInfo.driver}</span>
                      </div>
                    </div>
                  </div>
                )}

                {/* Battery Card */}
                {batteryInfo?.detected === true && (
                  <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
                    <div className="flex items-center justify-between mb-8">
                      <h3 className="text-xl font-bold text-text uppercase tracking-widest flex items-center gap-3"><Battery size={24} className="text-success" /> Battery</h3>
                      <span className={cn(
                        'text-xs font-bold px-2 py-0.5 rounded border uppercase tracking-widest',
                        batteryInfo.charging ? 'bg-success/20 text-success border-success/30' : 'bg-warning/20 text-warning border-warning/30'
                      )}>
                        {batteryInfo.charging ? 'Charging' : 'Discharging'}
                      </span>
                    </div>
                    <div className="space-y-4">
                      <div className="flex items-center justify-between">
                        <span className="text-sm font-bold text-text-faint uppercase">Charge Level</span>
                        <span className="text-sm font-bold text-[var(--color-text)] tabular-nums">{batteryInfo.percent.toFixed(0)}%</span>
                      </div>
                      <Bar label="Charge" value={batteryInfo.percent} color={batteryInfo.percent > 50 ? '#4ade80' : batteryInfo.percent > 20 ? '#fbbf24' : '#f87171'} showLabel={false} />
                      {batteryInfo.time_left_sec > 0 && (
                        <div className="flex items-center justify-between">
                          <span className="text-sm font-bold text-text-faint uppercase">Time Remaining</span>
                          <span className="text-sm font-bold text-[var(--color-text)] tabular-nums">
                            {Math.floor(batteryInfo.time_left_sec / 3600)}h {Math.floor((batteryInfo.time_left_sec % 3600) / 60)}m
                          </span>
                        </div>
                      )}
                      <div className="flex items-center justify-between">
                        <span className="text-sm font-bold text-text-faint uppercase">Status</span>
                        <span className="text-sm font-bold text-[var(--color-text)]">{batteryInfo.status}</span>
                      </div>
                    </div>
                  </div>
                )}

              </div>
            </div>

            {/* Metric History Trends */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
              <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
                <h3 className="text-xl font-bold text-text uppercase tracking-widest flex items-center gap-3">
                  <Cpu size={24} className="text-accent" /> CPU Trend
                </h3>
                <div className="mt-6 min-h-[200px]">
                  <ResponsiveContainer width="100%" height={200}>
                    <RechartsAreaChart data={cpuHistory.map(d => ({ time: format(new Date(d.Time), 'HH:mm'), value: d.Value }))}>
                      <defs>
                        <linearGradient id="sysopsCpuGrad" x1="0" y1="0" x2="0" y2="1">
                          <stop offset="0%" stopColor="var(--color-accent)" stopOpacity={0.3} />
                          <stop offset="100%" stopColor="var(--color-accent)" stopOpacity={0} />
                        </linearGradient>
                      </defs>
                      <CartesianGrid strokeDasharray="4 4" stroke="var(--color-border)" vertical={false} strokeOpacity={0.5} />
                      <XAxis dataKey="time" hide />
                      <YAxis hide domain={[0, 100]} />
                      <Tooltip
                        contentStyle={{ backgroundColor: 'var(--color-panel-3)', border: 'none', borderRadius: '12px' }}
                        formatter={(value: any) => [`${Number(value).toFixed(1)}%`, 'CPU']}
                      />
                      <Area type="monotone" dataKey="value" stroke="var(--color-accent)" strokeWidth={2} fill="url(#sysopsCpuGrad)" isAnimationActive={false} />
                    </RechartsAreaChart>
                  </ResponsiveContainer>
                </div>
                {cpuHistory.length === 0 && (
                  <p className="text-text-faint text-sm text-center py-4">No history data</p>
                )}
              </div>

              <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
                <h3 className="text-xl font-bold text-text uppercase tracking-widest flex items-center gap-3">
                  <MemoryStick size={24} className="text-success" /> Memory Trend
                </h3>
                <div className="mt-6 min-h-[200px]">
                  <ResponsiveContainer width="100%" height={200}>
                    <RechartsAreaChart data={memHistory.map(d => ({ time: format(new Date(d.Time), 'HH:mm'), value: d.Value }))}>
                      <defs>
                        <linearGradient id="sysopsMemGrad" x1="0" y1="0" x2="0" y2="1">
                          <stop offset="0%" stopColor="var(--color-success)" stopOpacity={0.3} />
                          <stop offset="100%" stopColor="var(--color-success)" stopOpacity={0} />
                        </linearGradient>
                      </defs>
                      <CartesianGrid strokeDasharray="4 4" stroke="var(--color-border)" vertical={false} strokeOpacity={0.5} />
                      <XAxis dataKey="time" hide />
                      <YAxis hide domain={[0, 100]} />
                      <Tooltip
                        contentStyle={{ backgroundColor: 'var(--color-panel-3)', border: 'none', borderRadius: '12px' }}
                        formatter={(value: any) => [`${Number(value).toFixed(1)}%`, 'Memory']}
                      />
                      <Area type="monotone" dataKey="value" stroke="var(--color-success)" strokeWidth={2} fill="url(#sysopsMemGrad)" isAnimationActive={false} />
                    </RechartsAreaChart>
                  </ResponsiveContainer>
                </div>
                {memHistory.length === 0 && (
                  <p className="text-text-faint text-sm text-center py-4">No history data</p>
                )}
              </div>

              <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
                <h3 className="text-xl font-bold text-text uppercase tracking-widest flex items-center gap-3">
                  <Disc size={24} className="text-warning" /> Disk Trend
                </h3>
                <div className="mt-6 min-h-[200px]">
                  <ResponsiveContainer width="100%" height={200}>
                    <RechartsAreaChart data={diskHistory.map(d => ({ time: format(new Date(d.Time), 'HH:mm'), value: d.Value }))}>
                      <defs>
                        <linearGradient id="sysopsDiskGrad" x1="0" y1="0" x2="0" y2="1">
                          <stop offset="0%" stopColor="var(--color-warning)" stopOpacity={0.3} />
                          <stop offset="100%" stopColor="var(--color-warning)" stopOpacity={0} />
                        </linearGradient>
                      </defs>
                      <CartesianGrid strokeDasharray="4 4" stroke="var(--color-border)" vertical={false} strokeOpacity={0.5} />
                      <XAxis dataKey="time" hide />
                      <YAxis hide domain={[0, 100]} />
                      <Tooltip
                        contentStyle={{ backgroundColor: 'var(--color-panel-3)', border: 'none', borderRadius: '12px' }}
                        formatter={(value: any) => [`${Number(value).toFixed(1)}%`, 'Disk']}
                      />
                      <Area type="monotone" dataKey="value" stroke="var(--color-warning)" strokeWidth={2} fill="url(#sysopsDiskGrad)" isAnimationActive={false} />
                    </RechartsAreaChart>
                  </ResponsiveContainer>
                </div>
                {diskHistory.length === 0 && (
                  <p className="text-text-faint text-sm text-center py-4">No history data</p>
                )}
              </div>
            </div>

            {/* Process Count Cards */}
            <div className="grid grid-cols-3 gap-6">
              <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-2xl flex items-center gap-5">
                <div className="w-12 h-12 rounded-xl bg-success/20 border border-success/30 flex items-center justify-center">
                  <Play size={20} className="text-success" />
                </div>
                <div>
                  <p className="text-2xl font-bold text-text tabular-nums">
                    {processes.filter(p => {
                      const s = p.status.toLowerCase()
                      return s !== 'stopped' && s !== 'zombie'
                    }).length}
                  </p>
                  <p className="text-xs font-bold text-text-faint uppercase tracking-widest">Running</p>
                </div>
              </div>
              <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-2xl flex items-center gap-5">
                <div className="w-12 h-12 rounded-xl bg-warning/20 border border-warning/30 flex items-center justify-center">
                  <PauseCircle size={20} className="text-warning" />
                </div>
                <div>
                  <p className="text-2xl font-bold text-text tabular-nums">
                    {processes.filter(p => p.status.toLowerCase() === 'stopped').length}
                  </p>
                  <p className="text-xs font-bold text-text-faint uppercase tracking-widest">Stopped</p>
                </div>
              </div>
              <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-2xl flex items-center gap-5">
                <div className="w-12 h-12 rounded-xl bg-danger/20 border border-danger/30 flex items-center justify-center">
                  <AlertTriangle size={20} className="text-danger" />
                </div>
                <div>
                  <p className="text-2xl font-bold text-text tabular-nums">
                    {processes.filter(p => p.status.toLowerCase() === 'zombie').length}
                  </p>
                  <p className="text-xs font-bold text-text-faint uppercase tracking-widest">Zombie</p>
                </div>
              </div>
            </div>

            {/* Top Consumers */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
              {/* Top CPU */}
              <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
                <div className="flex items-center gap-3 mb-6">
                  <Timer size={20} className="text-accent" />
                  <h3 className="text-lg font-bold text-text uppercase tracking-widest">Top CPU</h3>
                </div>
                <div className="space-y-4">
                  {[...topProcesses].sort((a, b) => b.cpu - a.cpu).slice(0, 5).map(p => (
                    <div key={p.pid} className="flex items-center justify-between">
                      <div className="flex flex-col min-w-0">
                        <span className="text-sm font-medium text-text truncate">{p.name}</span>
                        <span className="text-xs font-bold text-text-faint">PID {p.pid}</span>
                      </div>
                      <span className="text-sm font-bold text-accent tabular-nums ml-4">{p.cpu.toFixed(1)}%</span>
                    </div>
                  ))}
                  {topProcesses.length === 0 && <p className="text-text-dim text-sm text-center py-4">No data</p>}
                </div>
              </div>

              {/* Top RAM */}
              <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
                <div className="flex items-center gap-3 mb-6">
                  <MemoryStick size={20} className="text-success" />
                  <h3 className="text-lg font-bold text-text uppercase tracking-widest">Top RAM</h3>
                </div>
                <div className="space-y-4">
                  {[...topProcesses].sort((a, b) => b.mem_pct - a.mem_pct).slice(0, 5).map(p => (
                    <div key={p.pid} className="flex items-center justify-between">
                      <div className="flex flex-col min-w-0">
                        <span className="text-sm font-medium text-text truncate">{p.name}</span>
                        <span className="text-xs font-bold text-text-faint">PID {p.pid}</span>
                      </div>
                      <span className="text-sm font-bold text-success tabular-nums ml-4">{p.mem_pct.toFixed(1)}%</span>
                    </div>
                  ))}
                  {topProcesses.length === 0 && <p className="text-text-dim text-sm text-center py-4">No data</p>}
                </div>
              </div>

              {/* Top Disk IO */}
              <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
                <div className="flex items-center gap-3 mb-6">
                  <HardDrive size={20} className="text-warning" />
                  <h3 className="text-lg font-bold text-text uppercase tracking-widest">Top Disk IO</h3>
                </div>
                <div className="flex items-center justify-center py-12">
                  <p className="text-sm font-bold text-text-faint uppercase tracking-widest">Coming Soon</p>
                </div>
              </div>

              {/* Recommendations */}
              {recommendations.length > 0 && (
                <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
                  <div className="flex items-center gap-3 mb-6">
                    <Lightbulb size={20} className="text-warning" />
                    <h3 className="text-lg font-bold text-text uppercase tracking-widest">System Recommendations</h3>
                  </div>
                  <div className="space-y-4">
                    {recommendations.map((rec, i) => {
                      const styles = severityStyles[rec.severity] || severityStyles.info
                      return (
                        <div key={i} className={cn('flex items-start gap-4 p-4 rounded-[var(--radius-lg)] border', styles.bg, styles.border)}>
                          <div className={cn('flex-shrink-0 mt-0.5', styles.icon)}>
                            {rec.severity === 'critical' ? <AlertTriangle size={18} /> : rec.severity === 'warning' ? <AlertTriangle size={18} /> : <Info size={18} />}
                          </div>
                          <div className="flex-1 min-w-0">
                            <div className="flex items-center gap-2 mb-1">
                              <span className={cn('text-xs font-bold px-2 py-0.5 rounded border uppercase tracking-widest', styles.bg, styles.color, styles.border)}>
                                {rec.severity}
                              </span>
                              <span className="text-xs font-bold text-text-faint uppercase">{rec.category}</span>
                            </div>
                            <p className="text-sm font-medium text-[var(--color-text)] leading-snug">{rec.message}</p>
                          </div>
                        </div>
                      )
                    })}
                  </div>
                </div>
              )}
            </div>
          </div>
        )}

        {activeTab === 'processes' && (
          <div className="space-y-8">
            <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
              <div className="lg:col-span-1">
                <SectionBriefing
                  title="Runtime Audit"
                  steps={[
                    "Search for unauthorized background executables.",
                    "Sort by CPU to identify 'Resource Hogs'.",
                    "Check 'FDs' (File Descriptors) for potential handle leaks.",
                    "Use 'Force Kill' only for non-responsive applications."
                  ]}
                />
              </div>
              <div className="lg:col-span-3 space-y-6">
                <div className="flex items-center gap-6 bg-panel-2 border border-border p-6 rounded-[var(--radius-lg)] shadow-inner">
                  <div className="relative group flex-1">
                    <Search size={24} className="absolute left-5 top-1/2 -translate-y-1/2 text-text-faint group-focus-within:text-accent transition-colors" />
                    <input
                      type="text"
                      placeholder="Filter active threads..."
                      value={search}
                      onChange={(e) => setSearch(e.target.value)}
                      className="w-full bg-panel border border-border rounded-2xl pl-16 pr-4 py-3 text-sm font-medium text-[var(--color-text)] placeholder-[var(--color-text-faint)] focus:outline-none focus:border-accent shadow-xl"
                    />
                  </div>
                  <div className="flex gap-1 bg-panel border border-border rounded-xl p-1 shadow-lg">
                    <button
                      onClick={() => setProcessView('list')}
                      className={cn("px-4 py-2 rounded-lg text-xs font-bold transition-all", processView === 'list' ? "bg-accent text-white" : "text-text-faint hover:text-text")}
                    >
                      Flat List
                    </button>
                    <button
                      onClick={() => setProcessView('tree')}
                      className={cn("px-4 py-2 rounded-lg text-xs font-bold transition-all", processView === 'tree' ? "bg-accent text-white" : "text-text-faint hover:text-text")}
                    >
                      Process Tree
                    </button>
                  </div>
                  <div className="px-8 py-4 bg-panel border border-border rounded-2xl shadow-lg">
                    <span className="text-sm font-semibold text-text tabular-nums">{processes.length} active</span>
                  </div>
                </div>

                <div className="bg-panel border border-border rounded-[var(--radius-lg)] overflow-hidden shadow-2xl">
                  <div className="max-h-[600px] overflow-y-auto">
                    {processView === 'list' ? (
                      <table className="w-full text-left border-collapse">
                        <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border">
                          <tr>
                            <th className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Process Name</th>
                            <th className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider text-right">CPU %</th>
                            <th className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider text-right">RAM (MB)</th>
                            <th className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider text-right">Impact</th>
                            <th className="px-6 py-4 w-20" />
                          </tr>
                        </thead>
                        <tbody>
                          {processes.filter(p => p.name.toLowerCase().includes(search.toLowerCase())).length === 0 ? (
                            <tr>
                              <td colSpan={5} className="px-6 py-10 text-center">
                                <p className="text-text-faint text-lg font-bold">No processes match your filter.</p>
                                <p className="text-text-faint text-sm mt-2">Try a different search term or clear the filter.</p>
                              </td>
                            </tr>
                          ) : processes.filter(p => p.name.toLowerCase().includes(search.toLowerCase())).map(p => (
                            <tr key={p.pid} className="border-b border-border/20 hover:bg-[var(--color-sidebar-hover)] transition-all group">
                              <td className="px-6 py-4">
                                <div className="flex flex-col">
                                  <span className="text-sm font-medium text-text">{p.name}</span>
                                  <span className="text-sm font-bold text-text-faint uppercase tracking-tighter">PID: {p.pid} • {p.status}</span>
                                </div>
                              </td>
                              <td className="px-6 py-4 text-right font-semibold text-sm text-[var(--color-accent)] tabular-nums">{p.cpu.toFixed(1)}%</td>
                              <td className="px-6 py-4 text-right font-medium text-sm text-[var(--color-text-dim)] tabular-nums">{p.memory.toFixed(0)}</td>
                              <td className="px-6 py-4 text-right">
                                <span className={cn("px-4 py-1.5 rounded-full text-xs font-bold uppercase border", p.cpu > 5 ? "bg-danger/10 text-danger border-danger/30" : "bg-success/10 text-success border-success/30")}>
                                  {p.cpu > 5 ? 'High Impact' : 'Nominal'}
                                </span>
                              </td>
                              <td className="px-6 py-4">
                                <button onClick={() => setKillTarget({ pid: p.pid, name: p.name })} aria-label={`Kill process ${p.name} (PID ${p.pid})`} className="p-3 text-text-faint hover:text-danger hover:bg-danger/10 rounded-xl transition-all">
                                  <Trash2 size={24} />
                                </button>
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    ) : (
                      <div className="flex flex-col">
                        <div className="sticky top-0 z-10 grid grid-cols-[1fr_120px_120px_60px] bg-panel-2 border-b border-border px-8 py-4">
                          <div className="text-xs font-semibold text-text-dim uppercase">Process Tree</div>
                          <div className="text-xs font-semibold text-text-dim uppercase text-right">CPU %</div>
                          <div className="text-xs font-semibold text-text-dim uppercase text-right">RAM (MB)</div>
                          <div />
                        </div>
                        <div className="p-4">
                          {processTree
                            .filter(p => !processTree.some(parent => parent.pid === p.ppid && parent.pid !== p.pid))
                            .filter(p => p.name.toLowerCase().includes(search.toLowerCase()) || processTree.some(c => c.ppid === p.pid && c.name.toLowerCase().includes(search.toLowerCase())))
                            .map(p => (
                              <ProcessTreeItem
                                key={p.pid}
                                proc={p}
                                allProcs={processTree}
                                onKill={(target) => setKillTarget({ pid: target.pid, name: target.name })}
                              />
                            ))}
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'system-info' && (
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-12">
            <div className="bg-panel border border-border rounded-[var(--radius-xl)] p-10 shadow-2xl space-y-8">
              <h3 className="text-2xl font-bold text-text uppercase tracking-widest flex items-center gap-4"><Disc size={32} className="text-warning" /> Operating Logic</h3>
              <div className="space-y-2">
                <InfoRow label="Hostname" value={sysInfo.hostname} copyable />
                <InfoRow label="Platform" value={sysInfo.platform} />
                <InfoRow label="Kernel" value={sysInfo.kernel_version} />
                <InfoRow label="Build" value={sysInfo.platform_version} />
              </div>
            </div>

            <div className="bg-panel border border-border rounded-[var(--radius-xl)] p-10 shadow-2xl space-y-8">
              <h3 className="text-2xl font-bold text-text uppercase tracking-widest flex items-center gap-4"><Cpu size={32} className="text-accent" /> Hardware Tier</h3>
              <div className="space-y-2">
                <InfoRow label="Architecture" value={sysInfo.kernel_arch} />
                <InfoRow label="Logic Threads" value={cpuInfo.core_count} />
                <InfoRow label="Virtualization" value={sysInfo.virtualization || 'None'} />
                <InfoRow label="Uptime" value={sysInfo.uptime} />
              </div>
            </div>

            <div className="lg:col-span-1">
              <SectionBriefing
                title="Asset Inventory"
                steps={[
                  "Hostname confirms network identity.",
                  "Kernel version defines security patch level.",
                  "Thread count limits parallel task execution.",
                  "Uptime identifies pending reboot cycles."
                ]}
              />
            </div>
          </div>
        )}
      </div>
    </div >
  )
}
