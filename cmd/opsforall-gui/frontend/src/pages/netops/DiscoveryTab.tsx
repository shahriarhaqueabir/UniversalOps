import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import {
  Search,
  Radio,
  Clock,
  Hash,
  Wifi,
  RefreshCw,
} from 'lucide-react'
import { SectionBriefing } from '@/components/ui/SectionBriefing'
import { SearchInput } from '@/components/ui/SearchInput'
import type { DiscoveryResultData } from '@/types'

// ── Main DiscoveryTab ──

export function DiscoveryTab() {
  const { call } = useBackend()
  const [subnet, setSubnet] = useState('')
  const [hasRun, setHasRun] = useState(false)

  const { data: result, isLoading, refetch } = useQuery<DiscoveryResultData>({
    queryKey: ['netops-discovery', subnet],
    queryFn: async () => {
      const res = (await call('NetOps.RunNetworkDiscovery', subnet)) as DiscoveryResultData
      setHasRun(true)
      return res
    },
    enabled: false,
    retry: false,
  })

  const runDiscovery = () => {
    if (!subnet.trim()) return
    refetch()
  }

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <SectionBriefing
        title="Network Discovery"
        objective="Scan the local subnet to discover devices. Combines ARP table data with ping sweep to map active hosts on the network."
        checklist={[
          'ARP entries reveal recently-seen devices',
          'Ping sweep finds responsive hosts',
          'Response time indicates network proximity',
          'Vendor lookup identifies device manufacturer',
        ]}
      />

      {/* ── Subnet Input ── */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
        <div className="flex items-center gap-4">
          <div className="flex-1">
            <SearchInput
              size="lg"
              value={subnet}
              onChange={(e) => setSubnet(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && runDiscovery()}
              placeholder="Enter subnet prefix (e.g., 192.168.1)"
            />
          </div>
          <button
            onClick={runDiscovery}
            disabled={isLoading || !subnet.trim()}
            className={cn(
              'flex items-center gap-3 px-6 py-3 text-sm font-semibold rounded-xl transition-all shadow-xl',
              isLoading || !subnet.trim()
                ? 'bg-panel-3 text-text-faint border border-border cursor-not-allowed'
                : 'bg-accent text-white hover:bg-accent/90',
            )}
          >
            {isLoading ? (
              <RefreshCw size={16} className="animate-spin" />
            ) : (
              <Wifi size={16} />
            )}
            {isLoading ? 'Scanning...' : 'Discover'}
          </button>
        </div>
      </div>

      {/* ── Loading State ── */}
      {isLoading && (
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-16 shadow-xl flex flex-col items-center justify-center">
          <div className="w-12 h-12 rounded-xl bg-panel-3 flex items-center justify-center border border-border mb-4">
            <Radio size={24} className="text-accent animate-pulse" />
          </div>
          <p className="text-sm font-bold text-text uppercase tracking-widest mb-1">Scanning Subnet</p>
          <p className="text-xs font-medium text-text-faint">
            ARP lookup and ping sweep in progress: this may take a moment...
          </p>
        </div>
      )}

      {/* ── Results ── */}
      {result && !isLoading && (
        <>
          {/* ── Summary Bar ── */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {[
              { label: 'Devices Found', value: result.devices.length, icon: <Hash size={18} /> },
              { label: 'Subnet Scanned', value: result.subnet, icon: <Radio size={18} /> },
              { label: 'Scan Time', value: `${result.scan_time_ms}ms`, icon: <Clock size={18} /> },
            ].map(card => (
              <div
                key={card.label}
                className="bg-panel border border-border rounded-xl p-5 flex items-center gap-4 hover:border-accent/30 transition-colors"
              >
                <div className="w-11 h-11 rounded-lg flex items-center justify-center shrink-0 bg-panel-3 border border-border text-accent">
                  {card.icon}
                </div>
                <div>
                  <p className="text-[10px] font-bold text-text-faint uppercase tracking-wider mb-1">
                    {card.label}
                  </p>
                  <p className="text-lg font-bold text-text tabular-nums truncate">{card.value}</p>
                </div>
              </div>
            ))}
          </div>

          {/* ── Device Table ── */}
          <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-9 h-9 rounded-lg flex items-center justify-center bg-panel-3 border border-border">
                <Wifi size={18} className="text-accent" />
              </div>
              <h3 className="text-sm font-bold text-text uppercase tracking-widest">Discovered Devices</h3>
              <span className="ml-auto px-2.5 py-0.5 text-[10px] font-bold rounded-full bg-accent/15 text-accent border border-accent/30">
                {result.devices.length}
              </span>
            </div>

            {result.devices.length === 0 ? (
              <p className="text-sm font-medium text-text-faint text-center py-8">
                No devices found on this subnet
              </p>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left">
                  <thead>
                    <tr className="border-b border-border">
                      {['IP Address', 'MAC Address', 'Vendor', 'Hostname', 'Response Time'].map(
                        h => (
                          <th
                            key={h}
                            className="px-4 py-3 text-[10px] font-bold text-text-faint uppercase tracking-widest"
                          >
                            {h}
                          </th>
                        ),
                      )}
                    </tr>
                  </thead>
                  <tbody>
                    {result.devices.map((device, idx) => (
                      <tr
                        key={idx}
                        className="border-b border-border/50 transition-colors hover:bg-panel-2/50"
                      >
                        <td className="px-4 py-3">
                          <span className="text-xs font-bold text-text tabular-nums">
                            {device.ip}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          <span className="text-xs font-medium text-text-dim font-mono">
                            {device.mac}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          <span className="text-xs font-medium text-text-dim">
                            {device.vendor || 'Unknown'}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          <span className="text-xs font-medium text-text-dim">
                            {device.hostname || 'N/A'}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          <span
                            className={cn(
                              'text-xs font-bold tabular-nums',
                              device.response_time_ms < 5
                                ? 'text-success'
                                : device.response_time_ms < 20
                                  ? 'text-warning'
                                  : 'text-danger',
                            )}
                          >
                            {device.response_time_ms}ms
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </>
      )}

      {/* ── Empty State ── */}
      {!hasRun && !isLoading && !result && (
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-16 shadow-xl flex flex-col items-center justify-center text-center">
          <div className="w-16 h-16 rounded-2xl bg-panel-3 flex items-center justify-center border border-border mb-4">
            <Search size={32} className="text-text-faint" />
          </div>
          <p className="text-sm font-bold text-text uppercase tracking-widest mb-1">
            No Discovery Data
          </p>
          <p className="text-xs font-medium text-text-faint">
            Enter a subnet prefix to begin discovery
          </p>
        </div>
      )}
    </div>
  )
}
