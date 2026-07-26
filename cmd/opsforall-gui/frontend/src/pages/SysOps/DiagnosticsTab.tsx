import { useQuery } from '@tanstack/react-query'
import { Activity, CheckCircle, AlertTriangle, XCircle, Clock, ChevronRight, History, Info } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { ExtendedDiagnosticResult, ReportRecord } from '@/types'
import { cn } from '@/lib/utils'
import { useState } from 'react'

// ── Helpers ──

/** Extract numeric percentage from a value string like "95.2%" or "72.1°C" */
function extractPercent(value: string): number {
  const m = value.match(/^([\d.]+)/)
  return m ? parseFloat(m[1]) : 0
}

/** Determine check status color classes with smooth gradient for near-threshold states */
function statusVisuals(status: string) {
  if (status === 'pass') return {
    text: 'text-success',
    bg: 'bg-success/5',
    border: 'border-success/20',
    bar: 'bg-gradient-to-r from-success/80 to-success',
    glow: 'shadow-[0_0_8px_var(--color-success)]',
    pulse: false,
  }
  if (status === 'warn') return {
    text: 'text-warning',
    bg: 'bg-warning/5',
    border: 'border-warning/20',
    bar: 'bg-gradient-to-r from-warning/80 to-warning',
    glow: 'shadow-[0_0_8px_var(--color-warning)]',
    pulse: false,
  }
  return {
    text: 'text-danger',
    bg: 'bg-danger/5',
    border: 'border-danger/20',
    bar: 'bg-gradient-to-r from-danger/80 to-danger',
    glow: 'shadow-[0_0_8px_var(--color-danger)]',
    pulse: true,
  }
}

/** Health bar component with threshold markers */
function HealthBar({ value, warnAt = 80, failAt = 90, className }: { value: number; warnAt?: number; failAt?: number; className?: string }) {
  const clamped = Math.min(value, 100)
  const barColor = value >= failAt ? 'bg-danger' : value >= warnAt ? 'bg-warning' : 'bg-success'

  return (
    <div className={cn('relative h-2 bg-[var(--color-panel-2)] rounded-full overflow-hidden', className)}>
      {/* Fill */}
      <div
        className={cn('h-full rounded-full transition-all duration-700 ease-out', barColor)}
        style={{ width: `${clamped}%` }}
      />
      {/* Warn threshold marker */}
      <div className="absolute top-0 bottom-0 w-0.5 bg-text-faint/30" style={{ left: `${warnAt}%` }} title={`Warn threshold: ${warnAt}%`} />
      {/* Fail threshold marker */}
      <div className="absolute top-0 bottom-0 w-0.5 bg-danger/50" style={{ left: `${failAt}%` }} title={`Fail threshold: ${failAt}%`} />
    </div>
  )
}

