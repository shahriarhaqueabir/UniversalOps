import { useState, useRef } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { SystemLogsResult } from '@/types'
import { cn } from '@/lib/utils'
import { useVirtualizer } from '@tanstack/react-virtual'

export function LogsTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const [source, setSource] = useState('system')
  const [count, setCount] = useState(50)

  const { data: logs } = useQuery<SystemLogsResult>({
    queryKey: ['sysops-logs', source, count],
    queryFn: async () => { const r = await call('SysOps.GetSystemLogs', count, source); return r as SystemLogsResult },
    refetchInterval: refreshInterval,
  })

  const entries = logs?.entries ?? []

  const parentRef = useRef<HTMLDivElement>(null)

  const rowVirtualizer = useVirtualizer({
    count: entries.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 40,
    overscan: 10,
  })

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <div className="flex gap-1 bg-[var(--color-panel)] border border-[var(--color-border)] rounded-lg p-1">
          {['system', 'application', 'security'].map(s => (
            <button key={s} onClick={() => setSource(s)} className={cn("px-3 py-1.5 rounded text-xs font-bold capitalize", source === s ? "bg-[var(--color-accent)] text-white" : "text-[var(--color-text-faint)]")}>
              {s}
            </button>
          ))}
        </div>
        <select value={count} onChange={(e) => setCount(Number(e.target.value))} className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-lg px-3 py-1.5 text-sm text-[var(--color-text)]">
          <option value={25}>25 lines</option>
          <option value={50}>50 lines</option>
          <option value={100}>100 lines</option>
          <option value={200}>200 lines</option>
        </select>
        <span className="text-sm text-[var(--color-text-faint)]">{logs?.total || 0} entries</span>
      </div>

      <div className="bg-[var(--color-panel-3)] border border-[var(--color-border)] rounded-xl overflow-hidden">
        <div ref={parentRef} className="max-h-[600px] overflow-y-auto p-4">
          {entries.length === 0 ? (
            <p className="text-[var(--color-text-faint)] text-sm text-center py-8">No log entries</p>
          ) : (
            <div style={{ height: `${rowVirtualizer.getTotalSize()}px`, position: 'relative' }}>
              {rowVirtualizer.getVirtualItems().map(virtualRow => {
                const entry = entries[virtualRow.index]
                return (
                  <div
                    key={virtualRow.key}
                    data-index={virtualRow.index}
                    ref={rowVirtualizer.measureElement}
                    style={{
                      position: 'absolute',
                      top: 0,
                      left: 0,
                      width: '100%',
                      transform: `translateY(${virtualRow.start}px)`,
                    }}
                    className="flex items-start gap-3 py-1.5 border-b border-[var(--color-border)]/20 last:border-0"
                  >
                    <span className={cn('text-xs font-bold px-1.5 py-0.5 rounded min-w-[60px] text-center', entry.level === 'error' ? 'bg-[var(--color-danger)]/20 text-[var(--color-danger)]' : entry.level === 'warning' ? 'bg-[var(--color-warning)]/20 text-[var(--color-warning)]' : 'bg-[var(--color-panel-2)] text-[var(--color-text-faint)]')}>
                      {entry.level.toUpperCase()}
                    </span>
                    <span className="text-xs text-[var(--color-text-faint)] min-w-[140px]">{entry.timestamp}</span>
                    <span className="text-sm font-mono text-[var(--color-text-dim)] flex-1 break-all">{entry.message}</span>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
