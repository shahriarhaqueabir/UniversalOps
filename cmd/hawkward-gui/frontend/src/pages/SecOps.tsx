import { useState, useEffect, useCallback } from 'react'
import { cn } from '@/lib/utils'
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
  ShieldAlert,
  History,
  Zap,
} from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { ConfirmDialog } from '@/components/dialogs/ConfirmDialog'
import type {
  FirewallRule,
  UserInfo,
  DefenderStatus,
  SecurityEvent,
  ListeningPort,
} from '@/types'

// eslint-disable-next-line @typescript-eslint/no-explicit-any -- Wails bridge type
type BackendCall = (method: string, ...args: any[]) => Promise<any>

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
    <span className={cn('inline-block px-3 py-1 text-xs font-black uppercase tracking-widest rounded-full border shadow-sm', colorMap[s] || 'bg-text-faint/20 text-text-faint border-border')}>
      {status}
    </span>
  )
}

// ── Firewall Tab ──

function FirewallTab({ call }: { call: BackendCall }) {
  const [rules, setRules] = useState<FirewallRule[]>([])
  const [search, setSearch] = useState('')
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [pendingAction, setPendingAction] = useState<{ index: number, name: string } | null>(null)

  const fetchRules = useCallback(async () => {
    try {
      const data = await call('SecOps.GetFirewallRules')
      setRules((data as FirewallRule[]) || [])
    } catch (err) {
      console.error(err)
      setRules([])
    }
  }, [call])

  useEffect(() => { fetchRules() }, [fetchRules])

  const toggleRule = async (index: number) => {
    const rule = rules[index]
    const success = await call('SecOps.SetFirewallRuleState', rule.name, !rule.enabled)
    if (success) {
      setRules(prev => prev.map((r, i) => i === index ? { ...r, enabled: !r.enabled } : r))
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

      <div className="flex items-center gap-6 bg-panel-2 border border-border p-6 rounded-2xl shadow-inner">
        <div className="relative group w-96">
          <Search size={24} className="absolute left-4 top-1/2 -translate-y-1/2 text-text-faint group-focus-within:text-accent transition-colors" />
          <input
            type="text"
            placeholder="Search firewall policy..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full bg-panel border border-border rounded-xl pl-14 pr-4 py-4 text-xl text-text placeholder-text-faint focus:outline-none focus:border-accent shadow-inner"
          />
        </div>
        <div className="flex-1" />
        <div className="flex items-center gap-6">
          <div className="flex items-center gap-4 bg-panel border border-border px-6 py-4 rounded-xl shadow-lg">
            <ShieldCheck size={20} className="text-success" />
            <span className="text-lg font-black text-text uppercase tabular-nums">{rules.filter(r => r.enabled).length} ACTIVE RULES</span>
          </div>
          {rules.some(r => r.is_high_risk) && (
            <div className="flex items-center gap-4 bg-danger/10 border border-danger/30 px-6 py-4 rounded-xl shadow-lg animate-pulse">
              <AlertTriangle size={20} className="text-danger" />
              <span className="text-lg font-black text-danger uppercase tabular-nums">
                {rules.filter(r => r.is_high_risk).length} HIGH RISK
              </span>
            </div>
          )}
        </div>
      </div>

      <div className="bg-panel border border-border rounded-[24px] overflow-hidden shadow-2xl">
        <div className="overflow-x-auto h-full max-h-[600px] overflow-y-auto">
          <table className="w-full text-left border-collapse">
            <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border shadow-sm">
              <tr>
                <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest">Policy Name</th>
                <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest">Action</th>
                <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest">Protocol</th>
                <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest">Local Port</th>
                <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest text-center">State</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((rule, i) => (
                <tr key={i} className="border-b border-border/20 hover:bg-white/5 transition-all">
                  <td className="px-8 py-5 text-xl font-black text-text truncate max-w-md">
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
                    <button
                      onClick={() => {
                        setPendingAction({ index: i, name: rule.name })
                        setConfirmOpen(true)
                      }}
                      className={cn(
                        'relative inline-flex h-8 w-14 items-center rounded-full transition-all shadow-inner border border-border',
                        rule.enabled ? 'bg-success/30 border-success/50' : 'bg-panel-3',
                      )}
                    >
                      <span className={cn('h-5 w-5 rounded-full bg-white shadow-lg transition-transform', rule.enabled ? 'translate-x-7' : 'translate-x-1')} />
                    </button>
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
  const [users, setUsers] = useState<UserInfo[]>([])

  useEffect(() => {
    call('SecOps.GetUsers').then(res => setUsers((res as UserInfo[]) || []))
  }, [call])

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <ExpertInsight
        title="Identity & Access"
        content="Local user accounts with 'IsAdmin' status have full control over the workstation. Audit these regularly to ensure the Principle of Least Privilege (PoLP) is maintained."
      />

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {users.map(user => (
          <div key={user.username} className="bg-panel border border-border rounded-[24px] p-8 shadow-xl flex items-center gap-8">
            <div className={cn("w-20 h-20 rounded-2xl flex items-center justify-center shrink-0 border shadow-inner", user.is_admin ? "bg-warning/10 border-warning/30 text-warning" : "bg-accent/10 border-accent/30 text-accent")}>
              {user.is_admin ? <ShieldAlert size={40} /> : <UserCheck size={40} />}
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-4 mb-1">
                <h3 className="text-2xl font-black text-text truncate">{user.username}</h3>
                {user.is_admin && <span className="px-3 py-1 rounded-full bg-warning text-black text-xs font-black uppercase tracking-tighter">Admin</span>}
              </div>
              <p className="text-text-dim text-lg mb-4">{user.full_name || 'System Account'}</p>
              <div className="flex items-center gap-4">
                <div className="px-4 py-1.5 rounded-full bg-panel-3 border border-border text-sm font-bold text-text-faint tabular-nums truncate">{user.sid}</div>
                <div className={cn("w-3 h-3 rounded-full shadow-lg", user.is_enabled ? "bg-success" : "bg-danger")} />
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
  const [status, setStatus] = useState<DefenderStatus | null>(null)

  useEffect(() => {
    call('SecOps.GetDefenderStatus').then(res => setStatus(res as DefenderStatus))
  }, [call])

  if (!status) return null

  const metrics = [
    { label: 'Antivirus Engine', active: status.enabled, icon: <Shield size={24} /> },
    { label: 'Real-time Protection', active: status.real_time_protection, icon: <Activity size={24} /> },
    { label: 'Cloud-delivered Protection', active: status.cloud_protection, icon: <Zap size={24} /> },
    { label: 'Signature Baseline', active: status.up_to_date, icon: <ShieldCheck size={24} /> },
  ]

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <ExpertInsight
        title="Endpoint Hardening"
        content="Windows Defender provides multiple layers of protection. Ensure 'Real-time' and 'Cloud' protection are both active to detect modern zero-day threats through heuristic analysis."
      />

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {metrics.map(m => (
          <div key={m.label} className={cn("bg-panel border rounded-[24px] p-8 shadow-xl transition-all border-l-[8px]", m.active ? "border-l-success border-border" : "border-l-danger border-danger/30")}>
            <div className={cn("w-12 h-12 rounded-xl flex items-center justify-center mb-6 border shadow-inner", m.active ? "bg-success/10 border-success/30 text-success" : "bg-danger/10 border-danger/30 text-danger")}>
              {m.icon}
            </div>
            <h4 className="text-lg font-black text-text mb-2">{m.label}</h4>
            <span className={cn("text-xs font-bold uppercase tracking-widest px-3 py-1 rounded-full", m.active ? "bg-success text-white" : "bg-danger text-white")}>
              {m.active ? 'Operational' : 'Disabled'}
            </span>
          </div>
        ))}
      </div>

      <div className="bg-panel border border-border rounded-[24px] p-10 shadow-2xl">
        <h3 className="text-xl font-black text-text uppercase tracking-widest mb-8 flex items-center gap-4">
          <History size={24} className="text-accent" /> Scan & Integrity Timeline
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-10">
          <div>
            <p className="text-sm font-bold text-text-faint uppercase mb-2">Last Threat Scan</p>
            <p className="text-2xl font-black text-text">{status.last_scan}</p>
          </div>
          <div>
            <p className="text-sm font-bold text-text-faint uppercase mb-2">Signature Age</p>
            <p className="text-2xl font-black text-success tabular-nums">{status.signature_age}</p>
          </div>
          <div>
            <p className="text-sm font-bold text-text-faint uppercase mb-2">Full Scan Recency</p>
            <p className="text-2xl font-black text-text tabular-nums">{status.full_scan_age} Days</p>
          </div>
        </div>
      </div>
    </div>
  )
}

// ── Listening Tab ──

function ListeningTab({ call }: { call: BackendCall }) {
  const [ports, setPorts] = useState<ListeningPort[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    call('SecOps.GetListeningPorts')
      .then(res => {
        setPorts((res as ListeningPort[]) || [])
        setLoading(false)
      })
      .catch((err: unknown) => {
        console.error(err)
        setError('Failed to retrieve listening ports.')
        setLoading(false)
      })
  }, [call])

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <ExpertInsight
        title="Network Surveillance"
        content="Every open port is a potential vector for unauthorized access. Processes listed here are actively listening for inbound connections. Scrutinize unfamiliar process names, unexpected protocols, or ports outside the well-known range (0-1023)."
      />

      {loading && (
        <div className="flex flex-col items-center justify-center py-32 opacity-40">
          <Radio size={64} className="animate-pulse mb-6 text-accent" />
          <p className="text-2xl font-black uppercase tracking-[0.15em]">Scanning ports...</p>
        </div>
      )}

      {error && !loading && (
        <div className="bg-panel border border-danger/30 rounded-[24px] p-10 shadow-2xl">
          <div className="flex items-center gap-6">
            <div className="w-16 h-16 rounded-2xl bg-danger/20 flex items-center justify-center shrink-0">
              <ShieldAlert size={36} className="text-danger" />
            </div>
            <div>
              <h3 className="text-2xl font-black text-danger mb-2">Scan Failed</h3>
              <p className="text-xl text-text-dim">{error}</p>
            </div>
          </div>
        </div>
      )}

      {!loading && !error && ports.length === 0 && (
        <div className="flex flex-col items-center justify-center py-32 opacity-40">
          <ShieldCheck size={64} className="mb-6 text-success" />
          <p className="text-2xl font-black uppercase tracking-[0.15em]">No listening ports found</p>
        </div>
      )}

      {!loading && !error && ports.length > 0 && (
        <div className="bg-panel border border-border rounded-[24px] overflow-hidden shadow-2xl">
          <div className="overflow-x-auto max-h-[600px] overflow-y-auto">
            <table className="w-full text-left border-collapse">
              <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border shadow-sm">
                <tr>
                  <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest">Port</th>
                  <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest">Protocol</th>
                  <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest">Process</th>
                  <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest">PID</th>
                  <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest">State</th>
                </tr>
              </thead>
              <tbody>
                {ports.map((p, i) => (
                  <tr key={i} className={cn("border-b border-border/20 hover:bg-white/5 transition-all", p.is_external ? "bg-warning/5" : "")}>
                    <td className="px-8 py-5 text-xl font-black text-accent tabular-nums">
                      <div className="flex items-center gap-3">
                        {p.port}
                        {p.is_external && (
                          <span className="px-2 py-0.5 rounded-md bg-warning/20 text-warning text-[10px] font-black uppercase tracking-widest border border-warning/30">External</span>
                        )}
                      </div>
                    </td>
                    <td className="px-8 py-5 font-bold text-lg text-text-faint tabular-nums uppercase">{p.protocol}</td>
                    <td className="px-8 py-5 text-xl font-bold text-text truncate max-w-md">{p.process_name}</td>
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
  const [events, setEvents] = useState<SecurityEvent[]>([])

  useEffect(() => {
    call('SecOps.GetSecurityEvents').then(res => setEvents((res as SecurityEvent[]) || []))
  }, [call])

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <div className="bg-panel border border-border rounded-[24px] overflow-hidden shadow-2xl">
        <div className="max-h-[700px] overflow-y-auto">
          <table className="w-full text-left border-collapse">
            <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border shadow-sm">
              <tr>
                <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest">ID</th>
                <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest">Level</th>
                <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest">Origin</th>
                <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest">Message</th>
              </tr>
            </thead>
            <tbody>
              {events.map((e, i) => (
                <tr key={i} className={cn("border-b border-border/20 hover:bg-white/5 transition-all group", e.level === 'Error' ? "bg-danger/5" : "")}>
                  <td className="px-8 py-5 font-bold text-lg text-text-faint tabular-nums">{e.id}</td>
                  <td className="px-8 py-5">
                    <span className={cn("px-4 py-1 rounded-full text-xs font-black uppercase tracking-tighter border", e.level === 'Error' ? "bg-danger text-white border-danger/30 shadow-[0_0_12px_rgba(251,93,107,0.4)]" : "bg-panel-3 text-text-dim border-border")}>
                      {e.level}
                    </span>
                  </td>
                  <td className="px-8 py-5 text-xl font-bold text-text truncate max-w-[200px]">{e.provider}</td>
                  <td className="px-8 py-5 text-lg text-text-dim leading-relaxed">{e.message}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}

// ── Main Page ──

export function SecOps() {
  const { call } = useBackend()
  const [activeTab, setActiveTab] = useState<SecOpsTab>('firewall')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // Brief initial render before sub-tabs load their data
    const t = setTimeout(() => setLoading(false), 50)
    return () => clearTimeout(t)
  }, [])

  if (loading) {
    return (
      <div className="p-6 space-y-4 animate-pulse">
        <div className="h-8 w-48 bg-panel-2 rounded" />
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="h-32 bg-panel-2 rounded" />
          <div className="h-32 bg-panel-2 rounded" />
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full bg-[var(--color-bg)]">
      <div className="p-8 border-b border-border bg-panel-2 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-text flex items-center gap-4">
            <Shield size={32} className="text-danger" /> Security Operations
          </h1>
          <p className="text-text-dim text-lg mt-2">Threat surface analysis, access control, and endpoint protection status.</p>
        </div>
        <div className="flex gap-1 bg-panel border border-border rounded-2xl p-1.5 shadow-inner">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={cn(
                'flex items-center gap-3 px-6 py-3 rounded-xl text-lg font-bold transition-all',
                activeTab === tab.id ? 'bg-danger text-white shadow-lg' : 'text-text-dim hover:text-text hover:bg-white/5',
              )}
            >
              {tab.icon}
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-10">
        {activeTab === 'firewall' && <FirewallTab call={call} />}
        {activeTab === 'users' && <UsersTab call={call} />}
        {activeTab === 'defender' && <DefenderTab call={call} />}
        {activeTab === 'events' && <EventsTab call={call} />}
        {activeTab === 'listening' && <ListeningTab call={call} />}
      </div>
    </div>
  )
}
