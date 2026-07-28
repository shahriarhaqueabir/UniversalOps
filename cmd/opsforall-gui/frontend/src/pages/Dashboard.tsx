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
  Shield,
  Terminal,
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
import { cn, formatSafeDate } from '@/lib/utils'
import { DataFreshnessIndicator } from '@/components/ui/DataFreshnessIndicator'
import { Panel } from '@/components/ui/Panel'
import type { DashboardData, SLOSummary } from '@/types'

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

/* ── Cross-Pillar Summary Types ── */

interface SecurityScore {
  score: number
  grade: string
  breakdown: Record<string, number>
  recommendations: string[]
}

interface SecuritySummary {
  score: number
  summary: string
  risks: string[]
  recommendations: string[]
  analyzedAt: string
}

interface DevOpsSummary {
  serviceCount: number
  runningCount: number
  dockerInstalled: boolean
  dockerRunning: boolean
  containerCount: number
  k8sInstalled: boolean
  k8sConnected: boolean
  k8sPods: number
  summary: string
}

interface AIInsight {
  category: string
  severity: string
  title: string
  message: string
  action: string
  actionPage: string
  timestamp: string
}

interface AIOpsSummary {
  ollamaAvailable: boolean
  ollamaModel: string
  anomalyCount: number
  criticalAnomalies: number
  recentInsights: AIInsight[]
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

function getHealthScoreColor(score: number) {
  if (score <= 10) return 'var(--color-danger)'
  if (score <= 20) return 'var(--color-warning)'
  return 'var(--color-success)'
}

function formatThroughput(bps: number): string {
  if (bps <= 0) return '0'
  const units = ['', 'K', 'M', 'G', 'T']
  const magnitude = Math.min(units.length - 1, Math.floor(Math.log10(bps) / 3))
  const scaled = bps / Math.pow(10, magnitude * 3)
  return scaled >= 10 ? `${Math.round(scaled)}${units[magnitude]}` : `${scaled.toFixed(1)}${units[magnitude]}`
}

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

  // Use the backend drift-weighted health score, fall back to naive computation
  const systemHealth = stats.health_score != null ? stats.health_score : clamp(100 - Math.max(stats.cpu?.value ?? 0, stats.memory?.value ?? 0, stats.disk?.value ?? 0))
  const healthTrend = stats.health_trend ?? []
  const r = 44
  const circumference = 2 * Math.PI * r
  const dash = (systemHealth / 100) * circumference
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
            stroke={getHealthScoreColor(systemHealth)}
            strokeWidth="10"
            strokeLinecap="round"
            strokeDasharray={`${dash} ${gap}`}
            transform="rotate(-90 60 60)"
            style={{ transition: 'stroke-dasharray 1.2s cubic-bezier(0.34, 1.56, 0.64, 1)' }}
          />
          <text x="60" y="54" textAnchor="middle" fill="var(--color-text)" fontSize="28" fontWeight="900" dominantBaseline="middle" className="tabular-nums">
            {Math.round(systemHealth)}%
          </text>
          <text x="60" y="78" textAnchor="middle" fill="var(--color-text-faint)" fontSize="12" fontWeight="bold" style={{ textTransform: 'uppercase' }} dominantBaseline="middle">
            HEALTH
          </text>
        </svg>
      </div>

      <div className="flex-1 min-w-0 w-full">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-4">
            <span className={cn(
              "px-3 py-1 rounded-full border text-xs font-semibold uppercase tracking-wider transition-colors",
              systemHealth > 80 ? "bg-success/10 border-success/30 text-success" :
              systemHealth > 50 ? "bg-warning/10 border-warning/30 text-warning" :
              "bg-danger/10 border-danger/30 text-danger animate-pulse"
            )}>
              {systemHealth > 80 ? "System Nominal" : systemHealth > 50 ? "Performance Degraded" : "Critical State"}
            </span>
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
          Drift-weighted health score computed by the backend engine from CPU, memory, and disk deviation analysis.
          Current status indicates <span className={cn("font-semibold", systemHealth > 80 ? "text-success" : systemHealth > 50 ? "text-warning" : "text-danger")}>
            {systemHealth > 80 ? "Stable Operation" : systemHealth > 50 ? "Pressure Detected" : "High Instability Risks"}
          </span> across all monitored subsystems.
        </p>

