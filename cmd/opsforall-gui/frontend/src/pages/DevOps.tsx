import { useState, useRef, useCallback, useEffect, memo } from 'react'
import { motion } from 'motion/react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import {
  Terminal, Server, Play, Trash2,
  PlayCircle, StopCircle, Zap, Activity, Globe,
  TerminalSquare, Box, Container, Variable,
  RefreshCw,
  Shield, Layers, RotateCcw,
  Upload,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import * as Tabs from '@radix-ui/react-tabs'
import { useBackend } from '@/hooks/useBackend'
import { useEvents } from '@/hooks/useEvents'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { ConfirmDialog } from '@/components/dialogs/ConfirmDialog'
import { StatusBadge } from '@/components/ui/StatusBadge'
import { SearchInput } from '@/components/ui/SearchInput'
import { Panel } from '@/components/ui/Panel'
import { nanoid } from 'nanoid'
import type {
  CommandResult, ServiceEntry, ContainerSummary,
  LocalServer, EnvironmentInfo, DockerStatus, KubernetesStatus,
  ServiceGroupSummary,
  DockerStatsEntry, DockerComposeProject, DockerNetworkInfo, DockerVolumeInfo,
  K8sResourceItem, K8sRolloutStatus, K8sEvent, K8sNamespaceInfo,
} from '@/types'

function stripAnsi(text: string): string {
  // eslint-disable-next-line no-control-regex
  return text.replace(/\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])/g, '')
}

type TabId = 'overview' | 'terminal' | 'powershell-pro' | 'services' | 'docker' | 'servers' | 'environment' | 'kubernetes' | 'diagnostics'

const ActionButton = memo(function ActionButton({ icon, label, onClick, variant, disabled }: { icon: React.ReactNode; label: string; onClick: () => void; variant?: string; disabled?: boolean }) {
  const colors: Record<string, string> = {
    primary: 'bg-accent text-white hover:bg-accent/90',
    danger: 'bg-danger/20 text-danger hover:bg-danger/30',
    warning: 'bg-warning/20 text-warning hover:bg-warning/30',
    ghost: 'text-text-dim hover:bg-panel-3',
  }
  return (
    <button onClick={onClick} disabled={disabled}
      className={cn('flex items-center gap-2 px-4 py-2 rounded-xl text-sm font-bold transition-all disabled:opacity-50', colors[variant || 'ghost'])}>
      {icon}{label}
    </button>
  )
})

const TAB_LIST = [
  { id: 'overview', label: 'Overview', icon: <Activity size={20} className="text-accent" /> },
  { id: 'terminal', label: 'Terminal', icon: <Terminal size={20} /> },
  { id: 'powershell-pro', label: 'PowerShell', icon: <Zap size={20} className="text-warning" /> },
  { id: 'docker', label: 'Docker', icon: <Container size={20} /> },
  { id: 'kubernetes', label: 'K8s', icon: <Layers size={20} /> },
  { id: 'diagnostics', label: 'Health', icon: <Shield size={20} /> },
  { id: 'services', label: 'Services', icon: <Server size={20} /> },
  { id: 'servers', label: 'Servers', icon: <Globe size={20} /> },
  { id: 'environment', label: 'Env', icon: <Variable size={20} /> },
] as const

export function DevOps() {
  const [activeTab, setActiveTab] = useState<TabId>('overview')

  return (
    <div className="flex flex-col h-full bg-[var(--color-bg)] animate-in fade-in duration-500">
      <div className="py-8 border-b border-[var(--color-border)] bg-[var(--color-panel-2)]/50 flex items-center justify-between px-10">
        <div>
          <div className="flex items-center gap-3 mb-2">
            <div className="w-8 h-8 rounded-lg bg-accent/10 flex items-center justify-center text-accent border border-accent/20">
               <TerminalSquare size={18} />
            </div>
            <h1 className="text-sm font-black text-[var(--color-text)] uppercase tracking-[0.25em]">Unified Control Center</h1>
          </div>
          <p className="text-3xl font-bold text-[var(--color-text)] tracking-tight">DevOps Console</p>
          <p className="text-[var(--color-text-dim)] text-xs font-semibold uppercase tracking-widest mt-2">Build, deploy, observe, and operate with local high-density telemetry</p>
        </div>
      </div>

      <Tabs.Root defaultValue="overview" onValueChange={(v) => setActiveTab(v as TabId)} className="flex-1 flex flex-col min-w-0">
        <Tabs.List className="flex border-b border-[var(--color-border)] bg-[var(--color-panel)] px-6 overflow-x-auto no-scrollbar">
          {TAB_LIST.map((tab) => (
            <Tabs.Trigger key={tab.id} value={tab.id}
              data-automation-id={`devops-tab-${tab.id}`}
              className={cn(
                'flex items-center gap-3 px-6 py-5 text-sm font-bold transition-all border-b-2 border-transparent whitespace-nowrap relative',
                activeTab === tab.id ? 'text-accent' : 'text-text-faint hover:text-text hover:bg-[var(--color-sidebar-hover)]',
              )}>
              {tab.icon}
              <span className="uppercase tracking-widest text-[10px] font-black">{tab.label}</span>
              {activeTab === tab.id && (
                <motion.div
                  layoutId="devops-tab-indicator"
                  className="absolute bottom-0 left-0 right-0 h-0.5 bg-accent"
                  transition={{ type: 'spring', stiffness: 500, damping: 30 }}
                />
              )}
            </Tabs.Trigger>
          ))}
        </Tabs.List>

        <div className="flex-1 overflow-hidden">
          <Tabs.Content value="overview" className="h-full"><OverviewTab /></Tabs.Content>
          <Tabs.Content value="terminal" className="h-full"><TerminalTab /></Tabs.Content>
          <Tabs.Content value="powershell-pro" className="h-full"><PowerShellProTab /></Tabs.Content>
          <Tabs.Content value="docker" className="h-full"><DockerTabExpanded /></Tabs.Content>
          <Tabs.Content value="kubernetes" className="h-full"><KubernetesTab /></Tabs.Content>
          <Tabs.Content value="diagnostics" className="h-full"><DiagnosticsTab /></Tabs.Content>
          <Tabs.Content value="services" className="h-full"><ServicesTab /></Tabs.Content>
          <Tabs.Content value="servers" className="h-full"><ServersTab /></Tabs.Content>
          <Tabs.Content value="environment" className="h-full"><EnvironmentTab /></Tabs.Content>
        </div>
      </Tabs.Root>
    </div>
  )
}

