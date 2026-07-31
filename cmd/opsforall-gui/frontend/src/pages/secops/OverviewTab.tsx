import { useQuery } from '@tanstack/react-query'
import { Shield, ShieldCheck, ShieldAlert, Users, Radio, AlertTriangle, Zap } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { SectionBriefing } from '@/components/ui/SectionBriefing'
import { MiniStat } from '@/components/ui/MiniStat'
import { StatusBadge } from '@/components/ui/StatusBadge'
import { Panel } from '@/components/ui/Panel'
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
        <Panel padding="lg" category="security" className="flex items-center gap-12 group overflow-hidden relative">
          <div className={cn(
            "absolute top-0 right-0 w-64 h-64 opacity-5 rounded-bl-full pointer-events-none transition-opacity group-hover:opacity-10",
            gradeColor[score.grade]?.replace('text-', 'bg-') || 'bg-danger'
          )} />

          <div className="text-center relative z-10">
            <p className="text-8xl font-black tabular-nums text-text tracking-tighter drop-shadow-md">{score.score}</p>
            <p className={cn("text-2xl font-black uppercase tracking-[0.2em] mt-2", gradeColor[score.grade] || 'text-danger')}>
              Grade {score.grade}
            </p>
          </div>
          <div className="flex-1 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 relative z-10">
            {Object.entries(score.breakdown).map(([cat, val]) => (
              <div key={cat} className="bg-panel-2 border border-border rounded-2xl p-5 hover:border-accent/40 transition-colors group/card">
                <p className="text-[10px] font-black text-text-faint uppercase tracking-[0.2em] mb-2">{cat}</p>
                <div className="flex items-baseline gap-1">
                  <p className="text-2xl font-black text-text tabular-nums">{val}</p>
                  <span className="text-[10px] font-bold text-text-faint">/ 100</span>
                </div>
                <div className="h-1 w-full bg-panel-3 rounded-full mt-3 overflow-hidden border border-border/50">
                  <div
                    className={cn("h-full transition-all duration-1000", (val >= 80 ? 'bg-success' : val >= 60 ? 'bg-warning' : 'bg-danger'))}
                    style={{ width: `${val}%` }}
                  />
                </div>
              </div>
            ))}
          </div>
        </Panel>
      )}

      {/* Quick Stats */}
      <div className="grid grid-cols-4 gap-4">
        <MiniStat label="Security Score" value={score?.score ?? 'N/A'} icon={<Shield size={24} />} variant={score && score.score >= 75 ? 'success' : score && score.score >= 50 ? 'warning' : 'danger'} />
        <MiniStat label="Active Admins" value={users.filter(u => u.is_admin && u.is_enabled).length} icon={<Users size={24} />} variant={users.filter(u => u.is_admin && u.is_enabled).length <= 2 ? 'success' : 'warning'} />
        <MiniStat label="High-Risk Ports" value={ports.filter(p => p.risk_level === 'high').length} icon={<Radio size={24} />} variant={ports.filter(p => p.risk_level === 'high').length === 0 ? 'success' : 'danger'} />
        <MiniStat label="Security Events" value={events.length} icon={<AlertTriangle size={24} />} variant={events.filter(e => e.level === 'Error').length === 0 ? 'success' : 'danger'} />
      </div>

      {/* Defender & Firewall Status */}
      <div className="grid grid-cols-2 gap-8">
        {defender && (
          <Panel padding="lg" category="security" className="group">
            <h3 className="text-lg font-black text-text uppercase tracking-[0.2em] mb-8 flex items-center gap-4">
              <div className="p-2 rounded-xl bg-accent/10 border border-accent/20">
                <ShieldCheck size={24} className="text-accent" />
              </div>
              Endpoint Protection
            </h3>
            <div className="space-y-5">
              {[
                { label: 'DEFENDER SERVICE', active: defender.enabled },
                { label: 'REAL-TIME ANALYSIS', active: defender.real_time_protection },
                { label: 'NEURAL CLOUD SYNC', active: defender.cloud_protection },
                { label: 'THREAT SIGNATURES', active: defender.up_to_date },
              ].map(m => (
                <div key={m.label} className="flex items-center justify-between group/row">
                  <span className="text-[11px] font-black text-text-dim uppercase tracking-wider group-hover/row:text-text transition-colors">{m.label}</span>
                  <StatusBadge status={m.active ? 'enabled' : 'disabled'} />
                </div>
              ))}
            </div>
          </Panel>
        )}

        {fwStatus && (
          <Panel padding="lg" category="security" className="group">
            <h3 className="text-lg font-black text-text uppercase tracking-[0.2em] mb-8 flex items-center gap-4">
              <div className="p-2 rounded-xl bg-warning/10 border border-warning/20">
                <ShieldAlert size={24} className="text-warning" />
              </div>
              Perimeter Defense
            </h3>
            <div className="space-y-5">
              <div className="flex items-center justify-between group/row">
                <span className="text-[11px] font-black text-text-dim uppercase tracking-wider group-hover/row:text-text transition-colors">GLOBAL FIREWALL</span>
                <StatusBadge status={fwStatus.enabled ? 'enabled' : 'disabled'} />
              </div>
              {fwStatus.profiles.map(p => (
                <div key={p.name} className="flex items-center justify-between group/row">
                  <span className="text-[11px] font-black text-text-dim uppercase tracking-wider group-hover/row:text-text transition-colors">{p.name.toUpperCase()} PROFILE</span>
                  <StatusBadge status={p.enabled ? 'enabled' : 'disabled'} />
                </div>
              ))}
            </div>
          </Panel>
        )}
      </div>

      {/* Recommendations */}
      {score && score.recommendations.length > 0 && (
        <Panel padding="lg" category="security">
          <h3 className="text-lg font-black text-text uppercase tracking-[0.2em] mb-8 flex items-center gap-4">
            <div className="p-2 rounded-xl bg-warning/10 border border-warning/20">
              <Zap size={24} className="text-warning" />
            </div>
            Tactical Recommendations
          </h3>
          <div className="grid grid-cols-2 gap-4">
            {score.recommendations.map((rec) => (
              <div key={rec} className="flex items-start gap-4 bg-panel-2 border border-border rounded-2xl px-6 py-5 hover:border-warning/40 transition-all group">
                <AlertTriangle size={18} className="text-warning mt-0.5 shrink-0 transition-transform group-hover:scale-110" />
                <span className="text-sm font-bold text-text-dim leading-relaxed group-hover:text-text transition-colors">{rec}</span>
              </div>
            ))}
          </div>
        </Panel>
      )}

      {/* ── Exposed Services ── */}
      {ports.length > 0 && (
        <Panel padding="lg" category="security">
          <h3 className="text-lg font-black text-text uppercase tracking-[0.2em] mb-8 flex items-center gap-4">
            <div className="p-2 rounded-xl bg-accent/10 border border-accent/20">
              <Radio size={24} className="text-accent" />
            </div>
            Exposed Services
            <span className="ml-auto text-xs font-bold text-text-faint tabular-nums">{ports.length} ports</span>
          </h3>
          <div className="overflow-x-auto">
            <table className="w-full text-left">
              <thead>
                <tr className="border-b border-border/50 text-[10px] font-black text-text-faint uppercase tracking-wider">
                  <th className="pb-3 pr-4">Port</th>
                  <th className="pb-3 pr-4">Protocol</th>
                  <th className="pb-3 pr-4">Process</th>
                  <th className="pb-3 pr-4">PID</th>
                  <th className="pb-3 pr-4">Service</th>
                  <th className="pb-3 pr-4">External</th>
                  <th className="pb-3 text-right">Risk</th>
                </tr>
              </thead>
              <tbody>
                {ports.slice(0, 12).map((p) => (
                  <tr key={`${p.port}-${p.protocol}`} className="border-b border-border/30 hover:bg-panel-2 transition-colors">
                    <td className="py-3 pr-4 text-sm font-bold text-text tabular-nums">{p.port}</td>
                    <td className="py-3 pr-4 text-xs font-semibold text-text-dim">{p.protocol}</td>
                    <td className="py-3 pr-4 text-xs font-semibold text-text-dim max-w-[160px] truncate">{p.process_name}</td>
                    <td className="py-3 pr-4 text-xs font-semibold text-text-dim tabular-nums">{p.pid}</td>
                    <td className="py-3 pr-4 text-xs font-semibold text-text-dim">{p.service_name || '—'}</td>
                    <td className="py-3 pr-4">
                      <span className={cn(
                        "text-[10px] font-black uppercase tracking-wider px-2 py-0.5 rounded",
                        p.is_external ? "bg-danger/10 text-danger border border-danger/20" : "bg-success/10 text-success border border-success/20"
                      )}>
                        {p.is_external ? 'Yes' : 'No'}
                      </span>
                    </td>
                    <td className="py-3 text-right">
                      <span className={cn(
                        "text-[10px] font-black uppercase tracking-wider px-2 py-0.5 rounded",
                        p.risk_level === 'high' ? "bg-danger/10 text-danger border border-danger/20" :
                        p.risk_level === 'medium' ? "bg-warning/10 text-warning border border-warning/20" :
                        "bg-success/10 text-success border border-success/20"
                      )}>
                        {p.risk_level}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {ports.length > 12 && (
              <p className="text-[10px] text-text-faint font-bold uppercase tracking-wider text-center mt-4">
                +{ports.length - 12} more ports
              </p>
            )}
          </div>
        </Panel>
      )}

      {/* ── User Accounts Breakdown ── */}
      {users.length > 0 && (
        <Panel padding="lg" category="security">
          <h3 className="text-lg font-black text-text uppercase tracking-[0.2em] mb-8 flex items-center gap-4">
            <div className="p-2 rounded-xl bg-accent/10 border border-accent/20">
              <Users size={24} className="text-accent" />
            </div>
            User Accounts
            <span className="ml-auto text-xs font-bold text-text-faint tabular-nums">{users.length} total</span>
          </h3>
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
            <div className="bg-panel-2 border border-border rounded-2xl p-5 text-center hover:border-accent/30 transition-all">
              <p className="text-3xl font-black text-text tabular-nums">{users.filter(u => u.is_admin && u.is_enabled).length}</p>
              <p className="text-[10px] font-black text-text-faint uppercase tracking-wider mt-1">Active Admins</p>
            </div>
            <div className="bg-panel-2 border border-border rounded-2xl p-5 text-center hover:border-accent/30 transition-all">
              <p className="text-3xl font-black text-text tabular-nums">{users.filter(u => !u.is_enabled).length}</p>
              <p className="text-[10px] font-black text-text-faint uppercase tracking-wider mt-1">Disabled</p>
            </div>
            <div className="bg-panel-2 border border-border rounded-2xl p-5 text-center hover:border-accent/30 transition-all">
              <p className="text-3xl font-black text-text tabular-nums">{users.filter(u => u.password_never_expires).length}</p>
              <p className="text-[10px] font-black text-text-faint uppercase tracking-wider mt-1">Password Never Expires</p>
            </div>
            <div className="bg-panel-2 border border-border rounded-2xl p-5 text-center hover:border-accent/30 transition-all">
              <p className="text-3xl font-black text-text tabular-nums">{users.filter(u => !u.is_enabled || !u.last_logon).length}</p>
              <p className="text-[10px] font-black text-text-faint uppercase tracking-wider mt-1">Inactive</p>
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-left">
              <thead>
                <tr className="border-b border-border/50 text-[10px] font-black text-text-faint uppercase tracking-wider">
                  <th className="pb-3 pr-4">Username</th>
                  <th className="pb-3 pr-4">Full Name</th>
                  <th className="pb-3 pr-4">Group</th>
                  <th className="pb-3 pr-4">Admin</th>
                  <th className="pb-3 pr-4">Enabled</th>
                  <th className="pb-3 pr-4">Pwd Expires</th>
                  <th className="pb-3 text-right">Last Logon</th>
                </tr>
              </thead>
              <tbody>
                {users.slice(0, 10).map((u) => (
                  <tr key={u.username} className="border-b border-border/30 hover:bg-panel-2 transition-colors">
                    <td className="py-3 pr-4 text-sm font-bold text-text">{u.username}</td>
                    <td className="py-3 pr-4 text-xs font-semibold text-text-dim">{u.full_name || '—'}</td>
                    <td className="py-3 pr-4 text-xs font-semibold text-text-dim">{u.group || '—'}</td>
                    <td className="py-3 pr-4">
                      <StatusBadge status={u.is_admin ? 'enabled' : 'disabled'} />
                    </td>
                    <td className="py-3 pr-4">
                      <StatusBadge status={u.is_enabled ? 'enabled' : 'disabled'} />
                    </td>
                    <td className="py-3 pr-4">
                      <StatusBadge status={u.password_never_expires ? 'disabled' : 'enabled'} />
                    </td>
                    <td className="py-3 text-right text-xs font-semibold text-text-dim tabular-nums">{u.last_logon || '—'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
            {users.length > 10 && (
              <p className="text-[10px] text-text-faint font-bold uppercase tracking-wider text-center mt-4">
                +{users.length - 10} more users
              </p>
            )}
          </div>
        </Panel>
      )}

      {/* ── Recent Security Events ── */}
      {events.length > 0 && (
        <Panel padding="lg" category="security">
          <h3 className="text-lg font-black text-text uppercase tracking-[0.2em] mb-8 flex items-center gap-4">
            <div className="p-2 rounded-xl bg-danger/10 border border-danger/20">
              <AlertTriangle size={24} className="text-danger" />
            </div>
            Recent Security Events
            <span className="ml-auto text-xs font-bold text-text-faint tabular-nums">{events.length} events</span>
          </h3>
          <div className="space-y-2 max-h-[400px] overflow-y-auto pr-1">
            {events.slice(0, 20).map((e) => (
              <div key={e.id} className="flex items-start gap-4 p-4 rounded-xl bg-panel-2 border border-border/50 hover:border-border transition-all">
                <div className={cn(
                  "mt-0.5 w-2 h-2 rounded-full shrink-0",
                  e.level === 'Error' ? 'bg-danger shadow-[0_0_6px_var(--color-danger)]' :
                  e.level === 'Warning' ? 'bg-warning shadow-[0_0_6px_var(--color-warning)]' :
                  'bg-accent'
                )} />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-3 mb-1">
                    <span className={cn(
                      "text-[10px] font-black uppercase tracking-wider px-1.5 py-0.5 rounded",
                      e.level === 'Error' ? 'bg-danger/10 text-danger' :
                      e.level === 'Warning' ? 'bg-warning/10 text-warning' :
                      'bg-accent/10 text-accent'
                    )}>
                      {e.level}
                    </span>
                    <span className="text-[10px] font-bold text-text-faint tabular-nums">{e.time}</span>
                    <span className="text-[10px] font-bold text-text-faint">{e.provider}</span>
                  </div>
                  <p className="text-xs font-semibold text-text-dim leading-relaxed">{e.message}</p>
                </div>
              </div>
            ))}
          </div>
          {events.length > 20 && (
            <p className="text-[10px] text-text-faint font-bold uppercase tracking-wider text-center mt-4">
              +{events.length - 20} more events
            </p>
          )}
        </Panel>
      )}
    </div>
  )
}
