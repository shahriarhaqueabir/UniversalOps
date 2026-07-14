import { useQuery } from '@tanstack/react-query'
import { Shield, ShieldCheck, ShieldAlert, Users, Radio, AlertTriangle, Zap } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { SectionBriefing, MiniStat, StatusBadge } from './components'
import type { SecurityScore, DefenderStatus, FirewallStatus, UserInfo, ListeningPort, SecurityEvent } from '@/types'

export function OverviewTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: score } = useQuery<SecurityScore>({
    queryKey: ['secops-score'],
    queryFn: async () => await call('SecOps.GetSecurityScore') as SecurityScore,
    refetchInterval: refreshInterval,
  })

  const { data: defender } = useQuery<DefenderStatus | null>({
    queryKey: ['secops-defender'],
    queryFn: async () => await call('SecOps.GetDefenderStatus') as DefenderStatus,
    refetchInterval: refreshInterval,
  })

  const { data: fwStatus } = useQuery<FirewallStatus | null>({
    queryKey: ['secops-firewall-status'],
    queryFn: async () => await call('SecOps.GetFirewallStatus') as FirewallStatus,
    refetchInterval: refreshInterval,
  })

  const { data: users = [] } = useQuery<UserInfo[]>({
    queryKey: ['secops-users'],
    queryFn: async () => (await call('SecOps.GetUsers') as UserInfo[]) || [],
    refetchInterval: refreshInterval,
  })

  const { data: ports = [] } = useQuery<ListeningPort[]>({
    queryKey: ['secops-listening'],
    queryFn: async () => (await call('SecOps.GetListeningPorts') as ListeningPort[]) || [],
    refetchInterval: refreshInterval,
  })

  const { data: events = [] } = useQuery<SecurityEvent[]>({
    queryKey: ['secops-events'],
    queryFn: async () => (await call('SecOps.GetSecurityEvents') as SecurityEvent[]) || [],
    refetchInterval: refreshInterval,
  })

  const gradeColor: Record<string, string> = {
    A: 'text-success', B: 'text-accent', C: 'text-warning', D: 'text-danger', F: 'text-danger',
  }

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <SectionBriefing
        title="Security Operations Center"
        objective="Unified view of your workstation's security posture across firewall, endpoint protection, identity, and network exposure."
        checklist={['Firewall & perimeter defense', 'Endpoint protection status', 'User account audit', 'Listening port exposure', 'Security event analysis']}
      />

      {/* Score */}
      {score && (
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl flex items-center gap-8">
          <div className="text-center">
            <p className="text-6xl font-black tabular-nums text-text">{score.score}</p>
            <p className={`text-2xl font-bold ${gradeColor[score.grade] || 'text-danger'}`}>Grade {score.grade}</p>
          </div>
          <div className="flex-1 grid grid-cols-3 gap-4">
            {Object.entries(score.breakdown).map(([cat, val]) => (
              <div key={cat} className="bg-panel-2 border border-border rounded-xl p-4">
                <p className="text-xs font-bold text-text-faint uppercase tracking-wider mb-1">{cat}</p>
                <p className="text-xl font-bold text-text tabular-nums">{val}</p>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Quick Stats */}
      <div className="grid grid-cols-4 gap-4">
        <MiniStat label="Security Score" value={score?.score ?? '—'} icon={<Shield size={24} />} variant={score && score.score >= 75 ? 'success' : score && score.score >= 50 ? 'warning' : 'danger'} />
        <MiniStat label="Active Admins" value={users.filter(u => u.is_admin && u.is_enabled).length} icon={<Users size={24} />} variant={users.filter(u => u.is_admin && u.is_enabled).length <= 2 ? 'success' : 'warning'} />
        <MiniStat label="External Ports" value={ports.filter(p => p.is_external).length} icon={<Radio size={24} />} variant={ports.filter(p => p.is_external).length === 0 ? 'success' : 'warning'} />
        <MiniStat label="Security Events" value={events.length} icon={<AlertTriangle size={24} />} variant={events.filter(e => e.level === 'Error').length === 0 ? 'success' : 'danger'} />
      </div>

      {/* Defender & Firewall Status */}
      <div className="grid grid-cols-2 gap-8">
        {defender && (
          <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-xl">
            <h3 className="text-lg font-bold text-text uppercase tracking-widest mb-6 flex items-center gap-3">
              <ShieldCheck size={22} className="text-accent" /> Endpoint Protection
            </h3>
            <div className="space-y-4">
              {[
                { label: 'Defender', active: defender.enabled },
                { label: 'Real-time Protection', active: defender.real_time_protection },
                { label: 'Cloud Protection', active: defender.cloud_protection },
                { label: 'Signatures Up-to-date', active: defender.up_to_date },
              ].map(m => (
                <div key={m.label} className="flex items-center justify-between">
                  <span className="text-sm font-bold text-text-dim">{m.label}</span>
                  <StatusBadge status={m.active ? 'enabled' : 'disabled'} />
                </div>
              ))}
            </div>
          </div>
        )}

        {fwStatus && (
          <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-xl">
            <h3 className="text-lg font-bold text-text uppercase tracking-widest mb-6 flex items-center gap-3">
              <ShieldAlert size={22} className="text-warning" /> Perimeter Defense
            </h3>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm font-bold text-text-dim">Global Firewall</span>
                <StatusBadge status={fwStatus.enabled ? 'enabled' : 'disabled'} />
              </div>
              {fwStatus.profiles.map(p => (
                <div key={p.name} className="flex items-center justify-between">
                  <span className="text-sm font-bold text-text-dim">{p.name}</span>
                  <StatusBadge status={p.enabled ? 'enabled' : 'disabled'} />
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Recommendations */}
      {score && score.recommendations.length > 0 && (
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-xl">
          <h3 className="text-lg font-bold text-text uppercase tracking-widest mb-6 flex items-center gap-3">
            <Zap size={22} className="text-warning" /> Recommendations
          </h3>
          <div className="space-y-3">
            {score.recommendations.map((rec, i) => (
              <div key={i} className="flex items-start gap-3 bg-panel-2 border border-border rounded-xl px-4 py-3">
                <AlertTriangle size={14} className="text-warning mt-0.5 shrink-0" />
                <span className="text-sm text-text-dim leading-snug">{rec}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