function DiagnosticsTab() {
  const { call } = useBackend()
  const [isRunning, setIsRunning] = useState(false)
  const [result, setResult] = useState<DevOpsDiagResult | null>(null)

  const runCheck = async () => {
    setIsRunning(true)
    try {
      const res = await call('DevOps.RunDevOpsDiagnostics') as DevOpsDiagResult
      setResult(res)
      toast.success('Health check completed')
    } catch {
      toast.error('Health check failed')
    } finally {
      setIsRunning(false)
    }
  }

  return (
    <div className="flex flex-col h-full p-10 space-y-8 overflow-y-auto">
      <div className="flex items-center justify-between bg-panel border border-border p-8 rounded-3xl shadow-xl">
        <div className="flex items-center gap-6">
          <div className="w-14 h-14 rounded-2xl bg-accent/10 flex items-center justify-center text-accent border border-accent/20">
            <Shield size={28} />
          </div>
          <div>
            <h3 className="text-xl font-bold text-text uppercase tracking-tight">System Node Health</h3>
            <p className="text-sm text-text-dim font-medium">Verify toolchain integrity and local daemon status</p>
          </div>
        </div>
        <button
          onClick={runCheck}
          disabled={isRunning}
          className="flex items-center gap-3 px-8 py-4 bg-accent text-white rounded-2xl font-black uppercase tracking-widest hover:opacity-90 transition-all shadow-lg active:scale-95 disabled:opacity-50"
        >
          {isRunning ? <RefreshCw size={20} className="animate-spin" /> : <Play size={20} fill="currentColor" />}
          {isRunning ? 'Running Analysis...' : 'Run Health Check'}
        </button>
      </div>

      {result ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {result.checks.map((check, i) => (
            <div key={i} className={cn("p-6 rounded-2xl border transition-all hover:scale-[1.02]",
              check.status === 'pass' ? "bg-success/5 border-success/20 hover:border-success/40" :
              check.status === 'warn' ? "bg-warning/5 border-warning/20 hover:border-warning/40" :
              "bg-danger/5 border-danger/20 hover:border-danger/40")}>
              <div className="flex items-center justify-between mb-4">
                <span className="text-xs font-black uppercase tracking-widest text-text-faint">{check.name}</span>
                <StatusBadge status={check.status} />
              </div>
              <p className="text-lg font-bold text-text mb-2 tracking-tight">{check.value}</p>
              <p className="text-xs text-text-dim font-medium leading-relaxed">{check.message}</p>
            </div>
          ))}
        </div>
      ) : (
        <div className="flex-1 flex flex-col items-center justify-center opacity-30 py-20">
           <Shield size={80} className="text-accent mb-6" />
           <p className="text-xl font-black uppercase tracking-[0.3em]">Ready for Analysis</p>
           <p className="text-sm font-bold text-text-dim mt-2 tracking-widest">Execute check to verify node integrity</p>
        </div>
      )}
    </div>
  )
}

interface DevOpsDiagCheck {
  name: string
  status: string
  message: string
  value: string
}

interface DevOpsDiagResult {
  checks: DevOpsDiagCheck[]
  score: number
  timestamp: string
}

function OverviewTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: dockerData } = useQuery<DockerStatus>({
    queryKey: ['devops-docker-status'],
    queryFn: async () => {
      try { return await call('DevOps.GetDockerStatus') as DockerStatus }
      catch { return { installed: false, running: false, version: '', containers: { running: 0, stopped: 0, failed: 0, total: 0, containers: [] } } }
    },
    refetchInterval: refreshInterval,
  })

  const { data: k8sData } = useQuery<KubernetesStatus>({
    queryKey: ['devops-k8s-status'],
    queryFn: async () => {
      try { return await call('DevOps.GetKubernetesStatus') as KubernetesStatus }
      catch { return { installed: false, connected: false, cluster: '', nodes: 0, pods: 0 } }
    },
    refetchInterval: refreshInterval,
  })

  const { data: summary } = useQuery<ServiceGroupSummary>({
    queryKey: ['devops-service-summary'],
    queryFn: async () => {
      try { return await call('DevOps.GetServiceGroupSummary') as ServiceGroupSummary }
      catch { return { databases: 0, messageQueues: 0, webServers: 0, containers: 0, other: 0, running: 0, stopped: 0 } }
    },
    refetchInterval: refreshInterval,
  })

  return (
    <div className="flex flex-col h-full space-y-6 overflow-y-auto p-10">
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <Panel variant="elevated" padding="lg" category="devops" className="group hover:border-accent/30 transition-all">
          <div className="flex items-center justify-between mb-8">
            <div className="flex items-center gap-4">
              <div className="w-10 h-10 rounded-xl bg-accent/10 border border-accent/20 flex items-center justify-center text-accent group-hover:scale-110 transition-transform">
                <Box size={20} />
              </div>
              <div>
                <h3 className="text-xs font-black text-text uppercase tracking-[0.2em]">Docker Engine</h3>
                <p className="text-[10px] text-text-faint font-bold uppercase tracking-wider">{dockerData?.version || 'N/A'}</p>
              </div>
            </div>
            <span className={cn('text-[10px] font-black px-3 py-1.5 rounded-full uppercase tracking-[0.15em] border',
              dockerData?.installed ? (dockerData?.running ? 'bg-success/10 text-success border-success/30' : 'bg-warning/10 text-warning border-warning/30') : 'bg-text-faint/10 text-text-faint border-text-faint/30')}>
              {dockerData?.installed ? (dockerData?.running ? 'Running' : 'Installed') : 'Not Found'}
            </span>
          </div>
          {dockerData?.installed ? (
            <div className="grid grid-cols-4 gap-4">
              {[{ label: 'Running', value: dockerData.containers?.running ?? 0, cl: 'text-success' },
                { label: 'Stopped', value: dockerData.containers?.stopped ?? 0, cl: 'text-warning' },
                { label: 'Failed', value: dockerData.containers?.failed ?? 0, cl: 'text-danger' },
                { label: 'Total', value: dockerData.containers?.total ?? 0, cl: 'text-text' }].map(item => (
                  <div key={item.label} className="flex flex-col items-center bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-xl p-5 transition-all hover:border-accent/20">
                    <span className={cn('text-3xl font-black tabular-nums tracking-tighter', item.cl)}>{item.value}</span>
                    <span className="text-[9px] font-black text-text-faint uppercase tracking-widest mt-1">{item.label}</span>
                  </div>
              ))}
            </div>
          ) : <p className="text-text-faint text-xs font-medium italic">Docker subsystem not detected on this node.</p>}
        </Panel>

        <Panel variant="elevated" padding="lg" category="devops" className="group hover:border-accent/30 transition-all">
          <div className="flex items-center justify-between mb-8">
            <div className="flex items-center gap-4">
              <div className="w-10 h-10 rounded-xl bg-accent/10 border border-accent/20 flex items-center justify-center text-accent group-hover:scale-110 transition-transform">
                <Layers size={20} />
              </div>
              <div>
                <h3 className="text-xs font-black text-text uppercase tracking-[0.2em]">Kubernetes</h3>
                <p className="text-[10px] text-text-faint font-bold uppercase tracking-wider">{k8sData?.cluster || 'Standalone Node'}</p>
              </div>
            </div>
            <span className={cn('text-[10px] font-black px-3 py-1.5 rounded-full uppercase tracking-[0.15em] border',
              k8sData?.installed ? (k8sData?.connected ? 'bg-success/10 text-success border-success/30' : 'bg-warning/10 text-warning border-warning/30') : 'bg-text-faint/10 text-text-faint border-text-faint/30')}>
              {k8sData?.installed ? (k8sData?.connected ? 'Connected' : 'Offline') : 'Not Found'}
            </span>
          </div>
          {k8sData?.installed ? (
            <div className="grid grid-cols-2 gap-4">
              <div className="flex flex-col items-center bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-xl p-5 transition-all hover:border-accent/20">
                <span className="text-3xl font-black tabular-nums text-text tracking-tighter">{k8sData.nodes}</span>
                <span className="text-[9px] font-black text-text-faint uppercase tracking-widest mt-1">Nodes</span>
              </div>
              <div className="flex flex-col items-center bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-xl p-5 transition-all hover:border-accent/20">
                <span className="text-3xl font-black tabular-nums text-accent tracking-tighter">{k8sData.pods}</span>
                <span className="text-[9px] font-black text-text-faint uppercase tracking-widest mt-1">Pods</span>
              </div>
            </div>
          ) : <p className="text-text-faint text-xs font-medium italic">kubectl orchestrator not detected on this node.</p>}
        </Panel>
      </div>

      <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-6 shadow-2xl">
        <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
          {[
            { label: 'Databases', value: summary?.databases ?? 0, cl: 'text-accent' },
            { label: 'Queues', value: summary?.messageQueues ?? 0, cl: 'text-accent-2' },
            { label: 'Web Servers', value: summary?.webServers ?? 0, cl: 'text-success' },
            { label: 'Running', value: summary?.running ?? 0, cl: 'text-success' },
            { label: 'Stopped', value: summary?.stopped ?? 0, cl: 'text-danger' },
          ].map(kpi => (
            <div key={kpi.label} className="flex flex-col items-center gap-1 bg-[var(--color-panel-2)] border border-[var(--color-border)]/50 rounded-xl p-6 transition-all hover:border-accent/20">
              <span className={cn('text-3xl font-black tabular-nums tracking-tighter', kpi.cl)}>{kpi.value}</span>
              <span className="text-[9px] font-black text-text-faint uppercase tracking-widest">{kpi.label}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

function TerminalTab() {
  const { call } = useBackend()
  const [input, setInput] = useState('')
  const [output, setOutput] = useState<string[]>([`Unified Terminal: Type a command...\n`])
  const [history, setHistory] = useState<string[]>([])
  const [historyIndex, setHistoryIndex] = useState(-1)
  const [isRunning, setIsRunning] = useState(false)
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [pendingCmd, setPendingCmd] = useState('')
  const [currentCmdId, setCurrentCmdId] = useState<string | null>(null)
  const outputRef = useRef<HTMLDivElement>(null)

  useEvents('cmd:line', (data: any) => {
    if (data.id === currentCmdId) setOutput(prev => [...prev, data.line])
  })

  useEvents('cmd:done', (id) => {
    if (id === currentCmdId) { setIsRunning(false); setCurrentCmdId(null) }
  })

  const runCommand = useCallback(async (cmd: string) => {
    if (!cmd.trim() || isRunning) return
    const id = nanoid()
    setIsRunning(true); setCurrentCmdId(id)
    setHistory(prev => [...prev, cmd])
    setOutput(prev => [...prev, `$ ${cmd}`])
    try {
      const res = await call('DevOps.RunCommandLive', cmd, id) as CommandResult
      if (res.error) setOutput(prev => [...prev, `\u001b[31mError: ${res.error}\u001b[0m`])
      else if (res.output) setOutput(prev => [...prev, res.output])
    } catch (err) {
      setOutput(prev => [...prev, `\u001b[31mExecution Error: ${String(err)}\u001b[0m`])
      setIsRunning(false)
    }
  }, [call, isRunning])

  const handleImportScript = async () => {
    try {
      const path = await call('App.OpenFileDialog', 'Import Execution Script', ['Scripts|*.ps1;*.sh;*.bat;*.cmd', 'All Files|*.*'])
      if (path) {
        const content = await call('App.ReadTextFile', path)
        if (content) {
          setInput(content as string)
          toast.success('Script loaded into buffer')
        }
      }
    } catch (err: any) {
      toast.error(err?.message || 'Import failed')
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      if (input.trim()) {
        const lowerCmd = input.trim().toLowerCase()
        const dangerous = ['rm ', 'del ', 'format', 'mkfs', 'kill ', 'shutdown', 'restart', 'stop-service']
        if (dangerous.some(d => lowerCmd.includes(d))) { setPendingCmd(input); setConfirmOpen(true) }
        else { runCommand(input); setInput(''); setHistoryIndex(-1) }
      }
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      if (history.length > 0) {
        const newIdx = historyIndex === -1 ? history.length - 1 : Math.max(historyIndex - 1, 0)
        setHistoryIndex(newIdx); setInput(history[newIdx])
      }
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      if (historyIndex >= 0 && historyIndex < history.length - 1) {
        const newIdx = historyIndex + 1; setHistoryIndex(newIdx); setInput(history[newIdx])
      } else { setHistoryIndex(-1); setInput('') }
    }
  }

  useEffect(() => {
    if (outputRef.current) outputRef.current.scrollTop = outputRef.current.scrollHeight
  }, [output])

  return (
    <div className="flex flex-col h-full p-8 space-y-6">
      <ConfirmDialog open={confirmOpen} title="Impactful Command Detected"
        description={`Run: "${pendingCmd}"?`} type="danger" confirmText="Execute"
        onConfirm={() => { runCommand(pendingCmd); setInput(''); setPendingCmd('') }}
        onClose={() => setConfirmOpen(false)} />
      <div className="flex items-center gap-4">
        <div className="relative flex-1 group">
          <input type="text" value={input} onChange={(e) => setInput(e.target.value)} onKeyDown={handleKeyDown}
            placeholder="Enter shell command..." disabled={isRunning}
            className="w-full bg-[var(--color-bg)] border border-[var(--color-border)] rounded-2xl pl-12 pr-6 py-4 text-sm font-mono text-text focus:outline-none focus:border-accent transition-all shadow-inner group-hover:border-accent/30" />
          <span className="absolute left-5 top-1/2 -translate-y-1/2 text-success text-sm font-black font-mono opacity-60 group-focus-within:opacity-100 transition-opacity">$</span>
        </div>
        <button onClick={handleImportScript} disabled={isRunning}
          className="p-4 bg-panel-3 border border-border text-text-dim hover:text-white rounded-2xl transition-all active:scale-95" title="Import Script">
          <Upload size={18} />
        </button>
        <button onClick={() => { if (input.trim()) runCommand(input) }} disabled={isRunning || !input.trim()}
          className="flex items-center gap-2.5 px-8 py-4 text-xs font-black uppercase tracking-widest bg-accent text-white rounded-2xl hover:bg-accent/90 disabled:opacity-50 transition-all shadow-xl active:scale-95">
          <Play size={16} /> Run Execution
        </button>
        <button onClick={() => setOutput([`Unified Terminal: Type a command...\n`])} className="px-6 py-4 text-sm font-bold text-text-faint border border-border rounded-2xl hover:bg-panel-3 transition-all hover:text-danger">
          <Trash2 size={18} />
        </button>
      </div>
      <div ref={outputRef} className="flex-1 bg-[var(--color-terminal-bg)] border border-[var(--color-border)] rounded-3xl p-8 overflow-y-auto font-mono text-sm leading-relaxed whitespace-pre-wrap shadow-2xl relative">
        <div className="absolute top-0 left-0 right-0 h-12 bg-gradient-to-b from-[var(--color-terminal-bg)] to-transparent pointer-events-none z-10" />
        <div className="relative z-0">
          {output.map((block, i) => <div key={i} className="mb-1">{stripAnsi(block)}</div>)}
          {isRunning && <div className="flex items-center gap-3 mt-4"><span className="inline-block w-2.5 h-5 bg-success animate-pulse" /><span className="text-[10px] font-black text-success uppercase tracking-[0.2em] animate-pulse">Processing Stream...</span></div>}
        </div>
      </div>
    </div>
  )
}

function PowerShellProTab() {
  const { call } = useBackend()
  const [isRunning, setIsRunning] = useState(false)
  const [output, setOutput] = useState('')
  const [selectedWorkflow, setSelectedWorkflow] = useState('')
  const [confirmOpen, setConfirmOpen] = useState(false)
  const { data: workflows = [] } = useQuery<string[]>({
    queryKey: ['devops-workflows'],
    queryFn: async () => (await call('DevOps.GetPowerShellWorkflows') as string[]) || []
  })
  const runWorkflow = async (name: string) => {
    setIsRunning(true); setOutput(`Running workflow: ${name}...\n`)
    try {
      const res = await call('DevOps.RunPowerShell', name) as CommandResult
      if (res.error) setOutput(prev => prev + `Error: ${res.error}\n`)
      else setOutput(prev => prev + res.output)
    } catch (err) { setOutput(prev => prev + `Execution Error: ${String(err)}\n`) }
    finally { setIsRunning(false) }
  }
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 h-full gap-6 overflow-hidden p-6">
      <ConfirmDialog open={confirmOpen} title="Execute PowerShell Workflow" description={`Run "${selectedWorkflow}"?`}
        onConfirm={() => runWorkflow(selectedWorkflow)} onClose={() => setConfirmOpen(false)} />
      <div className="col-span-1 space-y-3 overflow-y-auto pr-2">
        <h3 className="text-sm font-bold text-text-dim uppercase tracking-widest mb-3">Diagnostic Workflows</h3>
        {workflows.map(wf => (
          <button key={wf} onClick={() => { setSelectedWorkflow(wf); setConfirmOpen(true) }} disabled={isRunning}
            className="w-full text-left bg-panel border border-border rounded-xl p-5 transition-all hover:border-accent/50 hover:bg-accent/5 group disabled:opacity-50">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-panel-3 flex items-center justify-center text-text-dim group-hover:text-accent border border-border">
                <Zap size={16} />
              </div>
              <span className="text-sm font-bold text-text group-hover:text-accent transition-colors">{wf.replace('Invoke-Hawk', '')}</span>
            </div>
          </button>
        ))}
      </div>
      <div className="col-span-2 flex flex-col space-y-3">
        <div className="flex-1 bg-[var(--color-terminal-bg)] border border-border rounded-2xl p-6 overflow-y-auto font-mono text-sm leading-relaxed whitespace-pre shadow-inner">
          {stripAnsi(output) || 'Select a workflow to begin.'}
        </div>
        <button onClick={() => setOutput('')} className="self-end px-4 py-2 text-sm font-bold text-text-dim border border-border rounded-xl hover:bg-panel-3 transition-all">Clear</button>
      </div>
    </div>
  )
}

function DockerTabExpanded() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const [subTab, setSubTab] = useState<'containers' | 'stats' | 'compose' | 'networks' | 'volumes'>('containers')
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [pendingAction, setPendingAction] = useState<{ label: string; action: () => void } | null>(null)

  const { data: containers } = useQuery<ContainerSummary>({
    queryKey: ['devops-containers'],
    queryFn: async () => (await call('DevOps.GetContainers') as ContainerSummary) || { containers: [] },
  })

  const { data: stats } = useQuery<DockerStatsEntry[]>({
    queryKey: ['devops-docker-stats'],
    queryFn: async () => (await call('DevOps.GetDockerStats') as DockerStatsEntry[]) || [],
    enabled: subTab === 'stats',
    refetchInterval: refreshInterval,
  })

  const { data: compose } = useQuery<DockerComposeProject[]>({
    queryKey: ['devops-docker-compose'],
    queryFn: async () => (await call('DevOps.DockerComposeList') as DockerComposeProject[]) || [],
    enabled: subTab === 'compose',
  })

  const { data: networks } = useQuery<DockerNetworkInfo[]>({
    queryKey: ['devops-docker-networks'],
    queryFn: async () => (await call('DevOps.GetDockerNetworks') as DockerNetworkInfo[]) || [],
    enabled: subTab === 'networks',
  })

  const { data: volumes } = useQuery<DockerVolumeInfo[]>({
    queryKey: ['devops-docker-volumes'],
    queryFn: async () => (await call('DevOps.GetDockerVolumes') as DockerVolumeInfo[]) || [],
    enabled: subTab === 'volumes',
  })

  const dockerAction = async (action: string, id: string) => {
    try {
      await call(`DevOps.Docker${action}`, id)
    } catch (err) { console.error(err) }
  }

  return (
    <div className="flex flex-col h-full p-6 space-y-4">
      <ConfirmDialog open={confirmOpen} title="Confirm Docker Action"
        description={pendingAction?.label || ''}
        onConfirm={() => { pendingAction?.action() }}
        onClose={() => setConfirmOpen(false)} />

      <div className="flex gap-2 border-b border-border pb-2">
        {(['containers', 'stats', 'compose', 'networks', 'volumes'] as const).map(tab => (
          <button key={tab} onClick={() => setSubTab(tab)}
            className={cn('px-4 py-2 text-sm font-bold rounded-t-lg transition-colors',
              subTab === tab ? 'text-accent border-b-2 border-accent' : 'text-text-faint hover:text-text')}>
            {tab.charAt(0).toUpperCase() + tab.slice(1)}
          </button>
        ))}
        <div className="ml-auto flex gap-2">
          <ActionButton icon={<Trash2 size={14} />} label="Prune" variant="danger"
            onClick={() => { setPendingAction({ label: 'Prune unused Docker resources?', action: async () => { await call('DevOps.DockerPrune') } }); setConfirmOpen(true) }} />
        </div>
      </div>

      {subTab === 'containers' && (
        <div className="flex-1 overflow-y-auto space-y-2">
          {(containers?.containers ?? []).map(c => (
            <div key={c.id} className="bg-panel border border-border rounded-xl p-5 flex items-center justify-between">
              <div className="flex items-center gap-4">
                <span className={cn('w-2 h-2 rounded-full', c.state === 'running' ? 'bg-success' : 'bg-text-faint')} />
                <div>
                  <p className="font-bold text-sm tracking-tight">{c.name} <span className="text-[10px] font-black font-mono text-text-faint ml-2 uppercase tracking-widest">{c.id.slice(0, 12)}</span></p>
                  <p className="text-[10px] text-text-dim font-mono font-medium mt-1">{c.image} \u2022 {c.status}</p>
                </div>
              </div>
              <div className="flex items-center gap-2">
                {c.ports && <span className="text-xs font-mono text-text-faint">{c.ports}</span>}
                <StatusBadge status={c.state} />
                <div className="flex gap-1">
                  {c.state !== 'running' && <ActionButton icon={<Play size={12} />} label="" onClick={() => dockerAction('Kill', c.id)} />}
                  {/* Using Kill as workaround - the actual start/stop is via ControlContainer but that's different */}
                </div>
              </div>
            </div>
          ))}
          {containers?.containers?.length === 0 && <p className="text-text-faint text-sm text-center py-8">No containers</p>}
        </div>
      )}

      {subTab === 'stats' && (
        <div className="flex-1 overflow-y-auto">
          <table className="w-full text-left text-sm">
            <thead><tr className="text-xs text-text-dim uppercase border-b border-border">
              <th className="px-3 py-2">Name</th><th className="px-3 py-2">CPU</th><th className="px-3 py-2">Memory</th><th className="px-3 py-2">Net I/O</th><th className="px-3 py-2">Block I/O</th><th className="px-3 py-2">PIDs</th>
            </tr></thead>
            <tbody>
              {(stats ?? []).map(s => (
                <tr key={s.container_id} className="border-b border-border/20 hover:bg-panel-3">
                  <td className="px-3 py-2 font-mono text-xs">{s.name}</td>
                  <td className="px-3 py-2 font-mono text-xs">{s.cpu_percent}</td>
                  <td className="px-3 py-2 font-mono text-xs">{s.memory_usage} / {s.memory_limit}</td>
                  <td className="px-3 py-2 font-mono text-xs">{s.net_io}</td>
                  <td className="px-3 py-2 font-mono text-xs">{s.block_io}</td>
                  <td className="px-3 py-2 font-mono text-xs">{s.pid_count}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {subTab === 'compose' && (
        <div className="flex-1 overflow-y-auto space-y-3">
          {(compose ?? []).map(p => (
            <div key={p.project} className="bg-panel border border-border rounded-xl p-5">
              <div className="flex items-center justify-between mb-2">
                <h4 className="font-bold text-sm">{p.project}</h4>
                <StatusBadge status={p.status} />
              </div>
              {p.services.map(s => (
                <div key={s.name} className="flex items-center gap-3 text-xs font-mono px-4 py-1">
                  <span className="text-text-dim">{s.name}</span>
                  <StatusBadge status={s.state} />
                  {s.ports && <span className="text-text-faint">{s.ports}</span>}
                </div>
              ))}
              <div className="flex gap-2 mt-2">
                <ActionButton icon={<Play size={12} />} label="Up" variant="primary" onClick={() => dockerAction('ComposeUp', p.work_dir)} />
                <ActionButton icon={<StopCircle size={12} />} label="Down" variant="danger" onClick={() => dockerAction('ComposeDown', p.work_dir)} />
              </div>
            </div>
          ))}
        </div>
      )}

      {subTab === 'networks' && (
        <div className="flex-1 overflow-y-auto space-y-2">
          {(networks ?? []).map(n => (
            <div key={n.id} className="bg-panel border border-border rounded-xl p-5 flex items-center justify-between">
              <div>
                <p className="font-bold text-sm font-mono">{n.name}</p>
                <p className="text-xs text-text-dim">Driver: {n.driver} | Scope: {n.scope} | Subnet: {n.subnet || 'N/A'}</p>
              </div>
              <span className="text-xs text-text-faint">{n.containers} container(s)</span>
            </div>
          ))}
        </div>
      )}

      {subTab === 'volumes' && (
        <div className="flex-1 overflow-y-auto space-y-2">
          {(volumes ?? []).map(v => (
            <div key={v.name} className="bg-panel border border-border rounded-xl p-5 flex items-center justify-between">
              <div>
                <p className="font-bold text-sm font-mono">{v.name}</p>
                <p className="text-xs text-text-dim">Driver: {v.driver} | Size: {v.size || 'N/A'}</p>
              </div>
              <span className="text-xs font-mono text-text-faint">{v.mountpoint}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

function KubernetesTab() {
  const { call } = useBackend()
  const [namespace, setNamespace] = useState('')
  const [subTab, setSubTab] = useState<'deployments' | 'services' | 'pods' | 'events' | 'rollouts'>('deployments')

  const { data: namespaces = [] } = useQuery<K8sNamespaceInfo[]>({
    queryKey: ['devops-k8s-ns'],
    queryFn: async () => (await call('DevOps.GetK8sNamespaces') as K8sNamespaceInfo[]) || [],
  })

  const { data: deployments = [] } = useQuery<K8sResourceItem[]>({
    queryKey: ['devops-k8s-deploy', namespace],
    queryFn: async () => (await call('DevOps.GetK8sDeployments', namespace) as K8sResourceItem[]) || [],
    enabled: subTab === 'deployments',
  })

  const { data: services = [] } = useQuery<K8sResourceItem[]>({
    queryKey: ['devops-k8s-svc', namespace],
    queryFn: async () => (await call('DevOps.GetK8sServices', namespace) as K8sResourceItem[]) || [],
    enabled: subTab === 'services',
  })

  const { data: pods = [] } = useQuery<K8sResourceItem[]>({
    queryKey: ['devops-k8s-pods', namespace],
    queryFn: async () => (await call('DevOps.GetK8sPods', namespace) as K8sResourceItem[]) || [],
    enabled: subTab === 'pods',
  })

  const { data: events = [] } = useQuery<K8sEvent[]>({
    queryKey: ['devops-k8s-events', namespace],
    queryFn: async () => (await call('DevOps.GetK8sEvents', namespace, 30) as K8sEvent[]) || [],
    enabled: subTab === 'events',
  })

  const { data: rollouts = [] } = useQuery<K8sRolloutStatus[]>({
    queryKey: ['devops-k8s-rollouts', namespace],
    queryFn: async () => (await call('DevOps.GetK8sRollouts', namespace) as K8sRolloutStatus[]) || [],
    enabled: subTab === 'rollouts',
  })

  const k8sAction = async (action: string, ...args: string[]) => {
    try { await call(`DevOps.K8s${action}`, ...args) } catch (err) { console.error(err) }
  }

  return (
    <div className="flex flex-col h-full p-6 space-y-4">
      <div className="flex items-center gap-3">
        <select value={namespace} onChange={(e) => setNamespace(e.target.value)}
          className="bg-panel border border-border rounded-xl px-4 py-2 text-sm font-mono text-text focus:outline-none focus:border-accent">
          <option value="">All Namespaces</option>
          {namespaces.map(n => <option key={n.name} value={n.name}>{n.name} ({n.status})</option>)}
        </select>
      </div>
      <div className="flex gap-2 border-b border-border pb-2">
        {(['deployments', 'services', 'pods', 'rollouts', 'events'] as const).map(tab => (
          <button key={tab} onClick={() => setSubTab(tab)}
            className={cn('px-4 py-2 text-sm font-bold rounded-t-lg transition-colors',
              subTab === tab ? 'text-accent border-b-2 border-accent' : 'text-text-faint hover:text-text')}>
            {tab.charAt(0).toUpperCase() + tab.slice(1)}
          </button>
        ))}
      </div>

      <div className="flex-1 overflow-y-auto space-y-2">
        {subTab === 'deployments' && deployments.map(d => (
          <div key={d.name + d.namespace} className="bg-panel border border-border rounded-xl p-5 flex items-center justify-between">
            <div>
              <p className="font-bold text-sm tracking-tight">{d.name}</p>
              <p className="text-[10px] text-text-dim font-mono font-medium mt-1 uppercase tracking-tighter">{d.namespace} \u2022 {d.details}</p>
            </div>
            <div className="flex items-center gap-2">
              <StatusBadge status={d.status} />
              <ActionButton icon={<RefreshCw size={12} />} label="Restart" onClick={() => k8sAction('RestartDeployment', d.name, d.namespace)} />
              <ActionButton icon={<RotateCcw size={12} />} label="Rollback" variant="warning" onClick={() => k8sAction('RollbackDeployment', d.name, d.namespace, '0')} />
            </div>
          </div>
        ))}
        {subTab === 'services' && services.map(s => (
          <div key={s.name + s.namespace} className="bg-panel border border-border rounded-xl p-5 flex items-center justify-between">
            <div>
              <p className="font-bold text-sm tracking-tight">{s.name}</p>
              <p className="text-[10px] text-text-dim font-mono font-medium mt-1 uppercase tracking-tighter">{s.namespace} \u2022 {s.details}</p>
            </div>
            <StatusBadge status={s.status} />
          </div>
        ))}
        {subTab === 'pods' && pods.map(p => (
          <div key={p.name + p.namespace} className="bg-panel border border-border rounded-xl p-5 flex items-center justify-between">
            <div>
              <p className="font-bold text-sm">{p.name}</p>
              <p className="text-xs text-text-dim font-mono">{p.namespace}</p>
            </div>
            <StatusBadge status={p.status} />
          </div>
        ))}
        {subTab === 'rollouts' && rollouts.map(r => (
          <div key={r.name} className="bg-panel border border-border rounded-xl p-5 flex items-center justify-between">
            <div>
              <p className="font-bold text-sm">{r.name} ({r.kind})</p>
              <p className="text-xs text-text-dim">Replicas: {r.replicas} | Updated: {r.updated} | Available: {r.available}</p>
            </div>
            <span className={cn('text-xs font-bold', r.ready ? 'text-success' : 'text-warning')}>{r.ready ? 'Ready' : 'In Progress'}</span>
          </div>
        ))}
        {subTab === 'events' && events.map((e, i) => (
          <div key={i} className="bg-panel border border-border rounded-xl p-3 flex items-start gap-3">
            <span className={cn('w-2 h-2 rounded-full mt-1 shrink-0', e.type === 'Normal' ? 'bg-success' : 'bg-warning')} />
            <div className="min-w-0">
              <p className="text-[10px] font-black uppercase tracking-tight"><span className="text-text-faint">{e.last_seen}</span> {e.reason} \u2022 {e.object}</p>
              <p className="text-[11px] text-text-dim truncate font-medium">{e.message}</p>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

function ServicesTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const [search, setSearch] = useState('')
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [pendingAction, setPendingAction] = useState<{ name: string; action: string } | null>(null)
  const queryClient = useQueryClient()

  const { data: services = [] } = useQuery<ServiceEntry[]>({
    queryKey: ['devops-services'],
    queryFn: async () => (await call('DevOps.GetServices') as ServiceEntry[]) || [],
    refetchInterval: refreshInterval,
  })

  const controlService = async (name: string, action: string) => {
    const success = await call('DevOps.ControlService', name, action)
    if (success) queryClient.invalidateQueries({ queryKey: ['devops-services'] })
    setPendingAction(null)
  }

  const filtered = services.filter(s => s.name.toLowerCase().includes(search.toLowerCase()) || s.display_name.toLowerCase().includes(search.toLowerCase()))

  return (
    <div className="flex flex-col h-full p-6 space-y-4">
      <ConfirmDialog open={confirmOpen} title="Service Control" description={`${pendingAction?.action} "${pendingAction?.name}"?`}
        type={pendingAction?.action === 'stop' ? 'danger' : 'warning'} confirmText="Execute"
        onConfirm={() => pendingAction && controlService(pendingAction.name, pendingAction.action)} onClose={() => setConfirmOpen(false)} />
      <div className="flex items-center gap-4">
        <div className="bg-panel border border-border px-4 py-2 rounded-xl flex gap-6">
          <span className="text-sm font-bold text-success">{services.filter(s => s.status === 'Running').length} Running</span>
          <span className="text-sm font-bold text-danger">{services.filter(s => s.status === 'Stopped').length} Stopped</span>
        </div>
        <div className="flex-1 max-w-xs ml-auto">
          <SearchInput size="sm" value={search} onChange={(e) => setSearch(e.target.value)} placeholder="Search services..." />
        </div>
      </div>
      <div className="flex-1 bg-[var(--color-bg)] border border-border rounded-2xl overflow-hidden shadow-inner">
        <div className="overflow-y-auto h-full">
          <table className="w-full text-left border-collapse">
            <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border">
              <tr>
                <th className="px-4 py-3 text-xs font-semibold text-text-dim uppercase">Name</th>
                <th className="px-4 py-3 text-xs font-semibold text-text-dim uppercase">Status</th>
                <th className="px-4 py-3 text-xs font-semibold text-text-dim uppercase text-right">Control</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map(svc => (
                <tr key={svc.name} className="border-b border-border/20 hover:bg-panel-3 group transition-colors">
                  <td className="px-4 py-3"><p className="text-sm font-bold text-text">{svc.display_name}</p><p className="text-xs font-mono text-text-faint">{svc.name}</p></td>
                  <td className="px-4 py-3"><StatusBadge status={svc.status} /></td>
                  <td className="px-4 py-3 text-right opacity-0 group-hover:opacity-100 transition-opacity">
                    <div className="flex justify-end gap-1">
                      <button onClick={() => { setPendingAction({ name: svc.name, action: 'start' }); setConfirmOpen(true) }} className="p-1.5 text-success hover:bg-success/10 rounded-lg"><PlayCircle size={16} /></button>
                      <button onClick={() => { setPendingAction({ name: svc.name, action: 'stop' }); setConfirmOpen(true) }} className="p-1.5 text-danger hover:bg-danger/10 rounded-lg"><StopCircle size={16} /></button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}

function ServersTab() {
  const { call } = useBackend()
  const { data: servers = [] } = useQuery<LocalServer[]>({
    queryKey: ['devops-servers'],
    queryFn: async () => (await call('DevOps.GetLocalServers') as LocalServer[]) || []
  })
  return (
    <div className="p-6 space-y-3 overflow-y-auto h-full">
      {servers.map(s => (
        <div key={`${s.port}-${s.pid}`} className="p-5 rounded-xl border border-border bg-panel flex justify-between items-center">
          <div><p className="font-bold text-lg">:{s.port}</p><p className="text-xs font-mono text-text-faint">{s.process}</p></div>
          <span className="text-xs px-2 py-1 bg-accent/20 text-accent rounded uppercase font-bold">{s.framework}</span>
        </div>
      ))}
    </div>
  )
}

function EnvironmentTab() {
  const { call } = useBackend()
  const { data: env } = useQuery<EnvironmentInfo>({
    queryKey: ['devops-env'],
    queryFn: async () => (await call('DevOps.GetEnvironment') as EnvironmentInfo)
  })
  return (
    <div className="p-6 space-y-6 overflow-y-auto h-full">
      <section>
        <h3 className="font-bold mb-3 text-sm">Key Variables</h3>
        <div className="space-y-1">
          {env?.key_vars.map(v => (
            <div key={v.name} className="flex gap-4 text-xs font-mono">
              <span className="text-accent font-bold">{v.name}</span>
              <span className="text-text-dim truncate">{v.value}</span>
            </div>
          ))}
        </div>
      </section>
      <section>
        <h3 className="font-bold mb-3 text-sm">SDKs</h3>
        <div className="flex gap-2 flex-wrap">
          {env?.sdks.map(s => (
            <span key={s.name} className="px-3 py-1 bg-panel border border-border rounded-full text-xs font-mono">{s.name} {s.version}</span>
          ))}
        </div>
      </section>
      <section>
        <h3 className="font-bold mb-3 text-sm">Package Managers</h3>
        <div className="flex gap-2 flex-wrap">
          {env?.package_managers.map(pm => (
            <span key={pm.name} className="px-3 py-1 bg-panel border border-border rounded-full text-xs font-mono">{pm.name} {pm.version}</span>
          ))}
        </div>
      </section>
    </div>
  )
}
