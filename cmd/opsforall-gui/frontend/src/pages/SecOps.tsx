import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import * as Switch from '@radix-ui/react-switch'
import { cn } from '@/lib/utils'
import { EmptyState } from '@/components/ui/EmptyState'
import {
  Shield,
  Search,
  Users,
  Radio,
  Activity,
  AlertTriangle,
  Lightbulb,
  ShieldCheck,
  UserCheck,
  CheckCircle2,
  Clock,
  ShieldAlert,
  History,
  Zap,
  FileText,
} from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { ConfirmDialog } from '@/components/dialogs/ConfirmDialog'
import { DataFreshnessIndicator } from '@/components/ui/DataFreshnessIndicator'
import type {
  FirewallRule,
  UserInfo,
  DefenderStatus,
  SecurityEvent,
  ListeningPort,
  SecurityScore,
  FirewallStatus,
  RiskInfo,
  SecuritySummary,
  ScheduledTask,
} from '@/types'

type BackendCall = (method: string, ...args: unknown[]) => Promise<unknown>

// ── Tab definitions ──

type SecOpsTab = 'firewall' | 'users' | 'listening' | 'defender' | 'events'

const tabs: { id: SecOpsTab; label: string; icon: React.ReactNode }[] = [
  { id: 'firewall', label: 'Firewall', icon: <Shield size={20} /> },
  { id: 'users', label: 'Users', icon: <Users size={20} /> },
  { id: 'listening', label: 'Listening', icon: <Radio size={20} /> },
  { id: 'defender', label: 'Defender', icon: <Activity size={20} /> },
  { id: 'events', label: 'Events', icon: <AlertTriangle size={20} /> },
]

// ── Inline helpers ──

function ExpertInsight({ title, content }: { title: string, content: string }) {
  return (
    <div className="bg-accent-soft border border-accent/20 rounded-2xl p-8 flex gap-8 animate-in slide-in-from-left-4 duration-500">
      <div className="w-14 h-14 rounded-2xl bg-accent flex items-center justify-center shrink-0 shadow-xl">
        <Lightbulb size={32} className="text-white" />
      </div>
      <div>
        <h4 className="text-2xl font-bold text-text mb-2">{title}</h4>
        <p className="text-text-dim text-xl leading-relaxed">{content}</p>
      </div>
    </div>
  )
}

function SecurityStatusBadge({ status }: { status: string }) {
  const colorMap: Record<string, string> = {
    allow: 'bg-success/20 text-success border-success/30',
    block: 'bg-danger/20 text-danger border-danger/30',
    enabled: 'bg-success/20 text-success border-success/30',
    disabled: 'bg-text-faint/20 text-text-faint border-border',
    listening: 'bg-accent/15 text-accent border-accent/30',
    error: 'bg-danger/20 text-danger border-danger/30',
    warning: 'bg-warning/20 text-warning border-warning/30',
  }
  const s = status.toLowerCase()
  return (
    <span className={cn('inline-block px-3 py-1 text-xs font-bold uppercase tracking-widest rounded-full border shadow-sm', colorMap[s] || 'bg-text-faint/20 text-text-faint border-border')}>
      {status}
    </span>
  )
}

// ── Firewall Tab ──

