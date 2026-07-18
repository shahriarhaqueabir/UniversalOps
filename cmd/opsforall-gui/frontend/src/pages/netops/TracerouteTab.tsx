import { useState, useCallback } from 'react'
import { useBackend } from '@/hooks/useBackend'
import { cn } from '@/lib/utils'
import {
  Globe,
  Map,
  RefreshCw,
} from 'lucide-react'
import type { TraceResult } from '@/types'
import { SectionBriefing } from '@/components/ui/SectionBriefing'

export function TracerouteTab() {
  const { call } = useBackend()
  const [traceTarget, setTraceTarget] = useState('8.8.8.8')
  const [traceResult, setTraceResult] = useState<TraceResult | null>(null)
  const [traceRunning, setTraceRunning] = useState(false)

  const executeTrace = useCallback(async () => {
    setTraceRunning(true); setTraceResult(null)
    try {
      const res = await call('NetOps.Traceroute', traceTarget) as TraceResult
      setTraceResult(res)
    } catch (err: unknown) {
      console.error('Traceroute failed:', err)
      setTraceResult({ target: traceTarget, hops: [], error: String(err) })
    } finally {
      setTraceRunning(false)
    }
  }, [call, traceTarget])

  return (
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
      <div className="flex items-center gap-6 bg-panel-2 border border-border p-6 rounded-[var(--radius-lg)] shadow-inner">
        <div className="relative group flex-1">
          <Globe size={24} className="absolute left-5 top-1/2 -translate-y-1/2 text-text-faint group-focus-within:text-accent transition-colors" />
          <input
            type="text"
            value={traceTarget}
            onChange={(e) => setTraceTarget(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && executeTrace()}
            placeholder="Target hostname or IP"
            className="w-full bg-panel border border-border rounded-2xl pl-16 pr-4 py-3 text-sm font-medium text-[var(--color-text)] placeholder-[var(--color-text-faint)] focus:outline-none focus:border-accent shadow-xl"
          />
        </div>
        <button onClick={executeTrace} disabled={traceRunning} className="flex items-center gap-3 px-5 py-2.5 bg-[var(--color-accent)] text-white text-sm font-semibold rounded-xl hover:bg-accent/90 shadow-xl transition-all disabled:opacity-50">
          {traceRunning ? <RefreshCw size={24} className="animate-spin" /> : <Map size={24} />}
          {traceRunning ? 'TRACING...' : 'TRACE ROUTE'}
        </button>
      </div>

      {/* Results Table */}
      {traceResult && !traceRunning && (
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] overflow-hidden shadow-2xl">
          <div className="px-8 py-6 bg-panel-2 border-b border-border flex items-center justify-between">
            <h3 className="text-xl font-bold text-text uppercase tracking-widest flex items-center gap-3">
              <Globe size={20} className="text-accent" /> {traceResult.target}
            </h3>
            {traceResult.error && (
              <span className="px-4 py-1.5 text-sm font-bold text-danger bg-danger/10 rounded-full border border-danger/30 uppercase tracking-widest">
                FAILED
              </span>
            )}
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
                {traceResult.hops.map((hop) => (
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
