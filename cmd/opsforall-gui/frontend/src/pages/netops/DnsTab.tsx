import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import {
  Globe,
  Search,
  Server,
  RefreshCw,
  Info,
  ChevronRight,
  AlertTriangle,
} from 'lucide-react'
import type { DNSResult } from '@/types'
import { SectionBriefing } from '@/components/ui/SectionBriefing'
import { SearchInput } from '@/components/ui/SearchInput'

export function DnsTab() {
  const { call } = useBackend()
  const { dnsTimeout } = useSettingsStore()
  const [dnsHost, setDnsHost] = useState('google.com')
  const [dnsServer, setDnsServer] = useState('')

  const dnsMutation = useMutation({
    mutationFn: async () => {
      return await call('NetOps.DNSLookup', dnsHost, dnsServer, dnsTimeout) as DNSResult
    },
    onError: (err: unknown) => {
      console.error('DNS lookup failed:', err)
    },
  })

  const handleDns = () => dnsMutation.mutate()

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
      <div className="flex items-center gap-4 bg-[var(--color-panel-2)] border border-[var(--color-border)] p-5 rounded-[var(--radius-lg)]">
        <SearchInput
          size="lg"
          value={dnsHost}
          onChange={(e) => setDnsHost(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleDns()}
          placeholder="Hostname (e.g. google.com)"
          className="flex-[2]"
        />
        <SearchInput
          size="lg"
          icon={<Server size={18} />}
          value={dnsServer}
          onChange={(e) => setDnsServer(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleDns()}
          placeholder="Resolver (e.g. 8.8.8.8)"
          className="flex-1"
        />
        <button onClick={handleDns} disabled={dnsMutation.isPending} className="flex items-center gap-2 px-5 py-2.5 bg-[var(--color-accent)] text-white text-sm font-bold rounded-xl hover:bg-[var(--color-accent)]/90 transition-all h-12 shrink-0">
          {dnsMutation.isPending ? <RefreshCw size={16} className="animate-spin" /> : <Search size={16} />}
          {dnsMutation.isPending ? 'RESOLVING...' : 'RESOLVE'}
        </button>
      </div>

      {/* Loading skeleton */}
      {dnsMutation.isPending && (
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
      {dnsMutation.data && !dnsMutation.isPending && !dnsMutation.data.error && (
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-10 shadow-xl">
          <div className="flex items-center gap-4 mb-8">
            <Globe size={28} className="text-success" />
            <h3 className="text-2xl font-bold text-text uppercase tracking-tight">
              {dnsMutation.data.hostname}
            </h3>
          </div>

          <div className="grid grid-cols-2 gap-6">
            {[
              { label: 'A Records', values: dnsMutation.data.a, icon: <Globe size={18} /> },
              { label: 'AAAA Records', values: dnsMutation.data.aaaa, icon: <Globe size={18} /> },
              { label: 'MX Records', values: dnsMutation.data.mx, icon: <Server size={18} /> },
              { label: 'NS Records', values: dnsMutation.data.ns, icon: <Server size={18} /> },
              { label: 'TXT Records', values: dnsMutation.data.txt, icon: <Info size={18} /> },
              { label: 'CNAME', values: dnsMutation.data.cname ? [dnsMutation.data.cname] : [], icon: <ChevronRight size={18} /> },
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
        </div>
      )}

      {/* Mutation error */}
      {dnsMutation.isError && (
        <div className="bg-danger/10 border border-danger/30 rounded-[var(--radius-lg)] p-6 flex items-start gap-3">
          <AlertTriangle size={20} className="text-danger shrink-0 mt-0.5" />
          <div>
            <p className="text-sm font-semibold text-danger">DNS lookup failed</p>
            <p className="text-sm text-[var(--color-text-dim)] mt-1">{String(dnsMutation.error)}</p>
          </div>
        </div>
      )}
    </div>
  )
}