export function DiagnosticsTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const [selectedReport, setSelectedReport] = useState<string | null>(null)

  const { data: diagnostics, isLoading } = useQuery<ExtendedDiagnosticResult>({
    queryKey: ['sysops-diagnostics'],
    queryFn: async () => { const r = await call('SysOps.RunExtendedDiagnostics'); return r as ExtendedDiagnosticResult },
    refetchInterval: refreshInterval,
  })

  const { data: history } = useQuery<ReportRecord[]>({
    queryKey: ['sysops-diagnostics-history'],
    queryFn: async () => { const r = await call('SysOps.ListHistoricalHealthReports'); return r as ReportRecord[] },
    refetchInterval: refreshInterval,
  })

  const { data: detailReport } = useQuery<ExtendedDiagnosticResult>({
    queryKey: ['sysops-diagnostics-detail', selectedReport],
    queryFn: async () => { const r = await call('SysOps.GetHistoricalHealthReport', selectedReport!); return r as ExtendedDiagnosticResult },
    enabled: !!selectedReport,
  })

  if (isLoading) {
    return <div className="animate-pulse space-y-4">{[1, 2, 3].map(i => <div key={i} className="h-16 bg-[var(--color-panel-2)] rounded-xl" />)}</div>
  }

  const scoreColor = (diagnostics?.score || 0) >= 80 ? 'text-[var(--color-success)]' : (diagnostics?.score || 0) >= 60 ? 'text-[var(--color-warning)]' : 'text-[var(--color-danger)]'

  return (
    <div className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-700">
      {/* Score Card */}
      <div className="bg-panel border border-border rounded-[2.5rem] p-10 shadow-2xl text-center group relative overflow-hidden">
        <div className="absolute inset-0 bg-accent/5 opacity-0 group-hover:opacity-100 transition-opacity duration-700 pointer-events-none" />
        <Activity size={40} className="text-accent mx-auto mb-6 transition-transform duration-500 group-hover:scale-110" />
        <p className={cn('text-8xl font-black tabular-nums tracking-tighter drop-shadow-sm', scoreColor)}>{diagnostics?.score || 0}</p>
        <p className="text-xs font-black text-text-faint uppercase tracking-[0.3em] mt-4">System Health Index</p>
      </div>

      {/* Check Results with Health Bars */}
      <div className="bg-panel border border-border rounded-[2rem] p-10 shadow-2xl">
        <h3 className="text-xl font-black text-text uppercase tracking-widest mb-8 flex items-center gap-3">
          <div className="w-1.5 h-6 bg-accent rounded-full" />
          Diagnostic Core Signals
        </h3>
        <div className="space-y-3">
          {diagnostics?.checks.map((check, i) => {
            const pct = extractPercent(check.value)
            const v = statusVisuals(check.status)
            const Icon = check.status === 'pass' ? CheckCircle : check.status === 'warn' ? AlertTriangle : XCircle

            return (
              <div key={i} className={cn('p-5 rounded-2xl border transition-all hover:translate-x-1 duration-300 group', v.bg, v.border)}>
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-4">
                    <div className={cn('p-2.5 rounded-xl border bg-panel transition-transform duration-300 group-hover:scale-110', v.border)}>
                      <Icon size={20} className={v.text} />
                    </div>
                    <div>
                      <span className="text-sm font-black text-text uppercase tracking-wider">{check.name}</span>
                      <p className={cn('text-[10px] font-semibold leading-relaxed max-w-lg mt-0.5', v.text)}>{check.message}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-4 flex-shrink-0">
                    <span className={cn('text-lg font-black tabular-nums tracking-tight', v.text)}>{check.value}</span>
                    <div className={cn("w-2 h-2 rounded-full", v.pulse ? "bg-danger animate-pulse" : `${v.glow} bg-current`)} />
                  </div>
                </div>
                {/* Health bar */}
                <HealthBar
                  value={pct}
                  warnAt={80}
                  failAt={check.name === 'CPU Usage' || check.name === 'Memory Usage' ? 90 : check.name.startsWith('Disk') ? 95 : check.name === 'CPU Temperature' ? 85 : 90}
                />
              </div>
            )
          })}
        </div>
      </div>

      {/* Baseline Verification with Health Bars */}
      <div className="bg-panel border border-border rounded-[2rem] p-10 shadow-2xl">
        <h3 className="text-lg font-black text-text uppercase tracking-[0.2em] mb-6 flex items-center gap-3">
          <div className="w-1.5 h-6 bg-accent rounded-full" />
          <Info size={16} className="text-accent" />
          Baseline Verification
        </h3>

        {(() => {
          const cpuCheck = diagnostics?.checks.find(c => c.name === 'CPU Usage')
          const memCheck = diagnostics?.checks.find(c => c.name === 'Memory Usage')
          const diskChecks = diagnostics?.checks.filter(c => c.name.startsWith('Disk'))
          const swapCheck = diagnostics?.checks.find(c => c.name === 'Swap Usage')
          const tempCheck = diagnostics?.checks.find(c => c.name === 'CPU Temperature')
          const diskFail = diskChecks?.some(c => c.status === 'fail')
          const diskWarn = diskChecks?.some(c => c.status === 'warn')

          const baselines = [
            { label: 'CPU LOAD < 80%', check: cpuCheck, pass: cpuCheck?.status === 'pass', value: cpuCheck?.value || '-', pct: extractPercent(cpuCheck?.value || '0') },
            { label: 'MEMORY LOAD < 80%', check: memCheck, pass: memCheck?.status === 'pass', value: memCheck?.value || '-', pct: extractPercent(memCheck?.value || '0') },
            { label: 'DISK HEADROOM > 15%', check: diskChecks?.[0], pass: !diskFail, value: diskFail ? 'CRITICAL' : diskWarn ? 'LOW' : 'OK', pct: diskChecks ? Math.max(...diskChecks.map(c => extractPercent(c.value))) : 0 },
            { label: 'SWAP UTILIZATION NOMINAL', check: swapCheck, pass: swapCheck?.status !== 'fail', value: swapCheck?.value || '-', pct: extractPercent(swapCheck?.value || '0') },
            { label: 'THERMAL ENVELOPE STABLE', check: tempCheck, pass: tempCheck?.status !== 'fail', value: tempCheck?.value || '-', pct: extractPercent(tempCheck?.value || '0') },
          ]

          return (
            <div className="grid grid-cols-2 gap-4">
              {baselines.map((item, i) => {
                const ok = item.pass
                const status = item.check?.status || 'pass'
                const color = ok ? 'text-success' : 'text-danger'
                const bg = ok ? 'bg-success/5' : status === 'warn' ? 'bg-warning/5' : 'bg-danger/5'
                const border = ok ? 'border-success/20' : status === 'warn' ? 'border-warning/20' : 'border-danger/20'

                return (
                  <div key={i} className={cn('p-4 rounded-2xl border transition-all active:scale-[0.98]', bg, border)}>
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center gap-3">
                        <div className={cn('p-1 rounded-full border', ok ? 'border-success/30' : 'border-danger/30')}>
                          {ok ? <CheckCircle size={12} className="text-success" /> : <XCircle size={12} className="text-danger" />}
                        </div>
                        <span className={cn('text-[10px] font-black uppercase tracking-widest', color)}>{item.label}</span>
                      </div>
                      <span className={cn('text-xs font-bold tabular-nums', color)}>{item.value}</span>
                    </div>
                    {/* Mini health bar */}
                    {item.pct > 0 && (
                      <div className="relative h-1.5 bg-panel-2 rounded-full overflow-hidden ml-7">
                        <div
                          className={cn('h-full rounded-full transition-all duration-700 ease-out', ok ? 'bg-success' : status === 'warn' ? 'bg-warning' : 'bg-danger')}
                          style={{ width: `${Math.min(item.pct, 100)}%` }}
                        />
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          )
        })()}
      </div>

      {/* Historical Reports */}
      <div className="bg-panel border border-border rounded-[2rem] p-10 shadow-2xl">
        <h3 className="text-lg font-black text-text uppercase tracking-[0.2em] mb-8 flex items-center gap-3">
          <div className="w-1.5 h-6 bg-accent rounded-full" />
          <History size={20} className="text-accent" />
          Diagnostic History
        </h3>

        {selectedReport && detailReport ? (
          /* Detail View */
          <div className="space-y-6">
            <button
              onClick={() => setSelectedReport(null)}
              className="text-xs font-black text-accent uppercase tracking-widest hover:opacity-70 transition-opacity"
            >
              &larr; Back to History
            </button>
            <div className="grid grid-cols-2 gap-4 mb-6">
              <div className="bg-panel-2 rounded-2xl p-5 border border-border">
                <p className="text-[10px] font-black text-text-faint uppercase tracking-widest mb-1">Score</p>
                <p className={cn('text-3xl font-black tabular-nums', detailReport.score >= 80 ? 'text-success' : detailReport.score >= 60 ? 'text-warning' : 'text-danger')}>
                  {detailReport.score}
                </p>
              </div>
              <div className="bg-panel-2 rounded-2xl p-5 border border-border">
                <p className="text-[10px] font-black text-text-faint uppercase tracking-widest mb-1">Timestamp</p>
                <p className="text-sm font-bold text-text tabular-nums">{new Date(detailReport.timestamp).toLocaleString()}</p>
              </div>
            </div>
            <div className="space-y-3">
              {detailReport.checks.map((check, i) => {
                const Icon = check.status === 'pass' ? CheckCircle : check.status === 'warn' ? AlertTriangle : XCircle
                const color = check.status === 'pass' ? 'text-success' : check.status === 'warn' ? 'text-warning' : 'text-danger'
                return (
                  <div key={i} className="flex items-center justify-between p-4 rounded-xl bg-panel-2 border border-border">
                    <div className="flex items-center gap-3">
                      <Icon size={16} className={color} />
                      <span className="text-xs font-black text-text uppercase tracking-wider">{check.name}</span>
                    </div>
                    <span className={cn('text-xs font-bold tabular-nums', color)}>{check.value}</span>
                  </div>
                )
              })}
            </div>
          </div>
        ) : (
          /* List View */
          <div className="space-y-3">
            {history && history.length > 0 ? (
              history.map((report) => {
                const hColor = report.score >= 80 ? 'text-success' : report.score >= 60 ? 'text-warning' : 'text-danger'
                const hBg = report.score >= 80 ? 'bg-success/5' : report.score >= 60 ? 'bg-warning/5' : 'bg-danger/5'
                const hBorder = report.score >= 80 ? 'border-success/20' : report.score >= 60 ? 'border-warning/20' : 'border-danger/20'
                return (
                  <button
                    key={report.id}
                    onClick={() => setSelectedReport(report.id)}
                    className={cn('w-full flex items-center justify-between p-5 rounded-2xl border transition-all hover:translate-x-1 duration-300 group', hBg, hBorder)}
                  >
                    <div className="flex items-center gap-4">
                      <div className={cn('p-2.5 rounded-xl border bg-panel', hBorder)}>
                        <Clock size={18} className={hColor} />
                      </div>
                      <div className="text-left">
                        <p className="text-xs font-black text-text uppercase tracking-wider">{new Date(report.timestamp).toLocaleDateString()}</p>
                        <p className="text-[10px] font-bold text-text-faint uppercase tracking-widest mt-0.5">{new Date(report.timestamp).toLocaleTimeString()}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-4">
                      <span className={cn('text-2xl font-black tabular-nums tracking-tight', hColor)}>{report.score}</span>
                      <ChevronRight size={18} className="text-text-faint transition-transform duration-300 group-hover:translate-x-1" />
                    </div>
                  </button>
                )
              })
            ) : (
              <div className="text-center py-12">
                <Clock size={32} className="text-text-faint mx-auto mb-4" />
                <p className="text-sm font-black text-text-faint uppercase tracking-widest">No historical reports yet</p>
                <p className="text-[10px] font-bold text-text-faint/60 uppercase tracking-widest mt-2">Run diagnostics to populate history</p>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
