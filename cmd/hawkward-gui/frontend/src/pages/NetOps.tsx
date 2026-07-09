import { useState, useEffect, useCallback } from 'react'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import {
  Activity,
  Globe,
  Cable,
  Play,
  Square,
  Network,
  Timer,
  Signal,
  RefreshCw,
  Server,
  Wifi,
  ShieldCheck,
  Search,
  Info,
  ChevronRight,
  BookOpen,
  Map,
} from 'lucide-react'
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
} from 'recharts'
import type {
  PingEntry,
  DNSResult,
  ConnectionInfo,
  InterfaceInfo,
  TraceResult,
  PortResult,
} from '@/types'

type NetOpsTab = 'ping' | 'dns' | 'connections' | 'interfaces' | 'traceroute' | 'portscan' | 'bandwidth'

const tabs: { id: NetOpsTab; label: string; icon: React.ReactNode }[] = [
  { id: 'ping', label: 'Probes', icon: <Activity size={20} /> },
  { id: 'dns', label: 'Resolution', icon: <Globe size={20} /> },
  { id: 'connections', label: 'Endpoints', icon: <Cable size={20} /> },
  { id: 'interfaces', label: 'Hardware', icon: <Wifi size={20} /> },
  { id: 'traceroute', label: 'Route Trace', icon: <Map size={20} /> },
  { id: 'portscan', label: 'Port Scan', icon: <Search size={20} /> },
  { id: 'bandwidth', label: 'Traffic', icon: <Signal size={20} /> },
]

// ── Enhanced Components ──

