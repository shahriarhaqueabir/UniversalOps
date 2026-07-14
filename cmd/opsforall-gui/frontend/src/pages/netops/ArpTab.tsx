import { useQuery } from '@tanstack/react-query'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import {
  Network,
  ListChecks,
  Info,
  HelpCircle,
} from 'lucide-react'
import type { ARPEntryData } from '@/types'

// ── Helpers ──

function SectionHeader({ icon, title }: { icon: React.ReactNode; title: string }) {
  return (
    <div className="flex items-center gap-3 mb-4">
      <div className="w-9 h-9 rounded-lg flex items-center justify-center bg-panel-3 border border-border">
        {icon}
      </div>
      <h3 className="text-sm font-bold text-text uppercase tracking-widest">{title}</h3>
    </div>
  )
}

// ── Main ArpTab ──

export function ArpTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: entries = [], isLoading } = useQuery<ARPEntryData[]>({
    queryKey: ['arp-table'],
    queryFn: async () => {
      const res = await call('NetOps.GetARPTable') as ARPEntryData[]
      return res || []
    },
    refetchInterval: refreshInterval,
  })

  return (
    <div className="space-y-6 animate-in fade-in duration-500">

      {/* ── Briefing ── */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
        <div className="flex items-start gap-3">
          <div className="w-9 h-9 rounded-lg flex items-center justify-center bg-panel-3 border border-border shrink-0 mt-0.5">
            <Info size={18} className="text-accent" />
          </div>
          <p className="text-sm text-text-dim leading-relaxed">
            ARP Table — Maps IP addresses to MAC addresses on the local network segment. Use to identify devices, detect spoofing, and audit network access.
          </p>
        </div>
      </div>

      {/* ── Checklist ── */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
        <div className="flex items-center gap-3 mb-4">
          <div className="w-9 h-9 rounded-lg flex items-center justify-center bg-panel-3 border border-border">
            <ListChecks size={18} className="text-accent" />
          </div>
          <h3 className="text-sm font-bold text-text uppercase tracking-widest">What to Look For</h3>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {[
            'IP-MAC mappings reveal device identities',
            'Vendor lookup identifies hardware manufacturer',
            'Duplicate IPs may signal ARP spoofing',
            'Stale entries indicate disconnected devices',
          ].map(item => (
            <div key={item} className="flex items-center gap-2.5 bg-panel-2 border border-border rounded-xl px-4 py-2.5">
              <HelpCircle size={14} className="text-accent shrink-0" />
              <span className="text-xs font-medium text-text-dim">{item}</span>
            </div>
          ))}
        </div>
      </div>

      {/* ── ARP Table ── */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
        <SectionHeader icon={<Network size={18} className="text-accent" />} title="ARP Table" />

        {isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i} className="h-12 bg-panel-2 border border-border rounded-xl animate-pulse" />
            ))}
          </div>
        ) : entries.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12">
            <Network size={28} className="text-text-faint mb-3" />
            <p className="text-sm font-medium text-text-faint">No ARP entries found.</p>
            <p className="text-xs text-text-faint mt-1">Ensure you have network interfaces active.</p>
          </div>
        ) : (
          <>
            <div className="flex items-center justify-between mb-3">
              <p className="text-xs font-medium text-text-faint">
                {entries.length} entr{entries.length === 1 ? 'y' : 'ies'}
              </p>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full text-left">
                <thead>
                  <tr className="border-b border-border">
                    {['IP Address', 'MAC Address', 'Vendor', 'Interface'].map(h => (
                      <th key={h} className="px-4 py-2.5 text-[10px] font-bold text-text-faint uppercase tracking-widest">{h}</th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {entries.map((entry, idx) => (
                    <tr
                      key={`${entry.ip}-${idx}`}
                      className={cn(
                        'border-b border-border last:border-0 hover:bg-panel-2 transition-colors',
                        idx % 2 === 0 ? 'bg-panel-2/30' : 'bg-transparent',
                      )}
                    >
                      <td className="px-4 py-3 text-xs font-bold text-text tabular-nums">{entry.ip}</td>
                      <td className="px-4 py-3 text-xs font-mono text-text-dim">{entry.mac}</td>
                      <td className="px-4 py-3 text-xs font-medium text-text-dim truncate max-w-[200px]">{entry.vendor || 'Unknown'}</td>
                      <td className="px-4 py-3">
                        <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-accent/15 text-accent border border-accent/30 uppercase tracking-wider">
                          {entry.interface}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
