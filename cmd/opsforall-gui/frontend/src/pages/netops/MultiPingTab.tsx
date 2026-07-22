import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import {
  Play,
  ListChecks,
  Info,
  HelpCircle,
  Loader2,
  Target,
  Clock,
  Activity,
  AlertTriangle,
  ArrowUpDown,
} from 'lucide-react'
import { Panel, PanelHeader } from '@/components/ui/Panel'
import type { PingResultMultiData, PingStatsData } from '@/types'

function ResultCard({ result }: { result: PingResultMultiData }) {
  return (
    <div className={cn(
      'bg-[var(--color-panel-2)] border rounded-2xl p-5 transition-all shadow-sm',
      result.success ? 'border-success/30 hover:border-success/50' : 'border-danger/30 hover:border-danger/50',
    )}>
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2 min-w-0">
          <Target size={14} className="text-accent shrink-0" />
          <span className="text-xs font-black text-[var(--color-text)] uppercase tracking-widest truncate">{result.target}</span>
        </div>
        {result.success ? (
          <span className="flex items-center gap-1 text-[9px] font-black text-success uppercase border border-success/30 px-2 py-0.5 rounded-lg bg-success/10">
            OK
          </span>
        ) : (
          <span className="flex items-center gap-1 text-[9px] font-black text-danger uppercase border border-danger/30 px-2 py-0.5 rounded-lg bg-danger/10">
            FAIL
          </span>
        )}
      </div>

      {result.success ? (
        <div className="grid grid-cols-2 gap-4">
          <div className="bg-[var(--color-panel-3)] p-3 rounded-xl border border-[var(--color-border)]/50">
            <p className="text-[9px] font-black text-text-faint uppercase tracking-widest mb-1 opacity-60">Avg Latency</p>
            <p className={cn(
              'text-sm font-black tabular-nums',
              result.avg_ms < 50 ? 'text-success' : result.avg_ms < 150 ? 'text-warning' : 'text-danger',
            )}>
              {result.avg_ms.toFixed(1)}ms
            </p>
          </div>
          <div className="bg-[var(--color-panel-3)] p-3 rounded-xl border border-[var(--color-border)]/50">
            <p className="text-[9px] font-black text-text-faint uppercase tracking-widest mb-1 opacity-60">Jitter</p>
            <p className={cn(
              'text-sm font-black tabular-nums',
              result.jitter_ms < 10 ? 'text-success' : result.jitter_ms < 30 ? 'text-warning' : 'text-danger',
            )}>
              {result.jitter_ms.toFixed(1)}ms
            </p>
          </div>
          <div className="bg-[var(--color-panel-3)] p-3 rounded-xl border border-[var(--color-border)]/50">
            <p className="text-[9px] font-black text-text-faint uppercase tracking-widest mb-1 opacity-60">Packet Loss</p>
            <p className={cn(
              'text-sm font-black tabular-nums',
              result.packet_loss === 0 ? 'text-success' : result.packet_loss < 25 ? 'text-warning' : 'text-danger',
            )}>
              {result.packet_loss.toFixed(1)}%
            </p>
          </div>
          <div className="bg-[var(--color-panel-3)] p-3 rounded-xl border border-[var(--color-border)]/50">
            <p className="text-[9px] font-black text-text-faint uppercase tracking-widest mb-1 opacity-60">Consistency</p>
            <p className={cn(
              'text-sm font-black tabular-nums',
              result.stddev_ms < 10 ? 'text-success' : result.stddev_ms < 30 ? 'text-warning' : 'text-danger',
            )}>
              {result.stddev_ms.toFixed(1)}ms
            </p>
          </div>
        </div>
      ) : (
        <div className="flex items-center gap-2 mt-2 p-3 bg-danger/5 border border-danger/10 rounded-xl">
          <AlertTriangle size={12} className="text-danger shrink-0" />
          <p className="text-[10px] font-bold text-danger uppercase tracking-tight">{result.error || 'Connection Refused'}</p>
        </div>
      )}
    </div>
  )
}

