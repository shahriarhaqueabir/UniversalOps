import { useState, useRef, useCallback, useEffect } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Terminal, Server, Folder, Play, Trash2, Search, ChevronRight,
  FileText, X, PlayCircle, StopCircle, Zap, Activity, Globe,
  TerminalSquare, GitBranch, Box, Wrench, Container, Variable,
  Rocket, GanttChart, GitMerge, GitPullRequest, RefreshCw,
  Shield, BarChart3, Layers, Ship, RotateCcw,
  Download, Upload, Eye,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import * as Tabs from '@radix-ui/react-tabs'
import { useBackend } from '@/hooks/useBackend'
import { useEvents } from '@/hooks/useEvents'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { ConfirmDialog } from '@/components/dialogs/ConfirmDialog'
import { nanoid } from 'nanoid'
import type {
  CommandResult, ServiceEntry, FileEntry, ToolInfo, ContainerSummary,
  GitSummary, LocalServer, EnvironmentInfo, DockerStatus, KubernetesStatus,
  ServiceGroupSummary, GitBranchInfo, GitTagInfo, GitStashEntry, GitRemoteInfo,
  DockerStatsEntry, DockerComposeProject, DockerNetworkInfo, DockerVolumeInfo,
  K8sResourceItem, K8sRolloutStatus, K8sEvent, K8sNamespaceInfo,
  BuildSystemInfo, BuildTargetInfo, CICDStatus, ReleaseHistory, DeploymentRecord,
  DORAMetrics, DevOpsDiagResult,
} from '@/types'

function stripAnsi(text: string): string {
  return text.replace(/\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])/g, '')
}

type TabId = 'overview' | 'terminal' | 'powershell-pro' | 'services' | 'file-browser' | 'toolbox' | 'docker' | 'git' | 'servers' | 'environment' | 'kubernetes' | 'pipelines' | 'releases' | 'diagnostics'

function StatusBadge({ status, variant }: { status: string; variant?: string }) {
  const colors: Record<string, string> = {
    running: 'bg-success/20 text-success',
    stopped: 'bg-danger/20 text-danger',
    auto: 'bg-accent/20 text-accent',
    manual: 'bg-warning/20 text-warning',
    disabled: 'bg-text-faint/20 text-text-faint',
  }
  const v = variant || status
  return (
    <span className={cn('px-2 py-0.5 rounded text-xs font-medium border border-current opacity-80', colors[v.toLowerCase()] || 'bg-text-faint/20 text-text-faint')}>
      {status}
    </span>
  )
}

function ActionButton({ icon, label, onClick, variant, disabled }: { icon: React.ReactNode; label: string; onClick: () => void; variant?: string; disabled?: boolean }) {
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
}