function ProtocolReference() {
  const commonPorts = [
    { p: 22, n: 'SSH', d: 'Remote Shell (Encrypted)' },
    { p: 80, n: 'HTTP', d: 'Web Traffic (Clear)' },
    { p: 443, n: 'HTTPS', d: 'Web Traffic (SSL)' },
    { p: 3389, n: 'RDP', d: 'Windows Remote Desktop' },
    { p: 53, n: 'DNS', d: 'Domain Resolution' },
  ]
  return (
    <div className="bg-panel border border-border rounded-[24px] p-8 shadow-xl">
      <div className="flex items-center gap-4 mb-6">
        <BookOpen size={24} className="text-accent" />
        <h3 className="text-xl font-black text-text uppercase tracking-widest">Protocol Intel</h3>
      </div>
      <div className="space-y-4">
        {commonPorts.map(port => (
          <div key={port.p} className="flex items-center justify-between p-3 bg-panel-3 rounded-xl border border-border group hover:border-accent/30 transition-all">
            <div className="flex items-center gap-4">
              <span className="w-12 text-lg font-black text-accent">{port.p}</span>
              <div className="flex flex-col">
                <span className="text-sm font-black text-text uppercase tracking-tighter">{port.n}</span>
                <span className="text-xs text-text-faint font-medium">{port.d}</span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

function SectionBriefing({ title, objective, checklist }: { title: string, objective: string, checklist: string[] }) {
  return (
    <div className="bg-panel-2 border border-border rounded-[24px] p-8 shadow-xl mb-8">
      <div className="flex items-center gap-4 mb-4">
        <Info size={24} className="text-accent" />
        <h3 className="text-2xl font-black text-text uppercase tracking-widest">{title}</h3>
      </div>
      <p className="text-lg text-text-dim leading-relaxed mb-6 italic">{objective}</p>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {checklist.map((item, i) => (
          <div key={i} className="flex items-center gap-3">
            <div className="w-1.5 h-1.5 rounded-full bg-accent shadow-[0_0_6px_var(--color-accent)]" />
            <span className="text-sm font-bold text-text-faint">{item}</span>
          </div>
        ))}
      </div>
    </div>
  )
}

function StatusBadge({ status }: { status: string }) {
  const colorMap: Record<string, string> = {
    success: 'bg-success/15 text-success border-success/30',
    timeout: 'bg-danger/15 text-danger border-danger/30',
    listening: 'bg-success/15 text-success border-success/30',
    established: 'bg-accent/15 text-accent border-accent/30',
  }
  return (
    <span className={cn('inline-block px-3 py-1 text-xs font-black uppercase tracking-widest rounded-full border shadow-sm', colorMap[status.toLowerCase()] || 'bg-text-faint/20 text-text-faint border-border')}>
      {status.replace('_', ' ')}
    </span>
  )
}

function MiniStat({ label, value, icon, unit }: { label: string; value: string | number; icon?: React.ReactNode; unit?: string }) {
  return (
    <div className="bg-panel border border-border rounded-2xl p-6 flex items-center gap-6 shadow-lg transition-all hover:scale-105 active:scale-95 cursor-default group">
      <div className="w-14 h-14 rounded-2xl bg-panel-3 flex items-center justify-center text-accent border border-border shadow-inner group-hover:bg-accent-soft group-hover:text-white transition-all">
        {icon}
      </div>
      <div>
        <p className="text-sm font-bold text-text-faint uppercase tracking-widest mb-1">{label}</p>
        <p className="text-3xl font-black text-text tabular-nums leading-none">
          {value}{unit && <span className="text-base text-text-faint ml-1 font-medium">{unit}</span>}
        </p>
      </div>
    </div>
  )
}

// ── Main Page ──
export function NetOps() {
  const { call } = useBackend()
  const [activeTab, setActiveTab] = useState<NetOpsTab>('ping')
  const [pingTarget, setPingTarget] = useState('8.8.8.8')
  const [pingRunning, setPingRunning] = useState(false)
  const [pingEntries, setPingEntries] = useState<PingEntry[]>([])
  const [connections, setConnections] = useState<ConnectionInfo[]>([])
  const [interfaces, setInterfaces] = useState<InterfaceInfo[]>([])
  const [dnsHost, setDnsHost] = useState('google.com')
  const [dnsServer, setDnsServer] = useState('')
  const [dnsResult, setDnsResult] = useState<DNSResult | null>(null)
  const [dnsLoading, setDnsLoading] = useState(false)
  const [traceTarget, setTraceTarget] = useState('8.8.8.8')
  const [traceResult, setTraceResult] = useState<TraceResult | null>(null)
  const [traceRunning, setTraceRunning] = useState(false)
  const [portScanTarget, setPortScanTarget] = useState('127.0.0.1')
  const [portScanPorts, setPortScanPorts] = useState('21,22,23,25,53,80,110,143,443,445,993,1433,1521,3306,3389,5432,6379,8080,8443,27017')
  const [portScanResults, setPortScanResults] = useState<PortResult[]>([])
  const [portScanLoading, setPortScanLoading] = useState(false)
  const [initialLoading, setInitialLoading] = useState(true)

  // Data Fetching
  const fetchNetData = useCallback(async () => {
    try {
      if (activeTab === 'connections') {
        const res = await call('NetOps.GetConnections')
        setConnections((res as ConnectionInfo[]) || [])
      }
      if (activeTab === 'interfaces' || activeTab === 'bandwidth') {
        const res = await call('NetOps.GetInterfaces') as InterfaceInfo[]
        if (res) { setInterfaces(res) }
      }
    } catch (err) { console.error(err) }
    finally { setInitialLoading(false) }
  }, [call, activeTab])

  useEffect(() => { fetchNetData(); const t = setInterval(fetchNetData, 2000); return () => clearInterval(t) }, [fetchNetData])

  const executePing = useCallback(async () => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any -- Wails bridge returns dynamic type
    const res = await call('NetOps.Ping', pingTarget, 1) as any
    if (res?.error) {
      setPingEntries(prev => [...prev.slice(-49), {
        seq: prev.length + 1,
        ip: pingTarget,
        rtt_ms: null,
        jitter_ms: null,
        ttl: null,
        status: 'timeout'
      } as PingEntry])
    } else if (res) {
      setPingEntries(prev => {
        const lastEntry = prev[prev.length - 1]
        let currentJitter = 0
        if (lastEntry && lastEntry.rtt_ms !== null && res.avg_ms !== undefined) {
          currentJitter = Math.abs(res.avg_ms - lastEntry.rtt_ms)
        }
        return [...prev.slice(-49), {
          seq: prev.length + 1,
          ip: res.ip,
          rtt_ms: res.avg_ms || res.min_ms,
          jitter_ms: currentJitter,
          ttl: res.ttl,
          status: res.lost > 0 ? 'timeout' : 'success'
        } as PingEntry]
      })
    }
  }, [call, pingTarget])

  useEffect(() => {
    if (pingRunning) {
      const t = setInterval(executePing, 1000)
      return () => clearInterval(t)
    }
  }, [pingRunning, executePing])

  const handleDns = async () => {
    setDnsLoading(true); setDnsResult(null)
    try {
      const res = await call('NetOps.DNSLookup', dnsHost, dnsServer) as DNSResult
      setDnsResult(res)
    } catch (err) {
      console.error('DNS lookup failed:', err)
      setDnsResult({ hostname: dnsHost, a: [], aaaa: [], mx: [], ns: [], cname: '', txt: [], error: String(err) })
    } finally {
      setDnsLoading(false)
    }
  }

  const executeTrace = useCallback(async () => {
    setTraceRunning(true); setTraceResult(null)
    try {
      const res = await call('NetOps.Traceroute', traceTarget) as TraceResult
      setTraceResult(res)
    } catch (err) {
      console.error('Traceroute failed:', err)
      setTraceResult({ target: traceTarget, hops: [], error: String(err) })
    } finally {
      setTraceRunning(false)
    }
  }, [call, traceTarget])

  const handlePortScan = useCallback(async () => {
    setPortScanLoading(true); setPortScanResults([])
    try {
      const ports = portScanPorts.split(',').map(p => parseInt(p.trim(), 10)).filter(p => !isNaN(p))
      const res = await call('NetOps.PortScan', portScanTarget, ports) as PortResult[]
      setPortScanResults(res || [])
    } catch (err) {
      console.error('Port scan failed:', err)
    } finally {
      setPortScanLoading(false)
    }
  }, [call, portScanTarget, portScanPorts])

  if (initialLoading) {
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
          <h1 className="text-3xl font-black text-text flex items-center gap-4">
            <Network size={32} className="text-accent" /> NETWORK OPERATIONS
          </h1>
          <p className="text-text-dim text-lg mt-2">Fabric probes, resolver triage, and cumulative traffic heuristics.</p>
        </div>
        <div className="flex gap-1 bg-panel border border-border rounded-2xl p-1.5 shadow-inner overflow-x-auto max-w-[800px]">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={cn(
                'flex items-center gap-3 px-8 py-3 rounded-xl text-lg font-bold transition-all whitespace-nowrap',
                activeTab === tab.id ? 'bg-accent text-white shadow-lg' : 'text-text-dim hover:text-text hover:bg-white/5',
              )}
            >
              {tab.icon}
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-10 space-y-12">
        {activeTab === 'ping' && (
          <div className="space-y-8 animate-in fade-in duration-500">
            <SectionBriefing
              title="ICMP Probe Analysis"
              objective="Measure Round-Trip Time (RTT) to determine routing stability. Latency spikes correlate with packet-shaping or bottlenecked gateways."
              checklist={[
                "Ideal RTT: < 50ms for low-latency nodes.",
                "Jitter: Monitor for inconsistent response times.",
                "TTL: Verify hop-count to identify path changes.",
                "Packet Loss: 0% is the target for stable links."
              ]}
            />
            <div className="flex items-center gap-6 bg-panel-2 border border-border p-6 rounded-[24px] shadow-inner">
              <div className="relative group flex-1">
                <Globe size={24} className="absolute left-5 top-1/2 -translate-y-1/2 text-text-faint group-focus-within:text-accent transition-colors" />
                <input
                  type="text"
                  value={pingTarget}
                  onChange={(e) => setPingTarget(e.target.value)}
                  className="w-full bg-panel border border-border rounded-2xl pl-16 pr-4 py-5 text-2xl font-bold text-text placeholder-text-faint focus:outline-none focus:border-accent shadow-xl"
                />
              </div>
              <button onClick={() => setPingRunning(!pingRunning)} className={cn("flex items-center gap-3 px-10 py-5 text-xl font-black rounded-2xl transition-all shadow-xl", pingRunning ? "bg-danger text-white hover:bg-danger/90" : "bg-accent text-white hover:bg-accent/90")}>
                {pingRunning ? <Square size={24} fill="currentColor" /> : <Play size={24} fill="currentColor" />}
                {pingRunning ? 'STOP PROBE' : 'START PROBE'}
              </button>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
              <MiniStat label="Latency" value={pingEntries.length > 0 ? pingEntries[pingEntries.length - 1].rtt_ms?.toFixed(1) || '—' : '—'} unit="ms" icon={<Timer size={24} />} />
              <MiniStat
                label="Jitter"
                value={pingEntries.length > 1 ? (pingEntries.slice(-10).reduce((acc, curr, i, arr) => {
                  if (i === 0 || curr.rtt_ms === null || arr[i - 1].rtt_ms === null) return acc
                  return acc + Math.abs(curr.rtt_ms - arr[i - 1].rtt_ms!)
                }, 0) / (pingEntries.slice(-10).filter(e => e.rtt_ms !== null).length - 1 || 1)).toFixed(2) : '0.00'}
                unit="ms"
                icon={<Activity size={24} />}
              />
              <MiniStat label="Reliability" value={pingEntries.length > 0 ? (100 - (pingEntries.filter(e => e.status === 'timeout').length / pingEntries.length * 100)).toFixed(1) : '100'} unit="%" icon={<ShieldCheck size={24} />} />
              <MiniStat label="Signal" value={pingEntries.length > 0 ? pingEntries[pingEntries.length - 1].ttl || '—' : '—'} unit="ttl" icon={<Signal size={24} />} />
            </div>

            {/* Latency Chart */}
            <div className="bg-panel border border-border rounded-[24px] p-8 shadow-xl">
              <div className="flex items-center justify-between mb-8">
                <h3 className="text-xl font-black text-text uppercase tracking-widest flex items-center gap-3">
                  <Activity size={20} className="text-accent" /> Latency History
                </h3>
                <div className="flex items-center gap-4 text-xs font-bold text-text-faint">
                  <span className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded-full bg-accent" /> RTT (ms)
                  </span>
                </div>
              </div>
              <div className="h-[300px] w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={pingEntries.map(e => ({ seq: e.seq, rtt: e.rtt_ms || 0 }))}>
                    <defs>
                      <linearGradient id="colorRtt" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="var(--color-accent)" stopOpacity={0.3} />
                        <stop offset="95%" stopColor="var(--color-accent)" stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.05)" vertical={false} />
                    <XAxis dataKey="seq" hide />
                    <YAxis
                      stroke="rgba(255,255,255,0.3)"
                      fontSize={12}
                      tickLine={false}
                      axisLine={false}
                      tickFormatter={(v) => `${v}ms`}
                    />
                    <Tooltip
                      contentStyle={{ backgroundColor: 'rgba(0,0,0,0.8)', border: '1px solid rgba(255,255,255,0.1)', borderRadius: '12px' }}
                      itemStyle={{ color: 'var(--color-accent)' }}
                      labelStyle={{ display: 'none' }}
                    />
                    <Area
                      type="monotone"
                      dataKey="rtt"
                      stroke="var(--color-accent)"
                      strokeWidth={3}
                      fillOpacity={1}
                      fill="url(#colorRtt)"
                      animationDuration={300}
                    />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'dns' && (
          <div className="space-y-8 animate-in fade-in duration-500">
            <SectionBriefing
              title="Domain Resolution"
              objective="Identify local cache poisoning or upstream resolver failures. Slow DNS lookup times directly impact application perceived performance."
              checklist={[
                "A-Records: Verify IPv4 host identity.",
                "AAAA-Records: Check for modern IPv6 support.",
                "NS-Records: Audit authoritative nameservers.",
                "MX-Records: Confirm mail routing topology."
              ]}
            />
            <div className="flex items-center gap-6 bg-panel-2 border border-border p-6 rounded-[24px] shadow-inner">
              <div className="relative group flex-[2]">
                <Search size={24} className="absolute left-5 top-1/2 -translate-y-1/2 text-text-faint group-focus-within:text-accent transition-colors" />
                <input
                  type="text"
                  value={dnsHost}
                  onChange={(e) => setDnsHost(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleDns()}
                  className="w-full bg-panel border border-border rounded-2xl pl-16 pr-4 py-5 text-2xl font-bold text-text placeholder-text-faint focus:outline-none focus:border-accent shadow-xl"
                  placeholder="Hostname (e.g. google.com)"
                />
              </div>
              <div className="relative group flex-1">
                <Server size={24} className="absolute left-5 top-1/2 -translate-y-1/2 text-text-faint group-focus-within:text-accent transition-colors" />
                <input
                  type="text"
                  value={dnsServer}
                  onChange={(e) => setDnsServer(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleDns()}
                  className="w-full bg-panel border border-border rounded-2xl pl-16 pr-4 py-5 text-lg font-bold text-text placeholder-text-faint focus:outline-none focus:border-accent shadow-xl"
                  placeholder="Resolver (e.g. 8.8.8.8)"
                />
              </div>
              <button onClick={handleDns} disabled={dnsLoading} className="flex items-center gap-3 px-10 py-5 bg-accent text-white text-xl font-black rounded-2xl hover:bg-accent/90 shadow-xl transition-all">
                {dnsLoading ? <RefreshCw size={24} className="animate-spin" /> : <Search size={24} />}
                {dnsLoading ? 'RESOLVING...' : 'RESOLVE'}
              </button>
            </div>

            {/* Loading skeleton */}
            {dnsLoading && (
              <div className="bg-panel border border-border rounded-[24px] p-10 shadow-xl animate-pulse">
                <div className="h-6 bg-panel-3 rounded w-1/3 mb-6" />
                <div className="space-y-4">
                  <div className="h-4 bg-panel-3 rounded w-3/4" />
                  <div className="h-4 bg-panel-3 rounded w-1/2" />
                  <div className="h-4 bg-panel-3 rounded w-2/3" />
                  <div className="h-4 bg-panel-3 rounded w-1/2" />
                </div>
              </div>
            )}

            {/* DNS Results */}
            {dnsResult && !dnsLoading && (
              <div className="bg-panel border border-border rounded-[24px] p-10 shadow-xl">
                <div className="flex items-center gap-4 mb-8">
                  <Globe size={28} className={dnsResult.error ? 'text-danger' : 'text-success'} />
                  <h3 className="text-2xl font-black text-text uppercase tracking-tight">
                    {dnsResult.hostname}
                  </h3>
                  {dnsResult.error && (
                    <span className="px-4 py-1 text-sm font-black text-danger bg-danger/10 rounded-full border border-danger/30 uppercase tracking-widest">
                      Failed
                    </span>
                  )}
                </div>

                {dnsResult.error ? (
                  <div className="bg-danger/5 border border-danger/20 rounded-2xl p-6">
                    <p className="text-lg font-bold text-danger">{dnsResult.error}</p>
                  </div>
                ) : (
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    {[
                      { label: 'A Records', values: dnsResult.a, icon: <Globe size={18} /> },
                      { label: 'AAAA Records', values: dnsResult.aaaa, icon: <Globe size={18} /> },
                      { label: 'MX Records', values: dnsResult.mx, icon: <Server size={18} /> },
                      { label: 'NS Records', values: dnsResult.ns, icon: <Server size={18} /> },
                      { label: 'TXT Records', values: dnsResult.txt, icon: <Info size={18} /> },
                      { label: 'CNAME', values: dnsResult.cname ? [dnsResult.cname] : [], icon: <ChevronRight size={18} /> },
                    ].map(section => (
                      <div key={section.label} className="bg-panel-2 border border-border rounded-2xl p-6">
                        <div className="flex items-center gap-2 mb-4">
                          <span className="text-accent">{section.icon}</span>
                          <span className="text-sm font-black text-text-dim uppercase tracking-widest">{section.label}</span>
                        </div>
                        {section.values.length > 0 ? (
                          <div className="space-y-2">
                            {section.values.map((v, i) => (
                              <div key={i} className="px-4 py-2 bg-panel-3 border border-border rounded-xl text-lg font-bold text-text tabular-nums font-[JetBrains_Mono]">
                                {v}
                              </div>
                            ))}
                          </div>
                        ) : (
                          <p className="text-base text-text-faint italic">No records found</p>
                        )}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}
          </div>
        )}

        {activeTab === 'connections' && (
          <div className="grid grid-cols-1 lg:grid-cols-4 gap-12 animate-in fade-in duration-500">
            <div className="lg:col-span-1 space-y-8">
              <SectionBriefing
                title="Endpoint Matrix"
                objective="Audit established sockets to detect unauthorized data exfiltration or active C2 channels."
                checklist={[
                  "ESTABLISHED: Active traffic session.",
                  "LISTENING: Local port waiting for input.",
                  "TIME_WAIT: Socket closing sequence.",
                  "CLOSE_WAIT: Pending remote termination."
                ]}
              />
              <ProtocolReference />
            </div>
            <div className="lg:col-span-3">
              <div className="bg-panel border border-border rounded-[28px] overflow-hidden shadow-2xl">
                <div className="max-h-[800px] overflow-y-auto">
                  <table className="w-full text-left border-collapse">
                    <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border">
                      <tr>
                        <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest">Protocol</th>
                        <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest">Endpoint Node</th>
                        <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest">Process Origin</th>
                        <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest text-right">State</th>
                      </tr>
                    </thead>
                    <tbody>
                      {connections.map((c, i) => (
                        <tr key={i} className="border-b border-border/20 hover:bg-white/5 transition-all group">
                          <td className="px-8 py-4 font-black text-accent">{c.state}</td>
                          <td className="px-8 py-4">
                            <div className="flex flex-col">
                              <span className="text-lg font-black text-text">{c.remote_addr}:{c.remote_port}</span>
                              <span className="text-sm font-bold text-text-faint uppercase tabular-nums">LOCAL: {c.local_addr}:{c.local_port}</span>
                            </div>
                          </td>
                          <td className="px-8 py-4">
                            <div className="flex flex-col">
                              <span className="text-xl font-black text-text">{c.process_name || 'System Core'}</span>
                              <span className="text-xs font-bold text-text-faint uppercase tracking-widest">PID: {c.pid}</span>
                            </div>
                          </td>
                          <td className="px-8 py-4 text-right"><StatusBadge status={c.state} /></td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'interfaces' && (
          <div className="space-y-8 animate-in fade-in duration-500">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
              {interfaces.map(iface => (
                <div key={iface.name} className="bg-panel border border-border rounded-[24px] p-10 shadow-xl relative overflow-hidden group">
                  <div className="absolute top-0 right-0 w-32 h-32 bg-accent/5 rounded-bl-full pointer-events-none" />
                  <div className="flex items-center gap-6 mb-8">
                    <div className={cn("w-16 h-16 rounded-2xl flex items-center justify-center border shadow-inner transition-all", iface.is_up ? "bg-success/10 border-success/30 text-success" : "bg-danger/10 border-danger/30 text-danger")}>
                      <Wifi size={32} />
                    </div>
                    <div>
                      <h3 className="text-3xl font-black text-text uppercase tracking-tighter">{iface.name}</h3>
                      <div className="flex items-center gap-2 mt-1">
                        <span className={cn("w-2 h-2 rounded-full", iface.is_up ? "bg-success shadow-[0_0_8px_var(--color-success)]" : "bg-danger")} />
                        <span className="text-sm font-black text-text-faint uppercase tracking-widest">{iface.is_up ? 'ACTIVE NODE' : 'DISCONNECTED'}</span>
                      </div>
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-8">
                    <div>
                      <p className="text-xs font-black text-text-faint uppercase mb-2">Physical MAC</p>
                      <p className="text-xl font-bold text-text tabular-nums">{iface.mac}</p>
                    </div>
                    <div>
                      <p className="text-xs font-black text-text-faint uppercase mb-2">Link Capacity</p>
                      <p className="text-xl font-bold text-accent">{iface.speed}</p>
                    </div>
                  </div>
                  <div className="mt-8 pt-8 border-t border-border flex items-center gap-4 flex-wrap">
                    {iface.ips.map((ip, idx) => (
                      <div key={idx} className="px-4 py-1.5 bg-panel-3 border border-border rounded-full text-sm font-black text-accent tabular-nums flex items-center gap-2">
                        <Globe size={14} /> {ip}
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {activeTab === 'traceroute' && (
          <div className="space-y-8 animate-in fade-in duration-500">
            <SectionBriefing
              title="Route Trace Analysis"
              objective="Map the network path to identify routing loops, high-latency hops, or packet loss between the local node and the target."
              checklist={[
                "Each hop represents a router along the path.",
                "High RTT at a single hop may indicate congestion.",
                "Multiple timeouts in sequence suggests a routing blackhole.",
                "Consistent latency gradient across hops is expected."
              ]}
            />

            {/* Input + Execute */}
            <div className="flex items-center gap-6 bg-panel-2 border border-border p-6 rounded-[24px] shadow-inner">
              <div className="relative group flex-1">
                <Globe size={24} className="absolute left-5 top-1/2 -translate-y-1/2 text-text-faint group-focus-within:text-accent transition-colors" />
                <input
                  type="text"
                  value={traceTarget}
                  onChange={(e) => setTraceTarget(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && executeTrace()}
                  placeholder="Target hostname or IP"
                  className="w-full bg-panel border border-border rounded-2xl pl-16 pr-4 py-5 text-2xl font-bold text-text placeholder-text-faint focus:outline-none focus:border-accent shadow-xl"
                />
              </div>
              <button onClick={executeTrace} disabled={traceRunning} className="flex items-center gap-3 px-10 py-5 bg-accent text-white text-xl font-black rounded-2xl hover:bg-accent/90 shadow-xl transition-all disabled:opacity-50">
                {traceRunning ? <RefreshCw size={24} className="animate-spin" /> : <Map size={24} />}
                {traceRunning ? 'TRACING...' : 'TRACE ROUTE'}
              </button>
            </div>

            {/* Results Table */}
            {traceResult && !traceRunning && (
              <div className="bg-panel border border-border rounded-[28px] overflow-hidden shadow-2xl">
                <div className="px-8 py-6 bg-panel-2 border-b border-border flex items-center justify-between">
                  <h3 className="text-xl font-black text-text uppercase tracking-widest flex items-center gap-3">
                    <Globe size={20} className="text-accent" /> {traceResult.target}
                  </h3>
                  {traceResult.error && (
                    <span className="px-4 py-1.5 text-sm font-black text-danger bg-danger/10 rounded-full border border-danger/30 uppercase tracking-widest">
                      FAILED
                    </span>
                  )}
                </div>
                <div className="overflow-x-auto">
                  <table className="w-full text-left border-collapse">
                    <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border">
                      <tr>
                        <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest w-24">Hop</th>
                        <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest">Host / IP</th>
                        <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest text-right">RTT (ms)</th>
                        <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest text-right w-28">Status</th>
                      </tr>
                    </thead>
                    <tbody>
                      {traceResult.hops.map((hop) => (
                        <tr key={hop.number} className="border-b border-border/20 hover:bg-white/5 transition-all group">
                          <td className="px-8 py-5 text-2xl font-black text-accent tabular-nums">{hop.number}</td>
                          <td className="px-8 py-5">
                            <div className="flex flex-col">
                              <span className="text-xl font-black text-text">{hop.host || 'Unknown'}</span>
                              <span className="text-sm font-bold text-text-faint uppercase">{hop.ip || '—'}</span>
                            </div>
                          </td>
                          <td className="px-8 py-5 text-right">
                            <span className="text-xl font-black text-text tabular-nums">
                              {hop.rtts_ms.length > 0 ? hop.rtts_ms.join(', ') : '—'}
                            </span>
                          </td>
                          <td className="px-8 py-5 text-right">
                            <span className={cn(
                              "px-4 py-1.5 rounded-full text-sm font-black uppercase tracking-widest",
                              hop.timed ? "bg-warning/10 text-warning border border-warning/30" : "bg-success/10 text-success border border-success/30"
                            )}>
                              {hop.timed ? 'TIMED' : 'REACHED'}
                            </span>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}
          </div>
        )}

        {activeTab === 'portscan' && (
          <div className="space-y-8 animate-in fade-in duration-500">
            <SectionBriefing
              title="Port Scan"
              objective="Scan a target host for open ports to identify running services, exposed attack surfaces, or unauthorized listeners."
              checklist={[
                "Open ports indicate services accepting connections.",
                "Common ports (21,22,23,25,53,80,110,143,443,445,993,1433,1521,3306,3389,5432,6379,8080,8443,27017) are pre-configured.",
                "Unrecognized open ports may signal unauthorized services or malware.",
                "Closed or filtered ports are typically invisible to scanners."
              ]}
            />

            {/* Input + Scan Button */}
            <div className="flex items-center gap-6 bg-panel-2 border border-border p-6 rounded-[24px] shadow-inner">
              <div className="relative group flex-1">
                <Globe size={24} className="absolute left-5 top-1/2 -translate-y-1/2 text-text-faint group-focus-within:text-accent transition-colors" />
                <input
                  type="text"
                  value={portScanTarget}
                  onChange={(e) => setPortScanTarget(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handlePortScan()}
                  placeholder="Target hostname or IP"
                  className="w-full bg-panel border border-border rounded-2xl pl-16 pr-4 py-5 text-2xl font-bold text-text placeholder-text-faint focus:outline-none focus:border-accent shadow-xl"
                />
              </div>
              <div className="relative group w-96">
                <Network size={24} className="absolute left-5 top-1/2 -translate-y-1/2 text-text-faint group-focus-within:text-accent transition-colors" />
                <input
                  type="text"
                  value={portScanPorts}
                  onChange={(e) => setPortScanPorts(e.target.value)}
                  placeholder="Ports (e.g. 80,443,8080 or comma-separated)"
                  className="w-full bg-panel border border-border rounded-2xl pl-16 pr-4 py-5 text-lg font-bold text-text placeholder-text-faint focus:outline-none focus:border-accent shadow-xl"
                />
              </div>
              <button onClick={handlePortScan} disabled={portScanLoading} className="flex items-center gap-3 px-10 py-5 bg-accent text-white text-xl font-black rounded-2xl hover:bg-accent/90 shadow-xl transition-all disabled:opacity-50">
                {portScanLoading ? <RefreshCw size={24} className="animate-spin" /> : <Search size={24} />}
                {portScanLoading ? 'SCANNING...' : 'SCAN'}
              </button>
            </div>

            {/* Results Table */}
            {portScanResults.length > 0 && !portScanLoading && (
              <div className="bg-panel border border-border rounded-[28px] overflow-hidden shadow-2xl">
                <div className="px-8 py-6 bg-panel-2 border-b border-border flex items-center justify-between">
                  <h3 className="text-xl font-black text-text uppercase tracking-widest flex items-center gap-3">
                    <Globe size={20} className="text-accent" /> {portScanTarget}
                  </h3>
                  <span className="px-4 py-1.5 text-sm font-black text-text-dim bg-panel-3 rounded-full border border-border/30 uppercase tracking-widest">
                    {portScanResults.filter(p => p.open).length}/{portScanResults.length} OPEN
                  </span>
                </div>
                <div className="overflow-x-auto">
                  <table className="w-full text-left border-collapse">
                    <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border">
                      <tr>
                        <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest">Port</th>
                        <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest">Service</th>
                        <th className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest text-right">Status</th>
                      </tr>
                    </thead>
                    <tbody>
                      {portScanResults.map((p, i) => (
                        <tr key={i} className="border-b border-border/20 hover:bg-white/5 transition-all group">
                          <td className="px-8 py-5">
                            <span className="text-2xl font-black text-accent tabular-nums">{p.port}</span>
                          </td>
                          <td className="px-8 py-5">
                            <span className="text-xl font-black text-text">{p.service || 'Unknown'}</span>
                          </td>
                          <td className="px-8 py-5 text-right">
                            <span className={cn(
                              "px-4 py-1.5 rounded-full text-sm font-black uppercase tracking-widest",
                              p.open
                                ? "bg-success/10 text-success border border-success/30"
                                : "bg-danger/10 text-danger border border-danger/30"
                            )}>
                              {p.open ? 'OPEN' : 'CLOSED'}
                            </span>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}

            {/* Empty state */}
            {portScanResults.length === 0 && !portScanLoading && (
              <div className="bg-panel border border-border rounded-[24px] p-12 shadow-xl text-center">
                <Search size={48} className="mx-auto mb-4 text-text-faint" />
                <p className="text-xl font-bold text-text-dim">
                  Enter a target host and click SCAN to begin.
                </p>
              </div>
            )}
          </div>
        )}

        {activeTab === 'bandwidth' && (
          <div className="space-y-8 animate-in fade-in duration-500">
            <SectionBriefing
              title="Traffic Analysis"
              objective="Monitor bandwidth utilization across all network interfaces. Track real-time throughput and historical traffic patterns to identify congestion and plan capacity."
              checklist={[
                "Combined throughput shows aggregate bandwidth demand across all links.",
                "RX/TX balance indicates asymmetric link usage or duplex mismatches.",
                "Sustained saturation suggests capacity planning is needed.",
                "Spike patterns may correlate with large transfers or anomalies."
              ]}
            />

            {interfaces.length === 0 ? (
              <div className="bg-panel border border-border rounded-[24px] p-12 shadow-xl text-center">
                <Signal size={48} className="mx-auto mb-4 text-text-faint" />
                <p className="text-xl font-bold text-text-dim">
                  No interface data available. Visit the Hardware tab first.
                </p>
              </div>
            ) : (
              <>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  <MiniStat
                    label="Total Interfaces"
                    value={interfaces.length}
                    icon={<Wifi size={24} />}
                  />
                  <MiniStat
                    label="Combined RX"
                    value={(interfaces.reduce((sum, i) => sum + i.rx_rate_bps, 0) / 1_000_000).toFixed(2)}
                    unit="Mbps"
                    icon={<Signal size={24} />}
                  />
                  <MiniStat
                    label="Combined TX"
                    value={(interfaces.reduce((sum, i) => sum + i.tx_rate_bps, 0) / 1_000_000).toFixed(2)}
                    unit="Mbps"
                    icon={<Signal size={24} />}
                  />
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                  {interfaces.map((iface) => {
                    const hasHistory = iface.rx_history?.length > 0
                    const chartData = hasHistory
                      ? iface.rx_history.map((_, idx) => ({
                        point: idx,
                        rx: +(iface.rx_history[idx] / 1_000_000).toFixed(4),
                        tx: +(iface.tx_history[idx] / 1_000_000).toFixed(4),
                      }))
                      : []
                    const gradId = `bw-${iface.name.replace(/[^a-zA-Z0-9]/g, '')}`

                    return (
                      <div key={iface.name} className="bg-panel border border-border rounded-[24px] p-6 md:p-8 shadow-xl">
                        <div className="flex items-center justify-between mb-6 flex-wrap gap-3">
                          <h3 className="text-lg font-black text-text uppercase tracking-wider">{iface.name}</h3>
                          <div className="flex items-center gap-4 text-sm font-bold tabular-nums">
                            <span className="flex items-center gap-1.5">
                              <span className="w-2.5 h-2.5 rounded-full bg-[#14b8a6]" />
                              <span className="text-text-dim">RX <span className="text-text">{(iface.rx_rate_bps / 1_000_000).toFixed(2)}</span> Mbps</span>
                            </span>
                            <span className="flex items-center gap-1.5">
                              <span className="w-2.5 h-2.5 rounded-full bg-[#3b82f6]" />
                              <span className="text-text-dim">TX <span className="text-text">{(iface.tx_rate_bps / 1_000_000).toFixed(2)}</span> Mbps</span>
                            </span>
                          </div>
                        </div>
                        {hasHistory ? (
                          <ResponsiveContainer width="100%" height={200}>
                            <AreaChart data={chartData}>
                              <defs>
                                <linearGradient id={`${gradId}-rx`} x1="0" y1="0" x2="0" y2="1">
                                  <stop offset="5%" stopColor="#14b8a6" stopOpacity={0.35} />
                                  <stop offset="95%" stopColor="#14b8a6" stopOpacity={0} />
                                </linearGradient>
                                <linearGradient id={`${gradId}-tx`} x1="0" y1="0" x2="0" y2="1">
                                  <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.35} />
                                  <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                                </linearGradient>
                              </defs>
                              <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.04)" />
                              <XAxis dataKey="point" tick={false} axisLine={false} />
                              <YAxis tick={{ fill: 'rgba(255,255,255,0.3)', fontSize: 11 }} tickFormatter={(v) => `${v}`} width={40} axisLine={false} />
                              <Tooltip
                                contentStyle={{
                                  background: 'rgba(0,0,0,0.85)',
                                  border: '1px solid rgba(255,255,255,0.1)',
                                  borderRadius: '12px',
                                  backdropFilter: 'blur(8px)',
                                }}
                                labelStyle={{ display: 'none' }}
                                formatter={(value: any) => [`${Number(value).toFixed(2)} Mbps`]}
                              />
                              <Area type="monotone" dataKey="rx" stackId="1" stroke="#14b8a6" fill={`url(#${gradId}-rx)`} strokeWidth={2} dot={false} />
                              <Area type="monotone" dataKey="tx" stackId="1" stroke="#3b82f6" fill={`url(#${gradId}-tx)`} strokeWidth={2} dot={false} />
                            </AreaChart>
                          </ResponsiveContainer>
                        ) : (
                          <div className="h-[200px] flex items-center justify-center">
                            <p className="text-sm font-bold text-text-faint">No traffic history available</p>
                          </div>
                        )}
                      </div>
                    )
                  })}
                </div>
              </>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