// ── Main MultiPingTab ──

export function MultiPingTab() {
  const { call } = useBackend()

  const [targetsInput, setTargetsInput] = useState('8.8.8.8, 1.1.1.1, google.com')
  const [count, setCount] = useState(4)
  const [results, setResults] = useState<PingResultMultiData[]>([])
  const [stats, setStats] = useState<PingStatsData | null>(null)
  const [error, setError] = useState<string | null>(null)

  const runMutation = useMutation({
    mutationFn: async () => {
      const targets = targetsInput
        .split(',')
        .map(t => t.trim())
        .filter(t => t.length > 0)

      if (targets.length === 0) {
        throw new Error('Enter at least one target')
      }

      const pingResults = await call('NetOps.PingMultiTarget', targets, count) as PingResultMultiData[]
      const pingStats = await call('NetOps.GetPingStats', pingResults || []) as PingStatsData

      return { results: pingResults || [], stats: pingStats }
    },
    onSuccess: (data) => {
      setResults(data.results)
      setStats(data.stats)
      setError(null)
    },
    onError: (err: Error) => {
      setResults([])
      setStats(null)
      setError(err.message || 'Multi-ping failed')
    },
  })

  const handleRun = () => {
    setError(null)
    setResults([])
    setStats(null)
    runMutation.mutate()
  }

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
              Multi-Ping — Simultaneously ping multiple targets to compare latency and loss across different endpoints. Ideal for benchmarking and failover testing.
            </p>
          </div>
        </div>
      </Panel>

      {/* ── Checklist ── */}
      <Panel variant="elevated" padding="md">
        <PanelHeader icon={<ListChecks size={20} />} title="Path Integrity Checklist" />
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {[
            'Concurrent pings reveal path-specific issues',
            'StdDev indicates consistency of response times',
            'Jitter measures inter-packet delay variation',
            'Aggregate stats highlight the worst-performing target',
          ].map(item => (
            <div key={item} className="flex items-center gap-3 bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-2xl px-5 py-3 shadow-sm">
              <HelpCircle size={14} className="text-accent shrink-0" />
              <span className="text-[11px] font-bold text-text-dim uppercase tracking-wider">{item}</span>
            </div>
          ))}
        </div>
      </Panel>

      {/* ── Controls ── */}
      <Panel variant="default" padding="md" category="network">
        <PanelHeader icon={<Target size={20} />} title="Multi-Ping Orchestrator" category="network" />
        <div className="space-y-6">
          <div className="space-y-2">
            <label className="text-[10px] font-black text-text-faint uppercase tracking-[0.2em] block">Target Infrastructure (Comma-Separated)</label>
            <input
              type="text"
              value={targetsInput}
              onChange={(e) => setTargetsInput(e.target.value)}
              placeholder="8.8.8.8, 1.1.1.1, google.com"
              className="w-full bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-xl px-5 py-3 text-sm font-bold text-text font-mono focus:outline-none focus:border-accent transition-all shadow-inner"
              onKeyDown={(e) => {
                if (e.key === 'Enter' && !runMutation.isPending) handleRun()
              }}
            />
          </div>
          <div className="flex items-center gap-6">
            <div className="w-40 space-y-2">
              <label className="text-[10px] font-black text-text-faint uppercase tracking-[0.2em] block">Probe Count</label>
              <input
                type="number"
                value={count}
                onChange={(e) => setCount(Math.max(1, Math.min(20, parseInt(e.target.value) || 1)))}
                min={1}
                max={20}
                className="w-full bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-xl px-5 py-3 text-sm font-black text-text font-mono tabular-nums focus:outline-none focus:border-accent transition-all shadow-inner"
              />
            </div>
            <div className="flex-1" />
            <button
              onClick={handleRun}
              disabled={runMutation.isPending || !targetsInput.trim()}
              className={cn(
                'flex items-center gap-3 px-8 py-3.5 text-xs font-black uppercase tracking-widest rounded-xl transition-all shadow-xl',
                runMutation.isPending
                  ? 'bg-panel-3 text-text-faint border border-border cursor-not-allowed'
                  : 'bg-accent text-white hover:opacity-90 active:scale-95 shadow-[0_15px_30px_rgba(var(--color-accent-rgb),0.3)]',
              )}
            >
              {runMutation.isPending ? (
                <>
                  <Loader2 size={16} className="animate-spin" />
                  PROBING...
                </>
              ) : (
                <>
                  <Play size={16} fill="currentColor" />
                  INITIATE CLUSTER PING
                </>
              )}
            </button>
          </div>
        </div>
      </Panel>

      {/* ── Error ── */}
      {error && (
        <div className="flex items-center gap-3 bg-danger/10 border border-danger/30 rounded-2xl p-5 text-danger font-black uppercase text-xs animate-in shake duration-500">
          <AlertTriangle size={18} className="shrink-0" />
          {error}
        </div>
      )}

      {/* ── Summary Stats ── */}
      {stats && (
        <Panel variant="elevated" padding="lg">
          <PanelHeader icon={<Activity size={20} />} title="Network Cluster Summary" />
          <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
            <div className="bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-2xl p-5 shadow-inner">
              <div className="flex items-center gap-2 mb-2 opacity-50">
                <Clock size={12} />
                <p className="text-[9px] font-black uppercase tracking-widest">Avg Latency</p>
              </div>
              <p className={cn(
                'text-xl font-black tabular-nums leading-none',
                stats.avg_latency < 50 ? 'text-success' : stats.avg_latency < 150 ? 'text-warning' : 'text-danger',
              )}>
                {stats.avg_latency.toFixed(1)}<span className="text-[10px] ml-0.5">ms</span>
              </p>
            </div>
            <div className="bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-2xl p-5 shadow-inner">
              <div className="flex items-center gap-2 mb-2 opacity-50">
                <ArrowUpDown size={12} />
                <p className="text-[9px] font-black uppercase tracking-widest">Peak Latency</p>
              </div>
              <p className={cn(
                'text-xl font-black tabular-nums leading-none',
                stats.max_latency < 100 ? 'text-success' : stats.max_latency < 300 ? 'text-warning' : 'text-danger',
              )}>
                {stats.max_latency.toFixed(1)}<span className="text-[10px] ml-0.5">ms</span>
              </p>
            </div>
            <div className="bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-2xl p-5 shadow-inner">
              <div className="flex items-center gap-2 mb-2 opacity-50">
                <AlertTriangle size={12} />
                <p className="text-[9px] font-black uppercase tracking-widest">Drop Rate</p>
              </div>
              <p className={cn(
                'text-xl font-black tabular-nums leading-none',
                stats.total_loss === 0 ? 'text-success' : stats.total_loss < 10 ? 'text-warning' : 'text-danger',
              )}>
                {stats.total_loss.toFixed(1)}%
              </p>
            </div>
            <div className="bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-2xl p-5 shadow-inner">
              <div className="flex items-center gap-2 mb-2 opacity-50">
                <Target size={12} />
                <p className="text-[9px] font-black uppercase tracking-widest">Worst Node</p>
              </div>
              <p className="text-sm font-black text-danger truncate uppercase tracking-tight">{stats.worst_target || 'N/A'}</p>
            </div>
          </div>
        </Panel>
      )}

      {/* ── Per-Target Results ── */}
      {results.length > 0 && (
        <Panel variant="default" padding="md">
          <PanelHeader icon={<Target size={20} />} title="Node Inspection" action={<span className="text-[10px] font-black text-text-faint uppercase tracking-widest">{results.length} Targets</span>} />
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {results.map((result) => (
              <ResultCard key={result.target} result={result} />
            ))}
          </div>
        </Panel>
      )}
    </div>
  )
}
