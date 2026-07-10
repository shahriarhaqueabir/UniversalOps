import { useState, useCallback, useMemo, useRef } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Cpu,
  MemoryStick,
  HardDrive,
  Activity,
  LayoutDashboard,
  Network,
  Shield,
  Terminal,
  Brain,
  AlertTriangle,
  Info,
  Zap,
  Target,
  FileSearch,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Loader2,
  ScrollText,
} from 'lucide-react'
import { format } from 'date-fns'
import {
  AreaChart as RechartsAreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
} from 'recharts'
import { useBackend } from '@/hooks/useBackend'
import { useEvents } from '@/hooks/useEvents'
import { cn } from '@/lib/utils'
import type { DashboardData, TimeSeriesPoint } from '@/types'
import type { Page } from '@/App'

/* ── Dashboard Backend Types ── */

interface DiagnosticResult {
  category: string
  status: 'pass' | 'warn' | 'fail' | 'info'
  message: string
  value: number
  unit: string
}

interface BriefingSection {
  title: string
  content: string
  level: 'info' | 'warning' | 'critical'
}

/* ───────────────────────────────────────────
   Helpers
   ─────────────────────────────────────────── */

function clamp(v: number, min = 0, max = 100) {
  return Math.min(max, Math.max(min, v))
}

function healthColor(pct: number) {
  if (pct >= 90) return 'var(--color-danger)'
  if (pct >= 80) return 'var(--color-warning)'
  return 'var(--color-success)'
}

/* ───────────────────────────────────────────
   Enhanced Components
   ─────────────────────────────────────────── */

