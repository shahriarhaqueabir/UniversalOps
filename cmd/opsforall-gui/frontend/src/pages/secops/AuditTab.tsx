import { useMutation } from '@tanstack/react-query'
import { ClipboardCheck, CheckCircle2, AlertTriangle, RefreshCw } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { SectionBriefing } from '@/components/ui/SectionBriefing'
import { StatusBadge } from '@/components/ui/StatusBadge'
import { Panel } from '@/components/ui/Panel'
import type { SecurityAuditResult } from '@/types'

export function AuditTab() {
  const { call } = useBackend()

  const auditMutation = useMutation({
    mutationFn: async () => {
      return await call('SecOps.RunSecurityAuditChecklist') as SecurityAuditResult
    },
  })

  const runAudit = () => auditMutation.mutate()

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <SectionBriefing
        title="Security Audit"
        objective="Run a comprehensive one-click security audit that checks firewall, encryption, certificates, passwords, and more."
        checklist={['Firewall enabled', 'Secure Boot active', 'Disk encrypted', 'No exposed RDP/SSH', 'Password policy compliant', 'Certificates valid', 'No suspicious persistence']}
      />

      <div className="flex items-center gap-4">
        <button
          onClick={runAudit}
          disabled={auditMutation.isPending}
          className="flex items-center gap-3 px-6 py-3 rounded-xl bg-accent text-white font-bold text-sm uppercase tracking-wider shadow-lg hover:bg-accent/80 disabled:opacity-50 transition-all"
        >
          {auditMutation.isPending ? <RefreshCw size={18} className="animate-spin" /> : <ClipboardCheck size={18} />}
          {auditMutation.isPending ? 'Running Audit...' : 'Run Security Audit'}
        </button>
      </div>

      {auditMutation.isPending && (
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
          <div className="flex items-center gap-4 mb-8">
            <div className="h-6 w-48 bg-panel-3 rounded animate-pulse" />
            <div className="h-6 w-32 bg-panel-3 rounded animate-pulse ml-auto" />
          </div>
          <div className="space-y-4">
            {[1,2,3,4,5,6,7,8].map((i) => (
              <div key={i} className="flex items-center gap-4">
                <div className="h-5 w-5 bg-panel-3 rounded animate-pulse" />
                <div className="h-4 flex-1 bg-panel-3 rounded animate-pulse" />
                <div className="h-5 w-20 bg-panel-3 rounded animate-pulse" />
              </div>
            ))}
          </div>
        </div>
      )}

      {auditMutation.isError && (
        <div className="bg-danger/10 border border-danger/30 rounded-[var(--radius-lg)] p-6 flex items-start gap-3">
          <AlertTriangle size={20} className="text-danger shrink-0 mt-0.5" />
          <div>
            <p className="text-sm font-semibold text-danger">Audit failed</p>
            <p className="text-sm text-[var(--color-text-dim)] mt-1">{String(auditMutation.error)}</p>
          </div>
        </div>
      )}

      {auditMutation.data && (
        <>
          {/* Score */}
          <Panel padding="lg" category="security" className="flex items-center gap-8">
            <div className="text-center">
              <p className="text-6xl font-black tabular-nums text-text">{auditMutation.data.score}%</p>
              <p className="text-sm font-bold text-text-faint uppercase tracking-wider mt-2">
                {auditMutation.data.passed}/{auditMutation.data.total} passed
              </p>
            </div>
            <div className="flex-1">
              <div className="h-4 bg-panel-3 rounded-full overflow-hidden">
                <div
                  className="h-full rounded-full transition-all duration-700"
                  style={{
                    width: `${auditMutation.data.score}%`,
                    backgroundColor: auditMutation.data.score >= 80 ? 'var(--color-success)' : auditMutation.data.score >= 50 ? 'var(--color-warning)' : 'var(--color-danger)',
                  }}
                />
              </div>
              <p className="text-xs text-text-faint mt-2">Last run: {new Date(auditMutation.data.timestamp).toLocaleString()}</p>
            </div>
          </Panel>

          {/* Checklist Items */}
          <Panel padding="lg" category="security">
            <h3 className="text-lg font-bold text-text uppercase tracking-widest mb-6">Audit Results</h3>
            <div className="space-y-3">
              {auditMutation.data.items.map((item, i) => (
                <div key={i} className={`flex items-center justify-between bg-panel-2 border rounded-xl px-5 py-3 ${item.passed ? 'border-border' : 'border-danger/30 bg-danger/5'}`}>
                  <div className="flex items-center gap-4">
                    {item.passed ? (
                      <CheckCircle2 size={18} className="text-success shrink-0" />
                    ) : (
                      <AlertTriangle size={18} className="text-danger shrink-0" />
                    )}
                    <div>
                      <span className="text-sm font-bold text-text">{item.check}</span>
                      <span className="text-xs text-text-faint ml-2">({item.category})</span>
                    </div>
                  </div>
                  <div className="flex items-center gap-4">
                    <StatusBadge status={item.passed ? 'passed' : 'failed'} />
                    {!item.passed && (
                      <span className="text-xs text-text-dim max-w-xs truncate">{item.remediation}</span>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </Panel>
        </>
      )}

      {!auditMutation.data && !auditMutation.isPending && !auditMutation.isError && (
        <Panel padding="lg" category="security" className="text-center">
          <ClipboardCheck size={48} className="text-text-faint mx-auto mb-4" />
          <p className="text-text-dim text-sm">Click "Run Security Audit" to analyze your system's security posture.</p>
        </Panel>
      )}
    </div>
  )
}
