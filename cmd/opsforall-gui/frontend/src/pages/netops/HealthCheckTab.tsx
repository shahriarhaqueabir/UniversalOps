import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import {
  Stethoscope,
  Play,
  CheckCircle2,
  AlertTriangle,
  XCircle,
  RefreshCw,
  Clock,
  BarChart3,
} from 'lucide-react'
import { SectionBriefing } from '@/components/ui/SectionBriefing'
import type { HealthReportData } from '@/types'

// ── Score Ring ──

function ScoreRing({ score }: { score: number }) {
  const color =
    score >= 80
      ? 'var(--color-success)'
      : score >= 60
        ? 'var(--color-warning)'
        : 'var(--color-danger)'

  const circumference = 2 * Math.PI * 54
  const offset = circumference - (score / 100) * circumference

  return (
    <div className="relative w-36 h-36 shrink-0">
      <svg viewBox="0 0 120 120" className="w-full h-full -rotate-90">
        <circle cx="60" cy="60" r="54" fill="none" stroke="var(--color-panel-3)" strokeWidth="10" />
        <circle
          cx="60"
          cy="60"
          r="54"
          fill="none"
          stroke={color}
          strokeWidth="10"
          strokeLinecap="round"
          strokeDasharray={circumference}
          strokeDashoffset={offset}
          className="transition-all duration-1000 ease-out"
          style={{ filter: `drop-shadow(0 0 8px ${color})` }}
        />
      </svg>
      <div className="absolute inset-0 flex flex-col items-center justify-center">
        <span className="text-4xl font-bold tabular-nums" style={{ color }}>
          {score}
        </span>
        <span className="text-[10px] font-bold text-text-faint uppercase tracking-widest mt-0.5">
          / 100
        </span>
      </div>
    </div>
  )
}

// ── Status Icon ──

function StatusIcon({ status }: { status: 'pass' | 'warn' | 'fail' }) {
  if (status === 'pass') return <CheckCircle2 size={18} className="text-success shrink-0" />
  if (status === 'warn') return <AlertTriangle size={18} className="text-warning shrink-0" />
  return <XCircle size={18} className="text-danger shrink-0" />
}

// ── Main HealthCheckTab ──