function FirewallTab({ call }: { call: BackendCall }) {
  const { refreshInterval } = useSettingsStore()
  const [search, setSearch] = useState('')
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [pendingAction, setPendingAction] = useState<{ index: number, name: string } | null>(null)
  const queryClient = useQueryClient()

  const { data: rules = [] } = useQuery<FirewallRule[]>({
    queryKey: ['secops-firewall'],
    queryFn: async () => {
      const data = await call('SecOps.GetFirewallRules')
      return (data as FirewallRule[]) || []
    },
    refetchInterval: refreshInterval,
  })

  const { data: fwStatus } = useQuery<FirewallStatus | null>({
    queryKey: ['secops-firewall-status'],
    queryFn: async () => {
      const res = await call('SecOps.GetFirewallStatus')
      return (res as FirewallStatus) || null
    },
    refetchInterval: refreshInterval,
  })

  const toggleRule = async (index: number) => {
    const rule = rules[index]
    const success = await call('SecOps.SetFirewallRuleState', rule.name, !rule.enabled)
    if (success) {
      queryClient.invalidateQueries({ queryKey: ['secops-firewall'] })
    }
  }

  const filtered = rules.filter(r => r.name.toLowerCase().includes(search.toLowerCase()))

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <ConfirmDialog
        open={confirmOpen}
        title="Modify Network Perimeter"
        description={`You are about to change the state of firewall rule "${pendingAction?.name}". This affects how your workstation accepts or denies network traffic.`}
        type="warning"
        confirmText="Update Rule"
        onConfirm={() => toggleRule(pendingAction!.index)}
        onClose={() => setConfirmOpen(false)}
      />

      <ExpertInsight
        title="Perimeter Defense"
        content="The Host Firewall is your first line of defense. Rules marked 'Allow' from 'Any' IP are high-risk nodes. Always prefer 'Block' by default and only allow specific protocols on trusted interfaces."
      />

      {fwStatus && (
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
          <h3 className="text-xl font-bold text-text uppercase tracking-widest mb-6 flex items-center gap-4">
            <Shield size={24} className="text-accent" /> Firewall Status
          </h3>
          <div className="flex items-center gap-8 mb-6">
            <div className="flex items-center gap-4">
              <div className={cn("w-4 h-4 rounded-full shadow-lg", fwStatus.enabled ? "bg-success" : "bg-danger")} />
              <span className="text-2xl font-bold text-text">Global: {fwStatus.enabled ? 'ON' : 'OFF'}</span>
            </div>
          </div>
          {(fwStatus.profiles ?? []).length > 0 && (
            <div className="flex items-center gap-4">
              {(fwStatus.profiles ?? []).map(p => (
                <div key={p.name} className="flex items-center gap-2">
                  <span className={cn("px-3 py-1 text-xs font-bold uppercase tracking-widest rounded-full border shadow-sm",
                    p.enabled
                      ? 'bg-success/20 text-success border-success/30'
                      : 'bg-danger/20 text-danger border-danger/30'
                  )}>
                    {p.name}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      <div className="flex items-center gap-6 bg-panel-2 border border-border p-6 rounded-2xl shadow-inner">
        <div className="relative group w-96">
          <Search size={24} className="absolute left-4 top-1/2 -translate-y-1/2 text-text-faint group-focus-within:text-accent transition-colors" />
          <input
            type="text"
            placeholder="Search firewall policy..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full bg-panel border border-border rounded-xl pl-14 pr-4 py-2.5 text-sm text-[var(--color-text)] placeholder-[var(--color-text-faint)] focus:outline-none focus:border-accent shadow-inner"
          />
        </div>
        <div className="flex-1" />
        <div className="flex items-center gap-6">
          <div className="flex items-center gap-4 bg-panel border border-border px-6 py-4 rounded-xl shadow-lg">
            <ShieldCheck size={20} className="text-success" />
            <span className="text-sm font-semibold text-[var(--color-text)] uppercase tracking-wider tabular-nums">{rules.filter(r => r.enabled).length} ACTIVE RULES</span>
          </div>
          {rules.some(r => r.is_high_risk) && (
            <div className="flex items-center gap-4 bg-danger/10 border border-danger/30 px-6 py-4 rounded-xl shadow-lg animate-pulse">
              <AlertTriangle size={20} className="text-danger" />
              <span className="text-sm font-semibold text-[var(--color-danger)] uppercase tracking-wider tabular-nums">
                {rules.filter(r => r.is_high_risk).length} HIGH RISK
              </span>
            </div>
          )}
        </div>
      </div>

      <div className="bg-panel border border-border rounded-[var(--radius-lg)] overflow-hidden shadow-2xl">
        <div className="overflow-x-auto h-full max-h-[600px] overflow-y-auto">
          <table className="w-full text-left border-collapse">
            <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border shadow-sm">
              <tr>
                <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Policy Name</th>
                <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Action</th>
                <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Protocol</th>
                <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Local Port</th>
                <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider text-center">State</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((rule, i) => (
                <tr key={i} className="border-b border-border/20 hover:bg-[var(--color-sidebar-hover)] transition-all">
                  <td className="px-6 py-4 text-sm font-medium text-[var(--color-text)] truncate max-w-md">
                    <div className="flex items-center gap-3">
                      {rule.is_high_risk && (
                        <AlertTriangle size={20} className="text-danger animate-pulse shrink-0" />
                      )}
                      <span className={cn(rule.is_high_risk ? "text-danger" : "")}>{rule.name}</span>
                    </div>
                  </td>
                  <td className="px-8 py-5">
                    <SecurityStatusBadge status={rule.action} />
                  </td>
                  <td className="px-8 py-5 font-bold text-lg text-text-faint tabular-nums">{rule.protocol}</td>
                  <td className="px-8 py-5 font-bold text-lg text-accent tabular-nums">{rule.local_port}</td>
                  <td className="px-8 py-5 text-center">
                    <Switch.Root
                      checked={rule.enabled}
                      onCheckedChange={() => {
                        setPendingAction({ index: i, name: rule.name })
                        setConfirmOpen(true)
                      }}
                      className={cn(
                        'relative inline-flex h-6 w-11 items-center rounded-full transition-colors cursor-pointer',
                        rule.enabled ? 'bg-[var(--color-success)]' : 'bg-[var(--color-panel-3)]'
                      )}
                    >
                      <Switch.Thumb
                        className={cn(
                          'block h-4 w-4 rounded-full bg-white shadow-lg transition-transform will-change-transform',
                          rule.enabled ? 'translate-x-6' : 'translate-x-1'
                        )}
                      />
                    </Switch.Root>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}

// ── Users Tab ──

function UsersTab({ call }: { call: BackendCall }) {
  const { refreshInterval } = useSettingsStore()
  const { data: users = [] } = useQuery<UserInfo[]>({
    queryKey: ['secops-users'],
    queryFn: async () => {
      const res = await call('SecOps.GetUsers')
      return (res as UserInfo[]) || []
    },
    refetchInterval: refreshInterval,
  })

  const adminCount = users.filter(u => u.is_admin).length
  const disabledCount = users.filter(u => !u.is_enabled).length
  const pwdNeverExpiresCount = users.filter(u => u.password_never_expires).length

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <ExpertInsight
        title="Identity & Access"
        content="Local user accounts with 'IsAdmin' status have full control over the workstation. Audit these regularly to ensure the Principle of Least Privilege (PoLP) is maintained."
      />

      {/* ── Summary Stats ── */}
      {users.length > 0 && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {([
            { label: 'Admins', value: adminCount, cls: 'text-[var(--color-warning)]' },
            { label: 'Disabled', value: disabledCount, cls: 'text-[var(--color-danger)]' },
            { label: 'Password Never Expires', value: pwdNeverExpiresCount, cls: 'text-[var(--color-accent)]' },
            { label: 'Total Accounts', value: users.length, cls: 'text-[var(--color-success)]' },
          ] as const).map(s => (
            <div key={s.label} className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-5 shadow-lg">
              <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase tracking-wider mb-1">{s.label}</p>
              <p className={`text-3xl font-bold tabular-nums ${s.cls}`}>{s.value}</p>
            </div>
          ))}
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {users.length === 0 ? (
          <div className="lg:col-span-2">
            <EmptyState
              icon={<Users size={28} />}
              title="No Users Found"
              description="No local user accounts were detected on this system."
            />
          </div>
        ) : users.map(user => (
          <div key={user.username} className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-xl flex items-center gap-8">
            <div className={cn("w-20 h-20 rounded-2xl flex items-center justify-center shrink-0 border shadow-inner", user.is_admin ? "bg-warning/10 border-warning/30 text-warning" : "bg-accent/10 border-accent/30 text-accent")}>
              {user.is_admin ? <ShieldAlert size={40} /> : <UserCheck size={40} />}
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-4 mb-1">
                <h3 className="text-2xl font-bold text-text truncate">{user.username}</h3>
                {user.is_admin && <span className="px-3 py-1 rounded-full bg-warning text-black text-xs font-bold uppercase tracking-tighter">Admin</span>}
                {!user.is_enabled && <span className="px-3 py-1 rounded-full bg-danger text-white text-xs font-bold uppercase tracking-tighter">Disabled</span>}
              </div>
              <p className="text-text-dim text-lg mb-3">{user.full_name || 'System Account'}</p>
              <div className="flex flex-wrap items-center gap-3 text-xs">
                <div className="px-3 py-1 rounded-full bg-panel-3 border border-border font-bold text-text-faint tabular-nums truncate">{user.sid}</div>
                {user.last_logon && (
                  <div className="px-3 py-1 rounded-full bg-panel-3 border border-border font-bold text-text-faint">
                    Last logon: {user.last_logon}
                  </div>
                )}
                {user.password_never_expires && (
                  <div className="px-3 py-1 rounded-full bg-accent/10 border border-accent/30 font-bold text-accent">
                    Password Never Expires
                  </div>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

// ── Defender Tab ──

function DefenderTab({ call }: { call: BackendCall }) {
  const { refreshInterval } = useSettingsStore()
  const { data: status } = useQuery<DefenderStatus | null>({
    queryKey: ['secops-defender'],
    queryFn: async () => {
      const res = await call('SecOps.GetDefenderStatus')
      return (res as DefenderStatus) || null
    },
    refetchInterval: refreshInterval,
  })

  if (!status) return null

  const metrics = [
    { label: 'Antivirus Engine', active: status.enabled, icon: <Shield size={24} /> },
    { label: 'Real-time Protection', active: status.real_time_protection, icon: <Activity size={24} /> },
    { label: 'Cloud-delivered Protection', active: status.cloud_protection, icon: <Zap size={24} /> },
    { label: 'Signature Baseline', active: status.up_to_date, icon: <ShieldCheck size={24} /> },
    { label: 'Threats', active: status.threats_detected === 0, icon: <AlertTriangle size={24} />, count: status.threats_detected, countDangerous: status.threats_detected > 0 },
  ]

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <ExpertInsight
        title="Endpoint Hardening"
        content="Windows Defender provides multiple layers of protection. Ensure 'Real-time' and 'Cloud' protection are both active to detect modern zero-day threats through heuristic analysis."
      />

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-6">
        {metrics.map(m => (
          <div key={m.label} className={cn("bg-panel border rounded-[var(--radius-lg)] p-8 shadow-xl transition-all border-l-[8px]", m.active ? "border-l-success border-border" : "border-l-danger border-danger/30")}>
            <div className={cn("w-12 h-12 rounded-xl flex items-center justify-center mb-6 border shadow-inner", m.active ? "bg-success/10 border-success/30 text-success" : "bg-danger/10 border-danger/30 text-danger")}>
              {m.icon}
            </div>
            <h4 className="text-lg font-bold text-text mb-2">{m.label}</h4>
            {'countDangerous' in m && m.countDangerous ? (
              <span className="text-xs font-bold uppercase tracking-widest px-3 py-1 rounded-full bg-danger text-white">
                {m.count} detected
              </span>
            ) : (
              <span className={cn("text-xs font-bold uppercase tracking-widest px-3 py-1 rounded-full", m.active ? "bg-success text-white" : "bg-danger text-white")}>
                {m.active ? 'Operational' : 'Disabled'}
              </span>
            )}
          </div>
        ))}
      </div>

      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-10 shadow-2xl">
        <h3 className="text-xl font-bold text-text uppercase tracking-widest mb-8 flex items-center gap-4">
          <History size={24} className="text-accent" /> Scan & Integrity Timeline
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-10">
          <div>
            <p className="text-sm font-bold text-text-faint uppercase mb-2">Last Threat Scan</p>
            <p className="text-2xl font-bold text-text">{status.last_scan}</p>
          </div>
          <div>
            <p className="text-sm font-bold text-text-faint uppercase mb-2">Signature Age</p>
            <p className="text-2xl font-bold text-success tabular-nums">{status.signature_age}</p>
          </div>
          <div>
            <p className="text-sm font-bold text-text-faint uppercase mb-2">Full Scan Recency</p>
            <p className="text-2xl font-bold text-text tabular-nums">{status.full_scan_age} Days</p>
          </div>
        </div>
      </div>
    </div>
  )
}

// ── Listening Tab ──

function ListeningTab({ call }: { call: BackendCall }) {
  const { refreshInterval } = useSettingsStore()
  const { data: ports = [], isLoading: loading, error: fetchError } = useQuery<ListeningPort[]>({
    queryKey: ['secops-listening'],
    queryFn: async () => {
      const res = await call('SecOps.GetListeningPorts')
      return (res as ListeningPort[]) || []
    },
    refetchInterval: refreshInterval,
  })

  const error = fetchError ? 'Failed to retrieve listening ports.' : null
  const totalCount = ports.length
  const externalCount = ports.filter(p => p.is_external).length
  const unknownCount = ports.filter(p => p.process_name.startsWith('pid:')).length

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <ExpertInsight
        title="Network Surveillance"
        content="Every open port is a potential vector for unauthorized access. Processes listed here are actively listening for inbound connections. Scrutinize unfamiliar process names, unexpected protocols, or ports outside the well-known range (0-1023)."
      />

      {/* ── Summary Stats ── */}
      {ports.length > 0 && (
        <div className="grid grid-cols-3 gap-4">
          {([
            { label: 'Total Listening', value: totalCount, cls: 'text-[var(--color-accent)]' },
            { label: 'External Bindings', value: externalCount, cls: 'text-[var(--color-warning)]' },
            { label: 'Unknown Processes', value: unknownCount, cls: 'text-[var(--color-danger)]' },
          ] as const).map(s => (
            <div key={s.label} className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-5 shadow-lg">
              <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase tracking-wider mb-1">{s.label}</p>
              <p className={`text-3xl font-bold tabular-nums ${s.cls}`}>{s.value}</p>
            </div>
          ))}
        </div>
      )}

      {loading && (
        <div className="flex flex-col items-center justify-center py-20">
          <div className="w-16 h-16 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)] flex items-center justify-center text-[var(--color-accent)] mb-4">
            <Radio size={28} className="animate-pulse" />
          </div>
          <p className="text-sm font-semibold text-[var(--color-text-dim)]">Scanning ports...</p>
        </div>
      )}

      {error && !loading && (
        <div className="bg-[var(--color-panel)] border border-[var(--color-danger)]/30 rounded-[var(--radius-lg)] p-6 shadow-xl">
          <div className="flex items-center gap-4">
            <div className="w-10 h-10 rounded-lg bg-[var(--color-danger)]/10 flex items-center justify-center shrink-0">
              <ShieldAlert size={20} className="text-[var(--color-danger)]" />
            </div>
            <div>
              <h3 className="text-sm font-bold text-[var(--color-danger)] mb-0.5">Scan Failed</h3>
              <p className="text-sm text-[var(--color-text-dim)]">{error}</p>
            </div>
          </div>
        </div>
      )}

      {!loading && !error && ports.length === 0 && (
        <EmptyState
          icon={<ShieldCheck size={28} />}
          title="No Listening Ports"
          description="No open ports detected on this system. All services appear to be bound to internal interfaces."
        />
      )}

      {!loading && !error && ports.length > 0 && (
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] overflow-hidden shadow-2xl">
          <div className="overflow-x-auto max-h-[600px] overflow-y-auto">
            <table className="w-full text-left border-collapse">
              <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border shadow-sm">
                <tr>
                  <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Port</th>
                  <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Protocol</th>
                  <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Process</th>
                  <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">PID</th>
                  <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">State</th>
                </tr>
              </thead>
              <tbody>
                {ports.map((p, i) => (
                  <tr key={i} className={cn("border-b border-border/20 hover:bg-[var(--color-sidebar-hover)] transition-all", p.is_external ? "bg-warning/5" : "")}>
                    <td className="px-6 py-4 text-sm font-semibold text-[var(--color-accent)] tabular-nums">
                      <div className="flex items-center gap-3">
                        {p.port}
                        {p.is_external && (
                          <span className="px-2 py-0.5 rounded-md bg-warning/20 text-warning text-[10px] font-bold uppercase tracking-widest border border-warning/30">External</span>
                        )}
                      </div>
                    </td>
                    <td className="px-8 py-5 font-bold text-lg text-text-faint tabular-nums uppercase">{p.protocol}</td>
                    <td className="px-6 py-4 text-sm font-medium text-[var(--color-text)] truncate max-w-md">{p.process_name}</td>
                    <td className="px-8 py-5 font-bold text-lg text-text-faint tabular-nums">{p.pid}</td>
                    <td className="px-8 py-5">
                      <SecurityStatusBadge status={p.state} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}

// ── Events Tab ──

function EventsTab({ call }: { call: BackendCall }) {
  const { refreshInterval } = useSettingsStore()
  const [activeFilter, setActiveFilter] = useState<string>('all')
  const { data: events = [] } = useQuery<SecurityEvent[]>({
    queryKey: ['secops-events'],
    queryFn: async () => {
      const res = await call('SecOps.GetSecurityEvents')
      return (res as SecurityEvent[]) || []
    },
    refetchInterval: refreshInterval,
  })

  const { data: tasks = [] } = useQuery<ScheduledTask[]>({
    queryKey: ['secops-tasks'],
    queryFn: async () => {
      const res = await call('SecOps.GetScheduledTasks')
      return (res as ScheduledTask[]) || []
    },
    refetchInterval: refreshInterval * 2, // Tasks change less frequently
  })

  const categories = [
    { id: 'all', label: 'All', ids: null },
    { id: 'failed-logins', label: 'Failed Logins', ids: [4625] },
    { id: 'elevation', label: 'Elevation', ids: [4672, 4673] },
    { id: 'policy', label: 'Policy Changes', ids: [1102, 4719] },
    { id: 'account', label: 'Account Changes', ids: [4720, 4722, 4725, 4726] },
    { id: 'usb', label: 'USB', ids: [6416, 6417, 4663] },
  ]

  const filtered = activeFilter === 'all'
    ? events
    : events.filter(e => {
      const cat = categories.find(c => c.id === activeFilter)
      return cat?.ids?.includes(e.id)
    })

  const failedLoginsCount = events.filter(e => [4625].includes(e.id)).length
  const elevationCount = events.filter(e => [4672, 4673].includes(e.id)).length
  const policyCount = events.filter(e => [1102, 4719].includes(e.id)).length
  const usbCount = events.filter(e => [6416, 6417, 4663].includes(e.id)).length

  return (
    <div className="space-y-12 animate-in fade-in duration-500">
      <div className="space-y-8">
        <h3 className="text-xl font-bold text-text uppercase tracking-widest flex items-center gap-4">
          <History size={24} className="text-accent" /> Security Event Log
        </h3>

        {/* ── Summary Stats ── */}
        {events.length > 0 && (
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {([
              { label: 'Failed Logins', value: failedLoginsCount, cls: 'text-[var(--color-danger)]' },
              { label: 'Elevation', value: elevationCount, cls: 'text-[var(--color-warning)]' },
              { label: 'Policy Changes', value: policyCount, cls: 'text-[var(--color-accent)]' },
              { label: 'USB Devices', value: usbCount, cls: 'text-[var(--color-text-dim)]' },
            ] as const).map(s => (
              <div key={s.label} className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-5 shadow-lg">
                <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase tracking-wider mb-1">{s.label}</p>
                <p className={`text-3xl font-bold tabular-nums ${s.cls}`}>{s.value}</p>
              </div>
            ))}
          </div>
        )}

        <div className="flex items-center gap-3 flex-wrap">
          {categories.map(c => (
            <button
              key={c.id}
              onClick={() => setActiveFilter(c.id)}
              className={cn(
                'px-4 py-2 rounded-full text-sm font-bold uppercase tracking-wider border transition-all',
                activeFilter === c.id
                  ? 'bg-accent text-white border-accent shadow-lg'
                  : 'bg-panel border-border text-text-dim hover:text-text hover:bg-[var(--color-sidebar-hover)]'
              )}
            >
              {c.label}
              {c.ids && (
                <span className="ml-2 text-xs tabular-nums">
                  {events.filter(e => c.ids.includes(e.id)).length}
                </span>
              )}
            </button>
          ))}
        </div>

        <div className="bg-panel border border-border rounded-[var(--radius-lg)] overflow-hidden shadow-2xl">
          <div className="max-h-[500px] overflow-y-auto">
            {filtered.length === 0 ? (
              <EmptyState
                icon={<FileText size={28} />}
                title="No Security Events"
                description={activeFilter === 'all' ? "No security events have been recorded yet." : "No events match the selected filter."}
              />
            ) : (
              <table className="w-full text-left border-collapse">
                <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border shadow-sm">
                  <tr>
                    <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">ID</th>
                    <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Level</th>
                    <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Origin</th>
                    <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Message</th>
                  </tr>
                </thead>
                <tbody>
                  {filtered.map((e, i) => (
                    <tr key={i} className={cn("border-b border-border/20 hover:bg-[var(--color-sidebar-hover)] transition-all group", e.level === 'Error' ? "bg-danger/5" : "")}>
                      <td className="px-8 py-5 font-bold text-lg text-text-faint tabular-nums">{e.id}</td>
                      <td className="px-8 py-5">
                        <span className={cn("px-4 py-1 rounded-full text-xs font-bold uppercase tracking-tighter border", e.level === 'Error' ? "bg-danger text-white border-danger/30 shadow-[0_0_12px_rgba(251,93,107,0.4)]" : "bg-panel-3 text-text-dim border-border")}>
                          {e.level}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-sm font-medium text-[var(--color-text)] truncate max-w-[200px]">{e.provider}</td>
                      <td className="px-6 py-4 text-sm text-[var(--color-text-dim)] leading-relaxed">{e.message}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>
      </div>

      {/* ── Scheduled Tasks Section ── */}
      <div className="space-y-8">
        <h3 className="text-xl font-bold text-text uppercase tracking-widest flex items-center gap-4">
          <Clock size={24} className="text-warning" /> Scheduled Tasks
        </h3>

        <div className="bg-panel border border-border rounded-[var(--radius-lg)] overflow-hidden shadow-2xl">
          <div className="max-h-[500px] overflow-y-auto">
            {tasks.length === 0 ? (
              <div className="py-20">
                <EmptyState
                  icon={<Clock size={28} />}
                  title="No Scheduled Tasks"
                  description="No system-level scheduled tasks were detected."
                />
              </div>
            ) : (
              <table className="w-full text-left border-collapse">
                <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border shadow-sm">
                  <tr>
                    <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Task Name</th>
                    <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Status</th>
                    <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Last Run</th>
                    <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Next Run</th>
                  </tr>
                </thead>
                <tbody>
                  {tasks.map((t, i) => (
                    <tr key={i} className="border-b border-border/20 hover:bg-[var(--color-sidebar-hover)] transition-all">
                      <td className="px-8 py-5">
                        <p className="text-sm font-bold text-[var(--color-text)]">{t.name}</p>
                        <p className="text-[10px] text-text-faint font-medium uppercase tracking-tight mt-0.5">{t.author || 'System'}</p>
                      </td>
                      <td className="px-8 py-5">
                        <SecurityStatusBadge status={t.status} />
                      </td>
                      <td className="px-8 py-5 text-sm text-text-dim font-medium tabular-nums">{t.last_run || 'Never'}</td>
                      <td className="px-8 py-5 text-sm text-accent font-bold tabular-nums">{t.next_run || 'N/A'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

// ── Security Score Card ──

function securityGradeColor(grade: string) {
  switch (grade) {
    case 'A': return 'var(--color-success)'
    case 'B': return 'var(--color-accent)'
    case 'C': return 'var(--color-warning)'
    case 'D': return 'var(--color-danger)'
    default: return 'var(--color-danger)'
  }
}

const maxBreakdownValues: Record<string, number> = {
  Defender: 35,
  Firewall: 20,
  Users: 10,
  Ports: 10,
  Events: 10,
}

function SecurityScoreCard({ call }: { call: BackendCall }) {
  const { refreshInterval } = useSettingsStore()
  const { data, isLoading } = useQuery<SecurityScore>({
    queryKey: ['secops-security-score'],
    queryFn: async () => {
      const res = await call('SecOps.GetSecurityScore')
      return res as SecurityScore
    },
    refetchInterval: refreshInterval,
  })

  if (isLoading || !data) {
    return (
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 flex items-center justify-center">
        <div className="w-10 h-10 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)] flex items-center justify-center text-[var(--color-accent)]">
          <Shield size={20} className="animate-pulse" />
        </div>
        <p className="ml-4 text-sm font-semibold text-[var(--color-text-dim)]">Computing security score…</p>
      </div>
    )
  }

  const color = securityGradeColor(data.grade)
  const r = 48
  const circumference = 2 * Math.PI * r
  const dash = (data.score / 100) * circumference
  const gap = circumference - dash

  return (
    <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 flex flex-col lg:flex-row items-center gap-8 shadow-2xl relative overflow-hidden group">
      <div className="absolute top-0 right-0 w-48 h-48 rounded-bl-full pointer-events-none transition-all" style={{ backgroundColor: `${color}08` }} />

      {/* ── SVG Score Donut ── */}
      <div className="relative shrink-0">
        <svg width={140} height={140} viewBox="0 0 120 120" className="tabular-nums">
          <circle cx="60" cy="60" r={r} fill="none" stroke="var(--color-border)" strokeWidth="10" />
          <circle
            cx="60" cy="60" r={r}
            fill="none"
            stroke={color}
            strokeWidth="10"
            strokeLinecap="round"
            strokeDasharray={`${dash} ${gap}`}
            transform="rotate(-90 60 60)"
            style={{ transition: 'stroke-dasharray 1s ease' }}
          />
          <text x="60" y="54" textAnchor="middle" fill="var(--color-text)" fontSize="28" fontWeight="900" dominantBaseline="middle">
            {data.score}
          </text>
          <text x="60" y="78" textAnchor="middle" fill="var(--color-text-faint)" fontSize="11" fontWeight="bold" style={{ textTransform: 'uppercase' }} dominantBaseline="middle">
            SECURITY
          </text>
        </svg>
      </div>

      <div className="flex-1 min-w-0 w-full">
        <div className="flex items-center gap-4 mb-4">
          <span className="text-2xl font-bold text-[var(--color-text)]">Security Score</span>
          <span
            className="px-3 py-1 rounded-full border text-xs font-bold uppercase tracking-widest"
            style={{ borderColor: `${color}50`, color, backgroundColor: `${color}15` }}
          >
            Grade {data.grade}
          </span>
        </div>

        {/* Breakdown bars */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 mb-4">
          {Object.entries(data.breakdown).map(([category, value]) => {
            const maxVal = maxBreakdownValues[category] || 20
            const pct = Math.max(0, Math.min(100, (value / maxVal) * 100))
            return (
              <div key={category} className="flex flex-col gap-1">
                <div className="flex justify-between text-xs">
                  <span className="font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">{category}</span>
                  <span className="font-bold tabular-nums text-[var(--color-text)]">{value}/{maxVal}</span>
                </div>
                <div className="h-1.5 bg-panel-3 rounded-full overflow-hidden">
                  <div
                    className="h-full rounded-full transition-all duration-700"
                    style={{ width: `${pct}%`, backgroundColor: pct >= 70 ? 'var(--color-success)' : pct >= 40 ? 'var(--color-warning)' : 'var(--color-danger)' }}
                  />
                </div>
              </div>
            )
          })}
        </div>

        {/* Recommendations */}
        {(data.recommendations ?? []).length > 0 && (
          <div className="flex flex-col gap-2">
            {(data.recommendations ?? []).map((rec, i) => (
              <div key={i} className="flex items-start gap-3 text-sm">
                <Lightbulb size={14} className="text-warning mt-0.5 shrink-0" />
                <span className="text-[var(--color-text-dim)]">{rec}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

// ── AI Security Summary Panel ──

function SecuritySummaryPanel({ call }: { call: BackendCall }) {
  const { refreshInterval } = useSettingsStore()
  const { data, isLoading } = useQuery<SecuritySummary>({
    queryKey: ['secops-security-summary'],
    queryFn: async () => {
      const res = await call('SecOps.GetSecuritySummary')
      return res as SecuritySummary
    },
    refetchInterval: refreshInterval,
  })

  const scoreColor =
    (data?.score ?? 0) >= 75
      ? 'var(--color-success)'
      : (data?.score ?? 0) >= 50
        ? 'var(--color-warning)'
        : 'var(--color-danger)'

  return (
    <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
      <h3 className="text-xl font-bold text-text uppercase tracking-widest mb-6 flex items-center gap-4">
        <Lightbulb size={24} className="text-warning" /> AI Security Summary
      </h3>

      {isLoading || !data ? (
        <div className="flex items-center justify-center py-10">
          <div className="w-10 h-10 rounded-xl bg-panel-2 border border-border flex items-center justify-center text-accent">
            <Shield size={20} className="animate-pulse" />
          </div>
          <p className="ml-4 text-sm font-semibold text-text-dim">Analyzing security posture…</p>
        </div>
      ) : (
        <div className="space-y-6">
          {/* ── Summary header ── */}
          <div className="flex items-start gap-4">
            <div
              className="w-16 h-16 rounded-2xl flex items-center justify-center shrink-0 text-2xl font-black shadow-lg"
              style={{ backgroundColor: `${scoreColor}15`, color: scoreColor, border: `2px solid ${scoreColor}40` }}
            >
              {data.score}
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm text-text-dim leading-relaxed">{data.summary}</p>
              <div className="flex items-center gap-2 mt-2 text-[11px] text-text-faint">
                <Clock size={12} />
                <span>
                  Last analyzed{' '}
                  {new Date(data.analyzedAt).toLocaleString()}
                </span>
              </div>
            </div>
          </div>

          {/* ── Risks ── */}
          {(data.risks ?? []).length > 0 && (
            <div>
              <h4 className="text-xs font-bold uppercase tracking-widest text-text-dim mb-3">Risks Detected</h4>
              <div className="space-y-2">
                {(data.risks ?? []).map((risk, i) => (
                  <div key={i} className="flex items-start gap-3 bg-panel-2 border border-border rounded-xl px-4 py-3">
                    <AlertTriangle size={14} className="text-danger mt-0.5 shrink-0" />
                    <span className="text-sm text-text-dim leading-snug">{risk}</span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* ── Recommendations ── */}
          {(data.recommendations ?? []).length > 0 && (
            <div>
              <h4 className="text-xs font-bold uppercase tracking-widest text-text-dim mb-3">Recommendations</h4>
              <div className="space-y-2">
                {(data.recommendations ?? []).map((rec, i) => (
                  <div key={i} className="flex items-start gap-3 bg-panel-2 border border-border rounded-xl px-4 py-3">
                    <Lightbulb size={14} className="text-warning mt-0.5 shrink-0" />
                    <span className="text-sm text-text-dim leading-snug">{rec}</span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* ── Empty state ── */}
          {(data.risks ?? []).length === 0 && (data.recommendations ?? []).length === 0 && (
            <div className="flex items-center gap-4 py-4">
              <div className="w-12 h-12 rounded-2xl bg-success/10 border border-success/30 flex items-center justify-center">
                <CheckCircle2 size={24} className="text-success" />
              </div>
              <div>
                <h4 className="text-sm font-bold text-success">All Clear</h4>
                <p className="text-xs text-text-dim">No actionable risks or recommendations at this time.</p>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

// ── Risk Assessment Panel ──

function RiskAssessmentPanel({ call }: { call: BackendCall }) {
  const { refreshInterval } = useSettingsStore()
  const { data: risks = [], isLoading } = useQuery<RiskInfo[]>({
    queryKey: ['secops-risks'],
    queryFn: async () => {
      const res = await call('SecOps.GetRisks')
      return (res as RiskInfo[]) || []
    },
    refetchInterval: refreshInterval,
  })

  const severityColors: Record<string, string> = {
    critical: 'bg-danger/20 text-danger border-danger/30',
    high: 'bg-orange-500/20 text-orange-400 border-orange-500/30',
    medium: 'bg-warning/20 text-warning border-warning/30',
    low: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
  }

  const severityDots: Record<string, string> = {
    critical: 'bg-danger',
    high: 'bg-orange-400',
    medium: 'bg-warning',
    low: 'bg-blue-400',
  }

  return (
    <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
      <h3 className="text-xl font-bold text-text uppercase tracking-widest mb-6 flex items-center gap-4">
        <ShieldAlert size={24} className="text-danger" /> Risk Assessment
      </h3>

      {isLoading ? (
        <div className="flex items-center justify-center py-10">
          <div className="w-10 h-10 rounded-xl bg-panel-2 border border-border flex items-center justify-center text-accent">
            <Shield size={20} className="animate-pulse" />
          </div>
          <p className="ml-4 text-sm font-semibold text-text-dim">Analyzing risks…</p>
        </div>
      ) : risks.length === 0 ? (
        <div className="flex items-center gap-4 py-6">
          <div className="w-14 h-14 rounded-2xl bg-success/10 border border-success/30 flex items-center justify-center">
            <CheckCircle2 size={32} className="text-success" />
          </div>
          <div>
            <h4 className="text-lg font-bold text-success">No Risks Detected</h4>
            <p className="text-text-dim">Your system appears to be in a healthy security posture.</p>
          </div>
        </div>
      ) : (
        <div className="space-y-4">
          {risks.map((risk, i) => (
            <div key={i} className="bg-panel-2 border border-border rounded-xl p-6 flex gap-6">
              <div className="flex flex-col items-center gap-2 shrink-0">
                <div className={cn("w-3 h-3 rounded-full shadow-lg", severityDots[risk.severity] || 'bg-text-faint')} />
                <span className={cn('px-2 py-1 text-[10px] font-bold uppercase tracking-widest rounded-full border', severityColors[risk.severity] || 'bg-text-faint/20 text-text-faint border-border')}>
                  {risk.severity}
                </span>
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-3 mb-1">
                  <span className="px-2 py-0.5 rounded-md bg-panel-3 border border-border text-[10px] font-bold uppercase tracking-widest text-text-dim">{risk.category}</span>
                  <h4 className="text-lg font-bold text-text">{risk.title}</h4>
                </div>
                <p className="text-sm text-text-dim mb-2">{risk.description}</p>
                <div className="flex items-start gap-2">
                  <Lightbulb size={14} className="text-warning mt-0.5 shrink-0" />
                  <span className="text-sm text-text-faint">{risk.recommendation}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

// ── Main Page ──

export function SecOps() {
  const { call } = useBackend()
  const [activeTab, setActiveTab] = useState<SecOpsTab>('firewall')
  const [showOverview, setShowOverview] = useState(false)

  const { dataUpdatedAt: secUpdatedAt } = useQuery({
    queryKey: ['secops-health'],
    queryFn: async () => {
      const res = await call('SecOps.GetSecurityScore')
      return res
    },
    refetchInterval: 30000,
    staleTime: 15000,
  })

  return (
    <div className="flex flex-col h-full bg-[var(--color-bg)]">
      <div className="p-8 border-b border-border bg-panel-2 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-text flex items-center gap-4">
            <Shield size={32} className="text-danger" /> Security Operations
          </h1>
          <p className="text-text-dim text-lg mt-2">Threat surface analysis, access control, and endpoint protection status.</p>
          <div className="flex items-center gap-4 mt-2">
            <DataFreshnessIndicator lastUpdated={secUpdatedAt ? new Date(secUpdatedAt) : null} />
            <button
              onClick={() => setShowOverview(!showOverview)}
              className="text-xs font-bold uppercase tracking-wider text-accent hover:text-accent/80 transition-all border border-accent/20 px-2 py-0.5 rounded bg-accent/5 hover:bg-accent/10"
            >
              {showOverview ? 'Hide Threat Summary' : 'Show Threat Summary'}
            </button>
          </div>
        </div>
        <div className="flex gap-1 bg-panel border border-border rounded-2xl p-1.5 shadow-inner">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              role="tab"
              aria-selected={activeTab === tab.id}
              aria-label={`${tab.label} tab`}
              className={cn(
                'flex items-center gap-3 px-6 py-3 rounded-xl text-lg font-bold transition-all',
                activeTab === tab.id ? 'bg-danger text-white shadow-lg' : 'text-text-dim hover:text-text hover:bg-[var(--color-sidebar-hover)]',
              )}
            >
              {tab.icon}
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-6 space-y-6">
        {showOverview && (
          <div className="space-y-6 animate-in fade-in slide-in-from-top-2 duration-300">
            <SecurityScoreCard call={call} />
            <SecuritySummaryPanel call={call} />
            <RiskAssessmentPanel call={call} />
          </div>
        )}
        {activeTab === 'firewall' && <FirewallTab call={call} />}
        {activeTab === 'users' && <UsersTab call={call} />}
        {activeTab === 'defender' && <DefenderTab call={call} />}
        {activeTab === 'events' && <EventsTab call={call} />}
        {activeTab === 'listening' && <ListeningTab call={call} />}
      </div>
    </div>
  )
}
