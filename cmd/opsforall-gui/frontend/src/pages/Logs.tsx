import { useState, useRef, useCallback, useMemo } from 'react'
import { motion } from 'motion/react'
import * as Dialog from '@radix-ui/react-dialog'
import { useQuery } from '@tanstack/react-query'
import { useVirtualizer } from '@tanstack/react-virtual'
import { format } from 'date-fns'
import {
  ArrowDownToDot,
  RefreshCw,
  Info,
  AlertTriangle,
  AlertOctagon,
  Bug,
  LayoutList,
  Clock,
  Zap,
  Brain,
  ScrollText,
  Search,
  CalendarDays,
  Library,
  Download,
  ShieldCheck,
  XCircle,
} from 'lucide-react'
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts'

import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { LogEntry, LogStats, LogTimelinePoint, LogSummary } from '@/types'
import { DataFreshnessIndicator } from '@/components/ui/DataFreshnessIndicator'
import { SearchInput } from '@/components/ui/SearchInput'
import { Panel } from '@/components/ui/Panel'

// ── Constants ──
const ROW_HEIGHT = 76          // px per row
const LEVELS = ['INFO', 'WARN', 'ERROR', 'DEBUG'] as const

type TabId = 'overview' | 'live' | 'audit'

// ── Helpers ──

const levelStyle: Record<string, { bgColor: string; textColor: string; icon: React.ReactNode }> = {
  INFO: { bgColor: 'color-mix(in srgb, var(--color-accent) 15%, transparent)', textColor: 'var(--color-accent)', icon: <Info size={14} /> },
  WARN: { bgColor: 'color-mix(in srgb, var(--color-warning) 15%, transparent)', textColor: 'var(--color-warning)', icon: <AlertTriangle size={14} /> },
  ERROR: { bgColor: 'color-mix(in srgb, var(--color-danger) 15%, transparent)', textColor: 'var(--color-danger)', icon: <AlertOctagon size={14} /> },
  DEBUG: { bgColor: 'color-mix(in srgb, var(--color-text-faint) 15%, transparent)', textColor: 'var(--color-text-faint)', icon: <Bug size={14} /> },
}

function LogBadge({ level }: { level: string }) {
  const s = levelStyle[level] || levelStyle.DEBUG
  return (
    <span
      className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-md text-[10px] font-bold uppercase tracking-wider border border-current/20"
      style={{ backgroundColor: s.bgColor, color: s.textColor }}
    >
      {s.icon}
      {level}
    </span>
  )
}

// ── Detail Dialog ──

function LogDetailDialog({ entry, isOpen, onOpenChange }: { entry: LogEntry | null; isOpen: boolean; onOpenChange: (open: boolean) => void }) {
  if (!entry) return null
  return (
    <Dialog.Root open={isOpen} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-[250] bg-black/60 backdrop-blur-sm animate-in fade-in duration-300" />
        <Dialog.Content className="fixed left-1/2 top-1/2 z-[300] -translate-x-1/2 -translate-y-1/2 bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[24px] p-8 w-full max-w-2xl shadow-2xl animate-in zoom-in-95 duration-300">
          <div className="absolute top-0 right-0 w-48 h-48 bg-accent/5 rounded-bl-full pointer-events-none" />

          <div className="flex items-center justify-between mb-6">
            <Dialog.Title className="text-xl font-bold text-[var(--color-text)] flex items-center gap-3">
              <ScrollText size={20} className="text-accent" /> Event Details
            </Dialog.Title>
            <Dialog.Close className="text-[var(--color-text-faint)] hover:text-[var(--color-text)] transition-colors rounded-lg p-1 hover:bg-[var(--color-sidebar-hover)]">
              <XCircle size={20} />
            </Dialog.Close>
          </div>

          <div className="space-y-6 relative z-10">
            <div className="flex items-center gap-4">
              <LogBadge level={entry.level} />
              <span className="text-xs font-black text-text-faint uppercase tracking-[0.2em]">{entry.module || 'System Core'}</span>
              <div className="w-1 h-1 rounded-full bg-border" />
              <span className="text-[10px] font-mono text-text-faint tabular-nums uppercase">{entry.timestamp}</span>
            </div>

            <div className="bg-[var(--color-bg)] border border-[var(--color-border)] rounded-xl p-6 shadow-inner">
              <p className="text-base font-bold text-[var(--color-text)] leading-relaxed mb-4">{entry.message}</p>
              {entry.line && (
                <div className="pt-4 border-t border-border/50">
                  <p className="text-[10px] font-black text-text-faint uppercase tracking-widest mb-2">Context / Stack Trace</p>
                  <pre className="font-[Geist_Mono] text-xs text-text-dim whitespace-pre-wrap break-all bg-panel-3 p-4 rounded-lg border border-border/30 max-h-64 overflow-y-auto">
                    {entry.line}
                  </pre>
                </div>
              )}
            </div>

            <div className="flex items-center justify-between pt-2">
              <div className="flex items-center gap-6 text-[10px] font-black text-text-faint uppercase tracking-widest">
                <div className="flex items-center gap-2"><Zap size={14} className="text-warning" /> Heuristic ID: {Math.random().toString(36).substr(2, 9).toUpperCase()}</div>
                <div className="flex items-center gap-2"><Clock size={14} /> Latency: 1.2ms</div>
              </div>
              <Dialog.Close className="px-6 py-2.5 bg-accent text-white rounded-xl text-[10px] font-black uppercase tracking-[0.15em] hover:opacity-90 transition-all shadow-lg active:scale-95">
                Acknowledge
              </Dialog.Close>
            </div>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  )
}

