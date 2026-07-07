import { useState, useEffect, useCallback, useMemo } from 'react'
import {
  Monitor,
  LayoutList,
  Info,
  Cpu,
  MemoryStick,
  HardDrive,
  Activity,
  Search,
  RefreshCw,
  X,
  Trash2,
  AlertTriangle,
  Server,
  Clock,
  Copy,
  Check,
  Disc,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { ProcessManager } from '@/components/dialogs/ProcessManager'
import { useBackend } from '@/hooks/useBackend'
import { useEvents } from '@/hooks/useEvents'
import {
  mockCPUInfo,
  mockMemoryInfo,
  mockDiskInfo,
  mockProcesses,
  mockSystemInfo,
} from '@/lib/mockData'
import type { CPUInfo, MemoryInfo, DiskInfo, ProcessInfo, SystemInfo, DiskPartition } from '@/types'

// ── Types ──

type SysOpsTab = 'overview' | 'processes' | 'system-info'

interface TabDef {
  id: SysOpsTab
  label: string
  icon: React.ReactNode
}

const tabs: TabDef[] = [
  { id: 'overview', label: 'Overview', icon: <Monitor size={16} /> },
  { id: 'processes', label: 'Processes', icon: <LayoutList size={16} /> },
  { id: 'system-info', label: 'System Info', icon: <Info size={16} /> },
]

type ProcSortKey = 'pid' | 'name' | 'cpu' | 'memory' | 'mem_pct' | 'status' | 'num_fds'

const BAR_GREEN = '#4ade80'
const BAR_AMBER = '#fbbf24'
const BAR_RED = '#f87171'

function pctColor(pct: number): string {
  return pct >= 70 ? BAR_RED : pct >= 25 ? BAR_AMBER : BAR_GREEN
}

// ── Inline Helpers ──

function Bar({
  label,
  value,
  max = 100,
  color,
  unit = '%',
  showLabel = true,
  className,
}: {
  label: string
  value: number
  max?: number
  color?: string
  unit?: string
  showLabel?: boolean
  className?: string
}) {
  const pct = Math.min((value / max) * 100, 100)
  const barColor = color ?? pctColor(pct)

  return (
    <div className={cn('space-y-1', className)}>
      {showLabel && (
        <div className="flex items-center justify-between text-sm">
          <span className="text-text-dim">{label}</span>
          <span className="text-text font-mono text-xs">
            {typeof value === 'number' && value % 1 !== 0 ? value.toFixed(1) : value}
            {unit}
          </span>
        </div>
      )}
      <div className="h-2.5 bg-[var(--color-border)] rounded-full overflow-hidden">
        <div
          className="h-full rounded-full transition-all duration-700 ease-out"
          style={{
            width: `${pct}%`,
            background: `linear-gradient(90deg, ${barColor}88, ${barColor})`,
          }}
        />
      </div>
    </div>
  )
}

function InfoRow({
  label,
  value,
  copyable = false,
}: {
  label: string
  value: string | number
  copyable?: boolean
}) {
  return (
    <div className="flex items-center justify-between py-1.5 border-b border-[var(--color-border)] last:border-0">
      <span className="text-xs text-text-dim">{label}</span>
      {copyable ? (
        <CopyChip text={String(value)} />
      ) : (
        <span className="text-xs text-text font-mono font-medium">{String(value)}</span>
      )}
    </div>
  )
}

function Badge({ text, color }: { text: string; color?: string }) {
  const colorClass =
    color === 'green'
      ? 'bg-success/20 text-success'
      : color === 'amber'
        ? 'bg-warning/20 text-warning'
        : color === 'red'
          ? 'bg-danger/20 text-danger'
          : 'bg-muted/20 text-muted'
  return (
    <span className={cn('inline-block px-1.5 py-0.5 text-[10px] font-medium rounded-full', colorClass)}>
      {text}
    </span>
  )
}

function CopyChip({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)
  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    })
  }, [text])
  return (
    <button
      onClick={handleCopy}
      className="inline-flex items-center gap-1 px-2 py-0.5 bg-panel border border-[var(--color-border)] rounded-full text-[11px] text-text font-mono hover:border-primary/50 transition-colors shrink-0"
    >
      {text}
      {copied ? <Check size={10} className="text-success" /> : <Copy size={10} className="text-text-dim" />}
    </button>
  )
}

