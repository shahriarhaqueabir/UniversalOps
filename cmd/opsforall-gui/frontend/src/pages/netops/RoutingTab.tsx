import { useQuery } from '@tanstack/react-query'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import {
  Route,
  ListChecks,
  Info,
  HelpCircle,
  Globe,
} from 'lucide-react'
import { Panel, PanelHeader } from '@/components/ui/Panel'
import type { RouteEntryData } from '@/types'

export function RoutingTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: routes = [], isLoading } = useQuery<RouteEntryData[]>({
    queryKey: ['routing-table'],
    queryFn: async () => {
      const res = await call('NetOps.GetRoutingTable') as RouteEntryData[]
      return res || []
    },
    refetchInterval: refreshInterval,
  })

  const defaultRoutes = routes.filter(r => r.is_default)

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
              Routing Table — Displays the kernel routing table. Identify default gateways, static routes, and subnet paths for traffic engineering.
            </p>
          </div>
        </div>
      </Panel>

      {/* ── Checklist ── */}
      <Panel variant="elevated" padding="md">
        <PanelHeader icon={<ListChecks size={20} />} title="Route Verification Checklist" />
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {[
            'Default route (0.0.0.0) forwards unmatched traffic',
            'Gateway is the next-hop router IP',
            'Metric determines route preference (lower = preferred)',
            'Interface shows which adapter handles the route',
          ].map(item => (
            <div key={item} className="flex items-center gap-3 bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-2xl px-5 py-3 shadow-sm">
              <HelpCircle size={14} className="text-accent shrink-0" />
              <span className="text-[11px] font-bold text-text-dim uppercase tracking-wider">{item}</span>
            </div>
          ))}
        </div>
      </Panel>

      {/* ── Default Routes ── */}
      {defaultRoutes.length > 0 && (
        <Panel variant="default" padding="md" category="network" className="border-accent/30">
          <PanelHeader icon={<Globe size={20} />} title="Active Gateways" category="network" action={<span className="text-xs font-black px-4 py-1.5 rounded-full bg-panel-3 text-accent border border-border tabular-nums shadow-inner">{defaultRoutes.length} Default</span>} />
          <div className="space-y-3">
            {defaultRoutes.map((route, idx) => (
              <div
                key={`default-${idx}`}
                className="flex items-center gap-4 bg-[var(--color-bg)]/50 border border-[var(--color-border)] rounded-2xl px-6 py-4 shadow-sm"
              >
                <div className="w-10 h-10 rounded-xl flex items-center justify-center bg-accent/10 border border-accent/20">
                  <Globe size={18} className="text-accent" />
                </div>
                <div className="flex-1 min-w-0 grid grid-cols-4 gap-6">
                  <div>
                    <p className="text-[9px] font-black text-text-faint uppercase tracking-widest mb-1">Destination</p>
                    <p className="text-sm font-black text-accent tabular-nums">{route.destination}</p>
                  </div>
                  <div>
                    <p className="text-[9px] font-black text-text-faint uppercase tracking-widest mb-1">Gateway</p>
                    <p className="text-sm font-bold text-text tabular-nums">{route.gateway}</p>
                  </div>
                  <div>
                    <p className="text-[9px] font-black text-text-faint uppercase tracking-widest mb-1">Interface</p>
                    <p className="text-sm font-black text-[var(--color-text-dim)] uppercase truncate">{route.interface}</p>
                  </div>
                  <div>
                    <p className="text-[9px] font-black text-text-faint uppercase tracking-widest mb-1">Priority</p>
                    <p className="text-sm font-black text-text-dim tabular-nums">{route.metric}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </Panel>
      )}

      {/* ── All Routes Table ── */}
      <Panel variant="default" padding="none" category="network" className="overflow-hidden">
        <div className="p-8">
          <PanelHeader
            icon={<Route size={20} />}
            title="Full Routing Table"
            category="network"
            action={<span className="text-xs font-black px-4 py-1.5 rounded-full bg-panel-3 text-accent border border-border tabular-nums shadow-inner">{routes.length} Paths</span>}
          />

          {isLoading ? (
            <div className="space-y-3">
              {Array.from({ length: 8 }).map((_, i) => (
                <div key={i} className="h-14 bg-panel-2 border border-border rounded-xl animate-pulse" />
              ))}
            </div>
          ) : routes.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 opacity-30">
              <Route size={40} className="text-text-faint mb-4" />
              <p className="text-xs font-black uppercase tracking-widest text-text-faint">No Routing Paths Discovered</p>
            </div>
          ) : (
            <div className="overflow-x-auto -mx-8">
              <table className="w-full text-left border-collapse">
                <thead className="bg-[var(--color-panel-2)] border-y border-[var(--color-border)]/50">
                  <tr>
                    {['Destination', 'Mask', 'Gateway', 'Interface', 'Metric'].map(h => (
                      <th key={h} className="px-8 py-4 text-[10px] font-black text-text-faint uppercase tracking-[0.2em]">{h}</th>
                    ))}
                  </tr>
                </thead>
                <tbody className="divide-y divide-[var(--color-border)]/30">
                  {routes.map((route, idx) => (
                    <tr
                      key={`route-${idx}`}
                      className={cn(
                        'hover:bg-[var(--color-sidebar-hover)] transition-colors group',
                        route.is_default && 'bg-accent/5'
                      )}
                    >
                      <td className={cn('px-8 py-4 text-xs font-black tabular-nums', route.is_default ? 'text-accent' : 'text-text')}>
                        {route.destination}
                      </td>
                      <td className="px-8 py-4 text-xs font-mono font-bold text-text-dim">{route.mask}</td>
                      <td className="px-8 py-4 text-xs font-bold text-text tabular-nums">{route.gateway}</td>
                      <td className="px-8 py-4">
                        <span className="text-[9px] font-black px-3 py-1 rounded-lg bg-accent/10 text-accent border border-accent/20 uppercase tracking-widest group-hover:bg-accent group-hover:text-white transition-all">
                          {route.interface}
                        </span>
                      </td>
                      <td className="px-8 py-4 text-xs font-black text-text-dim tabular-nums">{route.metric}</td>
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
