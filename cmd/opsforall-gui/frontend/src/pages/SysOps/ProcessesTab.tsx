import { useState, useRef } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Search, Trash2, TreePine, List } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { ConfirmDialog } from '@/components/dialogs/ConfirmDialog'
import type { ProcessInfo } from '@/types'
import { cn } from '@/lib/utils'
import { useVirtualizer } from '@tanstack/react-virtual'

export function ProcessesTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const queryClient = useQueryClient()
  const [search, setSearch] = useState('')
  const [view, setView] = useState<'list' | 'tree'>('list')
  const [killTarget, setKillTarget] = useState<{ pid: number; name: string } | null>(null)

  const { data: processes = [] } = useQuery<ProcessInfo[]>({
    queryKey: ['sysops-processes'],
    queryFn: async () => { const r = await call('SysOps.ListAllProcesses', 100); return (r as ProcessInfo[]) || [] },
    refetchInterval: refreshInterval,
  })

  const killProcess = async (pid: number) => {
    await call('DevOps.KillProcess', pid)
    queryClient.invalidateQueries({ queryKey: ['sysops-processes'] })
    setKillTarget(null)
  }

  const filtered = processes.filter(p => p.name.toLowerCase().includes(search.toLowerCase()))

  const parentRef = useRef<HTMLDivElement>(null)

  const rowVirtualizer = useVirtualizer({
    count: filtered.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 49,
    overscan: 10,
  })

  return (
    <div className="space-y-6">
      <ConfirmDialog
        open={killTarget !== null}
        title="Kill Process"
        description={`Terminate "${killTarget?.name}" (PID: ${killTarget?.pid})?`}
        type="danger"
        confirmText="Kill"
        onConfirm={() => killProcess(killTarget!.pid)}
        onClose={() => setKillTarget(null)}
      />

      <div className="flex items-center gap-4">
        <div className="relative flex-1">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-text-faint)]" />
          <input
            type="text"
            placeholder="Filter processes..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full bg-[var(--color-panel)] border border-[var(--color-border)] rounded-xl pl-10 pr-4 py-2 text-sm text-[var(--color-text)] placeholder-[var(--color-text-faint)] focus:outline-none focus:border-[var(--color-accent)]"
          />
        </div>
        <div className="flex gap-1 bg-[var(--color-panel)] border border-[var(--color-border)] rounded-lg p-1">
          <button onClick={() => setView('list')} className={cn("px-3 py-1.5 rounded text-xs font-bold", view === 'list' ? "bg-[var(--color-accent)] text-white" : "text-[var(--color-text-faint)]")}>
            <List size={14} />
          </button>
          <button onClick={() => setView('tree')} className={cn("px-3 py-1.5 rounded text-xs font-bold", view === 'tree' ? "bg-[var(--color-accent)] text-white" : "text-[var(--color-text-faint)]")}>
            <TreePine size={14} />
          </button>
        </div>
        <span className="text-sm text-[var(--color-text-faint)]">{filtered.length} active</span>
      </div>

      <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-xl overflow-hidden">
        {/* Sticky header */}
        <table className="w-full text-left">
          <thead className="bg-[var(--color-panel-2)] border-b border-[var(--color-border)]">
            <tr>
              <th className="px-4 py-3 text-xs font-bold text-[var(--color-text-faint)] uppercase">Process</th>
              <th className="px-4 py-3 text-xs font-bold text-[var(--color-text-faint)] uppercase text-right">CPU %</th>
              <th className="px-4 py-3 text-xs font-bold text-[var(--color-text-faint)] uppercase text-right">RAM (MB)</th>
              <th className="px-4 py-3 text-xs font-bold text-[var(--color-text-faint)] uppercase text-right">FDs</th>
              <th className="px-4 py-3 w-12" />
            </tr>
          </thead>
        </table>
        {/* Virtualised body */}
        <div ref={parentRef} className="max-h-[600px] overflow-y-auto">
          <div style={{ height: `${rowVirtualizer.getTotalSize()}px`, position: 'relative' }}>
            <table className="w-full text-left" style={{ tableLayout: 'fixed' }}>
              <colgroup>
                <col />
                <col style={{ width: '80px' }} />
                <col style={{ width: '90px' }} />
                <col style={{ width: '70px' }} />
                <col style={{ width: '48px' }} />
              </colgroup>
              <tbody>
                {rowVirtualizer.getVirtualItems().map(virtualRow => {
                  const p = filtered[virtualRow.index]
                  return (
                    <tr
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
                      className="border-b border-[var(--color-border)]/20 hover:bg-[var(--color-sidebar-hover)] group"
                    >
                      <td className="px-4 py-3">
                        <span className="text-sm font-medium text-[var(--color-text)]">{p.name}</span>
                        <span className="text-xs text-[var(--color-text-faint)] ml-2">PID {p.pid}</span>
                      </td>
                      <td className="px-4 py-3 text-right text-sm font-bold text-[var(--color-accent)] tabular-nums">{p.cpu.toFixed(1)}%</td>
                      <td className="px-4 py-3 text-right text-sm text-[var(--color-text-dim)] tabular-nums">{p.memory.toFixed(0)}</td>
                      <td className="px-4 py-3 text-right text-sm text-[var(--color-text-faint)] tabular-nums">{p.num_fds}</td>
                      <td className="px-4 py-3">
                        <button onClick={() => setKillTarget({ pid: p.pid, name: p.name })} className="opacity-0 group-hover:opacity-100 p-1.5 text-[var(--color-text-faint)] hover:text-[var(--color-danger)] hover:bg-[var(--color-danger)]/10 rounded-lg transition-all">
                          <Trash2 size={14} />
                        </button>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  )
}
