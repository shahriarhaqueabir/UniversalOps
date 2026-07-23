import { useQuery } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { BrainCircuit, Cpu, Terminal, Box, GitBranch, CheckCircle2, AlertCircle } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { CapabilityInfo } from '@/types'

/**
 * Tools the app actually invokes at runtime.
 * Grouped by subsystem so users see exactly what touches their workstation.
 */
const RUNTIME_TOOLS: { id: string; label: string; icon: any; category: string }[] = [
  { id: 'ollama',      label: 'Ollama',                icon: BrainCircuit, category: 'AI' },
  { id: 'nvidia-smi',  label: 'NVIDIA SMI',            icon: Cpu,          category: 'Hardware' },
  { id: 'powershell',  label: 'Windows PowerShell',    icon: Terminal,     category: 'System' },
  { id: 'pwsh',        label: 'PowerShell 7',          icon: Terminal,     category: 'System' },
  { id: 'docker',      label: 'Docker',                icon: Box,          category: 'DevOps' },
  { id: 'git',         label: 'Git',                   icon: GitBranch,    category: 'DevOps' },
]

const CATEGORIES = ['AI', 'Hardware', 'System', 'DevOps']

/**
 * CapabilityMatrix — Shows only the external tools the app actually uses at runtime.
 * No external links, no browse-for-path — just visibility into what touches the workstation.
 */
export function CapabilityMatrix() {
  const { call } = useBackend()

  const { data: caps = [], isLoading } = useQuery<CapabilityInfo[]>({
    queryKey: ['system-capabilities'],
    queryFn: async () => {
      const res = await call('App.GetSystemCapabilities')
      return (res as CapabilityInfo[]) || []
    },
  })

  if (isLoading) {
    return <div className="p-8 text-center text-xs text-[var(--color-text-faint)] animate-pulse">Scanning system tools...</div>
  }

  const capMap = new Map(caps.map(c => [c.id, c]))

  return (
    <div className="space-y-6">
      {CATEGORIES.map(category => {
        const tools = RUNTIME_TOOLS.filter(t => t.category === category)
        if (tools.length === 0) return null
        return (
          <div key={category}>
            <h4 className="text-[10px] font-bold uppercase tracking-[0.15em] text-[var(--color-text-faint)] mb-3">{category}</h4>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              {tools.map(tool => {
                const cap = capMap.get(tool.id)
                const available = cap?.available ?? false
                const Icon = tool.icon
                return (
                  <div
                    key={tool.id}
                    className={cn(
                      'flex items-center gap-3 p-5 rounded-xl border transition-all duration-300',
                      available
                        ? 'bg-[var(--color-accent)]/5 border-[var(--color-accent)]/20 shadow-sm'
                        : 'bg-[var(--color-panel-3)]/30 border-[var(--color-border)]/50 grayscale opacity-60'
                    )}
                  >
                    <div className={cn(
                      'w-10 h-10 rounded-lg flex items-center justify-center shrink-0 border',
                      available
                        ? 'bg-[var(--color-accent)]/10 text-[var(--color-accent)] border-[var(--color-accent)]/20'
                        : 'bg-[var(--color-bg)] text-[var(--color-text-faint)] border-[var(--color-border)]'
                    )}>
                      <Icon size={20} />
                    </div>

                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2">
                        <p className="text-xs font-bold uppercase tracking-tight text-[var(--color-text)] truncate">
                          {tool.label}
                        </p>
                        {available ? (
                          <CheckCircle2 size={12} className="text-[var(--color-success)]" />
                        ) : (
                          <AlertCircle size={12} className="text-[var(--color-text-faint)]" />
                        )}
                      </div>
                      <p className="text-[10px] text-[var(--color-text-dim)] font-mono truncate mt-0.5">
                        {available ? (cap?.path || 'Detected') : 'Not detected'}
                      </p>
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        )
      })}
    </div>
  )
}
