import { useState, useEffect, useCallback } from 'react'
import { cn } from '@/lib/utils'
import {
  Shield,
  Search,
  Users,
  Radio,
  Activity,
  AlertTriangle,
  RefreshCw,
  Check,
  X,
  Copy,
  Globe,
  Wifi,
  Clock,
} from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import type {
  FirewallRule,
  UserInfo,
  ListeningPort,
  DefenderStatus,
  SecurityEvent,
} from '@/types'
import {
  mockFirewallRules,
  mockUsers,
  mockListeningPorts,
  mockDefenderStatus,
  mockSecurityEvents,
} from '@/lib/mockData'

// ── Tab definitions ──

type SecOpsTab = 'firewall' | 'users' | 'listening' | 'defender' | 'events'

const tabs: { id: SecOpsTab; label: string; icon: React.ReactNode }[] = [
  { id: 'firewall', label: 'Firewall', icon: <Shield size={16} /> },
  { id: 'users', label: 'Users', icon: <Users size={16} /> },
  { id: 'listening', label: 'Listening', icon: <Radio size={16} /> },
  { id: 'defender', label: 'Defender', icon: <Activity size={16} /> },
  { id: 'events', label: 'Events', icon: <AlertTriangle size={16} /> },
]

// ── Inline helpers ──

function StatusBadge({ status }: { status: string }) {
  const colorMap: Record<string, string> = {
    allow: 'bg-success/20 text-success',
    block: 'bg-danger/20 text-danger',
    enabled: 'bg-success/20 text-success',
    disabled: 'bg-[var(--color-border)]/50 text-text-dim',
    running: 'text-[var(--color-accent)] bg-[var(--color-accent)]/10',
    ready: 'bg-success/20 text-success',
    listening: 'bg-[var(--color-accent)]/10 text-[var(--color-accent)]',
    error: 'bg-danger/20 text-danger',
    warning: 'bg-warning/20 text-warning',
    info: 'bg-[var(--color-accent)]/10 text-[var(--color-accent)]',
  }
  const key = status.toLowerCase()
  return (
    <span
      className={cn(
        'inline-block px-2 py-0.5 text-[10px] font-medium rounded-full',
        colorMap[key] || 'bg-[var(--color-border)]/50 text-text-dim',
      )}
    >
      {status}
    </span>
  )
}

function SectionHeader({ title, subtitle }: { title: string; subtitle?: string }) {
  return (
    <div className="mb-3">
      <h3 className="text-xs font-semibold text-text-dim uppercase tracking-wider">{title}</h3>
      {subtitle && <p className="text-[11px] text-text-dim mt-0.5">{subtitle}</p>}
    </div>
  )
}

function CopyChip({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)
  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    })
  }, [text])
  return (
    <button
      onClick={handleCopy}
      className="inline-flex items-center gap-1 px-2 py-0.5 bg-panel border border-[var(--color-border)] rounded-full text-[11px] text-text font-mono hover:border-primary/50 transition-colors shrink-0"
    >
      {text}
      {copied ? <Check size={10} className="text-success" /> : <Copy size={10} className="text-text-dim" />}
    </button>
  )
}

function MiniStat({ label, value, icon, color }: { label: string; value: string | number; icon?: React.ReactNode; color?: string }) {
  return (
    <div className="bg-panel border border-[var(--color-border)] rounded-[12px] p-4 flex items-center gap-3">
      {icon && <span className={cn('shrink-0', color ? '' : 'text-[var(--color-accent)]')} style={color ? { color } : undefined}>{icon}</span>}
      <div>
        <p className="text-[10px] text-text-dim uppercase tracking-wider">{label}</p>
        <p className="text-sm font-semibold text-text">{value}</p>
      </div>
    </div>
  )
}

// ── Firewall Tab ──

