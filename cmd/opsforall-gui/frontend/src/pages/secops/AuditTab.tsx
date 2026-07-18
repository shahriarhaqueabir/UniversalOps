import { useState } from 'react'
import { ClipboardCheck, CheckCircle2, AlertTriangle, RefreshCw } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { SectionBriefing } from '@/components/ui/SectionBriefing'
import { StatusBadge } from '@/components/ui/StatusBadge'
import { Panel } from '@/components/ui/Panel'
import type { SecurityAuditResult } from '@/types'

export function AuditTab() {
  const { call } = useBackend()
  const [auditResult, setAuditResult] = useState<SecurityAuditResult | null>(null)
  const [isRunning, setIsRunning] = useState(false)

  const runAudit = async () => {
    setIsRunning(true)
    try {
      const result = await call('SecOps.RunSecurityAuditChecklist') as SecurityAuditResult
      setAuditResult(result)
    } finally {
      setIsRunning(false)
    }
  }

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
          disabled={isRunning}
          className="flex items-center gap-3 px-6 py-3 rounded-xl bg-accent text-white font-bold text-sm uppercase tracking-wider shadow-lg hover:bg-accent/80 disabled:opacity-50 transition-all"
        >
          {isRunning ? <RefreshCw size={18} className="animate-spin" /> : <ClipboardCheck size={18} />}
          {isRunning ? 'Running Audit...' : 'Run Security Audit'}
        </button>
      </div>

      {auditResult && (
        <>
          {/* Score */}
          <Panel padding="lg" category="security" className="flex items-center gap-8">
            <div className="text-center">
              <p className="text-6xl font-black tabular-nums text-text">{auditResult.score}%</p>
              <p className="text-sm font-bold text-text-faint uppercase tracking-wider mt-2">
                {auditResult.passed}/{auditResult.total} passed
              </p>
            </div>
            <div className="flex-1">
              <div className="h-4 bg-panel-3 rounded-full overflow-hidden">
                <div
                  className="h-full rounded-full transition-all duration-700"
                  style={{
                    width: `${auditResult.score}%`,
                    backgroundColor: auditResult.score >= 80 ? 'var(--color-success)' : auditResult.score >= 50 ? 'var(--color-warning)' : 'var(--color-danger)',
                  }}
                />
              </div>
              <p className="text-xs text-text-faint mt-2">Last run: {new Date(auditResult.timestamp).toLocaleString()}</p>
            </div>
          </Panel>

          {/* Checklist Items */}
          <Panel padding="lg" category="security">
            <h3 className="text-lg font-bold text-text uppercase tracking-widest mb-6">Audit Results</h3>
            <div className="space-y-3">
              {auditResult.items.map((item, i) => (
                <div key={i} className={`flex items-center justify-between bg-panel-2 border rounded-xl px-4 py-3 ${item.passed ? 'border-border' : 'border-danger/30 bg-danger/5'}`}>
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

      {!auditResult && !isRunning && (
        <Panel padding="lg" category="security" className="text-center">
          <ClipboardCheck size={48} className="text-text-faint mx-auto mb-4" />
          <p className="text-text-dim text-sm">Click "Run Security Audit" to analyze your system's security posture.</p>
        </Panel>
      )}
    </div>
  )
}
