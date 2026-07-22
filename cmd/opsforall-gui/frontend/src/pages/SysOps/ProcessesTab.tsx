import { useState, useRef, useMemo } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Trash2, TreePine, List, ChevronRight, ChevronDown } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { ConfirmationModal } from '@/components/dialogs/ConfirmationModal'
import { SearchInput } from '@/components/ui/SearchInput'
import { Panel } from '@/components/ui/Panel'
import type { ProcessInfo, ActionPreview } from '@/types'
import { cn } from '@/lib/utils'
import { useVirtualizer } from '@tanstack/react-virtual'
import { toast } from 'sonner'

interface ProcessNode extends ProcessInfo {
  children: ProcessNode[]
}

function ProcessTreeItem({ node, level, requestKill }: { node: ProcessNode, level: number, requestKill: (pid: number) => void }) {
  const [expanded, setExpanded] = useState(level < 1)
  const hasChildren = node.children.length > 0

  return (
    <div className="flex flex-col">
      <div
        className={cn(
          "flex items-center px-4 py-2 border-b border-[var(--color-border)]/10 hover:bg-[var(--color-sidebar-hover)] group transition-colors",
          level > 0 && "ml-4 border-l border-[var(--color-border)]/20"
        )}
      >
        <div className="flex-[3] min-w-0 flex items-center gap-2">
          {hasChildren ? (
            <button onClick={() => setExpanded(!expanded)} className="p-1 rounded hover:bg-[var(--color-panel-3)]">
              {expanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
            </button>
          ) : <div className="w-6" />}
          <div className="flex flex-col min-w-0">
            <span className="text-sm font-bold text-[var(--color-text)] truncate">{node.name}</span>
            <span className="text-[10px] font-black text-[var(--color-text-faint)] uppercase tracking-tighter">PID {node.pid}</span>
          </div>
        </div>
        <div className="flex-1 text-right text-xs font-bold text-[var(--color-accent)] tabular-nums">{node.cpu.toFixed(1)}%</div>
        <div className="flex-1 text-right text-xs font-bold text-[var(--color-text-dim)] tabular-nums">{node.memory.toFixed(0)}</div>
        <div className="w-12 text-right">
          <button onClick={() => requestKill(node.pid)} className="opacity-0 group-hover:opacity-100 p-1 text-[var(--color-text-faint)] hover:text-[var(--color-danger)] transition-all">
            <Trash2 size={12} />
          </button>
        </div>
      </div>
      {expanded && hasChildren && (
        <div className="flex flex-col">
          {node.children.map(child => (
            <ProcessTreeItem key={child.pid} node={child} level={level + 1} requestKill={requestKill} />
          ))}
        </div>
      )}
    </div>
  )
}

export function ProcessesTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const queryClient = useQueryClient()
  const [search, setSearch] = useState('')
  const [view, setView] = useState<'list' | 'tree'>('list')
  const [preview, setPreview] = useState<ActionPreview | null>(null)

  const { data: processes = [] } = useQuery<ProcessInfo[]>({
    queryKey: ['sysops-processes'],
    queryFn: async () => { const r = await call('SysOps.ListAllProcesses', 250); return (r as ProcessInfo[]) || [] },
    refetchInterval: refreshInterval,
  })

  // Build tree from flat list
  const processTree = useMemo(() => {
    if (view !== 'tree') return []
    const nodes = new Map<number, ProcessNode>()
    processes.forEach(p => nodes.set(p.pid, { ...p, children: [] }))

    const rootNodes: ProcessNode[] = []
    nodes.forEach(node => {
      const parent = nodes.get(node.ppid)
      if (parent && node.pid !== node.ppid) {
        parent.children.push(node)
      } else {
        rootNodes.push(node)
      }
    })
    return rootNodes.sort((a, b) => b.cpu - a.cpu)
  }, [processes, view])

  const requestKill = async (pid: number) => {
    const p = await call('SecOps.KillProcess', pid) as ActionPreview
    setPreview(p)
  }

  const handleConfirm = async () => {
    if (!preview) return
    const handshakeID = preview.handshake_id
    setPreview(null)
    const res = await call('App.ConfirmAction', handshakeID) as { success: boolean; error?: string }
    if (res.success) {
      toast.success('Process terminated')
    } else {
      toast.error(res.error || 'Failed to kill process')
    }
    queryClient.invalidateQueries({ queryKey: ['sysops-processes'] })
  }

  const filtered = processes.filter(p => p.name.toLowerCase().includes(search.toLowerCase()))

  const parentRef = useRef<HTMLDivElement>(null)

  const rowVirtualizer = useVirtualizer({
    count: filtered.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 60,
    overscan: 10,
  })

  return (
    <div className="space-y-6">
      <ConfirmationModal
        preview={preview}
        onConfirm={handleConfirm}
        onCancel={() => setPreview(null)}
      />

      <div className="flex items-center gap-4">
        <SearchInput size="md" value={search} onChange={(e) => setSearch(e.target.value)} placeholder="Filter processes..." className="flex-1" />
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

      <Panel variant="default" padding="none" category="system" className="flex flex-col overflow-hidden border border-[var(--color-border)]">
        {/* Header */}
        <div className="flex items-center bg-[var(--color-panel-2)] border-b border-[var(--color-border)] px-4 py-3 sticky top-0 z-20">
          <div className="flex-[3] text-[10px] font-black text-[var(--color-text-faint)] uppercase tracking-widest">Process</div>
          <div className="flex-1 text-[10px] font-black text-[var(--color-text-faint)] uppercase tracking-widest text-right">CPU %</div>
          <div className="flex-1 text-[10px] font-black text-[var(--color-text-faint)] uppercase tracking-widest text-right">RAM (MB)</div>
          <div className="flex-1 text-[10px] font-black text-[var(--color-text-faint)] uppercase tracking-widest text-right hidden sm:block">Status</div>
          <div className="w-12" />
        </div>

        {view === 'list' ? (
          /* Virtualised List */
          <div ref={parentRef} className="max-h-[650px] overflow-y-auto">
            <div style={{ height: `${rowVirtualizer.getTotalSize()}px`, position: 'relative', width: '100%' }}>
              {rowVirtualizer.getVirtualItems().map(virtualRow => {
                const p = filtered[virtualRow.index]
                if (!p) return null
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
                    className="flex items-center px-4 py-3 border-b border-[var(--color-border)]/10 hover:bg-[var(--color-sidebar-hover)] group transition-colors"
                  >
                    <div className="flex-[3] min-w-0 flex items-center gap-3">
                      <div className="flex flex-col min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="text-sm font-bold text-[var(--color-text)] truncate">{p.name}</span>
                          {p.is_signed && <span className="w-2 h-2 rounded-full bg-blue-500 shadow-[0_0_8px_rgba(59,130,246,0.5)]" title={`Signed by ${p.publisher}`} />}
                        </div>
                        <span className="text-[10px] font-black text-[var(--color-text-faint)] uppercase tracking-tighter">PID {p.pid}</span>
                      </div>
                    </div>
                    <div className="flex-1 text-right text-sm font-black text-[var(--color-accent)] tabular-nums">{p.cpu.toFixed(1)}%</div>
                    <div className="flex-1 text-right text-sm font-bold text-[var(--color-text-dim)] tabular-nums">{p.memory.toFixed(0)}</div>
                    <div className="flex-1 text-right hidden sm:block">
                      <span className={cn(
                        "text-[9px] font-black uppercase px-2 py-0.5 rounded border tracking-tighter",
                        p.status === 'running' ? "bg-green-500/10 text-green-500 border-green-500/20" : "bg-warning/10 text-warning border-warning/20"
                      )}>
                        {p.status}
                      </span>
                    </div>
                    <div className="w-12 text-right">
                      <button onClick={() => requestKill(p.pid)} className="opacity-0 group-hover:opacity-100 p-1.5 text-[var(--color-text-faint)] hover:text-[var(--color-danger)] hover:bg-[var(--color-danger)]/10 rounded-lg transition-all active:scale-[0.95]">
                        <Trash2 size={14} />
                      </button>
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        ) : (
          /* Tree View */
          <div className="max-h-[650px] overflow-y-auto p-2">
            {processTree.map(node => (
              <ProcessTreeItem key={node.pid} node={node} level={0} requestKill={requestKill} />
            ))}
          </div>
        )}
      </Panel>
    </div>
  )
}
