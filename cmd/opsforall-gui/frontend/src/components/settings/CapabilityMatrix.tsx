import { useQuery } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { Box, Network, BrainCircuit, Github, Terminal, CheckCircle2, AlertCircle } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { CapabilityInfo } from '@/types'

const capIcons: Record<string, any> = {
  nmap: Network,
  docker: Box,
  ollama: BrainCircuit,
  git: Github,
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
 * Implements the "Capability Gateway" UI by visualizing unlocked workstation powers.
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
              'flex items-center gap-3 p-3 rounded-xl border transition-all duration-300',
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
                <p className="text-xs font-black uppercase tracking-tight text-[var(--color-text)] truncate">
                  {capLabels[cap.id] || cap.id}
                </p>
                {cap.available ? (
                  <CheckCircle2 size={12} className="text-[var(--color-success)]" />
                ) : (
                  <AlertCircle size={12} className="text-[var(--color-text-faint)]" />
                )}
              </div>
              <p className="text-[10px] text-[var(--color-text-dim)] font-mono truncate mt-0.5">
                {cap.available ? cap.path : 'Executable not found in %PATH%'}
              </p>
            </div>
          </div>
        )
      })}
    </div>
  )
}
