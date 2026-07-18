import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { Box, Network, BrainCircuit, GitBranch, Terminal, CheckCircle2, AlertCircle, FolderSearch } from 'lucide-react'
import { cn } from '@/lib/utils'
import { toast } from 'sonner'
import type { CapabilityInfo } from '@/types'

const capIcons: Record<string, any> = {
  nmap: Network,
  docker: Box,
  ollama: BrainCircuit,
  git: GitBranch,
  pwsh: Terminal,
}

const capLabels: Record<string, string> = {
  nmap: 'Network Mapper (Nmap)',
  docker: 'Docker Engine',
  ollama: 'Local AI (Ollama)',
  git: 'Git CLI',
  pwsh: 'PowerShell 7',
}

/**
 * CapabilityMatrix — A high-density grid showing discovered tools and binaries.
 * Implements the "Capability Gateway" UI with support for manual path overrides.
 */
export function CapabilityMatrix() {
  const { call } = useBackend()
  const queryClient = useQueryClient()

  const { data: caps = [], isLoading } = useQuery<CapabilityInfo[]>({
    queryKey: ['system-capabilities'],
    queryFn: async () => {
      const res = await call('App.GetSystemCapabilities')
      return (res as CapabilityInfo[]) || []
    },
  })

  const handleBrowse = async (id: string) => {
    try {
      const path = await call('App.OpenFileDialog', `Select ${id} executable`, ['Executables|*.exe;*.sh;*'])
      if (path) {
        await call('App.SetCapabilityOverride', id, path)
        queryClient.invalidateQueries({ queryKey: ['system-capabilities'] })
        toast.success(`Updated path for ${id}`)
      }
    } catch (err) {
      toast.error('Failed to select file')
    }
  }

  if (isLoading) {
    return <div className="p-8 text-center text-xs text-[var(--color-text-faint)] animate-pulse">Scanning system PATH...</div>
  }

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
      {caps.map((cap) => {
        const Icon = capIcons[cap.id] || Terminal
        return (
          <div
            key={cap.id}
            className={cn(
              'flex items-center gap-3 p-3 rounded-xl border transition-all duration-300 group',
              cap.available
                ? 'bg-[var(--color-accent)]/5 border-[var(--color-accent)]/20 shadow-sm'
                : 'bg-[var(--color-panel-3)]/30 border-[var(--color-border)]/50 grayscale opacity-60'
            )}
          >
            <div className={cn(
              'w-10 h-10 rounded-lg flex items-center justify-center shrink-0 border',
              cap.available
                ? 'bg-[var(--color-accent)]/10 text-[var(--color-accent)] border-[var(--color-accent)]/20'
                : 'bg-[var(--color-bg)] text-[var(--color-text-faint)] border-[var(--color-border)]'
            )}>
              <Icon size={20} />
            </div>

            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2">
                <p className="text-xs font-bold uppercase tracking-tight text-[var(--color-text)] truncate">
                  {capLabels[cap.id] || cap.id}
                </p>
                {cap.available ? (
                  <CheckCircle2 size={12} className="text-[var(--color-success)]" />
                ) : (
                  <AlertCircle size={12} className="text-[var(--color-text-faint)]" />
                )}
              </div>
              <p className="text-[10px] text-[var(--color-text-dim)] font-mono truncate mt-0.5">
                {cap.available ? cap.path : 'Executable not found'}
              </p>
            </div>

            <button
              onClick={() => handleBrowse(cap.id)}
              className="p-2 rounded-lg bg-[var(--color-bg)] border border-[var(--color-border)] opacity-0 group-hover:opacity-100 transition-opacity hover:border-[var(--color-accent)] text-[var(--color-text-faint)] hover:text-[var(--color-accent)]"
              title="Manually set path"
            >
              <FolderSearch size={14} />
            </button>
          </div>
        )
      })}
    </div>
  )
}