        <div className="flex gap-1 items-end h-12 w-full bg-panel-3 p-2 rounded-xl border border-border shadow-inner">
          {healthTrend.length > 0 ? healthTrend.map((point, i) => (
            <div
              key={i}
              className="flex-1 rounded-[2px] transition-all relative group/tooltip"
              style={{
                height: `${Math.max(5, point.score)}%`,
                backgroundColor: point.score >= 80 ? 'var(--color-success)' : point.score >= 50 ? 'var(--color-warning)' : 'var(--color-danger)',
                opacity: 0.7,
              }}
              title={`${point.day}: ${point.score}%`}
            />
          )) : cpuHistoryBars.length > 0 ? cpuHistoryBars.map((val, i) => (
            <div
              key={i}
              className="flex-1 rounded-[2px] transition-all"
              style={{
                height: `${Math.max(5, val)}%`,
                backgroundColor: val >= 90 ? 'var(--color-danger)' : val >= 80 ? 'var(--color-warning)' : 'var(--color-success)',
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
  const displayVal = (metricKey === 'network')
    ? formatThroughput((data?.rx_rate ?? 0) + (data?.tx_rate ?? 0))
    : (metricKey === 'gpu' || metricKey === 'battery')
      ? (data?.detected ? (metricKey === 'gpu' ? data.vendor : val) : 'N/A')
      : val

  const status = (data?.value != null && data.value > 80) ? 'warning' : 'healthy'

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
        {unit && !(metricKey === 'network') && <span className="text-lg font-semibold text-[var(--color-text-faint)]">{unit}</span>}
      </div>
      {metricKey === 'network' ? (
        <div className="flex flex-col gap-1 border-t border-[var(--color-border)]/50 pt-3">
          <div className="flex items-center justify-between text-xs">
            <span className="text-[var(--color-text-faint)]">↓ RX</span>
            <span className="font-bold text-[var(--color-text)] tabular-nums">{formatThroughput(data?.rx_rate ?? 0)}<span className="font-semibold text-[var(--color-text-faint)] ml-1">bps</span></span>
          </div>
          <div className="flex items-center justify-between text-xs">
            <span className="text-[var(--color-text-faint)]">↑ TX</span>
            <span className="font-bold text-[var(--color-text)] tabular-nums">{formatThroughput(data?.tx_rate ?? 0)}<span className="font-semibold text-[var(--color-text-faint)] ml-1">bps</span></span>
          </div>
          <p className="text-xs text-[var(--color-text-faint)] leading-relaxed mt-1">{description}</p>
        </div>
      ) : (
        <p className="text-xs text-[var(--color-text-faint)] leading-relaxed border-t border-[var(--color-border)]/50 pt-3">
          {description}
        </p>
      )}
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
                <span className="text-[10px] text-[var(--color-text-faint)] tabular-nums ml-auto">{formatSafeDate(evt.timestamp, (d) => format(d, 'HH:mm:ss'))}</span>
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

/* ── Cross-Pillar Dashboard Panels ── */

const SecOpsPanel = memo(function SecOpsPanel() {
  const { call } = useBackend()
  const { navigate } = useNavigationStore()

  const { data: secScore } = useQuery<SecurityScore>({
    queryKey: ['dashboard-sec-score'],
    queryFn: () => call('Dashboard.GetSecurityScore') as Promise<SecurityScore>,
    staleTime: 30000,
    refetchInterval: 60000,
    refetchOnWindowFocus: false,
  })

  const { data: secSummary } = useQuery<SecuritySummary>({
    queryKey: ['dashboard-sec-summary'],
    queryFn: () => call('Dashboard.GetSecuritySummary') as Promise<SecuritySummary>,
    staleTime: 30000,
    refetchInterval: 60000,
    refetchOnWindowFocus: false,
  })

  const score = secScore?.score ?? 0
  const grade = secScore?.grade ?? 'N/A'
  const riskCount = secSummary?.risks?.length ?? 0
  const recCount = secSummary?.recommendations?.length ?? 0

  const scoreColor = score >= 80 ? 'var(--color-success)' : score >= 50 ? 'var(--color-warning)' : 'var(--color-danger)'
  const r = 32
  const circumference = 2 * Math.PI * r
  const dash = (score / 100) * circumference
  const gap = circumference - dash

  return (
    <Panel padding="md" className="group hover:border-[var(--color-accent)]/30 flex flex-col rounded-[2rem]">
      <div className="flex items-center gap-3 mb-4">
        <div className="w-10 h-10 rounded-2xl bg-danger/10 border border-danger/20 flex items-center justify-center text-danger shadow-sm group-hover:scale-110 transition-transform">
          <Shield size={20} />
        </div>
        <div className="flex-1">
          <h3 className="text-xs font-black text-danger uppercase tracking-[0.2em]">Security Posture</h3>
          <p className="text-[10px] text-text-faint font-bold uppercase tracking-wider">Threat & Compliance</p>
        </div>
        <button
          onClick={() => navigate('secops')}
          className="text-[10px] font-black uppercase tracking-tighter text-accent hover:text-accent/80 bg-accent/5 px-3 py-1.5 rounded-lg border border-accent/10 transition-all hover:bg-accent/10"
        >
          Open SecOps
        </button>
      </div>

      <div className="flex items-center gap-6 flex-1">
        <div className="shrink-0">
          <svg width={100} height={100} viewBox="0 0 80 80" className="tabular-nums drop-shadow-[0_0_6px_rgba(var(--color-danger-rgb),0.15)]">
            <circle cx="40" cy="40" r={r} fill="none" stroke="var(--color-border)" strokeWidth="8" />
            <circle
              cx="40" cy="40" r={r}
              fill="none"
              stroke={scoreColor}
              strokeWidth="8"
              strokeLinecap="round"
              strokeDasharray={`${dash} ${gap}`}
              transform="rotate(-90 40 40)"
              style={{ transition: 'stroke-dasharray 1.2s cubic-bezier(0.34, 1.56, 0.64, 1)' }}
            />
            <text x="40" y="36" textAnchor="middle" fill="var(--color-text)" fontSize="18" fontWeight="900" dominantBaseline="middle" className="tabular-nums">
              {score}
            </text>
            <text x="40" y="54" textAnchor="middle" fill="var(--color-text-faint)" fontSize="9" fontWeight="bold" dominantBaseline="middle">
              {grade}
            </text>
          </svg>
        </div>

        <div className="flex-1 space-y-2 min-w-0">
          <div className="flex items-center justify-between">
            <span className="text-[10px] font-bold text-text-faint uppercase tracking-wider">Risks</span>
            <span className={cn("text-xs font-black tabular-nums", riskCount > 0 ? "text-danger" : "text-success")}>{riskCount}</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-[10px] font-bold text-text-faint uppercase tracking-wider">Recommendations</span>
            <span className={cn("text-xs font-black tabular-nums", recCount > 0 ? "text-warning" : "text-success")}>{recCount}</span>
          </div>
          {secSummary?.summary && (
            <p className="text-[11px] text-text-dim leading-relaxed mt-2 line-clamp-2 border-t border-border/50 pt-2">
              {secSummary.summary}
            </p>
          )}
        </div>
      </div>
    </Panel>
  )
})

const DevOpsPanel = memo(function DevOpsPanel() {
  const { call } = useBackend()
  const { navigate } = useNavigationStore()

  const { data: devOpsSummary } = useQuery<DevOpsSummary>({
    queryKey: ['dashboard-devops-summary'],
    queryFn: () => call('Dashboard.GetDevOpsSummary') as Promise<DevOpsSummary>,
    staleTime: 30000,
    refetchInterval: 60000,
    refetchOnWindowFocus: false,
  })

  const svcRunning = devOpsSummary?.runningCount ?? 0
  const svcTotal = devOpsSummary?.serviceCount ?? 0
  const dockerOk = devOpsSummary?.dockerRunning ?? false
  const k8sOk = devOpsSummary?.k8sConnected ?? false
  const containerCount = devOpsSummary?.containerCount ?? 0
  const k8sPods = devOpsSummary?.k8sPods ?? 0

  return (
    <Panel padding="md" className="group hover:border-[var(--color-accent)]/30 flex flex-col rounded-[2rem]">
      <div className="flex items-center gap-3 mb-4">
        <div className="w-10 h-10 rounded-2xl bg-accent/10 border border-accent/20 flex items-center justify-center text-accent shadow-sm group-hover:scale-110 transition-transform">
          <Terminal size={20} />
        </div>
        <div className="flex-1">
          <h3 className="text-xs font-black text-accent uppercase tracking-[0.2em]">DevOps Health</h3>
          <p className="text-[10px] text-text-faint font-bold uppercase tracking-wider">Services & Containers</p>
        </div>
        <button
          onClick={() => navigate('devops')}
          className="text-[10px] font-black uppercase tracking-tighter text-accent hover:text-accent/80 bg-accent/5 px-3 py-1.5 rounded-lg border border-accent/10 transition-all hover:bg-accent/10"
        >
          Open DevOps
        </button>
      </div>

      <div className="grid grid-cols-2 gap-3 flex-1">
        <div className="bg-panel-2 rounded-xl p-3 border border-border/50">
          <div className="flex items-center gap-2 mb-2">
            <div className={cn("w-2 h-2 rounded-full", svcRunning === svcTotal && svcTotal > 0 ? "bg-success shadow-[0_0_6px_var(--color-success)]" : "bg-warning shadow-[0_0_6px_var(--color-warning)]")} />
            <span className="text-[10px] font-bold text-text-faint uppercase tracking-wider">Services</span>
          </div>
          <span className="text-xl font-black text-text tabular-nums">{svcRunning}<span className="text-sm font-semibold text-text-faint">/{svcTotal}</span></span>
        </div>
        <div className="bg-panel-2 rounded-xl p-3 border border-border/50">
          <div className="flex items-center gap-2 mb-2">
            <div className={cn("w-2 h-2 rounded-full", dockerOk ? "bg-success shadow-[0_0_6px_var(--color-success)]" : "bg-danger shadow-[0_0_6px_var(--color-danger)]")} />
            <span className="text-[10px] font-bold text-text-faint uppercase tracking-wider">Docker</span>
          </div>
          <span className="text-xl font-black text-text tabular-nums">{containerCount}<span className="text-sm font-semibold text-text-faint"> containers</span></span>
        </div>
        <div className="bg-panel-2 rounded-xl p-3 border border-border/50">
          <div className="flex items-center gap-2 mb-2">
            <div className={cn("w-2 h-2 rounded-full", k8sOk ? "bg-success shadow-[0_0_6px_var(--color-success)]" : "bg-text-faint/30")} />
            <span className="text-[10px] font-bold text-text-faint uppercase tracking-wider">Kubernetes</span>
          </div>
          <span className="text-xl font-black text-text tabular-nums">{k8sPods}<span className="text-sm font-semibold text-text-faint"> pods</span></span>
        </div>
        <div className="bg-panel-2 rounded-xl p-3 border border-border/50">
          <div className="flex items-center gap-2 mb-2">
            <span className="text-[10px] font-bold text-text-faint uppercase tracking-wider">Status</span>
          </div>
          <span className={cn("text-sm font-black", devOpsSummary?.summary?.includes('stopped') || devOpsSummary?.summary?.includes('not running') ? "text-warning" : "text-success")}>
            {devOpsSummary?.summary ?? 'Unknown'}
          </span>
        </div>
      </div>
    </Panel>
  )
})

const AIOpsPanel = memo(function AIOpsPanel() {
  const { call } = useBackend()
  const { navigate } = useNavigationStore()

  const { data: aiOpsSummary } = useQuery<AIOpsSummary>({
    queryKey: ['dashboard-aiops-summary'],
    queryFn: () => call('Dashboard.GetAIOpsSummary') as Promise<AIOpsSummary>,
    staleTime: 30000,
    refetchInterval: 60000,
    refetchOnWindowFocus: false,
  })

  const ollamaOk = aiOpsSummary?.ollamaAvailable ?? false
  const ollamaModel = aiOpsSummary?.ollamaModel ?? 'N/A'
  const anomalyCount = aiOpsSummary?.anomalyCount ?? 0
  const criticalAnomalies = aiOpsSummary?.criticalAnomalies ?? 0
  const insights = aiOpsSummary?.recentInsights ?? []

  return (
    <Panel padding="md" className="group hover:border-[var(--color-accent)]/30 flex flex-col rounded-[2rem]">
      <div className="flex items-center gap-3 mb-4">
        <div className="w-10 h-10 rounded-2xl bg-accent/10 border border-accent/20 flex items-center justify-center text-accent shadow-sm group-hover:scale-110 transition-transform">
          <Brain size={20} />
        </div>
        <div className="flex-1">
          <h3 className="text-xs font-black text-accent uppercase tracking-[0.2em]">AI Operations</h3>
          <p className="text-[10px] text-text-faint font-bold uppercase tracking-wider">Anomalies & Insights</p>
        </div>
        <button
          onClick={() => navigate('aiops')}
          className="text-[10px] font-black uppercase tracking-tighter text-accent hover:text-accent/80 bg-accent/5 px-3 py-1.5 rounded-lg border border-accent/10 transition-all hover:bg-accent/10"
        >
          Open AIOps
        </button>
      </div>

      <div className="grid grid-cols-2 gap-3 mb-3">
        <div className="bg-panel-2 rounded-xl p-3 border border-border/50">
          <div className="flex items-center gap-2 mb-2">
            <div className={cn("w-2 h-2 rounded-full", ollamaOk ? "bg-success shadow-[0_0_6px_var(--color-success)]" : "bg-danger shadow-[0_0_6px_var(--color-danger)]")} />
            <span className="text-[10px] font-bold text-text-faint uppercase tracking-wider">Ollama</span>
          </div>
          <span className="text-sm font-black text-text tabular-nums truncate block">{ollamaModel}</span>
        </div>
        <div className="bg-panel-2 rounded-xl p-3 border border-border/50">
          <div className="flex items-center gap-2 mb-2">
            <span className="text-[10px] font-bold text-text-faint uppercase tracking-wider">Anomalies</span>
          </div>
          <div className="flex items-baseline gap-2">
            <span className={cn("text-xl font-black tabular-nums", criticalAnomalies > 0 ? "text-danger" : anomalyCount > 0 ? "text-warning" : "text-success")}>
              {anomalyCount}
            </span>
            {criticalAnomalies > 0 && (
              <span className="text-[10px] font-bold text-danger uppercase tracking-wider">{criticalAnomalies} critical</span>
            )}
          </div>
        </div>
      </div>

      {insights.length > 0 && (
        <div className="space-y-1.5 border-t border-border/50 pt-3">
          <p className="text-[10px] font-bold text-text-faint uppercase tracking-wider mb-1.5">Recent Insights</p>
          {insights.slice(0, 2).map((insight, i) => (
            <div key={i} className="flex items-start gap-2 bg-panel-2 rounded-lg p-2 border border-border/50">
              <div className={cn(
                "w-1.5 h-1.5 rounded-full mt-1 shrink-0",
                insight.severity === 'critical' ? 'bg-danger' : insight.severity === 'warning' ? 'bg-warning' : 'bg-accent'
              )} />
              <div className="min-w-0">
                <p className="text-[11px] font-bold text-text leading-tight truncate">{insight.title}</p>
                <p className="text-[10px] text-text-dim leading-relaxed line-clamp-1">{insight.message}</p>
              </div>
            </div>
          ))}
        </div>
      )}
    </Panel>
  )
})

/* ── SLO/SLI Panel ── */

const SLOPanel = memo(function SLOPanel() {
  const { call } = useBackend()

  const { data: sloSummary, isLoading } = useQuery<SLOSummary>({
    queryKey: ['dashboard-slo-summary'],
    queryFn: () => call('Dashboard.GetSLOSummary') as Promise<SLOSummary>,
    staleTime: 30000,
    refetchInterval: 60000,
    refetchOnWindowFocus: false,
  })

  const total = sloSummary?.totalSLOs ?? 0
  const met = sloSummary?.metCount ?? 0
  const miss = sloSummary?.missCount ?? 0
  const overall = sloSummary?.overallPct ?? 100
  const results = sloSummary?.results ?? []

  return (
    <Panel padding="md" className="group hover:border-[var(--color-accent)]/30 flex flex-col rounded-[2rem]">
      <div className="flex items-center gap-3 mb-4">
        <div className="w-10 h-10 rounded-2xl bg-accent/10 border border-accent/20 flex items-center justify-center text-accent shadow-sm group-hover:scale-110 transition-transform">
          <Target size={20} />
        </div>
        <div className="flex-1">
          <h3 className="text-xs font-black text-accent uppercase tracking-[0.2em]">SLO / SLI</h3>
          <p className="text-[10px] text-text-faint font-bold uppercase tracking-wider">Service Level Objectives</p>
        </div>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center py-6">
          <Loader2 size={18} className="animate-spin text-text-faint" />
        </div>
      ) : total === 0 ? (
        <div className="flex flex-col items-center justify-center py-6 text-text-faint">
          <Target size={24} className="mb-2 opacity-40" />
          <p className="text-[11px] font-bold uppercase tracking-wider">No SLOs configured</p>
          <p className="text-[10px] text-text-dim mt-1">Metrics will populate as data is collected.</p>
        </div>
      ) : (
        <>
          {/* Overall compliance ring */}
          <div className="flex items-center gap-4 mb-4 bg-panel-2 rounded-xl p-3 border border-border/50">
            <div className={cn(
              "w-14 h-14 rounded-full flex items-center justify-center text-sm font-black tabular-nums border-2 shrink-0",
              overall >= 95 ? "border-success text-success" :
              overall >= 80 ? "border-warning text-warning" :
              "border-danger text-danger"
            )}>
              {overall.toFixed(0)}%
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-3 text-[11px] font-bold">
                <span className="flex items-center gap-1 text-success"><CheckCircle2 size={12} /> {met} met</span>
                {miss > 0 && <span className="flex items-center gap-1 text-danger"><XCircle size={12} /> {miss} miss</span>}
                <span className="text-text-faint">{total} total</span>
              </div>
              <div className="mt-1.5 h-1.5 bg-panel-3 rounded-full overflow-hidden">
                <div
                  className={cn("h-full rounded-full transition-all duration-500", overall >= 95 ? "bg-success" : overall >= 80 ? "bg-warning" : "bg-danger")}
                  style={{ width: `${Math.min(overall, 100)}%` }}
                />
              </div>
            </div>
          </div>

          {/* Individual SLO results */}
          <div className="space-y-2">
            {results.map((r) => (
              <div key={r.sloId} className="flex items-center gap-3 bg-panel-2 rounded-lg p-2.5 border border-border/50">
                <div className={cn(
                  "w-2 h-2 rounded-full shrink-0",
                  r.met ? "bg-success shadow-[0_0_6px_var(--color-success)]" : "bg-danger shadow-[0_0_6px_var(--color-danger)]"
                )} />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between gap-2">
                    <span className="text-[11px] font-bold text-text truncate">{r.sloName}</span>
                    <span className={cn(
                      "text-[11px] font-black tabular-nums shrink-0",
                      r.met ? "text-success" : "text-danger"
                    )}>
                      {r.compliantPct.toFixed(1)}%
                    </span>
                  </div>
                  <div className="flex items-center gap-2 mt-0.5">
                    <span className="text-[9px] text-text-faint font-bold uppercase tracking-wider">
                      target {r.targetPct}% · {r.samples} samples
                    </span>
                  </div>
                  <div className="mt-1 h-1 bg-panel-3 rounded-full overflow-hidden">
                    <div
                      className={cn("h-full rounded-full", r.met ? "bg-success" : "bg-danger")}
                      style={{ width: `${Math.min(r.compliantPct, 100)}%` }}
                    />
                  </div>
                </div>
              </div>
            ))}
          </div>
        </>
      )}
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

export function Dashboard() {
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

      {/* ── Top Issues ── */}
      <motion.div variants={itemVariants} className="space-y-5">
        <div className="flex items-center gap-4"><div className="h-px flex-1 bg-[var(--color-border)]" /><h2 className="text-xs font-semibold text-[var(--color-text-faint)] uppercase tracking-widest">Top Issues</h2><div className="h-px flex-1 bg-[var(--color-border)]" /></div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
          <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 hover:border-warning/30 transition-all group">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-10 h-10 rounded-lg bg-warning/10 border border-warning/20 flex items-center justify-center text-warning">
                <Cpu size={20} />
              </div>
              <div>
                <h3 className="text-sm font-bold text-text">High CPU Usage</h3>
                <p className="text-[10px] text-text-faint font-bold uppercase tracking-wider">Sustained Load</p>
              </div>
              <span className="ml-auto px-2.5 py-1 rounded-full bg-warning/10 border border-warning/20 text-[10px] font-black text-warning uppercase tracking-wider">Warning</span>
            </div>
            <p className="text-xs text-text-dim leading-relaxed mb-4">
              CPU sustained at {latest.cpu?.value != null ? Math.round(latest.cpu.value) : 'N/A'}% — above the 70% threshold. Consider terminating high-consumption processes or scaling compute resources.
            </p>
            <div className="flex items-center justify-between border-t border-border/50 pt-3">
              <span className="text-[10px] text-text-faint font-bold uppercase tracking-wider">Processes: {latest.processes ?? 'N/A'}</span>
              <button
                onClick={() => navigate('sysops', 'processes')}
                className="text-[10px] font-black uppercase tracking-tighter text-accent hover:text-accent/80 bg-accent/5 px-3 py-1.5 rounded-lg border border-accent/10 transition-all hover:bg-accent/10 flex items-center gap-1.5"
              >
                <FileSearch size={12} /> Investigate
              </button>
            </div>
          </div>
          <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 hover:border-warning/30 transition-all group">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-10 h-10 rounded-lg bg-accent/10 border border-accent/20 flex items-center justify-center text-accent">
                <Terminal size={20} />
              </div>
              <div>
                <h3 className="text-sm font-bold text-text">Browser Memory Pressure</h3>
                <p className="text-[10px] text-text-faint font-bold uppercase tracking-wider">Chrome / Edge</p>
              </div>
              <span className="ml-auto px-2.5 py-1 rounded-full bg-warning/10 border border-warning/20 text-[10px] font-black text-warning uppercase tracking-wider">Elevated</span>
            </div>
            <p className="text-xs text-text-dim leading-relaxed mb-4">
              Memory at {latest.memory?.value != null ? Math.round(latest.memory.value) : 'N/A'}% — browser processes consuming significant RAM. Close unused tabs or consider a memory management extension.
            </p>
            <div className="flex items-center justify-between border-t border-border/50 pt-3">
              <span className="text-[10px] text-text-faint font-bold uppercase tracking-wider">Memory: {latest.memory?.value != null ? Math.round(latest.memory.value) : 'N/A'}%</span>
              <button
                onClick={() => navigate('sysops', 'memory')}
                className="text-[10px] font-black uppercase tracking-tighter text-accent hover:text-accent/80 bg-accent/5 px-3 py-1.5 rounded-lg border border-accent/10 transition-all hover:bg-accent/10 flex items-center gap-1.5"
              >
                <FileSearch size={12} /> Investigate
              </button>
            </div>
          </div>
        </div>
      </motion.div>

      {/* ── Quick Actions ── */}
      <motion.div variants={itemVariants} className="space-y-5">
        <div className="flex items-center gap-4"><div className="h-px flex-1 bg-[var(--color-border)]" /><h2 className="text-xs font-semibold text-[var(--color-text-faint)] uppercase tracking-widest">Quick Actions</h2><div className="h-px flex-1 bg-[var(--color-border)]" /></div>
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4">
          <button
            onClick={runQuickDiag}
            className="flex flex-col items-center gap-3 p-5 rounded-[var(--radius-lg)] bg-panel border border-border hover:border-accent/30 hover:bg-panel-2 transition-all group active:scale-[0.97]"
          >
            <div className="w-12 h-12 rounded-xl bg-warning/10 border border-warning/20 flex items-center justify-center text-warning group-hover:scale-110 transition-transform">
              <Zap size={22} />
            </div>
            <span className="text-xs font-bold text-text-dim group-hover:text-text transition-colors">Scan System</span>
            <span className="text-[9px] text-text-faint font-semibold uppercase tracking-wider text-center">Quick diagnostic scan</span>
          </button>
          <button
            onClick={() => navigate('netops', 'overview')}
            className="flex flex-col items-center gap-3 p-5 rounded-[var(--radius-lg)] bg-panel border border-border hover:border-accent/30 hover:bg-panel-2 transition-all group active:scale-[0.97]"
          >
            <div className="w-12 h-12 rounded-xl bg-accent/10 border border-accent/20 flex items-center justify-center text-accent group-hover:scale-110 transition-transform">
              <Network size={22} />
            </div>
            <span className="text-xs font-bold text-text-dim group-hover:text-text transition-colors">Network Diagnostics</span>
            <span className="text-[9px] text-text-faint font-semibold uppercase tracking-wider text-center">Test connectivity & latency</span>
          </button>
          <button
            onClick={generateBriefing}
            className="flex flex-col items-center gap-3 p-5 rounded-[var(--radius-lg)] bg-panel border border-border hover:border-accent/30 hover:bg-panel-2 transition-all group active:scale-[0.97]"
          >
            <div className="w-12 h-12 rounded-xl bg-accent/10 border border-accent/20 flex items-center justify-center text-accent group-hover:scale-110 transition-transform">
              <FileSearch size={22} />
            </div>
            <span className="text-xs font-bold text-text-dim group-hover:text-text transition-colors">Generate Report</span>
            <span className="text-[9px] text-text-faint font-semibold uppercase tracking-wider text-center">Full operations briefing</span>
          </button>
          <button
            onClick={() => navigate('aiops')}
            className="flex flex-col items-center gap-3 p-5 rounded-[var(--radius-lg)] bg-panel border border-border hover:border-accent/30 hover:bg-panel-2 transition-all group active:scale-[0.97]"
          >
            <div className="w-12 h-12 rounded-xl bg-accent/10 border border-accent/20 flex items-center justify-center text-accent group-hover:scale-110 transition-transform">
              <Brain size={22} />
            </div>
            <span className="text-xs font-bold text-text-dim group-hover:text-text transition-colors">AI Analysis</span>
            <span className="text-[9px] text-text-faint font-semibold uppercase tracking-wider text-center">Heuristic insights & anomalies</span>
          </button>
          <button
            onClick={() => navigate('sysops', 'overview')}
            className="flex flex-col items-center gap-3 p-5 rounded-[var(--radius-lg)] bg-panel border border-border hover:border-accent/30 hover:bg-panel-2 transition-all group active:scale-[0.97]"
          >
            <div className="w-12 h-12 rounded-xl bg-accent/10 border border-accent/20 flex items-center justify-center text-accent group-hover:scale-110 transition-transform">
              <Activity size={22} />
            </div>
            <span className="text-xs font-bold text-text-dim group-hover:text-text transition-colors">Take Snapshot</span>
            <span className="text-[9px] text-text-faint font-semibold uppercase tracking-wider text-center">System state capture</span>
          </button>
        </div>
      </motion.div>

      <motion.div variants={itemVariants} className="space-y-5">
        <div className="flex items-center gap-4"><div className="h-px flex-1 bg-[var(--color-border)]" /><h2 className="text-xs font-semibold text-[var(--color-text-faint)] uppercase tracking-widest">Cross-Pillar Operations</h2><div className="h-px flex-1 bg-[var(--color-border)]" /></div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-2 xl:grid-cols-4 gap-5">
          <SecOpsPanel />
          <DevOpsPanel />
          <AIOpsPanel />
          <SLOPanel />
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
