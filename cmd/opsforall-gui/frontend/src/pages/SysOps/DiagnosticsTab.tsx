import { useQuery } from '@tanstack/react-query'
import { Activity, CheckCircle, AlertTriangle, XCircle } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { ExtendedDiagnosticResult } from '@/types'
import { cn } from '@/lib/utils'

export function DiagnosticsTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: diagnostics, isLoading } = useQuery<ExtendedDiagnosticResult>({
    queryKey: ['sysops-diagnostics'],
    queryFn: async () => { const r = await call('SysOps.RunExtendedDiagnostics'); return r as ExtendedDiagnosticResult },
    refetchInterval: refreshInterval,
  })

  if (isLoading) {
    return <div className="animate-pulse space-y-4">{[1, 2, 3].map(i => <div key={i} className="h-16 bg-[var(--color-panel-2)] rounded-xl" />)}</div>
  }

  const scoreColor = (diagnostics?.score || 0) >= 80 ? 'text-[var(--color-success)]' : (diagnostics?.score || 0) >= 60 ? 'text-[var(--color-warning)]' : 'text-[var(--color-danger)]'

  return (
    <div className="space-y-8">
      {/* Score Card */}
      <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-8 shadow-2xl text-center">
        <Activity size={32} className="text-[var(--color-accent)] mx-auto mb-4" />
        <p className={cn('text-6xl font-bold tabular-nums', scoreColor)}>{diagnostics?.score || 0}</p>
        <p className="text-sm font-bold text-[var(--color-text-faint)] uppercase tracking-widest mt-2">Health Score</p>
      </div>

      {/* Check Results */}
      <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-8 shadow-2xl">
        <h3 className="text-lg font-bold text-[var(--color-text)] uppercase tracking-widest mb-6">Diagnostic Checks</h3>
        <div className="space-y-3">
          {diagnostics?.checks.map((check, i) => {
            const Icon = check.status === 'pass' ? CheckCircle : check.status === 'warn' ? AlertTriangle : XCircle
            const color = check.status === 'pass' ? 'text-[var(--color-success)]' : check.status === 'warn' ? 'text-[var(--color-warning)]' : 'text-[var(--color-danger)]'
            const bg = check.status === 'pass' ? 'bg-[var(--color-success)]/10' : check.status === 'warn' ? 'bg-[var(--color-warning)]/10' : 'bg-[var(--color-danger)]/10'

            return (
              <div key={i} className={cn('flex items-center justify-between p-4 rounded-xl border', bg, 'border-[var(--color-border)]/50')}>
                <div className="flex items-center gap-3">
                  <Icon size={18} className={color} />
                  <span className="text-sm font-bold text-[var(--color-text)]">{check.name}</span>
                </div>
                <div className="flex items-center gap-4">
                  <span className="text-sm font-bold text-[var(--color-text)] tabular-nums">{check.value}</span>
                  <span className={cn('text-xs font-bold', color)}>{check.message}</span>
                </div>
              </div>
            )
          })}
        </div>
      </div>

      {/* Quick Health Summary */}
      <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-8 shadow-2xl">
        <h3 className="text-lg font-bold text-[var(--color-text)] uppercase tracking-widest mb-4">Quick Health Check</h3>
        <div className="grid grid-cols-2 gap-3">
          {[
            { label: 'CPU < 80%', pass: (diagnostics?.checks.find(c => c.name === 'CPU Usage')?.status || 'fail') === 'pass' },
            { label: 'Memory < 75%', pass: (diagnostics?.checks.find(c => c.name === 'Memory Usage')?.status || 'fail') === 'pass' },
            { label: 'Disk > 20% free', pass: !diagnostics?.checks.some(c => c.name.startsWith('Disk') && c.status === 'fail') },
            { label: 'Swap OK', pass: (diagnostics?.checks.find(c => c.name === 'Swap Usage')?.status || 'pass') !== 'fail' },
            { label: 'Temperature OK', pass: (diagnostics?.checks.find(c => c.name === 'CPU Temperature')?.status || 'pass') !== 'fail' },
          ].map((item, i) => (
            <div key={i} className={cn('flex items-center gap-2 px-3 py-2 rounded-lg border', item.pass ? 'bg-[var(--color-success)]/10 border-[var(--color-success)]/30' : 'bg-[var(--color-danger)]/10 border-[var(--color-danger)]/30')}>
              {item.pass ? <CheckCircle size={14} className="text-[var(--color-success)]" /> : <XCircle size={14} className="text-[var(--color-danger)]" />}
              <span className={cn('text-sm font-bold', item.pass ? 'text-[var(--color-success)]' : 'text-[var(--color-danger)]')}>{item.label}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
