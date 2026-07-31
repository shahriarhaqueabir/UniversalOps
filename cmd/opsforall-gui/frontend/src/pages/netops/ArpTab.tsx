import { useQuery } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import {
  Network,
  ListChecks,
  Info,
  HelpCircle,
} from 'lucide-react'
import { Panel, PanelHeader } from '@/components/ui/Panel'
import type { ARPEntryData } from '@/types'

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
      <Panel variant="default" padding="md">
        <div className="flex items-start gap-4">
          <div className="w-11 h-11 rounded-xl flex items-center justify-center bg-[var(--color-panel-3)] border border-[var(--color-border)] shrink-0 shadow-inner">
            <Info size={18} className="text-accent" />
          </div>
          <div className="pt-1">
            <p className="text-xs font-black text-[var(--color-text)] uppercase tracking-[0.2em] mb-1">Knowledge Node</p>
            <p className="text-sm text-text-dim leading-relaxed font-medium">
              ARP Table — Maps IP addresses to MAC addresses on the local network segment. Use to identify devices, detect spoofing, and audit network access.
            </p>
          </div>
        </div>
      </Panel>

      {/* ── Checklist ── */}
      <Panel variant="elevated" padding="md">
        <PanelHeader icon={<ListChecks size={20} />} title="Observation List" subtitle="Audit checklist for local segment" />
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {[
            'IP-MAC mappings reveal device identities',
            'Vendor lookup identifies hardware manufacturer',
            'Duplicate IPs may signal ARP spoofing',
            'Stale entries indicate disconnected devices',
          ].map(item => (
            <div key={item} className="flex items-center gap-3 bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-2xl px-5 py-3 shadow-sm transition-all hover:scale-[1.02]">
              <HelpCircle size={14} className="text-accent shrink-0" />
              <span className="text-[11px] font-bold text-text-dim uppercase tracking-wider">{item}</span>
            </div>
          ))}
        </div>
      </Panel>

      {/* ── ARP Table ── */}
      <Panel variant="default" padding="none" category="network" className="overflow-hidden">
        <div className="p-8">
          <PanelHeader
            icon={<Network size={20} />}
            title="ARP Table"
            category="network"
            action={<span className="text-xs font-black px-4 py-1.5 rounded-full bg-panel-3 text-accent border border-border tabular-nums shadow-inner">{entries.length} Nodes</span>}
          />

          {isLoading ? (
            <div className="space-y-3">
              {/* static skeleton */}
              {Array.from({ length: 6 }).map((_, i) => (
                <div key={i} className="h-14 bg-panel-2 border border-border rounded-xl animate-pulse" />
              ))}
            </div>
          ) : entries.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 opacity-30">
              <Network size={40} className="text-text-faint mb-4" />
              <p className="text-xs font-black uppercase tracking-widest text-text-faint">No Active Nodes Detected</p>
            </div>
          ) : (
            <div className="overflow-x-auto -mx-8">
              <table className="w-full text-left border-collapse">
                <thead className="bg-[var(--color-panel-2)] border-y border-[var(--color-border)]/50">
                  <tr>
                    {['IP Address', 'MAC Address', 'Vendor', 'Interface'].map(h => (
                      <th key={h} className="px-8 py-4 text-[10px] font-black text-text-faint uppercase tracking-[0.2em]">{h}</th>
                    ))}
                  </tr>
                </thead>
                <tbody className="divide-y divide-[var(--color-border)]/30">
                  {entries.map((entry, idx) => (
                    <tr
                      key={`${entry.ip}-${idx}`}
                      className="hover:bg-[var(--color-sidebar-hover)] transition-colors group"
                    >
                      <td className="px-8 py-4 text-xs font-black text-text tabular-nums">{entry.ip}</td>
                      <td className="px-8 py-4 text-xs font-mono font-bold text-text-dim">{entry.mac}</td>
                      <td className="px-8 py-4 text-xs font-black text-text-faint uppercase tracking-tighter truncate max-w-[200px]">{entry.vendor || 'UNKNOWN'}</td>
                      <td className="px-8 py-4">
                        <span className="text-[9px] font-black px-3 py-1 rounded-lg bg-accent/10 text-accent border border-accent/20 uppercase tracking-widest group-hover:bg-accent group-hover:text-white transition-all">
                          {entry.interface}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </Panel>
    </div>
  )
}
