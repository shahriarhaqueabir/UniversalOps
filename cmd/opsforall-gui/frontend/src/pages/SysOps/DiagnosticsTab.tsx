import { useQuery } from '@tanstack/react-query'
import { Activity, CheckCircle, AlertTriangle, XCircle, Clock, ChevronRight, History } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { ExtendedDiagnosticResult, ReportRecord } from '@/types'
import { cn } from '@/lib/utils'
import { useState } from 'react'

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

      {/* Check Results */}
      <div className="bg-panel border border-border rounded-[2rem] p-10 shadow-2xl">
        <h3 className="text-xl font-black text-text uppercase tracking-widest mb-8 flex items-center gap-3">
          <div className="w-1.5 h-6 bg-accent rounded-full" />
          Diagnostic Core Signals
        </h3>
        <div className="space-y-4">
          {diagnostics?.checks.map((check, i) => {
            const Icon = check.status === 'pass' ? CheckCircle : check.status === 'warn' ? AlertTriangle : XCircle
            const color = check.status === 'pass' ? 'text-success' : check.status === 'warn' ? 'text-warning' : 'text-danger'
            const bg = check.status === 'pass' ? 'bg-success/5' : check.status === 'warn' ? 'bg-warning/5' : 'bg-danger/5'
            const border = check.status === 'pass' ? 'border-success/20' : check.status === 'warn' ? 'border-warning/20' : 'border-danger/20'

            return (
              <div key={i} className={cn('flex items-center justify-between p-5 rounded-2xl border transition-all hover:translate-x-1 duration-300 group', bg, border)}>
                <div className="flex items-center gap-4">
                  <div className={cn('p-2.5 rounded-xl border bg-panel transition-transform duration-300 group-hover:scale-110', border)}>
                    <Icon size={20} className={color} />
                  </div>
                  <div>
                    <span className="text-sm font-black text-text uppercase tracking-wider">{check.name}</span>
                    <p className={cn('text-[10px] font-bold uppercase tracking-widest mt-0.5', color)}>{check.message}</p>
                  </div>
                </div>
                <div className="flex items-center gap-6">
                  <span className="text-lg font-black text-text tabular-nums tracking-tight">{check.value}</span>
                  <div className={cn("w-2 h-2 rounded-full", check.status === 'pass' ? "bg-success shadow-[0_0_8px_var(--color-success)]" : "bg-danger animate-pulse")} />
                </div>
              </div>
            )
          })}
        </div>
      </div>

      {/* Quick Health Summary */}
      <div className="bg-panel border border-border rounded-[2rem] p-10 shadow-2xl">
        <h3 className="text-lg font-black text-text uppercase tracking-[0.2em] mb-8">Baseline Verification</h3>
        <div className="grid grid-cols-2 gap-4">
          {[
            { label: 'CPU LOAD < 80%', pass: (diagnostics?.checks.find(c => c.name === 'CPU Usage')?.status || 'fail') === 'pass' },
            { label: 'MEMORY LOAD < 75%', pass: (diagnostics?.checks.find(c => c.name === 'Memory Usage')?.status || 'fail') === 'pass' },
            { label: 'DISK HEADROOM > 20%', pass: !diagnostics?.checks.some(c => c.name.startsWith('Disk') && c.status === 'fail') },
            { label: 'SWAP UTILIZATION NOMINAL', pass: (diagnostics?.checks.find(c => c.name === 'Swap Usage')?.status || 'pass') !== 'fail' },
            { label: 'THERMAL ENVELOPE STABLE', pass: (diagnostics?.checks.find(c => c.name === 'CPU Temperature')?.status || 'pass') !== 'fail' },
          ].map((item, i) => (
            <div key={i} className={cn('flex items-center gap-4 px-5 py-4 rounded-2xl border transition-all active:scale-[0.98]', item.pass ? 'bg-success/5 border-success/20' : 'bg-danger/5 border-danger/20')}>
              <div className={cn('p-1.5 rounded-full border', item.pass ? 'border-success/30' : 'border-danger/30')}>
                {item.pass ? <CheckCircle size={14} className="text-success" /> : <XCircle size={14} className="text-danger" />}
              </div>
              <span className={cn('text-[10px] font-black uppercase tracking-widest', item.pass ? 'text-success' : 'text-danger')}>{item.label}</span>
            </div>
          ))}
        </div>
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