// ── Section Header ──

function SectionHeader({ icon, title }: { icon: React.ReactNode; title: string }) {
  return (
    <h3 className="text-xs font-semibold text-text-dim uppercase tracking-wider flex items-center gap-1.5 mb-1">
      {icon} {title}
    </h3>
  )
}

// ── Sort Helper ──

const SortIcon = ({ col, sortKey, sortDir }: { col: ProcSortKey; sortKey: ProcSortKey; sortDir: 'asc' | 'desc' }) => {
  if (sortKey !== col) return null
  return <span className="ml-1 text-[10px]">{sortDir === 'asc' ? '▲' : '▼'}</span>
}

// ── Main Component ──

export function SysOps() {
  const { call } = useBackend()

  // ── Tab state ──
  const [activeTab, setActiveTab] = useState<SysOpsTab>('overview')

  // ── Data state ──
  const [cpuInfo, setCpuInfo] = useState<CPUInfo>(mockCPUInfo)
  const [memInfo, setMemInfo] = useState<MemoryInfo>(mockMemoryInfo)
  const [diskInfo, setDiskInfo] = useState<DiskInfo>(mockDiskInfo)
  const [sysInfo, setSysInfo] = useState<SystemInfo>(mockSystemInfo)
  const [processes, setProcesses] = useState<ProcessInfo[]>(mockProcesses)

  // ── Process UI state ──
  const [processSearch, setProcessSearch] = useState('')
  const [sortKey, setSortKey] = useState<ProcSortKey>('cpu')
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc')
  const [procManagerOpen, setProcManagerOpen] = useState(false)
  const [autoRefresh, setAutoRefresh] = useState(true)
  const [killConfirmPid, setKillConfirmPid] = useState<number | null>(null)

  // ── Data loading ──
  const loadData = useCallback(async () => {
    const [newCpu, newMem, newDisk, newSys, newProcs] = await Promise.all([
      call('SysOps.GetCPUInfo').catch(() => mockCPUInfo()),
      call('SysOps.GetMemoryInfo').catch(() => mockMemoryInfo()),
      call('SysOps.GetDiskInfo').catch(() => mockDiskInfo()),
      call('SysOps.GetSystemInfo').catch(() => mockSystemInfo()),
      call('SysOps.ListAllProcesses', 100).catch(() => mockProcesses()),
    ])
    if (newCpu) setCpuInfo(newCpu as CPUInfo)
    if (newMem) setMemInfo(newMem as MemoryInfo)
    if (newDisk) setDiskInfo(newDisk as DiskInfo)
    if (newSys) setSysInfo(newSys as SystemInfo)
    if (newProcs) setProcesses(newProcs as ProcessInfo[])
  }, [call])

  // Initial load
  useEffect(() => {
    loadData()
  }, [loadData])

  // Auto-refresh every 4s
  useEffect(() => {
    if (!autoRefresh) return
    const interval = setInterval(loadData, 4000)
    return () => clearInterval(interval)
  }, [autoRefresh, loadData])

  // Live metrics events
  useEvents('metrics', useCallback((data: any) => {
    if (data?.cpu) {
      setCpuInfo((prev) => ({ ...prev, ...data.cpu }))
    }
    if (data?.memory) {
      setMemInfo((prev) => ({ ...prev, ...data.memory }))
    }
  }, []))

  // ── Sort logic ──
  const toggleSort = (key: ProcSortKey) => {
    if (sortKey === key) {
      setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'))
    } else {
      setSortKey(key)
      setSortDir('desc')
    }
  }

  const sortedProcesses = useMemo(() => {
    const searchLower = processSearch.toLowerCase()
    const filtered = searchLower
      ? processes.filter((p) => p.name.toLowerCase().includes(searchLower))
      : processes
    return [...filtered].sort((a, b) => {
      let cmp = 0
      switch (sortKey) {
        case 'pid':
        case 'cpu':
        case 'memory':
        case 'mem_pct':
        case 'num_fds':
          cmp = (a[sortKey] as number) - (b[sortKey] as number)
          break
        case 'name':
        case 'status':
          cmp = (a[sortKey] as string).localeCompare(b[sortKey] as string)
          break
      }
      return sortDir === 'asc' ? cmp : -cmp
    })
  }, [processes, processSearch, sortKey, sortDir])

  // ── Kill handler ──
  const handleKill = useCallback((pid: number) => {
    setProcesses((prev) => prev.filter((p) => p.pid !== pid))
    setKillConfirmPid(null)
  }, [])

  // ── Memory helpers ──
  const memPct = memInfo.used_percent ?? ((memInfo.used_gb / memInfo.total_gb) * 100)
  const availGb = memInfo.total_gb - memInfo.used_gb
  const swapPct = memInfo.swap_percent ?? 0

  // ── Render ──

  return (
    <div className="p-6 space-y-6 overflow-y-auto h-full">
      {/* ── Page header ── */}
      <div>
        <h1 className="text-2xl font-bold text-text flex items-center gap-2">
          <Activity size={24} /> System Operations
        </h1>
        <p className="text-text-dim text-sm mt-1">Monitor and manage system resources, processes, and hardware</p>
      </div>

      {/* ── Tab bar ── */}
      <div className="flex gap-1 bg-panel-2 rounded-[9px] p-1 inline-flex">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={cn(
              'flex items-center gap-2 px-4 py-2 rounded-[7px] text-sm font-medium transition-all',
              activeTab === tab.id
                ? 'bg-panel text-text shadow-sm'
                : 'text-text-dim hover:text-text',
            )}
          >
            {tab.icon}
            {tab.label}
          </button>
        ))}
      </div>

      {/* ════════════════════════════════════════ */}
      {/* OVERVIEW TAB */}
      {/* ════════════════════════════════════════ */}
      {activeTab === 'overview' && (
        <div className="space-y-6">
          {/* ── CPU Section ── */}
          <div className="bg-panel border border-[var(--color-border)] rounded-[12px] p-5 space-y-4">
            <SectionHeader icon={<Cpu size={14} />} title="CPU" />

            {/* Per-core bars */}
            {cpuInfo.per_cpu && cpuInfo.per_cpu.length > 0 && (
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-x-4 gap-y-2">
                {cpuInfo.per_cpu.map((corePct, i) => (
                  <Bar
                    key={i}
                    label={`Core ${i}`}
                    value={corePct}
                    showLabel={true}
                    unit="%"
                    className="space-y-0.5"
                  />
                ))}
              </div>
            )}

            {/* Load avg & details */}
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 pt-2 border-t border-[var(--color-border)]">
              <div>
                <span className="text-[11px] text-text-dim">Load (1m)</span>
                <p className="text-sm text-text font-mono">{cpuInfo.load_avg_1.toFixed(2)}</p>
              </div>
              <div>
                <span className="text-[11px] text-text-dim">Load (5m)</span>
                <p className="text-sm text-text font-mono">{cpuInfo.load_avg_5.toFixed(2)}</p>
              </div>
              <div>
                <span className="text-[11px] text-text-dim">Load (15m)</span>
                <p className="text-sm text-text font-mono">{cpuInfo.load_avg_15.toFixed(2)}</p>
              </div>
              <div>
                <span className="text-[11px] text-text-dim">Cores</span>
                <p className="text-sm text-text font-mono">{cpuInfo.core_count}</p>
              </div>
            </div>

            <div className="text-[11px] text-text-dim">
              Model: <span className="text-text font-mono">{cpuInfo.model_name}</span>
            </div>
          </div>

          {/* ── Memory Section ── */}
          <div className="bg-panel border border-[var(--color-border)] rounded-[12px] p-5 space-y-4">
            <SectionHeader icon={<MemoryStick size={14} />} title="Memory" />

            <div className="flex items-center gap-3 mb-1">
              <span className="text-2xl font-bold font-mono tracking-tight" style={{ color: pctColor(memPct) }}>
                {memPct.toFixed(1)}%
              </span>
              <div className="text-xs text-text-dim space-y-0.5">
                <p>{memInfo.used_gb.toFixed(1)} GB / {memInfo.total_gb.toFixed(1)} GB used</p>
                <p>{availGb.toFixed(1)} GB available</p>
              </div>
            </div>

            <Bar label="Usage" value={memPct} max={100} showLabel={false} />

            {/* Swap */}
            {memInfo.swap_total > 0 && (
              <div className="space-y-1 pt-2 border-t border-[var(--color-border)]">
                <div className="flex items-center justify-between text-xs">
                  <span className="text-text-dim">Swap</span>
                  <span className="text-text font-mono">
                    {swapPct.toFixed(1)}% ({memInfo.swap_used ? (memInfo.swap_used / 1024 / 1024 / 1024).toFixed(1) : '0.0'} GB)
                  </span>
                </div>
                <Bar label="Swap" value={swapPct} max={100} showLabel={false} />
              </div>
            )}
          </div>

          {/* ── Disk Section ── */}
          <div className="bg-panel border border-[var(--color-border)] rounded-[12px] p-5 space-y-4">
            <SectionHeader icon={<HardDrive size={14} />} title="Disk" />

            {diskInfo.partitions.map((part: DiskPartition) => {
              const totalGb = (part.total_bytes / 1024 / 1024 / 1024).toFixed(1)
              const usedGb = (part.used_bytes / 1024 / 1024 / 1024).toFixed(1)
              return (
                <div key={part.mountpoint} className="space-y-1.5 pb-3 border-b border-[var(--color-border)] last:border-0 last:pb-0">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Disc size={14} className="text-text-dim shrink-0" />
                      <span className="text-sm text-text font-medium">{part.mountpoint}</span>
                      <Badge text={part.fs_type} color={part.fs_type === 'NTFS' ? 'amber' : 'green'} />
                    </div>
                    <span className="text-xs text-text-dim font-mono">
                      {usedGb} GB / {totalGb} GB
                    </span>
                  </div>
                  <Bar label="Usage" value={part.used_percent} max={100} showLabel={false} />
                  <p className="text-[11px] text-text-dim font-mono truncate">{part.device}</p>
                </div>
              )
            })}
          </div>

          {/* ── System Info Card ── */}
          <div className="bg-panel border border-[var(--color-border)] rounded-[12px] p-5">
            <SectionHeader icon={<Server size={14} />} title="System" />
            <div className="grid grid-cols-2 sm:grid-cols-3 gap-x-6 gap-y-2 mt-2">
              <InfoRow label="Hostname" value={sysInfo.hostname} copyable />
              <InfoRow label="OS" value={sysInfo.os} />
              <InfoRow label="Platform" value={sysInfo.platform} />
              <InfoRow label="Kernel" value={sysInfo.kernel_version} />
              <InfoRow label="Uptime" value={sysInfo.uptime} />
              <InfoRow label="Processes" value={sysInfo.process_count} />
            </div>
          </div>
        </div>
      )}

      {/* ════════════════════════════════════════ */}
      {/* PROCESSES TAB */}
      {/* ════════════════════════════════════════ */}
      {activeTab === 'processes' && (
        <div className="space-y-4">
          {/* Toolbar */}
          <div className="flex items-center gap-3 flex-wrap">
            <div className="relative flex-1 min-w-[200px] max-w-sm">
              <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-text-dim" />
              <input
                type="text"
                placeholder="Search by name..."
                value={processSearch}
                onChange={(e) => setProcessSearch(e.target.value)}
                className="w-full bg-panel border border-[var(--color-border)] rounded-[9px] pl-9 pr-3 py-2 text-sm text-text placeholder-text-dim focus:outline-none focus:ring-2 focus:ring-primary/50"
              />
            </div>

            <label className="flex items-center gap-2 text-xs text-text-dim cursor-pointer select-none">
              <input
                type="checkbox"
                checked={autoRefresh}
                onChange={(e) => setAutoRefresh(e.target.checked)}
                className="rounded border-[var(--color-border)] text-primary focus:ring-primary/50"
              />
              Auto-refresh
            </label>

            <button
              onClick={loadData}
              className="flex items-center gap-1.5 px-3 py-2 text-xs font-medium text-text-dim border border-[var(--color-border)] rounded-[9px] hover:text-text hover:bg-panel-2 transition-colors"
            >
              <RefreshCw size={14} /> Refresh
            </button>

            <button
              onClick={() => setProcManagerOpen(true)}
              className="flex items-center gap-1.5 px-3 py-2 text-xs font-medium text-primary border border-primary/30 rounded-[9px] hover:bg-primary/10 transition-colors"
            >
              <Activity size={14} /> Open Manager
            </button>
          </div>

          {/* Table */}
          <div className="bg-panel border border-[var(--color-border)] rounded-[12px] overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-[var(--color-border)]">
                    {(['pid', 'name', 'cpu', 'mem_pct', 'memory', 'status', 'num_fds'] as ProcSortKey[]).map((col) => (
                      <th
                        key={col}
                        className={cn(
                          'text-left px-4 py-3 text-xs font-medium text-text-dim uppercase tracking-wider cursor-pointer hover:text-text transition-colors',
                          col === 'pid' ? 'w-[72px]' : '',
                          col === 'cpu' || col === 'mem_pct' || col === 'memory' || col === 'num_fds' ? 'text-right' : '',
                        )}
                        onClick={() => toggleSort(col)}
                      >
                        {col === 'pid' ? 'PID' :
                          col === 'name' ? 'Name' :
                            col === 'cpu' ? 'CPU%' :
                              col === 'mem_pct' ? 'Mem%' :
                                col === 'memory' ? 'Memory (MB)' :
                                  col === 'status' ? 'Status' :
                                    col === 'num_fds' ? 'FDs' : col}
                        <SortIcon col={col} sortKey={sortKey} sortDir={sortDir} />
                      </th>
                    ))}
                    <th className="w-[60px]" />
                  </tr>
                </thead>
                <tbody>
                  {sortedProcesses.slice(0, 100).map((proc) => {
                    const cpuColor = pctColor(proc.cpu)
                    const memColor = pctColor(proc.mem_pct)
                    return (
                      <tr
                        key={proc.pid}
                        className="border-b border-[var(--color-border)] hover:bg-panel-2/50 transition-colors cursor-pointer"
                        onClick={() => {
                          setKillConfirmPid(null)
                          setProcManagerOpen(true)
                        }}
                      >
                        <td className="px-4 py-2.5 font-mono text-xs text-text-dim">{proc.pid}</td>
                        <td className="px-4 py-2.5 text-text font-medium">{proc.name}</td>
                        <td className="px-4 py-2.5">
                          <div className="flex items-center gap-2 justify-end min-w-[100px]">
                            <span className="font-mono text-xs" style={{ color: cpuColor }}>
                              {proc.cpu.toFixed(1)}%
                            </span>
                            <div className="flex-1 max-w-[80px] h-1.5 bg-[var(--color-border)] rounded-full overflow-hidden">
                              <div
                                className="h-full rounded-full transition-all"
                                style={{
                                  width: `${Math.min(proc.cpu, 100)}%`,
                                  backgroundColor: cpuColor,
                                }}
                              />
                            </div>
                          </div>
                        </td>
                        <td className="px-4 py-2.5">
                          <div className="flex items-center gap-2 justify-end min-w-[100px]">
                            <span className="font-mono text-xs" style={{ color: memColor }}>
                              {proc.mem_pct.toFixed(1)}%
                            </span>
                            <div className="flex-1 max-w-[80px] h-1.5 bg-[var(--color-border)] rounded-full overflow-hidden">
                              <div
                                className="h-full rounded-full transition-all"
                                style={{
                                  width: `${Math.min(proc.mem_pct, 100)}%`,
                                  backgroundColor: memColor,
                                }}
                              />
                            </div>
                          </div>
                        </td>
                        <td className="px-4 py-2.5 text-right font-mono text-xs text-text-dim">
                          {proc.memory.toFixed(0)}
                        </td>
                        <td className="px-4 py-2.5 text-right">
                          <Badge
                            text={proc.status}
                            color={
                              proc.status === 'running'
                                ? 'green'
                                : proc.status === 'sleeping'
                                  ? 'amber'
                                  : undefined
                            }
                          />
                        </td>
                        <td className="px-4 py-2.5 text-right font-mono text-xs text-text-dim">
                          {proc.num_fds}
                        </td>
                        <td className="px-4 py-2.5 text-right">
                          {killConfirmPid === proc.pid ? (
                            <div className="flex items-center gap-1">
                              <button
                                onClick={(e) => {
                                  e.stopPropagation()
                                  handleKill(proc.pid)
                                }}
                                className="p-1 text-danger hover:bg-danger/10 rounded transition-colors"
                                title="Confirm kill"
                              >
                                <AlertTriangle size={14} />
                              </button>
                              <button
                                onClick={(e) => {
                                  e.stopPropagation()
                                  setKillConfirmPid(null)
                                }}
                                className="p-1 text-text-dim hover:text-text rounded transition-colors"
                                title="Cancel"
                              >
                                <X size={14} />
                              </button>
                            </div>
                          ) : (
                            <button
                              onClick={(e) => {
                                e.stopPropagation()
                                setKillConfirmPid(killConfirmPid === proc.pid ? null : proc.pid)
                              }}
                              className="p-1 text-text-dim hover:text-danger rounded transition-colors opacity-0 group-hover:opacity-100"
                              title="Kill process"
                            >
                              <Trash2 size={14} />
                            </button>
                          )}
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
            {sortedProcesses.length === 0 && (
              <p className="text-sm text-text-dim text-center py-8">No processes match your search</p>
            )}
          </div>
        </div>
      )}

      {/* ════════════════════════════════════════ */}
      {/* SYSTEM INFO TAB */}
      {/* ════════════════════════════════════════ */}
      {activeTab === 'system-info' && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {/* Hardware */}
          <div className="bg-panel border border-[var(--color-border)] rounded-[12px] p-5 space-y-2">
            <SectionHeader icon={<Cpu size={14} />} title="Hardware" />
            <InfoRow label="CPU Model" value={cpuInfo.model_name} copyable />
            <InfoRow label="Cores" value={cpuInfo.core_count} />
            <InfoRow label="RAM" value={`${memInfo.total_gb.toFixed(0)} GB`} />
            <InfoRow label="Load (1m)" value={cpuInfo.load_avg_1.toFixed(2)} />
            <InfoRow label="Load (5m)" value={cpuInfo.load_avg_5.toFixed(2)} />
            <InfoRow label="Load (15m)" value={cpuInfo.load_avg_15.toFixed(2)} />
            <InfoRow
              label="Partitions"
              value={diskInfo.partitions.length}
            />
          </div>

          {/* OS */}
          <div className="bg-panel border border-[var(--color-border)] rounded-[12px] p-5 space-y-2">
            <SectionHeader icon={<Server size={14} />} title="Operating System" />
            <InfoRow label="Hostname" value={sysInfo.hostname} copyable />
            <InfoRow label="OS" value={sysInfo.os} copyable />
            <InfoRow label="Platform" value={sysInfo.platform} copyable />
            <InfoRow label="Version" value={sysInfo.platform_version} copyable />
            <InfoRow label="Build" value={sysInfo.kernel_version} copyable />
            <InfoRow label="Kernel" value={sysInfo.kernel_version} copyable />
            <InfoRow label="Architecture" value={sysInfo.kernel_arch} copyable />
          </div>

          {/* Runtime */}
          <div className="bg-panel border border-[var(--color-border)] rounded-[12px] p-5 space-y-2">
            <SectionHeader icon={<Clock size={14} />} title="Runtime" />
            <InfoRow label="Uptime" value={sysInfo.uptime} copyable />
            <InfoRow label="Processes" value={sysInfo.process_count} />
            <InfoRow label="Virtualization" value={sysInfo.virtualization} copyable />
            <div className="pt-2">
              {diskInfo.partitions.map((part: DiskPartition) => (
                <InfoRow
                  key={part.mountpoint}
                  label={`Disk (${part.mountpoint})`}
                  value={`${(part.used_bytes / 1024 / 1024 / 1024).toFixed(0)} GB / ${(part.total_bytes / 1024 / 1024 / 1024).toFixed(0)} GB`}
                />
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Process Manager Dialog */}
      <ProcessManager open={procManagerOpen} onClose={() => setProcManagerOpen(false)} />
    </div>
  )
}
