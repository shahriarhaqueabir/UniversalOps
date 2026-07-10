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
    <span className={cn('inline-flex items-center gap-2 px-3 py-1 rounded-full text-xs font-black uppercase tracking-widest border border-current opacity-80', s.bg, s.text)}>
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
          <span className="text-lg font-black text-text-faint uppercase tracking-[0.2em]">{entry.module || 'System Core'}</span>
        </div>
        <h4 className="text-2xl font-black text-text leading-relaxed mb-6">{entry.message}</h4>
        <div className="p-5 bg-panel-3 border border-border rounded-xl font-[JetBrains_Mono] text-base text-text-dim">
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

  // ── React Query for polling ──
  const { data: allLogs = [], isLoading, refetch } = useQuery<LogEntry[]>({
    queryKey: ['logs'],
    queryFn: async () => {
      const res = await call('Logs.GetLogs', '', '', 200) as LogEntry[]
      return res || []
    },
    refetchInterval: 2000,
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
      <div className="p-8 border-b border-border bg-panel-2 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-text flex items-center gap-4">
            <LayoutList size={32} className="text-accent" />
            Live Event Aggregator
          </h1>
          <p className="text-text-dim text-lg mt-2">
            Real-time audit trail of all system, network, and AI operations.
          </p>
        </div>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2 bg-panel border border-border px-4 py-2 rounded-xl">
            <span className="w-2 h-2 rounded-full bg-success animate-pulse shadow-[0_0_8px_var(--color-success)]" />
            <span className="text-sm font-black text-text-dim uppercase tracking-widest">Streaming</span>
          </div>
          <button onClick={() => refetch()} className="p-3 bg-panel border border-border rounded-xl hover:bg-panel-3 text-text-dim hover:text-text transition-all shadow-md">
            <RefreshCw size={24} className={cn(isLoading && "animate-spin")} />
          </button>
        </div>
      </div>

      {/* ── Toolbar ── */}
      <div className="border-b border-border p-8 bg-panel flex items-center gap-6 flex-wrap">
        <div className="relative group w-96">
          <Search size={24} className="absolute left-4 top-1/2 -translate-y-1/2 text-text-faint group-focus-within:text-accent transition-colors" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search live stream..."
            className="w-full bg-panel border border-border rounded-xl pl-14 pr-4 py-4 text-xl text-text placeholder-text-faint focus:outline-none focus:border-accent shadow-inner"
          />
        </div>

        <div className="flex items-center gap-3">
          {LEVELS.map((level) => (
            <button
              key={level}
              onClick={() => toggleLevel(level)}
              className={cn(
                'transition-all hover:scale-105 active:scale-95',
                activeLevels.has(level) ? 'opacity-100' : 'opacity-20 grayscale'
              )}
            >
              <LogBadge level={level} />
            </button>
          ))}
        </div>
        <div className="flex-1" />

        <button
          onClick={() => setAutoScroll((p) => !p)}
          className={cn(
            'flex items-center gap-3 px-6 py-4 text-lg font-bold rounded-xl border transition-all',
            autoScroll ? 'bg-accent/10 border-accent text-accent shadow-lg' : 'text-text-faint border-border hover:text-text'
          )}
        >
          <ArrowDownToDot size={20} />
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
          className="sticky top-0 z-10 grid grid-cols-[160px_130px_1fr_160px_40px] bg-panel-2 border-b border-border shadow-md"
          style={{ height: ROW_HEIGHT }}
        >
          <div className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest">Timestamp</div>
          <div className="px-4 py-6 text-sm font-black text-text-dim uppercase tracking-widest">Level</div>
          <div className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest">Event Message</div>
          <div className="px-8 py-6 text-sm font-black text-text-dim uppercase tracking-widest text-right">Module</div>
          <div />
        </div>

        {count === 0 ? (
          <div className="flex flex-col items-center justify-center text-text-faint opacity-20 py-40">
            <LayoutList size={120} className="mb-8" />
            <p className="text-4xl font-black uppercase tracking-[0.2em]">Idle Stream</p>
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
                      'grid grid-cols-[160px_130px_1fr_160px_40px] border-b border-border/20 cursor-pointer transition-colors group',
                      'hover:bg-[var(--color-sidebar-hover)]',
                      isExpanded && 'bg-accent/5',
                    )}
                    style={{ height: ROW_HEIGHT }}
                  >
                    <div className="px-8 py-5 text-lg text-text-faint font-bold font-[JetBrains_Mono] whitespace-nowrap truncate self-center">
                      {entry.timestamp ? entry.timestamp.split(' ').pop() : format(new Date(), 'HH:mm:ss')}
                    </div>
                    <div className="px-4 py-5 self-center">
                      <LogBadge level={entry.level} />
                    </div>
                    <div className="px-8 py-5 text-xl font-bold text-text truncate self-center">
                      {entry.message}
                    </div>
                    <div className="px-8 py-5 text-sm text-text-faint font-black uppercase tracking-widest text-right self-center group-hover:text-accent transition-colors">
                      {entry.module || 'SYSTEM'}
                    </div>
                    <div className="px-6 py-5 text-text-faint self-center">
                      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
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
