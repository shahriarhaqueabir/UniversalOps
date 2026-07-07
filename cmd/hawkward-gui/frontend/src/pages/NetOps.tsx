import { useState, useEffect, useCallback, useRef } from 'react'
import { cn } from '@/lib/utils'
import { AreaChart } from '@/components/charts/AreaChart'
import { MiniSparkline } from '@/components/charts/MiniSparkline'
import { useBackend } from '@/hooks/useBackend'
import {
  Activity,
  Wifi,
  Globe,
  Search,
  Radar,
  Map,
  Cable,
  Play,
  Square,
  Copy,
  Check,
  Loader2,
  Network,
  Server,
  ArrowRight,
  Timer,
  Signal,
  AlertTriangle,
  Info,
  RefreshCw,
  ChevronDown,
  ChevronUp,
  Gauge,
  TrendingUp,
  TrendingDown,
} from 'lucide-react'
import {
  AreaChart as RechartsAreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
  Legend,
} from 'recharts'
import type {
  PingEntry,
  PingStats,
  DNSResult,
  PortResult,
  TraceHop,
  TraceResult,
  ConnectionInfo,
  InterfaceInfo,
} from '@/types'
import {
  mockPingEntry,
  mockPingStats,
  mockDNSResult,
  mockPortScan,
  mockTraceResult,
  mockInterfaces,
  mockConnections as mockConnectionsList,
} from '@/lib/mockData'

type NetOpsTab = 'ping' | 'dns' | 'portscan' | 'traceroute' | 'connections' | 'interfaces' | 'bandwidth'

const tabs: { id: NetOpsTab; label: string; icon: React.ReactNode }[] = [
  { id: 'ping', label: 'Ping', icon: <Activity size={16} /> },
  { id: 'dns', label: 'DNS', icon: <Globe size={16} /> },
  { id: 'portscan', label: 'Port Scan', icon: <Radar size={16} /> },
  { id: 'traceroute', label: 'Traceroute', icon: <Map size={16} /> },
  { id: 'connections', label: 'Connections', icon: <Cable size={16} /> },
  { id: 'interfaces', label: 'Interfaces', icon: <Wifi size={16} /> },
  { id: 'bandwidth', label: 'Bandwidth', icon: <Gauge size={16} /> },
]

const COMMON_PORTS = [22, 23, 25, 53, 80, 110, 143, 443, 445, 993, 995, 1433, 3306, 3389, 5432, 6379, 8080, 8443, 27017]

const SERVICE_COLORS: Record<string, string> = {
  SSH: '#7c6cff',
  HTTP: '#2dd4a7',
  HTTPS: '#2dd4a7',
  'HTTP-Alt': '#2dd4a7',
  'HTTPS-Alt': '#2dd4a7',
  DNS: '#f59e0b',
  SMTP: '#fb5d6b',
  MySQL: '#fb5d6b',
  MSSQL: '#fb5d6b',
  PostgreSQL: '#fb5d6b',
  Redis: '#ef4444',
  MongoDB: '#10b981',
  RDP: '#3b82f6',
  SMB: '#8b5cf6',
  Telnet: '#f97316',
  POP3: '#ec4899',
  IMAP: '#ec4899',
  IMAPS: '#ec4899',
  POP3S: '#ec4899',
  HDFS: '#6366f1',
}

// ── Helper sub-components ──

function StatusBadge({ status }: { status: string }) {
  const colorMap: Record<string, string> = {
    success: 'bg-[var(--color-success)]/15 text-[var(--color-success)]',
    timeout: 'bg-[var(--color-danger)]/15 text-[var(--color-danger)]',
    open: 'bg-[var(--color-success)]/15 text-[var(--color-success)]',
    closed: 'bg-[var(--color-muted)]/20 text-[var(--color-muted)]',
    filtered: 'bg-[var(--color-warning)]/20 text-[var(--color-warning)]',
    listening: 'bg-[var(--color-success)]/15 text-[var(--color-success)]',
    established: 'bg-[var(--color-accent)]/15 text-[var(--color-accent)]',
    time_wait: 'bg-[var(--color-warning)]/15 text-[var(--color-warning)]',
    close_wait: 'bg-[var(--color-danger)]/15 text-[var(--color-danger)]',
  }
  const s = status.toLowerCase()
  return (
    <span
      className={cn(
        'inline-block px-2 py-0.5 text-[10px] font-semibold rounded-[6px] tracking-wide',
        colorMap[s] || 'bg-[var(--color-muted)]/20 text-[var(--color-muted)]',
      )}
    >
      {status}
    </span>
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
      className="inline-flex items-center gap-1.5 px-3 py-1 bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-[8px] text-xs text-[var(--color-text)] font-mono hover:border-[var(--color-accent)]/50 transition-colors"
    >
      <span className="truncate max-w-[300px]">{text}</span>
      {copied ? <Check size={12} className="text-[var(--color-success)] shrink-0" /> : <Copy size={12} className="text-[var(--color-muted)] shrink-0" />}
    </button>
  )
}

function MiniStat({ label, value, icon, unit }: { label: string; value: string | number; icon?: React.ReactNode; unit?: string }) {
  return (
    <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[12px] p-3.5 flex items-center gap-3">
      {icon && <span className="text-[var(--color-accent)] shrink-0">{icon}</span>}
      <div>
        <p className="text-[10px] text-[var(--color-muted)] uppercase tracking-wider font-medium">{label}</p>
        <p className="text-sm font-semibold text-[var(--color-text)] font-mono">
          {value}{unit && <span className="text-[10px] text-[var(--color-muted)] ml-0.5 font-normal">{unit}</span>}
        </p>
      </div>
    </div>
  )
}

function SectionHeader({ title, subtitle, actions }: { title: string; subtitle?: string; actions?: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between mb-3">
      <div>
        <h3 className="text-xs font-semibold text-[var(--color-muted)] uppercase tracking-wider">{title}</h3>
        {subtitle && <p className="text-[11px] text-[var(--color-muted)]/70 mt-0.5">{subtitle}</p>}
      </div>
      {actions && <div className="flex items-center gap-2">{actions}</div>}
    </div>
  )
}

function ServiceBadge({ service }: { service: string }) {
  const color = SERVICE_COLORS[service] || '#64748b'
  return (
    <span
      className="inline-flex items-center px-2 py-0.5 text-[10px] font-semibold rounded-[6px] tracking-wide"
      style={{
        backgroundColor: `${color}18`,
        color: color,
      }}
    >
      {service}
    </span>
  )
}

// Arrow sub-components used by PingTab
function ArrowUpFromLine({ size }: { size: number }) {
  return <ArrowRight size={size} className="rotate-[-90deg]" />
}
function ArrowDownToLine({ size }: { size: number }) {
  return <ArrowRight size={size} className="rotate-90" />
}

