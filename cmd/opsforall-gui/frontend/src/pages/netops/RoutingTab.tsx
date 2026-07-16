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
import { SectionHeader } from '@/components/ui/SectionHeader'
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
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
        <div className="flex items-start gap-3">
          <div className="w-9 h-9 rounded-lg flex items-center justify-center bg-panel-3 border border-border shrink-0 mt-0.5">
            <Info size={18} className="text-accent" />
          </div>
          <p className="text-sm text-text-dim leading-relaxed">
            Routing Table — Displays the kernel routing table. Identify default gateways, static routes, and subnet paths for traffic engineering.
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
            'Default route (0.0.0.0) forwards unmatched traffic',
            'Gateway is the next-hop router IP',
            'Metric determines route preference (lower = preferred)',
            'Interface shows which adapter handles the route',
          ].map(item => (
            <div key={item} className="flex items-center gap-2.5 bg-panel-2 border border-border rounded-xl px-4 py-2.5">
              <HelpCircle size={14} className="text-accent shrink-0" />
              <span className="text-xs font-medium text-text-dim">{item}</span>
            </div>
          ))}
        </div>
      </div>

      {/* ── Default Routes ── */}
      {defaultRoutes.length > 0 && (
        <div className="bg-panel border border-accent/30 rounded-[var(--radius-lg)] p-6 shadow-xl">
          <SectionHeader icon={<Globe size={18} className="text-accent" />} title="Default Routes" count={defaultRoutes.length} />
          <div className="space-y-2">
            {defaultRoutes.map((route, idx) => (
              <div
                key={`default-${idx}`}
                className="flex items-center gap-4 bg-accent/5 border border-accent/20 rounded-xl px-4 py-3"
              >
                <div className="w-8 h-8 rounded-lg flex items-center justify-center bg-accent/15 border border-accent/30">
                  <Globe size={16} className="text-accent" />
                </div>
                <div className="flex-1 min-w-0 grid grid-cols-4 gap-4 text-xs">
                  <div>
                    <p className="text-[10px] font-bold text-text-faint uppercase tracking-wider mb-0.5">Destination</p>
                    <p className="font-bold text-accent">{route.destination}</p>
                  </div>
                  <div>
                    <p className="text-[10px] font-bold text-text-faint uppercase tracking-wider mb-0.5">Gateway</p>
                    <p className="font-bold text-text">{route.gateway}</p>
                  </div>
                  <div>
                    <p className="text-[10px] font-bold text-text-faint uppercase tracking-wider mb-0.5">Interface</p>
                    <p className="font-bold text-text-dim">{route.interface}</p>
                  </div>
                  <div>
                    <p className="text-[10px] font-bold text-text-faint uppercase tracking-wider mb-0.5">Metric</p>
                    <p className="font-bold text-text-dim tabular-nums">{route.metric}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* ── All Routes Table ── */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
        <SectionHeader icon={<Route size={18} className="text-accent" />} title="Routing Table" count={routes.length} />

        {isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 8 }).map((_, i) => (
              <div key={i} className="h-12 bg-panel-2 border border-border rounded-xl animate-pulse" />
            ))}
          </div>
        ) : routes.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12">
            <Route size={28} className="text-text-faint mb-3" />
            <p className="text-sm font-medium text-text-faint">No routing entries found.</p>
            <p className="text-xs text-text-faint mt-1">The routing table may be empty or inaccessible.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left">
              <thead>
                <tr className="border-b border-border">
                  {['Destination', 'Mask', 'Gateway', 'Interface', 'Metric'].map(h => (
                    <th key={h} className="px-4 py-2.5 text-[10px] font-bold text-text-faint uppercase tracking-widest">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {routes.map((route, idx) => (
                  <tr
                    key={`route-${idx}`}
                    className={cn(
                      'border-b border-border last:border-0 transition-colors',
                      route.is_default
                        ? 'bg-accent/5 hover:bg-accent/10'
                        : 'hover:bg-panel-2',
                      !route.is_default && idx % 2 === 0 ? 'bg-panel-2/30' : 'bg-transparent',
                    )}
                  >
                    <td className={cn('px-4 py-3 text-xs font-bold tabular-nums', route.is_default ? 'text-accent' : 'text-text')}>
                      {route.destination}
                    </td>
                    <td className="px-4 py-3 text-xs font-mono text-text-dim">{route.mask}</td>
                    <td className="px-4 py-3 text-xs font-bold text-text tabular-nums">{route.gateway}</td>
                    <td className="px-4 py-3">
                      <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-accent/15 text-accent border border-accent/30 uppercase tracking-wider">
                        {route.interface}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-xs font-bold text-text-dim tabular-nums">{route.metric}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
