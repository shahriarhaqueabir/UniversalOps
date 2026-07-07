import { useState, useEffect } from 'react'
import { X, Activity, Trash2, RefreshCw, AlertTriangle } from 'lucide-react'
import { cn } from '@/lib/utils'
import { mockProcesses } from '@/lib/mockData'
import type { ProcessInfo } from '@/types'

interface ProcessManagerProps {
  open: boolean
  onClose: () => void
}

export function ProcessManager({ open, onClose }: ProcessManagerProps) {
  const [processes, setProcesses] = useState<ProcessInfo[]>([])
  const [selectedPid, setSelectedPid] = useState<number | null>(null)
  const [confirmKill, setConfirmKill] = useState(false)
  const [search, setSearch] = useState('')

  useEffect(() => {
    if (open) {
      setProcesses(mockProcesses())
    }
  }, [open])

  const filtered = processes.filter((p) =>
    p.name.toLowerCase().includes(search.toLowerCase()),
  )

  const handleKill = () => {
    if (selectedPid) {
      setProcesses((prev) => prev.filter((p) => p.pid !== selectedPid))
      setSelectedPid(null)
      setConfirmKill(false)
    }
  }

  if (!open) return null

  const selected = processes.find((p) => p.pid === selectedPid)

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <div className="bg-card border border-border rounded-xl shadow-2xl w-[640px] max-h-[80vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-border">
          <div className="flex items-center gap-2">
            <Activity size={18} className="text-primary" />
            <h2 className="text-lg font-semibold text-text">Process Manager</h2>
          </div>
          <button
            onClick={onClose}
            className="p-1 rounded hover:bg-sidebar-hover text-muted hover:text-text transition-colors"
          >
            <X size={18} />
          </button>
        </div>

        {/* Search */}
        <div className="p-4 border-b border-border">
          <input
            type="text"
            placeholder="Search processes..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full bg-background border border-border rounded-lg px-3 py-2 text-sm text-text placeholder-muted focus:outline-none focus:ring-2 focus:ring-primary/50"
          />
        </div>

        {/* List */}
        <div className="flex-1 overflow-y-auto p-4 space-y-1">
          {filtered.length === 0 ? (
            <p className="text-sm text-muted text-center py-8">No processes found</p>
          ) : (
            filtered.map((proc) => (
              <button
                key={proc.pid}
                onClick={() => {
                  setSelectedPid(proc.pid)
                  setConfirmKill(false)
                }}
                className={cn(
                  'w-full flex items-center justify-between px-3 py-2 rounded-lg text-sm transition-colors',
                  selectedPid === proc.pid
                    ? 'bg-primary/10 border border-primary/30'
                    : 'hover:bg-sidebar-hover border border-transparent',
                )}
              >
                <div className="flex items-center gap-3 min-w-0">
                  <span className="font-mono text-xs text-muted w-16 text-right">{proc.pid}</span>
                  <span className="text-text truncate">{proc.name}</span>
                </div>
                <div className="flex items-center gap-4 shrink-0">
                  <span className={cn('text-xs font-mono', proc.cpu > 50 ? 'text-warning' : 'text-muted')}>
                    {proc.cpu.toFixed(1)}%
                  </span>
                  <span className={cn('text-xs font-mono', proc.memory > 300 ? 'text-danger' : 'text-muted')}>
                    {proc.memory.toFixed(0)} MB
                  </span>
                  <span className={cn('text-xs px-1.5 py-0.5 rounded', proc.status === 'running' ? 'bg-success/20 text-success' : 'bg-muted/20 text-muted')}>
                    {proc.status}
                  </span>
                </div>
              </button>
            ))
          )}
        </div>

        {/* Detail / Action footer */}
        {selected && (
          <div className="border-t border-border p-4 flex items-center justify-between">
            <div className="flex items-center gap-4 text-sm">
              <span className="text-muted">
                PID <strong className="text-text">{selected.pid}</strong>
              </span>
              <span className="text-muted">
                CPU <strong className="text-text">{selected.cpu.toFixed(1)}%</strong>
              </span>
              <span className="text-muted">
                Mem <strong className="text-text">{selected.memory.toFixed(0)} MB</strong>
              </span>
              <span className="text-muted">
                FDs <strong className="text-text">{selected.num_fds}</strong>
              </span>
            </div>
            <div className="flex items-center gap-2">
              {confirmKill ? (
                <div className="flex items-center gap-2">
                  <span className="flex items-center gap-1 text-xs text-danger">
                    <AlertTriangle size={12} /> Kill {selected.name}?
                  </span>
                  <button
                    onClick={handleKill}
                    className="px-3 py-1.5 text-xs font-medium bg-danger/20 text-danger border border-danger/30 rounded-lg hover:bg-danger/30 transition-colors"
                  >
                    Confirm
                  </button>
                  <button
                    onClick={() => setConfirmKill(false)}
                    className="px-3 py-1.5 text-xs font-medium text-muted border border-border rounded-lg hover:bg-sidebar-hover transition-colors"
                  >
                    Cancel
                  </button>
                </div>
              ) : (
                <button
                  onClick={() => setConfirmKill(true)}
                  className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-danger border border-danger/30 rounded-lg hover:bg-danger/10 transition-colors"
                >
                  <Trash2 size={14} /> Kill Process
                </button>
              )}
            </div>
          </div>
        )}

        {/* Refresh hint */}
        <div className="flex items-center justify-between px-4 py-2 border-t border-border">
          <span className="text-xs text-muted">{processes.length} processes</span>
          <button
            onClick={() => setProcesses(mockProcesses())}
            className="flex items-center gap-1 text-xs text-muted hover:text-primary transition-colors"
          >
            <RefreshCw size={12} /> Refresh
          </button>
        </div>
      </div>
    </div>
  )
}