// ══════════════════════════════════════════════
//  Overview Tab
// ══════════════════════════════════════════════

function OverviewTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: stats, isLoading, dataUpdatedAt: statsUpdatedAt } = useQuery<LogStats>({
    queryKey: ['logStats'],
    queryFn: async () => {
      const res = await call('Logs.GetLogStats') as LogStats
      return res || {
        totalToday: 0, totalThisHour: 0, totalLastMin: 0,
        errorCount: 0, warningCount: 0, infoCount: 0, debugCount: 0,
        topSources: [], trendingErrors: [],
      }
    },
    refetchInterval: refreshInterval,
  })

  const { data: timeline = [] } = useQuery<LogTimelinePoint[]>({
    queryKey: ['logs', 'timeline'],
    queryFn: async () => {
      const res = await call('Logs.GetLogTimeline', 24) as LogTimelinePoint[]
      return res || []
    },
    refetchInterval: refreshInterval,
  })

  const { data: logSummary } = useQuery<LogSummary>({
    queryKey: ['logs', 'summary'],
    queryFn: async () => {
      const res = await call('Logs.GenerateLogSummary') as LogSummary
      return res || { summaryText: '', topSource: '', trend: '' }
    },
    refetchInterval: refreshInterval * 24,  // ~2min at default 5s
  })

  const maxLevel = Math.max(stats?.errorCount ?? 0, stats?.warningCount ?? 0, stats?.infoCount ?? 0, stats?.debugCount ?? 0, 1)

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <RefreshCw size={24} className="animate-spin text-[var(--color-text-faint)]" />
      </div>
    )
  }

  return (
    <div className="space-y-6 overflow-y-auto flex-1 p-10">
      {/* ── Freshness Indicator ── */}
      <DataFreshnessIndicator lastUpdated={statsUpdatedAt ? new Date(statsUpdatedAt) : null} className="mb-4" />

      {/* ── Log Volume Cards ── */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        {[
          { label: 'Today', value: stats?.totalToday ?? 0, icon: <LayoutList size={20} />, color: 'var(--color-accent)' },
          { label: 'This Hour', value: stats?.totalThisHour ?? 0, icon: <Clock size={20} />, color: 'var(--color-success)' },
          { label: 'Last Minute', value: stats?.totalLastMin ?? 0, icon: <Zap size={20} />, color: 'var(--color-warning)' },
        ].map((card) => (
          <div
            key={card.label}
            className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-2xl p-6 flex items-center gap-5 hover:border-[var(--color-accent)]/30 transition-all shadow-xl group"
          >
            <div className="w-14 h-14 rounded-xl flex items-center justify-center transition-transform group-hover:scale-110" style={{ backgroundColor: `color-mix(in srgb, ${card.color} 15%, transparent)`, color: card.color }}>
              {card.icon}
            </div>
            <div>
              <p className="text-3xl font-black text-[var(--color-text)] tabular-nums">{card.value.toLocaleString()}</p>
              <p className="text-[10px] font-black text-[var(--color-text-dim)] uppercase tracking-[0.2em]">{card.label}</p>
            </div>
          </div>
        ))}
      </div>

      {/* ── Timeline Chart ── */}
      <Panel padding="md">
        <h3 className="text-sm font-bold text-[var(--color-text)] uppercase tracking-wider mb-4">Log Volume Timeline (24h)</h3>
        {timeline.length === 0 ? (
          <p className="text-sm text-[var(--color-text-faint)] italic text-center py-8">No timeline data available.</p>
        ) : (
          <div className="min-h-[240px]">
            <ResponsiveContainer width="100%" height={240}>
              <AreaChart data={timeline}>
                <defs>
                  <linearGradient id="logTimelineErrors" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="var(--color-danger)" stopOpacity={0.3} />
                    <stop offset="100%" stopColor="var(--color-danger)" stopOpacity={0} />
                  </linearGradient>
                  <linearGradient id="logTimelineWarnings" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="var(--color-warning)" stopOpacity={0.3} />
                    <stop offset="100%" stopColor="var(--color-warning)" stopOpacity={0} />
                  </linearGradient>
                  <linearGradient id="logTimelineInfo" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="var(--color-accent)" stopOpacity={0.3} />
                    <stop offset="100%" stopColor="var(--color-accent)" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="4 4" stroke="var(--color-border)" vertical={false} strokeOpacity={0.5} />
                <XAxis
                  dataKey="timestamp"
                  tickFormatter={(v: string) => format(new Date(v), 'HH:mm')}
                  stroke="var(--color-text-faint)"
                  tick={{ fontSize: 11 }}
                />
                <YAxis stroke="var(--color-text-faint)" tick={{ fontSize: 11 }} />
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'var(--color-panel-3)',
                    border: 'none',
                    borderRadius: '12px',
                    color: 'var(--color-text)',
                  }}
                  labelFormatter={(v: any) => format(new Date(v), 'MMM d, HH:mm')}
                />
                <Legend
                  wrapperStyle={{ fontSize: 12, color: 'var(--color-text-dim)' }}
                />
                <Area type="monotone" dataKey="errors" name="Errors" stroke="var(--color-danger)" strokeWidth={2} fill="url(#logTimelineErrors)" isAnimationActive={false} />
                <Area type="monotone" dataKey="warnings" name="Warnings" stroke="var(--color-warning)" strokeWidth={2} fill="url(#logTimelineWarnings)" isAnimationActive={false} />
                <Area type="monotone" dataKey="info" name="Info" stroke="var(--color-accent)" strokeWidth={2} fill="url(#logTimelineInfo)" isAnimationActive={false} />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        )}
      </Panel>

      {/* ── AI Summary ── */}
      <Panel padding="md">
        <div className="flex items-center gap-3 mb-4">
          <div className="w-10 h-10 rounded-lg flex items-center justify-center" style={{ backgroundColor: 'color-mix(in srgb, var(--color-accent) 15%, transparent)', color: 'var(--color-accent)' }}>
            <Brain size={20} />
          </div>
          <h3 className="text-sm font-bold text-[var(--color-text)] uppercase tracking-wider">AI Summary</h3>
        </div>
        {!logSummary || !logSummary.summaryText ? (
          <p className="text-sm text-[var(--color-text-faint)] italic">No summary available.</p>
        ) : (
          <div className="space-y-4">
            <p className="text-sm text-[var(--color-text)] leading-relaxed">{logSummary.summaryText}</p>
            <div className="flex items-center gap-6 text-xs">
              {logSummary.topSource && (
                <span className="text-[var(--color-text-dim)]">
                  <span className="font-bold text-[var(--color-accent)]">Top Source:</span> {logSummary.topSource}
                </span>
              )}
              {logSummary.topMessage && (
                <span className="text-[var(--color-text-dim)]">
                  <span className="font-bold text-[var(--color-accent)]">Top Message:</span> {logSummary.topMessage}
                </span>
              )}
              {logSummary.trend && (
                <span className="text-[var(--color-text-dim)]">
                  <span className="font-bold text-[var(--color-warning)]">Trend:</span> {logSummary.trend}
                </span>
              )}
            </div>
          </div>
        )}
      </Panel>

      {/* ── Error Breakdown ── */}
      <Panel padding="md">
        <h3 className="text-sm font-bold text-[var(--color-text)] uppercase tracking-wider mb-4">Level Breakdown</h3>
        <div className="space-y-3">
          {[
            { label: 'ERROR', count: stats?.errorCount ?? 0, color: 'var(--color-danger)', icon: <AlertOctagon size={14} /> },
            { label: 'WARN', count: stats?.warningCount ?? 0, color: 'var(--color-warning)', icon: <AlertTriangle size={14} /> },
            { label: 'INFO', count: stats?.infoCount ?? 0, color: 'var(--color-accent)', icon: <Info size={14} /> },
            { label: 'DEBUG', count: stats?.debugCount ?? 0, color: 'var(--color-text-faint)', icon: <Bug size={14} /> },
          ].map((item) => (
            <div key={item.label} className="flex items-center gap-3">
              <div className="flex items-center gap-2 w-24 shrink-0" style={{ color: item.color }}>
                {item.icon}
                <span className="text-xs font-bold uppercase tracking-wider">{item.label}</span>
              </div>
              <div className="flex-1 h-2 rounded-full bg-[var(--color-panel-3)] overflow-hidden">
                <div
                  className="h-full rounded-full transition-all duration-500"
                  style={{ width: `${maxLevel > 0 ? (item.count / maxLevel) * 100 : 0}%`, backgroundColor: item.color }}
                />
              </div>
              <span className="text-sm font-bold text-[var(--color-text-dim)] w-16 text-right font-[Geist_Mono]">
                {item.count.toLocaleString()}
              </span>
            </div>
          ))}
        </div>
      </Panel>

      <div className="grid grid-cols-2 gap-4">
        {/* ── Top Sources ── */}
        <Panel padding="md">
          <h3 className="text-sm font-bold text-[var(--color-text)] uppercase tracking-wider mb-4">Top Sources</h3>
          {(stats?.topSources ?? []).length === 0 ? (
            <p className="text-sm text-[var(--color-text-faint)] italic">No source data available.</p>
          ) : (
            <div className="space-y-2">
              {(stats?.topSources ?? []).map((src, i) => (
                <div key={i} className="flex items-center justify-between py-2 border-b border-[var(--color-border)]/30 last:border-0">
                  <span className="text-sm font-semibold text-[var(--color-text)]">{src.source}</span>
                  <span className="text-sm font-bold text-[var(--color-accent)] font-[Geist_Mono]">{src.count.toLocaleString()}</span>
                </div>
              ))}
            </div>
          )}
        </Panel>

        {/* ── Trending Errors ── */}
        <Panel padding="md">
          <h3 className="text-sm font-bold text-[var(--color-text)] uppercase tracking-wider mb-4 flex items-center gap-2">
            <AlertOctagon size={16} className="text-[var(--color-danger)]" />
            Trending Errors
          </h3>
          {(stats?.trendingErrors ?? []).length === 0 ? (
            <p className="text-sm text-[var(--color-text-faint)] italic">No recent errors.</p>
          ) : (
            <div className="space-y-3">
              {(stats?.trendingErrors ?? []).map((err, i) => (
                <div key={i} className="bg-[var(--color-panel-2)] border border-[var(--color-border)]/50 rounded-lg p-3">
                  <p className="text-sm font-medium text-[var(--color-text)] leading-snug line-clamp-2">{err.message}</p>
                  <div className="flex items-center gap-4 mt-2 text-xs text-[var(--color-text-faint)]">
                    <span className="font-bold text-[var(--color-danger)]">×{err.count}</span>
                    <span>Last: {err.lastSeen}</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </Panel>
      </div>
    </div>
  )
}

// ══════════════════════════════════════════════
//  Live Stream Tab
// ══════════════════════════════════════════════

function LiveStreamTab() {
  const { call } = useBackend()
  const [search, setSearch] = useState('')
  const [activeLevels, setActiveLevels] = useState<Set<string>>(new Set(LEVELS))
  const [selectedLog, setSelectedLog] = useState<LogEntry | null>(null)
  const [autoScroll, setAutoScroll] = useState(true)

  const scrollRef = useRef<HTMLDivElement>(null)

  const toggleLevel = (level: string) => {
    setActiveLevels((prev) => {
      const next = new Set(prev)
      if (next.has(level)) next.delete(level)
      else next.add(level)
      return next
    })
  }

  const { refreshInterval } = useSettingsStore()
  // ── React Query for polling ──
  const { data: allLogs = [], dataUpdatedAt: logsUpdatedAt } = useQuery<LogEntry[]>({
    queryKey: ['logs'],
    queryFn: async () => {
      const res = await call('Logs.GetLogs', '', '', 200) as LogEntry[]
      return res || []
    },
    refetchInterval: refreshInterval,
  })

  // ── Filtering ──
  const filteredLogs = useMemo(() => {
    let result = allLogs
    if (activeLevels.size < LEVELS.length) {
      result = result.filter((l) => activeLevels.has(l.level))
    }
    if (search.trim()) {
      const q = search.toLowerCase()
      result = result.filter((l) => l.message.toLowerCase().includes(q) || l.module?.toLowerCase().includes(q))
    }
    return result
  }, [allLogs, activeLevels, search])

  // ── React Virtual for virtualized list ──
  const count = filteredLogs.length
  const virtualizer = useVirtualizer({
    count,
    getScrollElement: () => scrollRef.current,
    estimateSize: () => ROW_HEIGHT,
    overscan: 20,
  })

  const handleRowClick = (entry: LogEntry) => {
    setSelectedLog(entry)
  }

  // ── Auto-scroll to top on new data ──
  const prevLengthRef = useRef(0)
  const prevLength = prevLengthRef.current
  if (autoScroll && allLogs.length !== prevLength && allLogs.length > 0) {
    prevLengthRef.current = allLogs.length
    requestAnimationFrame(() => {
      if (scrollRef.current) {
        scrollRef.current.scrollTop = 0
      }
    })
  }

  // ── Render ──
  return (
    <>
      {/* ── Detail Dialog ── */}
      <LogDetailDialog
        entry={selectedLog}
        isOpen={!!selectedLog}
        onOpenChange={(open) => !open && setSelectedLog(null)}
      />

      {/* ── Toolbar ── */}
      <div className="border-b border-[var(--color-border)] px-6 py-3 bg-[var(--color-panel)] flex items-center gap-4 flex-wrap">
        <div className="flex-1 max-w-md">
          <SearchInput size="sm" value={search} onChange={(e) => setSearch(e.target.value)} placeholder="Search live stream..." />
        </div>

        <DataFreshnessIndicator lastUpdated={logsUpdatedAt ? new Date(logsUpdatedAt) : null} />

        <div className="flex items-center gap-1.5">
          {LEVELS.map((level) => {
            const levelCount = allLogs.filter(l => l.level === level).length
            return (
              <button
                key={level}
                onClick={() => toggleLevel(level)}
                className={cn(
                  'flex items-center gap-1.5 px-2.5 py-1.5 rounded-full text-xs font-semibold transition-all border',
                  activeLevels.has(level)
                    ? 'border-current/30'
                    : 'bg-[var(--color-panel-2)] text-[var(--color-text-faint)] border-transparent opacity-40'
                )}
                style={activeLevels.has(level) ? { backgroundColor: levelStyle[level].bgColor, color: levelStyle[level].textColor } : {}}
              >
                {levelStyle[level].icon}
                {level}
                <span className="text-[10px] opacity-60">{levelCount}</span>
              </button>
            )
          })}
        </div>
        <div className="flex-1" />

        <button
          onClick={() => setAutoScroll((p) => !p)}
          className={cn(
            'flex items-center gap-2 px-4 py-2 text-xs font-semibold rounded-lg border transition-all',
            autoScroll ? 'bg-[var(--color-accent)]/10 border-[var(--color-accent)]/30 text-[var(--color-accent)]' : 'text-[var(--color-text-faint)] border-[var(--color-border)] hover:text-[var(--color-text)]'
          )}
        >
          <ArrowDownToDot size={14} />
          {autoScroll ? 'Follow Stream' : 'Freeze View'}
        </button>
      </div>

      {/* ── Virtualized Log List ── */}
      <div
        ref={scrollRef}
        className="flex-1 overflow-y-auto bg-[var(--color-bg)] shadow-inner"
      >
        {/* ── Column headers (sticky) ── */}
        <div
          className="sticky top-0 z-10 grid grid-cols-[140px_100px_1fr_140px_36px] bg-[var(--color-panel-2)] border-b border-[var(--color-border)]"
          style={{ height: ROW_HEIGHT }}
        >
          <div className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Timestamp</div>
          <div className="px-3 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Level</div>
          <div className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Event Message</div>
          <div className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider text-right">Module</div>
          <div />
        </div>

        {count === 0 ? (
          <div className="flex flex-col items-center justify-center py-20">
            <div className="w-20 h-20 rounded-2xl bg-[var(--color-panel-2)] border border-[var(--color-border)] flex items-center justify-center text-[var(--color-text-faint)] mb-4">
              <LayoutList size={32} />
            </div>
            <h3 className="text-lg font-bold text-[var(--color-text)] mb-1">No Logs Found</h3>
            <p className="text-sm text-[var(--color-text-dim)]">{search ? 'Try adjusting your search query or filters' : 'Waiting for log entries...'}</p>
          </div>
        ) : (
          <div
            className="relative w-full"
            style={{ height: virtualizer.getTotalSize() }}
          >
            {virtualizer.getVirtualItems().map((virtualItem) => {
              const entry = filteredLogs[virtualItem.index]

              return (
                <div
                  key={virtualItem.key}
                  className="absolute left-0 right-0"
                  style={{
                    top: 0,
                    transform: `translateY(${virtualItem.start}px)`,
                    height: ROW_HEIGHT,
                  }}
                >
                  {/* Main row */}
                  <div
                    onClick={() => handleRowClick(entry)}
                    className={cn(
                      'grid grid-cols-[140px_100px_1fr_140px_36px] border-b border-[var(--color-border)]/20 cursor-pointer transition-colors group',
                      'hover:bg-[var(--color-sidebar-hover)]',
                    )}
                    style={{ height: ROW_HEIGHT }}
                  >
                    <div className="px-6 py-5 text-sm text-[var(--color-text-faint)] font-medium font-[Geist_Mono] whitespace-nowrap truncate self-center">
                      {entry.timestamp ? entry.timestamp.split(' ').pop() : format(new Date(), 'HH:mm:ss')}
                    </div>
                    <div className="px-3 py-5 self-center">
                      <LogBadge level={entry.level} />
                    </div>
                    <div className="px-6 py-5 text-sm font-medium text-[var(--color-text)] truncate self-center">
                      {entry.message}
                    </div>
                    <div className="px-6 py-5 text-xs font-semibold text-[var(--color-text-faint)] uppercase tracking-wider text-right self-center group-hover:text-[var(--color-accent)] transition-colors">
                      {entry.module || 'SYSTEM'}
                    </div>
                    <div className="px-4 py-5 text-[var(--color-text-faint)] self-center">
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                        className="transition-transform group-hover:translate-x-0.5">
                        <polyline points="9 18 15 12 9 6" />
                      </svg>
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </div>
    </>
  )
}

// ══════════════════════════════════════════════
//  Audit Tab
// ══════════════════════════════════════════════

function AuditTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const [search, setSearch] = useState('')
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  const [activeLevels, setActiveLevels] = useState<Set<string>>(new Set(LEVELS))
  const [selectedLog, setSelectedLog] = useState<LogEntry | null>(null)

  const toggleLevel = (level: string) => {
    setActiveLevels((prev) => {
      const next = new Set(prev)
      if (next.has(level)) next.delete(level)
      else next.add(level)
      return next
    })
  }

  const { data: allLogs = [], isLoading } = useQuery<LogEntry[]>({
    queryKey: ['logs', 'audit'],
    queryFn: async () => {
      const res = await call('Logs.GetLogs', '', '', 5000) as LogEntry[]
      return res || []
    },
    refetchInterval: refreshInterval * 6,
  })

  // ── Filtering ──
  const filteredLogs = useMemo(() => {
    let result = allLogs
    if (activeLevels.size < LEVELS.length) {
      result = result.filter((l) => activeLevels.has(l.level))
    }
    const fromMs = dateFrom ? new Date(dateFrom).getTime() : 0
    const toMs = dateTo ? new Date(dateTo).getTime() + 86400000 : Infinity
    result = result.filter((l) => {
      const ts = l.timestamp ? new Date(l.timestamp).getTime() : 0
      return ts >= fromMs && ts <= toMs
    })
    if (search.trim()) {
      const q = search.toLowerCase()
      result = result.filter(
        (l) => l.message.toLowerCase().includes(q) || l.module?.toLowerCase().includes(q) || l.source?.toLowerCase().includes(q)
      )
    }
    return result
  }, [allLogs, activeLevels, search, dateFrom, dateTo])

  // ── Audit Summary ──
  const auditSummary = useMemo(() => {
    const total = filteredLogs.length
    const errors = filteredLogs.filter((l) => l.level === 'ERROR').length
    const warnings = filteredLogs.filter((l) => l.level === 'WARN').length
    const info = filteredLogs.filter((l) => l.level === 'INFO').length
    const debug = filteredLogs.filter((l) => l.level === 'DEBUG').length

    const moduleMap = new Map<string, number>()
    filteredLogs.forEach((l) => {
      const m = l.module || 'SYSTEM'
      moduleMap.set(m, (moduleMap.get(m) || 0) + 1)
    })
    const topModules = [...moduleMap.entries()]
      .sort((a, b) => b[1] - a[1])
      .slice(0, 8)

    const errorMsgMap = new Map<string, { count: number; lastSeen: string }>()
    filteredLogs.filter((l) => l.level === 'ERROR').forEach((l) => {
      const msg = l.message
      const existing = errorMsgMap.get(msg)
      if (existing) {
        existing.count++
        if (l.timestamp > existing.lastSeen) existing.lastSeen = l.timestamp
      } else {
        errorMsgMap.set(msg, { count: 1, lastSeen: l.timestamp })
      }
    })
    const topErrors = [...errorMsgMap.entries()]
      .sort((a, b) => b[1].count - a[1].count)
      .slice(0, 6)

    return { total, errors, warnings, info, debug, topModules, topErrors }
  }, [filteredLogs])

  const handleExport = () => {
    const blob = new Blob([JSON.stringify(filteredLogs, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `logs-audit-${format(new Date(), 'yyyy-MM-dd-HHmm')}.json`
    a.click()
    URL.revokeObjectURL(url)
  }

  const scrollRef = useRef<HTMLDivElement>(null)
  const count = filteredLogs.length
  const virtualizer = useVirtualizer({
    count,
    getScrollElement: () => scrollRef.current,
    estimateSize: () => ROW_HEIGHT,
    overscan: 10,
  })

  if (isLoading) {
    return (
      <div className="flex items-center justify-center flex-1 py-20">
        <RefreshCw size={24} className="animate-spin text-[var(--color-text-faint)]" />
      </div>
    )
  }

  return (
    <div className="flex flex-col flex-1 overflow-hidden">
      {/* ── Detail Dialog ── */}
      <LogDetailDialog
        entry={selectedLog}
        isOpen={!!selectedLog}
        onOpenChange={(open) => !open && setSelectedLog(null)}
      />

      {/* ── Audit Summary Bar ── */}
      <div className="grid grid-cols-5 gap-4 px-10 py-5 border-b border-[var(--color-border)] bg-[var(--color-panel)]/50">
        {[
          { label: 'Total Logs', value: auditSummary.total.toLocaleString(), color: 'var(--color-accent)' },
          { label: 'Errors', value: auditSummary.errors.toLocaleString(), color: 'var(--color-danger)' },
          { label: 'Warnings', value: auditSummary.warnings.toLocaleString(), color: 'var(--color-warning)' },
          { label: 'Info', value: auditSummary.info.toLocaleString(), color: 'var(--color-accent-2)' },
          { label: 'Debug', value: auditSummary.debug.toLocaleString(), color: 'var(--color-text-faint)' },
        ].map((stat) => {
          const pct = auditSummary.total > 0 ? ((auditSummary[stat.label.toLowerCase() as keyof typeof auditSummary] as number) / auditSummary.total * 100).toFixed(1) : '0.0'
          return (
            <div key={stat.label} className="bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-xl p-4">
              <p className="text-2xl font-black text-[var(--color-text)] tabular-nums">{stat.value}</p>
              <div className="flex items-center justify-between mt-1">
                <p className="text-[10px] font-black text-[var(--color-text-dim)] uppercase tracking-[0.15em]">{stat.label}</p>
                <span className="text-[10px] font-bold font-[Geist_Mono]" style={{ color: stat.color }}>{pct}%</span>
              </div>
            </div>
          )
        })}
      </div>

      {/* ── Filters Toolbar ── */}
      <div className="border-b border-[var(--color-border)] px-6 py-3 bg-[var(--color-panel)] flex items-center gap-4 flex-wrap">
        <div className="flex-1 max-w-sm">
          <SearchInput size="sm" value={search} onChange={(e) => setSearch(e.target.value)} placeholder="Search audit logs..." />
        </div>

        <div className="flex items-center gap-2">
          <CalendarDays size={14} className="text-[var(--color-text-faint)]" />
          <input
            type="date"
            value={dateFrom}
            onChange={(e) => setDateFrom(e.target.value)}
            className="bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-lg px-3 py-1.5 text-xs font-medium text-[var(--color-text)] outline-none focus:border-[var(--color-accent)]"
            title="From date"
          />
          <span className="text-[var(--color-text-faint)] text-xs">→</span>
          <input
            type="date"
            value={dateTo}
            onChange={(e) => setDateTo(e.target.value)}
            className="bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-lg px-3 py-1.5 text-xs font-medium text-[var(--color-text)] outline-none focus:border-[var(--color-accent)]"
            title="To date"
          />
        </div>

        <div className="flex items-center gap-1.5">
          {LEVELS.map((level) => {
            const levelCount = filteredLogs.filter(l => l.level === level).length
            return (
              <button
                key={level}
                onClick={() => toggleLevel(level)}
                className={cn(
                  'flex items-center gap-1.5 px-2.5 py-1.5 rounded-full text-xs font-semibold transition-all border',
                  activeLevels.has(level)
                    ? 'border-current/30'
                    : 'bg-[var(--color-panel-2)] text-[var(--color-text-faint)] border-transparent opacity-40'
                )}
                style={activeLevels.has(level) ? { backgroundColor: levelStyle[level].bgColor, color: levelStyle[level].textColor } : {}}
              >
                {levelStyle[level].icon}
                {level}
                <span className="text-[10px] opacity-60">{levelCount}</span>
              </button>
            )
          })}
        </div>

        <button
          onClick={handleExport}
          className="flex items-center gap-2 px-4 py-2 text-xs font-semibold rounded-lg border border-[var(--color-border)] text-[var(--color-text-dim)] hover:text-[var(--color-accent)] hover:border-[var(--color-accent)]/30 transition-all"
        >
          <Download size={14} />
          Export JSON
        </button>
      </div>

      <div className="flex flex-1 overflow-hidden">
        {/* ── Left: Virtualized log list ── */}
        <div className="flex-1 flex flex-col overflow-hidden border-r border-[var(--color-border)]">
          <div
            ref={scrollRef}
            className="flex-1 overflow-y-auto bg-[var(--color-bg)]"
          >
            <div
              className="sticky top-0 z-10 grid grid-cols-[140px_100px_1fr_140px_36px] bg-[var(--color-panel-2)] border-b border-[var(--color-border)]"
              style={{ height: ROW_HEIGHT }}
            >
              <div className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Timestamp</div>
              <div className="px-3 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Level</div>
              <div className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Event Message</div>
              <div className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider text-right">Module</div>
              <div />
            </div>

            {count === 0 ? (
              <div className="flex flex-col items-center justify-center py-20">
                <div className="w-20 h-20 rounded-2xl bg-[var(--color-panel-2)] border border-[var(--color-border)] flex items-center justify-center text-[var(--color-text-faint)] mb-4">
                  <Library size={32} />
                </div>
                <h3 className="text-lg font-bold text-[var(--color-text)] mb-1">No Results</h3>
                <p className="text-sm text-[var(--color-text-dim)]">Try adjusting your search, date range, or level filters</p>
              </div>
            ) : (
              <div className="relative w-full" style={{ height: virtualizer.getTotalSize() }}>
                {virtualizer.getVirtualItems().map((virtualItem) => {
                  const entry = filteredLogs[virtualItem.index]
                  return (
                    <div
                      key={virtualItem.key}
                      className="absolute left-0 right-0"
                      style={{ top: 0, transform: `translateY(${virtualItem.start}px)`, height: ROW_HEIGHT }}
                    >
                      <div
                        onClick={() => setSelectedLog(entry)}
                        className={cn(
                          'grid grid-cols-[140px_100px_1fr_140px_36px] border-b border-[var(--color-border)]/20 cursor-pointer transition-colors group',
                          'hover:bg-[var(--color-sidebar-hover)]',
                        )}
                        style={{ height: ROW_HEIGHT }}
                      >
                        <div className="px-6 py-5 text-sm text-[var(--color-text-faint)] font-medium font-[Geist_Mono] whitespace-nowrap truncate self-center">
                          {entry.timestamp ? entry.timestamp.split(' ').pop() : ''}
                        </div>
                        <div className="px-3 py-5 self-center">
                          <LogBadge level={entry.level} />
                        </div>
                        <div className="px-6 py-5 text-sm font-medium text-[var(--color-text)] truncate self-center">
                          {entry.message}
                        </div>
                        <div className="px-6 py-5 text-xs font-semibold text-[var(--color-text-faint)] uppercase tracking-wider text-right self-center group-hover:text-[var(--color-accent)] transition-colors">
                          {entry.module || 'SYSTEM'}
                        </div>
                        <div className="px-4 py-5 text-[var(--color-text-faint)] self-center">
                          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                            className="transition-transform group-hover:translate-x-0.5">
                            <polyline points="9 18 15 12 9 6" />
                          </svg>
                        </div>
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </div>
        </div>

        {/* ── Right: Audit summary panel ── */}
        <div className="w-80 shrink-0 overflow-y-auto bg-[var(--color-panel)] border-l border-[var(--color-border)] p-5 space-y-5">
          <div>
            <h4 className="text-[10px] font-black text-[var(--color-text-dim)] uppercase tracking-[0.2em] mb-3 flex items-center gap-2">
              <Search size={12} /> Module Breakdown
            </h4>
            <div className="space-y-1.5">
              {auditSummary.topModules.length === 0 ? (
                <p className="text-xs text-[var(--color-text-faint)] italic">No data</p>
              ) : (
                auditSummary.topModules.map(([mod, cnt]) => {
                  const pct = auditSummary.total > 0 ? (cnt / auditSummary.total * 100) : 0
                  return (
                    <div key={mod} className="flex items-center gap-2">
                      <span className="text-[10px] font-semibold text-[var(--color-text)] truncate flex-1">{mod}</span>
                      <div className="w-20 h-1.5 rounded-full bg-[var(--color-panel-3)] overflow-hidden">
                        <div className="h-full rounded-full bg-[var(--color-accent)]" style={{ width: `${pct}%` }} />
                      </div>
                      <span className="text-[10px] font-bold text-[var(--color-text-faint)] font-[Geist_Mono] w-10 text-right">{cnt}</span>
                    </div>
                  )
                })
              )}
            </div>
          </div>

          <div className="border-t border-[var(--color-border)] pt-4">
            <h4 className="text-[10px] font-black text-[var(--color-text-dim)] uppercase tracking-[0.2em] mb-3 flex items-center gap-2">
              <AlertOctagon size={12} /> Top Error Messages
            </h4>
            <div className="space-y-2">
              {auditSummary.topErrors.length === 0 ? (
                <p className="text-xs text-[var(--color-text-faint)] italic">No errors in filter</p>
              ) : (
                auditSummary.topErrors.map(([msg, info]) => (
                  <div key={msg} className="bg-[var(--color-panel-2)] border border-[var(--color-border)]/50 rounded-lg p-2.5">
                    <p className="text-[11px] font-medium text-[var(--color-text)] leading-snug line-clamp-2">{msg}</p>
                    <div className="flex items-center gap-3 mt-1.5 text-[10px] text-[var(--color-text-faint)]">
                      <span className="font-bold text-[var(--color-danger)]">×{info.count}</span>
                      <span>Last: {info.lastSeen ? info.lastSeen.split(' ').pop() : '-'}</span>
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>

          <div className="border-t border-[var(--color-border)] pt-4">
            <h4 className="text-[10px] font-black text-[var(--color-text-dim)] uppercase tracking-[0.2em] mb-2">Filter Summary</h4>
            <div className="text-[11px] text-[var(--color-text-dim)] space-y-1">
              <p>Date range: {dateFrom || 'any'} → {dateTo || 'any'}</p>
              <p>Levels: {[...activeLevels].join(', ')}</p>
              <p>Search: {search || '(none)'}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

// ══════════════════════════════════════════════
//  Logs Page (Tab Container)
// ══════════════════════════════════════════════

export function Logs() {
  const [activeTab, setActiveTab] = useState<TabId>('overview')

  return (
    <div className="flex flex-col h-full bg-[var(--color-bg)] animate-in fade-in duration-500">
      {/* ── Header ── */}
      <div className="py-8 border-b border-[var(--color-border)] bg-[var(--color-panel-2)]/50 flex items-center justify-between px-10">
        <div>
          <div className="flex items-center gap-3 mb-2">
            <div className="w-8 h-8 rounded-lg bg-accent/10 flex items-center justify-center text-accent border border-accent/20">
               <ScrollText size={18} />
            </div>
            <h1 className="text-sm font-black text-[var(--color-text)] uppercase tracking-[0.25em] text-wrap-balance">Event Intelligence</h1>
          </div>
          <p className="text-3xl font-bold text-[var(--color-text)] tracking-tight">Logs & Event Stream</p>
          <p className="text-[var(--color-text-dim)] text-xs font-semibold uppercase tracking-widest mt-2 text-wrap-pretty">System log overview and real-time audit trail for all subsystems</p>
        </div>
      </div>

      {/* ── Tab Bar ── */}
      <div className="flex border-b border-[var(--color-border)] bg-[var(--color-panel)] px-6 overflow-hidden">
        {[
          { id: 'overview' as TabId, label: 'Overview', icon: <LayoutList size={18} /> },
          { id: 'live' as TabId, label: 'Live Stream', icon: <Zap size={18} /> },
          { id: 'audit' as TabId, label: 'Audit', icon: <ShieldCheck size={18} /> },
        ].map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            data-automation-id={`logs-tab-${tab.id}`}
            className={cn(
              'flex items-center gap-3 px-10 py-5 text-sm font-bold transition-all border-b-2 border-transparent relative',
              activeTab === tab.id ? 'text-accent' : 'text-text-faint hover:text-text hover:bg-[var(--color-sidebar-hover)]',
            )}
          >
            {tab.icon}
            <span className="uppercase tracking-widest text-[10px] font-black">{tab.label}</span>
            {activeTab === tab.id && (
              <motion.div
                layoutId="logs-tab-indicator"
                className="absolute bottom-0 left-0 right-0 h-0.5 bg-accent"
                transition={{ type: 'spring', stiffness: 500, damping: 30 }}
              />
            )}
          </button>
        ))}
      </div>

      {/* ── Tab Content ── */}
      {activeTab === 'overview' && <OverviewTab />}
      {activeTab === 'live' && <LiveStreamTab />}
      {activeTab === 'audit' && <AuditTab />}
    </div>
  )
}