function AnalystBriefing({ title, objective, redFlags }: { title: string, objective: string, redFlags: string[] }) {
  return (
    <div className="bg-accent-soft border border-accent/20 rounded-[24px] p-8 shadow-xl">
      <div className="flex items-center gap-4 mb-6">
        <div className="w-12 h-12 rounded-xl bg-accent flex items-center justify-center shadow-lg">
          <Target size={28} className="text-white" />
        </div>
        <h3 className="text-2xl font-black text-text uppercase tracking-widest">{title}</h3>
      </div>
      <div className="space-y-6">
        <div>
          <p className="text-xs font-black text-accent uppercase tracking-[0.2em] mb-2">Primary Objective</p>
          <p className="text-lg text-text-dim leading-relaxed">{objective}</p>
        </div>
        <div>
          <p className="text-xs font-black text-danger uppercase tracking-[0.2em] mb-3">Critical Red-Flags</p>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {redFlags.map((flag, i) => (
              <div key={i} className="flex items-center gap-3 bg-panel-3 p-3 rounded-xl border border-border">
                <AlertTriangle size={16} className="text-warning shrink-0" />
                <span className="text-sm font-bold text-text-faint">{flag}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}

function HeroSection({ stats }: { stats: DashboardData }) {
  const avgHealth = clamp((stats.cpu.value + stats.memory.value + stats.disk.value) / 3)
  const r = 44
  const circumference = 2 * Math.PI * r
  const dash = (avgHealth / 100) * circumference
  const gap = circumference - dash

  const cpuHistoryBars = useMemo(() => {
    const hist = stats.cpu.history || []
    if (hist.length === 0) return []
    // Sample to 90 bars
    const step = Math.max(1, Math.floor(hist.length / 90))
    const sampled: number[] = []
    for (let i = 0; i < 90 && i * step < hist.length; i++) {
      sampled.push(hist[i * step])
    }
    return sampled
  }, [stats.cpu.history])

  return (
    <div className="bg-panel border-2 border-accent/20 rounded-[28px] p-10 flex flex-col lg:flex-row items-center gap-12 shadow-2xl relative overflow-hidden group">
      <div className="absolute top-0 right-0 w-64 h-64 bg-accent/5 rounded-bl-full pointer-events-none transition-all group-hover:bg-accent/10" />

      {/* ── SVG Health Donut ── */}
      <div className="relative shrink-0 scale-110">
        <svg width={140} height={140} viewBox="0 0 120 120" className="tabular-nums">
          <circle cx="60" cy="60" r={r} fill="none" stroke="var(--color-border)" strokeWidth="10" />
          <circle
            cx="60" cy="60" r={r}
            fill="none"
            stroke={healthColor(avgHealth)}
            strokeWidth="10"
            strokeLinecap="round"
            strokeDasharray={`${dash} ${gap}`}
            transform="rotate(-90 60 60)"
            style={{ transition: 'stroke-dasharray 1s ease' }}
          />
          <text x="60" y="54" textAnchor="middle" fill="var(--color-text)" fontSize="28" fontWeight="900" dominantBaseline="middle">
            {Math.round(avgHealth)}%
          </text>
          <text x="60" y="78" textAnchor="middle" fill="var(--color-text-faint)" fontSize="12" fontWeight="bold" style={{ textTransform: 'uppercase' }} dominantBaseline="middle">
            HEALTH
          </text>
        </svg>
      </div>

      <div className="flex-1 min-w-0 w-full">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-4">
            <div className="px-4 py-1.5 rounded-full bg-success/10 border border-success/30 text-xs font-black text-success uppercase tracking-widest">System Nominal</div>
            {stats.alerts > 0 && (
              <div className="px-4 py-1.5 rounded-full bg-warning/10 border border-warning/30 text-xs font-black text-warning uppercase tracking-widest flex items-center gap-2">
                <AlertTriangle size={12} /> {stats.alerts} Active Alerts
              </div>
            )}
            <span className="text-3xl font-black text-text">Platform Vitality</span>
          </div>
          <div className="text-right">
            <p className="text-xs font-black text-text-faint uppercase tracking-widest">Active Uptime</p>
            <p className="text-xl font-bold text-accent tabular-nums">{stats.uptime}</p>
          </div>
        </div>

        <p className="text-text-dim text-xl leading-relaxed mb-8 max-w-2xl font-medium">
          Aggregate performance score derived from core compute, volatile memory, and storage throughput.
          Current status indicates <span className="text-text font-black">Stable Operation</span> across all monitored subsystems.
        </p>

        <div className="flex gap-1 items-end h-12 w-full bg-panel-3 p-2 rounded-xl border border-border shadow-inner">
          {cpuHistoryBars.length > 0 ? cpuHistoryBars.map((val, i) => (
            <div
              key={i}
              className="flex-1 rounded-[1px] transition-all hover:scale-y-125 hover:opacity-80"
              style={{
                height: `${Math.max(5, val)}%`,
                backgroundColor: healthColor(val),
                opacity: val > 80 ? 0.9 : 0.5,
              }}
            />
          )) : (
            <div className="w-full text-center text-xs font-bold text-text-faint">No history data</div>
          )}
        </div>
      </div>
    </div>
  )
}

function KpiCard({ icon, label, value, unit, status, description, onClick }: { icon: React.ReactNode, label: string, value: string, unit?: string, status: string, description: string, onClick?: () => void }) {
  return (
    <div
      onClick={onClick}
      className={cn(
        "bg-panel border border-border rounded-[24px] p-8 transition-all hover:border-accent/40 hover:shadow-xl group",
        onClick ? "cursor-pointer active:scale-95" : ""
      )}
    >
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-4">
          <div className="w-12 h-12 rounded-xl bg-panel-3 flex items-center justify-center text-text-faint group-hover:text-accent transition-colors border border-border shadow-inner">
            {icon}
          </div>
          <span className="text-sm font-black text-text-dim uppercase tracking-[0.2em]">{label}</span>
        </div>
        <div className={cn("w-3 h-3 rounded-full shadow-lg", status === 'healthy' ? "bg-success" : "bg-warning")} />
      </div>
      <div className="flex items-baseline gap-2 mb-4">
        <span className="text-5xl font-black text-text tabular-nums">{value}</span>
        {unit && <span className="text-xl font-bold text-text-faint">{unit}</span>}
      </div>
      <p className="text-base text-text-faint leading-relaxed border-t border-border/50 pt-4 font-medium italic">
        {description}
      </p>
    </div>
  )
}

/* ───────────────────────────────────────────
   Main Dashboard
   ─────────────────────────────────────────── */

function diagIcon(status: string) {
  switch (status) {
    case 'pass': return <CheckCircle2 size={20} className="text-success" />
    case 'warn': return <AlertCircle size={20} className="text-warning" />
    case 'fail': return <XCircle size={20} className="text-danger" />
    default: return <Info size={20} className="text-accent" />
  }
}

function diagColor(status: string) {
  switch (status) {
    case 'pass': return 'border-success/30 bg-success/5'
    case 'warn': return 'border-warning/30 bg-warning/5'
    case 'fail': return 'border-danger/30 bg-danger/5'
    default: return 'border-accent/30 bg-accent/5'
  }
}

function computeRedFlags(stats: DashboardData): string[] {
  const flags: string[] = []
  if (stats.cpu.value > 90) flags.push(`CPU usage at ${Math.round(stats.cpu.value)}% — critically high for sustained periods`)
  else if (stats.cpu.trend === 'rising' && stats.cpu.value > 70) flags.push(`CPU trend rising at ${Math.round(stats.cpu.value)}% — monitor for potential bottlenecks`)
  if (stats.memory.value > 92) flags.push(`Memory at ${Math.round(stats.memory.value)}% — swap pressure likely`)
  else if (stats.memory.trend === 'rising' && stats.memory.value > 80) flags.push(`Memory usage climbing at ${Math.round(stats.memory.value)}% — possible leak`)
  if (stats.disk.value > 95) flags.push(`Disk at ${Math.round(stats.disk.value)}% — critical, immediate cleanup needed`)
  else if (stats.disk.value > 85) flags.push(`Disk at ${Math.round(stats.disk.value)}% — low headroom affects performance`)
  if (stats.alerts > 0) flags.push(`${stats.alerts} active alert(s) require attention`)
  if (flags.length === 0) flags.push('All metrics within normal operating parameters')
  return flags
}

export function Dashboard({ onNavigate }: { onNavigate?: (page: Page) => void }) {
  const { call } = useBackend()
  const [data, setData] = useState<DashboardData | null>(null)
  const [cpuHistory, setCpuHistory] = useState<TimeSeriesPoint[]>([])
  const alertCount = useRef(0)

  // Quick Diag state
  const [diagOpen, setDiagOpen] = useState(false)
  const [diagResults, setDiagResults] = useState<DiagnosticResult[]>([])
  const [diagLoading, setDiagLoading] = useState(false)

  // Briefing state
  const [briefingOpen, setBriefingOpen] = useState(false)
  const [briefingSections, setBriefingSections] = useState<BriefingSection[]>([])
  const [briefingLoading, setBriefingLoading] = useState(false)

  // Initial data load via react-query (cached, deduped)
  useQuery<DashboardData>({
    queryKey: ['dashboard'],
    queryFn: async () => {
      const res = await call('Dashboard.GetDashboardData') as DashboardData
      setData(res)
      return res
    },
    staleTime: 10000,
    refetchOnWindowFocus: false,
  })

  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- Wails event payload is dynamic
  useEvents('metrics', useCallback((payload: any) => {
    const d = payload.data ?? payload
    setData(d)
    const t = format(new Date(), 'HH:mm:ss')
    setCpuHistory(prev => [...prev.slice(-59), { time: t, value: d.cpu.value }])
  }, []))

  useEvents('alert', useCallback(() => {
    alertCount.current++
    // handler signature matches useEvents hook
  }, []))

  const runQuickDiag = async () => {
    setDiagOpen(true)
    setDiagLoading(true)
    setDiagResults([])
    try {
      const res = await call('Dashboard.RunQuickDiag') as DiagnosticResult[]
      setDiagResults(res)
    } catch (err) {
      setDiagResults([{ category: 'Error', status: 'fail', message: String(err), value: 0, unit: '' }])
    } finally {
      setDiagLoading(false)
    }
  }

  const generateBriefing = async () => {
    setBriefingOpen(true)
    setBriefingLoading(true)
    setBriefingSections([])
    try {
      const res = await call('Dashboard.GenerateDashboardBriefing') as BriefingSection[]
      setBriefingSections(res)
    } catch (err) {
      setBriefingSections([{ title: 'Error', content: String(err), level: 'critical' }])
    } finally {
      setBriefingLoading(false)
    }
  }

  if (!data) return (
    <div className="p-10 space-y-12 overflow-y-auto h-full">
      <div className="animate-pulse space-y-6">
        <div className="h-10 w-64 bg-panel-2 rounded-xl" />
        <div className="h-6 w-96 bg-panel-2 rounded-lg" />
        <div className="h-48 bg-panel-2 rounded-[28px]" />
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          <div className="h-48 bg-panel-2 rounded-[24px]" />
          <div className="h-48 bg-panel-2 rounded-[24px]" />
          <div className="h-48 bg-panel-2 rounded-[24px]" />
        </div>
      </div>
    </div>
  )

  return (
    <div className="p-10 space-y-12 overflow-y-auto h-full scroll-smooth">
      {/* Header Strategy */}
      <div className="flex items-center justify-between border-b border-border pb-8">
        <div>
          <h1 className="text-4xl font-black text-text flex items-center gap-4">
            <LayoutDashboard size={44} className="text-accent" /> OPERATIONAL INTELLIGENCE
          </h1>
          <p className="text-xl text-text-dim mt-2 font-medium">Strategic overview of system-wide heuristics and critical subsystems.</p>
        </div>
        <div className="flex gap-4">
          <button onClick={runQuickDiag} className="flex items-center gap-3 px-6 py-3 rounded-xl bg-panel-2 border border-border text-text font-bold hover:bg-panel-3 transition-all shadow-lg active:scale-95">
            <Zap size={18} className="text-warning" /> QUICK DIAGNOSTIC
          </button>
          <button onClick={generateBriefing} className="flex items-center gap-3 px-6 py-3 rounded-xl bg-accent text-white font-bold hover:bg-accent/90 transition-all shadow-[0_10px_20px_rgba(124,108,255,0.2)] active:scale-95">
            <FileSearch size={18} /> GENERATE BRIEFING
          </button>
        </div>
      </div>

      <HeroSection stats={data} />

      {/* Structural Categories */}
      <div className="space-y-8">
        <div className="flex items-center gap-4">
          <div className="h-px flex-1 bg-border" />
          <h2 className="text-xs font-black text-text-faint uppercase tracking-[0.4em]">Resource Compute Layer</h2>
          <div className="h-px flex-1 bg-border" />
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          <KpiCard
            icon={<Cpu size={24} />}
            label="Processor"
            value={Math.round(data.cpu.value).toString()}
            unit="%"
            status={data.cpu.value > 80 ? 'warning' : 'healthy'}
            description="Measures aggregate clock-cycle pressure. Sustained high usage indicates thread-pool exhaustion."
            onClick={() => onNavigate?.('sysops')}
          />
          <KpiCard
            icon={<MemoryStick size={24} />}
            label="Memory"
            value={Math.round(data.memory.value).toString()}
            unit="%"
            status={data.memory.value > 85 ? 'warning' : 'healthy'}
            description="Percentage of volatile allocation. High occupancy forces the system to rely on disk-based swap files."
            onClick={() => onNavigate?.('sysops')}
          />
          <KpiCard
            icon={<HardDrive size={24} />}
            label="Storage"
            value={Math.round(data.disk.value).toString()}
            unit="%"
            status={data.disk.value > 90 ? 'warning' : 'healthy'}
            description="Local disk occupancy. Low headroom impacts filesystem performance and paging efficiency."
            onClick={() => onNavigate?.('sysops')}
          />
        </div>
      </div>

      {/* Analysis & Guidance Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <AnalystBriefing
          title="Compute Logic Analysis"
          objective="Monitor the relationship between CPU Spikes and RAM occupancy to identify potential memory leaks or runaway background services."
          redFlags={computeRedFlags(data)}
        />

        <div className="bg-panel border border-border rounded-[28px] p-8 shadow-2xl flex flex-col">
          <h3 className="text-xl font-black text-text uppercase tracking-widest mb-8 flex items-center gap-3">
            <Activity size={24} className="text-accent" /> Compute Timeline
          </h3>
          <div className="flex-1 min-h-[240px]">
            <ResponsiveContainer width="100%" height="100%">
              <RechartsAreaChart data={cpuHistory}>
                <defs>
                  <linearGradient id="cpuGrad" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="var(--color-accent)" stopOpacity={0.3} />
                    <stop offset="100%" stopColor="var(--color-accent)" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="4 4" stroke="var(--color-border)" vertical={false} strokeOpacity={0.5} />
                <XAxis dataKey="time" hide />
                <YAxis hide domain={[0, 100]} />
                <Tooltip contentStyle={{ backgroundColor: 'var(--color-panel-3)', border: 'none', borderRadius: '12px' }} />
                <Area type="monotone" dataKey="value" stroke="var(--color-accent)" strokeWidth={4} fill="url(#cpuGrad)" isAnimationActive={false} />
              </RechartsAreaChart>
            </ResponsiveContainer>
          </div>
          <p className="mt-6 text-sm font-bold text-text-faint italic flex items-center gap-2">
            <Info size={14} /> Statistical Trend: {data.cpu.trend.toUpperCase()}
          </p>
        </div>
      </div>

      {/* Layer Launchpad */}
      <div className="space-y-8">
        <div className="flex items-center gap-4">
          <div className="h-px flex-1 bg-border" />
          <h2 className="text-xs font-black text-text-faint uppercase tracking-[0.4em]">Functional Subsystems</h2>
          <div className="h-px flex-1 bg-border" />
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-6">
          {opsItems.map((item) => (
            <button
              key={item.id}
              onClick={() => onNavigate?.(item.id)}
              className="bg-panel-2 border border-border/50 rounded-[20px] p-6 transition-all duration-300 hover:bg-accent-soft hover:border-accent/40 group relative overflow-hidden text-left w-full"
            >
              <div className="absolute -bottom-4 -right-4 opacity-5 group-hover:opacity-10 transition-all group-hover:scale-150">
                {item.icon}
              </div>
              <div className="flex flex-col gap-4 relative z-10">
                <div className="w-10 h-10 rounded-xl bg-panel-3 border border-border flex items-center justify-center text-accent shadow-inner">
                  {item.icon}
                </div>
                <div>
                  <h4 className="text-lg font-black text-text mb-1">{item.label}</h4>
                  <p className="text-xs font-bold text-text-faint uppercase tracking-tighter group-hover:text-accent transition-colors">Launch Module</p>
                </div>
              </div>
            </button>
          ))}
        </div>
      </div>

      {/* ── Quick Diagnostic Overlay ── */}
      {diagOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={() => setDiagOpen(false)}>
          <div className="bg-panel border border-border rounded-[24px] p-8 w-full max-w-2xl max-h-[80vh] overflow-y-auto shadow-2xl mx-4" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-2xl font-black text-text flex items-center gap-3">
                <Zap size={24} className="text-warning" /> Quick Diagnostic Results
              </h2>
              <button onClick={() => setDiagOpen(false)} className="text-text-faint hover:text-text transition-colors">
                <XCircle size={24} />
              </button>
            </div>
            {diagLoading ? (
              <div className="flex items-center justify-center py-16">
                <Loader2 size={32} className="text-accent animate-spin" />
                <span className="ml-4 text-lg font-bold text-text-faint">Running diagnostics...</span>
              </div>
            ) : (
              <div className="space-y-4">
                {diagResults.map((r, i) => (
                  <div key={i} className={`flex items-start gap-4 p-4 rounded-xl border ${diagColor(r.status)}`}>
                    <div className="mt-1">{diagIcon(r.status)}</div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-3 mb-1">
                        <span className="text-sm font-black text-text uppercase tracking-widest">{r.category}</span>
                        <span className="text-xs font-bold text-text-faint">{r.value.toFixed(1)}{r.unit}</span>
                      </div>
                      <p className="text-sm text-text-dim">{r.message}</p>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}

      {/* ── Briefing Overlay ── */}
      {briefingOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={() => setBriefingOpen(false)}>
          <div className="bg-panel border border-border rounded-[24px] p-8 w-full max-w-3xl max-h-[80vh] overflow-y-auto shadow-2xl mx-4" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-2xl font-black text-text flex items-center gap-3">
                <ScrollText size={24} className="text-accent" /> Operations Briefing
              </h2>
              <button onClick={() => setBriefingOpen(false)} className="text-text-faint hover:text-text transition-colors">
                <XCircle size={24} />
              </button>
            </div>
            {briefingLoading ? (
              <div className="flex items-center justify-center py-16">
                <Loader2 size={32} className="text-accent animate-spin" />
                <span className="ml-4 text-lg font-bold text-text-faint">Synthesizing briefing...</span>
              </div>
            ) : briefingSections.length === 0 ? (
              <p className="text-text-faint text-center py-8">No briefing data available.</p>
            ) : (
              <div className="space-y-6">
                {briefingSections.map((s, i) => (
                  <div key={i} className="p-5 rounded-xl border border-border bg-panel-2">
                    <div className="flex items-center gap-3 mb-2">
                      {s.level === 'critical' ? <AlertTriangle size={18} className="text-danger" /> :
                        s.level === 'warning' ? <AlertCircle size={18} className="text-warning" /> :
                          <Info size={18} className="text-accent" />}
                      <h3 className="text-base font-black text-text uppercase tracking-widest">{s.title}</h3>
                    </div>
                    <p className="text-sm text-text-dim whitespace-pre-wrap leading-relaxed">{s.content}</p>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

const opsItems: { id: Page; label: string; icon: React.ReactNode }[] = [
  { id: 'sysops', label: 'System', icon: <Cpu size={20} /> },
  { id: 'netops', label: 'Network', icon: <Network size={20} /> },
  { id: 'secops', label: 'Security', icon: <Shield size={20} /> },
  { id: 'devops', label: 'DevOps', icon: <Terminal size={20} /> },
  { id: 'aiops', label: 'AI Analyst', icon: <Brain size={20} /> },
]