export function DevOps() {
  const [activeTab, setActiveTab] = useState<TabId>('overview')

  return (
    <div className="flex flex-col h-full bg-[var(--color-bg)]">
      <div className="p-8 border-b border-border bg-panel-2">
        <h1 className="text-3xl font-bold text-text flex items-center gap-3">
          <TerminalSquare size={32} className="text-accent" />
          DevOps Console
        </h1>
        <p className="text-text-dim text-lg mt-2">
          Build, deploy, observe, and operate — unified control center.
        </p>
      </div>

      <Tabs.Root defaultValue="overview" onValueChange={(v) => setActiveTab(v as TabId)} className="flex-1 flex flex-col min-w-0">
        <Tabs.List className="flex border-b border-border bg-panel px-4 overflow-x-auto no-scrollbar">
          {[
            { id: 'overview', label: 'Overview', icon: <Activity size={20} className="text-accent" /> },
            { id: 'terminal', label: 'Terminal', icon: <Terminal size={20} /> },
            { id: 'powershell-pro', label: 'PowerShell', icon: <Zap size={20} className="text-warning" /> },
            { id: 'git', label: 'Git', icon: <GitBranch size={20} /> },
            { id: 'docker', label: 'Docker', icon: <Container size={20} /> },
            { id: 'kubernetes', label: 'K8s', icon: <Layers size={20} /> },
            { id: 'pipelines', label: 'Pipelines', icon: <GanttChart size={20} /> },
            { id: 'releases', label: 'Releases', icon: <Rocket size={20} /> },
            { id: 'diagnostics', label: 'Health', icon: <Shield size={20} /> },
            { id: 'services', label: 'Services', icon: <Server size={20} /> },
            { id: 'servers', label: 'Servers', icon: <Globe size={20} /> },
            { id: 'environment', label: 'Env', icon: <Variable size={20} /> },
            { id: 'file-browser', label: 'Files', icon: <Folder size={20} /> },
            { id: 'toolbox', label: 'Toolbox', icon: <Wrench size={20} /> },
          ].map((tab) => (
            <Tabs.Trigger key={tab.id} value={tab.id}
              data-automation-id={`devops-tab-${tab.id}`}
              className={cn(
                'flex items-center gap-2 px-4 py-3 text-sm font-bold transition-all border-b-2 border-transparent whitespace-nowrap',
                activeTab === tab.id ? 'border-accent text-text bg-accent/5' : 'text-text-faint hover:text-text hover:bg-[var(--color-sidebar-hover)]',
              )}>
              {tab.icon}{tab.label}
            </Tabs.Trigger>
          ))}
        </Tabs.List>

        <div className="flex-1 overflow-hidden">
          <Tabs.Content value="overview" className="h-full"><OverviewTab /></Tabs.Content>
          <Tabs.Content value="terminal" className="h-full"><TerminalTab /></Tabs.Content>
          <Tabs.Content value="powershell-pro" className="h-full"><PowerShellProTab /></Tabs.Content>
          <Tabs.Content value="git" className="h-full"><GitTabExpanded /></Tabs.Content>
          <Tabs.Content value="docker" className="h-full"><DockerTabExpanded /></Tabs.Content>
          <Tabs.Content value="kubernetes" className="h-full"><KubernetesTab /></Tabs.Content>
          <Tabs.Content value="pipelines" className="h-full"><PipelinesTab /></Tabs.Content>
          <Tabs.Content value="releases" className="h-full"><ReleasesTab /></Tabs.Content>
          <Tabs.Content value="diagnostics" className="h-full"><DiagnosticsTab /></Tabs.Content>
          <Tabs.Content value="services" className="h-full"><ServicesTab /></Tabs.Content>
          <Tabs.Content value="file-browser" className="h-full"><FileBrowserTab /></Tabs.Content>
          <Tabs.Content value="toolbox" className="h-full"><ToolboxTab /></Tabs.Content>
          <Tabs.Content value="servers" className="h-full"><ServersTab /></Tabs.Content>
          <Tabs.Content value="environment" className="h-full"><EnvironmentTab /></Tabs.Content>
        </div>
      </Tabs.Root>
    </div>
  )
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

  const { data: dora } = useQuery<DORAMetrics>({
    queryKey: ['devops-dora'],
    queryFn: async () => {
      try { return await call('DevOps.GetDORAMetrics', '.') as DORAMetrics }
      catch { return null as unknown as DORAMetrics }
    },
  })

  return (
    <div className="flex flex-col h-full space-y-6 overflow-y-auto p-8">
      <div className="grid grid-cols-2 gap-6">
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-2xl">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-3">
              <Box size={24} className="text-accent" />
              <h3 className="text-lg font-bold text-text uppercase tracking-widest">Docker</h3>
            </div>
            <span className={cn('text-xs font-bold px-3 py-1 rounded-full uppercase tracking-widest border',
              dockerData?.installed ? (dockerData?.running ? 'bg-success/20 text-success border-success/30' : 'bg-warning/20 text-warning border-warning/30') : 'bg-text-faint/20 text-text-faint border-text-faint/30')}>
              {dockerData?.installed ? (dockerData?.running ? 'Running' : 'Installed') : 'Not Found'}
            </span>
          </div>
          {dockerData?.installed ? (
            <div className="grid grid-cols-4 gap-3">
              {[{ label: 'Running', value: dockerData.containers?.running ?? 0, cl: 'text-success' },
                { label: 'Stopped', value: dockerData.containers?.stopped ?? 0, cl: 'text-warning' },
                { label: 'Failed', value: dockerData.containers?.failed ?? 0, cl: 'text-danger' },
                { label: 'Total', value: dockerData.containers?.total ?? 0, cl: 'text-text-dim' }].map(item => (
                  <div key={item.label} className="flex flex-col items-center bg-panel-2 border border-border rounded-xl p-3">
                    <span className={cn('text-2xl font-bold tabular-nums', item.cl)}>{item.value}</span>
                    <span className="text-[10px] font-semibold text-text-faint uppercase tracking-wider">{item.label}</span>
                  </div>
              ))}
            </div>
          ) : <p className="text-text-faint text-sm">Docker not detected.</p>}
        </div>

        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-2xl">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-3">
              <Server size={24} className="text-info" />
              <h3 className="text-lg font-bold text-text uppercase tracking-widest">Kubernetes</h3>
            </div>
            <span className={cn('text-xs font-bold px-3 py-1 rounded-full uppercase tracking-widest border',
              k8sData?.installed ? (k8sData?.connected ? 'bg-success/20 text-success border-success/30' : 'bg-warning/20 text-warning border-warning/30') : 'bg-text-faint/20 text-text-faint border-text-faint/30')}>
              {k8sData?.installed ? (k8sData?.connected ? 'Connected' : 'Offline') : 'Not Found'}
            </span>
          </div>
          {k8sData?.installed ? (
            <div className="grid grid-cols-2 gap-3">
              <div className="flex flex-col items-center bg-panel-2 border border-border rounded-xl p-3">
                <span className="text-2xl font-bold tabular-nums text-info">{k8sData.nodes}</span>
                <span className="text-xs font-semibold text-text-faint uppercase">Nodes</span>
              </div>
              <div className="flex flex-col items-center bg-panel-2 border border-border rounded-xl p-3">
                <span className="text-2xl font-bold tabular-nums text-accent">{k8sData.pods}</span>
                <span className="text-xs font-semibold text-text-faint uppercase">Pods</span>
              </div>
            </div>
          ) : <p className="text-text-faint text-sm">kubectl not detected.</p>}
        </div>
      </div>

      {dora && (
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-2xl">
          <div className="flex items-center gap-3 mb-4">
            <BarChart3 size={24} className="text-accent" />
            <h3 className="text-lg font-bold text-text uppercase tracking-widest">DORA Metrics</h3>
            <span className="text-xs text-text-faint ml-auto">{dora.period}</span>
          </div>
          <div className="grid grid-cols-4 gap-4">
            {[
              { label: 'Deploy Frequency', value: dora.deployment_frequency, cl: 'text-accent' },
              { label: 'Lead Time', value: dora.lead_time_for_changes, cl: 'text-info' },
              { label: 'Change Failure Rate', value: dora.change_failure_rate, cl: dora.failure_pct > 15 ? 'text-danger' : 'text-success' },
              { label: 'MTTR', value: dora.mttr, cl: 'text-warning' },
            ].map(item => (
              <div key={item.label} className="flex flex-col items-center bg-panel-2 border border-border rounded-xl p-3">
                <span className={cn('text-lg font-bold tabular-nums', item.cl)}>{item.value}</span>
                <span className="text-[10px] font-semibold text-text-faint uppercase tracking-wider">{item.label}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-4 shadow-2xl">
        <div className="grid grid-cols-5 gap-4">
          {[
            { label: 'Databases', value: summary?.databases ?? 0, cl: 'text-accent' },
            { label: 'Queues', value: summary?.messageQueues ?? 0, cl: 'text-accent-2' },
            { label: 'Web Servers', value: summary?.webServers ?? 0, cl: 'text-success' },
            { label: 'Running', value: summary?.running ?? 0, cl: 'text-success' },
            { label: 'Stopped', value: summary?.stopped ?? 0, cl: 'text-danger' },
          ].map(kpi => (
            <div key={kpi.label} className="flex flex-col items-center gap-1 bg-panel-2 border border-border rounded-xl p-4">
              <span className={cn('text-2xl font-bold tabular-nums', kpi.cl)}>{kpi.value}</span>
              <span className="text-[10px] font-semibold text-text-faint uppercase tracking-wider">{kpi.label}</span>
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
  const [output, setOutput] = useState<string[]>([`Unified Terminal — Type a command...\n`])
  const [history, setHistory] = useState<string[]>([])
  const [historyIndex, setHistoryIndex] = useState(-1)
  const [isRunning, setIsRunning] = useState(false)
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [pendingCmd, setPendingCmd] = useState('')
  const [currentCmdId, setCurrentCmdId] = useState<string | null>(null)
  const outputRef = useRef<HTMLDivElement>(null)

  useEvents('DevOps.EventCmdLine', (data: any) => {
    if (data.id === currentCmdId) setOutput(prev => [...prev, data.line])
  })

  useEvents('DevOps.EventCmdDone', (id) => {
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
    <div className="flex flex-col h-full p-6 space-y-4">
      <ConfirmDialog open={confirmOpen} title="Impactful Command Detected"
        description={`Run: "${pendingCmd}"?`} type="danger" confirmText="Execute"
        onConfirm={() => { runCommand(pendingCmd); setInput(''); setPendingCmd('') }}
        onClose={() => setConfirmOpen(false)} />
      <div className="flex items-center gap-3">
        <div className="relative flex-1">
          <input type="text" value={input} onChange={(e) => setInput(e.target.value)} onKeyDown={handleKeyDown}
            placeholder="Enter shell command..." disabled={isRunning}
            className="w-full bg-[var(--color-bg)] border border-border rounded-xl pl-10 pr-4 py-3 text-sm font-mono text-text focus:outline-none focus:border-accent shadow-inner" />
          <span className="absolute left-3 top-1/2 -translate-y-1/2 text-success text-sm font-bold font-mono">$</span>
        </div>
        <button onClick={() => { if (input.trim()) runCommand(input) }} disabled={isRunning || !input.trim()}
          className="flex items-center gap-2 px-6 py-3 text-sm font-bold bg-accent text-white rounded-xl hover:bg-accent/90 disabled:opacity-50 transition-all shadow-lg active:scale-95">
          <Play size={16} /> Run
        </button>
        <button onClick={() => setOutput([`Output cleared.\n`])} className="px-4 py-3 text-sm font-bold text-text-dim border border-border rounded-xl hover:bg-panel-3 transition-all">
          <Trash2 size={16} />
        </button>
      </div>
      <div ref={outputRef} className="flex-1 bg-black border border-border rounded-2xl p-6 overflow-y-auto font-mono text-sm leading-relaxed whitespace-pre-wrap shadow-inner">
        {output.map((block, i) => <div key={i} className="mb-0.5">{stripAnsi(block)}</div>)}
        {isRunning && <div className="flex items-center gap-2 mt-2"><span className="inline-block w-2 h-4 bg-success animate-pulse" /><span className="text-xs font-bold text-success uppercase animate-pulse">Running...</span></div>}
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
    <div className="grid grid-cols-3 h-full gap-6 overflow-hidden p-6">
      <ConfirmDialog open={confirmOpen} title="Execute PowerShell Workflow" description={`Run "${selectedWorkflow}"?`}
        onConfirm={() => runWorkflow(selectedWorkflow)} onClose={() => setConfirmOpen(false)} />
      <div className="col-span-1 space-y-3 overflow-y-auto pr-2">
        <h3 className="text-sm font-bold text-text-dim uppercase tracking-widest mb-3">Diagnostic Workflows</h3>
        {workflows.map(wf => (
          <button key={wf} onClick={() => { setSelectedWorkflow(wf); setConfirmOpen(true) }} disabled={isRunning}
            className="w-full text-left bg-panel border border-border rounded-xl p-4 transition-all hover:border-accent/50 hover:bg-accent/5 group disabled:opacity-50">
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
        <div className="flex-1 bg-black border border-border rounded-2xl p-6 overflow-y-auto font-mono text-sm leading-relaxed whitespace-pre shadow-inner">
          {stripAnsi(output) || 'Select a workflow to begin.'}
        </div>
        <button onClick={() => setOutput('')} className="self-end px-4 py-2 text-sm font-bold text-text-dim border border-border rounded-xl hover:bg-panel-3 transition-all">Clear</button>
      </div>
    </div>
  )
}

function GitTabExpanded() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const [selectedRepo, setSelectedRepo] = useState('')
  const [branchName, setBranchName] = useState('')
  const [commitMsg, setCommitMsg] = useState('')
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [pendingAction, setPendingAction] = useState<{ label: string; action: () => void } | null>(null)
  const queryClient = useQueryClient()
  const [logOutput, setLogOutput] = useState('')

  const { data: gitSummary } = useQuery<GitSummary>({
    queryKey: ['devops-git'],
    queryFn: async () => (await call('DevOps.GetGitSummary') as GitSummary) || { repositories: [] },
    refetchInterval: refreshInterval,
  })

  const repos = gitSummary?.repositories || []
  const activeRepo = selectedRepo || repos[0]?.path || ''

  const { data: branches = [] } = useQuery<GitBranchInfo[]>({
    queryKey: ['devops-git-branches', activeRepo],
    queryFn: async () => (await call('DevOps.GetGitBranches', activeRepo) as GitBranchInfo[]) || [],
    enabled: !!activeRepo,
  })

  const { data: tags = [] } = useQuery<GitTagInfo[]>({
    queryKey: ['devops-git-tags', activeRepo],
    queryFn: async () => (await call('DevOps.GetGitTags', activeRepo) as GitTagInfo[]) || [],
    enabled: !!activeRepo,
  })

  const { data: stash = [] } = useQuery<GitStashEntry[]>({
    queryKey: ['devops-git-stash', activeRepo],
    queryFn: async () => (await call('DevOps.GetGitStash', activeRepo) as GitStashEntry[]) || [],
    enabled: !!activeRepo,
  })

  const { data: remotes = [] } = useQuery<GitRemoteInfo[]>({
    queryKey: ['devops-git-remotes', activeRepo],
    queryFn: async () => (await call('DevOps.GetGitRemotes', activeRepo) as GitRemoteInfo[]) || [],
    enabled: !!activeRepo,
  })

  const runGitAction = async (action: string, ...args: string[]) => {
    try {
      const res = await call(`DevOps.Git${action}`, activeRepo, ...args) as CommandResult
      if (res.error) console.error(res.error)
      queryClient.invalidateQueries({ queryKey: ['devops-git'] })
      return res
    } catch (err) { console.error(err) }
  }

  const fetchLog = async (branch = '') => {
    const log = await call('DevOps.GitLogExtended', activeRepo, 15, branch) as string
    setLogOutput(log || 'No commits')
  }

  useEffect(() => { if (activeRepo) fetchLog() }, [activeRepo])

  return (
    <div className="flex flex-col h-full p-6 space-y-4 overflow-y-auto">
      <ConfirmDialog open={confirmOpen} title="Confirm Git Action"
        description={pendingAction?.label || ''}
        onConfirm={() => { pendingAction?.action(); setConfirmOpen(false) }}
        onClose={() => setConfirmOpen(false)} />

      <div className="flex items-center gap-3 flex-wrap">
        <select value={activeRepo} onChange={(e) => setSelectedRepo(e.target.value)}
          className="bg-panel border border-border rounded-xl px-4 py-2 text-sm font-mono text-text focus:outline-none focus:border-accent">
          {repos.map(r => <option key={r.path} value={r.path}>{r.path.split(/[\\/]/).pop()} ({r.branch})</option>)}
        </select>
        <span className="text-xs font-mono text-text-faint">{activeRepo}</span>
        <div className="ml-auto flex gap-2">
          <ActionButton icon={<RefreshCw size={14} />} label="Fetch" variant="primary" onClick={() => runGitAction('Fetch', '', '')} />
          <ActionButton icon={<Download size={14} />} label="Pull" onClick={() => runGitAction('Pull', '', '')} />
          <ActionButton icon={<Upload size={14} />} label="Push" onClick={() => runGitAction('Push', '', '')} />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="bg-panel border border-border rounded-xl p-4">
          <h4 className="text-xs font-bold text-text-dim uppercase tracking-widest mb-3">Branches</h4>
          <div className="max-h-40 overflow-y-auto space-y-1">
            {branches.map(b => (
              <div key={b.name} className={cn('flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm', b.current ? 'bg-accent/10 text-accent' : 'hover:bg-panel-3')}>
                <GitBranch size={12} className={b.current ? 'text-accent' : 'text-text-faint'} />
                <span className={cn('font-mono', b.current && 'font-bold')}>{b.name} {b.current && '*'}</span>
                <span className="text-xs text-text-faint ml-auto">{b.last_commit}</span>
              </div>
            ))}
          </div>
          <div className="flex gap-2 mt-3">
            <input value={branchName} onChange={(e) => setBranchName(e.target.value)} placeholder="New branch..."
              className="flex-1 bg-[var(--color-bg)] border border-border rounded-lg px-3 py-1.5 text-xs font-mono" />
            <button onClick={() => runGitAction('CreateBranch', branchName)} disabled={!branchName}
              className="px-3 py-1.5 text-xs font-bold bg-accent text-white rounded-lg hover:bg-accent/90 disabled:opacity-50">Create</button>
          </div>
        </div>

        <div className="bg-panel border border-border rounded-xl p-4">
          <h4 className="text-xs font-bold text-text-dim uppercase tracking-widest mb-3">Tags & Stash</h4>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <p className="text-xs font-semibold text-text-faint mb-1">Tags ({tags.length})</p>
              <div className="max-h-20 overflow-y-auto space-y-1">
                {tags.map(t => <div key={t.name} className="flex items-center gap-2 text-xs font-mono"><span className="text-accent">●</span>{t.name}</div>)}
              </div>
            </div>
            <div>
              <p className="text-xs font-semibold text-text-faint mb-1">Stash ({stash.length})</p>
              <div className="max-h-20 overflow-y-auto space-y-1">
                {stash.map(s => <div key={s.index} className="flex items-center gap-2 text-xs font-mono"><span className="text-warning">●</span>{s.message}</div>)}
              </div>
            </div>
          </div>
          <div className="flex gap-2 mt-3">
            <input value={commitMsg} onChange={(e) => setCommitMsg(e.target.value)} placeholder="Commit message..."
              className="flex-1 bg-[var(--color-bg)] border border-border rounded-lg px-3 py-1.5 text-xs font-mono" />
            <button onClick={() => { if (commitMsg) { setPendingAction({ label: `Commit: "${commitMsg}"`, action: () => runGitAction('Commit', commitMsg) }); setConfirmOpen(true) } }}
              disabled={!commitMsg} className="px-3 py-1.5 text-xs font-bold bg-accent text-white rounded-lg hover:bg-accent/90 disabled:opacity-50">Commit</button>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="bg-panel border border-border rounded-xl p-4">
          <h4 className="text-xs font-bold text-text-dim uppercase tracking-widest mb-3">Remotes</h4>
          <div className="space-y-1">
            {remotes.map(r => (
              <div key={r.name} className="flex items-center gap-2 text-sm font-mono">
                <Globe size={12} className="text-text-faint" />
                <span className="font-bold text-text-dim">{r.name}:</span>
                <span className="text-text-faint truncate">{r.url}</span>
              </div>
            ))}
            {remotes.length === 0 && <p className="text-xs text-text-faint">No remotes configured</p>}
          </div>
        </div>

        <div className="bg-panel border border-border rounded-xl p-4">
          <h4 className="text-xs font-bold text-text-dim uppercase tracking-widest mb-3">Actions</h4>
          <div className="flex flex-wrap gap-2">
            <ActionButton icon={<GitMerge size={14} />} label="Merge" variant="warning"
              onClick={() => { const br = prompt('Branch to merge:'); if (br) runGitAction('Merge', br) }} />
            <ActionButton icon={<RotateCcw size={14} />} label="Rebase" variant="warning"
              onClick={() => { const br = prompt('Branch to rebase onto:'); if (br) runGitAction('Rebase', br) }} />
            <ActionButton icon={<Trash2 size={14} />} label="Clean" variant="danger"
              onClick={() => { setPendingAction({ label: 'Clean untracked files?', action: () => runGitAction('Clean') }); setConfirmOpen(true) }} />
            <ActionButton icon={<GitPullRequest size={14} />} label="Add All" onClick={() => runGitAction('Add', '.')} />
            <ActionButton icon={<GitBranch size={14} />} label="Status"
              onClick={async () => { const r = await runGitAction('Status'); if (r) setLogOutput(r.output || r.error || '') }} />
          </div>
        </div>
      </div>

      <div className="bg-black border border-border rounded-2xl p-4 overflow-y-auto max-h-48">
        <pre className="text-xs font-mono text-green-400 leading-relaxed whitespace-pre-wrap">{logOutput}</pre>
      </div>
    </div>
  )
}

function DockerTabExpanded() {
  const { call } = useBackend()
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
    refetchInterval: 5000,
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
        onConfirm={() => { pendingAction?.action(); setConfirmOpen(false) }}
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
            <div key={c.id} className="bg-panel border border-border rounded-xl p-4 flex items-center justify-between">
              <div className="flex items-center gap-4">
                <span className={cn('w-2 h-2 rounded-full', c.state === 'running' ? 'bg-success' : 'bg-text-faint')} />
                <div>
                  <p className="font-bold text-sm">{c.name} <span className="text-xs font-mono text-text-faint ml-2">{c.id.slice(0, 12)}</span></p>
                  <p className="text-xs text-text-dim font-mono">{c.image} — {c.status}</p>
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
            <div key={p.project} className="bg-panel border border-border rounded-xl p-4">
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
            <div key={n.id} className="bg-panel border border-border rounded-xl p-4 flex items-center justify-between">
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
            <div key={v.name} className="bg-panel border border-border rounded-xl p-4 flex items-center justify-between">
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
          <div key={d.name + d.namespace} className="bg-panel border border-border rounded-xl p-4 flex items-center justify-between">
            <div>
              <p className="font-bold text-sm">{d.name}</p>
              <p className="text-xs text-text-dim font-mono">{d.namespace} — {d.details}</p>
            </div>
            <div className="flex items-center gap-2">
              <StatusBadge status={d.status} />
              <ActionButton icon={<RefreshCw size={12} />} label="Restart" onClick={() => k8sAction('RestartDeployment', d.name, d.namespace)} />
              <ActionButton icon={<RotateCcw size={12} />} label="Rollback" variant="warning" onClick={() => k8sAction('RollbackDeployment', d.name, d.namespace, '0')} />
            </div>
          </div>
        ))}
        {subTab === 'services' && services.map(s => (
          <div key={s.name + s.namespace} className="bg-panel border border-border rounded-xl p-4 flex items-center justify-between">
            <div>
              <p className="font-bold text-sm">{s.name}</p>
              <p className="text-xs text-text-dim font-mono">{s.namespace} — {s.details}</p>
            </div>
            <StatusBadge status={s.status} />
          </div>
        ))}
        {subTab === 'pods' && pods.map(p => (
          <div key={p.name + p.namespace} className="bg-panel border border-border rounded-xl p-4 flex items-center justify-between">
            <div>
              <p className="font-bold text-sm">{p.name}</p>
              <p className="text-xs text-text-dim font-mono">{p.namespace}</p>
            </div>
            <StatusBadge status={p.status} />
          </div>
        ))}
        {subTab === 'rollouts' && rollouts.map(r => (
          <div key={r.name} className="bg-panel border border-border rounded-xl p-4 flex items-center justify-between">
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
              <p className="text-xs font-bold"><span className="text-text-faint">{e.last_seen}</span> {e.reason} — {e.object}</p>
              <p className="text-xs text-text-dim truncate">{e.message}</p>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

function PipelinesTab() {
  const { call } = useBackend()
  const [activeSubTab, setActiveSubTab] = useState<'cicd' | 'builds'>('cicd')

  const { data: cicd } = useQuery<CICDStatus>({
    queryKey: ['devops-cicd'],
    queryFn: async () => (await call('DevOps.GetCICDStatus', '.') as CICDStatus) || {},
  })

  const { data: buildSystems = [] } = useQuery<BuildSystemInfo[]>({
    queryKey: ['devops-build-systems'],
    queryFn: async () => (await call('DevOps.GetBuildSystems') as BuildSystemInfo[]) || [],
  })

  const { data: buildTargets = [] } = useQuery<BuildTargetInfo[]>({
    queryKey: ['devops-build-targets'],
    queryFn: async () => (await call('DevOps.FindBuildTargets', '.') as BuildTargetInfo[]) || [],
    enabled: activeSubTab === 'builds',
  })

  const runBuild = async (target: BuildTargetInfo, action: string) => {
    const r = await call('DevOps.RunBuildCommand', target.type, target.path, action) as CommandResult
    console.log(r.output)
  }

  return (
    <div className="flex flex-col h-full p-6 space-y-4 overflow-y-auto">
      <div className="flex gap-2 border-b border-border pb-2">
        {(['cicd', 'builds'] as const).map(tab => (
          <button key={tab} onClick={() => setActiveSubTab(tab)}
            className={cn('px-4 py-2 text-sm font-bold rounded-t-lg transition-colors',
              activeSubTab === tab ? 'text-accent border-b-2 border-accent' : 'text-text-faint hover:text-text')}>
            {tab === 'cicd' ? 'CI/CD' : 'Build Systems'}
          </button>
        ))}
      </div>

      {activeSubTab === 'cicd' && (
        <div className="space-y-3">
          <div className="grid grid-cols-5 gap-3 mb-4">
            {(cicd?.configs ?? []).map(c => (
              <div key={c.platform} className={cn('bg-panel border rounded-xl p-4 text-center', c.detected ? 'border-success/30' : 'border-border')}>
                <span className={cn('text-xs font-bold block mb-1', c.detected ? 'text-success' : 'text-text-faint')}>
                  {c.platform}
                </span>
                <span className={cn('text-2xl', c.detected ? 'text-success' : 'text-text-faint')}>
                  {c.detected ? '✓' : '○'}
                </span>
              </div>
            ))}
          </div>
          <h4 className="text-sm font-bold text-text-dim uppercase tracking-widest">Detected Pipelines</h4>
          {(cicd?.pipelines ?? []).map(p => (
            <div key={p.name} className="bg-panel border border-border rounded-xl p-4 flex items-center justify-between">
              <div>
                <p className="font-bold text-sm">{p.name}</p>
                <p className="text-xs text-text-dim">{p.branch && `Branch: ${p.branch}`}{p.duration && ` | ${p.duration}`}</p>
              </div>
              <StatusBadge status={p.status} />
            </div>
          ))}
          {(!cicd?.pipelines || cicd.pipelines.length === 0) && <p className="text-text-faint text-sm">No CI/CD pipelines detected. Configure one to get started.</p>}
        </div>
      )}

      {activeSubTab === 'builds' && (
        <div className="grid grid-cols-2 gap-6">
          <div>
            <h4 className="text-sm font-bold text-text-dim uppercase tracking-widest mb-3">Installed Build Systems</h4>
            <div className="space-y-2">
              {buildSystems.map(bs => (
                <div key={bs.name} className="bg-panel border border-border rounded-xl p-4 flex items-center justify-between">
                  <div>
                    <p className="font-bold text-sm">{bs.name}</p>
                    <p className="text-xs font-mono text-text-dim">{bs.version || 'not found'}</p>
                  </div>
                  <span className={cn('text-xs font-bold', bs.found ? 'text-success' : 'text-text-faint')}>
                    {bs.found ? 'Installed' : 'Missing'}
                  </span>
                </div>
              ))}
            </div>
          </div>
          <div>
            <h4 className="text-sm font-bold text-text-dim uppercase tracking-widest mb-3">Build Targets</h4>
            <div className="space-y-2">
              {buildTargets.map(t => (
                <div key={t.path} className="bg-panel border border-border rounded-xl p-4">
                  <div className="flex items-center justify-between mb-2">
                    <p className="font-bold text-sm">{t.name}</p>
                    <StatusBadge status={t.type} />
                  </div>
                  <div className="flex gap-3 text-xs text-text-dim mb-2">
                    {t.has_build && <span className="text-success">✓ Build</span>}
                    {t.has_test && <span className="text-success">✓ Test</span>}
                    {t.has_lint && <span className="text-success">✓ Lint</span>}
                    <span className="text-text-faint">{t.dep_count} deps</span>
                  </div>
                  <div className="flex gap-2">
                    <ActionButton icon={<Play size={12} />} label="Build" variant="primary" onClick={() => runBuild(t, 'build')} />
                    <ActionButton icon={<Eye size={12} />} label="Test" onClick={() => runBuild(t, 'test')} />
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function ReleasesTab() {
  const { call } = useBackend()
  const [subTab, setSubTab] = useState<'releases' | 'deployments'>('releases')

  const { data: releases } = useQuery<ReleaseHistory>({
    queryKey: ['devops-releases'],
    queryFn: async () => (await call('DevOps.GetReleases', '.') as ReleaseHistory) || { releases: [] },
  })

  const { data: deployments = [] } = useQuery<DeploymentRecord[]>({
    queryKey: ['devops-deployments'],
    queryFn: async () => (await call('DevOps.GetDeploymentHistory', '.') as DeploymentRecord[]) || [],
    enabled: subTab === 'deployments',
  })

  const { data: dora } = useQuery<DORAMetrics>({
    queryKey: ['devops-dora-metrics'],
    queryFn: async () => (await call('DevOps.GetDORAMetrics', '.') as DORAMetrics) || null as unknown as DORAMetrics,
  })

  return (
    <div className="flex flex-col h-full p-6 space-y-4 overflow-y-auto">
      {dora && (
        <div className="grid grid-cols-4 gap-3">
          {[
            { label: 'Deploy Frequency', value: dora.deployment_frequency, cl: 'text-accent' },
            { label: 'Lead Time', value: dora.lead_time_for_changes, cl: 'text-info' },
            { label: 'Change Failure Rate', value: dora.change_failure_rate, cl: dora.failure_pct > 15 ? 'text-danger' : 'text-success' },
            { label: 'MTTR', value: dora.mttr, cl: 'text-warning' },
          ].map(item => (
            <div key={item.label} className="bg-panel border border-border rounded-xl p-4 text-center">
              <span className={cn('text-xl font-bold tabular-nums', item.cl)}>{item.value}</span>
              <p className="text-[10px] font-semibold text-text-faint uppercase tracking-wider mt-1">{item.label}</p>
            </div>
          ))}
        </div>
      )}

      <div className="flex gap-2 border-b border-border pb-2">
        {(['releases', 'deployments'] as const).map(tab => (
          <button key={tab} onClick={() => setSubTab(tab)}
            className={cn('px-4 py-2 text-sm font-bold rounded-t-lg transition-colors',
              subTab === tab ? 'text-accent border-b-2 border-accent' : 'text-text-faint hover:text-text')}>
            {tab.charAt(0).toUpperCase() + tab.slice(1)}
          </button>
        ))}
      </div>

      {subTab === 'releases' && (releases?.releases ?? []).map(r => (
        <div key={r.version} className="bg-panel border border-border rounded-xl p-4 flex items-center justify-between">
          <div>
            <div className="flex items-center gap-3">
              <Ship size={16} className="text-accent" />
              <p className="font-bold text-sm">{r.version}</p>
              <StatusBadge status={r.status} />
            </div>
            <p className="text-xs text-text-dim mt-1">{r.notes || r.date}</p>
          </div>
          <span className="text-xs font-mono text-text-faint">{r.date ? new Date(r.date).toLocaleDateString() : ''}</span>
        </div>
      ))}
      {subTab === 'deployments' && deployments.map(d => (
        <div key={d.id} className="bg-panel border border-border rounded-xl p-4 flex items-center justify-between">
          <div>
            <div className="flex items-center gap-2">
              <p className="font-bold text-sm">{d.version}</p>
              <span className="text-xs font-mono text-text-faint">→ {d.environment}</span>
            </div>
            <p className="text-xs text-text-dim">{d.trigger} · {d.timestamp}</p>
          </div>
          <StatusBadge status={d.status} />
        </div>
      ))}
      {subTab === 'releases' && (!releases?.releases || releases.releases.length === 0) && <p className="text-text-faint text-sm text-center py-8">No releases found. Create a git tag to track releases.</p>}
    </div>
  )
}

function DiagnosticsTab() {
  const { call } = useBackend()
  const [result, setResult] = useState<DevOpsDiagResult | null>(null)
  const [running, setRunning] = useState(false)

  const runDiagnostics = async () => {
    setRunning(true)
    try {
      const res = await call('DevOps.RunDevOpsDiagnostics') as DevOpsDiagResult
      setResult(res)
    } catch (err) { console.error(err) }
    finally { setRunning(false) }
  }

  return (
    <div className="flex flex-col h-full p-6 space-y-6">
      <div className="flex items-center gap-4">
        <button onClick={runDiagnostics} disabled={running}
          className="flex items-center gap-3 px-8 py-4 text-lg font-bold bg-accent text-white rounded-xl hover:bg-accent/90 disabled:opacity-50 transition-all shadow-lg active:scale-95">
          <Activity size={24} className={cn(running && 'animate-pulse')} />
          {running ? 'Running Checks...' : 'Run DevOps Health Check'}
        </button>
      </div>

      {result && (
        <>
          <div className="flex items-center gap-4">
            <span className={cn('text-4xl font-bold', result.score >= 80 ? 'text-success' : result.score >= 50 ? 'text-warning' : 'text-danger')}>
              {result.score}%
            </span>
            <span className="text-sm text-text-dim">Health Score</span>
          </div>

          <div className="flex-1 overflow-y-auto space-y-2">
            {result.checks.map((check, i) => (
              <div key={i} className={cn('bg-panel border rounded-xl p-4 flex items-center justify-between',
                check.status === 'pass' ? 'border-success/20' : check.status === 'warn' ? 'border-warning/20' : 'border-danger/20')}>
                <div>
                  <p className="font-bold text-sm">{check.name}</p>
                  <p className="text-xs text-text-dim">{check.message}</p>
                </div>
                <div className="flex items-center gap-3">
                  <span className="text-xs font-mono text-text-faint">{check.value}</span>
                  <span className={cn('text-lg font-bold',
                    check.status === 'pass' ? 'text-success' : check.status === 'warn' ? 'text-warning' : 'text-danger')}>
                    {check.status === 'pass' ? '✓' : check.status === 'warn' ? '!' : '✗'}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </>
      )}
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
        <div className="relative flex-1 max-w-xs ml-auto">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-text-faint" />
          <input value={search} onChange={(e) => setSearch(e.target.value)} placeholder="Search services..."
            className="w-full bg-panel border border-border rounded-xl pl-10 pr-4 py-2 text-sm focus:outline-none focus:border-accent" />
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

function FileBrowserTab() {
  const { call } = useBackend()
  const [currentPath, setCurrentPath] = useState('')
  const [entries, setEntries] = useState<FileEntry[]>([])
  const [history, setHistory] = useState<string[]>([])
  const [, setLoading] = useState(false)
  const [previewFile, setPreviewFile] = useState<FileEntry | null>(null)
  const [fileContent, setPreviewContent] = useState('')

  const fetchDir = useCallback(async (path: string) => {
    setLoading(true)
    try {
      const res = await call('DevOps.ListDirectory', path)
      setEntries((res as FileEntry[]) || []); setCurrentPath(path)
    } catch (err) { console.error(err) } finally { setLoading(false) }
  }, [call])

  useEffect(() => {
    if (!currentPath) call('DevOps.GetDefaultPath').then(p => { if (p) { setCurrentPath(p as string); fetchDir(p as string) } })
    else fetchDir(currentPath)
  }, [fetchDir, currentPath, call])

  const openFile = async (file: FileEntry) => {
    if (file.is_dir) { setHistory(prev => [...prev, currentPath]); fetchDir(file.path); setPreviewFile(null) }
    else { setPreviewFile(file); setPreviewContent(await call('DevOps.ReadFile', file.path) as string) }
  }

  return (
    <div className="flex h-full p-6 gap-6">
      <div className={cn('flex flex-col flex-1 min-w-0 space-y-3', previewFile ? 'w-1/2' : 'w-full')}>
        <div className="flex items-center gap-3 bg-panel border border-border rounded-xl px-4 py-2">
          <button onClick={() => { const prev = history.pop(); if (prev) fetchDir(prev) }} disabled={history.length === 0} className="p-1 hover:bg-panel-3 rounded-lg"><ChevronRight size={16} className="rotate-180" /></button>
          <span className="text-sm font-mono truncate">{currentPath}</span>
        </div>
        <div className="flex-1 bg-panel border border-border rounded-2xl overflow-y-auto">
          <table className="w-full text-left">
            <tbody>
              {entries.map(file => (
                <tr key={file.path} onClick={() => openFile(file)} className="border-b border-border/10 hover:bg-panel-3 cursor-pointer group">
                  <td className="px-4 py-2 flex items-center gap-3">
                    {file.is_dir ? <Folder size={16} className="text-accent" /> : <FileText size={16} className="text-text-faint" />}
                    <span className="text-sm font-medium group-hover:text-accent">{file.name}</span>
                  </td>
                  <td className="px-4 py-2 text-right text-xs font-mono text-text-faint">{file.size}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
      {previewFile && (
        <div className="w-1/2 flex flex-col space-y-3">
          <div className="flex items-center justify-between"><h3 className="font-bold truncate">{previewFile.name}</h3><button onClick={() => setPreviewFile(null)}><X size={16} /></button></div>
          <div className="flex-1 bg-black border border-border rounded-2xl p-4 overflow-auto font-mono text-sm">
            {previewFile.is_binary ? 'Binary content hidden' : fileContent}
          </div>
        </div>
      )}
    </div>
  )
}

function ToolboxTab() {
  const { call } = useBackend()
  const { data: tools = [] } = useQuery<ToolInfo[]>({
    queryKey: ['devops-tools'],
    queryFn: async () => (await call('DevOps.GetInstalledTools') as ToolInfo[]) || []
  })
  return (
    <div className="grid grid-cols-3 gap-3 overflow-y-auto h-full p-6">
      {tools.map(t => (
        <div key={t.name} className="p-4 rounded-xl border border-border bg-panel flex justify-between items-center">
          <div><p className="font-bold">{t.name}</p><p className="text-xs font-mono text-text-faint">{t.version || 'not found'}</p></div>
          <StatusBadge status={t.status} />
        </div>
      ))}
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
        <div key={`${s.port}-${s.pid}`} className="p-4 rounded-xl border border-border bg-panel flex justify-between items-center">
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
