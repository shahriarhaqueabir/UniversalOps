import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { cn } from '@/lib/utils'
import {
  AlertTriangle,
  Globe,
  Map,
  RefreshCw,
} from 'lucide-react'
import type { TraceResult } from '@/types'
import { SectionBriefing } from '@/components/ui/SectionBriefing'

import { SearchInput } from '@/components/ui/SearchInput'

export function TracerouteTab() {
  const { call } = useBackend()
  const [traceTarget, setTraceTarget] = useState('8.8.8.8')

  const traceMutation = useMutation({
    mutationFn: async () => {
      return await call('NetOps.Traceroute', traceTarget) as TraceResult
    },
  })

  const executeTrace = () => traceMutation.mutate()

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <SectionBriefing
        title="Route Trace Analysis"
        objective="Map the network path and measure hop-by-hop latency to a target host."
        checklist={['Enter target hostname or IP', 'Click trace to begin', 'Review each hop\'s latency and IP']}
      />

      {/* Input + Execute */}
      <div className="flex items-center gap-6 bg-panel-2 border border-border p-6 rounded-[var(--radius-lg)] shadow-inner">
        <div className="flex-1">
          <SearchInput
            icon={<Globe size={18} />}
            value={traceTarget}
            onChange={(e) => setTraceTarget(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && executeTrace()}
            placeholder="Target hostname or IP"
            size="lg"
            className="shadow-xl"
          />
        </div>
        <button onClick={executeTrace} disabled={traceMutation.isPending} className="flex items-center gap-3 px-8 py-3.5 bg-[var(--color-accent)] text-white text-sm font-black uppercase tracking-widest rounded-xl hover:bg-accent/90 shadow-xl transition-all disabled:opacity-50 active:scale-95 h-12">
          {traceMutation.isPending ? <RefreshCw size={18} className="animate-spin" /> : <Map size={18} />}
          {traceMutation.isPending ? 'TRACING...' : 'TRACE ROUTE'}
        </button>
      </div>

      {/* Loading skeleton */}
      {traceMutation.isPending && (
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
          <div className="flex items-center gap-4 mb-8">
            <div className="w-8 h-8 rounded-full bg-panel-3 animate-pulse" />
            <div className="h-6 w-48 bg-panel-3 rounded animate-pulse" />
          </div>
          <div className="space-y-4">
            {/* static skeleton */}
            {[1,2,3,4,5].map((i) => (
              <div key={i} className="flex items-center gap-6">
                <div className="h-4 w-16 bg-panel-3 rounded animate-pulse" />
                <div className="h-4 w-40 bg-panel-3 rounded animate-pulse" />
                <div className="h-4 w-20 bg-panel-3 rounded animate-pulse ml-auto" />
                <div className="h-6 w-24 bg-panel-3 rounded animate-pulse ml-auto" />
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Error state */}
      {traceMutation.isError && (
        <div className="bg-danger/10 border border-danger/30 rounded-[var(--radius-lg)] p-6 flex items-start gap-3">
          <AlertTriangle size={20} className="text-danger shrink-0 mt-0.5" />
          <div>
            <p className="text-sm font-semibold text-danger">Trace failed</p>
            <p className="text-sm text-[var(--color-text-dim)] mt-1">{String(traceMutation.error)}</p>
          </div>
        </div>
      )}

      {/* Results Table */}
      {traceMutation.data && !traceMutation.isPending && !traceMutation.data.error && (
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] overflow-hidden shadow-2xl">
          <div className="px-8 py-6 bg-panel-2 border-b border-border flex items-center justify-between">
            <h3 className="text-xl font-bold text-text uppercase tracking-widest flex items-center gap-3">
              <Globe size={20} className="text-accent" /> {traceMutation.data.target}
            </h3>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border">
                <tr>
                  <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider w-24">Hop</th>
                  <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Host / IP</th>
                  <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider text-right">RTT (ms)</th>
                  <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider text-right w-28">Status</th>
                </tr>
              </thead>
              <tbody>
                {traceMutation.data.hops.map((hop) => (
                  <tr key={hop.number} className="border-b border-border/20 hover:bg-[var(--color-sidebar-hover)] transition-all group">
                    <td className="px-8 py-5 text-sm font-semibold text-[var(--color-accent)] tabular-nums">{hop.number}</td>
                    <td className="px-8 py-5">
                      <div className="flex flex-col">
                        <span className="text-sm font-medium text-[var(--color-text)]">{hop.host || 'Unknown'}</span>
                        <span className="text-sm font-bold text-text-faint uppercase">{hop.ip || 'N/A'}</span>
                      </div>
                    </td>
                    <td className="px-8 py-5 text-right">
                      <span className="text-sm font-medium text-[var(--color-text)] tabular-nums">
                        {hop.rtts_ms.length > 0 ? hop.rtts_ms.join(', ') : 'N/A'}
                      </span>
                    </td>
                    <td className="px-8 py-5 text-right">
                      <span className={cn(
                        "px-4 py-1.5 rounded-full text-sm font-bold uppercase tracking-widest",
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
  )
}