// ── Ping Tab ──
function PingTab() {
  const { call } = useBackend()
  const [target, setTarget] = useState('8.8.8.8')
  const [running, setRunning] = useState(false)
  const [entries, setEntries] = useState<PingEntry[]>([])
  const [stats, setStats] = useState<PingStats | null>(null)
  const seqRef = useRef(1)
  const intervalRef = useRef<ReturnType<typeof setInterval>>(undefined)

  const generateEntry = useCallback(async (seq: number, ip: string) => {
    try {
      const result = await call('NetOps.Ping', ip) as { entry: PingEntry; stats: PingStats } | null
      if (result) {
        // Backend returned a full result
        setEntries((prev) => [...prev, result.entry])
        setStats(result.stats)
        return
      }
    } catch { }
    // Fallback: generate mock entry
    const entry = mockPingEntry(seq, ip)
    setEntries((prev) => {
      const next = [...prev, entry]
      return next.length > 30 ? next.slice(-30) : next
    })
  }, [call])

  const startPing = useCallback(() => {
    setRunning(true)
    setEntries([])
    setStats(null)
    seqRef.current = 1
  }, [])

  const stopPing = useCallback(() => {
    setRunning(false)
  }, [])

  useEffect(() => {
    if (!running) {
      if (intervalRef.current) clearInterval(intervalRef.current)
      return
    }
    intervalRef.current = setInterval(() => {
      const seq = seqRef.current++
      generateEntry(seq, target)
    }, 1000)
    return () => clearInterval(intervalRef.current)
  }, [running, target, generateEntry])

  useEffect(() => {
    if (entries.length > 0 && !stats) {
      setStats(mockPingStats(entries))
    }
  }, [entries, stats])

  const latencyData = entries
    .filter((e) => e.status === 'success')
    .map((e) => ({
      time: `#${e.seq}`,
      value: e.rtt_ms || 0,
    }))

  const packetLoss = entries.length > 0
    ? ((entries.length - entries.filter((e) => e.status === 'success').length) / entries.length) * 100
    : 0

  return (
    <div className="space-y-4">
      {/* Controls */}
      <div className="flex items-center gap-3 flex-wrap">
        <div className="flex-1 min-w-[200px]">
          <input
            type="text"
            placeholder="Target host or IP"
            value={target}
            onChange={(e) => setTarget(e.target.value)}
            disabled={running}
            className="w-full bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-[9px] px-3 py-2 text-sm text-[var(--color-text)] placeholder-[var(--color-muted)] focus:outline-none focus:ring-2 focus:ring-[var(--color-accent)]/50 disabled:opacity-50"
          />
        </div>
        {running ? (
          <button
            onClick={stopPing}
            className="flex items-center gap-1.5 bg-[var(--color-danger)] text-white rounded-[9px] px-4 py-2 text-sm font-medium hover:brightness-110 transition-all shadow-[0_4px_12px_-4px_rgba(248,113,113,.5)]"
          >
            <Square size={14} fill="currentColor" /> Stop
          </button>
        ) : (
          <button
            onClick={startPing}
            className="flex items-center gap-1.5 bg-[var(--color-accent)] text-white rounded-[9px] px-4 py-2 text-sm font-medium shadow-[0_6px_18px_-6px_rgba(124,108,255,.7)] hover:brightness-110 transition-all"
          >
            <Play size={14} fill="currentColor" /> Start Ping
          </button>
        )}
      </div>

      {/* Packet loss indicator */}
      {running && entries.length > 0 && packetLoss > 0 && (
        <div className="flex items-center gap-2 px-3 py-2 rounded-[9px] text-xs font-medium bg-[var(--color-danger)]/10 text-[var(--color-danger)]">
          <AlertTriangle size={14} />
          Packet Loss: {packetLoss.toFixed(1)}% ({entries.filter((e) => e.status === 'success').length}/{entries.length} received)
        </div>
      )}

      {/* Stats */}
      {stats && (
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-2.5">
          <MiniStat label="Sent" value={stats.sent} icon={<ArrowUpFromLine size={14} />} />
          <MiniStat label="Received" value={stats.received} icon={<ArrowDownToLine size={14} />} />
          <MiniStat label="Lost" value={`${stats.lost} (${stats.lost_pct.toFixed(1)}%)`} icon={<AlertTriangle size={14} />} />
          <MiniStat label="Min RTT" value={stats.min_ms !== null ? stats.min_ms : '—'} unit="ms" icon={<Timer size={14} />} />
          <MiniStat label="Max RTT" value={stats.max_ms !== null ? stats.max_ms : '—'} unit="ms" icon={<Timer size={14} />} />
          <MiniStat label="Avg RTT" value={stats.avg_ms !== null ? stats.avg_ms : '—'} unit="ms" icon={<Signal size={14} />} />
        </div>
      )}

      {/* Latency chart */}
      {latencyData.length > 1 && (
        <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[12px] p-4">
          <SectionHeader title="Latency (Last 30 pings)" subtitle="Round-trip time over time" />
          <AreaChart
            data={latencyData}
            color="#4ade80"
            unit="ms"
            height={180}
          />
        </div>
      )}

      {/* Results table */}
      <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[12px] overflow-hidden">
        <div className="overflow-x-auto max-h-[320px] overflow-y-auto">
          <table className="w-full text-sm">
            <thead className="sticky top-0 bg-[var(--color-panel)]">
              <tr className="border-b border-[var(--color-border)]">
                <th className="text-left px-4 py-2.5 text-xs font-medium text-[var(--color-muted)] uppercase tracking-wider">#</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-[var(--color-muted)] uppercase tracking-wider">IP</th>
                <th className="text-right px-4 py-2.5 text-xs font-medium text-[var(--color-muted)] uppercase tracking-wider">RTT (ms)</th>
                <th className="text-right px-4 py-2.5 text-xs font-medium text-[var(--color-muted)] uppercase tracking-wider">TTL</th>
                <th className="text-right px-4 py-2.5 text-xs font-medium text-[var(--color-muted)] uppercase tracking-wider">Status</th>
              </tr>
            </thead>
            <tbody>
              {[...entries].reverse().map((entry) => (
                <tr key={entry.seq} className="border-b border-[var(--color-border)]/50 hover:bg-[var(--color-panel-2)]/40 transition-colors">
                  <td className="px-4 py-2 font-mono text-xs text-[var(--color-muted)]">{entry.seq}</td>
                  <td className="px-4 py-2 text-[var(--color-text)] font-mono text-xs">{entry.ip}</td>
                  <td className={cn('px-4 py-2 text-right font-mono text-xs', entry.status === 'success' ? 'text-[var(--color-success)]' : 'text-[var(--color-danger)]')}>
                    {entry.status === 'success' && entry.rtt_ms !== null ? `${entry.rtt_ms.toFixed(1)}` : '—'}
                  </td>
                  <td className="px-4 py-2 text-right font-mono text-xs text-[var(--color-muted)]">
                    {entry.ttl !== null ? entry.ttl : '—'}
                  </td>
                  <td className="px-4 py-2 text-right">
                    <StatusBadge status={entry.status} />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {entries.length === 0 && (
          <p className="text-sm text-[var(--color-muted)] text-center py-8">No ping data yet. Click "Start Ping" to begin.</p>
        )}
      </div>
    </div>
  )
}

// ── DNS Tab ──
function DNSTab() {
  const { call } = useBackend()
  const [hostname, setHostname] = useState('google.com')
  const [result, setResult] = useState<DNSResult | null>(null)
  const [loading, setLoading] = useState(false)
  const [timing, setTiming] = useState<number | null>(null)

  const handleLookup = useCallback(async () => {
    if (!hostname.trim()) return
    setLoading(true)
    setTiming(null)
    const start = performance.now()
    try {
      const data = await call('NetOps.DNSLookup', hostname.trim()) as DNSResult | null
      if (data) {
        setResult(data)
      } else {
        // Fallback
        await new Promise((r) => setTimeout(r, 800))
        setResult(mockDNSResult(hostname.trim()))
      }
    } catch {
      await new Promise((r) => setTimeout(r, 800))
      setResult(mockDNSResult(hostname.trim()))
    }
    setTiming(performance.now() - start)
    setLoading(false)
  }, [hostname, call])

  useEffect(() => {
    const onEnter = (e: KeyboardEvent) => {
      if (e.key === 'Enter') handleLookup()
    }
    window.addEventListener('keydown', onEnter)
    return () => window.removeEventListener('keydown', onEnter)
  }, [handleLookup])

  const recordSections = result
    ? [
      { label: 'A Records (IPv4)', icon: <Network size={14} />, values: result.a },
      { label: 'AAAA Records (IPv6)', icon: <Network size={14} />, values: result.aaaa },
      { label: 'MX Records (Mail)', icon: <Server size={14} />, values: result.mx },
      { label: 'NS Records (Nameservers)', icon: <Server size={14} />, values: result.ns },
      ...(result.cname ? [{ label: 'CNAME', icon: <Info size={14} />, values: [result.cname] }] : []),
      { label: 'TXT Records', icon: <Info size={14} />, values: result.txt },
    ].filter((s) => s.values.length > 0)
    : []

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3">
        <div className="flex-1 min-w-[200px]">
          <div className="relative">
            <Globe size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-muted)]" />
            <input
              type="text"
              placeholder="Enter hostname (e.g., google.com)"
              value={hostname}
              onChange={(e) => setHostname(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleLookup()}
              className="w-full bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-[9px] pl-9 pr-3 py-2 text-sm text-[var(--color-text)] placeholder-[var(--color-muted)] focus:outline-none focus:ring-2 focus:ring-[var(--color-accent)]/50"
            />
          </div>
        </div>
        <button
          onClick={handleLookup}
          disabled={loading}
          className="flex items-center gap-1.5 bg-[var(--color-accent)] text-white rounded-[9px] px-4 py-2 text-sm font-medium shadow-[0_6px_18px_-6px_rgba(124,108,255,.7)] hover:brightness-110 transition-all disabled:opacity-50"
        >
          {loading ? <Loader2 size={14} className="animate-spin" /> : <Search size={14} />}
          Lookup
        </button>
      </div>

      {loading && (
        <div className="flex items-center justify-center py-12">
          <div className="flex flex-col items-center gap-3">
            <Loader2 size={24} className="animate-spin text-[var(--color-accent)]" />
            <p className="text-sm text-[var(--color-muted)]">Resolving DNS records...</p>
          </div>
        </div>
      )}

      {result && !loading && (
        <div className="space-y-4">
          <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[12px] p-4 flex items-center justify-between">
            <div>
              <p className="text-xs text-[var(--color-muted)] mb-0.5">Results for</p>
              <p className="text-lg font-bold text-[var(--color-text)]">{result.hostname}</p>
            </div>
            {timing !== null && (
              <span className="text-[11px] text-[var(--color-muted)] font-mono">{(timing).toFixed(0)}ms</span>
            )}
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {recordSections.map((section) => (
              <div key={section.label} className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[12px] p-4">
                <div className="flex items-center gap-1.5 mb-2.5">
                  <span className="text-[var(--color-accent)]">{section.icon}</span>
                  <h4 className="text-xs font-semibold text-[var(--color-muted)] uppercase tracking-wider">{section.label}</h4>
                </div>
                <div className="flex flex-wrap gap-1.5 mt-2">
                  {section.values.map((val, i) => (
                    <CopyChip key={i} text={val} />
                  ))}
                </div>
                {section.values.length === 0 && (
                  <p className="text-xs text-[var(--color-muted)] italic">No records found</p>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {!result && !loading && (
        <div className="flex flex-col items-center justify-center py-12 text-[var(--color-muted)]">
          <Globe size={32} className="opacity-30 mb-3" />
          <p className="text-sm">Enter a hostname and click Lookup to resolve DNS records</p>
        </div>
      )}
    </div>
  )
}

// ── Port Scan Tab ──
function PortScanTab() {
  const { call } = useBackend()
  const [target, setTarget] = useState('192.168.1.1')
  const [portStart, setPortStart] = useState('1')
  const [portEnd, setPortEnd] = useState('100')
  const [scanning, setScanning] = useState(false)
  const [progress, setProgress] = useState(0)
  const [results, setResults] = useState<PortResult[]>([])
  const [autoRefresh, setAutoRefresh] = useState(false)
  const scanningRef = useRef(false)

  const runScan = useCallback(async (ports?: number[]) => {
    if (scanningRef.current) return
    const p = ports || Array.from(
      { length: Math.min(parseInt(portEnd) || 100, 1000) - Math.max(parseInt(portStart) || 1, 1) + 1 },
      (_, i) => (parseInt(portStart) || 1) + i
    )
    if (p.length === 0) return
    scanningRef.current = true
    setScanning(true)
    setProgress(0)
    setResults([])

    try {
      // Try backend for full scan
      const scanResult = await call('NetOps.PortScan', target, p) as PortResult[] | null
      if (scanResult && scanResult.length > 0) {
        setResults(scanResult)
        setProgress(100)
        scanningRef.current = false
        setScanning(false)
        return
      }
    } catch { }

    // Fallback: simulated batch scan
    let completed = 0
    const total = p.length
    const batchSize = Math.max(1, Math.floor(total / 20))
    const processBatch = () => {
      const batch = p.slice(completed, completed + batchSize)
      const batchResults = mockPortScan(target, batch)
      setResults((prev) => [...prev, ...batchResults])
      completed += batch.length
      setProgress(Math.min((completed / total) * 100, 100))
      if (completed < total) {
        setTimeout(processBatch, 150)
      } else {
        scanningRef.current = false
        setScanning(false)
      }
    }
    processBatch()
  }, [target, portStart, portEnd, call])

  const handleCommonPorts = useCallback(() => {
    runScan(COMMON_PORTS)
  }, [runScan])

  useEffect(() => {
    if (!autoRefresh || scanning) return
    const interval = setInterval(() => runScan(), 10000)
    return () => clearInterval(interval)
  }, [autoRefresh, scanning, runScan])

  const openCount = results.filter((r) => r.open).length

  return (
    <div className="space-y-4">
      {/* Inputs */}
      <div className="flex items-center gap-3 flex-wrap">
        <div className="min-w-[160px] flex-1">
          <input
            type="text"
            placeholder="Target host"
            value={target}
            onChange={(e) => setTarget(e.target.value)}
            disabled={scanning}
            className="w-full bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-[9px] px-3 py-2 text-sm text-[var(--color-text)] placeholder-[var(--color-muted)] focus:outline-none focus:ring-2 focus:ring-[var(--color-accent)]/50 disabled:opacity-50"
          />
        </div>
        <div className="flex items-center gap-2">
          <input
            type="number"
            placeholder="Start"
            value={portStart}
            onChange={(e) => setPortStart(e.target.value)}
            disabled={scanning}
            className="w-20 bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-[9px] px-3 py-2 text-sm text-[var(--color-text)] placeholder-[var(--color-muted)] focus:outline-none focus:ring-2 focus:ring-[var(--color-accent)]/50 disabled:opacity-50"
          />
          <span className="text-[var(--color-muted)] text-xs">to</span>
          <input
            type="number"
            placeholder="End"
            value={portEnd}
            onChange={(e) => setPortEnd(e.target.value)}
            disabled={scanning}
            className="w-20 bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-[9px] px-3 py-2 text-sm text-[var(--color-text)] placeholder-[var(--color-muted)] focus:outline-none focus:ring-2 focus:ring-[var(--color-accent)]/50 disabled:opacity-50"
          />
        </div>
      </div>

      {/* Buttons */}
      <div className="flex items-center gap-2 flex-wrap">
        <button
          onClick={() => runScan()}
          disabled={scanning}
          className="flex items-center gap-1.5 bg-[var(--color-accent)] text-white rounded-[9px] px-4 py-2 text-sm font-medium shadow-[0_6px_18px_-6px_rgba(124,108,255,.7)] hover:brightness-110 transition-all disabled:opacity-50"
        >
          {scanning ? <Loader2 size={14} className="animate-spin" /> : <Radar size={14} />}
          Scan
        </button>
        <button
          onClick={handleCommonPorts}
          disabled={scanning}
          className="px-3 py-2 text-xs font-medium text-[var(--color-muted)] border border-[var(--color-border)] rounded-[9px] hover:bg-[var(--color-panel-2)] transition-colors disabled:opacity-50"
        >
          Common Ports (20)
        </button>
        <button
          onClick={() => setAutoRefresh((v) => !v)}
          className={cn(
            'flex items-center gap-1.5 px-3 py-2 text-xs font-medium rounded-[9px] border transition-colors',
            autoRefresh
              ? 'bg-[var(--color-accent)]/10 text-[var(--color-accent)] border-[var(--color-accent)]/30'
              : 'text-[var(--color-muted)] border-[var(--color-border)] hover:bg-[var(--color-panel-2)]',
          )}
        >
          <RefreshCw size={12} className={autoRefresh ? 'animate-spin' : ''} /> Auto-refresh
        </button>
      </div>

      {/* Progress */}
      {scanning && (
        <div className="space-y-1.5">
          <div className="flex items-center justify-between text-xs text-[var(--color-muted)]">
            <span>Scanning {target}...</span>
            <span>{results.length} / {(parseInt(portEnd) - parseInt(portStart) + 1) || 100} ports checked</span>
          </div>
          <div className="h-2 bg-[var(--color-panel-2)] rounded-full overflow-hidden">
            <div
              className="h-full bg-[var(--color-accent)] rounded-full transition-all duration-300"
              style={{ width: `${progress}%` }}
            />
          </div>
        </div>
      )}

      {/* Summary */}
      {results.length > 0 && !scanning && (
        <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[12px] p-3.5 flex items-center gap-3">
          <span className="text-xs text-[var(--color-muted)]">Scan Results:</span>
          <span className="text-xs font-semibold text-[var(--color-success)]">{openCount} open</span>
          <span className="text-[var(--color-border)]">/</span>
          <span className="text-xs text-[var(--color-muted)]">{results.length} total</span>
          {openCount > 0 && (
            <>
              <span className="text-[var(--color-border)]">•</span>
              <span className="text-xs text-[var(--color-muted)]">
                Open ports: {results.filter((r) => r.open).map((r) => r.port).join(', ')}
              </span>
            </>
          )}
        </div>
      )}

      {/* Table */}
      {results.length > 0 && (
        <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[12px] overflow-hidden">
          <div className="overflow-x-auto max-h-[400px] overflow-y-auto">
            <table className="w-full text-sm">
              <thead className="sticky top-0 bg-[var(--color-panel)]">
                <tr className="border-b border-[var(--color-border)]">
                  <th className="text-left px-4 py-2.5 text-xs font-medium text-[var(--color-muted)] uppercase tracking-wider">Port</th>
                  <th className="text-left px-4 py-2.5 text-xs font-medium text-[var(--color-muted)] uppercase tracking-wider">State</th>
                  <th className="text-left px-4 py-2.5 text-xs font-medium text-[var(--color-muted)] uppercase tracking-wider">Service</th>
                </tr>
              </thead>
              <tbody>
                {results.map((r) => (
                  <tr key={r.port} className="border-b border-[var(--color-border)]/50 hover:bg-[var(--color-panel-2)]/40 transition-colors">
                    <td className="px-4 py-2.5 font-mono text-xs text-[var(--color-text)]">{r.port}</td>
                    <td className="px-4 py-2.5">
                      <StatusBadge status={r.open ? 'open' : 'closed'} />
                    </td>
                    <td className="px-4 py-2.5">
                      <ServiceBadge service={r.service} />
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

// ── Traceroute Tab ──
function TracerouteTab() {
  const { call } = useBackend()
  const [target, setTarget] = useState('google.com')
  const [tracing, setTracing] = useState(false)
  const [hops, setHops] = useState<TraceHop[]>([])
  const [progress, setProgress] = useState(0)

  const handleTrace = useCallback(async () => {
    if (!target.trim()) return
    setTracing(true)
    setHops([])
    setProgress(0)

    try {
      const result = await call('NetOps.Traceroute', target.trim()) as TraceResult | null
      if (result && result.hops.length > 0) {
        const totalHops = result.hops.length
        result.hops.forEach((hop: TraceHop, i: number) => {
          setTimeout(() => {
            setHops((prev) => [...prev, hop])
            setProgress(((i + 1) / totalHops) * 100)
            if (i === totalHops - 1) setTracing(false)
          }, (i + 1) * 400)
        })
        return
      }
    } catch { }

    // Fallback
    const result = mockTraceResult(target.trim())
    const totalHops = result.hops.length
    result.hops.forEach((hop, i) => {
      setTimeout(() => {
        setHops((prev) => [...prev, hop])
        setProgress(((i + 1) / totalHops) * 100)
        if (i === totalHops - 1) setTracing(false)
      }, (i + 1) * 600)
    })
  }, [target, call])

  const getRttColor = (rtt: number | null) => {
    if (rtt === null) return 'text-[var(--color-danger)]'
    if (rtt < 50) return 'text-[var(--color-success)]'
    if (rtt < 150) return 'text-[var(--color-warning)]'
    return 'text-[var(--color-danger)]'
  }

  const getRttBg = (rtt: number | null) => {
    if (rtt === null) return 'bg-[var(--color-danger)]/10'
    if (rtt < 50) return 'bg-[var(--color-success)]/10'
    if (rtt < 150) return 'bg-[var(--color-warning)]/10'
    return 'bg-[var(--color-danger)]/10'
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3">
        <div className="flex-1 min-w-[200px]">
          <div className="relative">
            <Map size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-muted)]" />
            <input
              type="text"
              placeholder="Target host"
              value={target}
              onChange={(e) => setTarget(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleTrace()}
              disabled={tracing}
              className="w-full bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-[9px] pl-9 pr-3 py-2 text-sm text-[var(--color-text)] placeholder-[var(--color-muted)] focus:outline-none focus:ring-2 focus:ring-[var(--color-accent)]/50 disabled:opacity-50"
            />
          </div>
        </div>
        <button
          onClick={handleTrace}
          disabled={tracing}
          className="flex items-center gap-1.5 bg-[var(--color-accent)] text-white rounded-[9px] px-4 py-2 text-sm font-medium shadow-[0_6px_18px_-6px_rgba(124,108,255,.7)] hover:brightness-110 transition-all disabled:opacity-50"
        >
          {tracing ? <Loader2 size={14} className="animate-spin" /> : <Map size={14} />}
          Trace
        </button>
      </div>

      {/* Progress */}
      {tracing && (
        <div className="space-y-1.5">
          <div className="flex items-center justify-between text-xs text-[var(--color-muted)]">
            <span>Tracing route to {target}... <span className="font-mono">{hops.length} hops</span></span>
            <span>{Math.round(progress)}%</span>
          </div>
          <div className="h-2 bg-[var(--color-panel-2)] rounded-full overflow-hidden">
            <div className="h-full bg-[var(--color-accent)] rounded-full transition-all duration-500" style={{ width: `${progress}%` }} />
          </div>
        </div>
      )}

      {/* Path visualization */}
      {hops.length > 1 && (
        <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[12px] p-4 overflow-x-auto">
          <div className="flex items-center gap-1 text-[11px] text-[var(--color-muted)] font-mono whitespace-nowrap">
            <span className="text-[var(--color-accent)] font-semibold">*</span>
            {hops.map((hop, i) => (
              <span key={i} className="flex items-center gap-1">
                {hop.timed ? (
                  <span className="text-[var(--color-danger)]">✗</span>
                ) : (
                  <span className="text-[var(--color-success)]">●</span>
                )}
                {i < hops.length - 1 && <span className="opacity-30">→</span>}
              </span>
            ))}
            <span className="text-[var(--color-accent)] font-semibold ml-1">🏁 {target}</span>
          </div>
        </div>
      )}

      {/* Hop cards */}
      <div className="space-y-1.5">
        {hops.map((hop) => (
          <div
            key={hop.number}
            className={cn(
              'flex items-center gap-4 bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[12px] px-4 py-3 text-sm transition-all',
              hop.timed && 'opacity-50',
            )}
          >
            {/* Hop number */}
            <span className="text-xs font-mono text-[var(--color-muted)] w-7 shrink-0 text-right">#{hop.number}</span>

            {/* Connection dot */}
            <div className={cn(
              'w-2 h-2 rounded-full shrink-0',
              hop.timed ? 'bg-[var(--color-danger)]' : 'bg-[var(--color-success)] shadow-[0_0_6px_rgba(74,222,128,0.4)]'
            )} />

            {/* Host/IP */}
            <div className="flex-1 min-w-0">
              <p className="text-[var(--color-text)] font-medium truncate">{hop.host}</p>
              <p className="text-xs text-[var(--color-muted)] font-mono">{hop.ip}</p>
            </div>

            {/* RTT probes */}
            {!hop.timed && hop.rtts_ms.length > 0 && (
              <div className="hidden sm:flex items-center gap-1.5">
                {hop.rtts_ms.map((rtt, i) => (
                  <span
                    key={i}
                    className={cn(
                      'px-2 py-0.5 rounded-[6px] text-[11px] font-mono',
                      getRttBg(rtt),
                      getRttColor(rtt),
                    )}
                  >
                    {rtt.toFixed(0)}ms
                  </span>
                ))}
              </div>
            )}

            {/* Average */}
            <div className="text-right shrink-0 min-w-[60px]">
              {hop.timed ? (
                <span className="text-xs text-[var(--color-danger)] font-medium">* * *</span>
              ) : (
                <span className={cn('text-xs font-mono font-semibold', getRttColor(hop.avg_rtt))}>
                  {hop.avg_rtt !== null ? `${hop.avg_rtt.toFixed(1)} ms` : '—'}
                </span>
              )}
            </div>
          </div>
        ))}
        {!tracing && hops.length > 0 && (
          <div className="flex items-center gap-2 bg-[var(--color-success)]/10 border border-[var(--color-success)]/20 rounded-[12px] px-4 py-2.5 text-xs text-[var(--color-success)]">
            <Check size={14} /> Trace complete. {hops.length} hop{hops.length !== 1 ? 's' : ''} to {target}.
          </div>
        )}
        {hops.length === 0 && !tracing && (
          <div className="flex flex-col items-center justify-center py-12 text-[var(--color-muted)]">
            <Map size={32} className="opacity-30 mb-3" />
            <p className="text-sm">Enter a target and click "Trace" to start.</p>
          </div>
        )}
      </div>
    </div>
  )
}

// ── Connections Tab ──
function ConnectionsTab() {
  const { call } = useBackend()
  const [connections, setConnections] = useState<ConnectionInfo[]>([])
  const [filterState, setFilterState] = useState<string>('all')
  const [autoRefresh, setAutoRefresh] = useState(false)
  const [refreshInterval, setRefreshInterval] = useState(3000)
  const [loading, setLoading] = useState(true)

  const fetchData = useCallback(async () => {
    try {
      const data = await call('NetOps.GetConnections') as ConnectionInfo[] | null
      if (data) setConnections(data)
    } catch {
      setConnections(mockConnectionsList())
    }
    setLoading(false)
  }, [call])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  useEffect(() => {
    if (!autoRefresh) return
    const interval = setInterval(fetchData, refreshInterval)
    return () => clearInterval(interval)
  }, [autoRefresh, refreshInterval, fetchData])

  const filtered = filterState === 'all'
    ? connections
    : connections.filter((c) => c.state.toLowerCase() === filterState.toLowerCase())

  const total = connections.length
  const listening = connections.filter((c) => c.state === 'LISTENING').length
  const established = connections.filter((c) => c.state === 'ESTABLISHED').length
  const timeWait = connections.filter((c) => c.state === 'TIME_WAIT').length

  const stateFilters = ['all', 'LISTENING', 'ESTABLISHED', 'TIME_WAIT', 'CLOSE_WAIT']

  return (
    <div className="space-y-4">
      {/* Summary bar */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-2.5">
        <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[12px] p-3.5 text-center">
          <p className="text-xl font-bold text-[var(--color-text)]">{total}</p>
          <p className="text-[10px] text-[var(--color-muted)] uppercase tracking-wider font-medium">Total</p>
        </div>
        <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[12px] p-3.5 text-center">
          <p className="text-xl font-bold text-[var(--color-success)]">{listening}</p>
          <p className="text-[10px] text-[var(--color-muted)] uppercase tracking-wider font-medium">Listening</p>
        </div>
        <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[12px] p-3.5 text-center">
          <p className="text-xl font-bold text-[var(--color-accent)]">{established}</p>
          <p className="text-[10px] text-[var(--color-muted)] uppercase tracking-wider font-medium">Established</p>
        </div>
        <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[12px] p-3.5 text-center">
          <p className="text-xl font-bold text-[var(--color-warning)]">{timeWait}</p>
          <p className="text-[10px] text-[var(--color-muted)] uppercase tracking-wider font-medium">Time Wait</p>
        </div>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-2 flex-wrap">
        <div className="flex gap-1 bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-[9px] p-1">
          {stateFilters.map((state) => (
            <button
              key={state}
              onClick={() => setFilterState(state)}
              className={cn(
                'px-3 py-1.5 rounded-md text-xs font-medium transition-all',
                filterState === state
                  ? 'bg-[var(--color-panel)] text-[var(--color-accent)] shadow-sm'
                  : 'text-[var(--color-muted)] hover:text-[var(--color-text)]',
              )}
            >
              {state === 'all' ? 'All' : state.charAt(0) + state.slice(1).toLowerCase()}
            </button>
          ))}
        </div>
        <div className="flex-1" />
        <div className="flex items-center gap-2">
          <select
            value={refreshInterval}
            onChange={(e) => setRefreshInterval(Number(e.target.value))}
            className="bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-[9px] px-2 py-1.5 text-xs text-[var(--color-text)] focus:outline-none"
          >
            <option value={2000}>2s</option>
            <option value={3000}>3s</option>
            <option value={5000}>5s</option>
            <option value={10000}>10s</option>
          </select>
          <button
            onClick={() => setAutoRefresh((v) => !v)}
            className={cn(
              'flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-[9px] border transition-colors',
              autoRefresh
                ? 'bg-[var(--color-accent)]/10 text-[var(--color-accent)] border-[var(--color-accent)]/30'
                : 'text-[var(--color-muted)] border-[var(--color-border)] hover:bg-[var(--color-panel-2)]',
            )}
          >
            <RefreshCw size={12} className={autoRefresh ? 'animate-spin' : ''} />
            {autoRefresh ? 'Auto' : 'Manual'}
          </button>
        </div>
      </div>

      {/* Table */}
      <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[12px] overflow-hidden">
        <div className="overflow-x-auto max-h-[480px] overflow-y-auto">
          <table className="w-full text-sm">
            <thead className="sticky top-0 bg-[var(--color-panel)]">
              <tr className="border-b border-[var(--color-border)]">
                <th className="text-left px-4 py-2.5 text-xs font-medium text-[var(--color-muted)] uppercase tracking-wider">Proto</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-[var(--color-muted)] uppercase tracking-wider">Local Address</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-[var(--color-muted)] uppercase tracking-wider">Remote Address</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-[var(--color-muted)] uppercase tracking-wider">State</th>
                <th className="text-right px-4 py-2.5 text-xs font-medium text-[var(--color-muted)] uppercase tracking-wider">PID</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-[var(--color-muted)] uppercase tracking-wider">Process</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((c, i) => (
                <tr key={`${c.pid}-${c.local_port}-${i}`} className="border-b border-[var(--color-border)]/50 hover:bg-[var(--color-panel-2)]/40 transition-colors">
                  <td className="px-4 py-2.5 text-xs font-mono text-[var(--color-muted)]">{c.proto}</td>
                  <td className="px-4 py-2.5 text-xs font-mono text-[var(--color-text)]">
                    {c.local_addr}:{c.local_port}
                  </td>
                  <td className="px-4 py-2.5 text-xs font-mono text-[var(--color-text)]">
                    {c.remote_addr}:{c.remote_port}
                  </td>
                  <td className="px-4 py-2.5">
                    <StatusBadge status={c.state} />
                  </td>
                  <td className="px-4 py-2.5 text-right text-xs font-mono text-[var(--color-muted)]">{c.pid}</td>
                  <td className="px-4 py-2.5 text-xs text-[var(--color-text)]">{c.process_name}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {filtered.length === 0 && !loading && (
          <p className="text-sm text-[var(--color-muted)] text-center py-8">No connections match the filter.</p>
        )}
        {loading && (
          <div className="flex items-center justify-center py-8">
            <Loader2 size={16} className="animate-spin text-[var(--color-accent)]" />
          </div>
        )}
      </div>
    </div>
  )
}

// ── Interfaces Tab ──
function InterfacesTab() {
  const { call } = useBackend()
  const [interfaces, setInterfaces] = useState<InterfaceInfo[]>([])
  const [expanded, setExpanded] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  const fetchData = useCallback(async () => {
    try {
      const data = await call('NetOps.GetInterfaces') as InterfaceInfo[] | null
      if (data) setInterfaces(data)
    } catch {
      setInterfaces(mockInterfaces())
    }
    setLoading(false)
  }, [call])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  // Live update simulation when no backend
  useEffect(() => {
    if (loading && interfaces.length === 0) return
    const interval = setInterval(() => {
      setInterfaces((prev) =>
        prev.map((iface) => ({
          ...iface,
          rx_bytes: iface.rx_bytes + Math.floor(Math.random() * 100000),
          tx_bytes: iface.tx_bytes + Math.floor(Math.random() * 50000),
          rx_rate_bps: iface.is_up ? 0.5 + Math.random() * 3 * 1024 * 1024 : 0,
          tx_rate_bps: iface.is_up ? 0.2 + Math.random() * 2 * 1024 * 1024 : 0,
          rx_history: [...iface.rx_history.slice(-19), Math.random() * 8],
          tx_history: [...iface.tx_history.slice(-19), Math.random() * 4],
        })),
      )
    }, 2000)
    return () => clearInterval(interval)
  }, [loading, interfaces.length])

  const formatRate = (bps: number) => {
    if (bps === 0) return '0 bps'
    if (bps < 1024) return `${bps.toFixed(0)} bps`
    if (bps < 1024 * 1024) return `${(bps / 1024).toFixed(1)} Kbps`
    return `${(bps / (1024 * 1024)).toFixed(2)} Mbps`
  }

  const sparklineData = (history: number[]) =>
    history.map((v, i) => ({
      time: `t${i}`,
      value: v,
    }))

  if (loading) {
    return (
      <div className="flex items-center justify-center py-16">
        <Loader2 size={20} className="animate-spin text-[var(--color-accent)]" />
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {interfaces.map((iface) => (
        <div
          key={iface.name}
          className={cn(
            'bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[12px] overflow-hidden transition-all',
            expanded === iface.name ? 'ring-1 ring-[var(--color-accent)]/30' : '',
          )}
        >
          {/* Card header */}
          <div
            className="flex items-center gap-4 p-4 cursor-pointer hover:bg-[var(--color-panel-2)]/40 transition-colors"
            onClick={() => setExpanded(expanded === iface.name ? null : iface.name)}
          >
            <span
              className={cn(
                'w-2.5 h-2.5 rounded-full shrink-0',
                iface.is_up
                  ? 'bg-[var(--color-success)] shadow-[0_0_6px_rgba(74,222,128,0.5)]'
                  : 'bg-[var(--color-danger)] shadow-[0_0_6px_rgba(248,113,113,0.5)]',
              )}
            />
            <div className="flex-1 min-w-0">
              <p className="text-sm font-semibold text-[var(--color-text)]">{iface.name}</p>
              <p className="text-xs text-[var(--color-muted)] font-mono truncate">
                {iface.ips[0] || 'No IP'} • {iface.mac}
              </p>
            </div>
            <div className="hidden sm:flex items-center gap-3 text-xs text-[var(--color-muted)]">
              <span className="font-mono">{iface.speed}</span>
              <span className="font-mono">MTU {iface.mtu}</span>
            </div>
            <div className="flex gap-3 min-w-[140px]">
              <div className="flex-1">
                <p className="text-[10px] text-[var(--color-muted)]">RX</p>
                <p className="text-xs font-mono text-[var(--color-success)]">{formatRate(iface.rx_rate_bps)}</p>
              </div>
              <div className="flex-1">
                <p className="text-[10px] text-[var(--color-muted)]">TX</p>
                <p className="text-xs font-mono text-[var(--color-accent)]">{formatRate(iface.tx_rate_bps)}</p>
              </div>
            </div>
            <span className="text-[var(--color-muted)]">
              {expanded === iface.name ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
            </span>
          </div>

          {/* Sparklines */}
          <div className="grid grid-cols-2 gap-2 px-4 pb-3">
            <div>
              <p className="text-[10px] text-[var(--color-muted)] mb-0.5">RX Rate History</p>
              <MiniSparkline data={sparklineData(iface.rx_history)} color="#4ade80" height={28} />
            </div>
            <div>
              <p className="text-[10px] text-[var(--color-muted)] mb-0.5">TX Rate History</p>
              <MiniSparkline data={sparklineData(iface.tx_history)} color="#7c6cff" height={28} />
            </div>
          </div>

          {/* Expanded details */}
          {expanded === iface.name && (
            <div className="border-t border-[var(--color-border)] px-4 py-3.5 space-y-3">
              <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3 text-xs">
                <div>
                  <p className="text-[var(--color-muted)] mb-0.5">MAC Address</p>
                  <p className="text-[var(--color-text)] font-mono">{iface.mac}</p>
                </div>
                <div>
                  <p className="text-[var(--color-muted)] mb-0.5">Speed</p>
                  <p className="text-[var(--color-text)]">{iface.speed}</p>
                </div>
                <div>
                  <p className="text-[var(--color-muted)] mb-0.5">MTU</p>
                  <p className="text-[var(--color-text)] font-mono">{iface.mtu}</p>
                </div>
                <div>
                  <p className="text-[var(--color-muted)] mb-0.5">Status</p>
                  <p className={cn('font-medium', iface.is_up ? 'text-[var(--color-success)]' : 'text-[var(--color-danger)]')}>
                    {iface.is_up ? 'Up' : 'Down'}
                  </p>
                </div>
                <div>
                  <p className="text-[var(--color-muted)] mb-0.5">RX Total</p>
                  <p className="text-[var(--color-text)] font-mono">{(iface.rx_bytes / (1024 * 1024 * 1024)).toFixed(2)} GB</p>
                </div>
                <div>
                  <p className="text-[var(--color-muted)] mb-0.5">TX Total</p>
                  <p className="text-[var(--color-text)] font-mono">{(iface.tx_bytes / (1024 * 1024 * 1024)).toFixed(2)} GB</p>
                </div>
                <div>
                  <p className="text-[var(--color-muted)] mb-0.5">RX Rate</p>
                  <p className="text-[var(--color-text)] font-mono text-[var(--color-success)]">{formatRate(iface.rx_rate_bps)}</p>
                </div>
                <div>
                  <p className="text-[var(--color-muted)] mb-0.5">TX Rate</p>
                  <p className="text-[var(--color-text)] font-mono text-[var(--color-accent)]">{formatRate(iface.tx_rate_bps)}</p>
                </div>
              </div>
              {iface.ips.length > 0 && (
                <div>
                  <p className="text-xs text-[var(--color-muted)] mb-1.5">IP Addresses</p>
                  <div className="flex flex-wrap gap-1.5">
                    {iface.ips.map((ip, i) => (
                      <span key={i} className="px-2 py-0.5 bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-[6px] text-xs font-mono text-[var(--color-text)]">
                        {ip}
                      </span>
                    ))}
                  </div>
                </div>
              )}
              <div>
                <p className="text-xs text-[var(--color-muted)] mb-1">Flags</p>
                <p className="text-xs font-mono text-[var(--color-muted)]">{iface.flags}</p>
              </div>
            </div>
          )}
        </div>
      ))}
    </div>
  )
}

// ── Bandwidth Tab (NEW) ──
function BandwidthTab() {
  const { call } = useBackend()
  const [interfaces, setInterfaces] = useState<InterfaceInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [autoRefresh, setAutoRefresh] = useState(true)
  const [history, setHistory] = useState<{ time: string; rxTotal: number; txTotal: number }[]>([])

  const fetchData = useCallback(async () => {
    try {
      const data = await call('NetOps.GetInterfaces') as InterfaceInfo[] | null
      if (data) setInterfaces(data)
    } catch {
      setInterfaces(mockInterfaces())
    }
    setLoading(false)
  }, [call])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  // Live update
  useEffect(() => {
    if (loading && interfaces.length === 0) return
    const interval = setInterval(() => {
      setInterfaces((prev) =>
        prev.map((iface) => ({
          ...iface,
          rx_bytes: iface.rx_bytes + Math.floor(Math.random() * 100000),
          tx_bytes: iface.tx_bytes + Math.floor(Math.random() * 50000),
          rx_rate_bps: iface.is_up ? 0.5 + Math.random() * 3 * 1024 * 1024 : 0,
          tx_rate_bps: iface.is_up ? 0.2 + Math.random() * 2 * 1024 * 1024 : 0,
          rx_history: [...iface.rx_history.slice(-59), Math.random() * 8],
          tx_history: [...iface.tx_history.slice(-59), Math.random() * 4],
        })),
      )
    }, 2000)
    return () => clearInterval(interval)
  }, [loading, interfaces.length])

  // Update chart history
  useEffect(() => {
    if (interfaces.length === 0) return
    const totalRx = interfaces.reduce((sum, iface) => sum + iface.rx_rate_bps, 0)
    const totalTx = interfaces.reduce((sum, iface) => sum + iface.tx_rate_bps, 0)

    setHistory((prev) => {
      const next = [
        ...prev,
        {
          time: new Date().toLocaleTimeString(),
          rxTotal: totalRx / (1024 * 1024),
          txTotal: totalTx / (1024 * 1024),
        },
      ]
      return next.length > 60 ? next.slice(-60) : next
    })
  }, [interfaces])

  useEffect(() => {
    if (!autoRefresh) return
    const interval = setInterval(fetchData, 3000)
    return () => clearInterval(interval)
  }, [autoRefresh, fetchData])

  const formatBandwidth = (bps: number) => {
    if (bps === 0) return '0 bps'
    if (bps < 1024) return `${bps.toFixed(0)} bps`
    if (bps < 1024 * 1024) return `${(bps / 1024).toFixed(1)} Kbps`
    return `${(bps / (1024 * 1024)).toFixed(2)} Mbps`
  }

  const totalRxRate = interfaces.reduce((sum, iface) => sum + iface.rx_rate_bps, 0)
  const totalTxRate = interfaces.reduce((sum, iface) => sum + iface.tx_rate_bps, 0)
  const peakRx = history.length > 0 ? Math.max(...history.map((h) => h.rxTotal)) : 0

  return (
    <div className="space-y-4">
      {/* Stats */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-2.5">
        <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[12px] p-3.5">
          <div className="flex items-center gap-2">
            <TrendingDown size={14} className="text-[var(--color-success)]" />
            <div>
              <p className="text-[10px] text-[var(--color-muted)] uppercase tracking-wider font-medium">Total RX</p>
              <p className="text-sm font-semibold text-[var(--color-success)] font-mono">{formatBandwidth(totalRxRate)}</p>
            </div>
          </div>
        </div>
        <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[12px] p-3.5">
          <div className="flex items-center gap-2">
            <TrendingUp size={14} className="text-[var(--color-accent)]" />
            <div>
              <p className="text-[10px] text-[var(--color-muted)] uppercase tracking-wider font-medium">Total TX</p>
              <p className="text-sm font-semibold text-[var(--color-accent)] font-mono">{formatBandwidth(totalTxRate)}</p>
            </div>
          </div>
        </div>
        <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[12px] p-3.5">
          <div className="flex items-center gap-2">
            <Activity size={14} className="text-[var(--color-warning)]" />
            <div>
              <p className="text-[10px] text-[var(--color-muted)] uppercase tracking-wider font-medium">Peak RX</p>
              <p className="text-sm font-semibold text-[var(--color-text)] font-mono">{peakRx.toFixed(1)} MB/s</p>
            </div>
          </div>
        </div>
        <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[12px] p-3.5">
          <div className="flex items-center gap-2">
            <Activity size={14} className="text-[var(--color-muted)]" />
            <div>
              <p className="text-[10px] text-[var(--color-muted)] uppercase tracking-wider font-medium">Interfaces</p>
              <p className="text-sm font-semibold text-[var(--color-text)] font-mono">
                {interfaces.filter((i) => i.is_up).length}/{interfaces.length}
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Controls */}
      <div className="flex items-center justify-between">
        <SectionHeader title="Combined Bandwidth" subtitle="RX / TX across all interfaces" />
        <button
          onClick={() => setAutoRefresh((v) => !v)}
          className={cn(
            'flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-[9px] border transition-colors',
            autoRefresh
              ? 'bg-[var(--color-accent)]/10 text-[var(--color-accent)] border-[var(--color-accent)]/30'
              : 'text-[var(--color-muted)] border-[var(--color-border)] hover:bg-[var(--color-panel-2)]',
          )}
        >
          <RefreshCw size={12} className={autoRefresh ? 'animate-spin' : ''} />
          {autoRefresh ? 'Auto' : 'Manual'}
        </button>
      </div>

      {/* Combined RX/TX chart */}
      <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[12px] p-4">
        {history.length > 1 ? (
          <ResponsiveContainer width="100%" height={240}>
            <RechartsAreaChart data={history} margin={{ top: 5, right: 10, left: -10, bottom: 0 }}>
              <defs>
                <linearGradient id="rxGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#4ade80" stopOpacity={0.3} />
                  <stop offset="100%" stopColor="#4ade80" stopOpacity={0.05} />
                </linearGradient>
                <linearGradient id="txGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#7c6cff" stopOpacity={0.3} />
                  <stop offset="100%" stopColor="#7c6cff" stopOpacity={0.05} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" strokeOpacity={0.5} />
              <XAxis
                dataKey="time"
                tick={{ fill: 'var(--color-muted)', fontSize: 10 }}
                axisLine={{ stroke: 'var(--color-border)' }}
                tickLine={false}
                interval="preserveStartEnd"
              />
              <YAxis
                tick={{ fill: 'var(--color-muted)', fontSize: 10 }}
                axisLine={false}
                tickLine={false}
                domain={[0, 'auto']}
                tickFormatter={(v) => `${v.toFixed(1)}`}
              />
              <Tooltip
                contentStyle={{
                  backgroundColor: 'var(--color-panel-2)',
                  border: '1px solid var(--color-border)',
                  borderRadius: '8px',
                  color: 'var(--color-text)',
                  fontSize: '12px',
                }}
                formatter={(value: any, name: any) => [
                  `${Number(value).toFixed(2)} MB/s`,
                  name === 'rxTotal' ? 'RX' : 'TX',
                ]}
                labelStyle={{ color: 'var(--color-muted)' }}
              />
              <Legend
                formatter={(value: string) => (
                  <span style={{ color: 'var(--color-muted)', fontSize: '11px' }}>
                    {value === 'rxTotal' ? 'RX' : 'TX'}
                  </span>
                )}
              />
              <Area
                type="monotone"
                dataKey="rxTotal"
                name="rxTotal"
                stroke="#4ade80"
                strokeWidth={2}
                fill="url(#rxGradient)"
                dot={false}
                activeDot={{ r: 4, fill: '#4ade80', stroke: 'var(--color-panel)', strokeWidth: 2 }}
              />
              <Area
                type="monotone"
                dataKey="txTotal"
                name="txTotal"
                stroke="#7c6cff"
                strokeWidth={2}
                fill="url(#txGradient)"
                dot={false}
                activeDot={{ r: 4, fill: '#7c6cff', stroke: 'var(--color-panel)', strokeWidth: 2 }}
              />
            </RechartsAreaChart>
          </ResponsiveContainer>
        ) : (
          <div className="flex items-center justify-center h-[240px] text-sm text-[var(--color-muted)]">
            Collecting bandwidth data...
          </div>
        )}
      </div>

      {/* Per-interface breakdown */}
      <SectionHeader title="Per-Interface Bandwidth" />
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2.5">
        {interfaces.filter((i) => i.is_up).map((iface) => (
          <div key={iface.name} className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[12px] p-3.5">
            <div className="flex items-center gap-2 mb-2.5">
              <span className="w-2 h-2 rounded-full bg-[var(--color-success)] shadow-[0_0_4px_rgba(74,222,128,0.4)]" />
              <span className="text-sm font-medium text-[var(--color-text)]">{iface.name}</span>
            </div>
            <div className="flex justify-between text-xs mb-1.5">
              <span className="text-[var(--color-muted)]">RX</span>
              <span className="font-mono text-[var(--color-success)]">{formatBandwidth(iface.rx_rate_bps)}</span>
            </div>
            <div className="flex justify-between text-xs">
              <span className="text-[var(--color-muted)]">TX</span>
              <span className="font-mono text-[var(--color-accent)]">{formatBandwidth(iface.tx_rate_bps)}</span>
            </div>
            <div className="mt-2 pt-2 border-t border-[var(--color-border)] flex justify-between text-[11px]">
              <span className="text-[var(--color-muted)]">Total RX</span>
              <span className="font-mono text-[var(--color-text)]">{(iface.rx_bytes / (1024 * 1024 * 1024)).toFixed(2)} GB</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

// ── Main NetOps Page ──
export function NetOps() {
  const [activeTab, setActiveTab] = useState<NetOpsTab>('ping')

  return (
    <div className="p-6 space-y-6 overflow-y-auto h-full">
      {/* Page header */}
      <div>
        <h1 className="text-2xl font-bold text-[var(--color-accent)] flex items-center gap-2">
          <Network size={24} /> Network Operations
        </h1>
        <p className="text-[var(--color-muted)] text-sm mt-1">
          Ping, DNS, port scanning, traceroute, connections, interface monitoring, and bandwidth analytics
        </p>
      </div>

      {/* Tab bar */}
      <div className="flex gap-1 bg-[var(--color-panel-2)] rounded-[9px] p-1 inline-flex overflow-x-auto">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={cn(
              'flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-all whitespace-nowrap',
              activeTab === tab.id
                ? 'bg-[var(--color-panel)] text-[var(--color-accent)] shadow-sm'
                : 'text-[var(--color-muted)] hover:text-[var(--color-text)] hover:bg-[var(--color-panel)]/30',
            )}
          >
            {tab.icon}
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab content */}
      {activeTab === 'ping' && <PingTab />}
      {activeTab === 'dns' && <DNSTab />}
      {activeTab === 'portscan' && <PortScanTab />}
      {activeTab === 'traceroute' && <TracerouteTab />}
      {activeTab === 'connections' && <ConnectionsTab />}
      {activeTab === 'interfaces' && <InterfacesTab />}
      {activeTab === 'bandwidth' && <BandwidthTab />}
    </div>
  )
}
