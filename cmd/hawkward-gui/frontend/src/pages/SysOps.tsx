import { useState, useEffect, useCallback } from 'react'
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
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { ConfirmDialog } from '@/components/dialogs/ConfirmDialog'
import type { CPUInfo, MemoryInfo, ProcessInfo, SystemInfo, DiskInfo } from '@/types'

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



const BAR_GREEN = '#4ade80'
const BAR_AMBER = '#fbbf24'
const BAR_RED = '#f87171'

function pctColor(pct: number): string {
  return pct >= 70 ? BAR_RED : pct >= 25 ? BAR_AMBER : BAR_GREEN
}

// ── Enhanced Components ──

function SectionBriefing({ title, steps }: { title: string, steps: string[] }) {
  return (
    <div className="bg-panel-2 border border-border rounded-[24px] p-8 shadow-xl">
      <div className="flex items-center gap-4 mb-6">
        <div className="w-10 h-10 rounded-xl bg-panel-3 border border-border flex items-center justify-center text-accent shadow-inner">
          <Info size={20} />
        </div>
        <h3 className="text-xl font-black text-text uppercase tracking-widest">{title}</h3>
      </div>
      <div className="space-y-4">
        {steps.map((step, i) => (
          <div key={i} className="flex gap-4 group">
            <div className="flex flex-col items-center">
              <div className="w-6 h-6 rounded-full bg-accent/20 border border-accent/40 flex items-center justify-center text-[10px] font-black text-accent">{i + 1}</div>
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
          <span className="text-text font-black text-lg tabular-nums">{value.toFixed(1)}{unit}</span>
        </div>
      )}
      <div className="h-4 bg-panel-3 rounded-full overflow-hidden border border-border shadow-inner">
        <div className="h-full rounded-full transition-all duration-700" style={{ width: `${pct}%`, background: `linear-gradient(90deg, ${barColor}88, ${barColor})` }} />
      </div>
    </div>
  )
}

function InfoRow({ label, value, copyable = false }: { label: string, value: string | number, copyable?: boolean }) {
  return (
    <div className="flex items-center justify-between py-4 border-b border-border last:border-0 group">
      <span className="text-lg text-text-faint font-bold uppercase tracking-tighter">{label}</span>
      {copyable ? (
        <button onClick={() => navigator.clipboard.writeText(String(value))} className="px-4 py-1.5 bg-panel-3 border border-border rounded-full text-base text-text font-bold hover:border-accent transition-all shadow-md active:scale-95">
          {String(value)}
        </button>
      ) : (
        <span className="text-lg text-text font-black tabular-nums">{String(value)}</span>
      )}
    </div>
  )
}

// ── Main Component ──

export function SysOps() {
  const { call } = useBackend()
  const [activeTab, setActiveTab] = useState<SysOpsTab>('overview')
  const [cpuInfo, setCpuInfo] = useState<CPUInfo | null>(null)
  const [memInfo, setMemInfo] = useState<MemoryInfo | null>(null)
  const [sysInfo, setSysInfo] = useState<SystemInfo | null>(null)
  const [diskInfo, setDiskInfo] = useState<DiskInfo | null>(null)
  const [processes, setProcesses] = useState<ProcessInfo[]>([])
  const [search, setSearch] = useState('')
  const [killTarget, setKillTarget] = useState<{ pid: number, name: string } | null>(null)

  const loadData = useCallback(async () => {
    try {
      const [c, m, s, p, d] = await Promise.all([
        call('SysOps.GetCPUInfo'),
        call('SysOps.GetMemoryInfo'),
        call('SysOps.GetSystemInfo'),
        call('SysOps.ListAllProcesses', 100),
        call('SysOps.GetDiskInfo'),
      ])
      setCpuInfo(c as CPUInfo); setMemInfo(m as MemoryInfo); setSysInfo(s as SystemInfo); setProcesses(p as ProcessInfo[]); setDiskInfo(d as DiskInfo)
    } catch (err) { console.error(err) }
  }, [call])

  useEffect(() => { loadData(); const t = setInterval(loadData, 5000); return () => clearInterval(t) }, [loadData])

  const killProcess = async (pid: number) => {
    await call('DevOps.KillProcess', pid)
    loadData()
    setKillTarget(null)
  }

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
          <h1 className="text-3xl font-black text-text flex items-center gap-4">
            <Cpu size={32} className="text-accent" /> SYSTEM OPERATIONS
          </h1>
          <p className="text-text-dim text-lg mt-2">Architecture monitoring, runtime thread audit, and resource inventory.</p>
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
                'flex items-center gap-3 px-8 py-3 rounded-xl text-lg font-black transition-all',
                activeTab === tab.id ? 'bg-accent text-white shadow-lg' : 'text-text-dim hover:text-text hover:bg-white/5',
              )}
            >
              {tab.icon}
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-10 space-y-12">
        {activeTab === 'overview' && (
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
              <div className="bg-panel border border-border rounded-[28px] p-8 shadow-2xl">
                <div className="flex items-center justify-between mb-8">
                  <div className="flex flex-col">
                    <h3 className="text-xl font-black text-text uppercase tracking-widest flex items-center gap-3"><Cpu size={24} className="text-accent" /> Processor Health</h3>
                    <p className="text-sm font-bold text-text-faint mt-1 uppercase tracking-tighter">
                      {cpuInfo.physical_cores} Physical • {cpuInfo.logical_cores} Logical Cores
                    </p>
                  </div>
                  <div className="flex flex-col items-end">
                    <span className="text-3xl font-black text-text">{cpuInfo.percent.toFixed(1)}%</span>
                    <span className={cn(
                      "text-[10px] font-black px-2 py-0.5 rounded border mt-1 uppercase tracking-widest",
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
              <div className="bg-panel border border-border rounded-[28px] p-8 shadow-2xl">
                <div className="flex items-center justify-between mb-8">
                  <h3 className="text-xl font-black text-text uppercase tracking-widest flex items-center gap-3"><MemoryStick size={24} className="text-success" /> Volatile RAM</h3>
                  <span className="text-3xl font-black text-success">{memInfo.used_percent.toFixed(1)}%</span>
                </div>
                <Bar label="Physical Allocation" value={memInfo.used_percent} color="#2dd4a7" showLabel={false} />
                <div className="mt-8 pt-8 border-t border-border grid grid-cols-2 gap-6">
                  <div>
                    <p className="text-xs font-black text-text-faint uppercase mb-1">Available</p>
                    <p className="text-2xl font-black text-text">{(memInfo.total_gb - memInfo.used_gb).toFixed(2)} GB</p>
                  </div>
                  <div>
                    <p className="text-xs font-black text-text-faint uppercase mb-1">Swap Usage</p>
                    <p className="text-2xl font-black text-warning">{memInfo.swap_percent.toFixed(1)}%</p>
                  </div>
                </div>
              </div>

              {/* Disk Card */}
              <div className="bg-panel border border-border rounded-[28px] p-8 shadow-2xl">
                <div className="flex items-center justify-between mb-8">
                  <h3 className="text-xl font-black text-text uppercase tracking-widest flex items-center gap-3"><Disc size={24} className="text-accent" /> Storage Analysis</h3>
                  <span className="text-3xl font-black text-accent">
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
                <div className="flex items-center gap-6 bg-panel-2 border border-border p-6 rounded-[24px] shadow-inner">
                  <div className="relative group flex-1">
                    <Search size={24} className="absolute left-5 top-1/2 -translate-y-1/2 text-text-faint group-focus-within:text-accent transition-colors" />
                    <input
                      type="text"
                      placeholder="Filter active threads..."
                      value={search}
                      onChange={(e) => setSearch(e.target.value)}
                      className="w-full bg-panel border border-border rounded-2xl pl-16 pr-4 py-5 text-2xl font-bold text-text placeholder-text-faint focus:outline-none focus:border-accent shadow-xl"
                    />
                  </div>
                  <div className="px-8 py-4 bg-panel border border-border rounded-2xl shadow-lg">
                    <span className="text-xl font-black text-text tabular-nums">{processes.length} ACTIVE</span>
                  </div>
                </div>

                <div className="bg-panel border border-border rounded-[28px] overflow-hidden shadow-2xl">
                  <div className="max-h-[600px] overflow-y-auto">
                    <table className="w-full text-left border-collapse">
                      <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border">
                        <tr>
                          <th className="px-10 py-6 text-sm font-black text-text-dim uppercase tracking-widest">Process Name</th>
                          <th className="px-10 py-6 text-sm font-black text-text-dim uppercase tracking-widest text-right">CPU %</th>
                          <th className="px-10 py-6 text-sm font-black text-text-dim uppercase tracking-widest text-right">RAM (MB)</th>
                          <th className="px-10 py-6 text-sm font-black text-text-dim uppercase tracking-widest text-right">Impact</th>
                          <th className="px-10 py-6 w-20" />
                        </tr>
                      </thead>
                      <tbody>
                        {processes.filter(p => p.name.toLowerCase().includes(search.toLowerCase())).length === 0 ? (
                          <tr>
                            <td colSpan={5} className="px-10 py-16 text-center">
                              <p className="text-text-faint text-lg font-bold">No processes match your filter.</p>
                              <p className="text-text-faint text-sm mt-2">Try a different search term or clear the filter.</p>
                            </td>
                          </tr>
                        ) : processes.filter(p => p.name.toLowerCase().includes(search.toLowerCase())).map(p => (
                          <tr key={p.pid} className="border-b border-border/20 hover:bg-white/5 transition-all group">
                            <td className="px-10 py-5">
                              <div className="flex flex-col">
                                <span className="text-xl font-black text-text">{p.name}</span>
                                <span className="text-sm font-bold text-text-faint uppercase tracking-tighter">PID: {p.pid} • {p.status}</span>
                              </div>
                            </td>
                            <td className="px-10 py-5 text-right font-black text-2xl text-accent tabular-nums">{p.cpu.toFixed(1)}%</td>
                            <td className="px-10 py-5 text-right font-bold text-xl text-text-dim tabular-nums">{p.memory.toFixed(0)}</td>
                            <td className="px-10 py-5 text-right">
                              <span className={cn("px-4 py-1.5 rounded-full text-xs font-black uppercase border", p.cpu > 5 ? "bg-danger/10 text-danger border-danger/30" : "bg-success/10 text-success border-success/30")}>
                                {p.cpu > 5 ? 'High Impact' : 'Nominal'}
                              </span>
                            </td>
                            <td className="px-10 py-5">
                              <button onClick={() => setKillTarget({ pid: p.pid, name: p.name })} aria-label={`Kill process ${p.name} (PID ${p.pid})`} className="p-3 text-text-faint hover:text-danger hover:bg-danger/10 rounded-xl transition-all">
                                <Trash2 size={24} />
                              </button>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'system-info' && (
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-12">
            <div className="bg-panel border border-border rounded-[32px] p-10 shadow-2xl space-y-8">
              <h3 className="text-2xl font-black text-text uppercase tracking-widest flex items-center gap-4"><Disc size={32} className="text-warning" /> Operating Logic</h3>
              <div className="space-y-2">
                <InfoRow label="Hostname" value={sysInfo.hostname} copyable />
                <InfoRow label="Platform" value={sysInfo.platform} />
                <InfoRow label="Kernel" value={sysInfo.kernel_version} />
                <InfoRow label="Build" value={sysInfo.platform_version} />
              </div>
            </div>

            <div className="bg-panel border border-border rounded-[32px] p-10 shadow-2xl space-y-8">
              <h3 className="text-2xl font-black text-text uppercase tracking-widest flex items-center gap-4"><Cpu size={32} className="text-accent" /> Hardware Tier</h3>
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
    </div>
  )
}
