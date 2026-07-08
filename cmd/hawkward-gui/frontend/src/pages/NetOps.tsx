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
} from 'lucide-react'
import type {
  PingEntry,
  DNSResult,
  ConnectionInfo,
  InterfaceInfo,
} from '@/types'

type NetOpsTab = 'ping' | 'dns' | 'connections' | 'interfaces' | 'bandwidth'

const tabs: { id: NetOpsTab; label: string; icon: React.ReactNode }[] = [
  { id: 'ping', label: 'Probes', icon: <Activity size={20} /> },
  { id: 'dns', label: 'Resolution', icon: <Globe size={20} /> },
  { id: 'connections', label: 'Endpoints', icon: <Cable size={20} /> },
  { id: 'interfaces', label: 'Hardware', icon: <Wifi size={20} /> },
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
    <span className={cn('inline-block px-3 py-1 text-xs font-black uppercase tracking-widest rounded-full border shadow-sm', colorMap[status.toLowerCase()] || 'bg-muted/20 text-muted border-border')}>
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
  const [dnsResult, setDnsResult] = useState<DNSResult | null>(null)
  const [dnsLoading, setDnsLoading] = useState(false)
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- bandwidth history items are mixed-type tuples
  const [_bandwidthHistory, setBandwidthHistory] = useState<any[]>([])
  void _bandwidthHistory;
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
        if (res) {
          setInterfaces(res)
          if (activeTab === 'bandwidth') {
            const totalRx = res.reduce((acc, i) => acc + i.rx_rate_bps, 0) / 1024 / 1024
            const totalTx = res.reduce((acc, i) => acc + i.tx_rate_bps, 0) / 1024 / 1024
            setBandwidthHistory(prev => [...prev.slice(-59), { time: new Date().toLocaleTimeString(), rx: totalRx, tx: totalTx }])
          }
        }
      }
    } catch (err) { console.error(err) }
    finally { setInitialLoading(false) }
  }, [call, activeTab])

  useEffect(() => { fetchNetData(); const t = setInterval(fetchNetData, 2000); return () => clearInterval(t) }, [fetchNetData])

  const executePing = useCallback(async () => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any -- Wails bridge returns dynamic type
    const res = await call('NetOps.Ping', pingTarget, 1) as any
    if (res) {
      setPingEntries(prev => [...prev.slice(-49), {
        seq: prev.length + 1,
        ip: res.ip,
        rtt_ms: res.avg_ms || res.min_ms,
        ttl: res.ttl,
        status: res.lost > 0 ? 'timeout' : 'success'
      } as PingEntry])
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
      const res = await call('NetOps.DNSLookup', dnsHost) as DNSResult
      setDnsResult(res)
    } catch (err) {
      console.error('DNS lookup failed:', err)
      setDnsResult({ hostname: dnsHost, a: [], aaaa: [], mx: [], ns: [], cname: '', txt: [], error: String(err) })
    } finally {
      setDnsLoading(false)
    }
  }

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
    <div className="flex flex-col h-full bg-background">
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
              <MiniStat label="Reliability" value={pingEntries.length > 0 ? (100 - (pingEntries.filter(e => e.status === 'timeout').length / pingEntries.length * 100)).toFixed(1) : '100'} unit="%" icon={<ShieldCheck size={24} />} />
              <MiniStat label="Sequence" value={pingEntries.length} unit="probes" icon={<Activity size={24} />} />
              <MiniStat label="Signal" value={pingEntries.length > 0 ? pingEntries[pingEntries.length - 1].ttl || '—' : '—'} unit="ttl" icon={<Signal size={24} />} />
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
              <div className="relative group flex-1">
                <Search size={24} className="absolute left-5 top-1/2 -translate-y-1/2 text-text-faint group-focus-within:text-accent transition-colors" />
                <input
                  type="text"
                  value={dnsHost}
                  onChange={(e) => setDnsHost(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleDns()}
                  className="w-full bg-panel border border-border rounded-2xl pl-16 pr-4 py-5 text-2xl font-bold text-text placeholder-text-faint focus:outline-none focus:border-accent shadow-xl"
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
                          <td className="px-8 py-4 font-black text-accent">{c.proto}</td>
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
      </div>
    </div>
  )
}
