import { useState, useRef, useCallback, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useVirtualizer } from '@tanstack/react-virtual'
import { format } from 'date-fns'
import {
  Search,
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

// ── Constants ──
const ROW_HEIGHT = 76          // px per collapsed row
const EXPANDED_HEIGHT = 380    // px for expanded detail area
const LEVELS = ['INFO', 'WARN', 'ERROR', 'DEBUG'] as const

type TabId = 'overview' | 'live'

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

// ── Detail Panel (memoized, renders when expanded) ──

function DetailPanel({ entry, idx }: { entry: LogEntry; idx: number }) {
  return (
    <div className="px-10 py-6 bg-panel-3 border-b border-border/50">
      <div className="bg-panel border border-border rounded-[24px] p-8 shadow-2xl relative overflow-hidden">
        <div className="absolute top-0 right-0 w-40 h-40 bg-accent/5 rounded-bl-full pointer-events-none" />
        <div className="flex items-center gap-6 mb-6">
          <LogBadge level={entry.level} />
          <span className="text-sm font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">{entry.module || 'System Core'}</span>
        </div>
        <h4 className="text-lg font-bold text-[var(--color-text)] leading-relaxed mb-6">{entry.message}</h4>
        <div className="p-5 bg-panel-3 border border-border rounded-xl font-[Geist_Mono] text-base text-text-dim">
          {entry.line || 'No additional context available.'}
        </div>
        <div className="mt-6 flex items-center gap-6 text-sm font-bold text-text-faint">
          <div className="flex items-center gap-2"><Clock size={16} /> {entry.timestamp}</div>
          <div className="w-1 h-1 rounded-full bg-border" />
          <div className="flex items-center gap-2"><Zap size={16} /> Trace: {idx.toString(16).toUpperCase().padStart(4, '0')}</div>
        </div>
      </div>
    </div>
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
    refetchInterval: 120000,
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
    <div className="space-y-6 overflow-y-auto flex-1">
      {/* ── Freshness Indicator ── */}
      <DataFreshnessIndicator lastUpdated={statsUpdatedAt ? new Date(statsUpdatedAt) : null} />

      {/* ── Log Volume Cards ── */}
      <div className="grid grid-cols-3 gap-4">
        {[
          { label: 'Today', value: stats?.totalToday ?? 0, icon: <LayoutList size={20} />, color: 'var(--color-accent)' },
          { label: 'This Hour', value: stats?.totalThisHour ?? 0, icon: <Clock size={20} />, color: 'var(--color-success)' },
          { label: 'Last Minute', value: stats?.totalLastMin ?? 0, icon: <Zap size={20} />, color: 'var(--color-warning)' },
        ].map((card) => (
          <div
            key={card.label}
            className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-xl p-5 flex items-center gap-4 hover:border-[var(--color-accent)]/30 transition-colors"
          >
            <div className="w-12 h-12 rounded-lg flex items-center justify-center" style={{ backgroundColor: `color-mix(in srgb, ${card.color} 15%, transparent)`, color: card.color }}>
              {card.icon}
            </div>
            <div>
              <p className="text-2xl font-bold text-[var(--color-text)]">{card.value.toLocaleString()}</p>
              <p className="text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">{card.label}</p>
            </div>
          </div>
        ))}
      </div>

      {/* ── Timeline Chart ── */}
      <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-xl p-6">
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
      </div>

      {/* ── AI Summary ── */}
      <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-xl p-6">
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
      </div>

      {/* ── Error Breakdown ── */}
      <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-xl p-6">
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
      </div>

      <div className="grid grid-cols-2 gap-4">
        {/* ── Top Sources ── */}
        <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-xl p-6">
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
        </div>

        {/* ── Trending Errors ── */}
        <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-xl p-6">
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
        </div>
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
  const [expandedIdx, setExpandedIdx] = useState<number | null>(null)
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
    estimateSize: useCallback((i: number) => {
      return expandedIdx === i ? ROW_HEIGHT + EXPANDED_HEIGHT : ROW_HEIGHT
    }, [expandedIdx]),
    overscan: 20,
  })

  const handleRowClick = (idx: number) => {
    setExpandedIdx(expandedIdx === idx ? null : idx)
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
      {/* ── Toolbar ── */}
      <div className="border-b border-[var(--color-border)] px-6 py-3 bg-[var(--color-panel)] flex items-center gap-4 flex-wrap">
        <div className="relative group flex-1 max-w-md">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-text-faint)] group-focus-within:text-[var(--color-accent)] transition-colors" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search live stream..."
            className="w-full bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-lg pl-10 pr-3 py-2 text-sm text-[var(--color-text)] placeholder-[var(--color-text-faint)] focus:outline-none focus:border-[var(--color-accent)]"
          />
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
              const isExpanded = expandedIdx === virtualItem.index

              return (
                <div
                  key={virtualItem.key}
                  className="absolute left-0 right-0"
                  style={{
                    top: 0,
                    transform: `translateY(${virtualItem.start}px)`,
                    height: isExpanded ? ROW_HEIGHT + EXPANDED_HEIGHT : ROW_HEIGHT,
                  }}
                >
                  {/* Main row */}
                  <div
                    onClick={() => handleRowClick(virtualItem.index)}
                    className={cn(
                      'grid grid-cols-[140px_100px_1fr_140px_36px] border-b border-[var(--color-border)]/20 cursor-pointer transition-colors group',
                      'hover:bg-[var(--color-sidebar-hover)]',
                      isExpanded && 'bg-[var(--color-accent)]/5',
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
                        className={cn('transition-transform', isExpanded && 'rotate-180')}>
                        <polyline points="6 9 12 15 18 9" />
                      </svg>
                    </div>
                  </div>

                  {/* Expanded detail */}
                  {isExpanded && (
                    <div style={{ height: EXPANDED_HEIGHT }}>
                      <DetailPanel entry={entry} idx={virtualItem.index} />
                    </div>
                  )}
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
//  Logs Page (Tab Container)
// ══════════════════════════════════════════════

export function Logs() {
  const [activeTab, setActiveTab] = useState<TabId>('overview')

  return (
    <div className="flex flex-col h-full bg-[var(--color-bg)]">
      {/* ── Header ── */}
      <div className="py-4 border-b border-[var(--color-border)] bg-[var(--color-panel-2)] flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-[var(--color-text)] flex items-center gap-3">
            <LayoutList size={24} className="text-[var(--color-accent)]" />
            Logs &amp; Event Stream
          </h1>
          <p className="text-[var(--color-text-dim)] text-sm mt-1">
            System log overview and real-time audit trail.
          </p>
        </div>
      </div>

      {/* ── Tab Bar ── */}
      <div className="flex border-b border-[var(--color-border)] bg-[var(--color-panel)] px-4">
        {[
          { id: 'overview' as TabId, label: 'Overview', icon: <LayoutList size={18} /> },
          { id: 'live' as TabId, label: 'Live Stream', icon: <Zap size={18} /> },
        ].map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            data-automation-id={`logs-tab-${tab.id}`}
            className={cn(
              'flex items-center gap-2 px-4 py-3 text-sm font-semibold border-b-2 transition-all',
              activeTab === tab.id
                ? 'border-[var(--color-accent)] text-[var(--color-accent)]'
                : 'border-transparent text-[var(--color-text-faint)] hover:text-[var(--color-text-dim)]'
            )}
          >
            {tab.icon}
            {tab.label}
          </button>
        ))}
      </div>

      {/* ── Tab Content ── */}
      {activeTab === 'overview' && <OverviewTab />}
      {activeTab === 'live' && <LiveStreamTab />}
    </div>
  )
}
