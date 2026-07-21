import React, { useState } from 'react'
import {
  CheckCircle2,
  XCircle,
  ChevronDown,
  ChevronUp,
  ExternalLink,
  RefreshCw,
  Info
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

interface DependencyItemProps {
  info: CapabilityInfo
  name: string
  description: string
  url: string
  instructions: string[]
  onVerify: (id: string) => Promise<void>
}

const DependencyItem = ({ info, name, description, url, instructions, onVerify }: DependencyItemProps) => {
  const [expanded, setExpanded] = useState(false)
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
      <div
        className="p-4 flex items-center justify-between cursor-pointer"
        onClick={() => setExpanded(!expanded)}
      >
        <div className="flex items-center gap-4">
          <div className={cn(
            "w-10 h-10 rounded-lg flex items-center justify-center",
            info.available ? "bg-[var(--color-success)]/10 text-[var(--color-success)]" : "bg-[var(--color-panel-2)] text-[var(--color-text-faint)]"
          )}>
            {info.available ? <CheckCircle2 size={20} /> : <XCircle size={20} />}
          </div>
          <div>
            <div className="flex items-center gap-2">
              <p className="font-bold text-[var(--color-text)] text-sm">{name}</p>
              {info.available && (
                <span className="text-[10px] px-1.5 py-0.5 rounded bg-[var(--color-success)]/10 text-[var(--color-success)] font-mono">
                  {info.version || 'Ready'}
                </span>
              )}
            </div>
            <p className="text-[11px] text-[var(--color-text-faint)]">{description}</p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          {!info.available && (
            <button
              onClick={(e) => { e.stopPropagation(); setExpanded(!expanded); }}
              className="text-[10px] font-bold uppercase tracking-wider text-[var(--color-accent)] hover:underline"
            >
              Setup Guide
            </button>
          )}
          {expanded ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
        </div>
      </div>

      {expanded && (
        <div className="px-4 pb-6 pt-2 border-t border-[var(--color-border)] space-y-4 animate-in fade-in slide-in-from-top-2 duration-300">
          <div className="space-y-2">
            <p className="text-[11px] font-bold text-[var(--color-text-faint)] uppercase tracking-wider flex items-center gap-2">
              <Info size={12} /> Instructions
            </p>
            <ul className="space-y-1.5">
              {instructions.map((step, i) => (
                <li key={i} className="text-[11px] text-[var(--color-text-dim)] flex gap-2">
                  <span className="text-[var(--color-accent)] font-bold">{i + 1}.</span>
                  {step}
                </li>
              ))}
            </ul>
          </div>

          <div className="flex items-center gap-3">
            <a
              href={url}
              target="_blank"
              rel="noopener noreferrer"
              onClick={(e) => e.stopPropagation()}
              className="flex items-center gap-2 px-4 py-2 rounded-lg bg-[var(--color-panel-2)] border border-[var(--color-border)] text-[11px] font-bold text-[var(--color-text)] hover:bg-[var(--color-panel-3)] transition-all"
            >
              <ExternalLink size={14} />
              Download {name}
            </a>
            <button
              onClick={handleVerify}
              disabled={verifying}
              className="flex items-center gap-2 px-4 py-2 rounded-lg bg-[var(--color-accent)] text-white text-[11px] font-bold hover:brightness-110 active:scale-95 transition-all disabled:opacity-50"
            >
              {verifying ? <RefreshCw size={14} className="animate-spin" /> : <RefreshCw size={14} />}
              Verify Now
            </button>
          </div>
        </div>
      )}
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
        toast.error(`${id} not found. Please ensure it is installed and running.`)
      }
    } catch (err: any) {
      toast.error(`Verification failed: ${err.message}`)
    }
  }

  const getMetadata = (id: string) => {
    switch(id) {
      case 'ollama':
        return {
          name: 'Ollama AI Engine',
          description: 'Local assistant brain for system briefings and RCA.',
          url: 'https://ollama.com',
          instructions: [
            'Download and install Ollama from the official website.',
            'Ensure the Ollama application is running in your system tray.',
            'Wait for the server to initialize (usually a few seconds).',
            'Click "Verify Now" to link it with AllOpsFull.'
          ]
        }
      case 'lhm':
        return {
          name: 'LibreHardwareMonitor',
          description: 'Kernel-level sensors for temperatures and fan speeds.',
          url: 'https://github.com/LibreHardwareMonitor/LibreHardwareMonitor/releases',
          instructions: [
            'Download the latest release ZIP and extract it.',
            'Run LibreHardwareMonitor.exe as Administrator.',
            'Go to Options -> Remote Control -> Enable WMI Provider.',
            'Click "Verify Now" below.'
          ]
        }
      case 'nvidia-smi':
        return {
          name: 'NVIDIA GPU Tools',
          description: 'NVIDIA-specific utilization and thermal monitoring.',
          url: 'https://www.nvidia.com/Download/index.aspx',
          instructions: [
            'Ensure you have the latest NVIDIA proprietary drivers installed.',
            'These tools are usually bundled with the driver package.',
            'Reboot your system if you just installed them.'
          ]
        }
      default:
        return {
          name: id.toUpperCase(),
          description: 'System capability tool.',
          url: '#',
          instructions: ['Ensure the tool is installed and available in your system PATH.']
        }
    }
  }

  // Filter to show major tools first
  const priorityIds = ['ollama', 'lhm', 'nvidia-smi']
  const sorted = [...dependencies].sort((a, b) => {
    const aIdx = priorityIds.indexOf(a.id)
    const bIdx = priorityIds.indexOf(b.id)
    if (aIdx !== -1 && bIdx === -1) return -1
    if (aIdx === -1 && bIdx !== -1) return 1
    return 0
  })

  return (
    <div className="space-y-3">
      {sorted.map(dep => (
        <DependencyItem
          key={dep.id}
          info={dep}
          {...getMetadata(dep.id)}
          onVerify={handleVerify}
        />
      ))}
    </div>
  )
}
