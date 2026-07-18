import { useQuery } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { Database, Trash2, Box, CheckCircle2, XCircle, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { useState, useEffect } from 'react'
import { cn } from '@/lib/utils'
import type { OllamaStatus, OllamaProgress } from '@/types'

/**
 * ModelManager — Integrated interface for Ollama model lifecycle.
 * Provides "ls", "pull", and "rm" capabilities directly in the Control Plane.
 */
export function ModelManager() {
  const { call } = useBackend()
  const [pullName, setPullName] = useState('')
  const [isPulling, setIsPulling] = useState(false)
  const [progress, setProgress] = useState<OllamaProgress | null>(null)

  const { data: status, isLoading, refetch } = useQuery<OllamaStatus>({
    queryKey: ['ollama-status'],
    queryFn: async () => {
      const res = await call('AIOps.GetOllamaStatus')
      return res as OllamaStatus
    },
    refetchInterval: 10000,
  })

  // Listen for pull progress from backend
  useEffect(() => {
    const runtime = (window as any).runtime
    if (runtime) {
      const handler = (p: any) => setProgress(p as OllamaProgress)
      runtime.EventsOn('ollama:progress', handler)
      return () => runtime.EventsOff('ollama:progress')
    }
  }, [])

  const handlePull = async () => {
    if (!pullName) return
    setIsPulling(true)
    const id = toast.loading(`Pulling ${pullName}...`)
    try {
      await call('AIOps.PullModel', pullName)
      toast.success(`Model ${pullName} pulled successfully`, { id })
      setPullName('')
      refetch()
    } catch (err) {
      toast.error(`Pull failed: ${err}`, { id })
    } finally {
      setIsPulling(false)
      setProgress(null)
    }
  }

  const handleDelete = async (name: string) => {
    if (!confirm(`Permanently delete model ${name}?`)) return
    const id = toast.loading(`Deleting ${name}...`)
    try {
      await call('AIOps.DeleteModel', name)
      toast.success(`Model ${name} removed`, { id })
      refetch()
    } catch (err) {
      toast.error(`Delete failed: ${err}`, { id })
    }
  }

  if (!status?.available && !isLoading) {
    return (
      <div className="p-6 rounded-2xl bg-danger/5 border border-danger/20 flex items-center gap-4">
        <XCircle className="text-danger" size={24} />
        <div>
          <p className="text-sm font-bold text-danger uppercase tracking-tight">Ollama Not Detected</p>
          <p className="text-xs text-[var(--color-text-dim)] mt-1">
            Ensure Ollama is running and accessible at http://localhost:11434.
            Checked PATH and common installation directories.
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Search/Pull Header */}
      <div className="flex gap-2">
        <div className="relative flex-1">
          <input
            type="text"
            value={pullName}
            onChange={(e) => setPullName(e.target.value)}
            placeholder="Enter model name (e.g. llama3, mistral)..."
            className="w-full bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-xl pl-4 pr-10 py-2.5 text-xs text-[var(--color-text)] focus:outline-none focus:border-[var(--color-accent)]"
          />
          <Box size={14} className="absolute right-4 top-3 text-[var(--color-text-faint)]" />
        </div>
        <button
          onClick={handlePull}
          disabled={isPulling || !pullName}
          className="px-6 py-2.5 bg-[var(--color-accent)] text-white text-[10px] font-black uppercase rounded-xl hover:opacity-90 transition-all disabled:opacity-50 disabled:grayscale"
        >
          {isPulling ? <Loader2 size={14} className="animate-spin" /> : 'Pull Model'}
        </button>
      </div>

      {/* Progress Bar for Pulls */}
      {progress && (
        <div className="p-4 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-accent)]/20 animate-in fade-in zoom-in-95">
          <div className="flex justify-between mb-2 text-[10px] font-bold uppercase tracking-widest text-[var(--color-accent)]">
            <span>{progress.status}</span>
            <span>{progress.percent.toFixed(1)}%</span>
          </div>
          <div className="h-1.5 bg-[var(--color-bg)] rounded-full overflow-hidden border border-[var(--color-border)]">
            <div
              className="h-full bg-[var(--color-accent)] transition-all duration-300"
              style={{ width: `${progress.percent}%` }}
            />
          </div>
        </div>
      )}

      {/* Models Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        {(status?.available_models || []).map((model) => (
          <div
            key={model}
            className={cn(
              "p-4 rounded-xl border flex items-center justify-between transition-all group",
              status?.model === model
                ? "bg-[var(--color-accent)]/10 border-[var(--color-accent)]/30"
                : "bg-[var(--color-panel-2)] border-[var(--color-border)]/50 hover:border-[var(--color-accent)]/50"
            )}
          >
            <div className="flex items-center gap-3 min-w-0">
              <div className={cn(
                "w-9 h-9 rounded-lg flex items-center justify-center shrink-0 border",
                status?.model === model ? "bg-[var(--color-accent)] text-white border-white/10" : "bg-[var(--color-bg)] text-[var(--color-text-faint)] border-[var(--color-border)]"
              )}>
                <Database size={16} />
              </div>
              <div className="min-w-0">
                <div className="flex items-center gap-2">
                  <p className="text-xs font-bold text-[var(--color-text)] truncate">{model}</p>
                  {status?.model === model && <CheckCircle2 size={12} className="text-[var(--color-accent)]" />}
                </div>
                <p className="text-[10px] text-[var(--color-text-dim)] uppercase tracking-tighter">Local Artifact</p>
              </div>
            </div>

            <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
              <button
                onClick={() => handleDelete(model)}
                className="p-2 rounded-lg hover:bg-danger/10 text-[var(--color-text-faint)] hover:text-danger transition-all"
                title="Remove Model"
              >
                <Trash2 size={14} />
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
