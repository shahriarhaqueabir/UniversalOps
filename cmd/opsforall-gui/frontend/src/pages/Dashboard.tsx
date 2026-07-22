import { useState, useMemo, useEffect, memo } from 'react'
import { motion } from 'motion/react'
import * as Dialog from '@radix-ui/react-dialog'
import { useQuery } from '@tanstack/react-query'
import {
  Cpu,
  MemoryStick,
  HardDrive,
  Gpu,
  Battery,
  Activity,
  LayoutDashboard,
  Network,
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
  Clock,
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
import { useSettingsStore, useMetricsStore, useNavigationStore } from '@/stores'
import { cn } from '@/lib/utils'
import { DataFreshnessIndicator } from '@/components/ui/DataFreshnessIndicator'
import { Panel } from '@/components/ui/Panel'
import type { DashboardData } from '@/types'
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

interface TimelineEvent {
  id: string
  timestamp: string
  category: string
  level: string
  title: string
  detail: string
  module: string
}

interface SystemSnapshot {
  metrics: DashboardData
  alerts: any[]
  timeline: TimelineEvent[]
  timestamp: string
}

function diagColor(status: string): string {
  switch (status) {
    case 'pass': return 'border-green-500/30 bg-green-500/5'
    case 'warn': return 'border-yellow-500/30 bg-yellow-500/5'
    case 'fail': return 'border-red-500/30 bg-red-500/5'
    default: return 'border-[var(--color-border)] bg-[var(--color-panel-2)]'
  }
}

function diagIcon(status: string): React.ReactNode {
  switch (status) {
    case 'pass': return <CheckCircle2 size={18} className="text-green-500 shrink-0" />
    case 'warn': return <AlertTriangle size={18} className="text-yellow-500 shrink-0" />
    case 'fail': return <XCircle size={18} className="text-red-500 shrink-0" />
    default: return <Info size={18} className="text-[var(--color-accent)] shrink-0" />
  }
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

// function formatBytes removed — unused

/* ───────────────────────────────────────────
   Enhanced Components
   ─────────────────────────────────────────── */

const AnalystBriefing = memo(function AnalystBriefing({ title, objective }: { title: string, objective: string }) {
  const latest = useMetricsStore(s => s.latest)
  const alerts = useMetricsStore(s => s.alerts)

  const redFlags = useMemo(() => {
    if (!latest) return ['No system data available for analysis']
    const flags: string[] = []
    const stats = latest
    if (stats.cpu?.value > 90) flags.push(`CPU usage at ${Math.round(stats.cpu.value)}%: critically high`)
    else if (stats.cpu?.trend === 'rising' && stats.cpu?.value > 70) flags.push(`CPU trend rising at ${Math.round(stats.cpu.value)}%`)
    if (stats.memory?.value > 92) flags.push(`Memory at ${Math.round(stats.memory.value)}%: swap pressure likely`)
    if (stats.disk?.value > 95) flags.push(`Disk at ${Math.round(stats.disk.value)}%: critical cleanup needed`)
    if (alerts.length > 0) flags.push(`${alerts.length} active alert(s) require attention`)
    if (flags.length === 0) flags.push('All metrics within normal operating parameters')
    return flags
  }, [latest, alerts])

  return (
    <Panel padding="md" className="group hover:border-[var(--color-accent)]/30 flex flex-col h-full rounded-[2rem]">
      <div className="flex items-center gap-3 mb-6">
        <div className="w-10 h-10 rounded-2xl bg-accent/10 border border-accent/20 flex items-center justify-center text-accent shadow-sm group-hover:scale-110 transition-transform">
          <Target size={20} />
        </div>
        <div>
          <h3 className="text-xs font-black text-accent uppercase tracking-[0.2em] text-wrap-balance">{title}</h3>
          <p className="text-[10px] text-text-faint font-bold uppercase tracking-wider">AI Heuristic Summary</p>
        </div>
      </div>
      <div className="space-y-6 flex-1">
        <div className="relative pl-4 border-l-2 border-accent/20">
          <p className="text-[10px] font-black text-text-faint uppercase tracking-wider mb-2">Primary Objective</p>
          <p className="text-sm text-text-dim leading-relaxed font-medium text-wrap-pretty">{objective}</p>
        </div>
        <div>
          <p className="text-[10px] font-black text-danger uppercase tracking-wider mb-3 flex items-center gap-2">
            <AlertTriangle size={12} /> Critical Red-Flags
          </p>
          <div className="grid grid-cols-1 gap-2">
            {redFlags.map((flag, i) => (
              <div key={i} className="flex items-start gap-3 bg-panel-2 p-4 rounded-2xl border border-border group/flag hover:border-danger/30 transition-colors">
                <div className="mt-0.5 w-1.5 h-1.5 rounded-full bg-danger/40 group-hover/flag:bg-danger transition-colors shrink-0" />
                <span className="text-xs font-semibold text-text-dim group-hover/flag:text-text transition-colors tabular-nums">{flag}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </Panel>
  )
})

const HeroSection = memo(function HeroSection() {
  const stats = useMetricsStore(s => s.latest)
  const alerts = useMetricsStore(s => s.alerts)
  const cpuHistory = useMetricsStore(s => s.history.cpu)
  const { navigate } = useNavigationStore()

  const cpuHistoryBars = useMemo(() => {
    if (cpuHistory.length === 0) return []
    const step = Math.max(1, Math.floor(cpuHistory.length / 90))
    const sampled: number[] = []
    for (let i = 0; i < 90 && i * step < cpuHistory.length; i++) {
      sampled.push(cpuHistory[i * step].value)
    }
    return sampled
  }, [cpuHistory])

  if (!stats) return null

  const avgHealth = clamp(((stats.cpu?.value ?? 0) + (stats.memory?.value ?? 0) + (stats.disk?.value ?? 0)) / 3)
  const r = 44
  const circumference = 2 * Math.PI * r
  const dash = (avgHealth / 100) * circumference
  const gap = circumference - dash

  const alertBreakdown = {
    critical: alerts.filter(a => a.level === 'CRITICAL').length,
    warning: alerts.filter(a => a.level === 'WARNING').length,
    info: alerts.filter(a => a.level === 'INFO').length
  }

  return (
    <div
      onClick={() => navigate('alerts')}
      className="bg-panel border-2 border-accent/20 rounded-[var(--radius-xl)] p-6 flex flex-row items-center gap-6 shadow-2xl relative overflow-hidden group animate-in fade-in slide-in-from-bottom-4 duration-700 cursor-pointer hover:border-accent/40 transition-all active:scale-[0.99]"
    >
      <div className="absolute top-0 right-0 w-64 h-64 bg-[var(--color-accent)]/5 rounded-bl-full pointer-events-none transition-all group-hover:bg-[var(--color-accent)]/10" />

      <div className="relative shrink-0">
        <svg width={140} height={140} viewBox="0 0 120 120" className="tabular-nums drop-shadow-[0_0_10px_rgba(var(--color-accent-rgb),0.2)]">
          <circle cx="60" cy="60" r={r} fill="none" stroke="var(--color-border)" strokeWidth="10" />
          <circle
            cx="60" cy="60" r={r}
            fill="none"
            stroke={healthColor(avgHealth)}
            strokeWidth="10"
            strokeLinecap="round"
            strokeDasharray={`${dash} ${gap}`}
            transform="rotate(-90 60 60)"
            style={{ transition: 'stroke-dasharray 1.2s cubic-bezier(0.34, 1.56, 0.64, 1)' }}
          />
          <text x="60" y="54" textAnchor="middle" fill="var(--color-text)" fontSize="28" fontWeight="900" dominantBaseline="middle" className="tabular-nums">
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
            <span className="px-3 py-1 rounded-full bg-[var(--color-success)]/10 border border-[var(--color-success)]/30 text-xs font-semibold text-[var(--color-success)] uppercase tracking-wider">System Nominal</span>
            {alerts.length > 0 && (
              <div className="flex items-center gap-2">
                {alertBreakdown.critical > 0 && (
                  <div className="px-3 py-1.5 rounded-full bg-danger/10 border border-danger/30 text-xs font-bold text-danger uppercase tracking-wider flex items-center gap-1.5 animate-pulse">
                    <AlertTriangle size={11} /> {alertBreakdown.critical} Critical
                  </div>
                )}
                {alertBreakdown.warning > 0 && (
                  <div className="px-3 py-1.5 rounded-full bg-warning/10 border border-warning/30 text-xs font-bold text-warning uppercase tracking-wider flex items-center gap-1.5">
                    <AlertCircle size={11} /> {alertBreakdown.warning} Warning
                  </div>
                )}
              </div>
            )}
            <span className="text-2xl font-bold text-[var(--color-text)]">System Health Score</span>
          </div>
          <div className="text-right">
            <p className="text-xs font-semibold text-[var(--color-text-faint)] uppercase tracking-widest">Active Uptime</p>
            <p className="text-xl font-bold text-accent tabular-nums">{stats.uptime}</p>
          </div>
        </div>

        <p className="text-[var(--color-text-dim)] text-sm leading-relaxed mb-6 max-w-2xl">
          Aggregate performance score derived from core compute, volatile memory, and storage throughput.
          Current status indicates <span className="text-[var(--color-text)] font-semibold">Stable Operation</span> across all monitored subsystems.
        </p>

        <div className="flex gap-1 items-end h-12 w-full bg-panel-3 p-2 rounded-xl border border-border shadow-inner">
          {cpuHistoryBars.length > 0 ? cpuHistoryBars.map((val, i) => (
            <div
              key={i}
              className="flex-1 rounded-[2px] transition-all"
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
})

const KpiCard = memo(function KpiCard({ icon, label, metricKey, unit, description, onClick, variant = 'default' }: { icon: React.ReactNode, label: string, metricKey: keyof DashboardData, unit?: string, description: string, onClick?: () => void, variant?: 'default' | 'accent' | 'success' | 'warning' }) {
  const data = useMetricsStore(s => (s.latest as any)?.[metricKey])

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if ((e.key === 'Enter' || e.key === ' ') && onClick) {
      e.preventDefault();
      onClick();
    }
  };
const variantStyles = {
    default: 'bg-[var(--color-panel)] border-[var(--color-border)]',
    accent: 'bg-[var(--color-accent)]/5 border-[var(--color-accent)]/20',
    success: 'bg-[var(--color-success)]/5 border-[var(--color-success)]/20',
    warning: 'bg-[var(--color-warning)]/5 border-[var(--color-warning)]/20',
  };

  const val = data?.value != null ? Math.round(data.value).toString() : '0'
  const displayVal = (metricKey === 'gpu' || metricKey === 'battery')
    ? (data?.detected ? (metricKey === 'gpu' ? data.vendor : val) : 'N/A')
    : val

  const status = (data?.value > 80) ? 'warning' : 'healthy'

  return (
    <div
      onClick={onClick}
      onKeyDown={handleKeyDown}
      role={onClick ? 'button' : undefined}
      tabIndex={onClick ? 0 : undefined}
      className={cn(
        "rounded-[var(--radius-lg)] p-6 transition-all hover:border-[var(--color-accent)]/30 hover:shadow-lg group card-hover focus-visible:outline-2 focus-visible:outline-[var(--color-accent)] focus-visible:outline-offset-2",
        onClick ? "cursor-pointer active:scale-[0.97]" : "",
        variantStyles[variant]
      )}
    >
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-lg bg-[var(--color-panel-3)] flex items-center justify-center text-[var(--color-text-faint)] group-hover:text-[var(--color-accent)] transition-colors border border-[var(--color-border)]">
            {icon}
          </div>
          <span className="text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">{label}</span>
        </div>
        <div className={cn("w-2.5 h-2.5 rounded-full transition-all duration-500", status === 'healthy' ? "bg-[var(--color-success)] shadow-[0_0_8px_var(--color-success)]" : "bg-[var(--color-warning)] shadow-[0_0_8px_var(--color-warning)]")} />
      </div>
      <div className="flex items-baseline gap-1.5 mb-3">
        <span className="text-3xl font-bold text-[var(--color-text)] tabular-nums">{displayVal}</span>
        {unit && <span className="text-lg font-semibold text-[var(--color-text-faint)]">{unit}</span>}
      </div>
      <p className="text-xs text-[var(--color-text-faint)] leading-relaxed border-t border-[var(--color-border)]/50 pt-3">
        {description}
      </p>
    </div>
  )
})

const RecentEventsPanel = memo(function RecentEventsPanel({ onExplain, explainingId }: { onExplain: (id: string) => void, explainingId: string | null }) {
  const timelineEvents = useMetricsStore(s => s.timeline)

  return (
    <Panel padding="md">
      <h3 className="text-base font-bold text-[var(--color-text)] uppercase tracking-wider mb-4 flex items-center gap-2">
        <Clock size={18} className="text-[var(--color-accent)]" /> Recent Events
      </h3>
      <div className="space-y-2 max-h-[320px] overflow-y-auto pr-1">
        {timelineEvents.map((evt) => (
          <div key={evt.id} className="flex items-start gap-3 p-4 rounded-lg bg-[var(--color-panel-2)] border border-[var(--color-border)]/50">
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 mb-0.5">
                <span className="px-1.5 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-[var(--color-panel-3)] text-[var(--color-text-faint)] border border-[var(--color-border)]/50">{evt.category}</span>
                <span className="text-[10px] text-[var(--color-text-faint)] tabular-nums ml-auto">{evt.timestamp ? format(new Date(evt.timestamp), 'HH:mm:ss') : ''}</span>
              </div>
              <p className="text-xs font-semibold text-[var(--color-text)] leading-snug">{evt.title}</p>
              <div className="flex items-center justify-between gap-4 mt-1">
                {evt.detail && <p className="text-[11px] text-[var(--color-text-dim)] leading-relaxed truncate flex-1">{evt.detail}</p>}
                <button
                  onClick={() => onExplain(evt.id)}
                  disabled={explainingId === evt.id}
                  className="text-[10px] font-black uppercase tracking-tighter text-accent hover:text-accent/80 flex items-center gap-1 shrink-0 bg-accent/5 px-2 py-0.5 rounded border border-accent/10"
                >
                  {explainingId === evt.id ? <Loader2 size={10} className="animate-spin" /> : <Brain size={10} />}Explain
                </button>
              </div>
            </div>
          </div>
        ))}
        {timelineEvents.length === 0 && (
          <div className="text-center py-10 opacity-20">
            <p className="text-xs font-bold uppercase tracking-widest">Waiting for events...</p>
          </div>
        )}
      </div>
    </Panel>
  )
})

const containerVariants = {
  hidden: { opacity: 0 },
  show: {
    opacity: 1,
    transition: { staggerChildren: 0.1, delayChildren: 0.1 }
  }
}

const itemVariants = {
  hidden: { opacity: 0, y: 15 },
  show: { opacity: 1, y: 0, transition: { duration: 0.5, ease: 'easeInOut' as const } }
}

export function Dashboard({ onNavigate }: { onNavigate?: (page: Page, tab?: string | null) => void }) {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const { navigate } = useNavigationStore()
  const setSnapshot = useMetricsStore(s => s.setSnapshot)
  const cpuHistory = useMetricsStore(s => s.history.cpu)
  const latest = useMetricsStore(s => s.latest)
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null)

  const { data: snapshotData, isLoading: queryLoading } = useQuery<SystemSnapshot>({
    queryKey: ['dashboard-snapshot'],
    queryFn: async () => {
      const res = await call('Dashboard.GetSystemSnapshot') as SystemSnapshot
      return res
    },
    staleTime: refreshInterval,
    refetchInterval: refreshInterval,
    refetchOnWindowFocus: false,
  })

  // Sync snapshot to metrics store AFTER render (breaking the setState-in-render loop)
  useEffect(() => {
    if (snapshotData) {
      setSnapshot(snapshotData)
      setLastUpdated(new Date())
    }
  }, [snapshotData, setSnapshot])

  const [diagOpen, setDiagOpen] = useState(false)
  const [diagResults, setDiagResults] = useState<DiagnosticResult[]>([])
  const [diagLoading, setDiagLoading] = useState(false)

  const runQuickDiag = async () => {
    setDiagOpen(true); setDiagLoading(true); setDiagResults([])
    try { setDiagResults(await call('Dashboard.RunQuickDiag') as DiagnosticResult[]) }
    catch (err: unknown) { setDiagResults([{ category: 'Error', status: 'fail', message: String(err), value: 0, unit: '' }]) }
    finally { setDiagLoading(false) }
  }

  const [briefingOpen, setBriefingOpen] = useState(false)
  const [briefingSections, setBriefingSections] = useState<BriefingSection[]>([])
  const [briefingLoading, setBriefingLoading] = useState(false)

  const generateBriefing = async () => {
    setBriefingOpen(true); setBriefingLoading(true); setBriefingSections([])
    try { setBriefingSections(await call('Dashboard.GenerateDashboardBriefing') as BriefingSection[]) }
    catch (err: unknown) { setBriefingSections([{ title: 'Error', content: String(err), level: 'critical' }]) }
    finally { setBriefingLoading(false) }
  }

  const [explanationOpen, setExplanationOpen] = useState(false)
  const [explanationText, setExplanationText] = useState('')
  const [explainingId, setExplainingId] = useState<string | null>(null)

  const handleExplain = async (eventId: string) => {
    setExplainingId(eventId); setExplanationOpen(true); setExplanationText('')
    try { setExplanationText(await call('Timeline.ExplainEvents', [eventId]) as string) }
    catch (err: unknown) { setExplanationText(`Analysis Error: ${String(err)}`) }
    finally { setExplainingId(null) }
  }

  if (queryLoading && !latest) return (
    <div className="space-y-12 animate-pulse p-10">
      <div className="h-10 w-64 bg-panel-2 rounded-xl" />
      <div className="h-48 bg-panel-2 rounded-[28px]" />
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-8"><div className="h-48 bg-panel-2 rounded-[24px]" /><div className="h-48 bg-panel-2 rounded-[24px]" /><div className="h-48 bg-panel-2 rounded-[24px]" /></div>
    </div>
  )

  if (!latest) return null

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="space-y-10 p-10 pb-12 scroll-smooth max-w-[1400px] mx-auto"
    >
      <motion.div variants={itemVariants} className="flex items-end justify-between border-b border-[var(--color-border)] pb-8 mb-4">
        <div>
          <div className="flex items-center gap-3 mb-2">
            <div className="w-8 h-8 rounded-lg bg-accent/10 flex items-center justify-center text-accent border border-accent/20">
               <LayoutDashboard size={18} />
            </div>
            <h1 className="text-sm font-black text-[var(--color-text)] uppercase tracking-[0.25em]">Operational Intelligence</h1>
          </div>
          <p className="text-3xl font-bold text-[var(--color-text)] tracking-tight">System Overview</p>
          <DataFreshnessIndicator lastUpdated={lastUpdated} className="mt-4" />
        </div>
        <div className="flex gap-4">
          <button onClick={runQuickDiag} className="flex items-center gap-2.5 px-6 py-3 rounded-xl bg-white/5 border border-white/10 text-[var(--color-text)] text-sm font-bold hover:bg-white/10 transition-all active:scale-95 shadow-lg group">
            <Zap size={16} className="text-warning group-hover:scale-110 transition-transform" />
            Run Diagnostics
          </button>
          <button onClick={generateBriefing} className="flex items-center gap-2.5 px-6 py-3 rounded-xl bg-accent text-white text-sm font-bold hover:opacity-90 transition-all active:scale-95 shadow-[0_0_20px_rgba(var(--color-accent-rgb),0.3)]">
            <FileSearch size={16} />
            Generate Briefing
          </button>
        </div>
      </motion.div>

      <motion.div variants={itemVariants}>
        <HeroSection />
      </motion.div>

      <motion.div variants={itemVariants} className="space-y-5">
        <div className="flex items-center gap-4"><div className="h-px flex-1 bg-[var(--color-border)]" /><h2 className="text-xs font-semibold text-[var(--color-text-faint)] uppercase tracking-widest">Resource Compute Layer</h2><div className="h-px flex-1 bg-[var(--color-border)]" /></div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-5">
           <KpiCard icon={<Cpu size={24} />} label="Processor" metricKey="cpu" unit="%" onClick={() => navigate('sysops', 'cpu')} variant="accent" description="Measures aggregate clock-cycle pressure." />
           <KpiCard icon={<MemoryStick size={24} />} label="Memory" metricKey="memory" unit="%" onClick={() => navigate('sysops', 'memory')} variant="success" description="Percentage of volatile allocation." />
           <KpiCard icon={<HardDrive size={24} />} label="Storage" metricKey="disk" unit="%" onClick={() => navigate('sysops', 'disk')} variant="warning" description="Local disk occupancy." />
           <KpiCard icon={<Gpu size={24} />} label="GPU" metricKey="gpu" onClick={() => navigate('sysops', 'hardware')} variant="default" description={latest.gpu?.detected ? `${latest.gpu.name}` : 'No GPU detected.'} />
           <KpiCard icon={<Battery size={24} />} label="Battery" metricKey="battery" unit="%" onClick={() => navigate('sysops', 'hardware')} variant="default" description={latest.battery?.detected ? `${latest.battery.status}` : 'AC-powered.'} />
           <KpiCard icon={<Network size={24} />} label="Network" metricKey="network" unit="/s" onClick={() => navigate('netops', 'overview')} variant="accent" description="Real-time throughput." />
         </div>
      </motion.div>

      <div className="grid grid-cols-12 gap-5">
        <motion.div variants={itemVariants} className="col-span-12 lg:col-span-5">
          <AnalystBriefing title="Compute Logic Analysis" objective="Monitor CPU vs RAM." />
        </motion.div>
        <motion.div variants={itemVariants} className="col-span-12 lg:col-span-7">
          <Panel padding="md" className="flex flex-col h-full">
          <h3 className="text-base font-bold text-[var(--color-text)] uppercase tracking-wider mb-4 flex items-center gap-2"><Activity size={18} className="text-[var(--color-accent)]" /> Compute Timeline</h3>
          <div className="flex-1" style={{ minHeight: '280px' }}>
            <ResponsiveContainer width="100%" height="100%">
              <RechartsAreaChart data={[...cpuHistory.map(p => ({ ...p, isForecast: false })), ...(latest.cpu.forecast || []).map((v, i) => ({ time: `+${i + 1}m`, value: v, isForecast: true }))]}>
                <defs><linearGradient id="cpuGrad" x1="0" y1="0" x2="0" y2="1"><stop offset="0%" stopColor="var(--color-accent)" stopOpacity={0.3} /><stop offset="100%" stopColor="var(--color-accent)" stopOpacity={0} /></linearGradient></defs>
                <CartesianGrid strokeDasharray="4 4" stroke="var(--color-border)" vertical={false} strokeOpacity={0.5} />
                <XAxis dataKey="time" hide /><YAxis hide domain={[0, 100]} />
                <Tooltip contentStyle={{ backgroundColor: 'var(--color-panel-3)', border: 'none', borderRadius: '12px' }} />
                <Area type="monotone" dataKey="value" stroke="var(--color-accent)" strokeWidth={4} fill="url(#cpuGrad)" isAnimationActive={false} connectNulls />
              </RechartsAreaChart>
            </ResponsiveContainer>
          </div>
          </Panel>
        </motion.div>
      </div>

      <motion.div variants={itemVariants}>
        <RecentEventsPanel onExplain={handleExplain} explainingId={explainingId} />
      </motion.div>

      <Dialog.Root open={explanationOpen} onOpenChange={setExplanationOpen}><Dialog.Portal><Dialog.Overlay className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm" /><Dialog.Content className="fixed left-1/2 top-1/2 z-50 -translate-x-1/2 -translate-y-1/2 bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[20px] p-6 w-full max-w-xl shadow-2xl"><div className="flex items-center justify-between mb-6"><Dialog.Title className="text-xl font-bold text-[var(--color-text)] flex items-center gap-3"><Brain size={20} className="text-accent" /> AI Event Analysis</Dialog.Title><Dialog.Close className="text-[var(--color-text-faint)] hover:text-[var(--color-text)] transition-colors rounded-lg p-1 hover:bg-[var(--color-sidebar-hover)]"><XCircle size={18} /></Dialog.Close></div><div className="bg-[var(--color-bg)] border border-[var(--color-border)] rounded-xl p-6 shadow-inner min-h-[120px]">{!explanationText ? <div className="flex flex-col items-center justify-center py-8 gap-3"><Loader2 size={24} className="text-accent animate-spin" /><p className="text-xs font-bold text-text-faint uppercase tracking-widest">Heuristic Synthesis...</p></div> : <p className="text-sm text-[var(--color-text-dim)] leading-relaxed whitespace-pre-wrap">{explanationText}</p>}</div><div className="mt-6 flex justify-end"><Dialog.Close className="px-5 py-2 rounded-lg bg-[var(--color-panel-3)] text-xs font-bold uppercase tracking-wider text-text hover:bg-panel transition-all">Acknowledge</Dialog.Close></div></Dialog.Content></Dialog.Portal></Dialog.Root>

      <Dialog.Root open={diagOpen} onOpenChange={setDiagOpen}><Dialog.Portal><Dialog.Overlay className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm" /><Dialog.Content className="fixed left-1/2 top-1/2 z-50 -translate-x-1/2 -translate-y-1/2 bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[20px] p-6 w-full max-w-2xl max-h-[80vh] overflow-y-auto shadow-2xl"><div className="flex items-center justify-between mb-6"><Dialog.Title className="text-2xl font-bold text-[var(--color-text)] flex items-center gap-3"><Zap size={24} className="text-[var(--color-warning)]" /> Quick Diagnostic Results</Dialog.Title><Dialog.Close className="text-[var(--color-text-faint)] hover:text-[var(--color-text)] transition-colors rounded-lg p-1 hover:bg-[var(--color-sidebar-hover)]"><XCircle size={20} /></Dialog.Close></div>{diagLoading ? <div className="flex items-center justify-center py-16"><Loader2 size={32} className="text-[var(--color-accent)] animate-spin" /><span className="ml-4 text-lg font-bold text-[var(--color-text-faint)]">Running diagnostics...</span></div> : <div className="space-y-3">{diagResults.map((r, i) => ( <div key={i} className={`flex items-start gap-4 p-5 rounded-xl border ${diagColor(r.status)}`}><div className="mt-1 shrink-0">{diagIcon(r.status)}</div><div className="flex-1 min-w-0"><div className="flex items-center gap-3 mb-1"><span className="text-sm font-bold text-[var(--color-text)] uppercase tracking-wider">{r.category}</span><span className="text-xs font-semibold text-[var(--color-text-faint)] tabular-nums">{r.value.toFixed(1)}{r.unit}</span></div><p className="text-sm text-[var(--color-text-dim)] leading-relaxed">{r.message}</p></div></div> ))}</div>}</Dialog.Content></Dialog.Portal></Dialog.Root>

      <Dialog.Root open={briefingOpen} onOpenChange={setBriefingOpen}><Dialog.Portal><Dialog.Overlay className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm" /><Dialog.Content className="fixed left-1/2 top-1/2 z-50 -translate-x-1/2 -translate-y-1/2 bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[20px] p-6 w-full max-w-3xl max-h-[80vh] overflow-y-auto shadow-2xl"><div className="flex items-center justify-between mb-6"><Dialog.Title className="text-2xl font-bold text-[var(--color-text)] flex items-center gap-3"><ScrollText size={24} className="text-[var(--color-accent)]" /> Operations Briefing</Dialog.Title><Dialog.Close className="text-[var(--color-text-faint)] hover:text-[var(--color-text)] transition-colors rounded-lg p-1 hover:bg-[var(--color-sidebar-hover)]"><XCircle size={20} /></Dialog.Close></div>{briefingLoading ? <div className="flex items-center justify-center py-16"><Loader2 size={32} className="text-[var(--color-accent)] animate-spin" /><span className="ml-4 text-lg font-bold text-[var(--color-text-faint)]">Synthesizing briefing...</span></div> : briefingSections.length === 0 ? <p className="text-[var(--color-text-faint)] text-center py-8">No briefing data available.</p> : <div className="space-y-4">{briefingSections.map((s, i) => ( <div key={i} className="p-5 rounded-xl border border-[var(--color-border)] bg-[var(--color-panel-2)]"><div className="flex items-center gap-3 mb-2">{s.level === 'critical' ? <AlertTriangle size={18} className="text-[var(--color-danger)]" /> : s.level === 'warning' ? <AlertCircle size={18} className="text-[var(--color-warning)]" /> : <Info size={18} className="text-[var(--color-accent)]" />}<h3 className="text-sm font-bold text-[var(--color-text)] uppercase tracking-wider">{s.title}</h3></div><p className="text-sm text-[var(--color-text-dim)] whitespace-pre-wrap leading-relaxed">{s.content}</p></div> ))}</div>}</Dialog.Content></Dialog.Portal></Dialog.Root>
    </motion.div>
  )
}
