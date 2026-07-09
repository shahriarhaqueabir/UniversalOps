import { useState, useEffect, useCallback, useMemo } from 'react'
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
import { useBackend } from '@/hooks/useBackend'
import { useEvents } from '@/hooks/useEvents'
import { cn } from '@/lib/utils'
import type { DashboardData, AlertInfo, TimeSeriesPoint } from '@/types'

/* ───────────────────────────────────────────
   Helpers
   ─────────────────────────────────────────── */

function clamp(v: number, min = 0, max = 100) {
  return Math.min(max, Math.max(min, v))
}

function healthColor(pct: number) {
  if (pct >= 90) return 'var(--color-success)'
  if (pct >= 80) return 'var(--color-warning)'
  return 'var(--color-danger)'
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

function KpiCard({ icon, label, value, unit, status, description }: { icon: React.ReactNode, label: string, value: string, unit?: string, status: string, description: string }) {
  return (
    <div className="bg-panel border border-border rounded-[24px] p-8 transition-all hover:border-accent/40 hover:shadow-xl group">
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

import type { Page } from '@/App'

export function Dashboard({ onNavigate }: { onNavigate?: (page: Page) => void }) {
  const { call } = useBackend()
  const [data, setData] = useState<DashboardData | null>(null)
  const [cpuHistory, setCpuHistory] = useState<TimeSeriesPoint[]>([])
  // Only need setAlerts for the event callback; alerts value is unused
  const setAlerts = useState<AlertInfo[]>([])[1];

  useEffect(() => {
    call('Dashboard.GetDashboardData').then(res => setData(res as DashboardData))
  }, [call])

  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- Wails event payload is dynamic
  useEvents('metrics', useCallback((payload: any) => {
    const d = payload.data ?? payload
    setData(d)
    const t = new Date().toLocaleTimeString()
    setCpuHistory(prev => [...prev.slice(-59), { time: t, value: d.cpu.value }])
  }, []))

  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- Wails event payload is dynamic
  useEvents('alert', useCallback((payload: any) => {
    setAlerts(prev => [payload, ...prev].slice(0, 50))
  }, []))

  if (!data) return <div className="flex h-full items-center justify-center opacity-20 animate-pulse text-2xl font-black uppercase tracking-[0.3em]">Synching Neural Bridge...</div>

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
          <button onClick={() => onNavigate?.('sysops')} className="flex items-center gap-3 px-6 py-3 rounded-xl bg-panel-2 border border-border text-text font-bold hover:bg-panel-3 transition-all shadow-lg active:scale-95">
            <Zap size={18} className="text-warning" /> QUICK DIAGNOSTIC
          </button>
          <button onClick={() => onNavigate?.('aiops')} className="flex items-center gap-3 px-6 py-3 rounded-xl bg-accent text-white font-bold hover:bg-accent/90 transition-all shadow-[0_10px_20px_rgba(124,108,255,0.2)] active:scale-95">
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
          />
          <KpiCard
            icon={<MemoryStick size={24} />}
            label="Memory"
            value={Math.round(data.memory.value).toString()}
            unit="%"
            status={data.memory.value > 85 ? 'warning' : 'healthy'}
            description="Percentage of volatile allocation. High occupancy forces the system to rely on disk-based swap files."
          />
          <KpiCard
            icon={<HardDrive size={24} />}
            label="Storage"
            value={Math.round(data.disk.value).toString()}
            unit="%"
            status={data.disk.value > 90 ? 'warning' : 'healthy'}
            description="Local disk occupancy. Low headroom impacts filesystem performance and paging efficiency."
          />
        </div>
      </div>

      {/* Analysis & Guidance Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <AnalystBriefing
          title="Compute Logic Analysis"
          objective="Monitor the relationship between CPU Spikes and RAM occupancy to identify potential memory leaks or runaway background services."
          redFlags={[
            "CPU > 90% for more than 60s",
            "RAM usage climbing without CPU activity",
            "System handle count exceeding 100k",
            "NTP drift greater than 500ms"
          ]}
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
