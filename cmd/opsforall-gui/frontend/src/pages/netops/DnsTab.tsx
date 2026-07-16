import { useState } from 'react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import {
  Globe,
  Search,
  Server,
  RefreshCw,
  Info,
  ChevronRight,
} from 'lucide-react'
import type { DNSResult } from '@/types'
import { SectionBriefing } from '@/components/ui/SectionBriefing'

export function DnsTab() {
  const { call } = useBackend()
  const { dnsTimeout } = useSettingsStore()
  const [dnsHost, setDnsHost] = useState('google.com')
  const [dnsServer, setDnsServer] = useState('')
  const [dnsResult, setDnsResult] = useState<DNSResult | null>(null)
  const [dnsLoading, setDnsLoading] = useState(false)

  const handleDns = async () => {
    setDnsLoading(true); setDnsResult(null)
    try {
      const res = await call('NetOps.DNSLookup', dnsHost, dnsServer, dnsTimeout) as DNSResult
      setDnsResult(res)
    } catch (err: unknown) {
      console.error('DNS lookup failed:', err)
      setDnsResult({ hostname: dnsHost, a: [], aaaa: [], mx: [], ns: [], cname: '', txt: [], error: String(err) })
    } finally {
      setDnsLoading(false)
    }
  }

  return (
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
      <div className="flex items-center gap-6 bg-panel-2 border border-border p-6 rounded-[var(--radius-lg)] shadow-inner">
        <div className="relative group flex-[2]">
          <Search size={24} className="absolute left-5 top-1/2 -translate-y-1/2 text-text-faint group-focus-within:text-accent transition-colors" />
          <input
            type="text"
            value={dnsHost}
            onChange={(e) => setDnsHost(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleDns()}
            className="w-full bg-panel border border-border rounded-2xl pl-16 pr-4 py-3 text-sm font-medium text-[var(--color-text)] placeholder-[var(--color-text-faint)] focus:outline-none focus:border-accent shadow-xl"
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
        <button onClick={handleDns} disabled={dnsLoading} className="flex items-center gap-3 px-5 py-2.5 bg-[var(--color-accent)] text-white text-sm font-semibold rounded-xl hover:bg-accent/90 shadow-xl transition-all">
          {dnsLoading ? <RefreshCw size={24} className="animate-spin" /> : <Search size={24} />}
          {dnsLoading ? 'RESOLVING...' : 'RESOLVE'}
        </button>
      </div>

      {/* Loading skeleton */}
      {dnsLoading && (
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-10 shadow-xl animate-pulse">
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
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-10 shadow-xl">
          <div className="flex items-center gap-4 mb-8">
            <Globe size={28} className={dnsResult.error ? 'text-danger' : 'text-success'} />
            <h3 className="text-2xl font-bold text-text uppercase tracking-tight">
              {dnsResult.hostname}
            </h3>
            {dnsResult.error && (
              <span className="px-4 py-1 text-sm font-bold text-danger bg-danger/10 rounded-full border border-danger/30 uppercase tracking-widest">
                Failed
              </span>
            )}
          </div>

          {dnsResult.error ? (
            <div className="bg-danger/5 border border-danger/20 rounded-2xl p-6">
              <p className="text-lg font-bold text-danger">{dnsResult.error}</p>
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-6">
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
                    <span className="text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">{section.label}</span>
                  </div>
                  {section.values && section.values.length > 0 ? (
                    <div className="space-y-2">
                      {section.values.map((v, i) => (
                        <div key={i} className="px-4 py-2 bg-panel-3 border border-border rounded-xl text-lg font-bold text-text tabular-nums font-[Geist_Mono]">
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
  )
}