function FirewallTab({ call }: { call: (method: string, ...args: any[]) => Promise<any> }) {
  const [rules, setRules] = useState<FirewallRule[]>(() => mockFirewallRules())
  const [search, setSearch] = useState('')
  const [dirFilter, setDirFilter] = useState<'all' | 'Inbound' | 'Outbound'>('all')
  const [actionFilter, setActionFilter] = useState<'all' | 'Allow' | 'Block'>('all')
  const [profileFilter, setProfileFilter] = useState<string>('all')

  useEffect(() => {
    call('SecOps.GetFirewallRules').then((data: any) => {
      if (data) setRules(data as FirewallRule[])
    })
  }, [call])

  const filtered = rules.filter((r) => {
    const matchesSearch = r.name.toLowerCase().includes(search.toLowerCase())
    const matchesDir = dirFilter === 'all' || r.direction === dirFilter
    const matchesAction = actionFilter === 'all' || r.action === actionFilter
    const matchesProfile = profileFilter === 'all' || r.profile.toLowerCase().includes(profileFilter.toLowerCase())
    return matchesSearch && matchesDir && matchesAction && matchesProfile
  })

  const allowCount = rules.filter((r) => r.action === 'Allow').length
  const blockCount = rules.filter((r) => r.action === 'Block').length
  const profiles = Array.from(new Set(rules.flatMap((r) => r.profile.split(',').map((p) => p.trim()))))

  const toggleRule = (index: number) => {
    setRules((prev) => prev.map((r, i) => (i === index ? { ...r, enabled: !r.enabled } : r)))
  }

  return (
    <div className="space-y-4">
      {/* Summary */}
      <div className="grid grid-cols-3 gap-3">
        <MiniStat label="Total Rules" value={rules.length} icon={<Shield size={14} />} />
        <MiniStat label="Allow" value={allowCount} icon={<Check size={14} />} color="#4ade80" />
        <MiniStat label="Block" value={blockCount} icon={<X size={14} />} color="#f87171" />
      </div>

      {/* Filters */}
      <div className="flex items-center gap-3 flex-wrap">
        <div className="relative flex-1 min-w-[200px]">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-text-dim" />
          <input
            type="text"
            placeholder="Search by rule name..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full bg-panel border border-[var(--color-border)] rounded-[9px] pl-9 pr-3 py-2 text-sm text-text placeholder-text-dim focus:outline-none focus:ring-2 focus:ring-[var(--color-accent)]/50"
          />
        </div>
        <div className="flex gap-1 bg-[var(--color-panel-2)] rounded-[9px] p-1">
          {(['all', 'Inbound', 'Outbound'] as const).map((d) => (
            <button
              key={d}
              onClick={() => setDirFilter(d)}
              className={cn('px-3 py-1.5 rounded-md text-xs font-medium transition-all whitespace-nowrap', dirFilter === d ? 'bg-panel text-[var(--color-accent)] shadow-sm' : 'text-text-dim hover:text-text')}
            >
              {d === 'all' ? 'All' : d}
            </button>
          ))}
        </div>
        <div className="flex gap-1 bg-[var(--color-panel-2)] rounded-[9px] p-1">
          {(['all', 'Allow', 'Block'] as const).map((a) => (
            <button
              key={a}
              onClick={() => setActionFilter(a)}
              className={cn('px-3 py-1.5 rounded-md text-xs font-medium transition-all', actionFilter === a ? 'bg-panel text-[var(--color-accent)] shadow-sm' : 'text-text-dim hover:text-text')}
            >
              {a === 'all' ? 'All' : a}
            </button>
          ))}
        </div>
        <div className="flex gap-1 bg-[var(--color-panel-2)] rounded-[9px] p-1">
          {['all', ...profiles].map((p) => (
            <button
              key={p}
              onClick={() => setProfileFilter(p)}
              className={cn('px-3 py-1.5 rounded-md text-xs font-medium transition-all', profileFilter === p ? 'bg-panel text-[var(--color-accent)] shadow-sm' : 'text-text-dim hover:text-text')}
            >
              {p === 'all' ? 'All' : p}
            </button>
          ))}
        </div>
      </div>

      {/* Table */}
      <div className="bg-panel border border-[var(--color-border)] rounded-[12px] overflow-hidden">
        <div className="overflow-x-auto max-h-[520px] overflow-y-auto">
          <table className="w-full text-sm">
            <thead className="sticky top-0 bg-panel z-10">
              <tr className="border-b border-[var(--color-border)]">
                <th className="text-left px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">Name</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">Direction</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">Action</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">Protocol</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">Local Port</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">Remote Port</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">Remote IP</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">Profile</th>
                <th className="text-center px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">Enabled</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((rule, i) => (
                <tr key={i} className="border-b border-[var(--color-border)]/50 hover:bg-[var(--color-panel-2)]/50 transition-colors">
                  <td className="px-4 py-2.5 text-text font-medium text-xs">{rule.name}</td>
                  <td className="px-4 py-2.5">
                    <StatusBadge status={rule.direction} />
                  </td>
                  <td className="px-4 py-2.5">
                    <span className={cn('inline-block px-1.5 py-0.5 text-[10px] font-medium rounded-full', rule.action === 'Allow' ? 'bg-success/20 text-success' : 'bg-danger/20 text-danger')}>
                      {rule.action}
                    </span>
                  </td>
                  <td className="px-4 py-2.5 text-xs font-mono text-text-dim">{rule.protocol}</td>
                  <td className="px-4 py-2.5 text-xs font-mono text-text-dim">{rule.local_port}</td>
                  <td className="px-4 py-2.5 text-xs font-mono text-text-dim">{rule.remote_port}</td>
                  <td className="px-4 py-2.5 text-xs font-mono text-text-dim max-w-[120px] truncate">{rule.remote_ip}</td>
                  <td className="px-4 py-2.5">
                    <StatusBadge status={rule.profile.split(',')[0]} />
                  </td>
                  <td className="px-4 py-2.5 text-center">
                    <button
                      onClick={() => toggleRule(i)}
                      className={cn(
                        'relative inline-flex h-5 w-9 items-center rounded-full transition-colors',
                        rule.enabled ? 'bg-success' : 'bg-[var(--color-border)]',
                      )}
                    >
                      <span
                        className={cn(
                          'inline-block h-3.5 w-3.5 rounded-full bg-white transition-transform shadow-sm',
                          rule.enabled ? 'translate-x-[18px]' : 'translate-x-[2px]',
                        )}
                      />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {filtered.length === 0 && (
          <p className="text-sm text-text-dim text-center py-8">No rules match your criteria.</p>
        )}
      </div>
    </div>
  )
}

// ── Users Tab ──

function UsersTab({ call }: { call: (method: string, ...args: any[]) => Promise<any> }) {
  const [users, setUsers] = useState<UserInfo[]>(() => mockUsers())
  const [search, setSearch] = useState('')

  useEffect(() => {
    call('SecOps.GetUsers').then((data: any) => {
      if (data) setUsers(data as UserInfo[])
    })
  }, [call])

  const filtered = users.filter((u) =>
    u.username.toLowerCase().includes(search.toLowerCase()) ||
    u.full_name.toLowerCase().includes(search.toLowerCase()),
  )

  const toggleUser = (index: number) => {
    setUsers((prev) => prev.map((u, i) => (i === index ? { ...u, is_enabled: !u.is_enabled } : u)))
  }

  const adminCount = users.filter((u) => u.is_admin).length
  const enabledCount = users.filter((u) => u.is_enabled).length

  return (
    <div className="space-y-4">
      {/* Summary */}
      <div className="grid grid-cols-3 gap-3">
        <MiniStat label="Total Users" value={users.length} icon={<Users size={14} />} />
        <MiniStat label="Admins" value={adminCount} icon={<Shield size={14} />} color="#fbbf24" />
        <MiniStat label="Enabled" value={enabledCount} icon={<Check size={14} />} color="#4ade80" />
      </div>

      {/* Search */}
      <div className="relative max-w-sm">
        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-text-dim" />
        <input
          type="text"
          placeholder="Search by username or name..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full bg-panel border border-[var(--color-border)] rounded-[9px] pl-9 pr-3 py-2 text-sm text-text placeholder-text-dim focus:outline-none focus:ring-2 focus:ring-[var(--color-accent)]/50"
        />
      </div>

      {/* Table */}
      <div className="bg-panel border border-[var(--color-border)] rounded-[12px] overflow-hidden">
        <div className="overflow-x-auto max-h-[520px] overflow-y-auto">
          <table className="w-full text-sm">
            <thead className="sticky top-0 bg-panel z-10">
              <tr className="border-b border-[var(--color-border)]">
                <th className="text-left px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">Username</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">Full Name</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">SID</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">Group</th>
                <th className="text-center px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">Admin</th>
                <th className="text-center px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">Enabled</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((user, i) => (
                <tr key={user.username} className="border-b border-[var(--color-border)]/50 hover:bg-[var(--color-panel-2)]/50 transition-colors">
                  <td className="px-4 py-2.5 text-text font-medium text-xs">{user.username}</td>
                  <td className="px-4 py-2.5 text-xs text-text-dim">{user.full_name}</td>
                  <td className="px-4 py-2.5 text-xs font-mono text-text-dim max-w-[160px] truncate">
                    <CopyChip text={user.sid} />
                  </td>
                  <td className="px-4 py-2.5">
                    <span className="inline-flex items-center gap-1 px-1.5 py-0.5 bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-full text-[10px] font-medium text-text-dim">
                      {user.group}
                    </span>
                  </td>
                  <td className="px-4 py-2.5 text-center">
                    {user.is_admin ? (
                      <span className="inline-flex items-center gap-1 px-2 py-0.5 bg-warning/20 text-warning text-[10px] font-medium rounded-full">Admin</span>
                    ) : (
                      <span className="text-text-dim text-[10px]">—</span>
                    )}
                  </td>
                  <td className="px-4 py-2.5 text-center">
                    <button
                      onClick={() => toggleUser(i)}
                      className={cn(
                        'relative inline-flex h-5 w-9 items-center rounded-full transition-colors',
                        user.is_enabled ? 'bg-success' : 'bg-[var(--color-border)]',
                      )}
                    >
                      <span
                        className={cn(
                          'inline-block h-3.5 w-3.5 rounded-full bg-white transition-transform shadow-sm',
                          user.is_enabled ? 'translate-x-[18px]' : 'translate-x-[2px]',
                        )}
                      />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {filtered.length === 0 && (
          <p className="text-sm text-text-dim text-center py-8">No users match your search.</p>
        )}
      </div>
    </div>
  )
}

// ── Listening Ports Tab ──

function ListeningTab({ call }: { call: (method: string, ...args: any[]) => Promise<any> }) {
  const [ports, setPorts] = useState<ListeningPort[]>(() => mockListeningPorts())
  const [protocol, setProtocol] = useState<'all' | 'TCP' | 'UDP'>('all')
  const [portRangeMin, setPortRangeMin] = useState('')
  const [portRangeMax, setPortRangeMax] = useState('')

  useEffect(() => {
    call('SecOps.GetListeningPorts').then((data: any) => {
      if (data) setPorts(data as ListeningPort[])
    })
  }, [call])

  const filtered = ports.filter((p) => {
    if (protocol !== 'all' && p.protocol !== protocol) return false
    if (portRangeMin && p.port < parseInt(portRangeMin)) return false
    if (portRangeMax && p.port > parseInt(portRangeMax)) return false
    return true
  })

  const tcpCount = ports.filter((p) => p.protocol === 'TCP').length
  const udpCount = ports.filter((p) => p.protocol === 'UDP').length

  const getPortColor = (port: number) => {
    if (port < 1024) return 'text-warning'
    if (port < 49152) return 'text-[var(--color-accent)]'
    return 'text-text-dim'
  }

  return (
    <div className="space-y-4">
      {/* Summary */}
      <div className="grid grid-cols-2 gap-3">
        <MiniStat label="TCP Ports" value={tcpCount} icon={<Globe size={14} />} />
        <MiniStat label="UDP Ports" value={udpCount} icon={<Wifi size={14} />} />
      </div>

      {/* Filters */}
      <div className="flex items-center gap-3 flex-wrap">
        <div className="flex gap-1 bg-[var(--color-panel-2)] rounded-[9px] p-1">
          {(['all', 'TCP', 'UDP'] as const).map((p) => (
            <button
              key={p}
              onClick={() => setProtocol(p)}
              className={cn('px-4 py-1.5 rounded-md text-xs font-medium transition-all', protocol === p ? 'bg-panel text-[var(--color-accent)] shadow-sm' : 'text-text-dim hover:text-text')}
            >
              {p === 'all' ? 'All' : p}
            </button>
          ))}
        </div>
        <div className="flex items-center gap-2">
          <input
            type="number"
            placeholder="Min port"
            value={portRangeMin}
            onChange={(e) => setPortRangeMin(e.target.value)}
            className="w-20 bg-panel border border-[var(--color-border)] rounded-[9px] px-3 py-1.5 text-xs text-text placeholder-text-dim focus:outline-none focus:ring-2 focus:ring-[var(--color-accent)]/50"
          />
          <span className="text-text-dim text-xs">to</span>
          <input
            type="number"
            placeholder="Max port"
            value={portRangeMax}
            onChange={(e) => setPortRangeMax(e.target.value)}
            className="w-20 bg-panel border border-[var(--color-border)] rounded-[9px] px-3 py-1.5 text-xs text-text placeholder-text-dim focus:outline-none focus:ring-2 focus:ring-[var(--color-accent)]/50"
          />
        </div>
        <div className="flex-1" />
        <span className="text-[11px] text-text-dim">
          {filtered.length} of {ports.length} ports
        </span>
      </div>

      {/* Table */}
      <div className="bg-panel border border-[var(--color-border)] rounded-[12px] overflow-hidden">
        <div className="overflow-x-auto max-h-[480px] overflow-y-auto">
          <table className="w-full text-sm">
            <thead className="sticky top-0 bg-panel z-10">
              <tr className="border-b border-[var(--color-border)]">
                <th className="text-left px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">Port</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">Protocol</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">Process Name</th>
                <th className="text-right px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">PID</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">State</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((p, i) => (
                <tr key={i} className="border-b border-[var(--color-border)]/50 hover:bg-[var(--color-panel-2)]/50 transition-colors">
                  <td className={cn('px-4 py-2.5 text-xs font-mono font-medium', getPortColor(p.port))}>{p.port}</td>
                  <td className="px-4 py-2.5">
                    <StatusBadge status={p.protocol} />
                  </td>
                  <td className="px-4 py-2.5 text-xs text-text">{p.process_name}</td>
                  <td className="px-4 py-2.5 text-right text-xs font-mono text-text-dim">{p.pid}</td>
                  <td className="px-4 py-2.5">
                    <StatusBadge status={p.state} />
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

// ── Defender Tab ──

function DefenderTab({ call }: { call: (method: string, ...args: any[]) => Promise<any> }) {
  const [status, setStatus] = useState<DefenderStatus>(() => mockDefenderStatus())

  useEffect(() => {
    call('SecOps.GetDefenderStatus').then((data: any) => {
      if (data) setStatus(data as DefenderStatus)
    })
  }, [call])

  const healthCards = [
    { label: 'Antivirus', active: status.enabled, icon: <Shield size={16} /> },
    { label: 'Up to Date', active: status.up_to_date, icon: <Clock size={16} /> },
    { label: 'Real-time Protection', active: status.real_time_protection, icon: <Activity size={16} /> },
    { label: 'Cloud Protection', active: status.cloud_protection, icon: <Wifi size={16} /> },
  ]

  return (
    <div className="space-y-4">
      {/* 2x2 Health Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        {healthCards.map((card) => (
          <div
            key={card.label}
            className={cn(
              'bg-panel border rounded-[12px] p-5 flex items-center justify-between transition-all',
              card.active ? 'border-[var(--color-border)]' : 'border-danger/30',
            )}
          >
            <div className="flex items-center gap-3">
              <span className={card.active ? 'text-success' : 'text-danger'}>{card.icon}</span>
              <span className="text-sm text-text font-medium">{card.label}</span>
            </div>
            <span
              className={cn(
                'inline-flex items-center gap-1.5 px-3 py-1 text-[11px] font-medium rounded-full',
                card.active ? 'bg-success/20 text-success' : 'bg-danger/20 text-danger',
              )}
            >
              <span className={cn('w-2 h-2 rounded-full', card.active ? 'bg-success' : 'bg-danger')} />
              {card.active ? 'On' : 'Off'}
            </span>
          </div>
        ))}
      </div>

      {/* Signature details */}
      <SectionHeader title="Scan & Signature Status" />
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
        <div className="bg-panel border border-[var(--color-border)] rounded-[12px] p-4">
          <p className="text-xs text-text-dim uppercase tracking-wider mb-1">Signature Age</p>
          <p className="text-sm text-text font-semibold">{status.signature_age}</p>
        </div>
        <div className="bg-panel border border-[var(--color-border)] rounded-[12px] p-4">
          <p className="text-xs text-text-dim uppercase tracking-wider mb-1">Last Quick Scan</p>
          <p className="text-sm text-text font-semibold">{status.quick_scan_age === 0 ? 'Today' : `${status.quick_scan_age} day(s) ago`}</p>
          <p className="text-xs text-text-dim mt-0.5">{status.last_scan}</p>
        </div>
        <div className="bg-panel border border-[var(--color-border)] rounded-[12px] p-4">
          <p className="text-xs text-text-dim uppercase tracking-wider mb-1">Last Full Scan</p>
          <p className="text-sm text-text font-semibold">{status.full_scan_age} day(s) ago</p>
        </div>
      </div>

      {/* Service status */}
      <SectionHeader title="Service Status" />
      <div className="bg-panel border border-[var(--color-border)] rounded-[12px] p-4">
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          {[
            { active: status.am_service_enabled, label: 'Anti-malware Service' },
            { active: status.antispyware_enabled, label: 'Anti-spyware Service' },
            { active: status.nis_enabled, label: 'Network Inspection System' },
          ].map((svc) => (
            <div key={svc.label} className="flex items-center gap-2">
              <span className={cn('w-2.5 h-2.5 rounded-full', svc.active ? 'bg-success shadow-[0_0_6px_rgba(74,222,128,0.5)]' : 'bg-danger shadow-[0_0_6px_rgba(248,113,113,0.5)]')} />
              <span className="text-xs text-text">{svc.label}</span>
              <span className={cn('text-xs font-medium ml-auto', svc.active ? 'text-success' : 'text-danger')}>
                {svc.active ? 'Running' : 'Stopped'}
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

// ── Events Tab ──

function EventsTab({ call }: { call: (method: string, ...args: any[]) => Promise<any> }) {
  const [events, setEvents] = useState<SecurityEvent[]>(() => mockSecurityEvents())
  const [levelFilter, setLevelFilter] = useState<'all' | 'Error' | 'Warning' | 'Info'>('all')
  const [autoRefresh, setAutoRefresh] = useState(false)

  useEffect(() => {
    call('SecOps.GetSecurityEvents').then((data: any) => {
      if (data) setEvents(data as SecurityEvent[])
    })
  }, [call])

  useEffect(() => {
    if (!autoRefresh) return
    const interval = setInterval(() => {
      call('SecOps.GetSecurityEvents').then((data: any) => {
        if (data) setEvents(data as SecurityEvent[])
      })
    }, 10000)
    return () => clearInterval(interval)
  }, [autoRefresh, call])

  const filtered = levelFilter === 'all' ? events : events.filter((e) => e.level === levelFilter)

  const getLevelColor = (level: string) => {
    switch (level) {
      case 'Error': return 'bg-danger/20 text-danger'
      case 'Warning': return 'bg-warning/20 text-warning'
      case 'Info': return 'bg-[var(--color-accent)]/10 text-[var(--color-accent)]'
      default: return 'bg-[var(--color-border)]/50 text-text-dim'
    }
  }

  const errorCount = events.filter((e) => e.level === 'Error').length
  const warnCount = events.filter((e) => e.level === 'Warning').length

  return (
    <div className="space-y-4">
      {/* Summary */}
      <div className="grid grid-cols-3 gap-3">
        <MiniStat label="Total Events" value={events.length} icon={<AlertTriangle size={14} />} />
        <MiniStat label="Errors" value={errorCount} icon={<X size={14} />} color="#f87171" />
        <MiniStat label="Warnings" value={warnCount} icon={<AlertTriangle size={14} />} color="#fbbf24" />
      </div>

      {/* Controls */}
      <div className="flex items-center gap-3 flex-wrap">
        <div className="flex gap-1 bg-[var(--color-panel-2)] rounded-[9px] p-1">
          {(['all', 'Error', 'Warning', 'Info'] as const).map((l) => (
            <button
              key={l}
              onClick={() => setLevelFilter(l)}
              className={cn('px-3 py-1.5 rounded-md text-xs font-medium transition-all', levelFilter === l ? 'bg-panel text-[var(--color-accent)] shadow-sm' : 'text-text-dim hover:text-text')}
            >
              {l === 'all' ? 'All' : l}
            </button>
          ))}
        </div>
        <div className="flex-1" />
        <span className="text-[11px] text-text-dim">{filtered.length} events shown</span>
        <button
          onClick={() => setAutoRefresh((v) => !v)}
          className={cn(
            'flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-[9px] border transition-colors',
            autoRefresh ? 'bg-[var(--color-accent)]/10 text-[var(--color-accent)] border-[var(--color-accent)]/30' : 'text-text-dim border-[var(--color-border)] hover:bg-[var(--color-panel-2)]',
          )}
        >
          <RefreshCw size={12} className={autoRefresh ? 'animate-spin' : ''} />
          Auto-refresh
        </button>
      </div>

      {/* Virtual-scrolled event list (overflow scroll) */}
      <div className="bg-panel border border-[var(--color-border)] rounded-[12px] overflow-hidden">
        <div className="overflow-y-auto max-h-[520px]">
          <table className="w-full text-sm">
            <thead className="sticky top-0 bg-panel z-10">
              <tr className="border-b border-[var(--color-border)]">
                <th className="text-left px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">Event ID</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">Level</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">Timestamp</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">Provider</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-text-dim uppercase tracking-wider">Message</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((event, i) => (
                <tr
                  key={i}
                  className={cn(
                    'border-b border-[var(--color-border)]/50 hover:bg-[var(--color-panel-2)]/50 transition-colors',
                    event.important && 'bg-warning/5',
                    event.level === 'Error' && 'bg-danger/5',
                  )}
                >
                  <td className="px-4 py-2.5 text-xs font-mono text-text">{event.id}</td>
                  <td className="px-4 py-2.5">
                    <span className={cn('inline-block px-1.5 py-0.5 text-[10px] font-medium rounded', getLevelColor(event.level))}>
                      {event.level}
                    </span>
                  </td>
                  <td className="px-4 py-2.5 text-xs text-text-dim font-mono whitespace-nowrap">{event.time}</td>
                  <td className="px-4 py-2.5 text-xs text-text-dim max-w-[200px] truncate">{event.provider}</td>
                  <td className="px-4 py-2.5 text-xs text-text max-w-[400px] truncate">{event.message}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {filtered.length === 0 && (
          <p className="text-sm text-text-dim text-center py-8">No events match your criteria.</p>
        )}
      </div>
    </div>
  )
}

// ── Main SecOps Page ──

export function SecOps() {
  const { call } = useBackend()
  const [activeTab, setActiveTab] = useState<SecOpsTab>('firewall')

  return (
    <div className="p-6 space-y-6 overflow-y-auto h-full">
      {/* Page header */}
      <div>
        <h1 className="text-2xl font-bold text-text flex items-center gap-2">
          <Shield size={24} className="text-[var(--color-accent)]" /> Security Operations
        </h1>
        <p className="text-text-dim text-sm mt-1">Firewall rules, user accounts, listening ports, Windows Defender, and security events</p>
      </div>

      {/* Squib-style tab bar */}
      <div className="bg-[var(--color-panel-2)] rounded-[9px] p-1 inline-flex">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={cn(
              'flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-all whitespace-nowrap',
              activeTab === tab.id
                ? 'bg-panel text-[var(--color-accent)] shadow-sm'
                : 'text-text-dim hover:text-text',
            )}
          >
            {tab.icon}
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab content */}
      <div className="space-y-4">
        {activeTab === 'firewall' && <FirewallTab call={call} />}
        {activeTab === 'users' && <UsersTab call={call} />}
        {activeTab === 'listening' && <ListeningTab call={call} />}
        {activeTab === 'defender' && <DefenderTab call={call} />}
        {activeTab === 'events' && <EventsTab call={call} />}
      </div>
    </div>
  )
}
