import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import {
  Play,
  ListChecks,
  Info,
  HelpCircle,
  CheckCircle2,
  XCircle,
  Loader2,
  Target,
  Clock,
  Activity,
  AlertTriangle,
  ArrowUpDown,
} from 'lucide-react'
import { SectionHeader } from '@/components/ui/SectionHeader'
import type { PingResultMultiData, PingStatsData } from '@/types'

function ResultCard({ result }: { result: PingResultMultiData }) {
  return (
    <div className={cn(
      'bg-panel-2 border rounded-xl p-4 transition-all',
      result.success ? 'border-success/30 hover:border-success/50' : 'border-danger/30 hover:border-danger/50',
    )}>
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2 min-w-0">
          <Target size={14} className="text-accent shrink-0" />
          <span className="text-xs font-bold text-text uppercase tracking-wider truncate">{result.target}</span>
        </div>
        {result.success ? (
          <span className="flex items-center gap-1 text-[10px] font-bold text-success">
            <CheckCircle2 size={12} /> OK
          </span>
        ) : (
          <span className="flex items-center gap-1 text-[10px] font-bold text-danger">
            <XCircle size={12} /> FAIL
          </span>
        )}
      </div>

      {result.success ? (
        <div className="grid grid-cols-2 gap-3">
          <div>
            <p className="text-[10px] font-bold text-text-faint uppercase tracking-wider mb-0.5">Avg Latency</p>
            <p className={cn(
              'text-sm font-bold tabular-nums',
              result.avg_ms < 50 ? 'text-success' : result.avg_ms < 150 ? 'text-warning' : 'text-danger',
            )}>
              {result.avg_ms.toFixed(1)}ms
            </p>
          </div>
          <div>
            <p className="text-[10px] font-bold text-text-faint uppercase tracking-wider mb-0.5">Jitter</p>
            <p className={cn(
              'text-sm font-bold tabular-nums',
              result.jitter_ms < 10 ? 'text-success' : result.jitter_ms < 30 ? 'text-warning' : 'text-danger',
            )}>
              {result.jitter_ms.toFixed(1)}ms
            </p>
          </div>
          <div>
            <p className="text-[10px] font-bold text-text-faint uppercase tracking-wider mb-0.5">Packet Loss</p>
            <p className={cn(
              'text-sm font-bold tabular-nums',
              result.packet_loss === 0 ? 'text-success' : result.packet_loss < 25 ? 'text-warning' : 'text-danger',
            )}>
              {result.packet_loss.toFixed(1)}%
            </p>
          </div>
          <div>
            <p className="text-[10px] font-bold text-text-faint uppercase tracking-wider mb-0.5">StdDev</p>
            <p className={cn(
              'text-sm font-bold tabular-nums',
              result.stddev_ms < 10 ? 'text-success' : result.stddev_ms < 30 ? 'text-warning' : 'text-danger',
            )}>
              {result.stddev_ms.toFixed(1)}ms
            </p>
          </div>
        </div>
      ) : (
        <div className="flex items-center gap-2 mt-1">
          <AlertTriangle size={12} className="text-danger shrink-0" />
          <p className="text-xs font-medium text-danger">{result.error || 'Ping failed'}</p>
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
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
        <div className="flex items-start gap-3">
          <div className="w-9 h-9 rounded-lg flex items-center justify-center bg-panel-3 border border-border shrink-0 mt-0.5">
            <Info size={18} className="text-accent" />
          </div>
          <p className="text-sm text-text-dim leading-relaxed">
            Multi-Ping — Simultaneously ping multiple targets to compare latency and loss across different endpoints. Ideal for benchmarking and failover testing.
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
            'Concurrent pings reveal path-specific issues',
            'StdDev indicates consistency of response times',
            'Jitter measures inter-packet delay variation',
            'Aggregate stats highlight the worst-performing target',
          ].map(item => (
            <div key={item} className="flex items-center gap-2.5 bg-panel-2 border border-border rounded-xl px-4 py-2.5">
              <HelpCircle size={14} className="text-accent shrink-0" />
              <span className="text-xs font-medium text-text-dim">{item}</span>
            </div>
          ))}
        </div>
      </div>

      {/* ── Controls ── */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
        <SectionHeader icon={<Target size={18} className="text-accent" />} title="Multi-Ping Controls" />
        <div className="space-y-4">
          <div>
            <label className="text-[10px] font-bold text-text-faint uppercase tracking-widest mb-1.5 block">Targets (comma-separated)</label>
            <input
              type="text"
              value={targetsInput}
              onChange={(e) => setTargetsInput(e.target.value)}
              placeholder="8.8.8.8, 1.1.1.1, google.com"
              className="w-full bg-panel-2 border border-border rounded-xl px-4 py-2.5 text-sm text-text font-mono placeholder:text-text-faint focus:outline-none focus:border-accent/50 transition-colors"
              onKeyDown={(e) => {
                if (e.key === 'Enter' && !runMutation.isPending) handleRun()
              }}
            />
          </div>
          <div className="flex items-center gap-4">
            <div className="w-32">
              <label className="text-[10px] font-bold text-text-faint uppercase tracking-widest mb-1.5 block">Count</label>
              <input
                type="number"
                value={count}
                onChange={(e) => setCount(Math.max(1, Math.min(20, parseInt(e.target.value) || 1)))}
                min={1}
                max={20}
                className="w-full bg-panel-2 border border-border rounded-xl px-4 py-2.5 text-sm text-text font-mono tabular-nums focus:outline-none focus:border-accent/50 transition-colors"
              />
            </div>
            <div className="flex-1" />
            <button
              onClick={handleRun}
              disabled={runMutation.isPending || !targetsInput.trim()}
              className={cn(
                'flex items-center gap-3 px-6 py-2.5 text-sm font-semibold rounded-xl transition-all shadow-xl mt-5',
                runMutation.isPending
                  ? 'bg-panel-3 text-text-faint border border-border cursor-not-allowed'
                  : 'bg-accent text-white hover:bg-accent/90',
              )}
            >
              {runMutation.isPending ? (
                <>
                  <Loader2 size={16} className="animate-spin" />
                  Running...
                </>
              ) : (
                <>
                  <Play size={16} fill="currentColor" />
                  Run Multi-Ping
                </>
              )}
            </button>
          </div>
        </div>
      </div>

      {/* ── Error ── */}
      {error && (
        <div className="flex items-center gap-3 bg-danger/10 border border-danger/30 rounded-[var(--radius-lg)] p-4">
          <AlertTriangle size={16} className="text-danger shrink-0" />
          <p className="text-sm font-medium text-danger">{error}</p>
        </div>
      )}

      {/* ── Summary Stats ── */}
      {stats && (
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
          <SectionHeader icon={<Activity size={18} className="text-accent" />} title="Aggregate Summary" />
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="bg-panel-2 border border-border rounded-xl p-4">
              <div className="flex items-center gap-2 mb-1">
                <Clock size={12} className="text-text-faint" />
                <p className="text-[10px] font-bold text-text-faint uppercase tracking-wider">Avg Latency</p>
              </div>
              <p className={cn(
                'text-lg font-bold tabular-nums',
                stats.avg_latency < 50 ? 'text-success' : stats.avg_latency < 150 ? 'text-warning' : 'text-danger',
              )}>
                {stats.avg_latency.toFixed(1)}ms
              </p>
            </div>
            <div className="bg-panel-2 border border-border rounded-xl p-4">
              <div className="flex items-center gap-2 mb-1">
                <ArrowUpDown size={12} className="text-text-faint" />
                <p className="text-[10px] font-bold text-text-faint uppercase tracking-wider">Max Latency</p>
              </div>
              <p className={cn(
                'text-lg font-bold tabular-nums',
                stats.max_latency < 100 ? 'text-success' : stats.max_latency < 300 ? 'text-warning' : 'text-danger',
              )}>
                {stats.max_latency.toFixed(1)}ms
              </p>
            </div>
            <div className="bg-panel-2 border border-border rounded-xl p-4">
              <div className="flex items-center gap-2 mb-1">
                <AlertTriangle size={12} className="text-text-faint" />
                <p className="text-[10px] font-bold text-text-faint uppercase tracking-wider">Total Loss</p>
              </div>
              <p className={cn(
                'text-lg font-bold tabular-nums',
                stats.total_loss === 0 ? 'text-success' : stats.total_loss < 10 ? 'text-warning' : 'text-danger',
              )}>
                {stats.total_loss.toFixed(1)}%
              </p>
            </div>
            <div className="bg-panel-2 border border-border rounded-xl p-4">
              <div className="flex items-center gap-2 mb-1">
                <Target size={12} className="text-text-faint" />
                <p className="text-[10px] font-bold text-text-faint uppercase tracking-wider">Worst Target</p>
              </div>
              <p className="text-sm font-bold text-danger truncate">{stats.worst_target || 'N/A'}</p>
            </div>
          </div>
        </div>
      )}

      {/* ── Per-Target Results ── */}
      {results.length > 0 && (
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
          <SectionHeader icon={<Target size={18} className="text-accent" />} title="Per-Target Results" />
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {results.map((result) => (
              <ResultCard key={result.target} result={result} />
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
