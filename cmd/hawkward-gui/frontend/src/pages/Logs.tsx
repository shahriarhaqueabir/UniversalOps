import { useState, useEffect, useRef, useCallback, useMemo, Fragment } from 'react'
import {
  Search,
  Trash2,
  ChevronDown,
  ChevronRight,
  ArrowDownToDot,
  RefreshCw,
  X,
  Info,
  AlertTriangle,
  AlertOctagon,
  Bug,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { mockLogEntries } from '@/lib/mockData'
import type { LogEntry } from '@/lib/mockData'

// ── LogBadge ──

function LogBadge({ level }: { level: string }) {
  const styles: Record<string, { bg: string; text: string; icon: React.ReactNode }> = {
    INFO: { bg: 'bg-blue-500/15', text: 'text-blue-400', icon: <Info size={10} /> },
    WARN: { bg: 'bg-amber-500/15', text: 'text-amber-400', icon: <AlertTriangle size={10} /> },
    ERROR: { bg: 'bg-red-500/15', text: 'text-red-400', icon: <AlertOctagon size={10} /> },
    DEBUG: { bg: 'bg-gray-500/15', text: 'text-gray-400', icon: <Bug size={10} /> },
  }
  const s = styles[level] || styles.DEBUG
  return (
    <span className={cn('inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-semibold', s.bg, s.text)}>
      {s.icon}
      {level}
    </span>
  )
}

// ── LogRow ──

function LogRow({ entry, expanded, onClick }: { entry: LogEntry; expanded: boolean; onClick: () => void }) {
  return (
    <tr
      onClick={onClick}
      className={cn(
        'border-b border-border/20 cursor-pointer transition-colors',
        expanded ? 'bg-primary/5' : 'hover:bg-white/[0.03]',
      )}
    >
      <td className="px-3 py-2 text-xs text-muted font-mono whitespace-nowrap w-24">
        {new Date(entry.timestamp).toLocaleTimeString()}
      </td>
      <td className="px-2 py-2 w-20">
        <LogBadge level={entry.level} />
      </td>
      <td className="px-3 py-2 text-xs text-text truncate max-w-0">
        {entry.message}
      </td>
      <td className="px-3 py-2 text-[11px] text-muted font-mono w-24 text-right">
        {entry.module}
      </td>
      <td className="px-2 py-2 w-6 text-muted">
        {expanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
      </td>
    </tr>
  )
}

// ── LogDetail ──

function LogDetail({ entry }: { entry: LogEntry }) {
  return (
    <div className="bg-card border border-border rounded-lg p-4 space-y-3">
      <div className="grid grid-cols-2 gap-3">
        <div>
          <p className="text-[10px] text-muted uppercase tracking-wider mb-0.5">Timestamp</p>
          <p className="text-xs text-text">{new Date(entry.timestamp).toLocaleString()}</p>
        </div>
        <div>
          <p className="text-[10px] text-muted uppercase tracking-wider mb-0.5">Level</p>
          <LogBadge level={entry.level} />
        </div>
        <div>
          <p className="text-[10px] text-muted uppercase tracking-wider mb-0.5">Source</p>
          <p className="text-xs text-text capitalize">{entry.source}</p>
        </div>
        <div>
          <p className="text-[10px] text-muted uppercase tracking-wider mb-0.5">Module</p>
          <p className="text-xs text-text font-mono">{entry.module}</p>
        </div>
      </div>

      <div>
        <p className="text-[10px] text-muted uppercase tracking-wider mb-0.5">Message</p>
        <div className="bg-background border border-border rounded px-3 py-2">
          <p className="text-xs text-text">{entry.message}</p>
        </div>
      </div>

      {entry.details && (
        <div>
          <p className="text-[10px] text-muted uppercase tracking-wider mb-0.5">Details</p>
          <div className="bg-background border border-border rounded px-3 py-2">
            <p className="text-xs text-text">{entry.details}</p>
          </div>
        </div>
      )}

      {entry.stackTrace && (
        <div>
          <p className="text-[10px] text-muted uppercase tracking-wider mb-0.5">Stack Trace</p>
          <div className="bg-red-950/20 border border-red-500/20 rounded px-3 py-2">
            <pre className="text-xs text-red-400 font-mono whitespace-pre-wrap">{entry.stackTrace}</pre>
          </div>
        </div>
      )}
    </div>
  )
}

// ── Logs Page ──

const LEVELS = ['INFO', 'WARN', 'ERROR', 'DEBUG'] as const
const SOURCES = ['system', 'network', 'security', 'devops', 'ai'] as const

export function Logs() {
  const [allLogs, setAllLogs] = useState<LogEntry[]>([])
  const [search, setSearch] = useState('')
  const [activeLevels, setActiveLevels] = useState<Set<string>>(new Set(LEVELS))
  const [activeSource, setActiveSource] = useState<string>('all')
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const [autoScroll, setAutoScroll] = useState(true)
  const [tailMode, setTailMode] = useState(false)

  const listRef = useRef<HTMLDivElement>(null)

  // Load initial data
  useEffect(() => {
    setAllLogs(mockLogEntries(200))
  }, [])

  // Tail mode: append new entries periodically
  useEffect(() => {
    if (!tailMode) return
    const interval = setInterval(() => {
      setAllLogs((prev) => {
        const fresh = mockLogEntries(1 + Math.floor(Math.random() * 2))
        return [...fresh, ...prev].slice(0, 500)
      })
    }, 3000)
    return () => clearInterval(interval)
  }, [tailMode])

  // Auto-scroll to top when new entries arrive in tail mode
  useEffect(() => {
    if (tailMode && autoScroll && listRef.current) {
      listRef.current.scrollTop = 0
    }
  }, [allLogs, tailMode, autoScroll])

  // Filter
  const filteredLogs = useMemo(() => {
    let result = allLogs

    // Level filter
    if (activeLevels.size < LEVELS.length) {
      result = result.filter((l) => activeLevels.has(l.level))
    }

    // Source filter
    if (activeSource !== 'all') {
      result = result.filter((l) => l.source === activeSource)
    }

    // Text search
    if (search.trim()) {
      const q = search.toLowerCase()
      result = result.filter(
        (l) =>
          l.message.toLowerCase().includes(q) ||
          l.module.toLowerCase().includes(q) ||
          l.source.toLowerCase().includes(q) ||
          l.level.toLowerCase().includes(q),
      )
    }

    return result
  }, [allLogs, activeLevels, activeSource, search])

  // Handlers
  const toggleLevel = useCallback((level: string) => {
    setActiveLevels((prev) => {
      const next = new Set(prev)
      if (next.has(level)) next.delete(level)
      else next.add(level)
      return next
    })
  }, [])

  const clearLogs = useCallback(() => {
    setAllLogs([])
    setExpandedId(null)
  }, [])

  const refreshLogs = useCallback(() => {
    setAllLogs(mockLogEntries(200))
  }, [])

  const handleRowClick = useCallback(
    (id: string) => {
      setExpandedId((prev) => (prev === id ? null : id))
    },
    [],
  )

  // Derived stats
  const errorCount = useMemo(() => allLogs.filter((l) => l.level === 'ERROR').length, [allLogs])
  const warnCount = useMemo(() => allLogs.filter((l) => l.level === 'WARN').length, [allLogs])

  return (
    <div className="flex flex-col h-full">
      {/* ── Filter Bar ── */}
      <div className="border-b border-border p-3 space-y-2">
        <div className="flex items-center gap-3 flex-wrap">
          {/* Search */}
          <div className="relative flex-1 min-w-[180px]">
            <Search size={13} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search logs..."
              className="w-full bg-background border border-border rounded-lg pl-8 pr-3 py-1.5 text-xs text-text placeholder-muted focus:outline-none focus:ring-1 focus:ring-primary"
            />
            {search && (
              <button
                onClick={() => setSearch('')}
                className="absolute right-2 top-1/2 -translate-y-1/2 text-muted hover:text-text"
              >
                <X size={13} />
              </button>
            )}
          </div>

          {/* Level pills */}
          <div className="flex items-center gap-1.5">
            {LEVELS.map((level) => {
              const active = activeLevels.has(level)
              return (
                <button
                  key={level}
                  onClick={() => toggleLevel(level)}
                  className={cn(
                    'transition-opacity',
                    active ? 'opacity-100' : 'opacity-30 hover:opacity-60',
                  )}
                >
                  <LogBadge level={level} />
                </button>
              )
            })}
          </div>

          {/* Source filter */}
          <select
            value={activeSource}
            onChange={(e) => setActiveSource(e.target.value)}
            className="bg-background border border-border rounded-lg px-2.5 py-1.5 text-xs text-text focus:outline-none focus:ring-1 focus:ring-primary"
          >
            <option value="all">All Sources</option>
            {SOURCES.map((s) => (
              <option key={s} value={s}>
                {s.charAt(0).toUpperCase() + s.slice(1)}
              </option>
            ))}
          </select>

          <span className="text-xs text-muted">|</span>

          {/* Actions */}
          <button
            onClick={clearLogs}
            className="flex items-center gap-1 px-2 py-1 text-xs text-muted hover:text-danger transition-colors"
            title="Clear logs"
          >
            <Trash2 size={12} /> Clear
          </button>
          <button
            onClick={refreshLogs}
            className="flex items-center gap-1 px-2 py-1 text-xs text-muted hover:text-text transition-colors"
            title="Refresh"
          >
            <RefreshCw size={12} /> Refresh
          </button>
          <button
            onClick={() => setAutoScroll((p) => !p)}
            className={cn(
              'flex items-center gap-1 px-2 py-1 text-xs rounded transition-colors',
              autoScroll
                ? 'bg-primary/10 text-primary'
                : 'text-muted hover:text-text',
            )}
            title="Auto-scroll to top on new entries"
          >
            <ArrowDownToDot size={12} /> Auto-scroll
          </button>
          <button
            onClick={() => setTailMode((p) => !p)}
            className={cn(
              'flex items-center gap-1 px-2 py-1 text-xs rounded transition-colors',
              tailMode
                ? 'bg-success/10 text-success'
                : 'text-muted hover:text-text',
            )}
            title="Tail mode - auto-follow new entries"
          >
            {tailMode ? '● Tail On' : '○ Tail Off'}
          </button>
        </div>

        {/* Stats */}
        <div className="flex items-center gap-3 text-[11px] text-muted">
          <span>{filteredLogs.length} of {allLogs.length} entries</span>
          {warnCount > 0 && <span className="text-amber-400">{warnCount} warnings</span>}
          {errorCount > 0 && <span className="text-red-400">{errorCount} errors</span>}
        </div>
      </div>

      {/* ── Log List ── */}
      <div ref={listRef} className="flex-1 overflow-y-auto">
        {filteredLogs.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-muted">
            <Search size={32} className="mb-2 opacity-30" />
            <p className="text-sm">No logs match the current filters</p>
            <p className="text-xs text-muted/60 mt-1">Try adjusting the filters above</p>
          </div>
        ) : (
          <table className="w-full text-sm">
            <thead className="sticky top-0 z-10">
              <tr className="text-[11px] text-muted border-b border-border bg-background/95 backdrop-blur">
                <th className="text-left px-3 py-2 font-medium w-24">Time</th>
                <th className="text-left px-2 py-2 font-medium w-20">Level</th>
                <th className="text-left px-3 py-2 font-medium">Message</th>
                <th className="text-right px-3 py-2 font-medium w-24">Module</th>
                <th className="w-6 px-2 py-2" />
              </tr>
            </thead>
            <tbody>
              {filteredLogs.map((entry) => (
                <Fragment key={entry.id}>
                  <LogRow
                    entry={entry}
                    expanded={expandedId === entry.id}
                    onClick={() => handleRowClick(entry.id)}
                  />
                  {expandedId === entry.id && (
                    <tr>
                      <td colSpan={5} className="px-4 py-3 bg-background/50">
                        <LogDetail entry={entry} />
                      </td>
                    </tr>
                  )}
                </Fragment>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}
