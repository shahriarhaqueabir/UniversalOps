import { useQuery } from '@tanstack/react-query'
import { ShieldCheck, ShieldAlert, AlertTriangle, CheckCircle2 } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { SectionBriefing, MiniStat, StatusBadge } from './components'
import type { HardeningCheck, SSHConfig } from '@/types'

export function HardeningTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: checks = [] } = useQuery<HardeningCheck[]>({
    queryKey: ['secops-hardening'],
    queryFn: async () => (await call('SecOps.GetHardeningChecks') as HardeningCheck[]) || [],
    refetchInterval: refreshInterval,
  })

  const { data: sshConfig } = useQuery<SSHConfig>({
    queryKey: ['secops-ssh-config'],
    queryFn: async () => await call('SecOps.GetSSHConfig') as SSHConfig,
    refetchInterval: refreshInterval,
  })

  const passedCount = checks.filter(c => c.passed).length
  const failedChecks = checks.filter(c => !c.passed)

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <SectionBriefing
        title="Security Hardening"
        objective="Evaluate system hardening against security baselines. Failed checks indicate deviations from recommended security configuration."
        checklist={['Firewall enabled', 'Antivirus active', 'Dangerous protocols disabled', 'SSH configuration review', 'Guest account disabled']}
      />

      <div className="grid grid-cols-4 gap-4">
        <MiniStat label="Total Checks" value={checks.length} icon={<ShieldCheck size={24} />} variant="default" />
        <MiniStat label="Passed" value={passedCount} icon={<CheckCircle2 size={24} />} variant="success" />
        <MiniStat label="Failed" value={failedChecks.length} icon={<ShieldAlert size={24} />} variant={failedChecks.length === 0 ? 'success' : 'danger'} />
        <MiniStat label="Score" value={checks.length > 0 ? `${Math.round((passedCount / checks.length) * 100)}%` : '—'} icon={<ShieldCheck size={24} />} variant={passedCount === checks.length ? 'success' : 'warning'} />
      </div>

      {/* Hardening Checks */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-xl">
        <h3 className="text-lg font-bold text-text uppercase tracking-widest mb-6">Hardening Checks</h3>
        <div className="space-y-3">
          {checks.map((c, i) => (
            <div key={i} className={`flex items-center justify-between bg-panel-2 border rounded-xl px-4 py-3 ${c.passed ? 'border-border' : 'border-danger/30 bg-danger/5'}`}>
              <div className="flex items-center gap-4">
                {c.passed ? (
                  <CheckCircle2 size={18} className="text-success shrink-0" />
                ) : (
                  <AlertTriangle size={18} className="text-danger shrink-0" />
                )}
                <div>
                  <span className="text-sm font-bold text-text">{c.check}</span>
                  <span className="text-xs text-text-faint ml-2">({c.category})</span>
                </div>
              </div>
              <div className="flex items-center gap-4">
                <StatusBadge status={c.severity} />
                {!c.passed && (
                  <span className="text-xs text-text-dim max-w-xs truncate">{c.remediation}</span>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* SSH Config */}
      {sshConfig && sshConfig.permit_root_login !== 'unknown' && (
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-xl">
          <h3 className="text-lg font-bold text-text uppercase tracking-widest mb-6 flex items-center gap-3">
            <ShieldAlert size={22} className="text-accent" /> SSH Configuration
          </h3>
          <div className="grid grid-cols-2 gap-4">
            {Object.entries(sshConfig).map(([key, val]) => (
              <div key={key} className="bg-panel-2 border border-border rounded-xl px-4 py-3 flex items-center justify-between">
                <span className="text-sm font-bold text-text-dim uppercase tracking-wider">{key.replace(/_/g, ' ')}</span>
                <StatusBadge status={val === 'no' || val === 'none' ? 'disabled' : val === 'yes' ? 'enabled' : val} />
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