export function HealthCheckTab() {
  const { call } = useBackend()
  const [report, setReport] = useState<HealthReportData | null>(null)
  const [hasRun, setHasRun] = useState(false)

  const { isLoading, refetch } = useQuery({
    queryKey: ['netops-health-check'],
    queryFn: async () => {
      const res = (await call('NetOps.RunNetworkHealthCheck')) as HealthReportData
      setReport(res)
      setHasRun(true)
      return res
    },
    enabled: false,
    retry: false,
  })

  const runCheck = () => {
    refetch()
  }

  const scoreColor =
    report && report.score >= 80
      ? 'text-success'
      : report && report.score >= 60
        ? 'text-warning'
        : 'text-danger'

  const scoreLabel =
    report && report.score >= 80
      ? 'Healthy Network'
      : report && report.score >= 60
        ? 'Degraded Performance'
        : 'Significant Issues'

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <SectionBriefing
        title="Health Check"
        objective="Run a comprehensive network health diagnostic. Scores your network from 0-100 across multiple dimensions including connectivity, DNS, gateway, latency, packet loss, interfaces, and VPN."
        checklist={[
          'Score 80+ = healthy network',
          'Score 60-79 = degraded performance',
          'Score below 60 = significant issues',
          'Individual checks reveal specific problems',
        ]}
      />

      {/* ── Run Button ── */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
        <div className="flex items-center gap-4">
          <button
            onClick={runCheck}
            disabled={isLoading}
            className={cn(
              'flex items-center gap-3 px-6 py-3 text-sm font-semibold rounded-xl transition-all shadow-xl',
              isLoading
                ? 'bg-panel-3 text-text-faint border border-border cursor-not-allowed'
                : 'bg-accent text-white hover:bg-accent/90',
            )}
          >
            {isLoading ? (
              <RefreshCw size={16} className="animate-spin" />
            ) : (
              <Play size={16} fill="currentColor" />
            )}
            {isLoading ? 'Running...' : 'Run Health Check'}
          </button>
          {!hasRun && !isLoading && (
            <p className="text-xs font-medium text-text-faint flex items-center gap-2">
              <Stethoscope size={14} className="text-accent" />
              Click Run Health Check to begin
            </p>
          )}
        </div>
      </div>

      {/* ── Results ── */}
      {report && (
        <>
          {/* ── Score + Summary ── */}
          <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
            <div className="flex items-center gap-8">
              <ScoreRing score={report.score} />
              <div className="flex-1 space-y-4">
                <div>
                  <h4 className={cn('text-xl font-bold uppercase tracking-widest', scoreColor)}>
                    {scoreLabel}
                  </h4>
                  <p className="text-sm text-text-dim mt-1 leading-relaxed">{report.summary}</p>
                </div>
                <div className="flex items-center gap-2 text-text-faint">
                  <Clock size={14} />
                  <span className="text-xs font-bold">Duration:</span>
                  <span className="text-xs font-medium tabular-nums">{report.duration}</span>
                </div>
              </div>
            </div>
          </div>

          {/* ── Per-Check Results ── */}
          <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-9 h-9 rounded-lg flex items-center justify-center bg-panel-3 border border-border">
                <BarChart3 size={18} className="text-accent" />
              </div>
              <h3 className="text-sm font-bold text-text uppercase tracking-widest">Check Results</h3>
              <span className="ml-auto px-2.5 py-0.5 text-[10px] font-bold rounded-full bg-accent/15 text-accent border border-accent/30">
                {report.checks.length}
              </span>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {report.checks.map((check, idx) => {
                const borderColor =
                  check.status === 'pass'
                    ? 'border-success/30 hover:border-success/50'
                    : check.status === 'warn'
                      ? 'border-warning/30 hover:border-warning/50'
                      : 'border-danger/30 hover:border-danger/50'

                const badgeStyle =
                  check.status === 'pass'
                    ? 'bg-success/15 text-success border-success/30'
                    : check.status === 'warn'
                      ? 'bg-warning/15 text-warning border-warning/30'
                      : 'bg-danger/15 text-danger border-danger/30'

                return (
                  <div
                    key={idx}
                    className={cn(
                      'bg-panel-2 border rounded-xl p-4 transition-all',
                      borderColor,
                    )}
                  >
                    <div className="flex items-center justify-between mb-3">
                      <div className="flex items-center gap-2 min-w-0">
                        <StatusIcon status={check.status} />
                        <span className="text-xs font-bold text-text uppercase tracking-wider truncate">
                          {check.name}
                        </span>
                      </div>
                      <span
                        className={cn(
                          'px-2 py-0.5 text-[10px] font-bold uppercase tracking-widest rounded-full border',
                          badgeStyle,
                        )}
                      >
                        {check.status}
                      </span>
                    </div>
                    <p className="text-[11px] font-medium text-text-dim mb-3 leading-relaxed">
                      {check.detail}
                    </p>
                    <div className="flex items-center justify-between">
                      <span className="text-[10px] font-bold text-text-faint uppercase tracking-wider">
                        Score
                      </span>
                      <span className="text-sm font-bold tabular-nums text-text">
                        {check.score}
                      </span>
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        </>
      )}

      {/* ── Empty State ── */}
      {!hasRun && !isLoading && !report && (
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-16 shadow-xl flex flex-col items-center justify-center text-center">
          <div className="w-16 h-16 rounded-2xl bg-panel-3 flex items-center justify-center border border-border mb-4">
            <Stethoscope size={32} className="text-text-faint" />
          </div>
          <p className="text-sm font-bold text-text uppercase tracking-widest mb-1">
            No Health Data
          </p>
          <p className="text-xs font-medium text-text-faint">
            Click Run Health Check to begin a network diagnostic
          </p>
        </div>
      )}
    </div>
  )
}
