import React, { useState } from 'react'
import {
  CheckCircle2,
  XCircle,
  RefreshCw,
  BrainCircuit,
  Cpu,
  Terminal,
  Box,
  GitBranch,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { toast } from 'sonner'
import { useBackend } from '@/hooks/useBackend'

export interface CapabilityInfo {
  id: string
  available: boolean
  path: string
  version: string
}

/** Tools the app actually invokes at runtime — the only ones worth showing. */
const RUNTIME_TOOLS: { id: string; name: string; purpose: string; icon: any }[] = [
  { id: 'ollama',      name: 'Ollama',      purpose: 'AI inference engine for system briefings and RCA', icon: BrainCircuit },
  { id: 'nvidia-smi',  name: 'NVIDIA SMI',  purpose: 'GPU utilization and performance monitoring',           icon: Cpu },
  { id: 'powershell',  name: 'Windows PowerShell', purpose: 'Legacy WinRM, AD, and Windows API fallback',   icon: Terminal },
  { id: 'pwsh',        name: 'PowerShell 7',   purpose: 'Modern system queries and execution engine',        icon: Terminal },
  { id: 'docker',      name: 'Docker',       purpose: 'Container status observation',                       icon: Box },
  { id: 'git',         name: 'Git',          purpose: 'Repository status observation',                      icon: GitBranch },
]

interface DependencyItemProps {
  info: CapabilityInfo
  name: string
  purpose: string
  icon: any
  onVerify: (id: string) => Promise<void>
}

const DependencyItem = ({ info, name, purpose, icon: Icon, onVerify }: DependencyItemProps) => {
  const [verifying, setVerifying] = useState(false)

  const handleVerify = async (e: React.MouseEvent) => {
    e.stopPropagation()
    setVerifying(true)
    try {
      await onVerify(info.id)
    } finally {
      setVerifying(false)
    }
  }

  return (
    <div className={cn(
      "border rounded-xl transition-all duration-300",
      info.available
        ? "bg-[var(--color-panel-2)] border-[var(--color-success)]/20"
        : "bg-[var(--color-panel-3)] border-[var(--color-border)]"
    )}>
      <div className="p-4 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <div className={cn(
            "w-10 h-10 rounded-lg flex items-center justify-center",
            info.available ? "bg-[var(--color-success)]/10 text-[var(--color-success)]" : "bg-[var(--color-panel-2)] text-[var(--color-text-faint)]"
          )}>
            {info.available ? <CheckCircle2 size={20} /> : <XCircle size={20} />}
          </div>
          <div>
            <div className="flex items-center gap-2">
              <Icon size={14} className="text-[var(--color-text-dim)]" />
              <p className="font-bold text-[var(--color-text)] text-sm">{name}</p>
              {info.available && info.version && (
                <span className="text-[10px] px-1.5 py-0.5 rounded bg-[var(--color-success)]/10 text-[var(--color-success)] font-mono">
                  {info.version}
                </span>
              )}
            </div>
            <p className="text-[11px] text-[var(--color-text-faint)]">{purpose}</p>
          </div>
        </div>
        <button
          onClick={handleVerify}
          disabled={verifying}
          className="p-2 rounded-lg bg-[var(--color-bg)] border border-[var(--color-border)] text-[var(--color-text-faint)] hover:text-[var(--color-accent)] hover:border-[var(--color-accent)] transition-all disabled:opacity-50"
          title="Re-scan for this tool"
        >
          {verifying ? <RefreshCw size={14} className="animate-spin" /> : <RefreshCw size={14} />}
        </button>
      </div>
    </div>
  )
}

interface DependencyChecklistProps {
  dependencies: CapabilityInfo[]
  onRefresh: () => Promise<void>
}

export const DependencyChecklist = ({ dependencies, onRefresh }: DependencyChecklistProps) => {
  const { call } = useBackend()

  const handleVerify = async (id: string) => {
    try {
      const res = await call('App.VerifyCapability', id) as CapabilityInfo
      if (res.available) {
        toast.success(`${id} detected successfully!`, {
          icon: <CheckCircle2 className="text-green-500" />
        })
        await onRefresh()
      } else {
        toast.error(`${id} not found. Some features may be unavailable.`)
      }
    } catch (err: any) {
      toast.error(`Verification failed: ${err.message}`)
    }
  }

  // Only show tools the app actually invokes — build a lookup from backend results
  const depMap = new Map(dependencies.map(d => [d.id, d]))

  return (
    <div className="space-y-3">
      {RUNTIME_TOOLS.map(tool => (
        <DependencyItem
          key={tool.id}
          info={depMap.get(tool.id) || { id: tool.id, available: false, path: '', version: '' }}
          name={tool.name}
          purpose={tool.purpose}
          icon={tool.icon}
          onVerify={handleVerify}
        />
      ))}
    </div>
  )
}
