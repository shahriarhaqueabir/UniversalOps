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
} from 'lucide-react'

import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { LogEntry } from '@/types'

// ── Constants ──
const ROW_HEIGHT = 76          // px per collapsed row
const EXPANDED_HEIGHT = 380    // px for expanded detail area
const LEVELS = ['INFO', 'WARN', 'ERROR', 'DEBUG'] as const

// ── Helpers ──

const levelStyle: Record<string, { bg: string; text: string; icon: React.ReactNode }> = {
  INFO: { bg: 'bg-blue-500/15', text: 'text-blue-400', icon: <Info size={14} /> },
  WARN: { bg: 'bg-amber-500/15', text: 'text-amber-400', icon: <AlertTriangle size={14} /> },
  ERROR: { bg: 'bg-red-500/15', text: 'text-red-400', icon: <AlertOctagon size={14} /> },
  DEBUG: { bg: 'bg-gray-500/15', text: 'text-gray-400', icon: <Bug size={14} /> },
}

function LogBadge({ level }: { level: string }) {
  const s = levelStyle[level] || levelStyle.DEBUG
  return (
    <span className={cn('inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-md text-[10px] font-bold uppercase tracking-wider border border-current/20', s.bg, s.text)}>
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

// ── Logs Page ──

export function Logs() {
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
  const { data: allLogs = [], isLoading, refetch } = useQuery<LogEntry[]>({
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
  // We reset scroll position when data refreshes if autoScroll is enabled
  const prevLengthRef = useRef(0)
  const prevLength = prevLengthRef.current
  if (autoScroll && allLogs.length !== prevLength && allLogs.length > 0) {
    prevLengthRef.current = allLogs.length
    // schedule for after render
    requestAnimationFrame(() => {
      if (scrollRef.current) {
        scrollRef.current.scrollTop = 0
      }
    })
  }

  // ── Render ──
  return (
    <div className="flex flex-col h-full bg-[var(--color-bg)]">
      {/* ── Header ── */}
      <div className="p-6 border-b border-[var(--color-border)] bg-[var(--color-panel-2)] flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-[var(--color-text)] flex items-center gap-3">
            <LayoutList size={24} className="text-[var(--color-accent)]" />
            Live Event Aggregator
          </h1>
          <p className="text-[var(--color-text-dim)] text-sm mt-1">
            Real-time audit trail of all system, network, and AI operations.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2 bg-[var(--color-panel)] border border-[var(--color-border)] px-3 py-1.5 rounded-lg">
            <span className="w-1.5 h-1.5 rounded-full bg-[var(--color-success)] animate-pulse shadow-[0_0_8px_var(--color-success)]" />
            <span className="text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Streaming</span>
          </div>
          <button onClick={() => refetch()} className="p-2 bg-[var(--color-panel)] border border-[var(--color-border)] rounded-lg hover:bg-[var(--color-panel-3)] text-[var(--color-text-dim)] hover:text-[var(--color-text)] transition-all">
            <RefreshCw size={18} className={cn(isLoading && "animate-spin")} />
          </button>
        </div>
      </div>

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

        <div className="flex items-center gap-1.5">
          {LEVELS.map((level) => {
            const count = allLogs.filter(l => l.level === level).length
            return (
              <button
                key={level}
                onClick={() => toggleLevel(level)}
                className={cn(
                  'flex items-center gap-1.5 px-2.5 py-1.5 rounded-full text-xs font-semibold transition-all border',
                  activeLevels.has(level)
                    ? `${levelStyle[level].bg} ${levelStyle[level].text} border-current/30`
                    : 'bg-[var(--color-panel-2)] text-[var(--color-text-faint)] border-transparent opacity-40'
                )}
              >
                {levelStyle[level].icon}
                {level}
                <span className="text-[10px] opacity-60">{count}</span>
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
    </div>
  )
}
