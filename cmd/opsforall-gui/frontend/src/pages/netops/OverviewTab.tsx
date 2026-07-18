import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import {
  Globe,
  Wifi,
  Server,
  ShieldCheck,
  Signal,
  ArrowUpRight,
  ArrowDownRight,
  Activity,
  Play,
  RefreshCw,
  Network,
  Clock,
  ArrowUpCircle,
  ArrowDownCircle,
  PlusCircle,
  MinusCircle,
  CircleDot,
  CircleOff,
  Zap,
  AlertTriangle,
  CheckCircle2,
  Radio,
} from 'lucide-react'
import { SectionHeader } from '@/components/ui/SectionHeader'
import { Panel } from '@/components/ui/Panel'
import type {
  ConnectionInfo,
  InterfaceInfo,
  NetworkChange,
  NetworkSummary,
  GatewayInfo,
  PingStats,
} from '@/types'

// ── Types ──

type ConnectivityStatus = 'ok' | 'error' | 'unknown' | 'checking'

interface ConnectivityCardData {
  label: string
  icon: React.ReactNode
  status: ConnectivityStatus
  detail: string
}

// ── Helpers ──

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${units[i]}`
}

// ── Connectivity Card ──

function ConnectivityCard({ label, icon, status, detail }: ConnectivityCardData) {
  const statusStyle = {
    ok: 'border-success/40 bg-success/5 hover:border-success/60',
    error: 'border-danger/40 bg-danger/5 hover:border-danger/60',
    unknown: 'border-warning/40 bg-warning/5 hover:border-warning/60',
    checking: 'border-border bg-panel-2',
  }[status]

  const dotStyle = {
    ok: 'bg-success shadow-[0_0_8px_var(--color-success)]',
    error: 'bg-danger shadow-[0_0_8px_var(--color-danger)]',
    unknown: 'bg-warning shadow-[0_0_8px_var(--color-warning)]',
    checking: 'bg-text-faint animate-pulse',
  }[status]

  return (
    <div className={cn('border rounded-2xl p-5 flex items-center gap-4 transition-all duration-300 group hover:shadow-lg', statusStyle)}>
      <div className="w-12 h-12 rounded-xl bg-panel flex items-center justify-center border border-border shadow-inner group-hover:scale-110 transition-transform">
        {icon}
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-1">
          <span className={cn('w-2 h-2 rounded-full shrink-0', dotStyle)} />
          <span className="text-[10px] font-black text-text uppercase tracking-[0.2em]">{label}</span>
        </div>
        <p className="text-xs font-bold text-text-dim truncate tabular-nums">{detail}</p>
      </div>
    </div>
  )
}

// ── Main OverviewTab ──

export function OverviewTab() {
  const { call } = useBackend()
  const { refreshInterval, dnsTimeout } = useSettingsStore()

  // ── Section 1: Connectivity Status ──

  const internetQuery = useQuery({
    queryKey: ['ov-internet'],
    queryFn: async () => {
      try {
        const res = await call('NetOps.Ping', '8.8.8.8', 1) as PingStats & { error?: string }
        if (res?.error) return { status: 'error' as ConnectivityStatus, detail: String(res.error) }
        if (res?.lost > 0) return { status: 'error' as ConnectivityStatus, detail: 'Packet loss detected' }
        return { status: 'ok' as ConnectivityStatus, detail: `${res.avg_ms}ms latency` }
      } catch { /* ignore */
        return { status: 'error' as ConnectivityStatus, detail: 'Ping failed' }
      }
    },
    refetchInterval: refreshInterval,
    retry: false,
  })

  const dnsProbeQuery = useQuery({
    queryKey: ['ov-dns-probe'],
    queryFn: async () => {
      try {
        const res = await call('NetOps.DNSLookup', 'google.com', '', Math.min(dnsTimeout, 3000)) as { a?: string[]; error?: string }
        if (res?.error) return { status: 'error' as ConnectivityStatus, detail: res.error }
        const count = res?.a?.length || 0
        return { status: 'ok' as ConnectivityStatus, detail: `Resolved ${count} A record${count !== 1 ? 's' : ''}` }
      } catch { /* ignore */
        return { status: 'error' as ConnectivityStatus, detail: 'Resolution failed' }
      }
    },
    refetchInterval: refreshInterval,
    retry: false,
  })

  const { data: connInterfaces = [], isLoading: ifacesLoading } = useQuery<InterfaceInfo[]>({
    queryKey: ['ov-interfaces'],
    queryFn: async () => {
      const res = await call('NetOps.GetInterfaces') as InterfaceInfo[]
      return res || []
    },
    refetchInterval: refreshInterval,
  })

  const { data: gatewayData } = useQuery({
    queryKey: ['ov-gateway'],
    queryFn: async () => {
      try {
        const res = await call('NetOps.GetDefaultGateway') as GatewayInfo
        return res
      } catch { /* ignore */
        return null
      }
    },
    refetchInterval: refreshInterval,
    retry: false,
  })

  const nonLoopbackUp = connInterfaces.filter(i => i.is_up && !i.name.includes('[Loopback]'))

  const lanStatus: { status: ConnectivityStatus; detail: string } = nonLoopbackUp.length > 0
    ? { status: 'ok', detail: `${nonLoopbackUp.length} active adapter${nonLoopbackUp.length !== 1 ? 's' : ''}` }
    : { status: 'error', detail: 'No active adapters' }

  const vpnPatterns = /vpn|wireguard|tunnel|ppp|tun|tap/i
  const vpnIface = connInterfaces.find(i => vpnPatterns.test(i.name))
  const vpnStatus: { status: ConnectivityStatus; detail: string } = vpnIface
    ? { status: 'ok', detail: vpnIface.name.replace(/^\[.*?\]\s*/, '') }
    : { status: 'unknown', detail: 'No VPN adapter detected' }

  const gatewayStatus: { status: ConnectivityStatus; detail: string } = gatewayData
    ? { status: gatewayData.reachable ? 'ok' : 'error', detail: gatewayData.ip }
    : { status: 'unknown', detail: 'Gateway detection N/A' }

  const connectivityCards: ConnectivityCardData[] = [
    { label: 'Internet', icon: <Globe size={20} className="text-accent" />, status: internetQuery.data?.status || 'checking', detail: internetQuery.isLoading ? 'Checking...' : internetQuery.data?.detail || 'Unavailable' },
    { label: 'LAN', icon: <Wifi size={20} className="text-accent" />, ...lanStatus },
    { label: 'Gateway', icon: <Server size={20} className="text-accent" />, ...gatewayStatus },
    { label: 'DNS', icon: <Signal size={20} className="text-accent" />, status: dnsProbeQuery.data?.status || 'checking', detail: dnsProbeQuery.isLoading ? 'Checking...' : dnsProbeQuery.data?.detail || 'Unavailable' },
    { label: 'VPN', icon: <ShieldCheck size={20} className="text-accent" />, ...vpnStatus },
  ]

  // ── Section 2: Active Connections ──

  const { data: connections = [] } = useQuery<ConnectionInfo[]>({
    queryKey: ['ov-connections'],
    queryFn: async () => {
      const res = await call('NetOps.GetConnections')
      return (res as ConnectionInfo[]) || []
    },
    refetchInterval: refreshInterval,
  })

  const establishedCount = connections.filter(c => c.state === 'ESTABLISHED').length
  const listeningCount = connections.filter(c => c.state === 'LISTEN').length
  const timeWaitCount = connections.filter(c => c.state === 'TIME_WAIT').length

  // ── Section 3: Interface Health (reuse connInterfaces from connectivity) ──

  // ── Section 4: Network Quality (Ping) ──

  const [pingRunning, setPingRunning] = useState(false)
  const [pingResult, setPingResult] = useState<any>(null)

  const runPingTest = async () => {
    setPingRunning(true)
    setPingResult(null)
    try {
      const res = await call('NetOps.Ping', '8.8.8.8', 4) as PingStats & { error?: string }
      setPingResult(res)
    } catch { /* ignore */
      setPingResult({ error: 'Ping test failed' })
    } finally {
      setPingRunning(false)
    }
  }

  // ── Section 5: DNS Summary ──

  const { data: dnsSummaryResult } = useQuery({
    queryKey: ['ov-dns-summary'],
    queryFn: async () => {
      try {
        const start = Date.now()
        const res = await call('NetOps.DNSLookup', 'example.com', '', Math.min(dnsTimeout, 3000)) as { a?: string[]; ns?: string[]; error?: string }
        const latency = Date.now() - start
        if (res?.error) return { resolver: 'N/A', latency, error: res.error }
        return { resolver: res.ns?.[0] || 'System default', latency, error: null }
      } catch { /* ignore */
        return { resolver: 'N/A', latency: 0, error: 'Lookup failed' }
      }
    },
    refetchInterval: refreshInterval,
    retry: false,
  })

  // ── Section 6: Recent Changes ──

  const { data: recentChanges = [] } = useQuery<NetworkChange[]>({
    queryKey: ['ov-recent-changes'],
    queryFn: async () => {
      const res = await call('NetOps.GetRecentChanges') as NetworkChange[]
      return res || []
    },
    refetchInterval: refreshInterval,
  })

  // ── Section 7: AI Summary ──

  const { data: netSummary } = useQuery<NetworkSummary>({
    queryKey: ['ov-network-summary'],
    queryFn: async () => {
      try {
        const res = await call('NetOps.GetNetworkSummary') as NetworkSummary
        return res || { summaryText: '', topInterface: '', issues: [] }
      } catch { /* ignore */
        return { summaryText: '', topInterface: '', issues: [] }
      }
    },
    refetchInterval: refreshInterval,
    retry: false,
  })

  // ── Render ──

  return (
    <div className="space-y-6 animate-in fade-in duration-500">

      {/* ── Section 1: Connectivity Status ── */}
      <Panel padding="md" category="network">
        <SectionHeader icon={<Signal size={18} className="text-accent" />} title="Connectivity Status" />
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4">
          {connectivityCards.map(card => (
            <ConnectivityCard key={card.label} {...card} />
          ))}
        </div>
      </Panel>

      {/* ── Section 2: Active Connections ── */}
      <Panel padding="lg" category="network">
        <SectionHeader icon={<Activity size={20} className="text-accent" />} title="Active Connections" count={connections.length} />
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {[
            { label: 'Established', value: establishedCount, color: 'var(--color-success)', icon: <ArrowUpRight size={20} /> },
            { label: 'Listening', value: listeningCount, color: 'var(--color-accent)', icon: <Radio size={20} /> },
            { label: 'Time Wait', value: timeWaitCount, color: 'var(--color-warning)', icon: <Clock size={20} /> },
          ].map(card => (
            <div
              key={card.label}
              className="bg-panel-2 border border-border rounded-2xl p-6 flex items-center gap-5 hover:border-accent/40 transition-all hover:shadow-lg group"
            >
              <div
                className="w-14 h-14 rounded-2xl flex items-center justify-center shrink-0 shadow-inner group-hover:scale-105 transition-transform"
                style={{ backgroundColor: `color-mix(in srgb, ${card.color} 10%, transparent)`, color: card.color, border: `1px solid color-mix(in srgb, ${card.color} 20%, transparent)` }}
              >
                {card.icon}
              </div>
              <div>
                <p className="text-3xl font-black text-text tabular-nums tracking-tighter">{card.value}</p>
                <p className="text-[10px] font-black text-text-faint uppercase tracking-[0.2em] mt-1">{card.label}</p>
              </div>
            </div>
          ))}
        </div>
      </Panel>

      {/* ── Section 3: Interface Health ── */}
      <Panel padding="md" category="network">
        <SectionHeader icon={<Wifi size={18} className="text-accent" />} title="Interface Health" count={connInterfaces.length} />
        {ifacesLoading ? (
          <div className="flex items-center justify-center py-10">
            <RefreshCw size={20} className="animate-spin text-text-faint" />
          </div>
        ) : connInterfaces.length === 0 ? (
          <p className="text-sm font-medium text-text-faint text-center py-8">No interfaces detected.</p>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {connInterfaces.map(iface => (
              <div
                key={iface.name}
                className={cn(
                  'bg-panel-2 border rounded-xl p-4 transition-all',
                  iface.is_up ? 'border-success/30 hover:border-success/50' : 'border-danger/30 hover:border-danger/50',
                )}
              >
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2 min-w-0">
                    <span className={cn('w-2 h-2 rounded-full shrink-0', iface.is_up ? 'bg-success shadow-[0_0_6px_var(--color-success)]' : 'bg-danger shadow-[0_0_6px_var(--color-danger)]')} />
                    <span className="text-xs font-bold text-text uppercase tracking-wider truncate">{iface.name}</span>
                  </div>
                  <span className={cn('text-[10px] font-bold px-2 py-0.5 rounded-full border uppercase tracking-widest',
                    iface.is_up ? 'bg-success/15 text-success border-success/30' : 'bg-danger/15 text-danger border-danger/30',
                  )}>
                    {iface.is_up ? 'UP' : 'DOWN'}
                  </span>
                </div>
                <div className="space-y-2 text-[11px]">
                  <div className="flex justify-between">
                    <span className="text-text-faint">Speed</span>
                    <span className="font-bold text-text-dim">{iface.speed || 'N/A'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-text-faint">IP</span>
                    <span className="font-bold text-text-dim truncate max-w-[140px]">{iface.ips?.[0] || 'N/A'}</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-text-faint">RX</span>
                    <span className="font-bold text-success flex items-center gap-1">
                      <ArrowDownRight size={12} />{formatBytes(iface.rx_bytes)}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-text-faint">TX</span>
                    <span className="font-bold text-accent flex items-center gap-1">
                      <ArrowUpRight size={12} />{formatBytes(iface.tx_bytes)}
                    </span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </Panel>

      {/* ── Section 4: Network Quality ── */}
      <Panel padding="lg" category="network">
        <SectionHeader icon={<Zap size={20} className="text-accent" />} title="Network Quality" />
        <div className="flex items-center gap-8">
          <button
            onClick={runPingTest}
            disabled={pingRunning}
            className={cn(
              'flex items-center gap-3 px-8 py-4 text-base font-black uppercase tracking-widest rounded-2xl transition-all shadow-xl active:scale-95',
              pingRunning
                ? 'bg-panel-3 text-text-faint border border-border cursor-not-allowed'
                : 'bg-accent text-white hover:bg-accent/90 shadow-accent/20 hover:shadow-accent/30',
            )}
          >
            {pingRunning ? <RefreshCw size={20} className="animate-spin" /> : <Play size={20} fill="currentColor" />}
            {pingRunning ? 'PROBING...' : 'RUN PING DIAGNOSTIC'}
          </button>
          {!pingRunning && !pingResult && (
            <p className="text-xs font-bold text-text-faint uppercase tracking-widest">Target: ICMP ECHO (8.8.8.8)</p>
          )}
        </div>

        {pingResult && !pingResult.error && (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mt-4">
            {(() => {
              // The backend never sends lost_pct (H12) — derive the
              // percentage from sent/lost instead of a phantom field.
              const lostPct = pingResult.sent > 0 ? (pingResult.lost / pingResult.sent) * 100 : null;
              return [
                { label: 'Latency', value: pingResult.avg_ms != null ? `${pingResult.avg_ms.toFixed(1)}ms` : 'N/A', color: 'var(--color-success)' },
                { label: 'Jitter', value: pingResult.avg_ms != null && pingResult.min_ms != null ? `${(pingResult.avg_ms - pingResult.min_ms).toFixed(1)}ms` : 'N/A', color: 'var(--color-warning)' },
                { label: 'Packet Loss', value: lostPct != null ? `${lostPct.toFixed(1)}%` : 'N/A', color: lostPct != null && lostPct > 0 ? 'var(--color-danger)' : 'var(--color-success)' },
              ];
            })().map(card => (
              <div key={card.label} className="bg-panel-2 border border-border rounded-xl p-4">
                <p className="text-[10px] font-bold text-text-faint uppercase tracking-wider mb-1">{card.label}</p>
                <p className="text-xl font-bold tabular-nums" style={{ color: card.color }}>{card.value}</p>
              </div>
            ))}
          </div>
        )}

        {pingResult?.error && (
          <div className="mt-4 flex items-center gap-3 bg-danger/10 border border-danger/30 rounded-xl p-4">
            <AlertTriangle size={16} className="text-danger shrink-0" />
            <p className="text-sm font-medium text-danger">{pingResult.error}</p>
          </div>
        )}
      </Panel>

      {/* ── Section 5: DNS Summary ── */}
      <Panel padding="md" category="network">
        <SectionHeader icon={<Globe size={18} className="text-accent" />} title="DNS Summary" />
        {dnsSummaryResult ? (
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-panel-2 border border-border rounded-xl p-4">
              <p className="text-[10px] font-bold text-text-faint uppercase tracking-wider mb-1">Resolver</p>
              <p className="text-sm font-bold text-text truncate">{dnsSummaryResult.resolver}</p>
            </div>
            <div className="bg-panel-2 border border-border rounded-xl p-4">
              <p className="text-[10px] font-bold text-text-faint uppercase tracking-wider mb-1">Lookup Latency</p>
              <p className={cn('text-sm font-bold tabular-nums', dnsSummaryResult.latency > 500 ? 'text-warning' : 'text-success')}>
                {dnsSummaryResult.latency}ms
              </p>
            </div>
          </div>
        ) : (
          <p className="text-sm font-medium text-text-faint text-center py-4">Loading DNS information...</p>
        )}
      </Panel>

      {/* ── Section 6: Recent Changes ── */}
      <Panel padding="md" category="network">
        <SectionHeader icon={<Network size={18} className="text-accent" />} title="Recent Changes" count={recentChanges.length} />
        {recentChanges.length === 0 ? (
          <p className="text-sm font-medium text-text-faint text-center py-6">No interface changes detected yet.</p>
        ) : (
          <div className="space-y-2">
            {recentChanges.slice(0, 5).map((change, idx) => {
              const iconMap: Record<string, React.ReactNode> = {
                up: <ArrowUpCircle size={14} className="text-success" />,
                down: <ArrowDownCircle size={14} className="text-danger" />,
                ip_added: <PlusCircle size={14} className="text-accent" />,
                ip_removed: <MinusCircle size={14} className="text-warning" />,
                appeared: <CircleDot size={14} className="text-success" />,
                disappeared: <CircleOff size={14} className="text-danger" />,
              }
              const labelMap: Record<string, string> = {
                up: 'UP',
                down: 'DOWN',
                ip_added: 'IP ADDED',
                ip_removed: 'IP REMOVED',
                appeared: 'APPEARED',
                disappeared: 'GONE',
              }
              const colorMap: Record<string, string> = {
                up: 'bg-success/15 text-success border-success/30',
                down: 'bg-danger/15 text-danger border-danger/30',
                ip_added: 'bg-accent/15 text-accent border-accent/30',
                ip_removed: 'bg-warning/15 text-warning border-warning/30',
                appeared: 'bg-success/15 text-success border-success/30',
                disappeared: 'bg-danger/15 text-danger border-danger/30',
              }
              return (
                <div key={idx} className="flex items-center gap-3 p-3 bg-panel-2 border border-border rounded-xl">
                  {iconMap[change.type]}
                  <span className={cn('px-2 py-0.5 text-[10px] font-bold uppercase tracking-widest rounded-full border', colorMap[change.type] || 'bg-text-faint/20 text-text-faint border-border')}>
                    {labelMap[change.type] || change.type}
                  </span>
                  <span className="text-xs font-bold text-text uppercase tracking-tight">{change.interface}</span>
                  <span className="text-xs font-medium text-text-dim flex-1 truncate">{change.detail}</span>
                  <span className="text-[10px] font-medium text-text-faint tabular-nums whitespace-nowrap">
                    {new Date(change.timestamp).toLocaleTimeString()}
                  </span>
                </div>
              )
            })}
          </div>
        )}
      </Panel>

      {/* ── Section 7: AI Summary ── */}
      <Panel padding="md" category="network">
        <div className="flex items-center gap-3 mb-4">
          <div className="w-9 h-9 rounded-lg flex items-center justify-center bg-panel-3 border border-border" style={{ color: 'var(--color-accent)' }}>
            <Zap size={18} />
          </div>
          <h3 className="text-sm font-bold text-text uppercase tracking-widest">AI Summary</h3>
        </div>
        {!netSummary || !netSummary.summaryText ? (
          <p className="text-sm font-medium text-text-faint italic">No summary available. AI analysis will appear once the backend generates a network briefing.</p>
        ) : (
          <div className="space-y-4">
            <p className="text-sm text-text-dim leading-relaxed">{netSummary.summaryText}</p>
            <div className="flex items-center gap-6 text-xs">
              {netSummary.topInterface && (
                <span className="text-text-faint">
                  <span className="font-bold text-accent">Top Interface:</span> {netSummary.topInterface}
                </span>
              )}
            </div>
            {netSummary.issues && netSummary.issues.length > 0 && (
              <div className="mt-3 space-y-2">
                <p className="text-xs font-bold text-warning uppercase tracking-wider flex items-center gap-2">
                  <AlertTriangle size={12} /> Issues Detected
                </p>
                {netSummary.issues.map((issue, i) => (
                  <div key={i} className="flex items-start gap-2 text-sm text-text-dim">
                    <CheckCircle2 size={14} className="text-warning shrink-0 mt-0.5" />
                    <span>{issue}</span>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </Panel>

    </div>
  )
}
